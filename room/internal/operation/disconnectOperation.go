package operation

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"room/internal/entity"

	"github.com/redis/go-redis/v9"
)

var (
	ErrUserIsInGame  = errors.New("user is in a game")
	ErrUserNotInGame = errors.New("user is no longer in a game")
	ErrStaleSession  = errors.New("disconnect notification is for a stale session")
)

func ensureCurrentSession(ctx context.Context, tx *redis.Tx, userID, sessionID string) error {
	currentSession, err := tx.Get(ctx, "session_account:"+userID).Result()
	if err == redis.Nil {
		return ErrStaleSession
	}
	if err != nil {
		return err
	}
	if currentSession != "session:"+sessionID {
		return ErrStaleSession
	}
	return nil
}

func FindRoomIDByUser(ctx context.Context, client *redis.Client, userID string) (string, error) {
	var cursor uint64
	for {
		keys, nextCursor, err := client.Scan(ctx, cursor, "room:*", 100).Result()
		if err != nil {
			return "", err
		}
		for _, key := range keys {
			roomJSON, err := client.Get(ctx, key).Result()
			if err == redis.Nil {
				continue
			}
			if err != nil {
				return "", err
			}
			var room entity.Room
			if err := json.Unmarshal([]byte(roomJSON), &room); err != nil {
				return "", err
			}
			for _, memberID := range room.UserIDs {
				if memberID == userID {
					if room.RoomID != "" {
						return room.RoomID, nil
					}
					return strings.TrimPrefix(key, "room:"), nil
				}
			}
		}
		cursor = nextCursor
		if cursor == 0 {
			return "", nil
		}
	}
}

// RemoveDisconnectedUser removes a disconnected non-game player and their
// user_status key in the same watched Redis transaction. ErrUserIsInGame tells
// the service that the room transitioned to a game and must retain the user.
func RemoveDisconnectedUser(ctx context.Context, client *redis.Client, roomID, userID, sessionID string) error {
	userStatusKey := "user_status:" + userID
	roomKey := "room:" + roomID
	sessionKey := "session_account:" + userID

	return client.Watch(ctx, func(tx *redis.Tx) error {
		if err := ensureCurrentSession(ctx, tx, userID, sessionID); err != nil {
			return err
		}
		roomJSON, err := tx.Get(ctx, roomKey).Result()
		if err == redis.Nil {
			_, err := tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
				pipe.Del(ctx, userStatusKey)
				return nil
			})
			return err
		}
		if err != nil {
			return err
		}
		var room entity.Room
		if err := json.Unmarshal([]byte(roomJSON), &room); err != nil {
			return err
		}
		if room.Status == entity.RoomStatusInGame {
			return ErrUserIsInGame
		}

		if err := DeleteUser(&room, userID); err != nil {
			return err
		}
		room.DisconnectedUserIDs = removeID(room.DisconnectedUserIDs, userID)
		newRoomJSON, err := json.Marshal(room)
		if err != nil {
			return err
		}
		_, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
			pipe.Del(ctx, userStatusKey)
			if len(room.UserIDs) == 0 {
				pipe.Del(ctx, roomKey)
			} else {
				pipe.Set(ctx, roomKey, newRoomJSON, 0)
			}
			return nil
		})
		return err
	}, userStatusKey, roomKey, sessionKey)
}

// MarkUserDisconnected keeps an in-game player's status persistent and marks
// the connection as disconnected without changing room membership.
func MarkUserDisconnected(ctx context.Context, client *redis.Client, roomID, userID, sessionID string) error {
	userStatusKey := "user_status:" + userID
	roomKey := "room:" + roomID
	sessionKey := "session_account:" + userID

	return client.Watch(ctx, func(tx *redis.Tx) error {
		if err := ensureCurrentSession(ctx, tx, userID, sessionID); err != nil {
			return err
		}
		var userStatus entity.UserStatus
		statusJSON, err := tx.Get(ctx, userStatusKey).Result()
		statusExists := err == nil
		if err == nil {
			if err := json.Unmarshal([]byte(statusJSON), &userStatus); err != nil {
				return err
			}
		} else if err != redis.Nil {
			return err
		}

		roomJSON, roomErr := tx.Get(ctx, roomKey).Result()
		if roomErr != nil && roomErr != redis.Nil {
			return roomErr
		}
		roomIsInGame := false
		var room entity.Room
		if roomErr == nil {
			if err := json.Unmarshal([]byte(roomJSON), &room); err != nil {
				return err
			}
			roomIsInGame = room.Status == entity.RoomStatusInGame && containsID(room.UserIDs, userID)
			if userStatus.RoomID == "" {
				userStatus.RoomID = roomID
			}
		}

		if roomErr == nil {
			if !roomIsInGame {
				return ErrUserNotInGame
			}
		} else if !statusExists || userStatus.Status != entity.UserStatusInGame {
			return ErrUserNotInGame
		}
		userStatus.UserID = userID
		userStatus.Status = entity.UserStatusInGame
		userStatus.RoomID = roomID
		userStatus.ConnectionStatus = entity.UserConnectionStatusDisconnected
		updatedStatus, err := json.Marshal(userStatus)
		if err != nil {
			return err
		}
		_, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
			pipe.Set(ctx, userStatusKey, updatedStatus, 0)
			if roomIsInGame {
				room.DisconnectedUserIDs = appendUniqueID(room.DisconnectedUserIDs, userID)
				updatedRoom, marshalErr := json.Marshal(room)
				if marshalErr != nil {
					return marshalErr
				}
				pipe.Set(ctx, roomKey, updatedRoom, 0)
			}
			return nil
		})
		return err
	}, userStatusKey, roomKey, sessionKey)
}

func containsID(ids []string, id string) bool {
	for _, candidate := range ids {
		if candidate == id {
			return true
		}
	}
	return false
}

func appendUniqueID(ids []string, id string) []string {
	if containsID(ids, id) {
		return ids
	}
	return append(ids, id)
}

func removeID(ids []string, id string) []string {
	filtered := ids[:0]
	for _, candidate := range ids {
		if candidate != id {
			filtered = append(filtered, candidate)
		}
	}
	return filtered
}
