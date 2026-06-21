package handlers

import (
	"encoding/json"
	"go-yzs/database"
	"go-yzs/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// ListReviews 质检列表
func ListReviews(c *gin.Context) {
	status := c.DefaultQuery("status", "pending") // pending, approved, all
	page := c.DefaultQuery("page", "1")
	size := c.DefaultQuery("size", "20")

	var pageInt, sizeInt int
	parseIntParam(page, &pageInt)
	parseIntParam(size, &sizeInt)
	if pageInt < 1 {
		pageInt = 1
	}
	if sizeInt < 1 {
		sizeInt = 20
	}

	user := c.MustGet("user").(*models.User)
	mine := c.Query("mine") == "true"

	query := database.DB.Model(&models.TradeReview{})
	if status != "all" {
		query = query.Where("review_status = ?", status)
	}
	if mine {
		query = query.Where("submitted_by_id = ?", user.ID)
	}

	var total int64
	query.Count(&total)

	var reviews []models.TradeReview
	offset := (pageInt - 1) * sizeInt
	query.Order("submitted_at DESC").Offset(offset).Limit(sizeInt).Find(&reviews)

	// 手动关联订单数据
	if len(reviews) > 0 {
		tradeIDs := make([]uint, 0, len(reviews))
		for _, r := range reviews {
			tradeIDs = append(tradeIDs, r.TradeID)
		}
		var trades []models.TradeAbnormal
		database.DB.Where("id IN ?", tradeIDs).Find(&trades)
		tradeMap := make(map[uint]models.TradeAbnormal, len(trades))
		for _, t := range trades {
			tradeMap[t.ID] = t
		}
		for i := range reviews {
			reviews[i].Trade = tradeMap[reviews[i].TradeID]
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"records": reviews,
			"total":   total,
			"page":    pageInt,
			"size":    sizeInt,
		},
	})
}

// ApproveReview 质检通过：预检外部状态 → 调外部 API → 更新本地
func ApproveReview(c *gin.Context) {
	id := c.Param("id")
	user := c.MustGet("user").(*models.User)

	var review models.TradeReview
	if err := database.DB.First(&review, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "记录不存在"})
		return
	}
	if review.ReviewStatus == "approved" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "该记录已审核通过"})
		return
	}

	var trade models.TradeAbnormal
	if err := database.DB.First(&trade, review.TradeID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "关联订单不存在"})
		return
	}

	if trade.IsHandled && trade.HandleSource == "external" {
		database.DB.Model(&review).Updates(map[string]any{
			"review_status":    "approved",
			"reviewed_by_id":   user.ID,
			"reviewed_by_name": user.Realname,
			"reviewed_at":      time.Now(),
			"review_remark":    "同步已标记外部处理，自动关闭",
		})
		c.JSON(http.StatusOK, gin.H{"code": 200, "message": "该订单已由外部系统处理，审核已自动关闭"})
		return
	}

	now := time.Now()
	reviewerName := user.Realname
	if reviewerName == "" {
		reviewerName = user.Username
	}

	alreadyHandled, err := checkExternalHandled(&trade)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"code": 503, "message": "网络异常，无法确认订单状态，请稍后重试"})
		return
	}

	closeReview := func(remark string) {
		if review.ReviewRemark != "" {
			remark = review.ReviewRemark + "；" + remark
		}
		database.DB.Model(&trade).Updates(map[string]any{
			"is_handled":      true,
			"handled_by_id":   review.SubmittedByID,
			"handled_by_name": review.SubmittedByName,
			"handled_at":      now,
			"handle_duration": review.Duration,
			"handle_goods":    review.GoodsJSON,
			"handle_remark":   review.OperatorRemark,
			"review_status":   "",
			"handle_source":   "internal",
		})
		database.DB.Model(&review).Updates(map[string]any{
			"review_status":    "approved",
			"reviewed_by_id":   user.ID,
			"reviewed_by_name": reviewerName,
			"reviewed_at":      now,
			"review_remark":    remark,
		})
	}

	if alreadyHandled {
		closeReview("外部系统已处理，未外发")
		c.JSON(http.StatusOK, gin.H{"code": 200, "message": "该订单已由外部系统处理，已自动标记完成"})
		return
	}

	var extResult *externalResult
	if review.ActionType == "pend" {
		extResult, err = callExternalAPI("/pend", map[string]any{
			"id":         trade.TradeID,
			"pendStatus": "PENDING",
		})
	} else {
		var goods []GoodsItem
		json.Unmarshal([]byte(review.GoodsJSON), &goods)
		extResult, err = callExternalAPI("/handle", map[string]any{
			"orderGoodsDetailList": goods,
			"outOrderNo":           trade.OutOrderNo,
			"handleUsername":       "prisonProject",
		})
	}

	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"code": 502, "message": "外部接口网络异常: " + err.Error()})
		return
	}

	if !extResult.Success {
		closeReview("外部系统已处理，未外发")
		c.JSON(http.StatusOK, gin.H{"code": 200, "message": "外部系统已处理，已自动标记完成"})
		return
	}

	tradeUpdates := map[string]any{
		"is_handled":      true,
		"handled_by_id":   review.SubmittedByID,
		"handled_by_name": review.SubmittedByName,
		"handled_at":      now,
		"handle_duration": review.Duration,
		"handle_goods":    review.GoodsJSON,
		"handle_remark":   review.OperatorRemark,
		"review_status":   "",
		"handle_source":   "internal",
	}
	if review.ActionType == "pend" {
		tradeUpdates["pend_status"] = "PENDING"
		tradeUpdates["pend_status_desc"] = "已挂起"
	}
	database.DB.Model(&trade).Updates(tradeUpdates)

	reviewRemark := review.ReviewRemark
	database.DB.Model(&review).Updates(map[string]any{
		"review_status":    "approved",
		"reviewed_by_id":   user.ID,
		"reviewed_by_name": reviewerName,
		"reviewed_at":      now,
		"review_remark":    reviewRemark,
	})

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "审核通过，已提交外部系统"})
}

// AddReviewRemark 添加/更新备注（不改变审核状态）
func AddReviewRemark(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		Remark string `json:"remark"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	var review models.TradeReview
	if err := database.DB.First(&review, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "记录不存在"})
		return
	}

	database.DB.Model(&review).Update("review_remark", req.Remark)
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "备注已保存"})
}
