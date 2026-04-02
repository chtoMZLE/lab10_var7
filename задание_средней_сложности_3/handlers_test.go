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

func TestCreateUser_Valid(t *testing.T) {
	resetState()
	gin.SetMode(gin.TestMode)
	r := setupRouter()

	body, _ := json.Marshal(map[string]interface{}{
		"name":  "Alice",
		"email": "alice@example.com",
		"age":   30,
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/users", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var u User
	json.Unmarshal(w.Body.Bytes(), &u)
	assert.Equal(t, "Alice", u.Name)
	assert.Equal(t, "alice@example.com", u.Email)
}

func TestCreateUser_MissingName(t *testing.T) {
	resetState()
	gin.SetMode(gin.TestMode)
	r := setupRouter()

	body, _ := json.Marshal(map[string]string{"email": "alice@example.com"})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/users", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	var resp map[string][]ValidationError
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NotEmpty(t, resp["errors"])
}

func TestCreateUser_InvalidEmail(t *testing.T) {
	resetState()
	gin.SetMode(gin.TestMode)
	r := setupRouter()

	body, _ := json.Marshal(map[string]string{"name": "Alice", "email": "not-email"})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/users", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestCreateUser_ShortName(t *testing.T) {
	resetState()
	gin.SetMode(gin.TestMode)
	r := setupRouter()

	body, _ := json.Marshal(map[string]string{"name": "A", "email": "alice@example.com"})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/users", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestCreateUser_InvalidAge(t *testing.T) {
	resetState()
	gin.SetMode(gin.TestMode)
	r := setupRouter()

	body, _ := json.Marshal(map[string]interface{}{"name": "Alice", "email": "a@b.com", "age": 999})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/users", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestCreateUser_BadJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/users", bytes.NewBufferString("{bad json"))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateUser_Valid(t *testing.T) {
	resetState()
	gin.SetMode(gin.TestMode)
	r := setupRouter()

	// создаём пользователя
	body, _ := json.Marshal(map[string]string{"name": "Bob", "email": "bob@example.com"})
	req, _ := http.NewRequest("POST", "/users", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(httptest.NewRecorder(), req)

	// обновляем
	upd, _ := json.Marshal(map[string]string{"name": "Robert"})
	w := httptest.NewRecorder()
	req2, _ := http.NewRequest("PUT", "/users/1", bytes.NewBuffer(upd))
	req2.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req2)

	assert.Equal(t, http.StatusOK, w.Code)
	var u User
	json.Unmarshal(w.Body.Bytes(), &u)
	assert.Equal(t, "Robert", u.Name)
}

func TestUpdateUser_NotFound(t *testing.T) {
	resetState()
	gin.SetMode(gin.TestMode)
	r := setupRouter()

	body, _ := json.Marshal(map[string]string{"name": "Xavier"})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/users/999", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}
