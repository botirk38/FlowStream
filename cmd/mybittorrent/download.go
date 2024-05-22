package main

import (
	"encoding/binary"
	"fmt"
	"os"
	"net"
	"io"
	"bytes"
	"math"
	"path/filepath"
)

type PeerMessage struct {
	lengthPrefix uint32
	id           uint8
	index        uint32
	begin        uint32
	length       uint32
}

func DownloadPiece(peerConn *PeerConnection, torrent *Torrent, pieceIndex int, outputPath string) error {
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Calculate the number of blocks and adjust the size of the last block of the last piece
	totalPieces := (torrent.Info.Length + torrent.Info.PieceLength - 1) / torrent.Info.PieceLength
	pieceSize := torrent.Info.PieceLength
	if pieceIndex == totalPieces-1 { // If it's the last piece
		lastPieceSize := torrent.Info.Length % torrent.Info.PieceLength
		if lastPieceSize != 0 {
			pieceSize = lastPieceSize
		}
	}
	numBlocks := int(math.Ceil(float64(pieceSize) / 16384))

	// Send interested message and wait for unchoke
	if err := sendInterestedMessage(peerConn.Conn); err != nil {
		return err
	}
	if err := waitForUnchoke(peerConn.Conn); err != nil {
		return err
	}

	var pieceData []byte
	for i := 0; i < numBlocks; i++ {
		begin := i * 16384
		blockSize := 16384
		if i == numBlocks-1 { // Adjust the size of the last block
			blockSize = pieceSize - begin
		}

		if err := sendRequest(peerConn.Conn, uint32(pieceIndex), uint32(begin), uint32(blockSize)); err != nil {
			return err
		}

		blockData, err := receiveBlock(peerConn.Conn)
		if err == io.EOF {
			fmt.Println("End of file reached, no more data available.")
			break
		} else if err != nil {
			return err
		}
		pieceData = append(pieceData, blockData...)
		fmt.Printf("Block %d downloaded.\n", i)
	}

	if len(pieceData) != pieceSize {
		return fmt.Errorf("incomplete piece: received %d bytes, expected %d bytes", len(pieceData), pieceSize)
	}

	if err := os.WriteFile(outputPath, pieceData, 0644); err != nil {
		return fmt.Errorf("failed to write piece to disk: %w", err)
	}

	fmt.Printf("Piece %d downloaded to %s.\n", pieceIndex, outputPath)
	return nil
}


func sendInterestedMessage(conn net.Conn) error {
	msg := []byte{0, 0, 0, 1, 2} // lengthPrefix = 1, ID = 2 (interested)
	_, err := conn.Write(msg)
	return err
}

func waitForUnchoke(conn net.Conn) error {
	for {
		header := make([]byte, 4)
		if _, err := io.ReadFull(conn, header); err != nil {
			return err
		}

		length := binary.BigEndian.Uint32(header)
		if length == 0 { // Keep-alive message
			continue
		}

		payload := make([]byte, length)
		if _, err := io.ReadFull(conn, payload); err != nil {
			return err
		}

		if payload[0] == 1 { // ID for unchoke
			break
		}
	}
	return nil
}

func sendRequest(conn net.Conn, index, begin, length uint32) error {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, uint32(13)) // length of the message including ID
	binary.Write(buf, binary.BigEndian, uint8(6))   // IDRequest
	binary.Write(buf, binary.BigEndian, index)
	binary.Write(buf, binary.BigEndian, begin)
	binary.Write(buf, binary.BigEndian, length)
	_, err := conn.Write(buf.Bytes())
	return err
}

func receiveBlock(conn net.Conn) ([]byte, error) {
	header := make([]byte, 4)
	if _, err := io.ReadFull(conn, header); err != nil {
		return nil, err
	}

	length := binary.BigEndian.Uint32(header)
	payload := make([]byte, length)
	if _, err := io.ReadFull(conn, payload); err != nil {
		return nil, err
	}

	if payload[0] != 7 { // ID for piece
		return nil, fmt.Errorf("expected piece message, got ID %d", payload[0])
	}

	// Return the block data excluding the first 9 bytes (index, begin)
	return payload[9:], nil
}
