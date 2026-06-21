package handlers

import (
	"bytes"
	"encoding/json"
	"go-yzs/config"
	"go-yzs/database"
	"go-yzs/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
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
			"review_status":    "pending",
			"locked_by_id":     nil,
			"locked_at":        nil,
			"handled_by_id":    user.ID,
			"handled_by_name":  handlerName,
			"handled_at":       now,
			"handle_duration":  pendReq.Duration,
			"handle_remark":    pendReq.Remark,
			"handle_source":    "internal",
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

type MyHandledRecord struct {
	models.TradeAbnormal
	ActionType string `json:"actionType"` // "submit" or "pend"
}

// ListMyHandled 查询当前操作员的已处理订单 + 待审核订单
func ListMyHandled(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	page := c.DefaultQuery("page", "1")
	size := c.DefaultQuery("size", "20")
	date := c.Query("date")                   // 可选日期过滤，格式 YYYY-MM-DD
	inspectStatus := c.Query("inspectStatus") // '' 全部, 'normal', 'abnormal', 'uninspected'

	// 查询该用户待审核的 trade_reviews（含 action_type）
	type pendingReview struct {
		TradeID    uint
		ActionType string
	}
	var pendingReviews []pendingReview
	database.DB.Model(&models.TradeReview{}).
		Select("trade_id, action_type").
		Where("submitted_by_id = ? AND review_status = ?", user.ID, "pending").
		Scan(&pendingReviews)

	pendingTradeIDs := make([]uint, 0, len(pendingReviews))
	pendingActionMap := make(map[uint]string, len(pendingReviews))
	for _, r := range pendingReviews {
		pendingTradeIDs = append(pendingTradeIDs, r.TradeID)
		pendingActionMap[r.TradeID] = r.ActionType
	}

	// baseWhere 封装重复的主条件，避免两处拼写
	baseWhere := func(q *gorm.DB) *gorm.DB {
		if len(pendingTradeIDs) > 0 {
			return q.Where(
				"(is_handled = ? AND handled_by_id = ? AND handled_by_name != ?) OR (review_status = ? AND id IN ?)",
				true, user.ID, "外部系统", "pending", pendingTradeIDs,
			)
		}
		return q.Where("is_handled = ? AND handled_by_id = ? AND handled_by_name != ?", true, user.ID, "外部系统")
	}

	// applyDateFilter 用范围比较代替 DATE() 函数，使 handled_at 索引生效
	applyDateFilter := func(q *gorm.DB) *gorm.DB {
		if date == "" {
			return q
		}
		loc, _ := time.LoadLocation("Asia/Shanghai")
		startOfDay, err := time.ParseInLocation("2006-01-02", date, loc)
		if err != nil {
			return q
		}
		endOfDay := startOfDay.Add(24 * time.Hour)
		return q.Where("handled_at >= ? AND handled_at < ?", startOfDay, endOfDay)
	}

	applyInspectFilter := func(q *gorm.DB) *gorm.DB {
		switch inspectStatus {
		case "normal", "abnormal":
			return q.Where("inspect_status = ?", inspectStatus)
		case "uninspected":
			return q.Where("inspect_status = '' OR inspect_status IS NULL")
		}
		return q
	}

	var trades []models.TradeAbnormal
	var total int64

	query := applyInspectFilter(applyDateFilter(baseWhere(database.DB.Model(&models.TradeAbnormal{}))))
	query.Count(&total)

	var pageInt, sizeInt int
	parseIntParam(page, &pageInt)
	parseIntParam(size, &sizeInt)
	if pageInt < 1 {
		pageInt = 1
	}
	if sizeInt < 1 {
		sizeInt = 20
	}

	offset := (pageInt - 1) * sizeInt
	query.Order("CASE WHEN review_status = 'pending' THEN 0 ELSE 1 END, handled_at DESC").
		Offset(offset).Limit(sizeInt).Find(&trades)

	// 组装 actionType
	records := make([]MyHandledRecord, 0, len(trades))
	for _, t := range trades {
		actionType := "submit"
		if t.ReviewStatus == "pending" {
			if at, ok := pendingActionMap[t.ID]; ok {
				actionType = at
			}
		} else {
			if t.PendStatusDesc != "" && (t.PendStatusDesc == "已挂起" || t.PendStatusDesc == "PENDING") {
				actionType = "pend"
			}
		}
		records = append(records, MyHandledRecord{TradeAbnormal: t, ActionType: actionType})
	}

	// 当日总金额：用 JSON_TABLE 在 SQL 层直接 SUM，无需把 JSON 传回 Go 解析
	var totalAmountResult struct {
		TotalAmount float64 `gorm:"column:total_amount"`
	}
	{
		// 构建与主查询相同的 WHERE 条件
		amountSQL := `
			SELECT COALESCE(SUM(jt.price * jt.cnt), 0) AS total_amount
			FROM trade_abnormals t
			CROSS JOIN JSON_TABLE(
				t.handle_goods,
				'$[*]' COLUMNS (price DOUBLE PATH '$.goodsPrice', cnt INT PATH '$.goodsCount')
			) jt
			WHERE t.handle_goods != '' AND t.handle_goods IS NOT NULL`
		var amountArgs []interface{}
		if len(pendingTradeIDs) > 0 {
			amountSQL += " AND ((t.is_handled = ? AND t.handled_by_id = ? AND t.handled_by_name != ?) OR (t.review_status = ? AND t.id IN (?)))"
			amountArgs = append(amountArgs, true, user.ID, "外部系统", "pending", pendingTradeIDs)
		} else {
			amountSQL += " AND t.is_handled = ? AND t.handled_by_id = ? AND t.handled_by_name != ?"
			amountArgs = append(amountArgs, true, user.ID, "外部系统")
		}
		if date != "" {
			loc, _ := time.LoadLocation("Asia/Shanghai")
			startOfDay, err := time.ParseInLocation("2006-01-02", date, loc)
			if err == nil {
				amountArgs = append(amountArgs, startOfDay, startOfDay.Add(24*time.Hour))
				amountSQL += " AND t.handled_at >= ? AND t.handled_at < ?"
			}
		}
		database.DB.Raw(amountSQL, amountArgs...).Scan(&totalAmountResult)
	}
	totalAmount := totalAmountResult.TotalAmount

	// ===== 累计统计（不按日期过滤）=====
	// JSON_TABLE 子查询：每条 trade 先按 id 聚合出单条金额，外层再统计件数和总金额
	// 三个原始查询（2×COUNT + 1×全量扫描）合并为一次 DB 往返
	type cumulativeResult struct {
		SubmitCount      int64   `gorm:"column:submit_count"`
		PendCount        int64   `gorm:"column:pend_count"`
		CumulativeAmount float64 `gorm:"column:cumulative_amount"`
	}
	var cumResult cumulativeResult
	database.DB.Raw(`
		SELECT
			SUM(CASE WHEN (sub.pend_status_desc IS NULL OR sub.pend_status_desc = '' OR sub.pend_status_desc != '已挂起') THEN 1 ELSE 0 END) AS submit_count,
			SUM(CASE WHEN sub.pend_status_desc = '已挂起'                                                                   THEN 1 ELSE 0 END) AS pend_count,
			COALESCE(SUM(CASE WHEN (sub.pend_status_desc IS NULL OR sub.pend_status_desc = '' OR sub.pend_status_desc != '已挂起') THEN sub.trade_amount ELSE 0 END), 0) AS cumulative_amount
		FROM (
			SELECT
				t.pend_status_desc,
				COALESCE(SUM(jt.price * jt.cnt), 0) AS trade_amount
			FROM trade_abnormals t
			LEFT JOIN JSON_TABLE(
				CASE WHEN t.handle_goods != '' AND t.handle_goods IS NOT NULL THEN t.handle_goods ELSE '[]' END,
				'$[*]' COLUMNS (price DOUBLE PATH '$.goodsPrice', cnt INT PATH '$.goodsCount')
			) jt ON TRUE
			WHERE t.is_handled = true AND t.handled_by_id = ? AND t.handled_by_name != ?
			GROUP BY t.id, t.pend_status_desc
		) sub`,
		user.ID, "外部系统",
	).Scan(&cumResult)

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"records":          records,
			"total":            total,
			"page":             pageInt,
			"size":             sizeInt,
			"totalAmount":      totalAmount,
			"cumulativeSubmit": cumResult.SubmitCount,
			"cumulativePend":   cumResult.PendCount,
			"cumulativeAmount": cumResult.CumulativeAmount,
		},
	})
}

