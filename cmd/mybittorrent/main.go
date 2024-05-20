package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"

	bencode "github.com/jackpal/bencode-go" // Available if you need it!
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

		urlLink, err := url.Parse(torrent.Announce)

		if err != nil {
			fmt.Println(err)
			return
		}

		query := urlLink.Query()
		query.Set("info_hash", string(infoHash))
		query.Set("peer_id", "00112233445566778898")
		query.Set("port", "6881")
		query.Set("uploaded", "0")
		query.Set("downloaded", "0")
		query.Set("left", strconv.Itoa(torrent.Info.PieceLength))
		query.Set("compact", "1")

		urlLink.RawQuery = query.Encode()

		response, err := http.Get(urlLink.String())

		if err != nil {
			fmt.Println(err)
			return
		}

		defer response.Body.Close()

		var trackerRes TrackerResponse

		if err := bencode.Unmarshal(response.Body, &trackerRes); err != nil {
			fmt.Println(err)
			return
		}

		for i := 0; i < len(trackerRes.Peers); i += 6 {
			ip := fmt.Sprintf("%d.%d.%d.%d", trackerRes.Peers[i], trackerRes.Peers[i+1], trackerRes.Peers[i+2], trackerRes.Peers[i+3])
			port := int(trackerRes.Peers[i+4])<<8 | int(trackerRes.Peers[i+5])
			fmt.Printf("%s:%d\n", ip, port)
		}

	default:
		fmt.Println("Invalid command")
		os.Exit(1)

	}

}
