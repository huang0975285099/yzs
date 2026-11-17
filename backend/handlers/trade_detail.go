package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go-yzs/database"
	"go-yzs/models"
	"go-yzs/scheduler"
	"io"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

const externalDetailURL = "https://api.uboxol.com/lotus/trade/abnormal/detail"
const externalProductURL = "https://api.uboxol.com/lotus/product/vmProductList"
const externalBranchProductURL = "https://api.uboxol.com/lotus/product/queryProductSale"
const externalProductPriceURL = "https://api.uboxol.com/lotus/product/queryProductPrice"

// GetRandomUnhandled 从最早的20条未处理且未被他人锁定的订单中随机返回1条的 id
func GetRandomUnhandled(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	expiry := time.Now().Add(-lockTTL)

	var trades []models.TradeAbnormal
	err := database.DB.
		Where(`is_handled = ? AND review_status = '' AND (
			locked_by_id IS NULL OR
			locked_by_id = ? OR
			locked_at < ?
		)`, false, user.ID, expiry).
		Order("create_time ASC"). // 取最早的20条
		Limit(20).
		Find(&trades).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询待处理订单失败"})
		return
	}
	if len(trades) == 0 {
		respondEmptyPool(c)
		return
	}
	trade := trades[rand.Intn(len(trades))]
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": gin.H{"id": trade.ID}})
}

// respondEmptyPool 返回可解释的“空池”状态，而不是笼统的 404。
// 如果本地长时间没有同步成功，会异步触发一次补同步，避免坐席一直空转。
func respondEmptyPool(c *gin.Context) {
	// 三条计数使用与取单相同的事实条件，保证 meta 与空池语义一致。
	// 均为局部变量：并发请求互不影响。
	var visible, locked, pendingReview int64
	database.DB.Model(&models.TradeAbnormal{}).
		Where("is_handled = ? AND review_status = ''", false).Count(&visible)
	database.DB.Model(&models.TradeAbnormal{}).
		Where("is_handled = ? AND review_status = '' AND locked_by_id IS NOT NULL AND locked_at >= ?", false, time.Now().Add(-lockTTL)).
		Count(&locked)
	database.DB.Model(&models.TradeAbnormal{}).
		Where("review_status = ?", "pending").Count(&pendingReview)

	lastSync := scheduler.LastSuccessTime()
	lastSyncText := ""
	stale := true
	if !lastSync.IsZero() {
		lastSyncText = lastSync.Format("2006-01-02 15:04:05")
		stale = time.Since(lastSync) > 5*time.Minute
	}

	// Redis 不可读时 stale 恒为 true，此时退化为“总是尝试触发”，
	// guard key 同样写在 Redis，也会一并失败，TriggerAsync 自身会静默放弃。
	triggered := visible == 0 && stale
	if triggered {
		scheduler.TriggerAsync()
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "暂无待处理订单",
		"data":    nil,
		"meta": gin.H{
			"unhandled":     visible,
			"locked":        locked,
			"pendingReview": pendingReview,
			"lastSyncTime":  lastSyncText,
			"syncTriggered": triggered,
		},
	})
}

// GetTradeDetail 代理调用外部 detail 接口 + 商品列表接口，合并返回
func GetTradeDetail(c *gin.Context) {
	id := c.Param("id")

	var trade models.TradeAbnormal
	if err := database.DB.First(&trade, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "记录不存在"})
		return
	}

	// 调外部 detail 接口
	detail, err := fetchDetail(trade.TradeID)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"code": 502, "message": "获取订单详情失败: " + err.Error()})
		return
	}

	// 用 innerCode 调商品列表
	innerCode, _ := detail["innerCode"].(string)
	products, err := fetchProducts(innerCode)
	if err != nil {
		products = []any{} // 商品列表失败不阻断，返回空
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"detail":          detail,
			"products":        products,
			"handleGoods":     trade.HandleGoods,
			"pendStatus":      trade.PendStatus,
			"handleRemark":    trade.HandleRemark,
			"inspectStatus":   trade.InspectStatus,
			"inspectRemark":   trade.InspectRemark,
			"inspectedByName": trade.InspectedByName,
			"inspectedAt":     trade.InspectedAt,
		},
	})
}

// QueryBranchProducts 搜索分公司商品库
func QueryBranchProducts(c *gin.Context) {
	id := c.Param("id")
	keyword := c.Query("keyword")

	var trade models.TradeAbnormal
	if err := database.DB.First(&trade, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "记录不存在"})
		return
	}

	resp, err := doPost(externalBranchProductURL, map[string]any{
		"keyCode": keyword,
		"limit":   10,
		"vmCode":  trade.InnerCode,
	})
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"code": 502, "message": "搜索失败: " + err.Error()})
		return
	}
	defer resp.Body.Close()

	var result struct {
		Code    int   `json:"code"`
		Success bool  `json:"success"`
		Data    []any `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"code": 502, "message": "解析响应失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": result.Data})
}

// QueryProductPrice 查询商品在指定机器上的实际价格
func QueryProductPrice(c *gin.Context) {
	id := c.Param("id")

	var trade models.TradeAbnormal
	if err := database.DB.First(&trade, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "记录不存在"})
		return
	}

	var req struct {
		ProductID int64 `json:"productId"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.ProductID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	resp, err := doPost(externalProductPriceURL, map[string]any{
		"productId": req.ProductID,
		"vmCode":    trade.InnerCode,
	})
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"code": 502, "message": "查询价格失败: " + err.Error()})
		return
	}
	defer resp.Body.Close()

	var result struct {
		Code    int  `json:"code"`
		Success bool `json:"success"`
		Data    struct {
			ProductPrice  int64  `json:"productPrice"`
			PriceTypeName string `json:"priceTypeName"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"code": 502, "message": "解析价格响应失败"})
		return
	}
	// 分转元
	priceYuan := float64(result.Data.ProductPrice) / 100.0
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"productPrice":  priceYuan,
			"priceTypeName": result.Data.PriceTypeName,
		},
	})
}

var extClient = &http.Client{Timeout: 30 * time.Second}

func doPost(url string, payload any) (*http.Response, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0")
	return extClient.Do(req)
}

func doGet(url string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")
	return extClient.Do(req)
}

func fetchDetail(tradeID int64) (map[string]any, error) {
	resp, err := doPost(externalDetailURL, map[string]any{"id": tradeID})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	log.Printf("[fetchDetail] tradeID=%d status=%d body=%s", tradeID, resp.StatusCode, string(raw))
	var result struct {
		Code    int            `json:"code"`
		Success bool           `json:"success"`
		Data    map[string]any `json:"data"`
	}
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w (body: %s)", err, string(raw))
	}
	if !result.Success {
		return nil, fmt.Errorf("detail API 返回 success=false (body: %s)", string(raw))
	}
	return result.Data, nil
}

func fetchProducts(innerCode string) ([]any, error) {
	rawURL := fmt.Sprintf("%s?vmId=%s", externalProductURL, innerCode)
	resp, err := doGet(rawURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Success bool  `json:"success"`
		Data    []any `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	if !result.Success {
		return nil, fmt.Errorf("product API 返回 success=false")
	}
	return result.Data, nil
}
