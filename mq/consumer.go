package mq

import (
	"bytes"
	"encoding/json"
	"go-xxl-admin/config"
	"go-xxl-admin/global"
	"go-xxl-admin/models"
	"log"
	"net/http"
	"time"
)

func StartConsumer(queueName string, electFunc func(appName string) (string, error)) {

	msgs, err := Channel.Consume(
		queueName, // 队列名
		"",        // consumer 名称，空字符串 = 自动生成
		false,     // autoAck: false = 手动确认（重要！）
		false,     // exclusive: 是否独占
		false,     // noLocal: 不支持了，填 false
		false,     // noWait: 不等服务器确认
		nil,
	)
	if err != nil {
		log.Printf("[MQ Consumer] 启动消费失败: %v", err)
		return
	}

	log.Printf("[MQ Consumer] 开始监听队列: %s", queueName)

	for msg := range msgs {

		//反序列化
		var task models.TriggerParam
		if err := json.Unmarshal(msg.Body, &task); err != nil {
			log.Printf("[MQ Consumer] JSON 解析失败: %v", err)
			msg.Nack(false, false) // 格式错误，不重新入队
			continue
		}

		//选举节点
		addr, err := electFunc(task.AppName)
		if err != nil {
			log.Printf("[MQ Consumer] 选举节点失败: %v", err)
			msg.Nack(false, true) // 重新入队
			continue
		}

		jsonData, _ := json.Marshal(task)
		runUrl := addr + "run"
		req, _ := http.NewRequest("POST", runUrl, bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("XXL-JOB-ACCESS-TOKEN", config.Cfg.AccessToken)

		//发请求
		client := &http.Client{Timeout: time.Duration(config.Cfg.HTTPTimeout) * time.Second}
		resp, err := client.Do(req)

		//判断结果
		if err == nil && resp.StatusCode == 200 {
			resp.Body.Close()
			global.DB.Model(&models.JobLog{}).Where("id= ?", task.LogId).Updates(map[string]interface{}{
				"trigger_code": "200",
				"trigger_msg":  "MQ下发成功",
			})
			log.Printf("[MQ下发成功] 目标: %s", runUrl)
			msg.Ack(false)
		} else {
			if resp != nil {
				resp.Body.Close()
			}

			msg.Nack(false, true)
			log.Printf("[MQ Consumer] 下发失败,重新入队")
		}

	}

}
