package handlers

import (
	"go-yzs/database"
	"go-yzs/models"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// GetOperatorStats 返回所有有处理记录的用户的统计数据（管理员/统计员可用）
func GetOperatorStats(c *gin.Context) {
	shanghaiLoc, _ := time.LoadLocation("Asia/Shanghai")
	today := time.Now().In(shanghaiLoc).Format("2006-01-02")

	// 查询所有有处理记录的用户ID（从 trade_abnormals 表）
	type handlerInfo struct {
		ID   uint
		Name string
	}
	var handlers []handlerInfo
	database.DB.Model(&models.TradeAbnormal{}).
		Select("handled_by_id as id, MAX(handled_by_name) as name").
		Where("is_handled = 1 AND handled_by_id IS NOT NULL AND handled_by_name != '外部系统' AND handled_by_name != ''").
		Group("handled_by_id").
		Scan(&handlers)

	if len(handlers) == 0 {
		c.JSON(http.StatusOK, gin.H{"code": 200, "data": []gin.H{}})
		return
	}

	// 收集用户 ID
	ids := make([]uint, len(handlers))
	for i, h := range handlers {
		ids[i] = h.ID
	}

	// 今日统计
	var todayStats []models.DailyStats
	database.DB.Where("user_id IN ? AND date = ?", ids, today).Find(&todayStats)
	todayMap := make(map[uint]*models.DailyStats, len(todayStats))
	for i := range todayStats {
		todayMap[todayStats[i].UserID] = &todayStats[i]
	}

	// 查询用户信息补充 username 和 realname
	var users []models.User
	database.DB.Where("id IN ?", ids).Find(&users)
	userMap := make(map[uint]*models.User, len(users))
	for i := range users {
		userMap[users[i].ID] = &users[i]
	}

	// 构建结果
	result := make([]gin.H, 0, len(handlers))
	for _, h := range handlers {
		ts := todayMap[h.ID]
		todayStart, todaySkip, todaySubmit, todayPend := 0, 0, 0, 0
		if ts != nil {
			todayStart = ts.StartCount
			todaySkip = ts.SkipCount
			todaySubmit = ts.SubmitCount
			todayPend = ts.PendCount
		}

		username := h.Name
		realname := h.Name
		if u := userMap[h.ID]; u != nil {
			username = u.Username
			realname = u.Realname
			if realname == "" {
				realname = u.Username
			}
		}

		result = append(result, gin.H{
			"userId":      h.ID,
			"username":    username,
			"realname":    realname,
			"todayStart":  todayStart,
			"todaySkip":   todaySkip,
			"todaySubmit": todaySubmit,
			"todayPend":   todayPend,
		})
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": result})
}

type opRecord struct {
	ID              uint       `json:"id"`
	TradeID         int64      `json:"tradeId"`
	ActionType      string     `json:"actionType"`
	HandledByName   string     `json:"handledByName"`
	HandledAt       time.Time  `json:"handledAt"`
	HandleRemark    string     `json:"handleRemark"`
	HandleGoods     string     `json:"handleGoods"`
	ReviewStatus    string     `json:"reviewStatus"`
	IsHandled       bool       `json:"isHandled"`
	OutOrderNo      string     `json:"outOrderNo"`
	NodeName        string     `json:"nodeName"`
	CreateTime      string     `json:"createTime"`
	InspectStatus   string     `json:"inspectStatus"`
	InspectedByName string     `json:"inspectedByName"`
	InspectedAt     *time.Time `json:"inspectedAt"`
	InspectRemark   string     `json:"inspectRemark"`
	VideoDuration   *int       `json:"videoDuration"`
	SortKey         time.Time  `json:"-"`
}

// GetOperatorRecords 分页查询处理记录，合并审核模式（reviews表）和直通模式（trade_abnormals表）
// 使用 UNION ALL + DB 层 ORDER BY / LIMIT / OFFSET，避免全量加载后在 Go 内存中排序分页
func GetOperatorRecords(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	sizeStr := c.DefaultQuery("size", "20")
	userIDStr := c.Query("userId")
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")
	actionType := c.Query("actionType")       // "submit" | "pend" | ""
	inspectStatus := c.Query("inspectStatus") // "none" | "normal" | "abnormal" | ""

	page, _ := strconv.Atoi(pageStr)
	size, _ := strconv.Atoi(sizeStr)
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 20
	}

	var uid uint64
	if userIDStr != "" {
		uid, _ = strconv.ParseUint(userIDStr, 10, 64)
	}

	// reviews 来源仅在 inspectStatus 不要求 "normal"/"abnormal" 时包含
	needReviews := (actionType == "" || actionType == "submit" || actionType == "pend") &&
		(inspectStatus == "" || inspectStatus == "none")

	var unionParts []string
	var unionArgs []interface{}

	// ===== 来源1：审核模式（trade_reviews JOIN trade_abnormals）=====
	if needReviews {
		sql := `SELECT r.id, r.trade_id,
			r.action_type,
			r.submitted_by_name  AS handled_by_name,
			r.submitted_at       AS handled_at,
			r.operator_remark    AS handle_remark,
			r.goods_json         AS handle_goods,
			r.review_status,
			COALESCE(t.is_handled,   0)  AS is_handled,
			COALESCE(t.out_order_no,'')  AS out_order_no,
			COALESCE(t.node_name,   '')  AS node_name,
			COALESCE(t.create_time, '')  AS create_time,
			'' AS inspect_status, '' AS inspected_by_name, NULL AS inspected_at, '' AS inspect_remark,
			t.video_duration
		FROM trade_reviews r
		LEFT JOIN trade_abnormals t ON t.id = r.trade_id
		WHERE r.submitted_by_name != '外部系统' AND r.submitted_by_name != ''`
		var sqlArgs []interface{}
		if uid > 0 {
			sql += " AND r.submitted_by_id = ?"
			sqlArgs = append(sqlArgs, uid)
		}
		if startDate != "" {
			sql += " AND r.submitted_at >= ?"
			sqlArgs = append(sqlArgs, startDate+" 00:00:00")
		}
		if endDate != "" {
			sql += " AND r.submitted_at <= ?"
			sqlArgs = append(sqlArgs, endDate+" 23:59:59")
		}
		if actionType != "" {
			sql += " AND r.action_type = ?"
			sqlArgs = append(sqlArgs, actionType)
		}
		unionParts = append(unionParts, sql)
		unionArgs = append(unionArgs, sqlArgs...)
	}

	// ===== 来源2：直通模式（trade_abnormals 排除已在 reviews 中的）=====
	// 用 NOT EXISTS 关联子查询代替 NOT IN (大列表)，大表时更稳定
	{
		notExists := "NOT EXISTS (SELECT 1 FROM trade_reviews r2 WHERE r2.trade_id = t.id"
		var neArgs []interface{}
		if uid > 0 {
			notExists += " AND r2.submitted_by_id = ?"
			neArgs = append(neArgs, uid)
		}
		notExists += ")"

		sql := `SELECT t.id, t.trade_id,
			CASE WHEN t.pend_status_desc = '已挂起' THEN 'pend' ELSE 'submit' END AS action_type,
			t.handled_by_name,
			t.handled_at,
			t.handle_remark,
			t.handle_goods,
			t.review_status,
			t.is_handled,
			t.out_order_no, t.node_name, t.create_time,
			t.inspect_status, t.inspected_by_name, t.inspected_at, t.inspect_remark,
			t.video_duration
		FROM trade_abnormals t
		WHERE t.is_handled = 1 AND t.handled_by_name != '外部系统' AND t.handled_by_name != ''
		  AND ` + notExists
		sqlArgs := append([]interface{}{}, neArgs...)

		if uid > 0 {
			sql += " AND t.handled_by_id = ?"
			sqlArgs = append(sqlArgs, uid)
		}
		if startDate != "" {
			sql += " AND t.handled_at >= ?"
			sqlArgs = append(sqlArgs, startDate+" 00:00:00")
		}
		if endDate != "" {
			sql += " AND t.handled_at <= ?"
			sqlArgs = append(sqlArgs, endDate+" 23:59:59")
		}
		switch actionType {
		case "submit":
			sql += " AND (t.pend_status_desc != '已挂起' OR t.pend_status_desc IS NULL OR t.pend_status_desc = '')"
		case "pend":
			sql += " AND t.pend_status_desc = '已挂起'"
		}
		switch inspectStatus {
		case "none":
			sql += " AND (t.inspect_status = '' OR t.inspect_status IS NULL)"
		case "normal", "abnormal":
			sql += " AND t.inspect_status = ?"
			sqlArgs = append(sqlArgs, inspectStatus)
		}
		unionParts = append(unionParts, sql)
		unionArgs = append(unionArgs, sqlArgs...)
	}

	unionSQL := strings.Join(unionParts, "\nUNION ALL\n")

	// COUNT（复用相同参数）
	var total int64
	database.DB.Raw("SELECT COUNT(*) FROM ("+unionSQL+") AS _u", unionArgs...).Scan(&total)

	// 分页查询：DB 层排序 + LIMIT/OFFSET，只传回一页数据
	type rawRow struct {
		ID              uint       `gorm:"column:id"`
		TradeID         int64      `gorm:"column:trade_id"`
		ActionType      string     `gorm:"column:action_type"`
		HandledByName   string     `gorm:"column:handled_by_name"`
		HandledAt       *time.Time `gorm:"column:handled_at"`
		HandleRemark    string     `gorm:"column:handle_remark"`
		HandleGoods     string     `gorm:"column:handle_goods"`
		ReviewStatus    string     `gorm:"column:review_status"`
		IsHandled       bool       `gorm:"column:is_handled"`
		OutOrderNo      string     `gorm:"column:out_order_no"`
		NodeName        string     `gorm:"column:node_name"`
		CreateTime      string     `gorm:"column:create_time"`
		InspectStatus   string     `gorm:"column:inspect_status"`
		InspectedByName string     `gorm:"column:inspected_by_name"`
		InspectedAt     *time.Time `gorm:"column:inspected_at"`
		InspectRemark   string     `gorm:"column:inspect_remark"`
		VideoDuration   *int       `gorm:"column:video_duration"`
	}
	offset := (page - 1) * size
	pagedSQL := "SELECT * FROM (" + unionSQL + ") AS _u ORDER BY handled_at DESC LIMIT ? OFFSET ?"
	pagedArgs := append(append([]interface{}{}, unionArgs...), size, offset)

	var rawRows []rawRow
	database.DB.Raw(pagedSQL, pagedArgs...).Scan(&rawRows)

	records := make([]opRecord, 0, len(rawRows))
	for _, r := range rawRows {
		var handledAt time.Time
		if r.HandledAt != nil {
			handledAt = *r.HandledAt
		}
		records = append(records, opRecord{
			ID:              r.ID,
			TradeID:         r.TradeID,
			ActionType:      r.ActionType,
			HandledByName:   r.HandledByName,
			HandledAt:       handledAt,
			HandleRemark:    r.HandleRemark,
			HandleGoods:     r.HandleGoods,
			ReviewStatus:    r.ReviewStatus,
			IsHandled:       r.IsHandled,
			OutOrderNo:      r.OutOrderNo,
			NodeName:        r.NodeName,
			CreateTime:      r.CreateTime,
			InspectStatus:   r.InspectStatus,
			InspectedByName: r.InspectedByName,
			InspectedAt:     r.InspectedAt,
			InspectRemark:   r.InspectRemark,
			VideoDuration:   r.VideoDuration,
		})
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

// GetDailyStats 返回每日提交处理和挂起数据，可按 userId 过滤
// 数据来源：审核模式查 trade_reviews（submitted_at），直通模式查 trade_abnormals（handled_at）
// 两者按日期合并，保证历史数据准确
func GetDailyStats(c *gin.Context) {
	type dailyRow struct {
		Date        string `json:"date"`
		SubmitCount int    `json:"submitCount"`
		PendCount   int    `json:"pendCount"`
	}

	userIDStr := c.Query("userId")
	var uid uint64
	if userIDStr != "" {
		uid, _ = strconv.ParseUint(userIDStr, 10, 64)
	}

	// 按日期聚合的 map
	dayMap := make(map[string]*dailyRow)
	ensure := func(date string) *dailyRow {
		if _, ok := dayMap[date]; !ok {
			dayMap[date] = &dailyRow{Date: date}
		}
		return dayMap[date]
	}

	// ===== 来源1：审核模式 —— trade_reviews 表 =====
	type reviewAgg struct {
		Date       string
		ActionType string
		Cnt        int
	}
	var reviewAggs []reviewAgg
	reviewQ := database.DB.Model(&models.TradeReview{}).
		Select("DATE(submitted_at) as date, action_type, COUNT(*) as cnt").
		Where("submitted_by_name != '外部系统' AND submitted_by_name != ''")
	if uid > 0 {
		reviewQ = reviewQ.Where("submitted_by_id = ?", uid)
	}
	reviewQ.Group("DATE(submitted_at), action_type").Scan(&reviewAggs)
	for _, r := range reviewAggs {
		row := ensure(r.Date)
		if r.ActionType == "pend" {
			row.PendCount += r.Cnt
		} else {
			row.SubmitCount += r.Cnt
		}
	}

	// ===== 来源2：直通模式 —— trade_abnormals 表（排除已在 reviews 中的） =====
	var reviewedTradeIDs []uint
	reviewIDQ := database.DB.Model(&models.TradeReview{}).Select("trade_id")
	if uid > 0 {
		reviewIDQ = reviewIDQ.Where("submitted_by_id = ?", uid)
	}
	reviewIDQ.Pluck("trade_id", &reviewedTradeIDs)

	type tradeAgg struct {
		Date           string
		PendStatusDesc string
		Cnt            int
	}
	var tradeAggs []tradeAgg
	directQ := database.DB.Model(&models.TradeAbnormal{}).
		Select("DATE(handled_at) as date, pend_status_desc, COUNT(*) as cnt").
		Where("is_handled = 1 AND handled_by_name != '外部系统' AND handled_by_name != '' AND handled_at IS NOT NULL")
	if len(reviewedTradeIDs) > 0 {
		directQ = directQ.Where("id NOT IN ?", reviewedTradeIDs)
	}
	if uid > 0 {
		directQ = directQ.Where("handled_by_id = ?", uid)
	}
	directQ.Group("DATE(handled_at), pend_status_desc").Scan(&tradeAggs)
	for _, r := range tradeAggs {
		row := ensure(r.Date)
		if r.PendStatusDesc == "已挂起" {
			row.PendCount += r.Cnt
		} else {
			row.SubmitCount += r.Cnt
		}
	}

	// 转为有序切片
	rows := make([]dailyRow, 0, len(dayMap))
	for _, v := range dayMap {
		rows = append(rows, *v)
	}
	// 按日期升序
	for i := 0; i < len(rows)-1; i++ {
		for j := i + 1; j < len(rows); j++ {
			if rows[j].Date < rows[i].Date {
				rows[i], rows[j] = rows[j], rows[i]
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": rows})
}

// GetOperatorRangeStats 查询指定操作员在精确时间段内的统计数据
// 开始/跳过 来自 daily_stats（仅日期精度），提交处理/挂起 来自 trade_reviews + trade_abnormals（秒级精度）
func GetOperatorRangeStats(c *gin.Context) {
	userIDsStr := c.Query("userIds")  // 逗号分隔的用户 ID
	startTime := c.Query("startTime") // YYYY-MM-DD HH:MM:SS
	endTime := c.Query("endTime")     // YYYY-MM-DD HH:MM:SS

	if userIDsStr == "" || startTime == "" || endTime == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "缺少参数"})
		return
	}

	// 解析用户 ID 列表
	parts := strings.Split(userIDsStr, ",")
	var userIDs []uint
	for _, p := range parts {
		id, err := strconv.ParseUint(strings.TrimSpace(p), 10, 64)
		if err == nil && id > 0 {
			userIDs = append(userIDs, uint(id))
		}
	}
	if len(userIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的用户ID"})
		return
	}

	// 查询用户基本信息
	var users []models.User
	database.DB.Where("id IN ?", userIDs).Find(&users)
	userMap := make(map[uint]*models.User, len(users))
	for i := range users {
		userMap[users[i].ID] = &users[i]
	}

	// ===== 开始/跳过：来自 daily_stats，按日期范围汇总 =====
	startDate := startTime[:10]
	endDate := endTime[:10]

	type dailyAgg struct {
		UserID     uint
		StartCount int
		SkipCount  int
	}
	var dailyAggs []dailyAgg
	database.DB.Model(&models.DailyStats{}).
		Select("user_id, SUM(start_count) as start_count, SUM(skip_count) as skip_count").
		Where("user_id IN ? AND date >= ? AND date <= ?", userIDs, startDate, endDate).
		Group("user_id").
		Scan(&dailyAggs)

	startMap := make(map[uint]int, len(userIDs))
	skipMap := make(map[uint]int, len(userIDs))
	for _, a := range dailyAggs {
		startMap[a.UserID] = a.StartCount
		skipMap[a.UserID] = a.SkipCount
	}

	// ===== 提交处理/挂起：来源1 —— trade_reviews（审核模式，秒级时间） =====
	type reviewAgg struct {
		SubmittedByID uint
		ActionType    string
		Cnt           int
	}
	var reviewAggs []reviewAgg
	database.DB.Model(&models.TradeReview{}).
		Select("submitted_by_id, action_type, COUNT(*) as cnt").
		Where("submitted_by_id IN ? AND submitted_at >= ? AND submitted_at <= ?", userIDs, startTime, endTime).
		Group("submitted_by_id, action_type").
		Scan(&reviewAggs)

	submitMap := make(map[uint]int, len(userIDs))
	pendMap := make(map[uint]int, len(userIDs))
	for _, a := range reviewAggs {
		if a.ActionType == "pend" {
			pendMap[a.SubmittedByID] += a.Cnt
		} else {
			submitMap[a.SubmittedByID] += a.Cnt
		}
	}

	// ===== 提交处理/挂起：来源2 —— trade_abnormals（直通模式，排除已在 reviews 中的） =====
	var allReviewedTradeIDs []uint
	database.DB.Model(&models.TradeReview{}).
		Select("trade_id").
		Where("submitted_by_id IN ?", userIDs).
		Pluck("trade_id", &allReviewedTradeIDs)

	type directAgg struct {
		HandledByID    uint
		PendStatusDesc string
		Cnt            int
	}
	var directAggs []directAgg
	directQ := database.DB.Model(&models.TradeAbnormal{}).
		Select("handled_by_id, pend_status_desc, COUNT(*) as cnt").
		Where("handled_by_id IN ? AND is_handled = 1 AND handled_at >= ? AND handled_at <= ?", userIDs, startTime, endTime).
		Where("handled_by_name != '外部系统' AND handled_by_name != ''")
	if len(allReviewedTradeIDs) > 0 {
		directQ = directQ.Where("id NOT IN ?", allReviewedTradeIDs)
	}
	directQ.Group("handled_by_id, pend_status_desc").Scan(&directAggs)

	for _, a := range directAggs {
		if a.PendStatusDesc == "已挂起" {
			pendMap[a.HandledByID] += a.Cnt
		} else {
			submitMap[a.HandledByID] += a.Cnt
		}
	}

	// 构建结果（保持入参顺序）
	result := make([]gin.H, 0, len(userIDs))
	for _, uid := range userIDs {
		username, realname := "", ""
		if u := userMap[uid]; u != nil {
			username = u.Username
			realname = u.Realname
			if realname == "" {
				realname = u.Username
			}
		}
		result = append(result, gin.H{
			"userId":      uid,
			"username":    username,
			"realname":    realname,
			"startCount":  startMap[uid],
			"skipCount":   skipMap[uid],
			"submitCount": submitMap[uid],
			"pendCount":   pendMap[uid],
		})
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": result})
}

// GetInspectExport 导出复查记录（按时间范围，不分页）
func GetInspectExport(c *gin.Context) {
	startTime := c.Query("startTime")
	endTime := c.Query("endTime")

	type exportRow struct {
		InspectedAt     *time.Time `json:"inspectedAt"`
		InspectedByName string     `json:"inspectedByName"`
		OutOrderNo      string     `json:"outOrderNo"`
		InspectStatus   string     `json:"inspectStatus"`
		InspectRemark   string     `json:"inspectRemark"`
		HandledByName   string     `json:"handledByName"`
	}

	q := database.DB.Model(&models.TradeAbnormal{}).
		Select("inspected_at, inspected_by_name, out_order_no, inspect_status, inspect_remark, handled_by_name").
		Where("inspect_status != ''")
	if startTime != "" {
		q = q.Where("inspected_at >= ?", startTime)
	}
	if endTime != "" {
		q = q.Where("inspected_at <= ?", endTime)
	}
	q = q.Order("inspected_at DESC")

	var rows []exportRow
	q.Scan(&rows)

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": rows})
}

// GetRandomUninspected 从未复查的已处理订单中随机返回1条的 id。
// 若质检员属于某个团队，则只从该团队成员处理的订单中随机；否则从全部订单随机。
func GetRandomUninspected(c *gin.Context) {
	inspector := c.MustGet("user").(*models.User)

	// Auth 中间件的 Redis 缓存不含 team_id，单独查一次
	var me models.User
	database.DB.Select("team_id").First(&me, inspector.ID)

	shanghaiLoc, _ := time.LoadLocation("Asia/Shanghai")
	today := time.Now().In(shanghaiLoc).Format("2006-01-02")

	const baseSQL = `
		SELECT id FROM trade_abnormals
		WHERE is_handled = 1
		  AND (inspect_status = '' OR inspect_status IS NULL)
		  AND handled_by_name != '外部系统' AND handled_by_name != ''
		  AND DATE(CONVERT_TZ(handled_at, '+00:00', '+08:00')) = ?`

	var result struct {
		ID uint `gorm:"column:id"`
	}

	if me.TeamID != nil {
		var teamUserIDs []uint
		database.DB.Model(&models.User{}).
			Where("team_id = ?", *me.TeamID).
			Pluck("id", &teamUserIDs)

		if len(teamUserIDs) == 0 {
			c.JSON(http.StatusOK, gin.H{"code": 404, "message": "没有待复查的订单"})
			return
		}
		database.DB.Raw(baseSQL+" AND handled_by_id IN ? ORDER BY RAND() LIMIT 1",
			today, teamUserIDs).Scan(&result)
	} else {
		database.DB.Raw(baseSQL+" ORDER BY RAND() LIMIT 1", today).Scan(&result)
	}

	if result.ID == 0 {
		c.JSON(http.StatusOK, gin.H{"code": 404, "message": "没有待复查的订单"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": gin.H{"id": result.ID}})
}

// GetInspectorStats 返回所有有复查记录的质检员统计数据（今日复查正常/异常/总计）
func GetInspectorStats(c *gin.Context) {
	shanghaiLoc, _ := time.LoadLocation("Asia/Shanghai")
	today := time.Now().In(shanghaiLoc).Format("2006-01-02")

	type inspectorInfo struct {
		ID   uint
		Name string
	}
	var inspectors []inspectorInfo
	database.DB.Model(&models.TradeAbnormal{}).
		Select("inspected_by_id as id, MAX(inspected_by_name) as name").
		Where("inspected_by_id > 0 AND inspect_status != ''").
		Group("inspected_by_id").
		Scan(&inspectors)

	if len(inspectors) == 0 {
		c.JSON(http.StatusOK, gin.H{"code": 200, "data": []gin.H{}})
		return
	}

	ids := make([]uint, len(inspectors))
	for i, h := range inspectors {
		ids[i] = h.ID
	}

	var users []models.User
	database.DB.Where("id IN ?", ids).Find(&users)
	userMap := make(map[uint]*models.User, len(users))
	for i := range users {
		userMap[users[i].ID] = &users[i]
	}

	type todayRow struct {
		InspectedByID uint
		InspectStatus string
		Cnt           int
	}
	var todayRows []todayRow
	database.DB.Model(&models.TradeAbnormal{}).
		Select("inspected_by_id, inspect_status, COUNT(*) as cnt").
		Where("inspected_by_id IN ? AND inspect_status != '' AND DATE(inspected_at) = ?", ids, today).
		Group("inspected_by_id, inspect_status").
		Scan(&todayRows)

	type dayStat struct {
		Normal   int
		Abnormal int
	}
	todayMap := make(map[uint]*dayStat, len(ids))
	for _, r := range todayRows {
		if todayMap[r.InspectedByID] == nil {
			todayMap[r.InspectedByID] = &dayStat{}
		}
		if r.InspectStatus == "normal" {
			todayMap[r.InspectedByID].Normal += r.Cnt
		} else if r.InspectStatus == "abnormal" {
			todayMap[r.InspectedByID].Abnormal += r.Cnt
		}
	}

	result := make([]gin.H, 0, len(inspectors))
	for _, h := range inspectors {
		ts := todayMap[h.ID]
		todayNormal, todayAbnormal := 0, 0
		if ts != nil {
			todayNormal = ts.Normal
			todayAbnormal = ts.Abnormal
		}
		username, realname := h.Name, h.Name
		if u := userMap[h.ID]; u != nil {
			username = u.Username
			realname = u.Realname
			if realname == "" {
				realname = u.Username
			}
		}
		result = append(result, gin.H{
			"userId":        h.ID,
			"username":      username,
			"realname":      realname,
			"todayNormal":   todayNormal,
			"todayAbnormal": todayAbnormal,
			"todayTotal":    todayNormal + todayAbnormal,
		})
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": result})
}

// GetInspectorRangeStats 查询指定质检员在精确时间段内的复查统计
func GetInspectorRangeStats(c *gin.Context) {
	userIDsStr := c.Query("userIds")
	startTime := c.Query("startTime")
	endTime := c.Query("endTime")

	if userIDsStr == "" || startTime == "" || endTime == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "缺少参数"})
		return
	}

	parts := strings.Split(userIDsStr, ",")
	var userIDs []uint
	for _, p := range parts {
		id, err := strconv.ParseUint(strings.TrimSpace(p), 10, 64)
		if err == nil && id > 0 {
			userIDs = append(userIDs, uint(id))
		}
	}
	if len(userIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的用户ID"})
		return
	}

	var users []models.User
	database.DB.Where("id IN ?", userIDs).Find(&users)
	userMap := make(map[uint]*models.User, len(users))
	for i := range users {
		userMap[users[i].ID] = &users[i]
	}

	type aggRow struct {
		InspectedByID uint
		InspectStatus string
		Cnt           int
	}
	var aggs []aggRow
	database.DB.Model(&models.TradeAbnormal{}).
		Select("inspected_by_id, inspect_status, COUNT(*) as cnt").
		Where("inspected_by_id IN ? AND inspect_status != '' AND inspected_at >= ? AND inspected_at <= ?", userIDs, startTime, endTime).
		Group("inspected_by_id, inspect_status").
		Scan(&aggs)

	type stat struct{ Normal, Abnormal int }
	statMap := make(map[uint]*stat, len(userIDs))
	for _, a := range aggs {
		if statMap[a.InspectedByID] == nil {
			statMap[a.InspectedByID] = &stat{}
		}
		if a.InspectStatus == "normal" {
			statMap[a.InspectedByID].Normal += a.Cnt
		} else if a.InspectStatus == "abnormal" {
			statMap[a.InspectedByID].Abnormal += a.Cnt
		}
	}

	result := make([]gin.H, 0, len(userIDs))
	for _, uid := range userIDs {
		username, realname := "", ""
		if u := userMap[uid]; u != nil {
			username = u.Username
			realname = u.Realname
			if realname == "" {
				realname = u.Username
			}
		}
		s := statMap[uid]
		normalCnt, abnormalCnt := 0, 0
		if s != nil {
			normalCnt = s.Normal
			abnormalCnt = s.Abnormal
		}
		result = append(result, gin.H{
			"userId":        uid,
			"username":      username,
			"realname":      realname,
			"normalCount":   normalCnt,
			"abnormalCount": abnormalCnt,
			"total":         normalCnt + abnormalCnt,
		})
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": result})
}
