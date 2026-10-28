package resiliencetest

import (
	"context"
	"net"
	"sync/atomic"
	"testing"
	"time"

	"gateway/internal/client/retry"
	roomclient "gateway/internal/client/room"
	userdataclient "gateway/internal/client/userdata"
	roompb "gateway/proto/room/gen"
	userdatapb "gateway/proto/userdata/gen"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type delayedUserData struct {
	userdatapb.UnimplementedUserDataServiceServer
}

func (delayedUserData) UserRegistration(ctx context.Context, _ *userdatapb.UserRegistrationRequest) (*userdatapb.UserRegistrationResponse, error) {
	<-ctx.Done()
	return nil, status.Error(codes.DeadlineExceeded, "test delay ended with context")
}

func (delayedUserData) UserLogin(ctx context.Context, _ *userdatapb.UserLoginRequest) (*userdatapb.UserLoginResponse, error) {
	<-ctx.Done()
	return nil, status.Error(codes.DeadlineExceeded, "test delay ended with context")
}

type delayedRoom struct {
	roompb.UnimplementedRoomServiceServer
}

type countedUnavailableUserData struct {
	userdatapb.UnimplementedUserDataServiceServer
	attempts  atomic.Int32
	messageID string
}

func (s *countedUnavailableUserData) UserRegistration(_ context.Context, request *userdatapb.UserRegistrationRequest) (*userdatapb.UserRegistrationResponse, error) {
	s.attempts.Add(1)
	s.messageID = request.GetMessageId()
	return nil, status.Error(codes.Unavailable, "injected unavailable error")
}

type countedUnavailableRoom struct {
	roompb.UnimplementedRoomServiceServer
	attempts atomic.Int32
}

func (s *countedUnavailableRoom) CreateRoom(context.Context, *roompb.CreateRoomRequest) (*roompb.CreateRoomResponse, error) {
	s.attempts.Add(1)
	return nil, status.Error(codes.Unavailable, "injected unavailable error")
}

type intermittentlyUnavailableUserData struct {
	userdatapb.UnimplementedUserDataServiceServer
	attempts   atomic.Int32
	messageIDs chan string
}

func (s *intermittentlyUnavailableUserData) UserLogin(_ context.Context, request *userdatapb.UserLoginRequest) (*userdatapb.UserLoginResponse, error) {
	s.messageIDs <- request.GetMessageId()
	if s.attempts.Add(1) < 3 {
		return nil, status.Error(codes.Unavailable, "injected temporary failure")
	}
	return &userdatapb.UserLoginResponse{MessageId: request.GetMessageId(), Result: true, SessionId: "test-session"}, nil
}

type intermittentlyUnavailableRoom struct {
	roompb.UnimplementedRoomServiceServer
	attempts atomic.Int32
}

func (s *intermittentlyUnavailableRoom) GetRoomInfo(context.Context, *roompb.GetRoomInfoRequest) (*roompb.GetRoomInfoResponse, error) {
	if s.attempts.Add(1) < 3 {
		return nil, status.Error(codes.Unavailable, "injected temporary failure")
	}
	return &roompb.GetRoomInfoResponse{RoomId: "room-1", UserIds: []string{"alice"}}, nil
}

func (delayedRoom) CreateRoom(ctx context.Context, _ *roompb.CreateRoomRequest) (*roompb.CreateRoomResponse, error) {
	<-ctx.Done()
	return nil, status.Error(codes.DeadlineExceeded, "test delay ended with context")
}

func (delayedRoom) GetRoomInfo(ctx context.Context, _ *roompb.GetRoomInfoRequest) (*roompb.GetRoomInfoResponse, error) {
	<-ctx.Done()
	return nil, status.Error(codes.DeadlineExceeded, "test delay ended with context")
}

func (delayedRoom) JoinRoom(ctx context.Context, _ *roompb.JoinRoomRequest) (*roompb.JoinRoomResponse, error) {
	<-ctx.Done()
	return nil, status.Error(codes.DeadlineExceeded, "test delay ended with context")
}

func (delayedRoom) LeaveRoom(ctx context.Context, _ *roompb.LeaveRoomRequest) (*roompb.LeaveRoomResponse, error) {
	<-ctx.Done()
	return nil, status.Error(codes.DeadlineExceeded, "test delay ended with context")
}

func (delayedRoom) SetPreparation(ctx context.Context, _ *roompb.SetPreparationRequest) (*roompb.SetPreparationResponse, error) {
	<-ctx.Done()
	return nil, status.Error(codes.DeadlineExceeded, "test delay ended with context")
}

func (delayedRoom) StartGame(ctx context.Context, _ *roompb.StartGameRequest) (*roompb.StartGameResponse, error) {
	<-ctx.Done()
	return nil, status.Error(codes.DeadlineExceeded, "test delay ended with context")
}

func (delayedRoom) EndGame(ctx context.Context, _ *roompb.EndGameRequest) (*roompb.EndGameResponse, error) {
	<-ctx.Done()
	return nil, status.Error(codes.DeadlineExceeded, "test delay ended with context")
}

func TestGatewayGRPCCallsHonorRequestDeadline(t *testing.T) {
	userDataAddr := startGRPCServer(t, func(server *grpc.Server) {
		userdatapb.RegisterUserDataServiceServer(server, delayedUserData{})
	})
	roomAddr := startGRPCServer(t, func(server *grpc.Server) {
		roompb.RegisterRoomServiceServer(server, delayedRoom{})
	})

	userData, err := userdataclient.NewClient(userDataAddr)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(userData.Close)
	rooms, err := roomclient.NewClient(roomAddr)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(rooms.Close)

	tests := []struct {
		name string
		call func(context.Context) error
	}{
		{"registration", func(ctx context.Context) error {
			_, err := userData.UserRegistration(ctx, &userdatapb.UserRegistrationRequest{})
			return err
		}},
		{"login", func(ctx context.Context) error {
			_, err := userData.UserLogin(ctx, &userdatapb.UserLoginRequest{})
			return err
		}},
		{"room creation", func(ctx context.Context) error {
			_, err := rooms.GetClient().CreateRoom(ctx, &roompb.CreateRoomRequest{})
			return err
		}},
		{"room lookup", func(ctx context.Context) error {
			_, err := rooms.GetClient().GetRoomInfo(ctx, &roompb.GetRoomInfoRequest{})
			return err
		}},
		{"room join", func(ctx context.Context) error {
			_, err := rooms.GetClient().JoinRoom(ctx, &roompb.JoinRoomRequest{})
			return err
		}},
		{"room leave", func(ctx context.Context) error {
			_, err := rooms.GetClient().LeaveRoom(ctx, &roompb.LeaveRoomRequest{})
			return err
		}},
		{"room preparation", func(ctx context.Context) error {
			_, err := rooms.GetClient().SetPreparation(ctx, &roompb.SetPreparationRequest{})
			return err
		}},
		{"room start game", func(ctx context.Context) error {
			_, err := rooms.GetClient().StartGame(ctx, &roompb.StartGameRequest{})
			return err
		}},
		{"room end game", func(ctx context.Context) error {
			_, err := rooms.GetClient().EndGame(ctx, &roompb.EndGameRequest{})
			return err
		}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
			defer cancel()
			started := time.Now()
			err := test.call(ctx)
			elapsed := time.Since(started)
			if status.Code(err) != codes.DeadlineExceeded {
				t.Fatalf("expected DeadlineExceeded, got %v", err)
			}
			if elapsed > time.Second {
				t.Fatalf("call exceeded its deadline by too much: %v", elapsed)
			}
		})
	}
}

