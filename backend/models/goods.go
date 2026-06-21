package models

import "time"

// Goods 商品表
type Goods struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	Gid       uint64    `gorm:"index" json:"gid"`
	Title     string    `gorm:"size:200;not null" json:"title"`
	Sn        string    `gorm:"size:30;not null" json:"sn"`
	FrontImg  string    `gorm:"size:512;not null" json:"frontImg"`
	BackImg   string    `gorm:"size:512;not null" json:"backImg"`
	LeftImg   string    `gorm:"size:512;not null" json:"leftImg"`
	RightImg  string    `gorm:"size:512;not null" json:"rightImg"`
	TopImg    string    `gorm:"size:512;not null" json:"topImg"`
	BottomImg string    `gorm:"size:512;not null" json:"bottomImg"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	DeletedAt time.Time `json:"deletedAt"`
}

func (Goods) TableName() string {
	return "goods"
}