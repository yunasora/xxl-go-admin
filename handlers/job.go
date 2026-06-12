package handlers

import (
	"go-xxl-admin/global"
	"go-xxl-admin/models"
	"strconv"

	"github.com/gin-gonic/gin"
)

func CreateJob(c *gin.Context) {

	var job models.JobInfo
	if err := c.ShouldBindJSON(&job); err != nil {
		c.JSON(200, models.XxlResponse{Code: 500, Msg: "参数错误"})
		return
	}
	if job.TriggerStatus == 0 {
		job.TriggerStatus = 1
	}
	global.DB.Create(&job)
	c.JSON(200, models.XxlResponse{Code: 200, Content: job})
}

func GetJob(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var job models.JobInfo
	if err := global.DB.First(&job, id).Error; err != nil {
		c.JSON(200, models.XxlResponse{Code: 500, Msg: "任务不存在"})
		return
	}
	c.JSON(200, models.XxlResponse{Code: 200, Content: job})
}

func ListJobs(c *gin.Context) {
	var jobs []models.JobInfo
	global.DB.Find(&jobs)
	c.JSON(200, models.XxlResponse{Code: 200, Content: jobs})
}

func UpdateJob(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var job models.JobInfo
	if err := global.DB.First(&job, id).Error; err != nil {
		c.JSON(200, models.XxlResponse{Code: 500, Msg: "参数错误"})
		return
	}
	var updates models.JobInfo
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(200, models.XxlResponse{Code: 500, Msg: "参数错误"})
		return
	}
	global.DB.Model(&job).Updates(updates)
	c.JSON(200, models.XxlResponse{Code: 200, Content: job})
}

func DeleteJob(c *gin.Context) {

	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var job models.JobInfo
	if err := global.DB.Delete(&job, id).Error; err != nil {
		c.JSON(200, models.XxlResponse{Code: 500, Msg: "任务不存在"})
		return
	}

	c.JSON(200, models.XxlResponse{Code: 200})
}

func StartJob(c *gin.Context) {

	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	global.DB.Model(&models.JobInfo{}).Where("id= ?", id).Update("trigger_status", 1)
	c.JSON(200, models.XxlResponse{Code: 200})

}

func StopJob(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	global.DB.Model(&models.JobInfo{}).Where("id= ?", id).Update("trigger_status", 0)
	c.JSON(200, models.XxlResponse{Code: 200})
}
