package scheduler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"go-yzs/database"
	"go-yzs/models"
	"log"
	"net/http"
	"time"

	"github.com/robfig/cron/v3"
	"gorm.io/gorm/clause"
)

const apiURL = "https://api.uboxol.com/lotus/trade/abnormal/page"
const pendingLockKey = "yzs:scheduler:sync:lock:pending"
const handledLockKey = "yzs:scheduler:sync:lock:handled"
const pendingLockTTL = 6 * time.Minute
const handledLockTTL = 12 * time.Minute
const lastSyncKey = "yzs:scheduler:last_sync"
const triggerGuardKey = "yzs:scheduler:trigger_guard"

// 共享 HTTP 客户端：复用连接，避免每个分页都重新握手 TLS。
var syncHTTPClient = &http.Client{
	Timeout: 30 * time.Second,
	Transport: &http.Transport{
		Proxy:               http.ProxyFromEnvironment,
		MaxIdleConns:        20,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
	},
}

type apiRequest struct {
	OperatingModeList []int  `json:"operatingModeList"`
	HandleStatus      string `json:"handleStatus"`
	PendStatus        string `json:"pendStatus"`
	Handler           string `json:"handler,omitempty"`
	Current           int    `json:"current"`
	Size              int    `json:"size"`
	StartCreateTime   string `json:"startCreateTime"`
	EndCreateTime     string `json:"endCreateTime"`
}

type apiRecord struct {
	ID                int64  `json:"id"`
	InnerCode         string `json:"innerCode"`
	OperatingModeDesc string `json:"operatingModeDesc"`
	VtDesc            string `json:"vtDesc"`
	AlgorithmDesc     string `json:"algorithmDesc"`
	NodeName          string `json:"nodeName"`
	CustomerID        string `json:"customerId"`
	OutOrderNo        string `json:"outOrderNo"`
	TransactionID     string `json:"transactionId"`
	Openid            string `json:"openid"`
	AbnormalTypeDesc  string `json:"abnormalTypeDesc"`
	AbnormalDesc      string `json:"abnormalDesc"`
	TradeStatusDesc   string `json:"tradeStatusDesc"`
	HandleStatusDesc  string `json:"handleStatusDesc"`
	CreateTime        string `json:"createTime"`
	DoorOpenTime      string `json:"doorOpenTime"`
	DoorCloseTime     string `json:"doorCloseTime"`
	AbnormalCode      string `json:"abnormalCode"`
	Handler           string `json:"handler"`
	PendStatus        string `json:"pendStatus"`
	PendStatusDesc    string `json:"pendStatusDesc"`
}

type apiResponse struct {
	Code    int  `json:"code"`
	Success bool `json:"success"`
	Data    struct {
		Records []apiRecord `json:"records"`
		Total   int         `json:"total"`
		Size    int         `json:"size"`
		Current int         `json:"current"`
		Pages   int         `json:"pages"`
	} `json:"data"`
}

func Start() {
	c := cron.New()
	// 待处理订单每 2 分钟拉一次，保证坐席取单的实时性；已处理回写每 30 分钟一轮。
	// 两者使用独立锁，避免回写批次把取单同步堵住。
	c.AddFunc("*/2 * * * *", func() { runPendingSync("pending") })
	c.AddFunc("*/30 * * * *", func() { runHandledSync("handled") })
	c.AddJob("5,35 * * * *", cron.NewChain(
		cron.SkipIfStillRunning(cron.DefaultLogger),
	).Then(cron.FuncJob(FillMissingVideoDurations)))

	c.Start()
	log.Println("[Scheduler] Started, pending sync every 2 minutes, handled sync every 30 minutes, video fill at :05 and :35")

	go func() {
		runPendingSync("initial")
		runHandledSync("initial")
	}()

	// 启动时立即补充一次视频时长（处理积压数据）
	go FillMissingVideoDurations()
}

// runSync 完整跑一轮（待处理 + 已处理），保留给手动调用。
func runSync(label string) {
	runPendingSync(label)
	runHandledSync(label)
}

func runPendingSync(label string) {
	withLock(pendingLockKey, pendingLockTTL, label+"/pending", syncPending)
}

func runHandledSync(label string) {
	withLock(handledLockKey, handledLockTTL, label+"/handled", func() error {
		syncHandled()
		return nil
	})
}

func withLock(key string, ttl time.Duration, label string, fn func() error) {
	if !acquireLock(key, ttl) {
		log.Printf("[Scheduler] %s sync skipped: another instance is running", label)
		return
	}
	defer releaseLock(key)

	started := time.Now()
	log.Printf("[Scheduler] %s sync start", label)
	if err := fn(); err != nil {
		log.Printf("[Scheduler] %s sync failed after %s: %v", label, time.Since(started).Round(time.Second), err)
		return
	}
	markSyncSuccess()
	log.Printf("[Scheduler] %s sync done in %s", label, time.Since(started).Round(time.Second))
}

