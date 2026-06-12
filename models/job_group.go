package models

type JobGroup struct {
	ID      int64  `gorm:"primaryKey;autoIncrement"`
	AppName string `gorm:"type:varchar(255);not null;uniqueIndex"`
	Title   string `gorm:"type:varchar(255)"`
}

func (JobGroup) TableName() string { return "job_group" }
