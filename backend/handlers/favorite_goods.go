package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"go-yzs/database"
	"go-yzs/models"
)

// AddFavoriteGoods 添加收藏
func AddFavoriteGoods(c *gin.Context) {
	user := c.MustGet("user").(*models.User)

	var req struct {
		GoodsID  uint64 `json:"goodsId" binding:"required"`
		Title    string `json:"title" binding:"required"`
		Sn       string `json:"sn" binding:"required"`
		FrontImg string `json:"frontImg"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	// 检查是否已收藏
	var existing models.FavoriteGoods
	if database.DB.Where("user_id = ? AND goods_id = ?", user.ID, req.GoodsID).First(&existing).Error == nil {
		c.JSON(http.StatusOK, gin.H{"code": 200, "message": "已收藏", "data": gin.H{"favoriteId": existing.ID}})
		return
	}

	favorite := models.FavoriteGoods{
		UserID:   uint64(user.ID),
		GoodsID:  req.GoodsID,
		Title:    req.Title,
		Sn:       req.Sn,
		FrontImg: req.FrontImg,
	}
	if err := database.DB.Create(&favorite).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "收藏失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "收藏成功", "data": gin.H{"favoriteId": favorite.ID}})
}

// RemoveFavoriteGoods 取消收藏
func RemoveFavoriteGoods(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	goodsID := c.Param("goodsId")

	if err := database.DB.Where("user_id = ? AND goods_id = ?", user.ID, goodsID).Delete(&models.FavoriteGoods{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "取消收藏失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "已取消收藏"})
}

// ListFavoriteGoods 获取收藏列表
func ListFavoriteGoods(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}

	var total int64
	database.DB.Model(&models.FavoriteGoods{}).Where("user_id = ?", user.ID).Count(&total)

	var favorites []models.FavoriteGoods
	offset := (page - 1) * size
	if err := database.DB.Where("user_id = ?", user.ID).Order("created_at DESC").Offset(offset).Limit(size).Find(&favorites).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"records": favorites,
			"total":   total,
			"page":    page,
			"size":    size,
		},
	})
}

// CheckFavoriteGoods 批量检查是否已收藏
func CheckFavoriteGoods(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	goodsIDs := c.Query("goodsIds")

	var ids []uint64
	for _, idStr := range splitGoodsIDs(goodsIDs) {
		ids = append(ids, parseUint64(idStr))
	}

	if len(ids) == 0 {
		c.JSON(http.StatusOK, gin.H{"code": 200, "data": gin.H{"favorites": []uint64{}}})
		return
	}

	var favorites []models.FavoriteGoods
	database.DB.Where("user_id = ? AND goods_id IN ?", user.ID, ids).Find(&favorites)

	favoriteIDs := make(map[uint64]bool)
	for _, f := range favorites {
		favoriteIDs[f.GoodsID] = true
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": gin.H{"favorites": favoriteIDs}})
}

func splitGoodsIDs(s string) []string {
	if s == "" {
		return []string{}
	}
	result := []string{}
	for _, id := range splitByComma(s) {
		if id != "" {
			result = append(result, id)
		}
	}
	return result
}

func splitByComma(s string) []string {
	result := []string{}
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == ',' {
			result = append(result, s[start:i])
			start = i + 1
		}
	}
	result = append(result, s[start:])
	return result
}

func parseUint64(s string) uint64 {
	var result uint64
	for _, c := range s {
		if c >= '0' && c <= '9' {
			result = result*10 + uint64(c-'0')
		}
	}
	return result
}
