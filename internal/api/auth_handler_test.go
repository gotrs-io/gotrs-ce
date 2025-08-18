package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/gotrs-io/gotrs-ce/internal/auth"
	"github.com/gotrs-io/gotrs-ce/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Login with valid request body", func(t *testing.T) {
		// Create mock database
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		jwtManager := auth.NewJWTManager("test-secret", 1*time.Hour)
		authService := auth.NewAuthService(db, jwtManager)
		handler := NewAuthHandler(authService)

		loginReq := models.LoginRequest{
			Email:    "test@example.com",
			Password: "password123",
		}

		// Create user with proper hashed password
		user := &models.User{Email: loginReq.Email}
		err = user.SetPassword(loginReq.Password)
		require.NoError(t, err)

		// Mock user query
		rows := sqlmock.NewRows([]string{
			"id", "login", "email", "pw", "title", "first_name", "last_name",
			"valid_id", "create_time", "create_by", "change_time", "change_by",
		}).AddRow(
			1, "testuser", loginReq.Email, user.Password, "Mr", "Test", "User",
			1, time.Now(), 1, time.Now(), 1,
		)

		mock.ExpectQuery("SELECT (.+) FROM users WHERE email = \\$1").
			WithArgs(loginReq.Email).
			WillReturnRows(rows)

		// Mock role determination
		mock.ExpectQuery("SELECT g.name FROM groups g").
			WithArgs(uint(1)).
			WillReturnError(sql.ErrNoRows)

		// Mock update for last login
		mock.ExpectExec("UPDATE users SET (.+) WHERE id = \\$1").
			WithArgs(1).
			WillReturnResult(sqlmock.NewResult(1, 1))

		// Create request
		body, _ := json.Marshal(loginReq)
		req := httptest.NewRequest("POST", "/auth/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		// Create gin context
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		// Call handler
		handler.Login(c)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)

		var responseWrapper struct {
			Success bool                 `json:"success"`
			Data    models.LoginResponse `json:"data"`
		}
		err = json.Unmarshal(w.Body.Bytes(), &responseWrapper)
		require.NoError(t, err)

		assert.True(t, responseWrapper.Success)
		assert.NotEmpty(t, responseWrapper.Data.Token)
		assert.NotEmpty(t, responseWrapper.Data.RefreshToken)
		assert.Equal(t, loginReq.Email, responseWrapper.Data.User.Email)
	})

	t.Run("Login with invalid request body", func(t *testing.T) {
		db, _, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		jwtManager := auth.NewJWTManager("test-secret", 1*time.Hour)
		authService := auth.NewAuthService(db, jwtManager)
		handler := NewAuthHandler(authService)

		// Invalid JSON
		req := httptest.NewRequest("POST", "/auth/login", bytes.NewBufferString("invalid json"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req

		handler.Login(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid request")
	})

	t.Run("Login with missing email", func(t *testing.T) {
		db, _, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		jwtManager := auth.NewJWTManager("test-secret", 1*time.Hour)
		authService := auth.NewAuthService(db, jwtManager)
		handler := NewAuthHandler(authService)

		loginReq := models.LoginRequest{
			Password: "password123",
		}

		body, _ := json.Marshal(loginReq)
		req := httptest.NewRequest("POST", "/auth/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req

		handler.Login(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid request format")
	})

	t.Run("Login with wrong password", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		jwtManager := auth.NewJWTManager("test-secret", 1*time.Hour)
		authService := auth.NewAuthService(db, jwtManager)
		handler := NewAuthHandler(authService)

		loginReq := models.LoginRequest{
			Email:    "test@example.com",
			Password: "wrongpassword",
		}

		// Create user with different password
		user := &models.User{Email: loginReq.Email}
		err = user.SetPassword("correctpassword")
		require.NoError(t, err)

		// Mock user query
		rows := sqlmock.NewRows([]string{
			"id", "login", "email", "pw", "title", "first_name", "last_name",
			"valid_id", "create_time", "create_by", "change_time", "change_by",
		}).AddRow(
			1, "testuser", loginReq.Email, user.Password, "Mr", "Test", "User",
			1, time.Now(), 1, time.Now(), 1,
		)

		mock.ExpectQuery("SELECT (.+) FROM users WHERE email = \\$1").
			WithArgs(loginReq.Email).
			WillReturnRows(rows)

		// Mock role determination
		mock.ExpectQuery("SELECT g.name FROM groups g").
			WithArgs(uint(1)).
			WillReturnError(sql.ErrNoRows)

		// Mock failed login update
		mock.ExpectExec("UPDATE users SET (.+)").
			WillReturnResult(sqlmock.NewResult(1, 1))

		body, _ := json.Marshal(loginReq)
		req := httptest.NewRequest("POST", "/auth/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req

		handler.Login(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid email or password")
	})

	t.Run("RefreshToken with valid token", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		jwtManager := auth.NewJWTManager("test-secret", 1*time.Hour)
		authService := auth.NewAuthService(db, jwtManager)
		handler := NewAuthHandler(authService)

		// Generate a valid refresh token
		refreshToken, err := jwtManager.GenerateRefreshToken(1, "test@example.com")
		require.NoError(t, err)

		refreshReq := models.RefreshTokenRequest{
			RefreshToken: refreshToken,
		}

		// Mock user query
		rows := sqlmock.NewRows([]string{
			"id", "login", "email", "pw", "title", "first_name", "last_name",
			"valid_id", "create_time", "create_by", "change_time", "change_by",
		}).AddRow(
			1, "testuser", "test@example.com", "hashedpw", "Mr", "Test", "User",
			1, time.Now(), 1, time.Now(), 1,
		)

		mock.ExpectQuery("SELECT (.+) FROM users WHERE email = \\$1 AND valid_id = 1").
			WithArgs("test@example.com").
			WillReturnRows(rows)

		// Mock role determination
		mock.ExpectQuery("SELECT g.name FROM groups g").
			WithArgs(uint(1)).
			WillReturnError(sql.ErrNoRows)

		body, _ := json.Marshal(refreshReq)
		req := httptest.NewRequest("POST", "/auth/refresh", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req

		handler.RefreshToken(c)

		// Debug: print response body if not OK
		if w.Code != http.StatusOK {
			t.Logf("Response body: %s", w.Body.String())
		}

		assert.Equal(t, http.StatusOK, w.Code)

		var responseWrapper struct {
			Success bool                 `json:"success"`
			Data    models.LoginResponse `json:"data"`
		}
		err = json.Unmarshal(w.Body.Bytes(), &responseWrapper)
		require.NoError(t, err)

		assert.True(t, responseWrapper.Success)
		assert.NotEmpty(t, responseWrapper.Data.Token)
		assert.NotEmpty(t, responseWrapper.Data.RefreshToken)
	})

	t.Run("RefreshToken with invalid token", func(t *testing.T) {
		db, _, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		jwtManager := auth.NewJWTManager("test-secret", 1*time.Hour)
		authService := auth.NewAuthService(db, jwtManager)
		handler := NewAuthHandler(authService)

		refreshReq := models.RefreshTokenRequest{
			RefreshToken: "invalid.token.here",
		}

		body, _ := json.Marshal(refreshReq)
		req := httptest.NewRequest("POST", "/auth/refresh", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req

		handler.RefreshToken(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid or expired refresh token")
	})

	t.Run("ChangePassword with valid request", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		jwtManager := auth.NewJWTManager("test-secret", 1*time.Hour)
		authService := auth.NewAuthService(db, jwtManager)
		handler := NewAuthHandler(authService)

		userID := uint(1)
		oldPassword := "oldpassword"
		newPassword := "newpassword123"

		changePassReq := models.ChangePasswordRequest{
			OldPassword: oldPassword,
			NewPassword: newPassword,
		}

		// Create user with old password
		user := &models.User{ID: userID}
		err = user.SetPassword(oldPassword)
		require.NoError(t, err)

		// Mock user query
		rows := sqlmock.NewRows([]string{
			"id", "login", "email", "pw", "title", "first_name", "last_name",
			"valid_id", "create_time", "create_by", "change_time", "change_by",
		}).AddRow(
			userID, "testuser", "test@example.com", user.Password, "Mr", "Test", "User",
			1, time.Now(), 1, time.Now(), 1,
		)

		mock.ExpectQuery("SELECT (.+) FROM users WHERE id = \\$1").
			WithArgs(userID).
			WillReturnRows(rows)

		// Mock role determination
		mock.ExpectQuery("SELECT g.name FROM groups g").
			WithArgs(userID).
			WillReturnError(sql.ErrNoRows)

		// Mock password update
		mock.ExpectExec("UPDATE users SET pw = \\$1, change_time = \\$2 WHERE id = \\$3").
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), userID).
			WillReturnResult(sqlmock.NewResult(1, 1))

		body, _ := json.Marshal(changePassReq)
		req := httptest.NewRequest("POST", "/auth/change-password", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("user_id", userID)

		handler.ChangePassword(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Password changed successfully")
	})

	t.Run("ChangePassword without authentication", func(t *testing.T) {
		db, _, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		jwtManager := auth.NewJWTManager("test-secret", 1*time.Hour)
		authService := auth.NewAuthService(db, jwtManager)
		handler := NewAuthHandler(authService)

		changePassReq := models.ChangePasswordRequest{
			OldPassword: "oldpass",
			NewPassword: "newpass123",
		}

		body, _ := json.Marshal(changePassReq)
		req := httptest.NewRequest("POST", "/auth/change-password", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req
		// No user_id set in context

		handler.ChangePassword(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "User not authenticated")
	})

	t.Run("Logout", func(t *testing.T) {
		db, _, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		jwtManager := auth.NewJWTManager("test-secret", 1*time.Hour)
		authService := auth.NewAuthService(db, jwtManager)
		handler := NewAuthHandler(authService)

		req := httptest.NewRequest("POST", "/auth/logout", nil)
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req

		handler.Logout(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Logged out successfully")
	})

	t.Run("GetCurrentUser with authentication", func(t *testing.T) {
		db, _, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		jwtManager := auth.NewJWTManager("test-secret", 1*time.Hour)
		authService := auth.NewAuthService(db, jwtManager)
		handler := NewAuthHandler(authService)

		req := httptest.NewRequest("GET", "/auth/me", nil)
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("user_id", uint(1))
		c.Set("user_email", "test@example.com")
		c.Set("user_role", "Agent")

		handler.GetCurrentUser(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, float64(1), response["id"])
		assert.Equal(t, "test@example.com", response["email"])
		assert.Equal(t, "Agent", response["role"])
	})

	t.Run("GetCurrentUser without authentication", func(t *testing.T) {
		db, _, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		jwtManager := auth.NewJWTManager("test-secret", 1*time.Hour)
		authService := auth.NewAuthService(db, jwtManager)
		handler := NewAuthHandler(authService)

		req := httptest.NewRequest("GET", "/auth/me", nil)
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req
		// No user context set

		handler.GetCurrentUser(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "User not authenticated")
	})
}

func TestNewAuthHandler(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	jwtManager := auth.NewJWTManager("secret", 1*time.Hour)
	authService := auth.NewAuthService(db, jwtManager)

	handler := NewAuthHandler(authService)
	assert.NotNil(t, handler)
	assert.NotNil(t, handler.authService)
}

func BenchmarkLogin(b *testing.B) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	jwtManager := auth.NewJWTManager("test-secret", 1*time.Hour)
	authService := auth.NewAuthService(db, jwtManager)
	handler := NewAuthHandler(authService)

	loginReq := models.LoginRequest{
		Email:    "bench@example.com",
		Password: "password123",
	}

	user := &models.User{Email: loginReq.Email}
	user.SetPassword(loginReq.Password)

	body, _ := json.Marshal(loginReq)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rows := sqlmock.NewRows([]string{
			"id", "login", "email", "pw", "title", "first_name", "last_name",
			"valid_id", "create_time", "create_by", "change_time", "change_by",
		}).AddRow(
			1, "testuser", loginReq.Email, user.Password, "Mr", "Test", "User",
			1, time.Now(), 1, time.Now(), 1,
		)

		mock.ExpectQuery("SELECT (.+)").WillReturnRows(rows)
		mock.ExpectQuery("SELECT g.name").WillReturnError(sql.ErrNoRows)
		mock.ExpectExec("UPDATE (.+)").WillReturnResult(sqlmock.NewResult(1, 1))

		req := httptest.NewRequest("POST", "/auth/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req

		handler.Login(c)
	}
}