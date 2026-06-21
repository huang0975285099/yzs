package main

import (
	"go-yzs/config"
	"go-yzs/database"
	"go-yzs/routes"
	"go-yzs/scheduler"
	"log"

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
	routes.Setup(r)

	addr := ":" + config.App.ServerPort
	log.Printf("Server starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
