package main

import (
	"go-xxl-admin/core"
	"go-xxl-admin/handlers"
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	// 1. 启动内存巡逻清理协程
	core.RegsC.StartClearloop()

	// 2. 异步启动 Gin 服务
	go func() {
		gin.SetMode(gin.ReleaseMode)
		r := gin.Default()

		r.POST("/api/registry", handlers.HandlerRegistry)
		r.POST("/api/callback", handlers.HandleCallBack)

		r.POST("/test/kill", func(c *gin.Context) {
		})

		log.Println("[GO Admin] Server 正在 8081 端口监听...")
		if err := r.Run(":8081"); err != nil {
			log.Fatalf("服务器启动失败: %v", err)
		}
	}()

	log.Println("[GO Admin] 启动成功，等待执行器心跳...")
	time.Sleep(5 * time.Second)

	targetAppName := "xxl-job-executor-sample"

	// 通过注册中心寻址后下发
	for {
		if addr, err := core.RegsC.ElectNode(targetAppName); err == nil {
			core.SendTrigger(targetAppName)

			time.Sleep(3 * time.Second)
			core.KillJob(addr, 1)
			break
		} else {
			log.Printf("[联调提示] 没能触发下发，原因: %v", err)
			time.Sleep(3 * time.Second)
		}
	}
}
