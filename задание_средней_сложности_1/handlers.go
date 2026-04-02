package main

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
)

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

var (
	users  = []User{}
	nextID = 1
	mu     sync.Mutex
)

func healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func getUsersHandler(c *gin.Context) {
	mu.Lock()
	defer mu.Unlock()
	c.JSON(http.StatusOK, users)
}

func getUserByIDHandler(c *gin.Context) {
	id := c.Param("id")
	mu.Lock()
	defer mu.Unlock()
	for _, u := range users {
		if fmt.Sprintf("%d", u.ID) == id {
			c.JSON(http.StatusOK, u)
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
}

func createUserHandler(c *gin.Context) {
	var newUser User
	if err := c.ShouldBindJSON(&newUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	mu.Lock()
	newUser.ID = nextID
	nextID++
	users = append(users, newUser)
	mu.Unlock()
	c.JSON(http.StatusCreated, newUser)
}
