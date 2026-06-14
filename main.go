package main

import (
	"go-xxl-admin/config"
	"go-xxl-admin/core"
	"go-xxl-admin/global"
	"go-xxl-admin/handlers"
	"go-xxl-admin/models"
	"go-xxl-admin/mq"
	"go-xxl-admin/redis"
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

	if config.Cfg.MQEnabled {
		if err := mq.Connect(config.Cfg.MQURL, config.Cfg.MQQueue); err != nil {
			log.Fatal("MQ链接失败", err)
		}
		go mq.StartConsumer(config.Cfg.MQQueue, core.RegsC.ElectNode)
	}

	if config.Cfg.RedisEnabled {
		if err := redis.Connect(config.Cfg.RedisAddr, config.Cfg.RedisPassword); err != nil {
			log.Fatal("Redis链接失败", err)
		}
	}

	// 2. 异步启动 Gin 服务
	r := gin.Default()
	r.StaticFile("/", "./web/index.html")
	r.Static("/web", "./web")
	r.GET("/api/ui-config", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"code": 200,
			"content": gin.H{
				"serverPort": config.Cfg.ServerPort,
				"appName": config.Cfg.AppName,
				"ginMode": config.Cfg.GinMode,
				"httpTimeout": config.Cfg.HTTPTimeout,
				"registryScanInterval": config.Cfg.RegistryScanInterval,
				"registryTimeout": config.Cfg.RegistryTimeout,
				"schedulerInterval": config.Cfg.SchedulerInterval,
				"mqEnabled": config.Cfg.MQEnabled,
				"redisEnabled": config.Cfg.RedisEnabled,
			},
		})
	})

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
