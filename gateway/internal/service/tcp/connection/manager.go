package connection

import (
	"sync"
)

type ConnectionManager struct {
	mu       sync.RWMutex
	sessions map[string]*Session
}

func NewConnectionManager() *ConnectionManager {
	return &ConnectionManager{
		sessions: make(map[string]*Session),
	}
}

func (cm *ConnectionManager) Add(session *Session) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.sessions[session.ID] = session
}

func (cm *ConnectionManager) Remove(id string, expected *Session) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Check if the connection exists and matches the expected connection
	// If it doesn't match( maybe new connection ), we don't remove it
	session, ok := cm.sessions[id]
	if !ok || session != expected {
		return
	}

	delete(cm.sessions, id)
	session.Close()
}
func (cm *ConnectionManager) Get(id string) (*Session, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	session, ok := cm.sessions[id]
	return session, ok
}
