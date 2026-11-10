package service

import (
	"context"
	"encoding/json"
	"strconv"
	"testing"
	"time"

	"room/internal/entity"
	"room/internal/service"

	"github.com/redis/go-redis/v9"
)

func TestHandleDisconnectOutsideRoomIsIdempotent(t *testing.T) {
	client := resilienceRedis(t)
	userID := "disconnect-lobby-" + strconv.FormatInt(time.Now().UnixNano(), 10)
	key := "user_status:" + userID
	t.Cleanup(func() { _ = client.Del(context.Background(), key).Err() })
	sessionID := setActiveSession(t, client, userID)
	statusJSON, _ := json.Marshal(entity.UserStatus{UserID: userID, Status: entity.UserStatusInLobby})
	if err := client.Set(context.Background(), key, statusJSON, 0).Err(); err != nil {
		t.Fatal(err)
	}

	for range 2 {
		if err := service.HandleDisconnect(context.Background(), client, userID, sessionID); err != nil {
			t.Fatalf("repeated lobby disconnect returned error: %v", err)
		}
	}
}

func TestHandleDisconnectIgnoresStaleSession(t *testing.T) {
	client := resilienceRedis(t)
	userID := "disconnect-stale-session-" + strconv.FormatInt(time.Now().UnixNano(), 10)
	statusKey := "user_status:" + userID
	sessionKey := "session_account:" + userID
	roomID := "disconnect-stale-room-" + strconv.FormatInt(time.Now().UnixNano(), 10)
	roomKey := "room:" + roomID
	t.Cleanup(func() { _ = client.Del(context.Background(), statusKey, sessionKey, roomKey).Err() })

	oldSessionID := "old-session"
	newSessionID := "new-session"
	if err := client.Set(context.Background(), sessionKey, "session:"+newSessionID, 7*24*time.Hour).Err(); err != nil {
		t.Fatal(err)
	}
	roomJSON, _ := json.Marshal(entity.Room{
		RoomID: roomID, MaximumSize: 4, UserIDs: []string{userID},
		HosterID: userID, Prepared: []string{}, Status: entity.RoomStatusInGame,
	})
	statusJSON, _ := json.Marshal(entity.UserStatus{
		UserID: userID, Status: entity.UserStatusInGame, RoomID: roomID,
	})
	if err := client.Set(context.Background(), roomKey, roomJSON, 0).Err(); err != nil {
		t.Fatal(err)
	}
	if err := client.Set(context.Background(), statusKey, statusJSON, 0).Err(); err != nil {
		t.Fatal(err)
	}

	if err := service.HandleDisconnect(context.Background(), client, userID, oldSessionID); err != nil {
		t.Fatalf("stale session disconnect returned error: %v", err)
	}
	storedStatus, err := client.Get(context.Background(), statusKey).Result()
	if err != nil {
		t.Fatal(err)
	}
	var gotStatus entity.UserStatus
	if err := json.Unmarshal([]byte(storedStatus), &gotStatus); err != nil {
		t.Fatal(err)
	}
	if gotStatus.ConnectionStatus == entity.UserConnectionStatusDisconnected {
		t.Fatalf("new session status was marked disconnected: %+v", gotStatus)
	}
	storedRoom, err := client.Get(context.Background(), roomKey).Result()
	if err != nil {
		t.Fatal(err)
	}
	var gotRoom entity.Room
	if err := json.Unmarshal([]byte(storedRoom), &gotRoom); err != nil {
		t.Fatal(err)
	}
	if len(gotRoom.UserIDs) != 1 || gotRoom.UserIDs[0] != userID || len(gotRoom.DisconnectedUserIDs) != 0 {
		t.Fatalf("stale session changed room membership/state: %+v", gotRoom)
	}
}

