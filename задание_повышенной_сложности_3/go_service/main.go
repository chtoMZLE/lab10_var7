package main

import "github.com/gin-gonic/gin"

func setupRouter() *gin.Engine {
	r := gin.Default()

	r.POST("/auth/login", loginHandler)

	api := r.Group("/api")
	api.Use(JWTMiddleware())
	{
		api.GET("/profile", profileHandler)
		api.GET("/items", listItemsHandler)
		api.POST("/items", createItemHandler)
	}

	return r
}

func main() {
	r := setupRouter()
	r.Run(":8080")
}
