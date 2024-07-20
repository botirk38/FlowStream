package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

// Example:
// - 5:hello -> hello
// - 10:hello12345 -> hello12345

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.

	command := os.Args[1]
	args := os.Args[2:]
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

		peerConnection, err := NewPeerConnection(peerAddr, infoHash)

		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Printf("Peer ID: %s\n", peerConnection.PeerID)
	case "download_piece":
		var torrentFile string
		var outputFile string
		var pieceIndexStr string

		if args[0] == "-o" {
			outputFile = args[1]
			torrentFile = args[2]
			pieceIndexStr = args[3]
		} else {
			torrentFile = args[0]
			pieceIndexStr = args[1]
			outputFile = "piece-" + pieceIndexStr
		}

		pieceIndex, err := strconv.Atoi(pieceIndexStr)
		if err != nil {
			fmt.Println("Invalid piece index:", err)
			return
		}

		torrent, err := readTorrentFile(torrentFile)
		if err != nil {
			fmt.Println("Error reading torrent file:", err)
			return
		}

		infoHash, err := calculateInfoHash(&torrent.Info)
		if err != nil {
			fmt.Println("Error calculating info hash:", err)
			return
		}

		peers, err := getPeers(torrent.Announce, infoHash, torrent.Info.Length)
		if err != nil {
			fmt.Println("Error getting peers:", err)
			return
		}

		if len(peers) == 0 {
			fmt.Println("No peers available")
			return
		}

		peerAddr := peers[0] // Use the first available peer
		peerConnection, err := NewPeerConnection(peerAddr, infoHash)
		if err != nil {
			fmt.Println("Error establishing connection to peer:", err)
			return
		}

		fmt.Printf("Output file: %s\n", outputFile)

		pieceData, err := DownloadPiece(peerConnection, torrent, pieceIndex)
		if err != nil {
			fmt.Println("Error downloading piece:", err)
			return
		}

		if err := os.WriteFile(outputFile, pieceData, 0644); err != nil {
			fmt.Printf("Failed to write piece to disk: %s\n", err)
			return
		}

		fmt.Printf("Piece %d downloaded to %s.\n", pieceIndex, outputFile)

	case "download":
		var outputFile string
		var torrentFile string
		if args[0] == "-o" {
			outputFile = args[1]
			torrentFile = args[2]
		} else {
			fmt.Println("Invalid arguments for download command")
			return
		}

		torrent, err := readTorrentFile(torrentFile)
		if err != nil {
			fmt.Println("Error reading torrent file:", err)
			return
		}

		infoHash, err := calculateInfoHash(&torrent.Info)
		if err != nil {
			fmt.Println("Error calculating info hash:", err)
			return
		}

		peers, err := getPeers(torrent.Announce, infoHash, torrent.Info.Length)
		if err != nil {
			fmt.Println("Error getting peers:", err)
			return
		}

		if len(peers) == 0 {
			fmt.Println("No peers available.")
			return
		}

		fileData, err := DownloadFile(torrent, peers, infoHash)
		if err != nil {
			fmt.Println("Error downloading file:", err)
			return
		}

		err = os.WriteFile(outputFile, fileData, 0644)
		if err != nil {
			fmt.Println("Error saving file:", err)
			return
		}

		fmt.Printf("Downloaded %s to %s.\n", filepath.Base(torrentFile), outputFile)

	default:
		fmt.Println("Invalid command")
		os.Exit(1)

	}

}
