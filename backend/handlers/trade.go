package handlers

import (
	"go-yzs/database"
	"go-yzs/models"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func ListTradeAbnormal(c *gin.Context) {
	page := c.DefaultQuery("page", "1")
	size := c.DefaultQuery("size", "20")
	keyword := c.Query("keyword")
	isHandled := c.Query("isHandled") // "true" | "false" | ""
	log.Printf("[ListTradeAbnormal] isHandled=%q", isHandled)
	abnormalType := c.Query("abnormalTypeDesc")
	startTime := c.Query("startDate")
	endTime := c.Query("endDate")

	var records []models.TradeAbnormal
	var total int64

	query := database.DB.Model(&models.TradeAbnormal{})

	if keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("node_name LIKE ? OR inner_code LIKE ? OR out_order_no LIKE ? OR transaction_id LIKE ?",
			like, like, like, like)
	}
	if isHandled == "true" {
		query = query.Where("is_handled = 1 AND handled_by_name != '外部系统'")
	} else if isHandled == "false" {
		query = query.Where("is_handled = 0 AND review_status = ''")
	} else if isHandled == "pending" {
		query = query.Where("review_status = 'pending'")
	}
	if abnormalType != "" {
		query = query.Where("abnormal_type_desc = ?", abnormalType)
	}
	if startTime != "" {
		query = query.Where("create_time >= ?", startTime+" 00:00:00")
	}
	if endTime != "" {
		query = query.Where("create_time <= ?", endTime+" 23:59:59")
	}

	query.Count(&total)

	// Convert page/size to int
	var pageInt, sizeInt int
	if _, err := parseIntParam(page, &pageInt); err != nil || pageInt < 1 {
		pageInt = 1
	}
	if _, err := parseIntParam(size, &sizeInt); err != nil || sizeInt < 1 {
		sizeInt = 20
	}

	offset := (pageInt - 1) * sizeInt
	query.Order("create_time DESC").Offset(offset).Limit(sizeInt).Find(&records)

	var lastSyncTime *time.Time
	database.DB.Model(&models.TradeAbnormal{}).Select("MAX(synced_at)").Scan(&lastSyncTime)

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"records":      records,
			"total":        total,
			"page":         pageInt,
			"size":         sizeInt,
			"pages":        (int(total) + sizeInt - 1) / sizeInt,
			"lastSyncTime": lastSyncTime,
		},
	})
}

func GetHourlyStats(c *gin.Context) {
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")
	keyword := c.Query("keyword")
	isHandled := c.Query("isHandled")

	query := database.DB.Model(&models.TradeAbnormal{})

	if keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("node_name LIKE ? OR inner_code LIKE ? OR out_order_no LIKE ? OR transaction_id LIKE ?",
			like, like, like, like)
	}
	if isHandled == "true" {
		query = query.Where("is_handled = 1 AND handled_by_name != '外部系统'")
	} else if isHandled == "false" {
		query = query.Where("is_handled = 0 AND review_status = ''")
	} else if isHandled == "pending" {
		query = query.Where("review_status = 'pending'")
	}

	if startDate != "" {
		query = query.Where("create_time >= ?", startDate+" 00:00:00")
	}
	if endDate != "" {
		query = query.Where("create_time <= ?", endDate+" 23:59:59")
	}

	type HourlyCount struct {
		Hour  int `json:"hour"`
		Count int `json:"count"`
	}
	var results []HourlyCount

	// Group by hour
	query.Select("HOUR(create_time) as hour, COUNT(*) as count").
		Group("HOUR(create_time)").
		Order("hour ASC").
		Scan(&results)

	// Ensure 24 hours are represented
	hoursMap := make(map[int]int)
	for _, r := range results {
		hoursMap[r.Hour] = r.Count
	}

	var finalResults []HourlyCount
	for i := 0; i < 24; i++ {
		finalResults = append(finalResults, HourlyCount{
			Hour:  i,
			Count: hoursMap[i],
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": finalResults,
	})
}

func parseIntParam(s string, v *int) (int, error) {
	n := 0
	_, err := parseIntStr(s, &n)
	*v = n
	return n, err
}

func parseIntStr(s string, v *int) (int, error) {
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			*v = n
			return n, nil
		}
		n = n*10 + int(c-'0')
	}
	*v = n
	return n, nil
}
