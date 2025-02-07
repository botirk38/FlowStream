package main

import (
	"bytes"
	"encoding/binary"
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
