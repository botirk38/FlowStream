package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
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

	case "handshake":
		torrentFile := os.Args[2]
		peerAddr := os.Args[3]

		torrent, err := readTorrentFile(torrentFile)

		if err != nil {
			fmt.Println(err)
			return
		}

		infoHash, err := calculateInfoHash(&torrent.Info)

		fmt.Printf("Info Hash: %x\n", infoHash)

		if err != nil {
			fmt.Println(err)
			return
		}

		conn, err := net.Dial("tcp", peerAddr)

		buffer := new(bytes.Buffer)
		buffer.WriteByte(19)
		buffer.WriteString("BitTorrent protocol")
		buffer.Write(make([]byte, 8))
		buffer.Write(infoHash)
		buffer.WriteString("00112233445566778899")

		_, err = conn.Write(buffer.Bytes())

		if err != nil {
			fmt.Println(err)
			return
		}

		response := make([]byte, 1024)

		_, err = conn.Read(response)

		if err != nil {
			fmt.Println(err)
			return
		}

		peerId := response[48:68]

		fmt.Printf("Peer ID: %x\n", peerId)

		err = conn.Close()

		if err != nil {
			fmt.Println(err)
			return
		}

	default:
		fmt.Println("Invalid command")
		os.Exit(1)

	}

}
