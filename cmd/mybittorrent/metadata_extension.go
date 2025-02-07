package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	bencode "github.com/jackpal/bencode-go"
	"net"
)

const (
	MetadataRequestType = 0
)

type MetadataRequest struct {
	MessageType int `bencode:"msg_type"`
	Piece       int `bencode:"piece"`
}

type MetadataResponse struct {
	MessageType int `bencode:"msg_type"`
	Piece       int `bencode:"piece"`
	TotalSize   int `bencode:"total_size"`
}

type MetadataRequestBuilder struct {
	extensionID uint8
	piece       int
}

func NewMetadataRequestBuilder() *MetadataRequestBuilder {
	return &MetadataRequestBuilder{
		piece: 0, // We'll always request piece 0 for this challenge
	}
}

func (b *MetadataRequestBuilder) WithExtensionID(id uint8) *MetadataRequestBuilder {
	b.extensionID = id
	return b
}

func (b *MetadataRequestBuilder) Build() []byte {
	request := MetadataRequest{
		MessageType: MetadataRequestType,
		Piece:       b.piece,
	}

	payload := bytes.Buffer{}
	_ = bencode.Marshal(&payload, request)

	messageLength := uint32(2 + payload.Len()) // 1 for message ID, 1 for extension ID
	message := make([]byte, 4+messageLength)

	binary.BigEndian.PutUint32(message[0:4], messageLength)
	message[4] = ExtensionMessageID
	message[5] = b.extensionID

	copy(message[6:], payload.Bytes())

	return message
}

func sendMetadataRequest(conn net.Conn, extensionID uint8) error {
	message := NewMetadataRequestBuilder().
		WithExtensionID(extensionID).
		Build()

	_, err := conn.Write(message)
	return err
}

func receiveMetadata(conn net.Conn) (*TorrentInfo, error) {
	// Read the metadata response message
	_, payload, err := ReadMessage(conn)
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata message: %w", err)
	}

	// Skip the extension message ID byte
	payloadData := payload[1:]

	// Find where the bencoded dictionary ends
	dictEnd := bytes.Index(payloadData, []byte("ee")) + 2
	if dictEnd == 1 {
		return nil, fmt.Errorf("invalid metadata response format")
	}

	// Parse the metadata response dictionary
	var response MetadataResponse
	if err := bencode.Unmarshal(bytes.NewReader(payloadData[:dictEnd]), &response); err != nil {
		return nil, fmt.Errorf("failed to decode metadata response: %w", err)
	}

	// Extract the actual metadata piece
	metadataBytes := payloadData[dictEnd:]

	// Parse the metadata info dictionary
	var metadata TorrentInfo
	if err := bencode.Unmarshal(bytes.NewReader(metadataBytes), &metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return &metadata, nil
}
