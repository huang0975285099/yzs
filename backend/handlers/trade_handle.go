package handlers

import (
	"bytes"
	"encoding/json"
	"go-yzs/config"
	"go-yzs/database"
	"go-yzs/models"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

const externalBaseURL = "https://api.uboxol.com/lotus/trade/abnormal"

type externalResult struct {
	Code    int    `json:"code"`
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// callExternalAPI 调用外部 API，返回响应结果
func callExternalAPI(path string, payload any) (*externalResult, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Post(externalBaseURL+path, "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result externalResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

// checkExternalHandled 检查订单在外部系统是否已不在未处理列表
// 返回 (alreadyHandled bool, err error)
func checkExternalHandled(trade *models.TradeAbnormal) (bool, error) {
	startTime := trade.CreateTime
	if startTime == "" {
		startTime = time.Now().AddDate(0, 0, -30).Format("2006-01-02") + " 00:00:00"
	}

	body, _ := json.Marshal(map[string]any{
		"operatingModeList": []int{21},
		"handleStatus":      "NOT_HANDLED",
		"pendStatus":        "NO_PENDING",
		"outOrderNo":        trade.OutOrderNo,
		"current":           1,
		"size":              5,
		"startCreateTime":   startTime,
		"endCreateTime":     time.Now().Format("2006-01-02") + " 23:59:59",
	})

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Post(externalBaseURL+"/page", "application/json", bytes.NewReader(body))
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	var pageResp struct {
		Success bool `json:"success"`
		Data    struct {
			Total int `json:"total"`
		} `json:"data"`
	}
	json.NewDecoder(resp.Body).Decode(&pageResp)

	if pageResp.Success && pageResp.Data.Total > 0 {
		return false, nil // 仍在未处理列表
	}
	return true, nil // 不在未处理列表，已被外部处理
}

// CheckTradeStatus 在打开处理表单前，先查询外部系统确认订单是否仍为未处理状态
func CheckTradeStatus(c *gin.Context) {
	id := c.Param("id")

	var trade models.TradeAbnormal
	if err := database.DB.First(&trade, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "记录不存在"})
		return
	}
	if trade.IsHandled {
		c.JSON(http.StatusOK, gin.H{"code": 200, "data": gin.H{"alreadyHandled": true, "message": "该订单已处理"}})
		return
	}

	alreadyHandled, err := checkExternalHandled(&trade)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"code": 503, "message": "网络异常，无法确认订单状态，请联系管理人员"})
		return
	}

	if !alreadyHandled {
		c.JSON(http.StatusOK, gin.H{"code": 200, "data": gin.H{"alreadyHandled": false}})
		return
	}

	// 外部已处理，更新本地
	now := time.Now()
	database.DB.Model(&trade).Updates(map[string]any{
		"is_handled":         true,
		"handled_by_id":      nil,
		"handled_by_name":    "外部系统",
		"handled_at":         now,
		"handle_status_desc": "客服已处理",
		"handle_source":      "external",
		"review_status":      "",
	})
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{"alreadyHandled": true, "message": "该订单已由外部系统处理，已从待处理列表移除"},
	})
}

type PendRequest struct {
	Duration int    `json:"duration"`
	Remark   string `json:"remark"`
}

