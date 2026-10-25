package service

import (
	"context"
	"encoding/json"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"

	"room/internal/entity"
	"room/internal/service"

	"github.com/redis/go-redis/v9"
)

func resilienceRedis(t *testing.T) *redis.Client {
	t.Helper()
	addr := os.Getenv("RESILIENCE_REDIS_ADDR")
	if addr == "" {
		t.Skip("set RESILIENCE_REDIS_ADDR to run this Redis integration test")
	}
	client := redis.NewClient(&redis.Options{Addr: addr})
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		t.Fatalf("Redis is unavailable at %s: %v", addr, err)
	}
	t.Cleanup(func() { _ = client.Close() })
	return client
}

func TestCreateRoomRetryIsNotIdempotent(t *testing.T) {
	client := resilienceRedis(t)
	hostID := "resilience-host-" + strconv.FormatInt(time.Now().UnixNano(), 10)
	t.Cleanup(func() {
		_ = client.Del(context.Background(), "user_status:"+hostID).Err()
	})

	firstRoomID, err := service.CreateRoom(client, hostID)
	if err != nil {
		t.Fatal(err)
	}
	secondRoomID, err := service.CreateRoom(client, hostID)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = client.Del(context.Background(), "room:"+firstRoomID, "room:"+secondRoomID).Err()
	})

	if firstRoomID == secondRoomID {
		t.Fatalf("expected the characterization check to expose a second room, got same ID %q", firstRoomID)
	}
	t.Logf("same logical create called twice returned two room IDs (%s, %s); a lost-response retry is not idempotent", firstRoomID, secondRoomID)
}

func TestConcurrentJoinsPreserveRoomAndUserStatus(t *testing.T) {
	client := resilienceRedis(t)
	suffix := strconv.FormatInt(time.Now().UnixNano(), 10)
	roomID := "resilience-room-" + suffix
	hostID := "resilience-host-" + suffix
	guestIDs := []string{"resilience-a-" + suffix, "resilience-b-" + suffix}
	roomKey := "room:" + roomID
	statusKeys := []string{"user_status:" + hostID, "user_status:" + guestIDs[0], "user_status:" + guestIDs[1]}
	t.Cleanup(func() {
		_ = client.Del(context.Background(), append([]string{roomKey}, statusKeys...)...).Err()
	})

	roomJSON, err := json.Marshal(entity.Room{
		RoomID:      roomID,
		MaximumSize: 4,
		UserIDs:     []string{hostID},
		HosterID:    hostID,
		Prepared:    []string{},
		Status:      entity.RoomStatusWaiting,
	})
	if err != nil {
		t.Fatal(err)
	}
	hostStatus, _ := json.Marshal(entity.UserStatus{UserID: hostID, Status: entity.UserStatusInRoom})
	if err := client.Set(context.Background(), roomKey, roomJSON, 0).Err(); err != nil {
		t.Fatal(err)
	}
	if err := client.Set(context.Background(), statusKeys[0], hostStatus, 0).Err(); err != nil {
		t.Fatal(err)
	}
	for _, guestID := range guestIDs {
		statusJSON, _ := json.Marshal(entity.UserStatus{UserID: guestID, Status: entity.UserStatusInLobby})
		if err := client.Set(context.Background(), "user_status:"+guestID, statusJSON, 0).Err(); err != nil {
			t.Fatal(err)
		}
	}

	errs := make(chan error, len(guestIDs))
	var wg sync.WaitGroup
	for _, guestID := range guestIDs {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()
			errs <- service.AddUserToRoom(client, roomID, id)
		}(guestID)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Errorf("concurrent join returned an error after retries: %v", err)
		}
	}

	stored, err := client.Get(context.Background(), roomKey).Result()
	if err != nil {
		t.Fatal(err)
	}
	var finalRoom entity.Room
	if err := json.Unmarshal([]byte(stored), &finalRoom); err != nil {
		t.Fatal(err)
	}
	members := make(map[string]bool, len(finalRoom.UserIDs))
	for _, userID := range finalRoom.UserIDs {
		if members[userID] {
			t.Fatalf("duplicate room member %q", userID)
		}
		members[userID] = true
	}
	if len(finalRoom.UserIDs) != 3 || !members[hostID] || !members[guestIDs[0]] || !members[guestIDs[1]] {
		t.Fatalf("room membership is inconsistent after joins: %+v", finalRoom.UserIDs)
	}
	for _, userID := range append([]string{hostID}, guestIDs...) {
		statusJSON, err := client.Get(context.Background(), "user_status:"+userID).Result()
		if err != nil {
			t.Fatal(err)
		}
		var userStatus entity.UserStatus
		if err := json.Unmarshal([]byte(statusJSON), &userStatus); err != nil {
			t.Fatal(err)
		}
		if userStatus.Status != entity.UserStatusInRoom || !members[userID] {
			t.Fatalf("room/user status diverged for %q: status=%q member=%v", userID, userStatus.Status, members[userID])
		}
	}
}
