package scheduler

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/robfig/cron/v3"
)

// CronTask 定时任务
type CronTask struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	CronSpec    string `json:"cronSpec"`
	WorkspaceId string `json:"workspaceId"`
	MainTaskId  string `json:"mainTaskId"`
	Config      string `json:"config"`
	Status      string `json:"status"` // enable/disable
	LastRunTime string `json:"lastRunTime"`
	NextRunTime string `json:"nextRunTime"`
	EntryId     cron.EntryID `json:"-"`
}

// CronManager 定时任务管理器
type CronManager struct {
	scheduler *Scheduler
	rdb       *redis.Client
	tasks     map[string]*CronTask
	cronKey   string
}

// NewCronManager 创建定时任务管理器
func NewCronManager(scheduler *Scheduler, rdb *redis.Client) *CronManager {
	return &CronManager{
		scheduler: scheduler,
		rdb:       rdb,
		tasks:     make(map[string]*CronTask),
		cronKey:   "cscan:cron:tasks",
	}
}

// LoadTasks 从Redis加载定时任务
func (m *CronManager) LoadTasks(ctx context.Context) error {
	data, err := m.rdb.HGetAll(ctx, m.cronKey).Result()
	if err != nil {
		return err
	}

	for id, taskData := range data {
		var task CronTask
		if err := json.Unmarshal([]byte(taskData), &task); err != nil {
			continue
		}
		task.Id = id
		if task.Status == "enable" {
			m.startTask(&task)
		}
		m.tasks[id] = &task
	}

	return nil
}

// AddTask 添加定时任务
func (m *CronManager) AddTask(ctx context.Context, task *CronTask) error {
	// 验证cron表达式
	parser := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	schedule, err := parser.Parse(task.CronSpec)
	if err != nil {
		return fmt.Errorf("invalid cron spec: %v", err)
	}

	task.NextRunTime = schedule.Next(time.Now()).Format("2006-01-02 15:04:05")
	task.Status = "enable"

	// 保存到Redis
	data, _ := json.Marshal(task)
	if err := m.rdb.HSet(ctx, m.cronKey, task.Id, data).Err(); err != nil {
		return err
	}

	// 启动任务
	m.startTask(task)
	m.tasks[task.Id] = task

	return nil
}

// RemoveTask 移除定时任务
func (m *CronManager) RemoveTask(ctx context.Context, taskId string) error {
	task, ok := m.tasks[taskId]
	if !ok {
		return fmt.Errorf("task not found: %s", taskId)
	}

	// 停止任务
	if task.EntryId > 0 {
		m.scheduler.RemoveCronTask(task.EntryId)
	}

	// 从Redis删除
	if err := m.rdb.HDel(ctx, m.cronKey, taskId).Err(); err != nil {
		return err
	}

	delete(m.tasks, taskId)
	return nil
}

// EnableTask 启用定时任务
func (m *CronManager) EnableTask(ctx context.Context, taskId string) error {
	task, ok := m.tasks[taskId]
	if !ok {
		return fmt.Errorf("task not found: %s", taskId)
	}

	if task.Status == "enable" {
		return nil
	}

	task.Status = "enable"
	m.startTask(task)

	// 更新Redis
	data, _ := json.Marshal(task)
	return m.rdb.HSet(ctx, m.cronKey, taskId, data).Err()
}

// DisableTask 禁用定时任务
func (m *CronManager) DisableTask(ctx context.Context, taskId string) error {
	task, ok := m.tasks[taskId]
	if !ok {
		return fmt.Errorf("task not found: %s", taskId)
	}

	if task.Status == "disable" {
		return nil
	}

	// 停止任务
	if task.EntryId > 0 {
		m.scheduler.RemoveCronTask(task.EntryId)
		task.EntryId = 0
	}

	task.Status = "disable"

	// 更新Redis
	data, _ := json.Marshal(task)
	return m.rdb.HSet(ctx, m.cronKey, taskId, data).Err()
}

// GetTasks 获取所有定时任务
func (m *CronManager) GetTasks() []*CronTask {
	tasks := make([]*CronTask, 0, len(m.tasks))
	for _, task := range m.tasks {
		tasks = append(tasks, task)
	}
	return tasks
}

// startTask 启动定时任务
func (m *CronManager) startTask(task *CronTask) {
	entryId, err := m.scheduler.AddCronTask(task.CronSpec, func() {
		m.executeTask(task)
	})
	if err != nil {
		return
	}
	task.EntryId = entryId
}

// executeTask 执行定时任务
func (m *CronManager) executeTask(task *CronTask) {
	ctx := context.Background()

	// 更新最后执行时间
	task.LastRunTime = time.Now().Format("2006-01-02 15:04:05")

	// 计算下次执行时间
	parser := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	schedule, _ := parser.Parse(task.CronSpec)
	task.NextRunTime = schedule.Next(time.Now()).Format("2006-01-02 15:04:05")

	// 更新Redis
	data, _ := json.Marshal(task)
	m.rdb.HSet(ctx, m.cronKey, task.Id, data)

	// 推送任务到队列
	taskInfo := &TaskInfo{
		MainTaskId:  task.MainTaskId,
		WorkspaceId: task.WorkspaceId,
		TaskName:    task.Name,
		Config:      task.Config,
		Priority:    0,
	}
	m.scheduler.PushTask(ctx, taskInfo)
}
