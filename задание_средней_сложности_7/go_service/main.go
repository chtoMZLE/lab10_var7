package main

import (
	"time"

	"github.com/gin-gonic/gin"
)

func setupRouter() *gin.Engine {
	r := gin.Default()
	r.GET("/health", healthHandler)
	r.GET("/users", getUsersHandler)
	r.GET("/users/:id", getUserByIDHandler)
	r.POST("/users", createUserHandler)
	return r
}

func main() {
	r := setupRouter()
	srv := NewServer(":8080", r)
	RunWithGracefulShutdown(srv, 5*time.Second)
}
