package resiliencetest

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

func TestRedisDefaultRetryCountIsBounded(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	var pingAttempts atomic.Int32
	serverDone := make(chan struct{})
	go func() {
		defer close(serverDone)
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go func(conn net.Conn) {
				defer conn.Close()
				reader := bufio.NewReader(conn)
				for {
					command, err := readRESPCommand(reader)
					if err != nil {
						return
					}
					if len(command) > 0 && strings.EqualFold(command[0], "PING") {
						pingAttempts.Add(1)
						_, _ = io.WriteString(conn, "-LOADING redis is loading the dataset in memory\r\n")
						continue
					}
					if len(command) > 0 && strings.EqualFold(command[0], "HELLO") {
						// Tell go-redis that RESP3 negotiation is unavailable, so it
						// falls back to the explicitly configured RESP2 protocol.
						_, _ = io.WriteString(conn, "-ERR unknown command 'HELLO'\r\n")
						continue
					}
					_, _ = io.WriteString(conn, "+OK\r\n")
				}
			}(conn)
		}
	}()

	client := redis.NewClient(&redis.Options{
		Addr:            listener.Addr().String(),
		Protocol:        2,
		DisableIdentity: true,
		MinRetryBackoff: -1,
		MaxRetryBackoff: -1,
		DialTimeout:     time.Second,
		ReadTimeout:     time.Second,
		WriteTimeout:    time.Second,
	})
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err = client.Ping(ctx).Err()
	if err == nil || !strings.Contains(err.Error(), "LOADING") {
		t.Fatalf("expected final LOADING error after retries, got %v", err)
	}
	if got := pingAttempts.Load(); got != 4 {
		t.Fatalf("PING attempts = %d, want initial attempt plus 3 default retries", got)
	}

	_ = listener.Close()
	<-serverDone
}

func readRESPCommand(reader *bufio.Reader) ([]string, error) {
	line, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	if len(line) < 3 || line[0] != '*' {
		return nil, fmt.Errorf("expected RESP array, got %q", line)
	}
	count, err := strconv.Atoi(strings.TrimSpace(line[1:]))
	if err != nil {
		return nil, err
	}
	command := make([]string, 0, count)
	for i := 0; i < count; i++ {
		lengthLine, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		if len(lengthLine) < 3 || lengthLine[0] != '$' {
			return nil, fmt.Errorf("expected RESP bulk string, got %q", lengthLine)
		}
		length, err := strconv.Atoi(strings.TrimSpace(lengthLine[1:]))
		if err != nil {
			return nil, err
		}
		value := make([]byte, length+2)
		if _, err := io.ReadFull(reader, value); err != nil {
			return nil, err
		}
		command = append(command, string(value[:length]))
	}
	return command, nil
}
