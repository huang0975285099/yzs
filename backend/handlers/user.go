package handlers

import (
	"net/http"
	"time"

	"go-yzs/database"
	"go-yzs/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CreateUserRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=6"`
	Realname string `json:"realname"`
	Role     string `json:"role" binding:"required,oneof=admin statistician operator inspector"`
	TeamID   *uint  `json:"teamId"`
}

type UpdateUserRequest struct {
	Password string `json:"password"`
	Realname string `json:"realname"`
	Role     string `json:"role" binding:"omitempty,oneof=admin statistician operator inspector"`
	TeamID   *uint  `json:"teamId"` // nil=不修改, 0=清除团队, >0=设置团队
}

func ListUsers(c *gin.Context) {
	var users []models.User
	database.DB.Preload("Team").Find(&users)
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": users})
}

func CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}

	// Check username uniqueness
	var count int64
	database.DB.Model(&models.User{}).Where("username = ?", req.Username).Count(&count)
	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "用户名已存在"})
		return
	}

	user := models.User{
		Username: req.Username,
		Realname: req.Realname,
		Role:     req.Role,
		TeamID:   req.TeamID,
	}
	if err := user.SetPassword(req.Password); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "密码加密失败"})
		return
	}

	if err := database.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "创建用户失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "创建成功", "data": user})
}

func UpdateUser(c *gin.Context) {
	id := c.Param("id")
	var user models.User
	if err := database.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "用户不存在"})
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}

	if req.Realname != "" {
		user.Realname = req.Realname
	}
	if req.Role != "" {
		user.Role = req.Role
	}
	if req.Password != "" {
		if err := user.SetPassword(req.Password); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "密码加密失败"})
			return
		}
	}
	if req.TeamID != nil {
		if *req.TeamID == 0 {
			user.TeamID = nil
		} else {
			user.TeamID = req.TeamID
		}
	}

	database.DB.Save(&user)
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "更新成功", "data": user})
}

func DeleteUser(c *gin.Context) {
	id := c.Param("id")

	// Prevent deleting yourself
	currentUser, _ := c.Get("user")
	cu := currentUser.(*models.User)

	var user models.User
	if err := database.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "用户不存在"})
		return
	}

	if cu.ID == user.ID {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "不能删除自己"})
		return
	}

	database.DB.Delete(&user)
	// Also delete sessions
	database.DB.Where("user_id = ?", user.ID).Delete(&models.UserSession{})
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "删除成功"})
}

// incrementDailyStats 增加用户当日统计
func incrementDailyStats(userID uint, field string) {
	today := time.Now().Format("2006-01-02")

	var stats models.DailyStats
	err := database.DB.Where("user_id = ? AND date = ?", userID, today).First(&stats).Error
	if err != nil {
		// 不存在，创建新记录
		stats = models.DailyStats{
			UserID:      userID,
			Date:        today,
			StartCount:  0,
			SkipCount:   0,
			SubmitCount: 0,
			PendCount:   0,
		}
		database.DB.Create(&stats)
	}

	// 增加对应字段
	switch field {
	case "start":
		database.DB.Model(&stats).UpdateColumn("start_count", gorm.Expr("start_count + 1"))
	case "skip":
		database.DB.Model(&stats).UpdateColumn("skip_count", gorm.Expr("skip_count + 1"))
	case "submit":
		database.DB.Model(&stats).UpdateColumn("submit_count", gorm.Expr("submit_count + 1"))
	case "pend":
		database.DB.Model(&stats).UpdateColumn("pend_count", gorm.Expr("pend_count + 1"))
	}
}

// IncrementStart 记录点击"开始处理"次数
func IncrementStart(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	database.DB.Model(user).UpdateColumn("start_count", gorm.Expr("start_count + 1"))
	incrementDailyStats(user.ID, "start")
	c.JSON(http.StatusOK, gin.H{"code": 200})
}

// IncrementSkip 记录点击"跳过"次数
func IncrementSkip(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	database.DB.Model(user).UpdateColumn("skip_count", gorm.Expr("skip_count + 1"))
	incrementDailyStats(user.ID, "skip")
	c.JSON(http.StatusOK, gin.H{"code": 200})
}

// GetMyInspectStats 质检员查看自己的复查统计
func GetMyInspectStats(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	date := c.Query("date")
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	type DayStat struct {
		Date          string `json:"date"`
		NormalCount   int    `json:"normalCount"`
		AbnormalCount int    `json:"abnormalCount"`
		Total         int    `json:"total"`
	}

	// 今日统计（范围比较代替 DATE() 函数，走 inspected_at 索引，使用 CST 时区）
	loc, _ := time.LoadLocation("Asia/Shanghai")
	startOfDay, err := time.ParseInLocation("2006-01-02", date, loc)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "日期格式错误"})
		return
	}
	endOfDay := startOfDay.Add(24 * time.Hour)
	var todayNormal, todayAbnormal int64
	database.DB.Model(&models.TradeAbnormal{}).
		Where("inspected_by_id = ? AND inspect_status = 'normal' AND inspected_at >= ? AND inspected_at < ?", user.ID, startOfDay, endOfDay).
		Count(&todayNormal)
	database.DB.Model(&models.TradeAbnormal{}).
		Where("inspected_by_id = ? AND inspect_status = 'abnormal' AND inspected_at >= ? AND inspected_at < ?", user.ID, startOfDay, endOfDay).
		Count(&todayAbnormal)

	// 最近30天历史（按日聚合）
	type Row struct {
		Date          string `json:"date"`
		NormalCount   int    `json:"normalCount"`
		AbnormalCount int    `json:"abnormalCount"`
	}
	var rows []Row
	database.DB.Model(&models.TradeAbnormal{}).
		Select("DATE(inspected_at) AS date, SUM(CASE WHEN inspect_status='normal' THEN 1 ELSE 0 END) AS normal_count, SUM(CASE WHEN inspect_status='abnormal' THEN 1 ELSE 0 END) AS abnormal_count").
		Where("inspected_by_id = ? AND inspect_status != '' AND inspected_at >= ?", user.ID, time.Now().AddDate(0, 0, -30)).
		Group("DATE(inspected_at)").
		Order("date DESC").
		Scan(&rows)

	history := make([]DayStat, 0, len(rows))
	for _, r := range rows {
		history = append(history, DayStat{
			Date:          r.Date,
			NormalCount:   r.NormalCount,
			AbnormalCount: r.AbnormalCount,
			Total:         r.NormalCount + r.AbnormalCount,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"date": date,
			"today": DayStat{
				Date:          date,
				NormalCount:   int(todayNormal),
				AbnormalCount: int(todayAbnormal),
				Total:         int(todayNormal + todayAbnormal),
			},
			"history": history,
		},
	})
}

// GetMyDailyStats 获取当前用户的每日统计数据
func GetMyDailyStats(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	date := c.Query("date") // 可选日期参数，格式 YYYY-MM-DD

	// 获取最近30天的统计数据
	var stats []models.DailyStats
	database.DB.Where("user_id = ?", user.ID).
		Order("date DESC").
		Limit(30).
		Find(&stats)

	// 获取指定日期的统计（默认今日）
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}
	var dayStats models.DailyStats
	database.DB.Where("user_id = ? AND date = ?", user.ID, date).First(&dayStats)

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"date": date,
			"today": gin.H{
				"startCount":  dayStats.StartCount,
				"skipCount":   dayStats.SkipCount,
				"submitCount": dayStats.SubmitCount,
				"pendCount":   dayStats.PendCount,
			},
			"history": stats,
		},
	})
}
