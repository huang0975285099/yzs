package handlers

import (
	"go-yzs/database"
	"go-yzs/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type InspectRequest struct {
	Status string `json:"status" binding:"required,oneof=normal abnormal"`
	Remark string `json:"remark"`
}

// InspectTrade 质检员对已处理订单进行复查（直通模式专用）
func InspectTrade(c *gin.Context) {
	id := c.Param("id")
	user := c.MustGet("user").(*models.User)

	var req InspectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}
	if req.Status == "abnormal" && req.Remark == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "标记异常时备注不能为空"})
		return
	}

	var trade models.TradeAbnormal
	if err := database.DB.First(&trade, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "记录不存在"})
		return
	}

	inspectorName := user.Realname
	if inspectorName == "" {
		inspectorName = user.Username
	}
	now := time.Now()

	database.DB.Model(&trade).Updates(map[string]any{
		"inspect_status":    req.Status,
		"inspect_remark":    req.Remark,
		"inspected_by_id":   user.ID,
		"inspected_by_name": inspectorName,
		"inspected_at":      now,
	})

	msg := "复查完成：正常"
	if req.Status == "abnormal" {
		msg = "复查完成：异常"
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": msg})
}
