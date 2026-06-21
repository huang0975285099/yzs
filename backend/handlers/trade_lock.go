package handlers

import (
	"go-yzs/database"
	"go-yzs/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

const lockTTL = 15 * time.Minute

// LockTrade 原子锁定订单：单条 UPDATE WHERE 避免并发竞态
func LockTrade(c *gin.Context) {
	id := c.Param("id")
	user := c.MustGet("user").(*models.User)

	expiry := time.Now().Add(-lockTTL)
	now := time.Now()

	result := database.DB.Model(&models.TradeAbnormal{}).
		Where(`id = ? AND (
			locked_by_id IS NULL OR
			locked_by_id = ? OR
			locked_at < ?
		)`, id, user.ID, expiry).
		Updates(map[string]any{
			"locked_by_id": user.ID,
			"locked_at":    now,
		})

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "操作失败"})
		return
	}
	if result.RowsAffected == 0 {
		// 0 行受影响：记录不存在 或 被他人锁定
		var exists int64
		database.DB.Model(&models.TradeAbnormal{}).Where("id = ?", id).Count(&exists)
		if exists == 0 {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "记录不存在"})
			return
		}
		c.JSON(http.StatusConflict, gin.H{"code": 409, "message": "该订单已被其他用户锁定"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200})
}

// UnlockTrade 解锁订单（跳过时调用）
func UnlockTrade(c *gin.Context) {
	id := c.Param("id")
	user := c.MustGet("user").(*models.User)

	database.DB.Model(&models.TradeAbnormal{}).
		Where("id = ? AND locked_by_id = ?", id, user.ID).
		Updates(map[string]any{"locked_by_id": nil, "locked_at": nil})
	c.JSON(http.StatusOK, gin.H{"code": 200})
}