func PendTrade(c *gin.Context) {
	id := c.Param("id")
	user := c.MustGet("user").(*models.User)

	var pendReq PendRequest
	_ = c.ShouldBindJSON(&pendReq)

	var trade models.TradeAbnormal
	if err := database.DB.First(&trade, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "记录不存在"})
		return
	}
	if trade.IsHandled {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "该订单已处理，无法挂起"})
		return
	}
	if trade.ReviewStatus == "pending" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "该订单已提交质检，请等待审核"})
		return
	}

	handlerName := user.Realname
	if handlerName == "" {
		handlerName = user.Username
	}

	if config.App.ReviewEnabled {
		// 审核模式：存入质检队列
		review := models.TradeReview{
			TradeID:         trade.ID,
			ActionType:      "pend",
			OperatorRemark:  pendReq.Remark,
			Duration:        pendReq.Duration,
			SubmittedByID:   user.ID,
			SubmittedByName: handlerName,
			SubmittedAt:     time.Now(),
			ReviewStatus:    "pending",
		}
		if err := database.DB.Create(&review).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "提交失败"})
			return
		}
		now := time.Now()
		database.DB.Model(&trade).Updates(map[string]any{
			"review_status":   "pending",
			"locked_by_id":    nil,
			"locked_at":       nil,
			"handled_by_id":   user.ID,
			"handled_by_name": handlerName,
			"handled_at":      now,
			"handle_duration": pendReq.Duration,
			"handle_remark":   pendReq.Remark,
			"handle_source":   "internal",
		})
		incrementDailyStats(user.ID, "pend")
		c.JSON(http.StatusOK, gin.H{"code": 200, "message": "已提交质检审核"})
		return
	}

	extResult, err := callExternalAPI("/pend", map[string]any{
		"id":         trade.TradeID,
		"pendStatus": "PENDING",
	})
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"code": 502, "message": "外部接口网络异常: " + err.Error()})
		return
	}

	now := time.Now()
	handledByName := handlerName
	if !extResult.Success {
		handledByName = "外部系统"
	}
	database.DB.Model(&trade).Updates(map[string]any{
		"is_handled":       true,
		"handled_by_id":    user.ID,
		"handled_by_name":  handledByName,
		"handled_at":       now,
		"handle_duration":  pendReq.Duration,
		"handle_remark":    pendReq.Remark,
		"pend_status":      "PENDING",
		"pend_status_desc": "已挂起",
		"locked_by_id":     nil,
		"locked_at":        nil,
		"handle_source":    "internal",
	})

	if extResult.Success {
		incrementDailyStats(user.ID, "pend")
		c.JSON(http.StatusOK, gin.H{"code": 200, "message": "挂起成功"})
	} else {
		c.JSON(http.StatusOK, gin.H{"code": 200, "message": "外部系统已处理，已标记完成"})
	}
}

// GoodsItem 提交处理时的商品明细
type GoodsItem struct {
	GoodsID    int64   `json:"goodsId"`
	GoodsName  string  `json:"goodsName"`
	GoodsPrice float64 `json:"goodsPrice"`
	GoodsImage string  `json:"goodsImage"`
	Type       int     `json:"type"`
	GoodsCount int     `json:"goodsCount"`
}

type SubmitRequest struct {
	OrderGoodsDetailList []GoodsItem `json:"orderGoodsDetailList"`
	Duration             int         `json:"duration"`
	Remark               string      `json:"remark"`
}

