package main

import (
	"encoding/json"
	"os"
	"fmt"
	// bencode "github.com/jackpal/bencode-go" // Available if you need it!
)

// Example:
// - 5:hello -> hello
// - 10:hello12345 -> hello12345


func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.

	command := os.Args[1]
	startIndex := 0

	if command == "decode" {
		bencodedValue := os.Args[2]

		decoded, err := Decode(bencodedValue, &startIndex) 
		if err != nil {
			fmt.Println(err)
			return
		}

		jsonOutput, _ := json.Marshal(decoded)
		fmt.Println(string(jsonOutput))
	} else {
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}
}
