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

	// 查询所有有处理事件的用户（审核中和已完成使用同一事实源）
	type handlerInfo struct {
		ID   uint
		Name string
	}
	var handlers []handlerInfo
	allEventsSQL, allEventsArgs := buildOperatorRecordUnion(operatorRecordFilter{})
	if err := database.DB.Raw(`SELECT handled_by_id AS id, MAX(handled_by_name) AS name
		FROM (`+allEventsSQL+`) AS events GROUP BY handled_by_id`, allEventsArgs...).Scan(&handlers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询操作员失败"})
		return
	}

	if len(handlers) == 0 {
		c.JSON(http.StatusOK, gin.H{"code": 200, "data": []gin.H{}})
		return
	}

	// 收集用户 ID
	ids := make([]uint, len(handlers))
	for i, h := range handlers {
		ids[i] = h.ID
	}

	// 今日开始/跳过来自点击计数；提交/挂起来自统一操作事件源。
	var todayStats []models.DailyStats
	database.DB.Where("user_id IN ? AND date = ?", ids, today).Find(&todayStats)
	todayMap := make(map[uint]*models.DailyStats, len(todayStats))
	for i := range todayStats {
		todayMap[todayStats[i].UserID] = &todayStats[i]
	}
	todayStart, _ := parseShanghaiDate(today)
	todayFilter := operatorRecordFilter{
		StartTime:        todayStart.Format("2006-01-02 15:04:05"),
		EndTimeExclusive: todayStart.AddDate(0, 0, 1).Format("2006-01-02 15:04:05"),
	}
	todayEventsSQL, todayEventsArgs := buildOperatorRecordUnion(todayFilter)
	type todayAction struct {
		UserID      uint `gorm:"column:user_id"`
		SubmitCount int  `gorm:"column:submit_count"`
		PendCount   int  `gorm:"column:pend_count"`
	}
	var todayActions []todayAction
	if err := database.DB.Raw(`SELECT handled_by_id AS user_id,
		SUM(CASE WHEN action_type = 'submit' THEN 1 ELSE 0 END) AS submit_count,
		SUM(CASE WHEN action_type = 'pend' THEN 1 ELSE 0 END) AS pend_count
		FROM (`+todayEventsSQL+`) AS events GROUP BY handled_by_id`, todayEventsArgs...).Scan(&todayActions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询今日统计失败"})
		return
	}
	actionMap := make(map[uint]todayAction, len(todayActions))
	for _, action := range todayActions {
		actionMap[action.UserID] = action
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
		todayStart, todaySkip := 0, 0
		if ts != nil {
			todayStart = ts.StartCount
			todaySkip = ts.SkipCount
		}
		todaySubmit := actionMap[h.ID].SubmitCount
		todayPend := actionMap[h.ID].PendCount

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
	HandledByID     uint       `json:"handledById"`
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

// operatorRecordFilter describes the only differences between operator and
// administrator views. The underlying event source is deliberately shared.
type operatorRecordFilter struct {
	UserIDs          []uint
	StartTime        string
	EndTimeExclusive string
	EndTimeInclusive string
	ActionType       string
	InspectStatus    string
}

func parseShanghaiDate(value string) (time.Time, error) {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		return time.Time{}, err
	}
	return time.ParseInLocation("2006-01-02", value, loc)
}

func parseShanghaiDateTime(value string) (time.Time, error) {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		return time.Time{}, err
	}
	for _, layout := range []string{"2006-01-02 15:04:05", "2006-01-02 15:04"} {
		if parsed, parseErr := time.ParseInLocation(layout, value, loc); parseErr == nil {
			return parsed, nil
		}
	}
	return time.Time{}, &time.ParseError{Layout: "2006-01-02 15:04[:05]", Value: value}
}

