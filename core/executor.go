package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go-xxl-admin/models"
	"log"
	"net/http"
	"time"
)

var httpClient = &http.Client{Timeout: 5 * time.Second}

func SendTrigger(appName string) {
	log.Printf("[Admin]正在为集群: %s 下发任务", appName)

	targetAddr, err := RegsC.ElectNode(appName)
	if err != nil {
		log.Printf("[调度错误],任务中心下发任务终止,%v", err)
		return
	}
	jsonBody := []byte(`{
		"jobId": 1,
		"executorHandler": "demoJobHandler",
		"executorParams": "Go callback test",
		"logId": 20240420,
		"logDateTime": 1690000000000,
		"glueType": "BEAN"
	}`)

	//拼装URL
	runUrl := targetAddr + "run"

	req, err := http.NewRequest("POST", runUrl, bytes.NewBuffer(jsonBody))

	if err != nil {
		fmt.Println("创建请求对象失败")
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("XXL-JOB-ACCESS-TOKEN", "default_token")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("向节点 %s 下发任务失败, %v", runUrl, err)
		return
	}
	defer resp.Body.Close()

	log.Printf("[下发成功] 目标节点: %s || [http码状态]: %s", runUrl, resp.StatusCode)
}

// 强杀执行器
func KillJob(targetAddr string, jobId int64) (*models.XxlResponse, error) {

	url := targetAddr + "kill"
	reqBody := models.KillRequest{JobId: jobId}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("JSON Marshal faild: %w", err)
	}

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("XXL-JOB-ACCESS-TOKEN", "default_token")

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
	req.Header.Set("XXL-JOB-ACCESS-TOKEN", "default_token")

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
