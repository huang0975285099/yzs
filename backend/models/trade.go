package models

import "time"

type TradeAbnormal struct {
	ID                uint      `gorm:"primarykey" json:"id"`
	TradeID           int64     `gorm:"uniqueIndex;not null" json:"tradeId"` // 外部 id，用于去重
	InnerCode         string    `gorm:"size:50" json:"innerCode"`
	OperatingModeDesc string    `gorm:"size:100" json:"operatingModeDesc"`
	VtDesc            string    `gorm:"size:100" json:"vtDesc"`
	AlgorithmDesc     string    `gorm:"size:100" json:"algorithmDesc"`
	NodeName          string    `gorm:"size:200" json:"nodeName"`
	CustomerID        string    `gorm:"size:50" json:"customerId"`
	OutOrderNo        string    `gorm:"size:100" json:"outOrderNo"`
	TransactionID     string    `gorm:"size:100" json:"transactionId"`
	Openid            string    `gorm:"size:100" json:"openid"`
	AbnormalTypeDesc  string    `gorm:"size:100" json:"abnormalTypeDesc"`
	AbnormalDesc      string    `gorm:"size:200" json:"abnormalDesc"`
	TradeStatusDesc   string    `gorm:"size:50" json:"tradeStatusDesc"`
	HandleStatusDesc  string    `gorm:"size:50" json:"handleStatusDesc"`
	CreateTime        string    `gorm:"size:30" json:"createTime"`
	DoorOpenTime      string    `gorm:"size:30" json:"doorOpenTime"`
	DoorCloseTime     string    `gorm:"size:30" json:"doorCloseTime"`
	AbnormalCode      string    `gorm:"size:20" json:"abnormalCode"`
	PendStatus        string     `gorm:"size:30" json:"pendStatus"`
	PendStatusDesc    string     `gorm:"size:50" json:"pendStatusDesc"`
	Handler           string     `gorm:"size:100" json:"handler"`
	SyncedAt          time.Time  `json:"syncedAt"`
	// 内部处理信息
	HandledByID   *uint      `gorm:"index" json:"handledById"`
	HandledByName string     `gorm:"size:100" json:"handledByName"`
	HandledAt     *time.Time `gorm:"index" json:"handledAt"`
	HandleRemark  string     `gorm:"size:500" json:"handleRemark"`
	IsHandled      bool       `gorm:"default:false;index" json:"isHandled"`
	HandleDuration int        `gorm:"default:0" json:"handleDuration"`
	HandleGoods    string     `gorm:"type:text" json:"handleGoods"`
	HandleSource   string     `gorm:"size:20;default:''" json:"handleSource"`
	// 视频时长（秒）NULL=未处理 0=无视频或获取失败 >0=实际秒数
	VideoDuration *int `json:"videoDuration"`
	// 处理锁
	LockedByID *uint      `gorm:"index" json:"lockedById"`
	LockedAt   *time.Time `json:"lockedAt"`
	// 质检状态（审核模式）
	ReviewStatus string `gorm:"size:20;default:''" json:"reviewStatus"` // '' 正常, 'pending' 待质检
	// 复查状态（直通模式，质检员事后复查）
	InspectStatus    string     `gorm:"size:20;default:''" json:"inspectStatus"`    // '' 未复查, 'normal' 正常, 'abnormal' 异常
	InspectRemark    string     `gorm:"size:500" json:"inspectRemark"`
	InspectedByID    uint       `gorm:"index" json:"inspectedById"`
	InspectedByName  string     `gorm:"size:100" json:"inspectedByName"`
	InspectedAt      *time.Time `json:"inspectedAt"`
}