// buildOperatorRecordUnion builds the canonical operator-event data set.
// Review-mode events keep their immutable submission time and action type;
// direct-mode events use the completed trade and exclude any reviewed trade.
func buildOperatorRecordUnion(filter operatorRecordFilter) (string, []interface{}) {
	applyFilters := func(sql string, args []interface{}, userColumn, timeColumn, actionExpr, inspectColumn string) (string, []interface{}) {
		if len(filter.UserIDs) > 0 {
			sql += " AND " + userColumn + " IN ?"
			args = append(args, filter.UserIDs)
		}
		if filter.StartTime != "" {
			sql += " AND " + timeColumn + " >= ?"
			args = append(args, filter.StartTime)
		}
		if filter.EndTimeExclusive != "" {
			sql += " AND " + timeColumn + " < ?"
			args = append(args, filter.EndTimeExclusive)
		} else if filter.EndTimeInclusive != "" {
			sql += " AND " + timeColumn + " <= ?"
			args = append(args, filter.EndTimeInclusive)
		}
		if filter.ActionType != "" {
			sql += " AND " + actionExpr + " = ?"
			args = append(args, filter.ActionType)
		}
		switch filter.InspectStatus {
		case "none", "uninspected":
			sql += " AND (" + inspectColumn + " = '' OR " + inspectColumn + " IS NULL)"
		case "normal", "abnormal":
			sql += " AND " + inspectColumn + " = ?"
			args = append(args, filter.InspectStatus)
		}
		return sql, args
	}

	reviewSQL := `SELECT t.id, COALESCE(t.trade_id, 0) AS trade_id,
		r.action_type,
		r.submitted_by_id AS handled_by_id,
		r.submitted_by_name AS handled_by_name,
		r.submitted_at AS handled_at,
		r.operator_remark AS handle_remark,
		r.goods_json AS handle_goods,
		r.review_status,
		COALESCE(t.is_handled, 0) AS is_handled,
		COALESCE(t.out_order_no, '') AS out_order_no,
		COALESCE(t.node_name, '') AS node_name,
		COALESCE(t.create_time, '') AS create_time,
		COALESCE(t.inspect_status, '') AS inspect_status,
		COALESCE(t.inspected_by_name, '') AS inspected_by_name,
		t.inspected_at,
		COALESCE(t.inspect_remark, '') AS inspect_remark,
		t.video_duration
	FROM trade_reviews r
	LEFT JOIN trade_abnormals t ON t.id = r.trade_id
	WHERE r.submitted_by_name != '外部系统' AND r.submitted_by_name != ''`
	reviewSQL, reviewArgs := applyFilters(reviewSQL, nil, "r.submitted_by_id", "r.submitted_at", "r.action_type", "t.inspect_status")

	directSQL := `SELECT t.id, t.trade_id,
		CASE WHEN t.pend_status_desc IN ('已挂起', 'PENDING') THEN 'pend' ELSE 'submit' END AS action_type,
		t.handled_by_id,
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
	WHERE t.is_handled = 1
	  AND t.handled_by_id IS NOT NULL
	  AND t.handled_at IS NOT NULL
	  AND t.handled_by_name != '外部系统'
	  AND t.handled_by_name != ''
	  AND NOT EXISTS (SELECT 1 FROM trade_reviews r2 WHERE r2.trade_id = t.id)`
	directActionExpr := "CASE WHEN t.pend_status_desc IN ('已挂起', 'PENDING') THEN 'pend' ELSE 'submit' END"
	directSQL, directArgs := applyFilters(directSQL, nil, "t.handled_by_id", "t.handled_at", directActionExpr, "t.inspect_status")

	return reviewSQL + "\nUNION ALL\n" + directSQL, append(reviewArgs, directArgs...)
}

