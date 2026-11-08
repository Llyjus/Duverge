package test

import (
	"context"
	"errors"
	"io"
	"net"
	"testing"
	"time"

	"gateway/internal/service/tcp/connection"
	"gateway/internal/service/tcp/handler"
)

type disconnectNotifierStub struct {
	calls     int
	userID    string
	sessionID string
	err       error
}

func (n *disconnectNotifierStub) NotifyUserDisconnected(_ context.Context, userID, sessionID string) error {
	n.calls++
	n.userID = userID
	n.sessionID = sessionID
	return n.err
}

func TestDisconnectRemovesAuthenticatedSessionAndNotifiesRoomOnce(t *testing.T) {
	manager := connection.NewConnectionManager()
	serverConn, peerConn := net.Pipe()
	t.Cleanup(func() { _ = peerConn.Close() })
	session := connection.NewSession(serverConn, time.Second)
	session.UserID = "user-123"
	session.SessionID = "session-456"
	session.AuthStatus = true
	manager.Add(session)
	notifier := &disconnectNotifierStub{}
	disconnect := connection.NewDisconnectHandler(manager, notifier, time.Second)

	disconnect.Handle(session)
	disconnect.Handle(session)

	if _, ok := manager.Get(session.UserID); ok {
		t.Fatal("authenticated session remains in ConnectionManager")
	}
	if notifier.calls != 1 {
		t.Fatalf("Room notification calls = %d, want 1", notifier.calls)
	}
	if notifier.userID != "user-123" || notifier.sessionID != "session-456" {
		t.Fatalf("Room notification IDs = (%q, %q), want (user-123, session-456)", notifier.userID, notifier.sessionID)
	}
	if _, err := peerConn.Read(make([]byte, 1)); !errors.Is(err, io.EOF) {
		t.Fatalf("peer read error = %v, want EOF after socket close", err)
	}
}

func TestDisconnectRemovesUnverifiedRegisteredSessionWithoutRoomNotification(t *testing.T) {
	manager := connection.NewConnectionManager()
	serverConn, peerConn := net.Pipe()
	t.Cleanup(func() { _ = peerConn.Close() })
	session := connection.NewSession(serverConn, time.Second)
	// This models authentication setup that registered a user before it failed.
	session.UserID = "user-123"
	session.SessionID = "session-456"
	manager.Add(session)
	notifier := &disconnectNotifierStub{}
	disconnect := connection.NewDisconnectHandler(manager, notifier, time.Second)

	disconnect.Handle(session)

	if _, ok := manager.Get(session.UserID); ok {
		t.Fatal("partially registered session remains in ConnectionManager")
	}
	if notifier.calls != 0 {
		t.Fatalf("Room notification calls = %d, want 0 for unauthenticated session", notifier.calls)
	}
}

func TestDisconnectNotificationFailureDoesNotPreventLocalCleanup(t *testing.T) {
	manager := connection.NewConnectionManager()
	serverConn, peerConn := net.Pipe()
	t.Cleanup(func() { _ = peerConn.Close() })
	session := connection.NewSession(serverConn, time.Second)
	session.UserID = "user-123"
	session.SessionID = "session-456"
	session.AuthStatus = true
	manager.Add(session)
	notifier := &disconnectNotifierStub{err: errors.New("Room unavailable")}
	disconnect := connection.NewDisconnectHandler(manager, notifier, time.Second)

	disconnect.Handle(session)

	if _, ok := manager.Get(session.UserID); ok {
		t.Fatal("session remains in ConnectionManager after Room notification failure")
	}
	if notifier.calls != 1 {
		t.Fatalf("Room notification calls = %d, want 1", notifier.calls)
	}
	if _, err := peerConn.Read(make([]byte, 1)); !errors.Is(err, io.EOF) {
		t.Fatalf("peer read error = %v, want EOF after socket close", err)
	}
}

func TestReadDeadlineTimeoutRunsDisconnectCleanup(t *testing.T) {
	manager := connection.NewConnectionManager()
	serverConn, peerConn := net.Pipe()
	t.Cleanup(func() { _ = peerConn.Close() })
	session := connection.NewSession(serverConn, time.Second)
	session.UserID = "user-123"
	session.SessionID = "session-456"
	session.AuthStatus = true
	manager.Add(session)
	notifier := &disconnectNotifierStub{}
	disconnect := connection.NewDisconnectHandler(manager, notifier, time.Second)
	done := make(chan struct{})

	go func() {
		defer close(done)
		handler.HandleConnection(
			session,
			manager,
			nil,
			connection.Timeouts{Idle: 30 * time.Millisecond, Operation: time.Second, Write: time.Second},
			disconnect,
		)
	}()

	if _, err := connection.ReadMessage(peerConn); err != nil {
		t.Fatalf("read timeout response: %v", err)
	}
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("TCP handler did not return after read timeout")
	}
	if _, ok := manager.Get(session.UserID); ok {
		t.Fatal("timed out session remains in ConnectionManager")
	}
	if notifier.calls != 1 {
		t.Fatalf("Room notification calls = %d, want 1 after authenticated read timeout", notifier.calls)
	}
}
