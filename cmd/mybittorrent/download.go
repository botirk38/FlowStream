package main

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"math"
	"sync"
)

func DownloadPiece(peerConn *PeerConnection, torrent *Torrent, pieceIndex int) ([]byte, error) {
	totalPieces := (torrent.Info.Length + torrent.Info.PieceLength - 1) / torrent.Info.PieceLength
	pieceSize := torrent.Info.PieceLength
	if pieceIndex == totalPieces-1 {
		lastPieceSize := torrent.Info.Length % torrent.Info.PieceLength
		if lastPieceSize != 0 {
			pieceSize = lastPieceSize
		}
	}
	numBlocks := int(math.Ceil(float64(pieceSize) / 16384))

	if err := sendInterestedMessage(peerConn.Conn); err != nil {
		return nil, err
	}
	if err := waitForUnchoke(peerConn.Conn); err != nil {
		return nil, err
	}

	pieceData := make([]byte, pieceSize)
	pendingRequests := make(chan int, 5) // Buffer for 5 pending requests
	results := make(chan struct {
		blockIndex int
		data       []byte
		err        error
	}, 5)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for blockIndex := range pendingRequests {
			begin := blockIndex * 16384
			blockSize := 16384
			if blockIndex == numBlocks-1 {
				blockSize = pieceSize - begin
			}
			if err := sendRequest(peerConn.Conn, uint32(pieceIndex), uint32(begin), uint32(blockSize)); err != nil {
				results <- struct {
					blockIndex int
					data       []byte
					err        error
				}{blockIndex, nil, err}
				return
			}
			blockData, err := receiveBlock(peerConn.Conn)
			results <- struct {
				blockIndex int
				data       []byte
				err        error
			}{blockIndex, blockData, err}
		}
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	for i := 0; i < numBlocks; i++ {
		pendingRequests <- i
		if i >= 4 { // Start collecting results after sending 5 requests
			result := <-results
			if result.err != nil {
				close(pendingRequests)
				return nil, result.err
			}
			copy(pieceData[result.blockIndex*16384:], result.data)
			fmt.Printf("Block %d downloaded.\n", result.blockIndex)
		}
	}
	close(pendingRequests)

	// Collect remaining results
	for result := range results {
		if result.err != nil {
			return nil, result.err
		}
		copy(pieceData[result.blockIndex*16384:], result.data)
		fmt.Printf("Block %d downloaded.\n", result.blockIndex)
	}

	if len(pieceData) != pieceSize {
		return nil, fmt.Errorf("incomplete piece: received %d bytes, expected %d bytes", len(pieceData), pieceSize)
	}

	return pieceData, nil
}

func DownloadFile(torrent *Torrent, peers []string, infoHash []byte) ([]byte, error) {
	numPieces := len(torrent.Info.Pieces) / 20
	fileSize := torrent.Info.Length

	pieces := make([][]byte, numPieces)
	pieceChannel := make(chan int, numPieces)
	resultChannel := make(chan struct {
		index int
		data  []byte
		err   error
	}, numPieces)

	// Initialize piece channel
	for i := 0; i < numPieces; i++ {
		pieceChannel <- i
	}

	// Use fewer workers to reduce connection overhead
	maxWorkers := 3
	var wg sync.WaitGroup

	// Start workers
	for i := 0; i < maxWorkers && i < len(peers); i++ {
		wg.Add(1)
		go func(peerAddr string) {
			defer wg.Done()

			// Create a single connection per worker
			peerConn, err := NewPeerConnection(peerAddr, infoHash)
			if err != nil {
				return
			}
			defer peerConn.Conn.Close()

			for pieceIndex := range pieceChannel {
				pieceData, err := DownloadPiece(peerConn, torrent, pieceIndex)
				if err != nil {
					resultChannel <- struct {
						index int
						data  []byte
						err   error
					}{pieceIndex, nil, err}
					continue
				}

				if verifyPiece(pieceData, []byte(torrent.Info.Pieces[pieceIndex*20:(pieceIndex+1)*20])) {
					resultChannel <- struct {
						index int
						data  []byte
						err   error
					}{pieceIndex, pieceData, nil}
				} else {
					resultChannel <- struct {
						index int
						data  []byte
						err   error
					}{pieceIndex, nil, fmt.Errorf("piece verification failed")}
				}
			}
		}(peers[i])
	}

	// Close channels when all workers are done
	go func() {
		wg.Wait()
		close(resultChannel)
		close(pieceChannel)
	}()

	// Collect results with timeout
	remainingPieces := numPieces
	for result := range resultChannel {
		if result.err != nil {
			// Retry failed piece
			select {
			case pieceChannel <- result.index:
			default:
				return nil, fmt.Errorf("failed to download piece %d: %v", result.index, result.err)
			}
		} else {
			pieces[result.index] = result.data
			remainingPieces--
		}

		if remainingPieces == 0 {
			break
		}
	}

	// Combine all pieces
	fileData := make([]byte, 0, fileSize)
	for _, piece := range pieces {
		if piece == nil {
			return nil, fmt.Errorf("incomplete download: missing pieces")
		}
		fileData = append(fileData, piece...)
	}

	return fileData[:fileSize], nil
}

func verifyPiece(pieceData []byte, expectedHash []byte) bool {
	hash := sha1.Sum(pieceData)
	return bytes.Equal(hash[:], expectedHash)
}
