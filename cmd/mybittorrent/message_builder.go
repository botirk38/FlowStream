package main

import (
	"encoding/binary"
)

// MessageBuilder handles construction of BitTorrent protocol messages
type MessageBuilder struct {
	id      uint8
	payload []byte
}

// NewMessageBuilder creates a new message builder instance
func NewMessageBuilder() *MessageBuilder {
	return &MessageBuilder{}
}

// WithID sets the message ID
func (b *MessageBuilder) WithID(id uint8) *MessageBuilder {
	b.id = id
	return b
}

// WithPayload sets the message payload
func (b *MessageBuilder) WithPayload(payload []byte) *MessageBuilder {
	b.payload = payload
	return b
}

// Build constructs the final message
func (b *MessageBuilder) Build() []byte {
	length := uint32(1)
	if b.payload != nil {
		length += uint32(len(b.payload))
	}

	buf := make([]byte, 4+length)
	binary.BigEndian.PutUint32(buf[0:4], length)
	buf[4] = b.id

	if b.payload != nil {
		copy(buf[5:], b.payload)
	}

	return buf
}

// Convenience methods for common message types
func NewChokeMessage() []byte {
	return NewMessageBuilder().WithID(IDChoke).Build()
}

func NewUnchokeMessage() []byte {
	return NewMessageBuilder().WithID(IDUnchoke).Build()
}

func NewInterestedMessage() []byte {
	return NewMessageBuilder().WithID(IDInterested).Build()
}

func NewNotInterestedMessage() []byte {
	return NewMessageBuilder().WithID(IDNotInterested).Build()
}

func NewHaveMessage(pieceIndex uint32) []byte {
	payload := make([]byte, 4)
	binary.BigEndian.PutUint32(payload, pieceIndex)
	return NewMessageBuilder().WithID(IDHave).WithPayload(payload).Build()
}

func NewRequestMessage(index, begin, length uint32) []byte {
	payload := make([]byte, 12)
	binary.BigEndian.PutUint32(payload[0:4], index)
	binary.BigEndian.PutUint32(payload[4:8], begin)
	binary.BigEndian.PutUint32(payload[8:12], length)
	return NewMessageBuilder().WithID(IDRequest).WithPayload(payload).Build()
}
