package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

// NewServer создаёт http.Server с заданным адресом и роутером.
func NewServer(addr string, r *gin.Engine) *http.Server {
	return &http.Server{
		Addr:    addr,
		Handler: r,
	}
}

// RunWithGracefulShutdown запускает сервер и корректно завершает его
// при получении SIGINT или SIGTERM.
func RunWithGracefulShutdown(srv *http.Server, shutdownTimeout time.Duration) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen error: %v", err)
		}
	}()

	log.Printf("Server started on %s", srv.Addr)
	<-quit
	log.Println("Shutdown signal received, waiting for active connections...")

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
	log.Println("Server exited cleanly")
}
