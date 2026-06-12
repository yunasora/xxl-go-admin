package core

import (
	"go-xxl-admin/config"
	"go-xxl-admin/global"
	"go-xxl-admin/models"
	"log"
	"time"

	"github.com/robfig/cron/v3"
)

type Scheduler struct {
}

var Sched = &Scheduler{}

func (s *Scheduler) Start() {
	log.Println("[调度器] 已启动，开始每秒扫描...")
	go func() {
		ticker := time.NewTicker(time.Duration(config.Cfg.SchedulerInterval) * time.Second)
		for range ticker.C {
			s.scanAndTrigger()
		}
	}()
}

func (s *Scheduler) scanAndTrigger() {
	var jobs []models.JobInfo

	nowMillis := time.Now().UnixMilli()

	global.DB.Where("trigger_status = 1 AND triggeer_next_time <= ?", nowMillis).Find(&jobs)
	log.Printf("[调度扫描] 当前时间=%d, 扫到任务数=%d", nowMillis, len(jobs))

	for _, job := range jobs {

		var group models.JobGroup

		if err := global.DB.First(&group, job.JobGroup).Error; err != nil {
			log.Printf("[调度错误] 查询job_group失败 groupId=%d: %v", job.JobGroup, err)
			continue
		}

		SendTrigger(group.AppName, job)
		log.Printf("[调度扫描] 触发任务ID=%d, Handler=%s, 参数=%s",
			job.ID, job.ExecutorHandler, job.ExecutorParam)
		parser := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
		schedule, err := parser.Parse(job.JobCron)
		if err != nil {
			log.Printf("解析cron错误:%v", err)
			continue
		}
		nextTime := schedule.Next(time.Now())

		//更新数据库
		global.DB.Model(&job).Updates(map[string]interface{}{
			"triggeer_last_time": time.Now().UnixMilli(),
			"triggeer_next_time": nextTime.UnixMilli(),
		})
	}
}
