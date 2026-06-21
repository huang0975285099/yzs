package models

import (
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	ID         uint           `gorm:"primarykey" json:"id"`
	Username   string         `gorm:"uniqueIndex;size:50;not null" json:"username"`
	Password   string         `gorm:"size:255;not null" json:"-"`
	Realname   string         `gorm:"size:100" json:"realname"`
	Role       string         `gorm:"size:20;default:'operator'" json:"role"` // admin, statistician, operator
	TeamID     *uint          `gorm:"index" json:"teamId"`
	Team       *Team          `gorm:"foreignKey:TeamID" json:"team,omitempty"`
	StartCount int            `gorm:"default:0" json:"startCount"`
	SkipCount  int            `gorm:"default:0" json:"skipCount"`
	CreatedAt  time.Time      `json:"createdAt"`
	UpdatedAt  time.Time      `json:"updatedAt"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

type UserSession struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	UserID    uint      `gorm:"index;not null" json:"userId"`
	Token     string    `gorm:"uniqueIndex;size:500" json:"token"`
	DeviceKey string    `gorm:"size:255" json:"deviceKey"`
	ExpiredAt time.Time `json:"expiredAt"`
	CreatedAt time.Time `json:"createdAt"`
}

func (u *User) SetPassword(password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hash)
	return nil
}

func (u *User) CheckPassword(password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)) == nil
}
