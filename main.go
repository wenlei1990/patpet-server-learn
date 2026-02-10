package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"patpet-server/handler"
	"patpet-server/middleware"
	"patpet-server/model"
)

func main() {
	cfg := LoadConfig()

	// 先启动 HTTP 服务并监听端口，再连数据库（避免 DB 连不上时完全连不上服务）
	r := gin.Default()
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Fatal("数据库连接失败:", err)
	}
	log.Println("✅ 数据库连接成功")

	db.AutoMigrate(&model.User{})
	log.Println("✅ 数据库表迁移完成")

	authHandler := &handler.AuthHandler{DB: db, JWTSecret: cfg.JWTSecret}
	profileHandler := &handler.ProfileHandler{DB: db}

	r.POST("/api/v1/register", authHandler.Register)
	r.POST("/api/v1/login", authHandler.Login)

	auth := r.Group("/api/v1")
	auth.Use(middleware.JWTAuth(cfg.JWTSecret))
	{
		auth.GET("/profile", profileHandler.GetProfile)
	}

	log.Printf("🚀 服务启动在 :%s", cfg.Port)
	r.Run(":" + cfg.Port)
}