// TriggerAsync 在订单池为空时由接口触发一次补同步。
// 用独立的 guard key 限频，避免坐席连续刷新打爆外部接口。
func TriggerAsync() {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	ok, err := database.RDB.SetNX(ctx, triggerGuardKey, "1", 2*time.Minute).Result()
	if err != nil || !ok {
		return
	}
	go runPendingSync("triggered")
}

func markSyncSuccess() {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_ = database.RDB.Set(ctx, lastSyncKey, time.Now().Format(time.RFC3339), 7*24*time.Hour).Err()
}

// LastSuccessTime 返回最近一次完整同步成功的时间；没有记录时返回零值。
func LastSuccessTime() time.Time {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	raw, err := database.RDB.Get(ctx, lastSyncKey).Result()
	if err != nil {
		return time.Time{}
	}
	t, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return time.Time{}
	}
	return t
}

func acquireLock(key string, ttl time.Duration) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	ok, err := database.RDB.SetNX(ctx, key, "1", ttl).Result()
	if err != nil {
		log.Printf("[Scheduler] Redis lock error: %v, proceeding without lock", err)
		return true
	}
	return ok
}

func releaseLock(key string) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	database.RDB.Del(ctx, key)
}

func SyncData() error {
	if err := syncPending(); err != nil {
		return err
	}
	syncHandled()
	return nil
}

// syncPending 只拉未处理订单，用于高频（2 分钟）补充待处理池。
func syncPending() error {
	startStr, endStr := syncWindow()

	round1Start := time.Now()
	n1, new1, err := syncPages(apiRequest{
		OperatingModeList: []int{21},
		HandleStatus:      "NOT_HANDLED",
		PendStatus:        "NO_PENDING",
		StartCreateTime:   startStr,
		EndCreateTime:     endStr,
	}, insertIfNew)
	if err != nil {
		return fmt.Errorf("sync pending error: %w", err)
	}
	log.Printf("[Scheduler] Round1 pending: total=%d new=%d elapsed=%s", n1, new1, time.Since(round1Start).Round(time.Millisecond))
	return nil
}

// syncHandled 回写已处理订单，频率较低（30 分钟）。
func syncHandled() {
	startStr, endStr := syncWindow()

	round2Start := time.Now()
	n2, upd2, err := syncPages(apiRequest{
		OperatingModeList: []int{21},
		HandleStatus:      "CUSTOMER_SERVICE_HANDLED",
		PendStatus:        "NO_PENDING",
		Handler:           "prisonProject",
		StartCreateTime:   startStr,
		EndCreateTime:     endStr,
	}, upsertHandled)
	if err != nil {
		log.Printf("[Scheduler] Round2 handled error: %v", err)
		return
	}
	log.Printf("[Scheduler] Round2 handled: total=%d changed=%d elapsed=%s", n2, upd2, time.Since(round2Start).Round(time.Millisecond))
}

func syncWindow() (string, string) {
	now := time.Now()
	start := now.AddDate(0, 0, -7)
	return start.Format("2006-01-02") + " 00:00:00", now.Format("2006-01-02") + " 23:59:59"
}

func syncPages(baseReq apiRequest, handle func(apiRecord) (bool, error)) (int, int, error) {
	count := 0
	current := 1
	pageSize := 100
	total := 0
	pages := 0
	failedPages := 0
	consecutiveFailures := 0
	deadline := time.Now().Add(8 * time.Minute)

	for {
		if time.Now().After(deadline) {
			log.Printf("[Scheduler] round exceeded 8m deadline at page %d, stopping early", current)
			break
		}
		baseReq.Current = current
		baseReq.Size = pageSize

		resp, err := fetchPage(baseReq)
		if err != nil {
			failedPages++
			consecutiveFailures++
			log.Printf("[Scheduler] fetchPage error at page %d: %v", current, err)
			// 第一页失败时拿不到总页数，本轮直接返回；后续页失败则跳过继续，
			// 避免单页抖动导致整轮同步（包括后面的已处理回写）全部中止。
			if pages == 0 {
				return total, count, fmt.Errorf("fetchPage error at page %d: %w", current, err)
			}
			// 连续失败说明外部接口整体不可用，立刻收手，避免 170 页 × 3 次重试拖垮整轮。
			if consecutiveFailures >= 5 {
				log.Printf("[Scheduler] abort round after %d consecutive page failures", consecutiveFailures)
				break
			}
			if current >= pages {
				break
			}
			current++
			continue
		}
		consecutiveFailures = 0
		pages = resp.Data.Pages
		total = resp.Data.Total

		for _, r := range resp.Data.Records {
			changed, err := handle(r)
			if err != nil {
				log.Printf("[Scheduler] handle error for id=%d: %v", r.ID, err)
				continue
			}
			if changed {
				count++
			}
		}

		if current >= resp.Data.Pages || len(resp.Data.Records) == 0 {
			break
		}
		current++
	}
	if failedPages > 0 {
		log.Printf("[Scheduler] finished with %d failed page(s), processed %d records", failedPages, count)
	}
	return total, count, nil
}

