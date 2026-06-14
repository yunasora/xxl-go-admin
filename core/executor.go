package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go-xxl-admin/config"
	"go-xxl-admin/global"
	"go-xxl-admin/models"
	"go-xxl-admin/mq"
	"log"
	"net/http"
	"time"
)

func SendTrigger(appName string, job models.JobInfo) {

	log.Printf("[Admin]正在为集群: %s 下发任务", appName)

	if config.Cfg.MQEnabled {
		logRecord := models.JobLog{
			JobId:           job.ID,
			ExecutorAddress: appName,
			ExecutorHandler: job.ExecutorHandler,
			ExecutorParam:   job.ExecutorParam,
			TriggerTime:     time.Now(),
			TriggerCode:     "0",
			TriggerMsg:      "已入队",
		}

		global.DB.Create(&logRecord)

		param := models.TriggerParam{
			JobId:           job.ID,
			ExecutorHandler: job.ExecutorHandler,
			ExecutorParams:  job.ExecutorParam,
			LogId:           logRecord.ID,
			LogDateTime:     time.Now().UnixMilli(),
			GlueType:        "BEAN",
			AppName:         appName,
		}
		if err := mq.PublishTask(config.Cfg.MQQueue, param); err != nil {

			log.Printf("[MQ] 发布失败: %v", err)
			global.DB.Model(&logRecord).Updates(map[string]interface{}{
				"trigger_code": "500",
				"trigger_msg":  "MQ发布失败",
			})

		}
		return
	}

	maxRetry := int(job.ExecutorFailRetryCount) + 1
	timeout := job.ExecutorTimeout

	if timeout == 0 {
		timeout = 5
	}

	var logRecord models.JobLog

	for attempt := 1; attempt <= maxRetry; attempt++ {
		//选举节点
		targetAddr, err := RegsC.ElectNode(appName)
		if err != nil {
			log.Printf("[重试%d / %d]	选举失败: %v", attempt, maxRetry, err)
			continue
		}
		if attempt == 1 {
			//首次尝试创建job_log
			logRecord = models.JobLog{
				JobId:           job.ID,
				ExecutorAddress: targetAddr,
				ExecutorHandler: job.ExecutorHandler,
				ExecutorParam:   job.ExecutorParam,
				TriggerTime:     time.Now(),
				TriggerCode:     "0",
				TriggerMsg:      "触发中",
			}
			global.DB.Create(&logRecord)
		}

		//拼出请求
		param := models.TriggerParam{
			JobId:           job.ID,
			ExecutorHandler: job.ExecutorHandler,
			ExecutorParams:  job.ExecutorParam,
			LogId:           logRecord.ID,
			LogDateTime:     time.Now().UnixMilli(),
			GlueType:        "BEAN",
		}

		jsonData, _ := json.Marshal(param)
		runUrl := targetAddr + "run"
		req, _ := http.NewRequest("POST", runUrl, bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("XXL-JOB-ACCESS-TOKEN", config.Cfg.AccessToken)

		//发请求
		client := &http.Client{Timeout: time.Duration(timeout) * time.Second}
		resp, err := client.Do(req)

		//判断结果
		if err == nil && resp.StatusCode == 200 {
			resp.Body.Close()
			global.DB.Model(&logRecord).Updates(map[string]interface{}{
				"trigger_code": "200",
				"trigger_msg":  "下发成功",
			})
			log.Printf("[下发成功] 第%d次尝试,目标: %s", attempt, runUrl)
			return
		}

		if resp != nil {
			resp.Body.Close()
		}

		if attempt == maxRetry {
			global.DB.Model(&logRecord).Updates(map[string]interface{}{
				"trigger_code": "500",
				"trigger_msg":  "最终失败",
			})
		} else {
			log.Printf("[重试 %d / %d]失败,准备重试....", attempt, maxRetry)
		}

	}
}

// 强杀执行器
func KillJob(targetAddr string, jobId int64) (*models.XxlResponse, error) {

	var httpClient = &http.Client{Timeout: time.Duration(config.Cfg.HTTPTimeout) * time.Second}
	url := targetAddr + "kill"
	reqBody := models.KillRequest{JobId: jobId}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("JSON Marshal faild: %w", err)
	}

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("XXL-JOB-ACCESS-TOKEN", config.Cfg.AccessToken)

	resp, err := httpClient.Do(req)

	if err != nil {
		return nil, fmt.Errorf("network failed on kill %w", err)
	}

	defer resp.Body.Close()

	var Xxlresp models.XxlResponse

	if err := json.NewDecoder(resp.Body).Decode(&Xxlresp); err != nil {
		return nil, fmt.Errorf("decode kill response fail :%w", err)
	}
	return &Xxlresp, err
}

// 跨终端拉取Executor日志
func FetchLog(targetAddr string, logDataTim int64, logId int64, fromLineNum int, logReq models.LogRequest) (*models.LogResultContent, error) {

	var httpClient = &http.Client{Timeout: time.Duration(config.Cfg.HTTPTimeout) * time.Second}
	url := targetAddr + "log"

	reqBody := models.LogRequest{
		LogDataTim:  logDataTim,
		LogId:       logId,
		FromLineNum: fromLineNum,
	}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json;charset=UTF-8")
	req.Header.Set("XXL-JOB-ACCESS-TOKEN", config.Cfg.AccessToken)

	var Xxlresp models.XxlResponse
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&Xxlresp); err != nil {
		return nil, fmt.Errorf("decode log response faild: %w", err)
	}
	if Xxlresp.Code != 200 {
		return nil, fmt.Errorf("Java Executor Error: %s", Xxlresp.Msg)
	}

	contentBytes, err := json.Marshal(Xxlresp.Content)

	var logContent models.LogResultContent
	if err := json.Unmarshal(contentBytes, &logContent); err != nil {
		return nil, fmt.Errorf("parse log content failed: %w", err)
	}
	return &logContent, nil
}
