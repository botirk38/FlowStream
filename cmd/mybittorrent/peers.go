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
)

// Message IDs for the BitTorrent protocol
const (
	IDChoke         = uint8(0)
	IDUnchoke       = uint8(1)
	IDInterested    = uint8(2)
	IDNotInterested = uint8(3)
	IDHave          = uint8(4)
	IDBitfield      = uint8(5)
	IDRequest       = uint8(6)
	IDPiece         = uint8(7)
	IDCancel        = uint8(8)
)

// Protocol constants
const (
	DefaultPort = 6881
	BlockSize   = 16384 // Standard BitTorrent block size (16KB)
)

type PeerConnection struct {
	InfoHash []byte
	PeerID   string
	Conn     net.Conn
}

type TrackerResponse struct {
	Interval int    `bencode:"interval"`
	Peers    string `bencode:"peers"`
}

type Message struct {
	Length  uint32
	ID      uint8
	Payload []byte
}

func newMagnetPeerConnection(peerAddr string, infoHash []byte) (*PeerConnection, error) {
	fmt.Printf("Connecting to peer at: %s\n", peerAddr)
	conn, err := net.Dial("tcp", peerAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to dial peer: %w", err)
	}

	fmt.Println("Successfully connected to peer")
	fmt.Println("Generating peer ID...")
	peerID, err := generatePeerID()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to generate peer id: %w", err)
	}

	fmt.Printf("Generated peer ID: %x\n", peerID)
	fmt.Println("Creating and sending handshake...")

	handshake := createMagnetHandshake(infoHash, peerID)

	if err := sendHandshake(conn, handshake); err != nil {
		conn.Close()
		return nil, err
	}

	fmt.Println("Handshake sent successfully")
	fmt.Println("Reading response handshake...")

	responseHandshake, err := readHandshake(conn)
	if err != nil {
		conn.Close()
		return nil, err
	}

	fmt.Printf("Handshake: %x\n", responseHandshake)
	fmt.Printf("Response handshake received, reserved bytes: %x\n", responseHandshake[5])
	fmt.Println("Waiting for bitfield message...")

	_, _, err = ReadMessage(conn)
	if err != nil {
		conn.Close()
		return nil, err
	}

	fmt.Println("Bitfield message received")
	extensionSupport := supportsExtensions(responseHandshake)
	fmt.Printf("Extension support check: %v\n", extensionSupport)

	if extensionSupport {
		fmt.Println("Peer supports extensions, sending extension handshake...")
		message := NewExtensionMessageBuilder().
			WithHandshakePayload().
			Build()

		_, err = conn.Write(message)
		if err != nil {
			conn.Close()
			return nil, fmt.Errorf("failed to send extension handshake: %w", err)
		}

		metadataID, err := handleExtensionHandshake(conn)
		if err != nil {
			conn.Close()
			return nil, fmt.Errorf("failed to handle extension handshake: %w", err)
		}
		fmt.Printf("Peer Metadata Extension ID: %d\n", metadataID)
		fmt.Println("Extension handshake sent successfully")
	}

	fmt.Println("Connection established successfully")
	return &PeerConnection{
		InfoHash: infoHash,
		PeerID:   string(responseHandshake[HandshakeLength-PeerIDLength:]),
		Conn:     conn,
	}, nil
}

func NewPeerConnection(peerAddr string, infoHash []byte) (*PeerConnection, error) {
	conn, err := net.Dial("tcp", peerAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to dial peer: %w", err)
	}

	peerID, err := generatePeerID()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to generate peer id: %w", err)
	}

	handshake := NewHandshakeBuilder().
		WithInfoHash(infoHash).
		WithPeerID(peerID).
		Build()

	if err := sendHandshake(conn, handshake); err != nil {
		conn.Close()
		return nil, err
	}

	responseHandshake, err := readHandshake(conn)
	if err != nil {
		conn.Close()
		return nil, err
	}

	return &PeerConnection{
		InfoHash: infoHash,
		PeerID:   hex.EncodeToString(responseHandshake[HandshakeLength-PeerIDLength:]),
		Conn:     conn,
	}, nil
}

