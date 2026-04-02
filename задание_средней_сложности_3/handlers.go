package main

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
)

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
	c.JSON(http.StatusNotFound, gin.H{"error": "пользователь не найден"})
}

func createUserHandler(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": formatValidationErrors(err)})
		return
	}

	if errs := ValidateStruct(req); errs != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"errors": errs})
		return
	}

	mu.Lock()
	u := User{ID: nextID, Name: req.Name, Email: req.Email, Age: req.Age}
	nextID++
	users = append(users, u)
	mu.Unlock()

	c.JSON(http.StatusCreated, u)
}

func updateUserHandler(c *gin.Context) {
	id := c.Param("id")

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": formatValidationErrors(err)})
		return
	}

	if errs := ValidateStruct(req); errs != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"errors": errs})
		return
	}

	mu.Lock()
	defer mu.Unlock()
	for i, u := range users {
		if fmt.Sprintf("%d", u.ID) == id {
			if req.Name != "" {
				users[i].Name = req.Name
			}
			if req.Email != "" {
				users[i].Email = req.Email
			}
			if req.Age != 0 {
				users[i].Age = req.Age
			}
			c.JSON(http.StatusOK, users[i])
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "пользователь не найден"})
}
