package scheduler

import (
	"context"

	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
)

// SyncInterface 同步接口
type SyncInterface interface {
	SyncNucleiTemplates()
	SyncWappalyzerFingerprints()
	ImportCustomPocAndFingerprints()
	RefreshTemplateCache()
}

// SchedulerService 调度器服务，实现 service.Service 接口
type SchedulerService struct {
	scheduler   *Scheduler
	cronManager *CronManager
	rdb         *redis.Client
	syncMethods SyncInterface
}

// NewSchedulerService 创建调度器服务
func NewSchedulerService(rdb *redis.Client, syncMethods SyncInterface) *SchedulerService {
	sched := NewScheduler(rdb)
	cronManager := NewCronManager(sched, rdb)

	return &SchedulerService{
		scheduler:   sched,
		cronManager: cronManager,
		rdb:         rdb,
		syncMethods: syncMethods,
	}
}

// Start 启动服务
func (s *SchedulerService) Start() {
	logx.Info("Starting scheduler service...")

	// 启动调度器
	s.scheduler.Start()

	// 加载定时任务
	s.cronManager.LoadTasks(context.Background())

	// 启动后台同步任务
	if s.syncMethods != nil {
		// 先加载缓存
		s.syncMethods.RefreshTemplateCache()

		// 异步同步模板和指纹
		go s.syncMethods.SyncNucleiTemplates()
		go s.syncMethods.SyncWappalyzerFingerprints()
		go s.syncMethods.ImportCustomPocAndFingerprints()
	}

	logx.Info("Scheduler service started")
}

// Stop 停止服务
func (s *SchedulerService) Stop() {
	logx.Info("Stopping scheduler service...")
	s.scheduler.Stop()
	logx.Info("Scheduler service stopped")
}

// GetScheduler 获取调度器
func (s *SchedulerService) GetScheduler() *Scheduler {
	return s.scheduler
}

// GetCronManager 获取定时任务管理器
func (s *SchedulerService) GetCronManager() *CronManager {
	return s.cronManager
}
