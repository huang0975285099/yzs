package models

import "time"

// HandledGoods 记录用户处理的商品明细
type HandledGoods struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	UserID      uint      `gorm:"index;not null" json:"userId"`
	Username    string    `gorm:"size:100" json:"username"`
	Realname    string    `gorm:"size:100" json:"realname"`
	TradeID     uint      `gorm:"index" json:"tradeId"`          // 关联的异常订单ID
	OutOrderNo  string    `gorm:"size:100" json:"outOrderNo"`    // 商户单号
	GoodsID     int64     `gorm:"index" json:"goodsId"`          // 商品ID
	GoodsName   string    `gorm:"size:200" json:"goodsName"`     // 商品名称
	GoodsPrice  float64   `json:"goodsPrice"`                    // 商品价格
	GoodsImage  string    `gorm:"size:500" json:"goodsImage"`    // 商品图片
	Type        int       `json:"type"`                          // 1=机器商品, 2=分公司商品
	GoodsCount  int       `json:"goodsCount"`                    // 数量
	Duration    int       `json:"duration"`                      // 作业时长(秒)
	Remark      string    `gorm:"size:500" json:"remark"`        // 备注
	CreatedAt   time.Time `json:"createdAt"`                     // 处理时间
}