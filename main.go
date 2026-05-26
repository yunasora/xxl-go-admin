package main

import (
	"go-xxl-admin/core"
	"go-xxl-admin/global"
	"go-xxl-admin/handlers"
	"go-xxl-admin/models"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {

	log.Println("[Go Admin] 正在启动基础服务.......")

	global.InitDB("xxl_job.db")

	err := global.DB.AutoMigrate(
		&models.JobRegistry{},
		&models.JobInfo{},
		&models.JobLog{},
	)

	if err != nil {
		log.Fatal("数据库自动建表失败", err)
	}

	log.Println("数据库的三张表registry、info、log创建成功")
	// 1. 启动内存巡逻清理协程
	core.RegsC.StartClearloop()

	// 2. 异步启动 Gin 服务
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	r.POST("/api/registry", handlers.HandlerRegistry)
	r.POST("/api/callback", handlers.HandleCallBack)
	r.POST("/test/kill", handlers.HandlerKillJob)

	log.Println("[GO Admin] Server 正在 8081 端口监听...")
	if err := r.Run(":8081"); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}

}
