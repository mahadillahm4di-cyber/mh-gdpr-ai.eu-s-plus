package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/mahadillahm4di-cyber/mh-gdpr-ai.eu-s-plus/internal/auth"
)

func TestAuthRequired_NoHeader(t *testing.T) {
	r := gin.New()
	r.Use(AuthRequired("secret"))
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAuthRequired_InvalidFormat(t *testing.T) {
	r := gin.New()
	r.Use(AuthRequired("secret"))
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "NotBearer token")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAuthRequired_InvalidToken(t *testing.T) {
	r := gin.New()
	r.Use(AuthRequired("secret"))
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid-jwt-token")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAuthRequired_ValidToken(t *testing.T) {
	secret := "test-secret-key-for-jwt-testing!"

	token, err := auth.GenerateAccessToken("user-123", secret)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	var capturedUserID string
	r := gin.New()
	r.Use(AuthRequired(secret))
	r.GET("/test", func(c *gin.Context) {
		capturedUserID = c.GetString("user_id")
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if capturedUserID != "user-123" {
		t.Errorf("expected user_id 'user-123', got %q", capturedUserID)
	}
}

func TestAuthRequired_BearerCaseInsensitive(t *testing.T) {
	secret := "test-secret-key-for-jwt-testing!"
	token, _ := auth.GenerateAccessToken("user-123", secret)

	r := gin.New()
	r.Use(AuthRequired(secret))
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "bearer "+token)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for lowercase 'bearer', got %d", w.Code)
	}
}
