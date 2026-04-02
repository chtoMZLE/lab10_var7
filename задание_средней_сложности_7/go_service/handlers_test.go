package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func resetState() {
	users = []User{}
	nextID = 1
}

func TestHealthHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := setupRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "ok", resp["status"])
}

func TestGetUsersEmpty(t *testing.T) {
	resetState()
	gin.SetMode(gin.TestMode)
	r := setupRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/users", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var resp []User
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Len(t, resp, 0)
}

func TestCreateAndGetUser(t *testing.T) {
	resetState()
	gin.SetMode(gin.TestMode)
	r := setupRouter()

	body, _ := json.Marshal(map[string]string{"name": "Alice", "email": "alice@example.com"})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/users", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/users/1", nil)
	r.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)
	var u User
	json.Unmarshal(w2.Body.Bytes(), &u)
	assert.Equal(t, "Alice", u.Name)
}