func TestHandleDisconnectRemovesWaitingRoomMemberAndStatus(t *testing.T) {
	client := resilienceRedis(t)
	suffix := strconv.FormatInt(time.Now().UnixNano(), 10)
	roomID := "disconnect-room-" + suffix
	hostID := "disconnect-host-" + suffix
	guestID := "disconnect-guest-" + suffix
	roomKey := "room:" + roomID
	hostStatusKey := "user_status:" + hostID
	guestStatusKey := "user_status:" + guestID
	t.Cleanup(func() { _ = client.Del(context.Background(), roomKey, hostStatusKey, guestStatusKey).Err() })
	sessionID := setActiveSession(t, client, guestID)

	roomJSON, _ := json.Marshal(entity.Room{
		RoomID: roomID, MaximumSize: 4, UserIDs: []string{hostID, guestID},
		HosterID: hostID, Prepared: []string{}, Status: entity.RoomStatusWaiting,
	})
	if err := client.Set(context.Background(), roomKey, roomJSON, 0).Err(); err != nil {
		t.Fatal(err)
	}
	for _, userID := range []string{hostID, guestID} {
		statusJSON, _ := json.Marshal(entity.UserStatus{UserID: userID, Status: entity.UserStatusInRoom, RoomID: roomID})
		if err := client.Set(context.Background(), "user_status:"+userID, statusJSON, 0).Err(); err != nil {
			t.Fatal(err)
		}
	}

	if err := service.HandleDisconnect(context.Background(), client, guestID, sessionID); err != nil {
		t.Fatalf("HandleDisconnect() error = %v", err)
	}
	if err := service.HandleDisconnect(context.Background(), client, guestID, sessionID); err != nil {
		t.Fatalf("repeated HandleDisconnect() error = %v", err)
	}
	if _, err := client.Get(context.Background(), guestStatusKey).Result(); err != redis.Nil {
		t.Fatalf("disconnected user's status error = %v, want redis.Nil", err)
	}
	storedRoom, err := client.Get(context.Background(), roomKey).Result()
	if err != nil {
		t.Fatal(err)
	}
	var updated entity.Room
	if err := json.Unmarshal([]byte(storedRoom), &updated); err != nil {
		t.Fatal(err)
	}
	if len(updated.UserIDs) != 1 || updated.UserIDs[0] != hostID {
		t.Fatalf("room members = %v, want only %q", updated.UserIDs, hostID)
	}
}

func TestHandleDisconnectDeletesEmptyWaitingRoom(t *testing.T) {
	client := resilienceRedis(t)
	suffix := strconv.FormatInt(time.Now().UnixNano(), 10)
	roomID := "disconnect-empty-" + suffix
	userID := "disconnect-only-" + suffix
	roomKey := "room:" + roomID
	statusKey := "user_status:" + userID
	t.Cleanup(func() { _ = client.Del(context.Background(), roomKey, statusKey).Err() })
	sessionID := setActiveSession(t, client, userID)
	roomJSON, _ := json.Marshal(entity.Room{
		RoomID: roomID, MaximumSize: 4, UserIDs: []string{userID},
		HosterID: userID, Prepared: []string{}, Status: entity.RoomStatusWaiting,
	})
	statusJSON, _ := json.Marshal(entity.UserStatus{UserID: userID, Status: entity.UserStatusInRoom, RoomID: roomID})
	if err := client.Set(context.Background(), roomKey, roomJSON, 0).Err(); err != nil {
		t.Fatal(err)
	}
	if err := client.Set(context.Background(), statusKey, statusJSON, 0).Err(); err != nil {
		t.Fatal(err)
	}

	if err := service.HandleDisconnect(context.Background(), client, userID, sessionID); err != nil {
		t.Fatalf("HandleDisconnect() error = %v", err)
	}
	if _, err := client.Get(context.Background(), roomKey).Result(); err != redis.Nil {
		t.Fatalf("empty room lookup error = %v, want redis.Nil", err)
	}
	if _, err := client.Get(context.Background(), statusKey).Result(); err != redis.Nil {
		t.Fatalf("user status lookup error = %v, want redis.Nil", err)
	}
}

