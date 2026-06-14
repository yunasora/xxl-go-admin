package models

type JobGroup struct {
	ID      int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	AppName string `gorm:"type:varchar(255);not null;uniqueIndex;not null" json:appName`
	Title   string `gorm:"type:varchar(255);not null" json:"title"`
}

func (JobGroup) TableName() string { return "job_group" }
