package main

import (
	"encoding/json"
	"fmt"
	"os"
	// bencode "github.com/jackpal/bencode-go" // Available if you need it!
)

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

		torrentInfo, err := os.ReadFile(torrentFile)

		if err != nil {
			fmt.Println(err)
			return
		}

		decoded, err := Decode(string(torrentInfo), &startIndex)

		if err != nil {
			fmt.Println(err)
			return

		}

		torrentDict, ok := decoded.(map[string]interface{})

		if !ok {
			fmt.Println("Error: Decoded value is not a dictionary")
			return
		}

		announce := torrentDict["announce"].(string)
		infoDict := torrentDict["info"].(map[string]interface{})
		fileLength := infoDict["length"].(int)

		fmt.Println("Tracker URL:", announce)
		fmt.Println("File Length:", fileLength)

	default:
		{
			fmt.Println("Invalid command")
			os.Exit(1)
		}

	}

}
