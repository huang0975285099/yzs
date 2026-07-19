package main

import (
	"go-yzs/config"
	"go-yzs/database"
	"go-yzs/routes"
	"go-yzs/scheduler"
	"log"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("未找到 .env 文件，使用系统环境变量")
	}
	config.Init()
	database.Init()
	database.InitRedis()
	scheduler.Start()

	r := gin.Default()

	// 配置信任的代理 IP，用于从 X-Forwarded-For 正确解析客户端真实 IP
	// TRUSTED_PROXIES 为空：不信任任何代理，c.ClientIP() 返回直连 IP
	// TRUSTED_PROXIES=*：信任所有代理（仅开发环境推荐，有伪造风险）
	// TRUSTED_PROXIES=127.0.0.1,10.0.0.0/8：仅信任特定代理（生产推荐）
	if config.App.TrustedProxies == "*" {
		_ = r.SetTrustedProxies([]string{"*"})
	} else if config.App.TrustedProxies != "" {
		proxies := strings.Split(config.App.TrustedProxies, ",")
		for i := range proxies {
			proxies[i] = strings.TrimSpace(proxies[i])
		}
		_ = r.SetTrustedProxies(proxies)
	} else {
		_ = r.SetTrustedProxies(nil) // 不信任任何代理
	}

	routes.Setup(r)

	addr := ":" + config.App.ServerPort
	log.Printf("Server starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
