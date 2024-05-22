package main

import (
	"fmt"
	bencode "github.com/jackpal/bencode-go"
	"net/http"
	"net/url"
	"strconv"
	"io"
	"encoding/binary"
	"net"
)

type TrackerResponse struct {
	Peers string `bencode:"peers"`
}



const (
	IDChoke        = 0
	IDUnchoke      = 1
	IDInterested   = 2
	IDHave         = 4
	IDBitfield     = 5
	IDRequest      = 6
	IDPiece        = 7
	IDCancel       = 8
)


func getPeers(announceURL string, infoHash []byte, left int) ([]string, error) {
	urlLink, err := url.Parse(announceURL)
	if err != nil {
		return nil, fmt.Errorf("error parsing URL: %w", err)
	}

	query := urlLink.Query()
	query.Set("info_hash", string(infoHash)) 
	query.Set("peer_id", "00112233445566778899") // Should ideally be a randomly generated peer ID
	query.Set("port", "6881")
	query.Set("uploaded", "0")
	query.Set("downloaded", "0")
	query.Set("left", strconv.Itoa(left))
	query.Set("compact", "1")

	urlLink.RawQuery = query.Encode()

	response, err := http.Get(urlLink.String())
	if err != nil {
		return nil, fmt.Errorf("error making HTTP request: %w", err)
	}
	defer response.Body.Close()

	var trackerRes TrackerResponse
	if err := bencode.Unmarshal(response.Body, &trackerRes); err != nil {
		return nil, fmt.Errorf("error decoding tracker response: %w", err)
	}

	return parsePeers(trackerRes.Peers), nil
}

func parsePeers(peersData string) []string {
	peers := []string{}
	for i := 0; i < len(peersData); i += 6 {
		ip := fmt.Sprintf("%d.%d.%d.%d", peersData[i], peersData[i+1], peersData[i+2], peersData[i+3])
		port := int(peersData[i+4])<<8 | int(peersData[i+5])
		peers = append(peers, fmt.Sprintf("%s:%d", ip, port))
	}
	return peers
}

func printPeers(peers []string) {
	for _, peer := range peers {
		fmt.Println(peer)
	}
}

// SendInterested sends an interested message to a peer
func SendInterested(conn net.Conn) error {
	return sendMessage(conn, IDInterested, nil)
}

// SendRequest sends a block request to a peer
func SendRequest(conn net.Conn, index, begin, length int) error {
	data := make([]byte, 12)
	binary.BigEndian.PutUint32(data[0:], uint32(index))
	binary.BigEndian.PutUint32(data[4:], uint32(begin))
	binary.BigEndian.PutUint32(data[8:], uint32(length))
	return sendMessage(conn, IDRequest, data)
}

// sendMessage sends a generic message to a peer
func sendMessage(conn net.Conn, id byte, payload []byte) error {
	length := 1 + len(payload)
	message := make([]byte, 4+length)
	binary.BigEndian.PutUint32(message[0:], uint32(length))
	message[4] = id
	copy(message[5:], payload)
	_, err := conn.Write(message)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}
	return nil
}

// ReadMessage reads a single message from the connection
func ReadMessage(conn net.Conn) (byte, []byte, error) {
	header := make([]byte, 4)
	if _, err := io.ReadFull(conn, header); err != nil {
		return 0, nil, err
	}
	length := binary.BigEndian.Uint32(header)
	if length == 0 { // keep-alive message
		return 0, nil, nil
	}
	message := make([]byte, length)
	if _, err := io.ReadFull(conn, message); err != nil {
		return 0, nil, err
	}
	return message[0], message[1:], nil
}
