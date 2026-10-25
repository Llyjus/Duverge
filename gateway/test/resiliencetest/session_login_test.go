package resiliencetest

import (
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"gateway/internal/service/html/sessionLogin"

	"github.com/redis/go-redis/v9"
)

func TestRedisFailureReturnsServiceUnavailable(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr:         "127.0.0.1:1",
		MaxRetries:   -1,
		DialTimeout:  100 * time.Millisecond,
		ReadTimeout:  100 * time.Millisecond,
		WriteTimeout: 100 * time.Millisecond,
	})
	t.Cleanup(func() { _ = client.Close() })

	request := httptest.NewRequest("POST", "/login_session", strings.NewReader(`{"sessionId":"probe"}`))
	response := httptest.NewRecorder()
	sessionLogin.HandleLoginSession(client)(response, request)

	if response.Code != 503 {
		t.Fatalf("HTTP status = %d, want 503; body: %s", response.Code, response.Body.String())
	}
	if !strings.Contains(response.Body.String(), "Session service unavailable") {
		t.Fatalf("unexpected response body: %s", response.Body.String())
	}
}
