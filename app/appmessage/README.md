# wire

[![ISC License](http://img.shields.io/badge/license-ISC-blue.svg)](https://choosealicense.com/licenses/isc/)
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg)](http://godoc.org/github.com/karlsen-network/karlsend/wire)

Package wire implements the karlsen wire protocol.

## Karlsen Message Overview

The karlsen protocol consists of exchanging messages between peers.
Each message is preceded by a header which identifies information
about it such as which karlsen network it is a part of, its type, how
big it is, and a checksum to verify validity. All encoding and
decoding of message headers is handled by this package.

To accomplish this, there is a generic interface for karlsen messages
named `Message` which allows messages of any type to be read, written,
or passed around through channels, functions, etc. In addition,
concrete implementations of most all karlsen messages are provided.
All of the details of marshalling and unmarshalling to and from the
wire using karlsen encoding are handled so the  caller doesn't have
to concern themselves with the specifics.

## Reading Messages Example

In order to unmarshal karlsen messages from the wire, use the
`ReadMessage` function. It accepts any `io.Reader`, but typically
this will be a `net.Conn` to a remote node running a karlsen peer.
Example syntax is:

```Go
// Use the most recent protocol version supported by the package and the
// main karlsen network.
pver := wire.ProtocolVersion
karlsennet := wire.Mainnet

// Reads and validates the next karlsen message from conn using the
// protocol version pver and the karlsen network karlsennet. The returns
// are a appmessage.Message, a []byte which contains the unmarshalled
// raw payload, and a possible error.
msg, rawPayload, err := wire.ReadMessage(conn, pver, karlsennet)
if err != nil {
	// Log and handle the error
}
```

See the package documentation for details on determining the message
type.

## Writing Messages Example

In order to marshal karlsen messages to the wire, use the
`WriteMessage` function. It accepts any `io.Writer`, but typically
this will be a `net.Conn` to a remote node running a karlsen peer.
Example syntax to request addresses from a remote peer is:

```Go
// Use the most recent protocol version supported by the package and the
// main bitcoin network.
pver := wire.ProtocolVersion
karlsennet := wire.Mainnet

// Create a new getaddr karlsen message.
msg := wire.NewMsgGetAddr()

// Writes a karlsen message msg to conn using the protocol version
// pver, and the karlsen network karlsennet. The return is a possible
// error.
err := wire.WriteMessage(conn, msg, pver, karlsennet)
if err != nil {
	// Log and handle the error
}
```
