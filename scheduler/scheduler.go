package scheduler

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/robfig/cron/v3"
)

// 任务状态常量
const (
	TaskStatusCreated = "CREATED"
	TaskStatusPending = "PENDING"
	TaskStatusStarted = "STARTED"
	TaskStatusSuccess = "SUCCESS"
	TaskStatusFailure = "FAILURE"
	TaskStatusRevoked = "REVOKED"
)

// TaskInfo 任务信息
type TaskInfo struct {
	TaskId      string `json:"taskId"`
	MainTaskId  string `json:"mainTaskId"`
	WorkspaceId string `json:"workspaceId"`
	TaskName    string `json:"taskName"`
	Config      string `json:"config"`
	Priority    int    `json:"priority"`
	CreateTime  string `json:"createTime"`
}

// Scheduler 任务调度器
type Scheduler struct {
	rdb         *redis.Client
	cron        *cron.Cron
	queueKey    string
	processingKey string
	mu          sync.Mutex
	handlers    map[string]TaskHandler
}

// TaskHandler 任务处理函数
type TaskHandler func(ctx context.Context, task *TaskInfo) error

// NewScheduler 创建调度器
func NewScheduler(rdb *redis.Client) *Scheduler {
	return &Scheduler{
		rdb:           rdb,
		cron:          cron.New(cron.WithSeconds()),
		queueKey:      "cscan:task:queue",
		processingKey: "cscan:task:processing",
		handlers:      make(map[string]TaskHandler),
	}
}

// RegisterHandler 注册任务处理器
func (s *Scheduler) RegisterHandler(taskName string, handler TaskHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.handlers[taskName] = handler
}

// Start 启动调度器
func (s *Scheduler) Start() {
	s.cron.Start()
}

// Stop 停止调度器
func (s *Scheduler) Stop() {
	s.cron.Stop()
}

// AddCronTask 添加定时任务
func (s *Scheduler) AddCronTask(spec string, taskFunc func()) (cron.EntryID, error) {
	return s.cron.AddFunc(spec, taskFunc)
}

// RemoveCronTask 移除定时任务
func (s *Scheduler) RemoveCronTask(id cron.EntryID) {
	s.cron.Remove(id)
}

// PushTask 推送任务到队列
func (s *Scheduler) PushTask(ctx context.Context, task *TaskInfo) error {
	if task.TaskId == "" {
		task.TaskId = uuid.New().String()
	}
	task.CreateTime = time.Now().Format("2006-01-02 15:04:05")

	data, err := json.Marshal(task)
	if err != nil {
		return err
	}

	// 使用优先级队列，分数越小优先级越高
	score := float64(time.Now().Unix()) - float64(task.Priority*1000)
	return s.rdb.ZAdd(ctx, s.queueKey, redis.Z{
		Score:  score,
		Member: data,
	}).Err()
}

// PopTask 从队列获取任务
func (s *Scheduler) PopTask(ctx context.Context) (*TaskInfo, error) {
	// 获取优先级最高的任务
	results, err := s.rdb.ZPopMin(ctx, s.queueKey, 1).Result()
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, nil
	}

	var task TaskInfo
	if err := json.Unmarshal([]byte(results[0].Member.(string)), &task); err != nil {
		return nil, err
	}

	// 添加到处理中集合
	s.rdb.SAdd(ctx, s.processingKey, task.TaskId)

	return &task, nil
}

// CompleteTask 完成任务
func (s *Scheduler) CompleteTask(ctx context.Context, taskId string) error {
	return s.rdb.SRem(ctx, s.processingKey, taskId).Err()
}

// GetQueueLength 获取队列长度
func (s *Scheduler) GetQueueLength(ctx context.Context) (int64, error) {
	return s.rdb.ZCard(ctx, s.queueKey).Result()
}

// GetProcessingCount 获取处理中任务数
func (s *Scheduler) GetProcessingCount(ctx context.Context) (int64, error) {
	return s.rdb.SCard(ctx, s.processingKey).Result()
}

