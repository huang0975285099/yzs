package handlers

import (
	"go-yzs/database"
	"go-yzs/models"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// SaveHandledGoodsRequest 保存处理商品请求
type SaveHandledGoodsRequest struct {
	TradeID    uint              `json:"tradeId"`
	OutOrderNo string            `json:"outOrderNo"`
	GoodsList  []GoodsItem       `json:"goodsList"`
	Duration   int               `json:"duration"`
	Remark     string            `json:"remark"`
}

// SaveHandledGoods 保存用户处理的商品记录
func SaveHandledGoods(c *gin.Context) {
	user := c.MustGet("user").(*models.User)

	var req SaveHandledGoodsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数格式错误"})
		return
	}

	if len(req.GoodsList) == 0 {
		c.JSON(http.StatusOK, gin.H{"code": 200, "message": "无商品记录"})
		return
	}

	now := time.Now()
	records := make([]models.HandledGoods, 0, len(req.GoodsList))

	for _, goods := range req.GoodsList {
		record := models.HandledGoods{
			UserID:     user.ID,
			Username:   user.Username,
			Realname:   user.Realname,
			TradeID:    req.TradeID,
			OutOrderNo: req.OutOrderNo,
			GoodsID:    goods.GoodsID,
			GoodsName:  goods.GoodsName,
			GoodsPrice: goods.GoodsPrice,
			GoodsImage: goods.GoodsImage,
			Type:       goods.Type,
			GoodsCount: goods.GoodsCount,
			Duration:   req.Duration,
			Remark:     req.Remark,
			CreatedAt:  now,
		}
		records = append(records, record)
	}

	if err := database.DB.Create(&records).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "保存失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "保存成功", "data": gin.H{"count": len(records)}})
}

// AggregatedGoods 聚合后的商品记录（按 goodsId+type 累计数量）
type AggregatedGoods struct {
	GoodsID    int64     `json:"goodsId"`
	GoodsName  string    `json:"goodsName"`
	GoodsPrice float64   `json:"goodsPrice"`
	GoodsImage string    `json:"goodsImage"`
	Type       int       `json:"type"`
	GoodsCount int       `json:"goodsCount"`
	CreatedAt  time.Time `json:"createdAt"`
}

// ListMyHandledGoods 查询当前用户处理的商品列表（按商品累计数量，分页）
func ListMyHandledGoods(c *gin.Context) {
	user := c.MustGet("user").(*models.User)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}
	offset := (page - 1) * size

	var total int64
	database.DB.Raw(
		"SELECT COUNT(*) FROM (SELECT goods_id FROM handled_goods WHERE user_id = ? GROUP BY goods_id, type) AS t",
		user.ID,
	).Scan(&total)

	records := make([]AggregatedGoods, 0)
	if err := database.DB.Model(&models.HandledGoods{}).
		Select("goods_id, goods_name, goods_price, goods_image, type, SUM(goods_count) AS goods_count, MAX(created_at) AS created_at").
		Where("user_id = ?", user.ID).
		Group("goods_id, goods_name, goods_price, goods_image, type").
		Order("MAX(created_at) DESC").
		Offset(offset).
		Limit(size).
		Scan(&records).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"records": records,
			"total":   total,
			"page":    page,
			"size":    size,
		},
	})
}