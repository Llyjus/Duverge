package connection

import (
	"net"
	"sync"
	"time"
)

type Timeouts struct {
	Unauthenticated time.Duration
	Idle            time.Duration
	Operation       time.Duration
	Write           time.Duration
}

type Session struct {
	Conn       net.Conn
	UserID     string
	AuthStatus bool

	// Add a mutex to prevent concurrent writes to the connection
	// and ensure thread safety.
	writeMutex sync.Mutex
	closeOnce  sync.Once
	// Add a write timeout to the session
	// to prevent indefinite blocking on writes.
	writeTimeout time.Duration
}

func NewSession(conn net.Conn, writeTimeout time.Duration) *Session {
	return &Session{
		Conn:         conn,
		writeTimeout: writeTimeout,
	}
}

func (s *Session) WriteMessage(body []byte) error {
	s.writeMutex.Lock()
	defer s.writeMutex.Unlock()

	if err := s.Conn.SetWriteDeadline(time.Now().Add(s.writeTimeout)); err != nil {
		_ = s.Close()
		return err
	}
	err := WriteMessage(s.Conn, body)
	if err != nil {
		_ = s.Close()
	}
	return err
}

func (s *Session) Close() error {
	var err error

	s.closeOnce.Do(func() {
		err = s.Conn.Close()
	})

	return err
}
