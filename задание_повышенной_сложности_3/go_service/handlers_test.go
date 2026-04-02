package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func resetItems() {
	items = []Item{}
	nextID = 1
}

func bearerHeader(t *testing.T, username string) string {
	t.Helper()
	tok, err := GenerateToken(username)
	require.NoError(t, err)
	return "Bearer " + tok
}

// TestLogin_Success — корректные учётные данные дают токен.
func TestLogin_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := setupRouter()

	body, _ := json.Marshal(map[string]string{"username": "admin", "password": "password123"})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NotEmpty(t, resp["token"])
}

// TestLogin_WrongPassword — неверный пароль → 401.
func TestLogin_WrongPassword(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := setupRouter()

	body, _ := json.Marshal(map[string]string{"username": "admin", "password": "wrong"})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestLogin_UnknownUser — несуществующий пользователь → 401.
func TestLogin_UnknownUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := setupRouter()

	body, _ := json.Marshal(map[string]string{"username": "ghost", "password": "x"})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestProfile_WithToken — валидный токен даёт профиль.
func TestProfile_WithToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/profile", nil)
	req.Header.Set("Authorization", bearerHeader(t, "admin"))
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "admin", resp["username"])
}

// TestProfile_NoToken — запрос без токена → 401.
func TestProfile_NoToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/profile", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestProfile_InvalidToken — подделанный токен → 401.
func TestProfile_InvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/profile", nil)
	req.Header.Set("Authorization", "Bearer fake.token.value")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestCreateItem_Authenticated — авторизованный пользователь создаёт item.
func TestCreateItem_Authenticated(t *testing.T) {
	resetItems()
	gin.SetMode(gin.TestMode)
	r := setupRouter()

	body, _ := json.Marshal(map[string]string{"name": "Widget"})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/items", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", bearerHeader(t, "user"))
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var item Item
	json.Unmarshal(w.Body.Bytes(), &item)
	assert.Equal(t, "Widget", item.Name)
	assert.Equal(t, "user", item.Owner)
}

// TestListItems_Authenticated — авторизованный пользователь видит список.
func TestListItems_Authenticated(t *testing.T) {
	resetItems()
	gin.SetMode(gin.TestMode)
	r := setupRouter()

	auth := bearerHeader(t, "admin")

	for i := 0; i < 3; i++ {
		body, _ := json.Marshal(map[string]string{"name": fmt.Sprintf("Item%d", i)})
		req, _ := http.NewRequest("POST", "/api/items", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", auth)
		r.ServeHTTP(httptest.NewRecorder(), req)
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/items", nil)
	req.Header.Set("Authorization", auth)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var result []Item
	json.Unmarshal(w.Body.Bytes(), &result)
	assert.Len(t, result, 3)
}

// TestTokenFromLogin_UsedForProtectedRoute — E2E: логин → используем токен.
func TestTokenFromLogin_UsedForProtectedRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := setupRouter()

	// 1. Логин
	body, _ := json.Marshal(map[string]string{"username": "user", "password": "secret"})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var loginResp map[string]string
	json.Unmarshal(w.Body.Bytes(), &loginResp)
	token := loginResp["token"]
	require.NotEmpty(t, token)

	// 2. Используем токен для профиля
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/api/profile", nil)
	req2.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)
}
