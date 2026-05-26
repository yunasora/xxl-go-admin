package models

import "time"

type JobLog struct {
	ID              int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	JobId           int64     `gorm:"not null;index" json:"jobId"`
	ExecutorAddress string    `gorm:"type:varchar(255);not null;index" json:"executorAddress"`
	ExecutorHandler string    `gorm:"type:varchar(255);not null;index" json:"executorHandler"`
	ExecutorParam   string    `gorm:"type:text" json:"executorParam"`
	TriggerTime     time.Time `gorm:"not null" json:"triggerTime"`
	TriggerCode     string    `gorm:"not null;default:0" json:"triggerCode"`
	TriggerMsg      string    `gorm:"type:text" json:"triggerMsg"`
	HandlerTime     time.Time `json:"handlerTime"`
	HandlerCode     string    `gorm:"not null;default:0" json:"handlerCode"`
	HandlerMsg      string    `gorm:"type:text" json:"handlerMsg"`
}

func (JobLog) TableName() string {
	return "job_log"
}
