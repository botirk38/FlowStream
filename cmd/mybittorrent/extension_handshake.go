package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	bencode "github.com/jackpal/bencode-go"
	"net"
)

// Constants for extension protocol
const (
	ExtensionMessageID   = 20
	HandshakeExtensionID = 0
	UTMetadataID         = 1
)

// ExtensionMessage represents a BitTorrent extension message
type ExtensionMessage struct {
	length    uint32
	messageID uint8
	payload   []byte
}

// ExtensionMessageBuilder handles building extension messages
type ExtensionMessageBuilder struct {
	message ExtensionMessage
}

// NewExtensionMessageBuilder creates a new extension message builder
func NewExtensionMessageBuilder() *ExtensionMessageBuilder {
	return &ExtensionMessageBuilder{
		message: ExtensionMessage{
			messageID: ExtensionMessageID,
		},
	}
}

// WithHandshakePayload adds handshake payload to the message
func (b *ExtensionMessageBuilder) WithHandshakePayload() *ExtensionMessageBuilder {
	handshake := struct {
		M map[string]int `bencode:"m"`
	}{
		M: map[string]int{
			"ut_metadata": UTMetadataID,
		},
	}

	var buf bytes.Buffer
	_ = bencode.Marshal(&buf, handshake)
	b.message.payload = buf.Bytes()
	b.message.length = uint32(2 + len(b.message.payload)) // 1 byte for message ID, 1 byte for extension ID

	return b
}

// Build constructs the final extension message
func (b *ExtensionMessageBuilder) Build() []byte {
	buf := make([]byte, 4+b.message.length)

	// Write message length
	binary.BigEndian.PutUint32(buf[0:4], b.message.length)

	// Write extension message ID
	buf[4] = b.message.messageID

	// Write handshake extension ID
	buf[5] = HandshakeExtensionID

	// Write payload
	copy(buf[6:], b.message.payload)

	return buf
}

func sendExtensionHandshake(conn net.Conn) error {
	message := NewExtensionMessageBuilder().
		WithHandshakePayload().
		Build()

	_, err := conn.Write(message)
	if err != nil {
		return fmt.Errorf("failed to send extension handshake: %w", err)
	}

	return nil
}

func supportsExtensions(handshake []byte) bool {
	// Check the 20th bit (5th byte, bit 4) for extension protocol support
	return (handshake[ExtensionMessageID+5] & 0x10) == 0x10
}
