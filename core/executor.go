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
	jsonData, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("XXL-JOB-ACCESS-TOKEN", "default_token")

	resp, err := httpClient.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var Xxlresp models.XxlResponse

	if err := json.NewDecoder(resp.Body).Decode(&Xxlresp); err != nil {
		return nil, err
	}
	return &Xxlresp, err
}

// 跨终端拉取日志
func FetchLog(targetAddr string, logReq models.LogRequest) (*models.LogResultContent, error) {

	url := targetAddr + "log"
	jsonData, _ := json.Marshal(logReq)

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
		return nil, err
	}
	if Xxlresp.Code != 200 {
		return nil, fmt.Errorf("executor exception: %s", Xxlresp.Msg)
	}

	contentBytes, _ := json.Marshal(Xxlresp.Content)

	var logContent models.LogResultContent
	_ = json.Unmarshal(contentBytes, &logContent)

	return &logContent, nil
}
