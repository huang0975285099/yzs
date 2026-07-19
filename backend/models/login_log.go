package models

import "time"

// LoginLog 登录日志（记录每次登录尝试，含成功/失败）
type LoginLog struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	UserID    uint      `gorm:"index" json:"userId"`        // 登录用户ID（失败时可能为0）
	Username  string    `gorm:"size:50;index" json:"username"` // 登录用户名输入
	IP        string    `gorm:"size:64;index" json:"ip"`     // 客户端IP（公网或内网，取决于部署）
	UA        string    `gorm:"size:500" json:"ua"`          // User-Agent
	Status    string    `gorm:"size:20;index" json:"status"` // success / failed
	Message   string    `gorm:"size:255" json:"message"`     // 失败原因等
	LoginAt   time.Time `gorm:"index" json:"loginAt"`
	CreatedAt time.Time `json:"createdAt"`
}
