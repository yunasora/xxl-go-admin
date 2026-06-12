package handlers

import (
	"fmt"
	"go-xxl-admin/global"
	"go-xxl-admin/models"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func HandleCallBack(c *gin.Context) {
	var reqs []models.CallbackRequest

	if err := c.ShouldBindJSON(&reqs); err != nil {
		fmt.Println("回调解析失败", err)
		c.JSON(http.StatusOK, models.XxlResponse{Code: 500, Msg: "invalid JSON"})
		return
	}

	for _, req := range reqs {
		status := "success"
		if req.HandleCode != 200 {
			status = "失败"
		}
		log.Printf("\n[收到汇报] 任务ID: %d |结果: %s |消息: %s\n", req.LogId, status, req.HandleMsg)

		handlerTime := time.UnixMilli(req.LogDateTime)
		result := global.DB.Model(&models.JobLog{}).
			Where("id = ?", req.LogId).
			Updates(map[string]interface{}{
				"handler_code": fmt.Sprintf("%d", req.HandleCode),
				"handler_msg":  req.HandleMsg,
				"handler_time": handlerTime,
			})
		if result.Error != nil {
			log.Printf("[持久化失败],logId=%d, err=%v", req.LogId, result.Error)
		}
	}
	c.JSON(http.StatusOK, models.XxlResponse{Code: 200})
}
