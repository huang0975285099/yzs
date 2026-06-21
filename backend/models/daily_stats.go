package models

import "time"

// DailyStats 用户每日操作统计
type DailyStats struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	UserID      uint      `gorm:"uniqueIndex:idx_user_date;not null" json:"userId"`
	Date        string    `gorm:"uniqueIndex:idx_user_date;size:10;not null" json:"date"` // YYYY-MM-DD
	StartCount  int       `gorm:"default:0" json:"startCount"`  // 点击"开始处理"次数
	SkipCount   int       `gorm:"default:0" json:"skipCount"`   // 点击"跳过"次数
	SubmitCount int       `gorm:"default:0" json:"submitCount"` // 提交处理完成数量
	PendCount   int       `gorm:"default:0" json:"pendCount"`   // 挂起数量
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}