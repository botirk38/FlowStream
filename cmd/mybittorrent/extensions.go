package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	bencode "github.com/jackpal/bencode-go"
	"io"
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

type ExtensionHandshake struct {
	M map[string]int `bencode:"m"`
}

type ExtensionMessageReader struct {
	conn io.Reader
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

func NewExtensionMessageReader(conn io.Reader) *ExtensionMessageReader {
	return &ExtensionMessageReader{conn: conn}
}

func (r *ExtensionMessageReader) ReadExtensionMessage() (*ExtensionHandshake, error) {
	lengthBuf := make([]byte, 4)
	if _, err := io.ReadFull(r.conn, lengthBuf); err != nil {
		return nil, err
	}

	length := binary.BigEndian.Uint32(lengthBuf)
	if length == 0 {
		return nil, nil
	}

	message := make([]byte, length)
	if _, err := io.ReadFull(r.conn, message); err != nil {
		return nil, err
	}

	// Skip the message ID and extension ID
	payload := message[2:]

	var handshake ExtensionHandshake
	if err := bencode.Unmarshal(bytes.NewReader(payload), &handshake); err != nil {
		return nil, err
	}

	return &handshake, nil
}

func GetMetadataExtensionID(handshake *ExtensionHandshake) int {
	if id, ok := handshake.M["ut_metadata"]; ok {
		return id
	}
	return 0
}
