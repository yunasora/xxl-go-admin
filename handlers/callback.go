package handlers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go-xxl-admin/models"
	"log"
	"net/http"
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
	}
	c.JSON(http.StatusOK, models.XxlResponse{Code: 200})
}
