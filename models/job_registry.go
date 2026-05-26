package models

import "time"

type JobRegistry struct {
	ID            int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	RegistryGroup string    `gorm:"type:varchar(50);not null;index:idx_group_key" json:"registryGroup"`
	RegistryKey   string    `gorm:"type:varchar(255);not null;index:idx_group_key" json:"registryKey"`
	RegistryValue string    `gorm:"type:varchar(255);not null;uniqueIndex:idx_val" json:"registryValue"`
	UpdateTime    time.Time `gorm:"autoUpdateTime" json:"updateTime"`
}

func (JobRegistry) TableName() string {
	return "job_registry"
}