// TaskConfig 任务配置
type TaskConfig struct {
	PortScan    *PortScanConfig    `json:"portscan,omitempty"`
	DomainScan  *DomainScanConfig  `json:"domainscan,omitempty"`
	Fingerprint *FingerprintConfig `json:"fingerprint,omitempty"`
	PocScan     *PocScanConfig     `json:"pocscan,omitempty"`
}

type PortScanConfig struct {
	Enable        bool   `json:"enable"`
	Tool          string `json:"tool"` // tcp, masscan, nmap
	Ports         string `json:"ports"`
	Rate          int    `json:"rate"`
	Timeout       int    `json:"timeout"`
	PortThreshold int    `json:"portThreshold"` // 开放端口数量阈值，超过则过滤该主机
}

type DomainScanConfig struct {
	Enable     bool `json:"enable"`
	Subfinder  bool `json:"subfinder"`
	Massdns    bool `json:"massdns"`
	Concurrent int  `json:"concurrent"`
}

type FingerprintConfig struct {
	Enable       bool `json:"enable"`
	Httpx        bool `json:"httpx"`
	IconHash     bool `json:"iconHash"`
	Wappalyzer   bool `json:"wappalyzer"`
	CustomEngine bool `json:"customEngine"` // 使用自定义指纹引擎（ARL格式）
	Screenshot   bool `json:"screenshot"`
}

type PocScanConfig struct {
	Enable            bool                `json:"enable"`
	PocTypes          []string            `json:"pocTypes"`          // nuclei, builtin
	PocFiles          []string            `json:"pocFiles"`          // 自定义POC文件
	UseNuclei         bool                `json:"useNuclei"`         // 使用Nuclei扫描
	AutoScan          bool                `json:"autoScan"`          // 基于自定义标签映射自动扫描
	AutomaticScan     bool                `json:"automaticScan"`     // 基于Wappalyzer内置映射自动扫描（类似nuclei -as）
	Severity          string              `json:"severity"`          // 严重级别过滤
	Tags              []string            `json:"tags"`              // 手动指定标签
	ExcludeTags       []string            `json:"excludeTags"`       // 排除标签
	RateLimit         int                 `json:"rateLimit"`         // 速率限制
	Concurrency       int                 `json:"concurrency"`       // 并发数
	CustomPocOnly     bool                `json:"customPocOnly"`     // 只使用自定义POC
	CustomTemplates   []string            `json:"customTemplates"`   // 自定义POC模板内容(YAML) - 已废弃
	NucleiTemplates   []string            `json:"nucleiTemplates"`   // 从数据库获取的模板内容(YAML) - 已废弃
	NucleiTemplateIds []string            `json:"nucleiTemplateIds"` // Nuclei模板ID列表（新）
	CustomPocIds      []string            `json:"customPocIds"`      // 自定义POC ID列表（新）
	TagMappings       map[string][]string `json:"tagMappings"`       // 应用名称到Nuclei标签的映射
}

// ParseTaskConfig 解析任务配置
func ParseTaskConfig(configStr string) (*TaskConfig, error) {
	var config TaskConfig
	if err := json.Unmarshal([]byte(configStr), &config); err != nil {
		return nil, err
	}
	return &config, nil
}

// BuildTaskConfig 构建任务配置
func BuildTaskConfig(config *TaskConfig) (string, error) {
	data, err := json.Marshal(config)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// TaskResult 任务结果
type TaskResult struct {
	TaskId     string `json:"taskId"`
	Status     string `json:"status"`
	Message    string `json:"message"`
	AssetCount int    `json:"assetCount"`
	VulCount   int    `json:"vulCount"`
	Duration   int64  `json:"duration"`
}

// FormatResult 格式化结果
func (r *TaskResult) FormatResult() string {
	return fmt.Sprintf("状态:%s 资产:%d 漏洞:%d 耗时:%ds",
		r.Status, r.AssetCount, r.VulCount, r.Duration)
}
