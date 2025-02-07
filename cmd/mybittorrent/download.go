package main

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"math"
	"sync"
)

const (
	maxRetries    = 3 // Maximum retries for piece download
	maxConcurrent = 5 // Maximum concurrent piece downloads
)

type pieceWork struct {
	index  int
	length int
	hash   []byte
	offset int64
}

func DownloadPiece(peerConn *PeerConnection, torrent *Torrent, pieceIndex int) ([]byte, error) {
	if err := setupConnection(peerConn); err != nil {
		return nil, fmt.Errorf("Failed to establish connection: %w", err)

	}

	pieceLength := calculatePieceLength(torrent, pieceIndex)
	return downloadPiece(peerConn, pieceIndex, pieceLength)
}

func downloadPiece(peerConn *PeerConnection, pieceIndex int, pieceLength int) ([]byte, error) {
	numBlocks := int(math.Ceil(float64(pieceLength) / float64(BlockSize)))
	pieceData := make([]byte, pieceLength)

	for blockIndex := 0; blockIndex < numBlocks; blockIndex++ {
		begin := blockIndex * BlockSize
		length := calculateBlockLength(begin, BlockSize, pieceLength)

		if err := downloadBlock(peerConn, pieceIndex, begin, length, pieceData); err != nil {
			return nil, fmt.Errorf("download block %d: %w", blockIndex, err)
		}
		fmt.Printf("Block %d downloaded.\n", blockIndex)
	}

	return pieceData, nil
}

func DownloadFile(torrent *Torrent, peers []string, infoHash []byte) ([]byte, error) {
	if len(peers) == 0 {
		return nil, fmt.Errorf("no peers available")
	}

	workQueue := make(chan pieceWork, len(torrent.Info.Pieces)/20)
	results := make(chan pieceWork, len(torrent.Info.Pieces)/20)

	// Initialize work queue with piece offsets
	initializeWorkQueue(torrent, workQueue)

	// Start workers
	var workers sync.WaitGroup
	numWorkers := min(len(peers), maxConcurrent)
	for i := 0; i < numWorkers; i++ {
		workers.Add(1)
		go func(peerAddr string) {
			defer workers.Done()
			worker(peerAddr, infoHash, workQueue, results)
		}(peers[i])
	}

	// Close results when all workers are done
	go func() {
		workers.Wait()
		close(results)
	}()

	// Collect and assemble results
	fileData := make([]byte, torrent.Info.Length)
	var downloaded int
	numPieces := len(torrent.Info.Pieces) / 20

	for piece := range results {
		copy(fileData[piece.offset:], piece.hash)
		downloaded++
		if downloaded == numPieces {
			break
		}
	}

	return fileData, nil
}

func worker(peerAddr string, infoHash []byte, workQueue chan pieceWork, results chan pieceWork) {
	peerConn, err := NewPeerConnection(peerAddr, infoHash)
	if err != nil {
		return
	}
	defer peerConn.Conn.Close()

	if err := setupConnection(peerConn); err != nil {
		return
	}

	for work := range workQueue {
		for retry := 0; retry < maxRetries; retry++ {
			pieceData, err := downloadPiece(peerConn, work.index, work.length)
			if err != nil {
				continue
			}

			if verifyPiece(pieceData, work.hash) {
				work.hash = pieceData // Store the actual data in the hash field
				results <- work
				fmt.Printf("Piece %d downloaded and verified.\n", work.index)
				break
			}
		}
	}
}

func initializeWorkQueue(torrent *Torrent, workQueue chan pieceWork) {
	numPieces := len(torrent.Info.Pieces) / 20
	var offset int64

	for i := 0; i < numPieces; i++ {
		pieceLength := calculatePieceLength(torrent, i)
		work := pieceWork{
			index:  i,
			length: pieceLength,
			hash:   []byte(torrent.Info.Pieces[i*20 : (i+1)*20]),
			offset: offset,
		}
		offset += int64(pieceLength)
		workQueue <- work
	}
	close(workQueue)
}

func calculatePieceLength(torrent *Torrent, pieceIndex int) int {
	totalPieces := (torrent.Info.Length + torrent.Info.PieceLength - 1) / torrent.Info.PieceLength
	if pieceIndex == totalPieces-1 {
		lastPieceSize := torrent.Info.Length % torrent.Info.PieceLength
		if lastPieceSize != 0 {
			return lastPieceSize
		}
	}
	return torrent.Info.PieceLength
}

func calculateBlockLength(begin, blockSize, pieceLength int) int {
	if begin+blockSize > pieceLength {
		return pieceLength - begin
	}
	return blockSize
}

func downloadBlock(peerConn *PeerConnection, pieceIndex, begin, length int, pieceData []byte) error {
	if err := sendRequest(peerConn.Conn, uint32(pieceIndex), uint32(begin), uint32(length)); err != nil {
		return err
	}

	blockData, err := receiveBlock(peerConn.Conn)
	if err != nil {
		return err
	}

	copy(pieceData[begin:], blockData)
	return nil
}

func verifyPiece(pieceData []byte, expectedHash []byte) bool {
	hash := sha1.Sum(pieceData)
	return bytes.Equal(hash[:], expectedHash)
}
