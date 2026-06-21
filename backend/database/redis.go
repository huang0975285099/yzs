package database

import (
	"context"
	"encoding/json"
	"fmt"
	"go-yzs/config"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

var RDB *redis.Client

const (
	sessionKeyPrefix  = "yzs:session:"   // yzs:session:{token} → CachedSession JSON
	userTokenKeyPrefix = "yzs:user_token:" // yzs:user_token:{userID} → token string（用于 SSO 踢出旧 session）
)

// CachedSession 是存入 Redis 的 session 缓存结构
type CachedSession struct {
	SessionID uint      `json:"sessionId"`
	UserID    uint      `json:"userId"`
	Username  string    `json:"username"`
	Realname  string    `json:"realname"`
	Role      string    `json:"role"`
	ExpiredAt time.Time `json:"expiredAt"`
}

func InitRedis() {
	RDB = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", config.App.RedisHost, config.App.RedisPort),
		Password: config.App.RedisPassword,
		DB:       0,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := RDB.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	log.Println("Redis connected successfully")
}

// SessionKey 返回 session 缓存 key
func SessionKey(token string) string {
	return sessionKeyPrefix + token
}

// UserTokenKey 返回用户 token 索引 key（用于 SSO 踢出）
func UserTokenKey(userID uint) string {
	return fmt.Sprintf("%s%d", userTokenKeyPrefix, userID)
}

// SetSession 将 session 写入 Redis，TTL 为 expiredAt 距今的时间
func SetSession(ctx context.Context, token string, s CachedSession) error {
	ttl := time.Until(s.ExpiredAt)
	if ttl <= 0 {
		return nil
	}
	data, err := json.Marshal(s)
	if err != nil {
		return err
	}
	pipe := RDB.Pipeline()
	pipe.Set(ctx, SessionKey(token), data, ttl)
	pipe.Set(ctx, UserTokenKey(s.UserID), token, ttl)
	_, err = pipe.Exec(ctx)
	return err
}

// GetSession 从 Redis 读取 session，未命中返回 nil
func GetSession(ctx context.Context, token string) *CachedSession {
	data, err := RDB.Get(ctx, SessionKey(token)).Bytes()
	if err != nil {
		return nil
	}
	var s CachedSession
	if err := json.Unmarshal(data, &s); err != nil {
		return nil
	}
	return &s
}

// DeleteSession 删除指定 token 的 Redis session
func DeleteSession(ctx context.Context, token string, userID uint) {
	RDB.Del(ctx, SessionKey(token), UserTokenKey(userID))
}

// DeleteUserSession 通过 userID 找到旧 token 并删除（SSO 登录时踢出旧 session）
func DeleteUserSession(ctx context.Context, userID uint) {
	oldToken, err := RDB.Get(ctx, UserTokenKey(userID)).Result()
	if err == nil && oldToken != "" {
		RDB.Del(ctx, SessionKey(oldToken))
	}
	RDB.Del(ctx, UserTokenKey(userID))
}