func TestHandleDisconnectPreservesInGameMemberAndDisconnectedState(t *testing.T) {
	client := resilienceRedis(t)
	suffix := strconv.FormatInt(time.Now().UnixNano(), 10)
	roomID := "disconnect-game-" + suffix
	userID := "disconnect-player-" + suffix
	otherID := "disconnect-other-" + suffix
	roomKey := "room:" + roomID
	statusKey := "user_status:" + userID
	otherStatusKey := "user_status:" + otherID
	t.Cleanup(func() { _ = client.Del(context.Background(), roomKey, statusKey, otherStatusKey).Err() })
	sessionID := setActiveSession(t, client, userID)
	roomJSON, _ := json.Marshal(entity.Room{
		RoomID: roomID, MaximumSize: 4, UserIDs: []string{userID, otherID},
		HosterID: userID, Prepared: []string{userID, otherID}, Status: entity.RoomStatusInGame,
	})
	userStatusJSON, _ := json.Marshal(entity.UserStatus{UserID: userID, Status: entity.UserStatusInGame, RoomID: roomID})
	otherStatusJSON, _ := json.Marshal(entity.UserStatus{UserID: otherID, Status: entity.UserStatusInGame, RoomID: roomID})
	if err := client.Set(context.Background(), roomKey, roomJSON, 0).Err(); err != nil {
		t.Fatal(err)
	}
	// Model an old/ordinary status TTL to prove the in-game disconnect makes it persistent.
	if err := client.Set(context.Background(), statusKey, userStatusJSON, 30*time.Second).Err(); err != nil {
		t.Fatal(err)
	}
	if err := client.Set(context.Background(), otherStatusKey, otherStatusJSON, 0).Err(); err != nil {
		t.Fatal(err)
	}

	if err := service.HandleDisconnect(context.Background(), client, userID, sessionID); err != nil {
		t.Fatalf("HandleDisconnect() error = %v", err)
	}
	storedStatus, err := client.Get(context.Background(), statusKey).Result()
	if err != nil {
		t.Fatal(err)
	}
	var updatedStatus entity.UserStatus
	if err := json.Unmarshal([]byte(storedStatus), &updatedStatus); err != nil {
		t.Fatal(err)
	}
	if updatedStatus.Status != entity.UserStatusInGame || updatedStatus.ConnectionStatus != entity.UserConnectionStatusDisconnected {
		t.Fatalf("user status = %+v, want in_game and connection_status=disconnected", updatedStatus)
	}
	if ttl, err := client.TTL(context.Background(), statusKey).Result(); err != nil || ttl != -1 {
		t.Fatalf("in-game status TTL = %s, error=%v; want persistent (-1)", ttl, err)
	}

	storedRoom, err := client.Get(context.Background(), roomKey).Result()
	if err != nil {
		t.Fatal(err)
	}
	var updatedRoom entity.Room
	if err := json.Unmarshal([]byte(storedRoom), &updatedRoom); err != nil {
		t.Fatal(err)
	}
	if len(updatedRoom.UserIDs) != 2 || !contains(updatedRoom.UserIDs, userID) || !contains(updatedRoom.DisconnectedUserIDs, userID) {
		t.Fatalf("room after disconnect = %+v, want member retained and marked disconnected", updatedRoom)
	}
	firstDisconnectStatus := storedStatus
	firstDisconnectRoom := storedRoom
	if err := service.HandleDisconnect(context.Background(), client, userID, sessionID); err != nil {
		t.Fatalf("repeated in-game HandleDisconnect() error = %v", err)
	}
	storedStatus, err = client.Get(context.Background(), statusKey).Result()
	if err != nil {
		t.Fatal(err)
	}
	storedRoom, err = client.Get(context.Background(), roomKey).Result()
	if err != nil {
		t.Fatal(err)
	}
	if storedStatus != firstDisconnectStatus || storedRoom != firstDisconnectRoom {
		t.Fatalf("repeated disconnect changed persisted data: status equal=%v room equal=%v", storedStatus == firstDisconnectStatus, storedRoom == firstDisconnectRoom)
	}

	if err := service.EndRoomGame(client, roomID); err != nil {
		t.Fatalf("EndRoomGame() error = %v", err)
	}
	firstEndRoom, err := client.Get(context.Background(), roomKey).Result()
	if err != nil {
		t.Fatal(err)
	}
	firstOnlineStatus, err := client.Get(context.Background(), otherStatusKey).Result()
	if err != nil {
		t.Fatal(err)
	}
	if err := service.EndRoomGame(client, roomID); err != nil {
		t.Fatalf("repeated EndRoomGame() error = %v", err)
	}
	secondEndRoom, err := client.Get(context.Background(), roomKey).Result()
	if err != nil {
		t.Fatal(err)
	}
	secondOnlineStatus, err := client.Get(context.Background(), otherStatusKey).Result()
	if err != nil {
		t.Fatal(err)
	}
	if firstEndRoom != secondEndRoom || firstOnlineStatus != secondOnlineStatus {
		t.Fatalf("repeated game cleanup changed state: room equal=%v online status equal=%v", firstEndRoom == secondEndRoom, firstOnlineStatus == secondOnlineStatus)
	}
	if _, err := client.Get(context.Background(), statusKey).Result(); err != redis.Nil {
		t.Fatalf("disconnected user's status after game end error = %v, want redis.Nil", err)
	}
	storedStatus, err = client.Get(context.Background(), otherStatusKey).Result()
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal([]byte(storedStatus), &updatedStatus); err != nil {
		t.Fatal(err)
	}
	if updatedStatus.Status != entity.UserStatusInRoom || updatedStatus.ConnectionStatus == entity.UserConnectionStatusDisconnected {
		t.Fatalf("connected user's status after game end = %+v, want in_room and connected", updatedStatus)
	}
	storedRoom, err = client.Get(context.Background(), roomKey).Result()
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal([]byte(storedRoom), &updatedRoom); err != nil {
		t.Fatal(err)
	}
	if updatedRoom.Status != entity.RoomStatusWaiting || len(updatedRoom.UserIDs) != 1 || updatedRoom.UserIDs[0] != otherID || len(updatedRoom.DisconnectedUserIDs) != 0 {
		t.Fatalf("room after game end = %+v, want only connected member and no disconnect markers", updatedRoom)
	}
}

