package main

import (
	"fmt"
	"strconv"
	"unicode"
)

// Decode dispatches the bencoded string to the appropriate decoder based on its first character.
// An index pointer is added to track parsing progress.
func Decode(bencodedString string, index *int) (interface{}, error) {
	if *index >= len(bencodedString) {
		return nil, fmt.Errorf("out of bounds")
	}

	switch bencodedString[*index] {
	case 'l':
		*index += 1 // Move past 'l'
		if bencodedString[len(bencodedString)-1] != 'e' {
			return nil, fmt.Errorf("invalid list format")
		}
		return decodeList(bencodedString, index)
	case 'i':
		return decodeInteger(bencodedString, index)
	case 'd':
		return decodeDict(bencodedString, index)
	default:
		if unicode.IsDigit(rune(bencodedString[*index])) {
			return decodeString(bencodedString, index)
		}
		return nil, fmt.Errorf("unsupported format or type")
	}
}

// decodeString handles decoding of bencoded strings, now using an index.
func decodeString(bencodedString string, index *int) (string, error) {
	start := *index
	firstColonIndex := findFirstColon(bencodedString[start:])
	if firstColonIndex == -1 {
		return "", fmt.Errorf("colon not found")
	}
	firstColonIndex += start // Adjust for current index
	lengthStr := bencodedString[start:firstColonIndex]
	length, err := strconv.Atoi(lengthStr)
	if err != nil {
		return "", fmt.Errorf("invalid length: %w", err)
	}
	contentStart := firstColonIndex + 1
	contentEnd := contentStart + length
	if contentEnd > len(bencodedString) {
		return "", fmt.Errorf("specified length %d exceeds string length %d", length, len(bencodedString))
	}
	*index = contentEnd // Move index past the current string
	return bencodedString[contentStart:contentEnd], nil
}

// decodeInteger handles decoding of bencoded integers, now using an index.
func decodeInteger(bencodedString string, index *int) (int, error) {
	if bencodedString[*index] != 'i' || bencodedString[len(bencodedString)-1] != 'e' {
		return 0, fmt.Errorf("invalid integer format")
	}
	start := *index + 1
	end := start

	for end < len(bencodedString) && bencodedString[end] != 'e' {
		end++
	}

	*index = end + 1 // Move past 'e'
	return strconv.Atoi(bencodedString[start:end])
}

// decodeList handles decoding of bencoded lists, now using an index.
func decodeList(bencodedString string, index *int) ([]interface{}, error) {
	list := make([]interface{}, 0)

	for *index < len(bencodedString) && bencodedString[*index] != 'e' {
		element, err := Decode(bencodedString, index)
		if err != nil {
			return nil, err
		}
		list = append(list, element)
	}
	*index += 1 // Move past 'e'
	return list, nil
}

func decodeDict(bencodedString string, index *int) (map[string]interface{}, error) {

	*index += 1 // Move past 'd'
	dict := make(map[string]interface{})

	for *index < len(bencodedString) && bencodedString[*index] != 'e' {
		key, err := decodeString(bencodedString, index)
		if err != nil {
			return nil, err
		}
		value, err := Decode(bencodedString, index)
		if err != nil {
			return nil, err
		}
		dict[key] = value
	}

	*index += 1 // Move past 'e'
	return dict, nil
}

// findFirstColon finds the index of the first colon in the substring of the string from the start.
func findFirstColon(s string) int {
	for i, c := range s {
		if c == ':' {
			return i
		}
	}
	return -1
}