func SubmitTrade(c *gin.Context) {
	id := c.Param("id")
	user := c.MustGet("user").(*models.User)

	var trade models.TradeAbnormal
	if err := database.DB.First(&trade, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "记录不存在"})
		return
	}
	if trade.IsHandled {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "该订单已处理"})
		return
	}
	if trade.ReviewStatus == "pending" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "该订单已提交质检，请等待审核"})
		return
	}

	var req SubmitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数格式错误"})
		return
	}
	if req.OrderGoodsDetailList == nil {
		req.OrderGoodsDetailList = []GoodsItem{}
	}
	for i := range req.OrderGoodsDetailList {
		if req.OrderGoodsDetailList[i].Type == 0 {
			req.OrderGoodsDetailList[i].Type = 1
		}
	}

	goodsJSON, _ := json.Marshal(req.OrderGoodsDetailList)
	handlerName := user.Realname
	if handlerName == "" {
		handlerName = user.Username
	}

	if config.App.ReviewEnabled {
		// 审核模式：存入质检队列
		review := models.TradeReview{
			TradeID:         trade.ID,
			ActionType:      "submit",
			GoodsJSON:       string(goodsJSON),
			OperatorRemark:  req.Remark,
			Duration:        req.Duration,
			SubmittedByID:   user.ID,
			SubmittedByName: handlerName,
			SubmittedAt:     time.Now(),
			ReviewStatus:    "pending",
		}
		if err := database.DB.Create(&review).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "提交失败"})
			return
		}
		now := time.Now()
		database.DB.Model(&trade).Updates(map[string]any{
			"review_status":   "pending",
			"locked_by_id":    nil,
			"locked_at":       nil,
			"handled_by_id":   user.ID,
			"handled_by_name": handlerName,
			"handled_at":      now,
			"handle_duration": req.Duration,
			"handle_goods":    string(goodsJSON),
			"handle_remark":   req.Remark,
			"handle_source":   "internal",
		})
		incrementDailyStats(user.ID, "submit")
		c.JSON(http.StatusOK, gin.H{"code": 200, "message": "已提交质检审核"})
		return
	}

	// 直通模式：直接调外部 API
	extResult, err := callExternalAPI("/handle", map[string]any{
		"orderGoodsDetailList": req.OrderGoodsDetailList,
		"outOrderNo":           trade.OutOrderNo,
		"handleUsername":       "prisonProject",
	})
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"code": 502, "message": "外部接口网络异常: " + err.Error()})
		return
	}

	now := time.Now()
	handledByName := handlerName
	if !extResult.Success {
		handledByName = "外部系统"
	}
	database.DB.Model(&trade).Updates(map[string]any{
		"is_handled":      true,
		"handled_by_id":   user.ID,
		"handled_by_name": handledByName,
		"handled_at":      now,
		"handle_duration": req.Duration,
		"handle_goods":    string(goodsJSON),
		"handle_remark":   req.Remark,
		"locked_by_id":    nil,
		"locked_at":       nil,
		"handle_source":   "internal",
	})

	if extResult.Success {
		incrementDailyStats(user.ID, "submit")
		c.JSON(http.StatusOK, gin.H{"code": 200, "message": "本订单处理成功"})
	} else {
		c.JSON(http.StatusOK, gin.H{"code": 200, "message": extResult.Message})
	}
}

type HandleRequest struct {
	Remark string `json:"remark"`
}

// HandleTrade 操作员处理一条异常订单（内部标记，无外部 API）
func HandleTrade(c *gin.Context) {
	id := c.Param("id")
	user := c.MustGet("user").(*models.User)

	var trade models.TradeAbnormal
	if err := database.DB.First(&trade, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "记录不存在"})
		return
	}
	if trade.IsHandled {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "该订单已处理"})
		return
	}

	var req HandleRequest
	c.ShouldBindJSON(&req)

	now := time.Now()
	trade.IsHandled = true
	trade.HandledByID = &user.ID
	trade.HandledByName = user.Realname
	if trade.HandledByName == "" {
		trade.HandledByName = user.Username
	}
	trade.HandledAt = &now
	trade.HandleRemark = req.Remark
	trade.HandleSource = "internal"

	database.DB.Save(&trade)
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "处理成功", "data": trade})
}

// ListMyHandled 查询当前操作员的已处理订单 + 待审核订单
func ListMyHandled(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	page, size := 1, 20
	parseIntParam(c.DefaultQuery("page", "1"), &page)
	parseIntParam(c.DefaultQuery("size", "20"), &size)
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 20
	}
	date := c.Query("date")                   // 可选日期过滤，格式 YYYY-MM-DD
	inspectStatus := c.Query("inspectStatus") // '' 全部, 'normal', 'abnormal', 'uninspected'
	filter := operatorRecordFilter{UserIDs: []uint{user.ID}, InspectStatus: inspectStatus}
	if date != "" {
		start, err := parseShanghaiDate(date)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "日期格式错误"})
			return
		}
		filter.StartTime = start.Format("2006-01-02 15:04:05")
		filter.EndTimeExclusive = start.AddDate(0, 0, 1).Format("2006-01-02 15:04:05")
	}

	records, total, err := queryOperatorRecords(filter, page, size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询处理记录失败"})
		return
	}
	daySummary, err := queryOperatorRecordSummary(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询当日统计失败"})
		return
	}
	cumulative, err := queryOperatorRecordSummary(operatorRecordFilter{UserIDs: []uint{user.ID}})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询累计统计失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"records":          records,
			"total":            total,
			"page":             page,
			"size":             size,
			"totalAmount":      daySummary.Amount,
			"cumulativeSubmit": cumulative.SubmitCount,
			"cumulativePend":   cumulative.PendCount,
			"cumulativeAmount": cumulative.Amount,
		},
	})
}

