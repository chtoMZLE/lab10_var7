package main

import (
	"context"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// startTestServer запускает сервер на случайном порту и возвращает адрес + функцию остановки.
func startTestServer(t *testing.T) (addr string, shutdown func()) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	srv := &http.Server{Handler: setupRouter()}

	go func() {
		srv.Serve(ln) //nolint:errcheck
	}()

	addr = "http://" + ln.Addr().String()
	shutdown = func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		srv.Shutdown(ctx) //nolint:errcheck
	}
	return addr, shutdown
}

func TestServerRespondsBeforeShutdown(t *testing.T) {
	addr, shutdown := startTestServer(t)
	defer shutdown()

	resp, err := http.Get(addr + "/health")
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestServerGracefulShutdown(t *testing.T) {
	gin.SetMode(gin.TestMode)

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	srv := &http.Server{Handler: setupRouter()}
	go func() {
		srv.Serve(ln) //nolint:errcheck
	}()

	serverAddr := "http://" + ln.Addr().String()

	// Сервер отвечает до остановки
	resp, err := http.Get(serverAddr + "/health")
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = srv.Shutdown(ctx)
	assert.NoError(t, err)

	// После остановки сервер не принимает запросы
	_, err = http.Get(serverAddr + "/health")
	assert.Error(t, err, "сервер должен отказать в соединении после shutdown")
}

func TestShutdownContextTimeout(t *testing.T) {
	gin.SetMode(gin.TestMode)

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	srv := &http.Server{Handler: setupRouter()}
	go func() {
		srv.Serve(ln) //nolint:errcheck
	}()

	// Используем очень короткий таймаут — Shutdown всё равно должен завершиться без паники
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()
	// Ошибка допустима (DeadlineExceeded), паники быть не должно
	_ = srv.Shutdown(ctx)
}

func TestNewServer(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := setupRouter()
	srv := NewServer(":9999", r)
	assert.Equal(t, ":9999", srv.Addr)
	assert.Equal(t, r, srv.Handler)
}
