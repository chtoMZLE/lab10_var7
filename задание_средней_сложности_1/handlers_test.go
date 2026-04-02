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
	assert.Equal(t, 0, len(resp))
}

func TestCreateUser(t *testing.T) {
	resetState()
	gin.SetMode(gin.TestMode)
	r := setupRouter()

	body, _ := json.Marshal(map[string]string{
		"name":  "Alice",
		"email": "alice@example.com",
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/users", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp User
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, 1, resp.ID)
	assert.Equal(t, "Alice", resp.Name)
	assert.Equal(t, "alice@example.com", resp.Email)
}

func TestGetUsersAfterCreate(t *testing.T) {
	resetState()
	gin.SetMode(gin.TestMode)
	r := setupRouter()

	body, _ := json.Marshal(map[string]string{"name": "Bob", "email": "bob@example.com"})
	req, _ := http.NewRequest("POST", "/users", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(httptest.NewRecorder(), req)

	w := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/users", nil)
	r.ServeHTTP(w, req2)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp []User
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, 1, len(resp))
	assert.Equal(t, "Bob", resp[0].Name)
}

func TestGetUserByID(t *testing.T) {
	resetState()
	gin.SetMode(gin.TestMode)
	r := setupRouter()

	body, _ := json.Marshal(map[string]string{"name": "Carol", "email": "carol@example.com"})
	req, _ := http.NewRequest("POST", "/users", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(httptest.NewRecorder(), req)

	w := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/users/1", nil)
	r.ServeHTTP(w, req2)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp User
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "Carol", resp.Name)
}

func TestGetUserByIDNotFound(t *testing.T) {
	resetState()
	gin.SetMode(gin.TestMode)
	r := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/users/999", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestCreateUserInvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/users", bytes.NewBufferString("not-json"))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
