package models

import "time"

type JobInfo struct {
	ID                      int64     `gorm:"primaryKey;autoIncrement" json:"Id"`
	JobGroup                int64     `gorm:"not null;index" json:"jobGroup"`
	JobDesc                 string    `gorm:"type:varchar(255);not null" json:"jobDesc"`
	ExecutorHandler         string    `gorm:"type:varchar(255);not null" json:"executorHandler"`
	JobCron                 string    `gorm:"type:varchar(128);not null" json:"jobCron"`
	ExecutorRoutingStrategy string    `gorm:"type:varchar(50);default:'FIRST'" json:"executorRoutingStrategy"`
	ExecutorParam           string    `gorm:"type:text" json:"executorParam"`
	ExecutorTimeout         int       `gorm:"not null;default:0" json:"executorTimeout"`
	ExecutorFailRetryCount  int       `gorm:"not null;default:0" json:"executorFailRetryCount"`
	TriggerStatus           int       `gorm:"not null;default:0" json:"triggerStatus"`
	TriggeerLastTime        int64     `gorm:"not null;default:0" json:"triggerLastTime"`
	TriggeerNextTime        int64     `gorm:"not null;default:0" json:"triggerNextTime"`
	CreateTime              time.Time `gorm:"autoCreateTime" json:"createTime"`
	UpdateTime              time.Time `gorm:"autoUpdateTime" json:"updateTime"`
}

func (JobInfo) TableName() string {
	return "job_info"
}
