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

func WriteMessage(conn net.Conn, body []byte) error {
	length := len(body)

	if length == 0 || length > MaxMessageSize {
		return fmt.Errorf("invalid message length: %d", length)
	}

	// create the 4-byte length header
	var header [4]byte
	binary.BigEndian.PutUint32(header[:], uint32(length))

	// combine the header and message body
	frame := make([]byte, 4+length)
	copy(frame[:4], header[:])
	copy(frame[4:], body)

	// write the complete message
	for len(frame) > 0 {
		n, err := conn.Write(frame)
		if err != nil {
			return err
		}
		if n == 0 {
			return io.ErrShortWrite
		}
		frame = frame[n:]
	}

	return nil
}
