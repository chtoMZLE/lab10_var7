package main

import "github.com/gin-gonic/gin"

func setupRouter() *gin.Engine {
	r := gin.Default()
	r.GET("/health", healthHandler)
	r.GET("/users", getUsersHandler)
	r.GET("/users/:id", getUserByIDHandler)
	r.POST("/users", createUserHandler)
	r.PUT("/users/:id", updateUserHandler)
	return r
}

func main() {
	r := setupRouter()
	r.Run(":8080")
}