func TestGatewayGRPCCallToUnavailableServiceReturns(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := listener.Addr().String()
	if err := listener.Close(); err != nil {
		t.Fatal(err)
	}

	client, err := userdataclient.NewClient(addr)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(client.Close)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	started := time.Now()
	_, err = client.UserRegistration(ctx, &userdatapb.UserRegistrationRequest{})
	if err == nil {
		t.Fatal("expected an error from the unavailable service")
	}
	if time.Since(started) > 2*time.Second {
		t.Fatalf("call waited past its one-second deadline: %v", time.Since(started))
	}
}

func TestGatewayDoesNotRetryNonIdempotentRPCs(t *testing.T) {
	userDataService := &countedUnavailableUserData{}
	userDataAddr := startGRPCServer(t, func(server *grpc.Server) {
		userdatapb.RegisterUserDataServiceServer(server, userDataService)
	})
	roomService := &countedUnavailableRoom{}
	roomAddr := startGRPCServer(t, func(server *grpc.Server) {
		roompb.RegisterRoomServiceServer(server, roomService)
	})

	userData, err := userdataclient.NewClient(userDataAddr)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(userData.Close)
	rooms, err := roomclient.NewClient(roomAddr)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(rooms.Close)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, _ = userData.UserRegistration(ctx, &userdatapb.UserRegistrationRequest{})
	_, _ = rooms.GetClient().CreateRoom(ctx, &roompb.CreateRoomRequest{})
	if got := userDataService.attempts.Load(); got != 1 {
		t.Errorf("UserData registration attempts = %d, want 1", got)
	}
	if got := roomService.attempts.Load(); got != 1 {
		t.Errorf("Room create attempts = %d, want 1", got)
	}
	if userDataService.messageID == "" {
		t.Error("UserData registration request has empty message_id")
	}
}

