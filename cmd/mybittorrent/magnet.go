package main

import (
	"fmt"
	"net/url"
	"strings"
)

type MagnetLink struct {
	InfoHash   string
	Name       string
	TrackerURL string
}

func ParseMagnetLink(magnetURL string) (*MagnetLink, error) {
	if !strings.HasPrefix(magnetURL, "magnet:?") {
		return nil, fmt.Errorf("invalid magnet link format")
	}

	// Remove the "magnet:?" prefix
	queryString := strings.TrimPrefix(magnetURL, "magnet:?")

	// Parse the query parameters
	values, err := url.ParseQuery(queryString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse magnet link: %v", err)
	}

	magnet := &MagnetLink{}

	// Extract info hash
	xt := values.Get("xt")
	if xt == "" {
		return nil, fmt.Errorf("missing info hash (xt parameter)")
	}
	if !strings.HasPrefix(xt, "urn:btih:") {
		return nil, fmt.Errorf("invalid info hash format")
	}
	magnet.InfoHash = strings.TrimPrefix(xt, "urn:btih:")

	// Extract name (optional)
	magnet.Name = values.Get("dn")

	// Extract tracker URL
	magnet.TrackerURL = values.Get("tr")

	return magnet, nil
}

