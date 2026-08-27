package handlers

import (
	"go-yzs/database"
	"go-yzs/models"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func ListTradeAbnormal(c *gin.Context) {
	page := c.DefaultQuery("page", "1")
	size := c.DefaultQuery("size", "20")
	keyword := c.Query("keyword")
	isHandled := c.Query("isHandled") // "true" | "false" | ""
	abnormalType := c.Query("abnormalTypeDesc")
	startTime := c.Query("startDate")
	endTime := c.Query("endDate")

	var records []models.TradeAbnormal
	var total int64

	startBound, endBound := "", ""
	if startTime != "" {
		start, err := parseShanghaiDate(startTime)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "开始日期格式错误"})
			return
		}
		startBound = start.Format("2006-01-02 15:04:05")
	}
	if endTime != "" {
		end, err := parseShanghaiDate(endTime)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "结束日期格式错误"})
			return
		}
		endBound = end.AddDate(0, 0, 1).Format("2006-01-02 15:04:05")
	}
	buildQuery := func() *gorm.DB {
		query := database.DB.Model(&models.TradeAbnormal{})
		if keyword != "" {
			like := "%" + keyword + "%"
			query = query.Where("node_name LIKE ? OR inner_code LIKE ? OR out_order_no LIKE ? OR transaction_id LIKE ?",
				like, like, like, like)
		}
		switch isHandled {
		case "true":
			query = query.Where("is_handled = 1 AND handled_by_name != '外部系统'")
		case "false":
			query = query.Where("is_handled = 0 AND review_status = ''")
		case "pending":
			query = query.Where("review_status = 'pending'")
		}
		if abnormalType != "" {
			query = query.Where("abnormal_type_desc = ?", abnormalType)
		}
		if startBound != "" {
			query = query.Where("create_time >= ?", startBound)
		}
		if endBound != "" {
			query = query.Where("create_time < ?", endBound)
		}
		return query
	}

	// Convert page/size to int
	var pageInt, sizeInt int
	if _, err := parseIntParam(page, &pageInt); err != nil || pageInt < 1 {
		pageInt = 1
	}
	if _, err := parseIntParam(size, &sizeInt); err != nil || sizeInt < 1 {
		sizeInt = 20
	}

	offset := (pageInt - 1) * sizeInt
	var lastSync struct {
		SyncedAt *time.Time
	}
	var queryErrors [3]error
	var wg sync.WaitGroup
	wg.Add(3)
	go func() {
		defer wg.Done()
		queryErrors[0] = buildQuery().Count(&total).Error
	}()
	go func() {
		defer wg.Done()
		queryErrors[1] = buildQuery().Order("create_time DESC").Offset(offset).Limit(sizeInt).Find(&records).Error
	}()
	go func() {
		defer wg.Done()
		queryErrors[2] = database.DB.Model(&models.TradeAbnormal{}).
			Select("synced_at").Where("synced_at IS NOT NULL").
			Order("synced_at DESC").Limit(1).Scan(&lastSync).Error
	}()
	wg.Wait()
	for _, err := range queryErrors {
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询订单失败"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"records":      records,
			"total":        total,
			"page":         pageInt,
			"size":         sizeInt,
			"pages":        (int(total) + sizeInt - 1) / sizeInt,
			"lastSyncTime": lastSync.SyncedAt,
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
		start, err := parseShanghaiDate(startDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "开始日期格式错误"})
			return
		}
		query = query.Where("create_time >= ?", start.Format("2006-01-02 15:04:05"))
	}
	if endDate != "" {
		end, err := parseShanghaiDate(endDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "结束日期格式错误"})
			return
		}
		query = query.Where("create_time < ?", end.AddDate(0, 0, 1).Format("2006-01-02 15:04:05"))
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
