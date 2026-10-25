package resiliencetest

import (
	"context"
	"net"
	"sync/atomic"
	"testing"
	"time"

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
	attempts atomic.Int32
}

func (s *countedUnavailableUserData) UserRegistration(context.Context, *userdatapb.UserRegistrationRequest) (*userdatapb.UserRegistrationResponse, error) {
	s.attempts.Add(1)
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

func TestGRPCClientsDoNotRetryReturnedUnavailableErrors(t *testing.T) {
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
