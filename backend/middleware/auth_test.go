package middleware

import (
	"go-yzs/models"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRolesRequired(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name       string
		user       *models.User
		wantStatus int
	}{
		{name: "administrator allowed", user: &models.User{Role: "admin"}, wantStatus: http.StatusNoContent},
		{name: "statistician allowed", user: &models.User{Role: "statistician"}, wantStatus: http.StatusNoContent},
		{name: "inspector allowed", user: &models.User{Role: "inspector"}, wantStatus: http.StatusNoContent},
		{name: "operator forbidden", user: &models.User{Role: "operator"}, wantStatus: http.StatusForbidden},
		{name: "missing user unauthorized", wantStatus: http.StatusUnauthorized},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.GET("/stats", func(c *gin.Context) {
				if tt.user != nil {
					c.Set("user", tt.user)
				}
				c.Next()
			}, RolesRequired("admin", "statistician", "inspector"), func(c *gin.Context) {
				c.Status(http.StatusNoContent)
			})

			response := httptest.NewRecorder()
			router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/stats", nil))
			if response.Code != tt.wantStatus {
				t.Fatalf("want status %d, got %d", tt.wantStatus, response.Code)
			}
		})
	}
}
