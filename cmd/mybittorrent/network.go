package main

import (
	"bytes"
	"fmt"
	"net"
)

type PeerConnection struct {
	InfoHash []byte
	PeerID   string
	Conn     net.Conn
}

func NewPeerConnection(peerAddr string, infoHash []byte) (*PeerConnection, error) {
	conn, err := net.Dial("tcp", peerAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to dial peer: %w", err)
	}

	// Preparing the handshake message
	buffer := new(bytes.Buffer)
	buffer.WriteByte(19) // Protocol string length
	buffer.WriteString("BitTorrent protocol")
	buffer.Write(make([]byte, 8)) // Reserved bytes
	buffer.Write(infoHash)
	buffer.WriteString("00112233445566778899") // Peer ID, should ideally be your client's peer ID

	// Sending the handshake
	_, err = conn.Write(buffer.Bytes())
	if err != nil {
		return nil, fmt.Errorf("failed to send handshake: %w", err)
	}

	// Reading the response
	response := make([]byte, 1024)
	_, err = conn.Read(response)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	return &PeerConnection{
		InfoHash: infoHash,
		PeerID:   fmt.Sprintf("%x", response[48:68]),
		Conn:     conn,
	}, nil
}

