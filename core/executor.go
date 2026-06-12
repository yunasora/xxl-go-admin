package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go-xxl-admin/config"
	"go-xxl-admin/global"
	"go-xxl-admin/models"
	"log"
	"net/http"
	"time"
)

func SendTrigger(appName string, job models.JobInfo) {

	log.Printf("[Admin]正在为集群: %s 下发任务", appName)

	targetAddr, err := RegsC.ElectNode(appName)
	if err != nil {
		log.Printf("[调度错误],任务中心下发任务终止,%v", err)
		return
	}

	logRecord := models.JobLog{
		JobId:           job.ID,
		ExecutorAddress: targetAddr,
		ExecutorHandler: job.ExecutorHandler,
		ExecutorParam:   job.ExecutorParam,
		TriggerTime:     time.Now(),
		TriggerCode:     "0",
		TriggerMsg:      "触发中",
	}

	global.DB.Create(&logRecord)

	param := models.TriggerParam{
		JobId:           job.ID,
		ExecutorHandler: job.ExecutorHandler,
		ExecutorParams:  job.ExecutorParam,
		LogId:           logRecord.ID,
		LogDateTime:     time.Now().UnixMilli(),
		GlueType:        "BEAN",
	}

	jsonData, err := json.Marshal(param)

	//拼装URL
	runUrl := targetAddr + "run"

	req, err := http.NewRequest("POST", runUrl, bytes.NewBuffer(jsonData))

	if err != nil {
		fmt.Println("创建请求对象失败")
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("XXL-JOB-ACCESS-TOKEN", config.Cfg.AccessToken)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("向节点 %s 下发任务失败, %v", runUrl, err)
		global.DB.Model(&logRecord).Updates(map[string]interface{}{
			"trigger_code": "500",
			"trigger_msg":  err.Error(),
		})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		global.DB.Model(&logRecord).Updates(map[string]interface{}{
			"trigger_code": "200",
			"trigger_msg":  "下发成功",
		})
		log.Printf("[下发成功] 目标节点: %s", runUrl)
	} else {
		global.DB.Model(&logRecord).Updates(map[string]interface{}{
			"trigger_code": "500",
			"trigger_msg":  fmt.Sprintf("Executor返回异常: %d", resp.StatusCode),
		})
		log.Printf("[下发失败] 目标节点: %s, 状态码: %d", runUrl, resp.StatusCode)
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
