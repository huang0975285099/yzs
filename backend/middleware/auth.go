package middleware

import (
	"context"
	"go-yzs/config"
	"go-yzs/database"
	"go-yzs/models"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c)
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未授权，请先登录"})
			c.Abort()
			return
		}

		ctx := context.Background()

		// 优先从 Redis 读取缓存 session
		if cached := database.GetSession(ctx, token); cached != nil {
			// 缓存命中：无需查 MySQL，直接构造 user 和 session 对象
			user := &models.User{
				ID:       cached.UserID,
				Username: cached.Username,
				Realname: cached.Realname,
				Role:     cached.Role,
			}
			session := &models.UserSession{
				ID:        cached.SessionID,
				UserID:    cached.UserID,
				Token:     token,
				ExpiredAt: cached.ExpiredAt,
			}
			c.Set("user", user)
			c.Set("session", session)
			c.Next()
			return
		}

		// 缓存未命中：回落到 MySQL 验证
		var session models.UserSession
		if err := database.DB.Where("token = ? AND expired_at > ?", token, time.Now()).First(&session).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "会话已失效，请重新登录"})
			c.Abort()
			return
		}

		claims, err := parseToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "Token 无效"})
			c.Abort()
			return
		}

		var user models.User
		if err := database.DB.First(&user, claims.UserID).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "用户不存在"})
			c.Abort()
			return
		}

		// 回填 Redis，下次命中缓存
		_ = database.SetSession(ctx, token, database.CachedSession{
			SessionID: session.ID,
			UserID:    user.ID,
			Username:  user.Username,
			Realname:  user.Realname,
			Role:      user.Role,
			ExpiredAt: session.ExpiredAt,
		})

		c.Set("user", &user)
		c.Set("session", &session)
		c.Next()
	}
}

func AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, _ := c.Get("user")
		u := user.(*models.User)
		if u.Role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "无权限，需要管理员角色"})
			c.Abort()
			return
		}
		c.Next()
	}
}

func extractToken(c *gin.Context) string {
	auth := c.GetHeader("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}
	return ""
}

type Claims struct {
	UserID uint   `json:"userId"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func parseToken(tokenStr string) (*Claims, error) {
	secret := []byte(config.App.JWTSecret)
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		return secret, nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, jwt.ErrTokenInvalidClaims
}
