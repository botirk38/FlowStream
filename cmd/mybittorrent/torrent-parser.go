package main

import (
	"crypto/sha1"
	"fmt"
	bencode "github.com/jackpal/bencode-go"
	"os"
)

type Torrent struct {
	Announce string
	Info     TorrentInfo `bencode:"info"`
}

type TorrentInfo struct {
	Length      int    `bencode:"length"`
	Name        string `bencode:"name"`
	PieceLength int    `bencode:"piece length"`
	Pieces      string `bencode:"pieces"`
}

func parseTorrentFile(torrentFilePath string, startIndex int) {

	torrentInfo, err := os.ReadFile(torrentFilePath)
	var torrent Torrent

	if err != nil {
		fmt.Println(err)
		return
	}

	decoded, err := Decode(string(torrentInfo), &startIndex)

	if err != nil {
		fmt.Println(err)
		return

	}

	torrentDict, ok := decoded.(map[string]interface{})

	if !ok {
		fmt.Println("Error: Decoded value is not a dictionary")
		return
	}

	torrent.Announce = torrentDict["announce"].(string)
	torrent.Info = TorrentInfo{
		Length:      torrentDict["info"].(map[string]interface{})["length"].(int),
		Name:        torrentDict["info"].(map[string]interface{})["name"].(string),
		PieceLength: torrentDict["info"].(map[string]interface{})["piece length"].(int),
		Pieces:      torrentDict["info"].(map[string]interface{})["pieces"].(string),
	}

	hash := sha1.New()

	if err := bencode.Marshal(hash, torrent.Info); err != nil {
		fmt.Println(err)
		return
	}

	infoHash := hash.Sum(nil)

	fmt.Printf("Tracker URL: %s\n", torrent.Announce)
	fmt.Printf("Length: %d\n", torrent.Info.Length)

	fmt.Printf("Info Hash: %x\n", infoHash)
	fmt.Printf("Piece Length: %d\n", torrent.Info.PieceLength)

	fmt.Printf("Piece Hashes: \n")

	for i := 0; i < len(torrent.Info.Pieces); i += 20 {
		fmt.Printf("%x\n", torrent.Info.Pieces[i:i+20])
	}

}
