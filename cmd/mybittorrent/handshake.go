package main

import (
	"bytes"
	"fmt"
	"io"
	"net"
)

const (
	ProtocolLength  = 19
	ProtocolString  = "BitTorrent protocol"
	ReservedBytes   = 8
	InfoHashLength  = 20
	PeerIDLength    = 20
	HandshakeLength = 1 + ProtocolLength + ReservedBytes + InfoHashLength + PeerIDLength
)

// HandshakeBuilder handles the construction of BitTorrent handshake messages
type HandshakeBuilder struct {
	protocol []byte
	reserved []byte
	infoHash []byte
	peerID   []byte
}

// NewHandshakeBuilder creates a new HandshakeBuilder with default protocol
func NewHandshakeBuilder() *HandshakeBuilder {
	return &HandshakeBuilder{
		protocol: []byte(ProtocolString),
		reserved: make([]byte, ReservedBytes),
	}
}

// WithInfoHash sets the info hash
func (b *HandshakeBuilder) WithInfoHash(infoHash []byte) *HandshakeBuilder {
	if len(infoHash) != InfoHashLength {
		panic("invalid info hash length")
	}
	b.infoHash = make([]byte, InfoHashLength)
	copy(b.infoHash, infoHash)
	return b
}

// WithPeerID sets the peer ID
func (b *HandshakeBuilder) WithPeerID(peerID []byte) *HandshakeBuilder {
	if len(peerID) != PeerIDLength {
		panic("invalid peer ID length")
	}
	b.peerID = make([]byte, PeerIDLength)
	copy(b.peerID, peerID)
	return b
}

// WithExtensions sets the reserved bytes with extension bit
func (b *HandshakeBuilder) WithExtensions() *HandshakeBuilder {
	reserved := make([]byte, ReservedBytes)
	// Set the 20th bit from right (5th byte, 5th bit)
	reserved[5] = 0x10
	b.reserved = reserved
	return b
}

// Build constructs and returns the complete handshake message
func (b *HandshakeBuilder) Build() []byte {
	if b.infoHash == nil || b.peerID == nil {
		panic("info hash and peer ID are required")
	}

	buffer := bytes.NewBuffer(make([]byte, 0, HandshakeLength))
	buffer.WriteByte(byte(ProtocolLength))
	buffer.Write(b.protocol)
	buffer.Write(b.reserved)
	buffer.Write(b.infoHash)
	buffer.Write(b.peerID)

	return buffer.Bytes()
}

func createHandshake(infoHash, peerID []byte) []byte {
	return NewHandshakeBuilder().
		WithInfoHash(infoHash).
		WithPeerID(peerID).
		Build()
}

func createMagnetHandshake(infoHash, peerID []byte) []byte {
	return NewHandshakeBuilder().
		WithInfoHash(infoHash).
		WithPeerID(peerID).
		WithExtensions().
		Build()
}

func sendHandshake(conn net.Conn, handshake []byte) error {
	_, err := conn.Write(handshake)
	if err != nil {
		return fmt.Errorf("failed to send handshake: %w", err)
	}
	return nil
}

func readHandshake(conn net.Conn) ([]byte, error) {
	response := make([]byte, HandshakeLength)
	_, err := io.ReadFull(conn, response)
	if err != nil {
		return nil, fmt.Errorf("failed to read handshake: %w", err)
	}
	return response, nil
}

func handleExtensionHandshake(conn net.Conn) (int, error) {
	reader := NewExtensionMessageReader(conn)
	handshake, err := reader.ReadExtensionMessage()
	if err != nil {
		return 0, err
	}

	return GetMetadataExtensionID(handshake), nil
}

