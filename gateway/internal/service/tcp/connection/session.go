package connection

import (
	"net"
	"sync"
)

type Session struct {
	ID   string
	Conn net.Conn

	closeOnce sync.Once
}

func NewSession(id string, conn net.Conn) *Session {
	return &Session{
		ID:   id,
		Conn: conn,
	}
}

func (s *Session) Close() error {
	var err error

	s.closeOnce.Do(func() {
		err = s.Conn.Close()
	})

	return err
}
