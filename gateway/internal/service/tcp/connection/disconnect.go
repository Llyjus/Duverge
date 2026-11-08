package connection

import (
	"context"
	"log"
	"time"
)

// RoomDisconnectNotifier is the boundary for notifying Room that an
// authenticated TCP session has disconnected. The current Room gRPC API does
// not yet expose a matching user_id/session_id RPC.
type RoomDisconnectNotifier interface {
	NotifyUserDisconnected(ctx context.Context, userID string, sessionID string) error
}

type DisconnectHandler struct {
	manager       *ConnectionManager
	roomNotifier  RoomDisconnectNotifier
	notifyTimeout time.Duration
}

func NewDisconnectHandler(
	manager *ConnectionManager,
	roomNotifier RoomDisconnectNotifier,
	notifyTimeout time.Duration,
) *DisconnectHandler {
	return &DisconnectHandler{
		manager:       manager,
		roomNotifier:  roomNotifier,
		notifyTimeout: notifyTimeout,
	}
}

// Handle runs once per session. It always clears the local connection before
// attempting the best-effort Room notification.
func (h *DisconnectHandler) Handle(session *Session) {
	session.disconnectOnce.Do(func() {
		wasAuthenticated := session.AuthStatus
		if session.AuthStatus && session.UserID != "" {
			h.manager.Remove(session.UserID, session)
		}
		_ = session.Close()

		if !wasAuthenticated {
			return
		}
		if h.roomNotifier == nil {
			log.Printf("[TCP] Room disconnect notifier is not configured for user %s", session.UserID)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), h.notifyTimeout)
		defer cancel()
		if err := h.roomNotifier.NotifyUserDisconnected(ctx, session.UserID, session.SessionID); err != nil {
			log.Printf("[TCP] failed to notify Room about disconnected user %s: %v", session.UserID, err)
		}
	})
}
