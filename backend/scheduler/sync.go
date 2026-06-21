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
const syncLockKey = "yzs:scheduler:sync:lock"
const syncLockTTL = 25 * time.Minute

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
	c.AddFunc("*/30 * * * *", func() {
		if !acquireSyncLock() {
			log.Println("[Scheduler] Another instance is syncing, skipping")
			return
		}
		defer releaseSyncLock()
		log.Println("[Scheduler] Starting data sync...")
		if err := SyncData(); err != nil {
			log.Printf("[Scheduler] Sync error: %v", err)
		}
	})
	c.AddJob("5,35 * * * *", cron.NewChain(
		cron.SkipIfStillRunning(cron.DefaultLogger),
	).Then(cron.FuncJob(FillMissingVideoDurations)))

	c.Start()
	log.Println("[Scheduler] Started, sync every 30 minutes, video fill at :05 and :35")

	go func() {
		if !acquireSyncLock() {
			log.Println("[Scheduler] Another instance is syncing, skipping initial sync")
			return
		}
		defer releaseSyncLock()
		log.Println("[Scheduler] Running initial sync...")
		if err := SyncData(); err != nil {
			log.Printf("[Scheduler] Initial sync error: %v", err)
		}
	}()

	// 启动时立即补充一次视频时长（处理积压数据）
	go FillMissingVideoDurations()
}

func acquireSyncLock() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	ok, err := database.RDB.SetNX(ctx, syncLockKey, "1", syncLockTTL).Result()
	if err != nil {
		log.Printf("[Scheduler] Redis lock error: %v, proceeding without lock", err)
		return true
	}
	return ok
}

func releaseSyncLock() {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	database.RDB.Del(ctx, syncLockKey)
}

func SyncData() error {
	now := time.Now()
	start := now.AddDate(0, 0, -7)
	startStr := start.Format("2006-01-02") + " 00:00:00"
	endStr := now.Format("2006-01-02") + " 23:59:59"

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
	log.Printf("[Scheduler] Round1 pending: total=%d new=%d", n1, new1)

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
	} else {
		log.Printf("[Scheduler] Round2 handled: total=%d changed=%d", n2, upd2)
	}

	return nil
}

func syncPages(baseReq apiRequest, handle func(apiRecord) (bool, error)) (int, int, error) {
	count := 0
	current := 1
	pageSize := 100
	total := 0

	for {
		baseReq.Current = current
		baseReq.Size = pageSize

		resp, err := fetchPage(baseReq)
		if err != nil {
			return 0, 0, fmt.Errorf("fetchPage error at page %d: %w", current, err)
		}
		if !resp.Success {
			return 0, 0, fmt.Errorf("API returned success=false at page %d", current)
		}
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
			database.DB.Model(&existing).Update("synced_at", now)
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
	client := &http.Client{Timeout: 30 * time.Second}
	httpResp, err := client.Post(apiURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	var resp apiResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