func generatePeerID() ([]byte, error) {
	id := make([]byte, PeerIDLength)
	_, err := rand.Read(id)
	if err != nil {
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

	query := NewTrackerQueryBuilder().
		WithInfoHash(infoHash).
		WithPeerID(peerID).
		WithPort(DefaultPort).
		WithLeft(left).
		Build()

	urlLink.RawQuery = query.Encode()

	resp, err := http.Get(urlLink.String())
	if err != nil {
		return nil, fmt.Errorf("error making HTTP request: %w", err)
	}
	defer resp.Body.Close()

	var trackerRes TrackerResponse
	if err := bencode.Unmarshal(resp.Body, &trackerRes); err != nil {
		return nil, fmt.Errorf("error decoding tracker response: %w", err)
	}

	return parsePeers(trackerRes.Peers), nil
}

func parsePeers(peersData string) []string {
	const peerSize = 6 // 4 bytes for IP + 2 bytes for port
	peers := make([]string, 0, len(peersData)/peerSize)

	for i := 0; i < len(peersData); i += peerSize {
		ip := net.IPv4(peersData[i], peersData[i+1], peersData[i+2], peersData[i+3])
		port := binary.BigEndian.Uint16([]byte{peersData[i+4], peersData[i+5]})
		peers = append(peers, fmt.Sprintf("%s:%d", ip, port))
	}
	return peers
}

func sendMessage(conn net.Conn, id uint8, payload []byte) error {
	message := NewMessageBuilder().
		WithID(id).
		WithPayload(payload).
		Build()

	_, err := conn.Write(message)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}
	return nil
}

func ReadMessage(conn net.Conn) (uint8, []byte, error) {
	lengthBuf := make([]byte, 4)
	if _, err := io.ReadFull(conn, lengthBuf); err != nil {
		return 0, nil, fmt.Errorf("failed to read message length: %w", err)
	}

	length := binary.BigEndian.Uint32(lengthBuf)
	if length == 0 {
		return 0, nil, nil // Keep-alive message
	}

	message := make([]byte, length)
	if _, err := io.ReadFull(conn, message); err != nil {
		return 0, nil, fmt.Errorf("failed to read message body: %w", err)
	}

	return message[0], message[1:], nil
}

func setupConnection(peerConn *PeerConnection) error {
	interestedMsg := NewMessageBuilder().
		WithID(IDInterested).
		Build()

	if _, err := peerConn.Conn.Write(interestedMsg); err != nil {
		return fmt.Errorf("failed to send interested message: %w", err)
	}

	if err := waitForBitfield(peerConn.Conn); err != nil {
		return fmt.Errorf("failed waiting for bitfield: %w", err)
	}

	return waitForUnchoke(peerConn.Conn)
}

func waitForBitfield(conn net.Conn) error {
	id, _, err := ReadMessage(conn)
	if err != nil {
		return err
	}

	if id != IDBitfield {
		return fmt.Errorf("expected bitfield message, got %d", id)
	}
	return nil
}

func waitForUnchoke(conn net.Conn) error {
	for {
		id, _, err := ReadMessage(conn)
		if err != nil {
			return fmt.Errorf("error waiting for unchoke: %w", err)
		}
		if id == IDUnchoke {
			return nil
		}
	}
}

func sendRequest(conn net.Conn, index, begin, length uint32) error {
	requestMsg := NewMessageBuilder().
		WithID(IDRequest).
		WithPayload(NewRequestPayloadBuilder().
			WithIndex(index).
			WithBegin(begin).
			WithLength(length).
			Build()).
		Build()

	_, err := conn.Write(requestMsg)
	return err
}

func receiveBlock(conn net.Conn) ([]byte, error) {
	id, payload, err := ReadMessage(conn)
	if err != nil {
		return nil, fmt.Errorf("failed to receive block: %w", err)
	}

	if id != IDPiece {
		return nil, fmt.Errorf("expected piece message, got %d", id)
	}

	return payload[8:], nil // Skip index and begin fields
}

func printPeers(peers []string) {
	for i, peer := range peers {
		fmt.Printf("Peer %d: %s\n", i+1, peer)
	}
}

