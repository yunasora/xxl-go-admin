package handlers

import (
	"go-xxl-admin/core"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func HandlerKillJob(c *gin.Context) {

	targetAppName := "xxl-job-executor-sample"

	go func() {
		log.Printf("[接口触发],开始为应用: %s,进行节点: 强杀", targetAppName)

		for {
			addr, err := core.RegsC.ElectNode(targetAppName)

			if err == nil {
				core.SendTrigger(targetAppName)
				time.Sleep(3 * time.Second)
				core.KillJob(addr, 1)
				log.Printf("成功为应用: %s 执行强杀", targetAppName)
				break
			} else {
				log.Printf("为能触发下发，原因是:%v", err)
				time.Sleep(3 * time.Second)
			}
		}
	}()
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    gin.H{"targetAppName": targetAppName},
	})
}
