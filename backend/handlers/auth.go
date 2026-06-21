package handlers

import (
	"context"
	"go-yzs/config"
	"go-yzs/database"
	"go-yzs/middleware"
	"go-yzs/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type LoginRequest struct {
	Username  string `json:"username" binding:"required"`
	Password  string `json:"password" binding:"required"`
	DeviceKey string `json:"deviceKey"`
}

func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	var user models.User
	if err := database.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "用户名或密码错误"})
		return
	}

	if !user.CheckPassword(req.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "用户名或密码错误"})
		return
	}

	expiredAt := time.Now().Add(24 * time.Hour)
	claims := middleware.Claims{
		UserID: user.ID,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiredAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(config.App.JWTSecret))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "生成 Token 失败"})
		return
	}

	ctx := context.Background()

	// SSO：先踢出 Redis 中旧 session，再删除 MySQL 旧 session
	database.DeleteUserSession(ctx, user.ID)
	database.DB.Where("user_id = ?", user.ID).Delete(&models.UserSession{})

	// 写入新 session 到 MySQL
	session := models.UserSession{
		UserID:    user.ID,
		Token:     tokenStr,
		DeviceKey: req.DeviceKey,
		ExpiredAt: expiredAt,
	}
	database.DB.Create(&session)

	// 写入新 session 到 Redis
	_ = database.SetSession(ctx, tokenStr, database.CachedSession{
		SessionID: session.ID,
		UserID:    user.ID,
		Username:  user.Username,
		Realname:  user.Realname,
		Role:      user.Role,
		ExpiredAt: expiredAt,
	})

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "登录成功",
		"data": gin.H{
			"token": tokenStr,
			"user": gin.H{
				"id":            user.ID,
				"username":      user.Username,
				"realname":      user.Realname,
				"role":          user.Role,
				"reviewEnabled": config.App.ReviewEnabled,
			},
		},
	})
}

func Logout(c *gin.Context) {
	session, _ := c.Get("session")
	s := session.(*models.UserSession)

	// 同时删除 Redis 和 MySQL
	database.DeleteSession(context.Background(), s.Token, s.UserID)
	database.DB.Delete(s)

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "已退出登录"})
}

func GetCurrentUser(c *gin.Context) {
	user, _ := c.Get("user")
	u := user.(*models.User)
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"id":            u.ID,
			"username":      u.Username,
			"realname":      u.Realname,
			"role":          u.Role,
			"reviewEnabled": config.App.ReviewEnabled,
		},
	})
}

type ChangePasswordRequest struct {
	OldPassword string `json:"oldPassword" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required,min=6"`
}

func ChangePassword(c *gin.Context) {
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	user, _ := c.Get("user")
	u := user.(*models.User)

	// 重新从数据库查询用户完整信息（包含密码字段）
	var fullUser models.User
	if err := database.DB.First(&fullUser, u.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "用户不存在"})
		return
	}

	// 验证旧密码
	if !fullUser.CheckPassword(req.OldPassword) {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "原密码错误"})
		return
	}

	// 设置新密码
	if err := fullUser.SetPassword(req.NewPassword); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "密码加密失败"})
		return
	}

	// 保存到数据库
	if err := database.DB.Save(&fullUser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "保存失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "密码修改成功"})
}
