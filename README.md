# socket.io-parser

![workflow](https://github.com/funcards/socket.io-parser/actions/workflows/workflow.yml/badge.svg)
[![Coverage Status](https://coveralls.io/repos/funcards/socket.io-parser/badge.svg?branch=main&service=github)](https://coveralls.io/github/funcards/socket.io-parser?branch=main)
[![GoDoc](https://godoc.org/github.com/funcards/socket.io-parser?status.svg)](https://pkg.go.dev/github.com/funcards/socket.io-parser/v5)
![License](https://img.shields.io/dub/l/vibe-d.svg)

socket.io encoder and decoder written in GO complying with version `5` of [socket.io-protocol](https://github.com/socketio/socket.io-protocol).

## TODO: reconstruct binary on decode

## Installation

Use go get.

```bash
go get github.com/funcards/socket.io-parser/v5
```

Then import the parser package into your own code.

```go
import "github.com/funcards/socket.io-parser/v5"
```

## How to use

The parser can encode/decode packets, payloads and payloads as binary.

Example:

```go
packet := sio_parser.Packet{
    Type: sio_parser.Connect,
    Nsp:  "/posts",
    Data: map[string]string{
        "sid": "unique-id",
    },
}
encoded := packet.Encode()
fmt.Println(encoded[0].(string))

fn := func(pkt Packet) {
    fmt.Println(pkt)
}

err := sio_parser.Decode(fn, encoded...)
```

## License

Distributed under MIT License, please see license file within the code for more details.
