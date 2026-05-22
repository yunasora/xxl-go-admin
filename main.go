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

		// 路由绑定
		r.POST("/api/registry", handlers.HandlerRegistry)
		r.POST("/api/callback", handlers.HandleCallBack) // 假设你放在了 handlers 中

		// ---- 今天新增的本地测试调试路由 ----
		r.POST("/test/kill", func(c *gin.Context) {
		})

		log.Println("[GO Admin] Server 正在 8081 端口监听...")
		if err := r.Run(":8081"); err != nil {
			log.Fatalf("服务器启动失败: %v", err)
		}
	}()

	// 3. 模拟生产调度流
	log.Println("[GO Admin] 启动成功，等待执行器心跳...")
	time.Sleep(35 * time.Second)

	targetAppName := "xxl-job-executor-sample"

	// 通过注册中心寻址后下发
	if addr, err := core.RegsC.ElectNode(targetAppName); err == nil {
		core.SendTrigger(addr)
	}

	select {} // 永久阻塞
}
