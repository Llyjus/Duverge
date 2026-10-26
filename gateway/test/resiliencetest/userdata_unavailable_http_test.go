package resiliencetest

import (
	"context"
	"net"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	userdataclient "gateway/internal/client/userdata"
	loginhandler "gateway/internal/service/html/login"
)

func TestLoginAndRegistrationReportUserDataUnavailable(t *testing.T) {
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
	handler := loginhandler.HandleLogin(client, nil)

	for _, operation := range []string{"register", "login"} {
		t.Run(operation, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			body := `{"type":"` + operation + `","accountId":"probe","password":"probe"}`
			request := httptest.NewRequest("POST", "/login", strings.NewReader(body)).WithContext(ctx)
			response := httptest.NewRecorder()

			handler(response, request)

			if response.Code != 503 {
				t.Fatalf("HTTP status = %d, want 503; body: %s", response.Code, response.Body.String())
			}
			if !strings.Contains(response.Body.String(), "UserData service unavailable") {
				t.Fatalf("unexpected response body: %s", response.Body.String())
			}
		})
	}
}
