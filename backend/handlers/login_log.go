package handlers

import (
	"go-yzs/database"
	"go-yzs/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ListLoginLogs 查询登录日志
// 查询参数：
//   - userId (可选) 按用户ID筛选
//   - username (可选) 按用户名筛选（模糊匹配）
//   - ip (可选) 按IP筛选（模糊匹配）
//   - status (可选) success / failed
//   - page (默认1)
//   - pageSize (默认20，最大100)
// 管理员可查看所有；普通用户只能查看自己的日志（强制按自身 userId 过滤）
func ListLoginLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	// 鉴权：非管理员强制只看自己
	user, _ := c.Get("user")
	u := user.(*models.User)
	isAdmin := u.Role == "admin"

	query := database.DB.Model(&models.LoginLog{})

	if !isAdmin {
		query = query.Where("user_id = ?", u.ID)
	} else if uidStr := c.Query("userId"); uidStr != "" {
		if uid, err := strconv.ParseUint(uidStr, 10, 64); err == nil {
			query = query.Where("user_id = ?", uint(uid))
		}
	}

	if username := c.Query("username"); username != "" {
		query = query.Where("username LIKE ?", "%"+username+"%")
	}
	if ip := c.Query("ip"); ip != "" {
		query = query.Where("ip LIKE ?", "%"+ip+"%")
	}
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	var total int64
	query.Count(&total)

	var logs []models.LoginLog
	if err := query.Order("login_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&logs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"list":     logs,
			"total":    total,
			"page":     page,
			"pageSize": pageSize,
		},
	})
}
