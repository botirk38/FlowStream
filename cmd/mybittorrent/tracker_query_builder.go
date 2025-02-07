package main

import (
	"net/url"
	"strconv"
)

type TrackerQueryBuilder struct {
	values url.Values
}

func NewTrackerQueryBuilder() *TrackerQueryBuilder {
	return &TrackerQueryBuilder{
		values: make(url.Values),
	}
}

func (b *TrackerQueryBuilder) WithInfoHash(infoHash []byte) *TrackerQueryBuilder {
	b.values.Set("info_hash", string(infoHash))
	return b
}

func (b *TrackerQueryBuilder) WithPeerID(peerID []byte) *TrackerQueryBuilder {
	b.values.Set("peer_id", string(peerID))
	return b
}

func (b *TrackerQueryBuilder) WithPort(port int) *TrackerQueryBuilder {
	b.values.Set("port", strconv.Itoa(port))
	b.values.Set("uploaded", "0")
	b.values.Set("downloaded", "0")
	b.values.Set("compact", "1")
	return b
}

func (b *TrackerQueryBuilder) WithLeft(left int) *TrackerQueryBuilder {
	b.values.Set("left", strconv.Itoa(left))
	return b
}

func (b *TrackerQueryBuilder) Build() url.Values {
	return b.values
}