func insertIfNew(r apiRecord) (bool, error) {
	record := models.TradeAbnormal{
		TradeID:           r.ID,
		InnerCode:         r.InnerCode,
		OperatingModeDesc: r.OperatingModeDesc,
		VtDesc:            r.VtDesc,
		AlgorithmDesc:     r.AlgorithmDesc,
		NodeName:          r.NodeName,
		CustomerID:        r.CustomerID,
		OutOrderNo:        r.OutOrderNo,
		TransactionID:     r.TransactionID,
		Openid:            r.Openid,
		AbnormalTypeDesc:  r.AbnormalTypeDesc,
		AbnormalDesc:      r.AbnormalDesc,
		TradeStatusDesc:   r.TradeStatusDesc,
		HandleStatusDesc:  r.HandleStatusDesc,
		CreateTime:        r.CreateTime,
		DoorOpenTime:      r.DoorOpenTime,
		DoorCloseTime:     r.DoorCloseTime,
		AbnormalCode:      r.AbnormalCode,
		Handler:           r.Handler,
		PendStatus:        r.PendStatus,
		PendStatusDesc:    r.PendStatusDesc,
		SyncedAt:          time.Now(),
	}
	result := database.DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "trade_id"}},
		DoNothing: true,
	}).Create(&record)
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

func upsertHandled(r apiRecord) (bool, error) {
	var existing models.TradeAbnormal
	err := database.DB.Where("trade_id = ?", r.ID).First(&existing).Error

	now := time.Now()

	if err == nil {
		if existing.HandleSource == "internal" {
			// 内部处理的订单不再每轮刷新 synced_at：7 天窗口内有一两万条，
			// 逐条 UPDATE 会制造大量无意义写入并拖慢整轮同步。
			return false, nil
		}

		updates := map[string]any{
			"trade_status_desc":  r.TradeStatusDesc,
			"handle_status_desc": r.HandleStatusDesc,
			"transaction_id":     r.TransactionID,
			"synced_at":          now,
		}

		if existing.PendStatusDesc == "" {
			updates["pend_status"] = r.PendStatus
			updates["pend_status_desc"] = r.PendStatusDesc
		}

		if !existing.IsHandled {
			updates["is_handled"] = true
			updates["handled_by_id"] = nil
			updates["handled_by_name"] = "外部系统"
			updates["handled_at"] = now
			updates["handle_source"] = "external"
			updates["review_status"] = ""
		}

		database.DB.Model(&existing).Updates(updates)
		return true, nil
	}

	record := models.TradeAbnormal{
		TradeID:           r.ID,
		InnerCode:         r.InnerCode,
		OperatingModeDesc: r.OperatingModeDesc,
		VtDesc:            r.VtDesc,
		AlgorithmDesc:     r.AlgorithmDesc,
		NodeName:          r.NodeName,
		CustomerID:        r.CustomerID,
		OutOrderNo:        r.OutOrderNo,
		TransactionID:     r.TransactionID,
		Openid:            r.Openid,
		AbnormalTypeDesc:  r.AbnormalTypeDesc,
		AbnormalDesc:      r.AbnormalDesc,
		TradeStatusDesc:   r.TradeStatusDesc,
		HandleStatusDesc:  r.HandleStatusDesc,
		CreateTime:        r.CreateTime,
		DoorOpenTime:      r.DoorOpenTime,
		DoorCloseTime:     r.DoorCloseTime,
		AbnormalCode:      r.AbnormalCode,
		Handler:           r.Handler,
		PendStatus:        r.PendStatus,
		PendStatusDesc:    r.PendStatusDesc,
		SyncedAt:          now,
		IsHandled:         true,
		HandledByName:     "外部系统",
		HandledAt:         &now,
		HandleSource:      "external",
	}
	result := database.DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "trade_id"}},
		DoNothing: true,
	}).Create(&record)
	if result.Error != nil {
		return false, result.Error
	}
	if result.RowsAffected == 0 {
		return upsertHandled(r)
	}
	return true, nil
}

func fetchPage(req apiRequest) (*apiResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	var lastErr error
	for attempt := 1; attempt <= 3; attempt++ {
		if attempt > 1 {
			time.Sleep(time.Duration(attempt-1) * time.Second)
		}
		httpResp, err := syncHTTPClient.Post(apiURL, "application/json", bytes.NewReader(body))
		if err != nil {
			lastErr = err
			continue
		}

		var resp apiResponse
		decodeErr := json.NewDecoder(httpResp.Body).Decode(&resp)
		httpResp.Body.Close()
		if decodeErr != nil {
			lastErr = decodeErr
			continue
		}
		if !resp.Success {
			lastErr = fmt.Errorf("API returned success=false")
			continue
		}
		return &resp, nil
	}
	return nil, lastErr
}