func TestGatewayRetriesSafeRPCsOnUnavailable(t *testing.T) {
	userDataService := &intermittentlyUnavailableUserData{messageIDs: make(chan string, 3)}
	userDataAddr := startGRPCServer(t, func(server *grpc.Server) {
		userdatapb.RegisterUserDataServiceServer(server, userDataService)
	})
	roomService := &intermittentlyUnavailableRoom{}
	roomAddr := startGRPCServer(t, func(server *grpc.Server) {
		roompb.RegisterRoomServiceServer(server, roomService)
	})

	policy := retry.Policy{MaxAttempts: 3, Interval: 5 * time.Millisecond}
	userData, err := userdataclient.NewClientWithRetryPolicy(userDataAddr, policy)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(userData.Close)
	rooms, err := roomclient.NewClientWithRetryPolicy(roomAddr, policy)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(rooms.Close)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	loginResponse, err := userData.UserLogin(ctx, &userdatapb.UserLoginRequest{})
	if err != nil || !loginResponse.GetResult() {
		t.Fatalf("UserData login after retry: response=%v err=%v", loginResponse, err)
	}
	if loginResponse.GetMessageId() == "" {
		t.Fatal("UserData login response has empty message_id")
	}
	roomResponse, err := rooms.GetRoomInfo(ctx, &roompb.GetRoomInfoRequest{})
	if err != nil || roomResponse.GetRoomId() != "room-1" {
		t.Fatalf("Room lookup after retry: response=%v err=%v", roomResponse, err)
	}
	if got := userDataService.attempts.Load(); got != 3 {
		t.Errorf("UserData login attempts = %d, want 3", got)
	}
	for attempt := 0; attempt < 3; attempt++ {
		if got := <-userDataService.messageIDs; got != loginResponse.GetMessageId() {
			t.Errorf("attempt %d message_id = %q, want %q", attempt+1, got, loginResponse.GetMessageId())
		}
	}
	if got := roomService.attempts.Load(); got != 3 {
		t.Errorf("Room lookup attempts = %d, want 3", got)
	}
}

func startGRPCServer(t *testing.T, register func(*grpc.Server)) string {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	server := grpc.NewServer()
	register(server)
	go func() { _ = server.Serve(listener) }()
	t.Cleanup(func() {
		server.Stop()
		_ = listener.Close()
	})
	return listener.Addr().String()
}
