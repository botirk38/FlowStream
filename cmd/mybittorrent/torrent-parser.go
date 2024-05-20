package main

import (
	"crypto/sha1"
	"fmt"
	bencode "github.com/jackpal/bencode-go"
	"os"
)

func parseTorrentFile(torrentFilePath string, startIndex int) {

	torrentInfo, err := os.ReadFile(torrentFilePath)

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

	announce := torrentDict["announce"].(string)
	infoDict := torrentDict["info"].(map[string]interface{})
	fileLength := infoDict["length"].(int)

	fmt.Println("Tracker URL:", announce)
	fmt.Println("File Length:", fileLength)

	hash := sha1.New()

	if err := bencode.Marshal(hash, infoDict); err != nil {
		fmt.Println(err)
		return
	}

	infoHash := hash.Sum(nil)

	fmt.Printf("Info Hash: %x\n", infoHash)

}
