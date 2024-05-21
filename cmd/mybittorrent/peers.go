package main

import (
	"fmt"
	bencode "github.com/jackpal/bencode-go"
	"net/http"
	"net/url"
	"strconv"
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
