package main

import "encoding/binary"

type RequestPayloadBuilder struct {
	index  uint32
	begin  uint32
	length uint32
}

func NewRequestPayloadBuilder() *RequestPayloadBuilder {
	return &RequestPayloadBuilder{}
}

func (b *RequestPayloadBuilder) WithIndex(index uint32) *RequestPayloadBuilder {
	b.index = index
	return b
}

func (b *RequestPayloadBuilder) WithBegin(begin uint32) *RequestPayloadBuilder {
	b.begin = begin
	return b
}

func (b *RequestPayloadBuilder) WithLength(length uint32) *RequestPayloadBuilder {
	b.length = length
	return b
}

func (b *RequestPayloadBuilder) Build() []byte {
	payload := make([]byte, 12)
	binary.BigEndian.PutUint32(payload[0:4], b.index)
	binary.BigEndian.PutUint32(payload[4:8], b.begin)
	binary.BigEndian.PutUint32(payload[8:12], b.length)
	return payload
}
