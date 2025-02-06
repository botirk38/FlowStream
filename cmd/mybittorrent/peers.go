package main

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	bencode "github.com/jackpal/bencode-go"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
)

const (
	IDChoke      = 0
	IDUnchoke    = 1
	IDInterested = 2
	IDHave       = 4
	IDBitfield   = 5
	IDRequest    = 6
	IDPiece      = 7
	IDCancel     = 8
)

const (
	PROTOCOL_LENGTH  = 19
	PROTOCOL_STRING  = "BitTorrent protocol"
	RESERVED_BYTES   = 8
	INFO_HASH_LENGTH = 20
	PEER_ID_LENGTH   = 20
	HANDSHAKE_LENGTH = 68 // 1 + 19 + 8 + 20 + 20
)

type PeerConnection struct {
	InfoHash []byte
	PeerID   string
	Conn     net.Conn
}

type TrackerResponse struct {
	Peers string `bencode:"peers"`
}

type Message struct {
	Payload []byte
	Length  uint32
	ID      byte
}

func generatePeerID() ([]byte, error) {
	id := make([]byte, PEER_ID_LENGTH)
	if _, err := rand.Read(id); err != nil {
		return nil, fmt.Errorf("failed to generate peer ID: %w", err)
	}
	return id, nil
}

func getPeers(announceURL string, infoHash []byte, left int) ([]string, error) {
	urlLink, err := url.Parse(announceURL)
	if err != nil {
		return nil, fmt.Errorf("error parsing URL: %w", err)
	}

	peerID, err := generatePeerID()
	if err != nil {
		return nil, fmt.Errorf("error generating peer ID: %w", err)
	}

	query := urlLink.Query()
	query.Set("info_hash", string(infoHash))
	query.Set("peer_id", string(peerID))
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
	peers := make([]string, 0, len(peersData)/6)
	for i := 0; i < len(peersData); i += 6 {
		ip := net.IPv4(peersData[i], peersData[i+1], peersData[i+2], peersData[i+3])
		port := binary.BigEndian.Uint16([]byte{peersData[i+4], peersData[i+5]})
		peers = append(peers, fmt.Sprintf("%s:%d", ip, port))
	}
	return peers
}

func SendInterested(conn net.Conn) error {
	return sendMessage(conn, IDInterested, nil)
}

func SendRequest(conn net.Conn, index, begin, length int) error {
	buf := make([]byte, 12)
	binary.BigEndian.PutUint32(buf[0:4], uint32(index))
	binary.BigEndian.PutUint32(buf[4:8], uint32(begin))
	binary.BigEndian.PutUint32(buf[8:12], uint32(length))
	return sendMessage(conn, IDRequest, buf)
}

func sendMessage(conn net.Conn, id byte, payload []byte) error {
	length := uint32(1 + len(payload))
	buf := make([]byte, 4+length)
	binary.BigEndian.PutUint32(buf[0:4], length)
	buf[4] = id
	if payload != nil {
		copy(buf[5:], payload)
	}

	_, err := conn.Write(buf)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}
	return nil
}

func ReadMessage(conn net.Conn) (byte, []byte, error) {
	lengthBuf := make([]byte, 4)
	if _, err := io.ReadFull(conn, lengthBuf); err != nil {
		return 0, nil, fmt.Errorf("failed to read message length: %w", err)
	}

	length := binary.BigEndian.Uint32(lengthBuf)
	if length == 0 {
		return 0, nil, nil
	}

	message := make([]byte, length)
	if _, err := io.ReadFull(conn, message); err != nil {
		return 0, nil, fmt.Errorf("failed to read message body: %w", err)
	}

	return message[0], message[1:], nil
}

func printPeers(peers []string) {
	for _, peer := range peers {
		fmt.Println(peer)
	}
}

func NewPeerConnection(peerAddr string, infoHash []byte) (*PeerConnection, error) {
	conn, err := net.Dial("tcp", peerAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to dial peer: %w", err)
	}

	peerID, err := generatePeerID()

	if err != nil {
		return nil, fmt.Errorf("failed to generate peer id: %w", err)
	}

	handshake := make([]byte, HANDSHAKE_LENGTH)
	handshake[0] = PROTOCOL_LENGTH
	copy(handshake[1:PROTOCOL_LENGTH+1], []byte(PROTOCOL_STRING))
	// Reserved bytes are already zeroed
	copy(handshake[PROTOCOL_LENGTH+RESERVED_BYTES+1:], infoHash)
	copy(handshake[PROTOCOL_LENGTH+RESERVED_BYTES+INFO_HASH_LENGTH+1:], peerID)

	if _, err = conn.Write(handshake); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to send handshake: %w", err)
	}

	response := make([]byte, HANDSHAKE_LENGTH)
	if _, err = io.ReadFull(conn, response); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	return &PeerConnection{
		InfoHash: infoHash,
		PeerID:   hex.EncodeToString(response[HANDSHAKE_LENGTH-PEER_ID_LENGTH:]),
		Conn:     conn,
	}, nil
}

func sendInterestedMessage(conn net.Conn) error {
	return sendMessage(conn, IDInterested, nil)
}

func waitForUnchoke(conn net.Conn) error {
	for {
		id, _, err := ReadMessage(conn)
		if err != nil {
			return err
		}
		if id == IDUnchoke {
			return nil
		}
	}
}

func sendRequest(conn net.Conn, index, begin, length uint32) error {
	buf := make([]byte, 12)
	binary.BigEndian.PutUint32(buf[0:4], index)
	binary.BigEndian.PutUint32(buf[4:8], begin)
	binary.BigEndian.PutUint32(buf[8:12], length)
	return sendMessage(conn, IDRequest, buf)
}

func receiveBlock(conn net.Conn) ([]byte, error) {
	id, payload, err := ReadMessage(conn)
	if err != nil {
		return nil, err
	}
	if id != IDPiece {
		return nil, fmt.Errorf("expected piece message, got %d", id)
	}
	return payload[8:], nil // Skip index and begin fields
}

