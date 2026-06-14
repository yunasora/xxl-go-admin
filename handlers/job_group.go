package handlers

import (
	"go-xxl-admin/global"
	"go-xxl-admin/models"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func CreateJobGrop(c *gin.Context) {

	var req models.JobGroup

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, models.XxlResponse{
			Code: 500,
			Msg:  "参数错误",
		})
		return
	}

	req.AppName = strings.TrimSpace(req.AppName)
	req.Title = strings.TrimSpace(req.Title)

	if req.AppName == "" || req.Title == "" {
		c.JSON(http.StatusOK, models.XxlResponse{
			Code: 500,
			Msg:  "AppName或title不得为空",
		})
		return
	}

	var count int64
	global.DB.Model(&models.JobGroup{}).Where("app_name = ?", req.AppName).Count(&count)

	if count > 0 {
		c.JSON(http.StatusOK, models.XxlResponse{
			Code: 500,
			Msg:  "appName已存在",
		})
		return
	}

	c.JSON(http.StatusOK, models.XxlResponse{
		Code:    200,
		Content: req,
	})

}
