package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"go-yzs/database"
	"go-yzs/models"
)

// ListGoods 获取商品列表（分页）
func ListGoods(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}
	keyword := c.Query("keyword")

	var total int64
	query := database.DB.Model(&models.Goods{})
	if keyword != "" {
		query = query.Where("title LIKE ?", "%"+keyword+"%")
	}
	query.Count(&total)

	var goods []models.Goods
	offset := (page - 1) * size
	if err := query.Offset(offset).Limit(size).Order("id DESC").Find(&goods).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"records": goods,
			"total":   total,
			"page":    page,
			"size":    size,
		},
	})
}
