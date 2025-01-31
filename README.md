# FLOWSTREAM

![FlowStream Logo](https://cdn-icons-png.flaticon.com/512/6295/6295417.png)

## High-Performance BitTorrent Client

![License](https://img.shields.io/github/license/botirk38/FlowStream?style=default&logo=opensourceinitiative&logoColor=white&color=0074ff)
![Last Commit](https://img.shields.io/github/last-commit/botirk38/FlowStream?style=default&logo=git&logoColor=white&color=0074ff)
![Top Language](https://img.shields.io/github/languages/top/botirk38/FlowStream?style=default&color=0074ff)
![Language Count](https://img.shields.io/github/languages/count/botirk38/FlowStream?style=default&color=0074ff)

## Table of Contents
- [Overview](#overview)
- [Features](#features)
- [Project Structure](#project-structure)
  - [Project Index](#project-index)
- [Getting Started](#getting-started)
  - [Prerequisites](#prerequisites)
  - [Installation](#installation)
  - [Usage](#usage)
  - [Testing](#testing)
- [Project Roadmap](#project-roadmap)
- [Contributing](#contributing)
- [License](#license)
- [Acknowledgments](#acknowledgments)

---

## Overview

FlowStream is an advanced, high-performance BitTorrent client engineered to optimize peer-to-peer file sharing. Utilizing cutting-edge algorithms and comprehensive network communication protocols, FlowStream is crafted to efficiently handle torrent decoding, peer discovery, and secure file transfers. This robust client is developed in Go, leveraging modular design principles for enhanced scalability and maintainability.

---

## Features

- **Advanced Torrent File Parsing**: Employs bencode for precision decoding of torrent files, extracting critical metadata including tracker details and file hashes.
- **Sophisticated Tracker Interaction**: Implements efficient communication protocols to announce client presence and retrieve peer lists from trackers.
- **Peer Management**: Manages connections with multiple peers, facilitating efficient data exchange and maintaining optimal download and upload rates.
- **Robust Data Integrity**: Utilizes SHA-1 hashing to verify the integrity of downloaded pieces, ensuring the accuracy and completeness of files.
- **Efficient Resource Utilization**: Designed with a focus on minimal resource consumption, ensuring high performance even on systems with limited capabilities.

---

## Project Structure

```sh
FlowStream/
├── README.md
├── cmd/
│   └── mybittorrent/
│       ├── decoder.go
│       ├── download.go
│       ├── main.go
│       ├── network.go
│       ├── peers.go
│       └── torrent-parser.go
├── codecrafters.yml
├── go.mod
├── go.sum
├── sample.torrent
└── your_bittorrent.sh
```

### Project Index

<details>
  <summary><b>cmd/mybittorrent/</b></summary>
  
  - **[decoder.go](cmd/mybittorrent/decoder.go)** - Handles the decoding of bencoded data from torrent files.
  - **[download.go](cmd/mybittorrent/download.go)** - Manages the downloading of file pieces from connected peers.
  - **[main.go](cmd/mybittorrent/main.go)** - The entry point of the application, initializing the client.
  - **[network.go](cmd/mybittorrent/network.go)** - Handles network communications, including peer connections.
  - **[peers.go](cmd/mybittorrent/peers.go)** - Manages peer information and interactions.
  - **[torrent-parser.go](cmd/mybittorrent/torrent-parser.go)** - Parses torrent files to extract metadata.
</details>

---

## Getting Started

### Prerequisites
- Install [Go](https://go.dev/dl/) (version 1.20+ recommended).

### Installation
```sh
git clone https://github.com/botirk38/FlowStream.git
cd FlowStream
go mod tidy
```

### Usage
```sh
go run cmd/mybittorrent/main.go sample.torrent
```

### Testing
```sh
go test ./...
```

---

## Project Roadmap
- [ ] Implement DHT support
- [ ] Optimize peer selection algorithm
- [ ] Add GUI frontend for user interaction

---

## Contributing
Pull requests are welcome! Please follow the coding guidelines and open an issue for major changes.

---

## License
This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

---

## Acknowledgments
- Inspired by open-source torrent clients like qBittorrent and Transmission.