func queryOperatorRecords(filter operatorRecordFilter, page, size int) ([]opRecord, int64, error) {
	unionSQL, unionArgs := buildOperatorRecordUnion(filter)
	var total int64
	if err := database.DB.Raw("SELECT COUNT(*) FROM ("+unionSQL+") AS events", unionArgs...).Scan(&total).Error; err != nil {
		return nil, 0, err
	}

	type rawRow struct {
		ID              uint       `gorm:"column:id"`
		TradeID         int64      `gorm:"column:trade_id"`
		ActionType      string     `gorm:"column:action_type"`
		HandledByID     uint       `gorm:"column:handled_by_id"`
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
	args := append(append([]interface{}{}, unionArgs...), size, offset)
	var rawRows []rawRow
	if err := database.DB.Raw("SELECT * FROM ("+unionSQL+") AS events ORDER BY handled_at DESC LIMIT ? OFFSET ?", args...).Scan(&rawRows).Error; err != nil {
		return nil, 0, err
	}

	records := make([]opRecord, 0, len(rawRows))
	for _, r := range rawRows {
		var handledAt time.Time
		if r.HandledAt != nil {
			handledAt = *r.HandledAt
		}
		records = append(records, opRecord{
			ID: r.ID, TradeID: r.TradeID, ActionType: r.ActionType,
			HandledByID: r.HandledByID, HandledByName: r.HandledByName, HandledAt: handledAt,
			HandleRemark: r.HandleRemark, HandleGoods: r.HandleGoods,
			ReviewStatus: r.ReviewStatus, IsHandled: r.IsHandled,
			OutOrderNo: r.OutOrderNo, NodeName: r.NodeName, CreateTime: r.CreateTime,
			InspectStatus: r.InspectStatus, InspectedByName: r.InspectedByName,
			InspectedAt: r.InspectedAt, InspectRemark: r.InspectRemark,
			VideoDuration: r.VideoDuration,
		})
	}
	return records, total, nil
}

type operatorRecordSummary struct {
	SubmitCount int64
	PendCount   int64
	Amount      float64
}

func queryOperatorRecordSummary(filter operatorRecordFilter) (operatorRecordSummary, error) {
	unionSQL, args := buildOperatorRecordUnion(filter)
	var result operatorRecordSummary
	if err := database.DB.Raw(`SELECT
		COALESCE(SUM(CASE WHEN action_type = 'submit' THEN 1 ELSE 0 END), 0) AS submit_count,
		COALESCE(SUM(CASE WHEN action_type = 'pend' THEN 1 ELSE 0 END), 0) AS pend_count
		FROM (`+unionSQL+") AS events", args...).Scan(&result).Error; err != nil {
		return result, err
	}
	var amountResult struct {
		Amount float64 `gorm:"column:amount"`
	}
	if err := database.DB.Raw(`SELECT COALESCE(SUM(jt.price * jt.cnt), 0) AS amount
		FROM (`+unionSQL+`) AS events
		JOIN JSON_TABLE(
			CASE WHEN JSON_VALID(events.handle_goods) THEN events.handle_goods ELSE '[]' END,
			'$[*]' COLUMNS (price DOUBLE PATH '$.goodsPrice', cnt INT PATH '$.goodsCount')
		) jt ON TRUE
		WHERE events.action_type = 'submit'`, args...).Scan(&amountResult).Error; err != nil {
		return result, err
	}
	result.Amount = amountResult.Amount
	return result, nil
}

// GetOperatorRecords 分页查询处理记录，合并审核模式（reviews表）和直通模式（trade_abnormals表）
// 使用 UNION ALL + DB 层 ORDER BY / LIMIT / OFFSET，避免全量加载后在 Go 内存中排序分页
func GetOperatorRecords(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	userIDStr := c.Query("userId")
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")
	actionType := c.Query("actionType")       // "submit" | "pend" | ""
	inspectStatus := c.Query("inspectStatus") // "none" | "normal" | "abnormal" | ""

	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 20
	}

	filter := operatorRecordFilter{ActionType: actionType, InspectStatus: inspectStatus}
	if userIDStr != "" {
		uid, err := strconv.ParseUint(userIDStr, 10, 64)
		if err != nil || uid == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的用户ID"})
			return
		}
		filter.UserIDs = []uint{uint(uid)}
	}
	if startDate != "" {
		start, err := parseShanghaiDate(startDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "开始日期格式错误"})
			return
		}
		filter.StartTime = start.Format("2006-01-02 15:04:05")
	}
	if endDate != "" {
		end, err := parseShanghaiDate(endDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "结束日期格式错误"})
			return
		}
		filter.EndTimeExclusive = end.AddDate(0, 0, 1).Format("2006-01-02 15:04:05")
	}

	records, total, err := queryOperatorRecords(filter, page, size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询处理记录失败"})
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

type dailyActionRow struct {
	Date        string `json:"date" gorm:"column:date"`
	SubmitCount int    `json:"submitCount" gorm:"column:submit_count"`
	PendCount   int    `json:"pendCount" gorm:"column:pend_count"`
}

func queryDailyActionStats(userIDs []uint) ([]dailyActionRow, error) {
	unionSQL, args := buildOperatorRecordUnion(operatorRecordFilter{UserIDs: userIDs})
	var rows []dailyActionRow
	err := database.DB.Raw(`SELECT DATE(handled_at) AS date,
		SUM(CASE WHEN action_type = 'submit' THEN 1 ELSE 0 END) AS submit_count,
		SUM(CASE WHEN action_type = 'pend' THEN 1 ELSE 0 END) AS pend_count
		FROM (`+unionSQL+`) AS events
		GROUP BY DATE(handled_at)
		ORDER BY date ASC`, args...).Scan(&rows).Error
	return rows, err
}

// GetDailyStats returns the same immutable operator events used by both the
// administrator record list and the operator's personal page.
func GetDailyStats(c *gin.Context) {
	var userIDs []uint
	if value := c.Query("userId"); value != "" {
		uid, err := strconv.ParseUint(value, 10, 64)
		if err != nil || uid == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的用户ID"})
			return
		}
		userIDs = []uint{uint(uid)}
	}
	rows, err := queryDailyActionStats(userIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询每日统计失败"})
		return
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
	startAt, startErr := parseShanghaiDateTime(startTime)
	endAt, endErr := parseShanghaiDateTime(endTime)
	if startErr != nil || endErr != nil || endAt.Before(startAt) {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "时间范围格式错误"})
		return
	}
	startTime = startAt.Format("2006-01-02 15:04:05")
	endTime = endAt.Format("2006-01-02 15:04:05")

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
	startDate := startAt.Format("2006-01-02")
	endDate := endAt.Format("2006-01-02")

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

	// ===== 提交处理/挂起：统一操作事件源，保留秒级精度 =====
	type actionAgg struct {
		HandledByID uint   `gorm:"column:handled_by_id"`
		ActionType  string `gorm:"column:action_type"`
		Cnt         int    `gorm:"column:cnt"`
	}
	unionSQL, unionArgs := buildOperatorRecordUnion(operatorRecordFilter{
		UserIDs: userIDs, StartTime: startTime, EndTimeInclusive: endTime,
	})
	var actionAggs []actionAgg
	if err := database.DB.Raw(`SELECT handled_by_id, action_type, COUNT(*) AS cnt
		FROM (`+unionSQL+`) AS events
		GROUP BY handled_by_id, action_type`, unionArgs...).Scan(&actionAggs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询时段统计失败"})
		return
	}
	submitMap := make(map[uint]int, len(userIDs))
	pendMap := make(map[uint]int, len(userIDs))
	for _, a := range actionAggs {
		if a.ActionType == "pend" {
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