func setActiveSession(t *testing.T, client *redis.Client, userID string) string {
	t.Helper()
	sessionID := "disconnect-session-" + strconv.FormatInt(time.Now().UnixNano(), 10)
	key := "session_account:" + userID
	if err := client.Set(context.Background(), key, "session:"+sessionID, 7*24*time.Hour).Err(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Del(context.Background(), key).Err() })
	return sessionID
}

func TestEndRoomGameDeletesRoomWhenAllMembersAreDisconnected(t *testing.T) {
	client := resilienceRedis(t)
	suffix := strconv.FormatInt(time.Now().UnixNano(), 10)
	roomID := "disconnect-all-game-" + suffix
	userID := "disconnect-all-player-" + suffix
	roomKey := "room:" + roomID
	statusKey := "user_status:" + userID
	t.Cleanup(func() { _ = client.Del(context.Background(), roomKey, statusKey).Err() })
	roomJSON, _ := json.Marshal(entity.Room{
		RoomID: roomID, MaximumSize: 4, UserIDs: []string{userID},
		DisconnectedUserIDs: []string{userID}, HosterID: userID,
		Prepared: []string{userID}, Status: entity.RoomStatusInGame,
	})
	statusJSON, _ := json.Marshal(entity.UserStatus{
		UserID: userID, Status: entity.UserStatusInGame, RoomID: roomID,
		ConnectionStatus: entity.UserConnectionStatusDisconnected,
	})
	if err := client.Set(context.Background(), roomKey, roomJSON, 0).Err(); err != nil {
		t.Fatal(err)
	}
	if err := client.Set(context.Background(), statusKey, statusJSON, 0).Err(); err != nil {
		t.Fatal(err)
	}

	if err := service.EndRoomGame(client, roomID); err != nil {
		t.Fatalf("EndRoomGame() error = %v", err)
	}
	if _, err := client.Get(context.Background(), roomKey).Result(); err != redis.Nil {
		t.Fatalf("room lookup error = %v, want redis.Nil", err)
	}
	if _, err := client.Get(context.Background(), statusKey).Result(); err != redis.Nil {
		t.Fatalf("user status lookup error = %v, want redis.Nil", err)
	}
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
