package main

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
)

// validUsers — статическое хранилище учётных данных (для демо).
var validUsers = map[string]string{
	"admin": "password123",
	"user":  "secret",
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type Item struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Owner string `json:"owner"`
}

var (
	items  = []Item{}
	nextID = 1
	mu     sync.Mutex
)

// loginHandler выдаёт JWT при корректных учётных данных.
func loginHandler(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	expected, ok := validUsers[req.Username]
	if !ok || expected != req.Password {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	token, err := GenerateToken(req.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

// profileHandler возвращает данные текущего авторизованного пользователя.
func profileHandler(c *gin.Context) {
	username, _ := c.Get("username")
	c.JSON(http.StatusOK, gin.H{
		"username": username,
		"role":     "user",
	})
}

// listItemsHandler возвращает список всех предметов (защищён JWT).
func listItemsHandler(c *gin.Context) {
	mu.Lock()
	defer mu.Unlock()
	c.JSON(http.StatusOK, items)
}

// createItemHandler создаёт предмет от имени текущего пользователя.
func createItemHandler(c *gin.Context) {
	var body struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	username, _ := c.Get("username")
	mu.Lock()
	item := Item{ID: nextID, Name: body.Name, Owner: fmt.Sprintf("%v", username)}
	nextID++
	items = append(items, item)
	mu.Unlock()

	c.JSON(http.StatusCreated, item)
}
