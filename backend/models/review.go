package models

import "time"

type TradeReview struct {
	ID              uint          `gorm:"primarykey" json:"id"`
	TradeID         uint          `gorm:"index;not null" json:"tradeId"`
	ActionType      string        `gorm:"size:10;not null" json:"actionType"` // submit, pend
	GoodsJSON       string        `gorm:"type:text" json:"goodsJson"`
	OperatorRemark  string        `gorm:"size:500" json:"operatorRemark"`
	Duration        int           `gorm:"default:0" json:"duration"`
	SubmittedByID   uint          `gorm:"index" json:"submittedById"`
	SubmittedByName string        `gorm:"size:100" json:"submittedByName"`
	SubmittedAt     time.Time     `gorm:"index" json:"submittedAt"`
	ReviewStatus    string        `gorm:"size:20;default:'pending'" json:"reviewStatus"` // pending, approved
	ReviewedByID    *uint         `json:"reviewedById"`
	ReviewedByName  string        `gorm:"size:100" json:"reviewedByName"`
	ReviewedAt      *time.Time    `json:"reviewedAt"`
	ReviewRemark    string        `gorm:"size:500" json:"reviewRemark"`
	Trade           TradeAbnormal `gorm:"-" json:"trade"`
}
