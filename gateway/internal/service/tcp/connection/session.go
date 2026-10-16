package connection

import (
	"net"
	"sync"
)

type Session struct {
	Conn       net.Conn
	UserID     string
	AuthStatus bool

	// Add a mutex to prevent concurrent writes to the connection
	// and ensure thread safety.
	writeMutex sync.Mutex
	closeOnce  sync.Once
}

func NewSession(conn net.Conn) *Session {
	return &Session{
		Conn:       conn,
		AuthStatus: false,
	}
}

func (s *Session) Close() error {
	var err error

	s.closeOnce.Do(func() {
		err = s.Conn.Close()
	})

	return err
}
