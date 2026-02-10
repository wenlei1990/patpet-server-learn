package main

import (
	"log"
	"net/http"
	"time"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"patpet-server/handler"
	"patpet-server/middleware"
	"patpet-server/model"
)

func connectDB(dsn string) *gorm.DB {
	for i := 0; i < 30; i++ {
		db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err == nil {
			log.Println("✅ 数据库连接成功")
			return db
		}
		log.Printf("⏳ 数据库连接失败，%d秒后重试... (%v)", 2, err)
		time.Sleep(2 * time.Second)
	}
	log.Fatal("❌ 数据库连接失败，已达最大重试次数")
	return nil
}

func main() {
	cfg := LoadConfig()

	r := gin.Default()

	// 健康检查端点 —— 即使数据库还没连上也能响应，让平台知道进程已启动
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// 先在后台启动 HTTP 监听，让平台端口检测通过
	go func() {
		log.Printf("🚀 服务启动在 :%s", cfg.Port)
		if err := r.Run(":" + cfg.Port); err != nil {
			log.Fatal("HTTP 服务启动失败:", err)
		}
	}()

	// 然后连接数据库（带重试）
	db := connectDB(cfg.DatabaseURL)

	db.AutoMigrate(&model.User{})
	log.Println("✅ 数据库表迁移完成")

	// 初始化 Cloudinary
	var cld *cloudinary.Cloudinary
	if cfg.CloudinaryURL != "" {
		var err error
		cld, err = cloudinary.NewFromURL(cfg.CloudinaryURL)
		if err != nil {
			log.Printf("⚠️ Cloudinary 初始化失败: %v（头像上传功能不可用）", err)
		} else {
			log.Println("✅ Cloudinary 初始化成功")
		}
	} else {
		log.Println("⚠️ 未配置 CLOUDINARY_URL，头像上传功能不可用")
	}

	authHandler := &handler.AuthHandler{DB: db, JWTSecret: cfg.JWTSecret}
	profileHandler := &handler.ProfileHandler{DB: db}
	uploadHandler := &handler.UploadHandler{DB: db, Cloudinary: cld}

	r.POST("/api/v1/register", authHandler.Register)
	r.POST("/api/v1/login", authHandler.Login)

	auth := r.Group("/api/v1")
	auth.Use(middleware.JWTAuth(cfg.JWTSecret))
	{
		auth.GET("/profile", profileHandler.GetProfile)
		auth.PUT("/profile", profileHandler.UpdateProfile)
		auth.PUT("/password", profileHandler.ChangePassword)
		auth.POST("/avatar", uploadHandler.UploadAvatar)
	}

	log.Println("✅ 所有路由注册完成，服务就绪")

	// 阻塞主 goroutine
	select {}
}
