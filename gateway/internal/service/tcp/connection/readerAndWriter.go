package connection

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
)

// message format:
// [4-byte header][body]
// at most 64KB body, header is the length of the body in bytes

const MaxMessageSize = 64 * 1024

func ReadMessage(conn net.Conn) ([]byte, error) {
	var header [4]byte

	// read the 4-byte header
	if _, err := io.ReadFull(conn, header[:]); err != nil {
		return nil, err
	}

	// convert the header to an integer to get the length of the body
	length := binary.BigEndian.Uint32(header[:])

	if length == 0 || length > MaxMessageSize {
		return nil, fmt.Errorf("invalid message length: %d", length)
	}

	// read the body based on the length specified in the header
	body := make([]byte, int(length))
	if _, err := io.ReadFull(conn, body); err != nil {
		return nil, err
	}

	return body, nil
}
