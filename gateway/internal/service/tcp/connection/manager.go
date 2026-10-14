package connection

import (
	"net"
	"sync"
)

type ConnectionManager struct {
	mu    sync.RWMutex
	conns map[string]net.Conn
}

func NewConnectionManager() *ConnectionManager {
	return &ConnectionManager{
		conns: make(map[string]net.Conn),
	}
}

func (cm *ConnectionManager) Add(id string, conn net.Conn) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.conns[id] = conn
}

func (cm *ConnectionManager) Remove(id string, expected net.Conn) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Check if the connection exists and matches the expected connection
	// If it doesn't match( maybe new connection ), we don't remove it
	conn, ok := cm.conns[id]
	if !ok || conn != expected {
		return
	}

	delete(cm.conns, id)
	conn.Close()
}
func (cm *ConnectionManager) Get(id string) (net.Conn, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	conn, ok := cm.conns[id]
	return conn, ok
}
