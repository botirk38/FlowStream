package main

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	bencode "github.com/jackpal/bencode-go"
	"os"
)

type Torrent struct {
	Announce string      `bencode:"announce"`
	Info     TorrentInfo `bencode:"info"`
}

type TorrentInfo struct {
	Length      int    `bencode:"length"`
	Name        string `bencode:"name"`
	PieceLength int    `bencode:"piece length"`
	Pieces      string `bencode:"pieces"`
}

// Read and decode a torrent file.
func readTorrentFile(filePath string) (*Torrent, error) {
	fileData, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	defer fileData.Close()

	var torrent Torrent
	if err := bencode.Unmarshal(fileData, &torrent); err != nil {
		return nil, fmt.Errorf("error unmarshalling torrent data: %w", err)
	}

	return &torrent, nil
}

func calculateInfoHash(info *TorrentInfo) ([]byte, error) {
	var b bytes.Buffer
	if err := bencode.Marshal(&b, *info); err != nil {
		return nil, fmt.Errorf("error marshalling torrent info for hashing: %w", err)
	}

	hash := sha1.New()
	hash.Write(b.Bytes())
	return hash.Sum(nil), nil
}

// Print torrent details.
func printTorrentDetails(torrent *Torrent, infoHash []byte) {
	fmt.Printf("Tracker URL: %s\n", torrent.Announce)
	fmt.Printf("Length: %d\n", torrent.Info.Length)
	fmt.Printf("Info Hash: %x\n", infoHash)
	fmt.Printf("Piece Length: %d\n", torrent.Info.PieceLength)

	fmt.Println("Piece Hashes: ")
	for i := 0; i < len(torrent.Info.Pieces); i += 20 {
		fmt.Printf("%x\n", torrent.Info.Pieces[i:i+20])
	}
}
