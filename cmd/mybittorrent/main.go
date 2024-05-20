package main

import (
	"encoding/json"
	"fmt"
	"os"

)

type TrackerResponse struct {
	Peers string `bencode:"peers"`
}

// Example:
// - 5:hello -> hello
// - 10:hello12345 -> hello12345

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.

	command := os.Args[1]
	startIndex := 0

	switch command {
	case "decode":
		bencodedValue := os.Args[2]

		decoded, err := Decode(bencodedValue, &startIndex)
		if err != nil {
			fmt.Println(err)
			return
		}

		jsonOutput, _ := json.Marshal(decoded)
		fmt.Println(string(jsonOutput))

	case "info":
		torrentFile := os.Args[2]

		torrent, err := readTorrentFile(torrentFile)

		if err != nil {
			fmt.Println(err)
			return
		}

		infoHash, err := calculateInfoHash(&torrent.Info)

		if err != nil {
			fmt.Println(err)
			return
		}

		printTorrentDetails(torrent, infoHash)

	case "peers":
		torrentFile := os.Args[2]

		torrent, err := readTorrentFile(torrentFile)

		if err != nil {
			fmt.Println(err)
			return
		}

		infoHash, err := calculateInfoHash(&torrent.Info)

		if err != nil {
			fmt.Println(err)
			return
		}

		peers, err := getPeers(torrent.Announce, infoHash, torrent.Info.Length)

		if err != nil {
			fmt.Println(err)
			return
		}

		printPeers(peers)

	default:
		fmt.Println("Invalid command")
		os.Exit(1)

	}

}
