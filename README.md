
# FlowStream 

FlowStream is an advanced, high-performance BitTorrent client engineered to optimize peer-to-peer file sharing. Utilizing cutting-edge algorithms and comprehensive network communication protocols, FlowStream is crafted to efficiently handle torrent decoding, peer discovery, and secure file transfers. This robust client is developed in Go, leveraging modular design principles for enhanced scalability and maintainability.

## Key Features

- **Advanced Torrent File Parsing**: Employs bencode for precision decoding of torrent files, extracting critical metadata including tracker details and file hashes.
- **Sophisticated Tracker Interaction**: Implements an intelligent algorithm to interact with torrent trackers, ensuring optimal peer discovery and network efficiency.
- **Robust Peer Communication Protocol**: Manages complex peer handshake mechanisms and supports dynamic file segmentation to optimize download speeds and maintain connection stability.
- **Enhanced File Download Mechanism**: Incorporates adaptive file download strategies that prioritize pieces based on availability and network conditions, aimed at reducing overall download time and improving data integrity.

## System Requirements

- Go programming language (version 1.15 or later)
- Dependencies:
  - `github.com/jackpal/bencode-go`: Required for the encoding and decoding of Bencode data.

## Installation Instructions

1. **Clone the Project Repository:**
   Clone FlowStream from its repository to your local machine:
   ```bash
   git clone https://github.com/botirk38/FlowStream
   cd cmd/bittorrent 
   ```

2. **Compile the Source Code:**
   Build the executable from the source code:
   ```bash
   go build -o mybittorrent
   ```

## Usage Overview

FlowStream supports various operational modes, tailored to different aspects of the BitTorrent protocol:

### Decode Torrent Metadata

Decode and display metadata from a bencoded string:

```bash
./mybittorrent decode <bencoded-string>
```

### Display Torrent Information

Extract and present detailed information from a specified torrent file:

```bash
./mybittorrent info <path-to-torrent-file>
```

### Retrieve Peer Information

Connect to a tracker and fetch peer details for initiating downloads:

```bash
./mybittorrent peers <path-to-torrent-file>
```

## Contributing

We encourage contributions from the community! If you are interested in enhancing FlowStream's capabilities or refining existing features, please fork the repository and submit your pull requests for review.

## Licensing

FlowStream is made available under the MIT License. For more details, please refer to the [LICENSE](LICENSE) document included in the repository.

