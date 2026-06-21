package database

import (
	"fmt"
	"go-yzs/config"
	"go-yzs/models"
	"log"
	"strings"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Init() {
	// 先不指定数据库名连接，确保数据库存在
	rootDSN := fmt.Sprintf("%s:%s@tcp(%s:%s)/?charset=utf8mb4&parseTime=True&loc=Local",
		config.App.DBUser,
		config.App.DBPassword,
		config.App.DBHost,
		config.App.DBPort,
	)
	rootDB, err := gorm.Open(mysql.Open(rootDSN), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to MySQL: %v", err)
	}
	createSQL := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci", config.App.DBName)
	if err := rootDB.Exec(createSQL).Error; err != nil {
		log.Fatalf("Failed to create database: %v", err)
	}
	sqlDB, _ := rootDB.DB()
	sqlDB.Close()

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.App.DBUser,
		config.App.DBPassword,
		config.App.DBHost,
		config.App.DBPort,
		config.App.DBName,
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// 配置连接池，避免空闲连接被 MySQL 服务器单侧关闭后出现 unexpected EOF
	sqlDB, err = db.DB()
	if err != nil {
		log.Fatalf("Failed to get sql.DB: %v", err)
	}
	sqlDB.SetMaxOpenConns(20)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(4 * time.Hour) // 小于 MySQL wait_timeout（默认 8h）
	sqlDB.SetConnMaxIdleTime(10 * time.Minute)

	// Auto migrate tables — migrate one by one to tolerate the MySQL
	// information_schema slow-query bug that causes a false "table not found"
	// check followed by a CREATE TABLE that fails with error 1050.
	for _, m := range []interface{}{
		&models.Team{},
		&models.User{},
		&models.UserSession{},
		&models.TradeAbnormal{},
		&models.DailyStats{},
		&models.TradeReview{},
		&models.HandledGoods{},
		&models.FavoriteGoods{},
	} {
		if err = db.AutoMigrate(m); err != nil {
			if strings.Contains(err.Error(), "1050") || strings.Contains(err.Error(), "already exists") {
				log.Printf("[Migrate] table already exists, skipping: %T", m)
				continue
			}
			log.Fatalf("Failed to migrate %T: %v", m, err)
		}
	}

	DB = db
	log.Println("Database connected and migrated successfully")

	migrateHandleSource(db)

	createDefaultAdmin(db)
}

func createDefaultAdmin(db *gorm.DB) {
	var count int64
	db.Model(&models.User{}).Count(&count)
	if count == 0 {
		admin := models.User{
			Username: "admin",
			Realname: "管理员",
			Role:     "admin",
		}
		admin.SetPassword("mzjy.com")
		db.Create(&admin)
		log.Println("Default admin user created: admin / mzjy.com")
	}
}

func migrateHandleSource(db *gorm.DB) {
	var count int64
	db.Model(&models.TradeAbnormal{}).Where("handle_source = '' AND is_handled = 1").Count(&count)
	if count == 0 {
		fixOverwrittenRecords(db)
		return
	}

	result := db.Exec(`
		UPDATE trade_abnormals
		SET handle_source = CASE
			WHEN handled_by_name = '外部系统' THEN 'external'
			ELSE 'internal'
		END
		WHERE is_handled = 1 AND handle_source = ''
	`)
	if result.Error != nil {
		log.Printf("[Migration] handle_source backfill error: %v", result.Error)
		return
	}
	log.Printf("[Migration] handle_source backfilled %d rows", result.RowsAffected)

	db.Exec(`
		UPDATE trade_abnormals
		SET handle_source = 'internal'
		WHERE is_handled = 1 AND handle_source = 'external' AND handle_goods != ''
	`)

	fixOverwrittenRecords(db)
}

func fixOverwrittenRecords(db *gorm.DB) {
	result := db.Exec(`
		UPDATE trade_abnormals t
		JOIN users u ON t.handled_by_id = u.id
		SET t.handled_by_name = u.realname,
		    t.handle_source = 'internal'
		WHERE t.handled_by_name = '外部系统'
		  AND t.handled_by_id IS NOT NULL
	`)
	if result.Error != nil {
		log.Printf("[Migration] fix overwritten records error: %v", result.Error)
		return
	}
	if result.RowsAffected > 0 {
		log.Printf("[Migration] fixed %d overwritten records: restored operator names from users table", result.RowsAffected)
	}

	result = db.Exec(`
		UPDATE trade_abnormals
		SET handle_source = 'internal'
		WHERE handled_by_name = '外部系统'
		  AND handled_by_id IS NULL
		  AND handle_goods != ''
	`)
	if result.Error != nil {
		log.Printf("[Migration] fix orphan goods records error: %v", result.Error)
		return
	}
	if result.RowsAffected > 0 {
		log.Printf("[Migration] fixed %d records: handle_source corrected to internal (goods present but operator lost)", result.RowsAffected)
	}
}