// GetStats 数据统计（用于看板）
func GetStats(c *gin.Context) {
	user := c.MustGet("user").(*models.User)

	type DayCount struct {
		Day   string `json:"day"`
		Count int    `json:"count"`
	}
	var dailyCounts []DayCount
	shanghaiLoc30, _ := time.LoadLocation("Asia/Shanghai")
	thirtyDaysAgo := time.Now().In(shanghaiLoc30).AddDate(0, 0, -30).Format("2006-01-02 00:00:00")
	database.DB.Raw(`
		SELECT DATE(create_time) as day, COUNT(*) as count
		FROM trade_abnormals
		WHERE create_time >= ?
		GROUP BY DATE(create_time)
		ORDER BY day ASC
	`, thirtyDaysAgo).Scan(&dailyCounts)

	type TypeCount struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}
	var typeCounts []TypeCount
	database.DB.Raw(`
		SELECT abnormal_type_desc as name, COUNT(*) as value
		FROM trade_abnormals
		GROUP BY abnormal_type_desc
		ORDER BY value DESC
	`).Scan(&typeCounts)

	var handledCount, unhandledCount int64
	database.DB.Model(&models.TradeAbnormal{}).Where("is_handled = ?", true).Count(&handledCount)
	database.DB.Model(&models.TradeAbnormal{}).Where("is_handled = ?", false).Count(&unhandledCount)

	var total int64
	database.DB.Model(&models.TradeAbnormal{}).Count(&total)

	var todayCount int64
	shanghaiLoc, _ := time.LoadLocation("Asia/Shanghai")
	todayStr := time.Now().In(shanghaiLoc).Format("2006-01-02")
	database.DB.Model(&models.TradeAbnormal{}).
		Where("DATE(create_time) = ?", todayStr).Count(&todayCount)

	// 待质检数量
	var pendingReviewCount int64
	database.DB.Model(&models.TradeReview{}).Where("review_status = ?", "pending").Count(&pendingReviewCount)

	type OpStat struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}
	var opStats []OpStat
	if user.Role == "admin" || user.Role == "statistician" {
		database.DB.Raw(`
			SELECT handled_by_name as name, COUNT(*) as value
			FROM trade_abnormals
			WHERE is_handled = true
			GROUP BY handled_by_id, handled_by_name
			ORDER BY value DESC
		`).Scan(&opStats)
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"total":              total,
			"todayCount":         todayCount,
			"handledCount":       handledCount,
			"unhandledCount":     unhandledCount,
			"pendingReviewCount": pendingReviewCount,
			"dailyCounts":        dailyCounts,
			"typeCounts":         typeCounts,
			"opStats":            opStats,
		},
	})
}
