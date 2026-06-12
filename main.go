package main

import (
	"go-xxl-admin/config"
	"go-xxl-admin/core"
	"go-xxl-admin/global"
	"go-xxl-admin/handlers"
	"go-xxl-admin/models"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	if err := config.Load("config.json"); err != nil {
		log.Fatal("配置加载失败:", err)
	}

	log.Println("[Go Admin] 正在启动基础服务.......")

	global.InitDB(config.Cfg.DBPath)

	if config.Cfg.GinMode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	err := global.DB.AutoMigrate(
		&models.JobRegistry{},
		&models.JobInfo{},
		&models.JobLog{},
		&models.JobGroup{},
	)

	if err != nil {
		log.Fatal("数据库自动建表失败", err)
	}

	log.Println("数据库的三张表registry、info、log创建成功")
	// 1. 启动内存巡逻清理协程
	core.RegsC.StartClearloop()
	core.Sched.Start()
	// 2. 异步启动 Gin 服务
	r := gin.Default()

	r.POST("/api/registry", handlers.HandlerRegistry)
	r.POST("/api/callback", handlers.HandleCallBack)
	r.POST("/test/kill", handlers.HandlerKillJob)
	r.POST("/api/job", handlers.CreateJob)
	r.GET("/api/job", handlers.ListJobs)
	r.GET("/api/job/:id", handlers.GetJob)
	r.PUT("/api/job/:id", handlers.UpdateJob)
	r.DELETE("/api/job/:id", handlers.DeleteJob)
	r.PUT("/api/job/:id/start", handlers.StartJob)
	r.PUT("/api/job/:id/stop", handlers.StopJob)
	log.Printf("[GO Admin] Server 正在 %s 端口监听...", config.Cfg.ServerPort)

	if err := r.Run(config.Cfg.ServerPort); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}

}