// GetStats 数据统计（用于看板）
func GetStats(c *gin.Context) {
	user := c.MustGet("user").(*models.User)

	shanghaiLoc, _ := time.LoadLocation("Asia/Shanghai")
	now := time.Now().In(shanghaiLoc)
	todayStr := now.Format("2006-01-02")
	thirtyDaysAgo := now.AddDate(0, 0, -30).Format("2006-01-02 00:00:00")
	todayStart := todayStr + " 00:00:00"
	todayEnd := todayStr + " 23:59:59"

	// 并发执行所有独立查询
	var wg sync.WaitGroup

	// 1. 最近30天每日统计（create_time 是字符串，用 LEFT 提取日期避免 DATE() 全表扫描）
	type DayCount struct {
		Day   string `json:"day"`
		Count int    `json:"count"`
	}
	var dailyCounts []DayCount
	wg.Add(1)
	go func() {
		defer wg.Done()
		database.DB.Raw(`
			SELECT LEFT(create_time, 10) as day, COUNT(*) as count
			FROM trade_abnormals
			WHERE create_time >= ?
			GROUP BY LEFT(create_time, 10)
			ORDER BY day ASC
		`, thirtyDaysAgo).Scan(&dailyCounts)
	}()

	// 2. 异常类型统计（限制最近30天，避免全表扫描）
	type TypeCount struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}
	var typeCounts []TypeCount
	wg.Add(1)
	go func() {
		defer wg.Done()
		database.DB.Raw(`
			SELECT abnormal_type_desc as name, COUNT(*) as value
			FROM trade_abnormals
			WHERE create_time >= ?
			GROUP BY abnormal_type_desc
			ORDER BY value DESC
		`, thirtyDaysAgo).Scan(&typeCounts)
	}()

	// 3. 合并 total / handledCount / unhandledCount 为单次查询
	type CountResult struct {
		Total      int64
		Handled    int64
		Unhandled  int64
		TodayCount int64
	}
	var cr CountResult
	wg.Add(1)
	go func() {
		defer wg.Done()
		database.DB.Raw(`
			SELECT
				COUNT(*) as total,
				COALESCE(SUM(CASE WHEN is_handled = 1 THEN 1 ELSE 0 END), 0) as handled,
				COALESCE(SUM(CASE WHEN is_handled = 0 THEN 1 ELSE 0 END), 0) as unhandled,
				COALESCE(SUM(CASE WHEN create_time >= ? AND create_time <= ? THEN 1 ELSE 0 END), 0) as today_count
			FROM trade_abnormals
		`, todayStart, todayEnd).Scan(&cr)
	}()

	// 4. 待质检数量
	var pendingReviewCount int64
	wg.Add(1)
	go func() {
		defer wg.Done()
		database.DB.Model(&models.TradeReview{}).Where("review_status = ?", "pending").Count(&pendingReviewCount)
	}()

	// 5. 操作员累计处理量排行（与操作员个人累计口径一致）
	type OpStat struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}
	var opStats []OpStat
	if user.Role == "admin" || user.Role == "statistician" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			unionSQL, args := buildOperatorRecordUnion(operatorRecordFilter{})
			database.DB.Raw(`SELECT handled_by_name AS name, COUNT(*) AS value
				FROM (`+unionSQL+`) AS events
				GROUP BY handled_by_id, handled_by_name
				ORDER BY value DESC
			`, args...).Scan(&opStats)
		}()
	}

	wg.Wait()

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"total":              cr.Total,
			"todayCount":         cr.TodayCount,
			"handledCount":       cr.Handled,
			"unhandledCount":     cr.Unhandled,
			"pendingReviewCount": pendingReviewCount,
			"dailyCounts":        dailyCounts,
			"typeCounts":         typeCounts,
			"opStats":            opStats,
		},
	})
}
