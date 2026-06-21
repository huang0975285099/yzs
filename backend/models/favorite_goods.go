package models

import "time"

// FavoriteGoods 用户收藏商品记录
type FavoriteGoods struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	UserID    uint64    `gorm:"index;not null" json:"userId"`
	GoodsID   uint64    `gorm:"index;not null" json:"goodsId"`
	Title     string    `gorm:"size:200;not null" json:"title"`
	Sn        string    `gorm:"size:30;not null" json:"sn"`
	FrontImg  string    `gorm:"size:512" json:"frontImg"`
	CreatedAt time.Time `json:"createdAt"`
}

func (FavoriteGoods) TableName() string {
	return "favorite_goods"
}