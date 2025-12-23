package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"cscan/model"
	"cscan/pkg/mapping"
	"cscan/rpc/task/pb"
	"cscan/scanner"
	"cscan/scheduler"

	"github.com/redis/go-redis/v9"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/zrpc"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/grpc"
)

// 使用标准 protobuf 编解码器

// WorkerConfig Worker配置
type WorkerConfig struct {
	Name        string `json:"name"`
	ServerAddr  string `json:"serverAddr"`
	RedisAddr   string `json:"redisAddr"`
	RedisPass   string `json:"redisPass"`
	Concurrency int    `json:"concurrency"`
	Timeout     int    `json:"timeout"`
}

// Worker 工作节点
type Worker struct {
	ctx         context.Context
	cancel      context.CancelFunc
	config      WorkerConfig
	rpcClient   pb.TaskServiceClient
	redisClient *redis.Client
	scanners    map[string]scanner.Scanner
	taskChan    chan *scheduler.TaskInfo
	resultChan  chan *scanner.ScanResult
	stopChan    chan struct{}
	wg          sync.WaitGroup
	mu          sync.Mutex

	taskStarted  int
	taskExecuted int
	isRunning    bool
	
	// 日志组件
	logger *WorkerLogger
}

// taskLog 发布任务级别日志
func (w *Worker) taskLog(taskId, level, format string, args ...interface{}) {
	logger := NewTaskLogger(w.redisClient, w.config.Name, taskId)
	switch level {
	case LevelError:
		logger.Error(format, args...)
	case LevelWarn:
		logger.Warn(format, args...)
	case LevelDebug:
		logger.Debug(format, args...)
	default:
		logger.Info(format, args...)
	}
}

// VulnerabilityBuffer 批量缓冲保存漏洞
type VulnerabilityBuffer struct {
	vuls      []*scanner.Vulnerability
	mu        sync.Mutex
	maxSize   int
	flushChan chan struct{}
}

// NewVulnerabilityBuffer 创建漏洞缓冲区
func NewVulnerabilityBuffer(maxSize int) *VulnerabilityBuffer {
	return &VulnerabilityBuffer{
		vuls:      make([]*scanner.Vulnerability, 0, maxSize),
		maxSize:   maxSize,
		flushChan: make(chan struct{}, 1),
	}
}

// Add 添加漏洞到缓冲区，返回是否需要刷�?
func (b *VulnerabilityBuffer) Add(vul *scanner.Vulnerability) {
	b.mu.Lock()
	b.vuls = append(b.vuls, vul)
	shouldFlush := len(b.vuls) >= b.maxSize
	b.mu.Unlock()

	if shouldFlush {
		select {
		case b.flushChan <- struct{}{}:
		default:
		}
	}
}

// Flush 刷新缓冲区，批量保存
func (b *VulnerabilityBuffer) Flush(ctx context.Context, saver func([]*scanner.Vulnerability)) {
	b.mu.Lock()
	vuls := b.vuls
	b.vuls = nil
	b.mu.Unlock()

	if len(vuls) > 0 {
		saver(vuls) // 批量保存
	}
}

// NewWorker 创建Worker
func NewWorker(config WorkerConfig) (*Worker, error) {
	// 创建RPC客户端，增加消息大小限制到100MB
	client, err := zrpc.NewClient(zrpc.RpcClientConf{
		Target: config.ServerAddr,
	}, zrpc.WithDialOption(grpc.WithDefaultCallOptions(
		grpc.MaxCallRecvMsgSize(100*1024*1024), // 100MB
		grpc.MaxCallSendMsgSize(100*1024*1024), // 100MB
	)))
	if err != nil {
		return nil, fmt.Errorf("connect to server failed: %v", err)
	}

	// 创建Redis客户端（用于日志推送）
	var redisClient *redis.Client
	if config.RedisAddr != "" {
		redisClient = redis.NewClient(&redis.Options{
			Addr:     config.RedisAddr,
			Password: config.RedisPass,
			DB:       0,
		})
		
		// 测试Redis连接
		ctx := context.Background()
		if err := redisClient.Ping(ctx).Err(); err != nil {
			fmt.Printf("[Worker] Redis connection failed: %v, logs will not be streamed\n", err)
			redisClient = nil
		} else {
			fmt.Printf("[Worker] Redis connected successfully at %s, logs will be streamed\n", config.RedisAddr)
			// 设置logx的输出Writer，将所有日志同时发送到Redis
			logWriter := NewRedisLogWriter(redisClient, config.Name)
			logx.SetWriter(logx.NewWriter(logWriter))
			// 写入一条测试日志确认日志系统工作
			NewLogPublisher(redisClient, config.Name).PublishWorkerLog(LevelInfo, "Worker日志系统已启动，Redis连接成功")
		}
	} else {
		fmt.Println("[Worker] Redis address not specified (-r flag), logs will not be streamed to Web")
	}

	// 创建可取消的Context
	ctx, cancel := context.WithCancel(context.Background())

	w := &Worker{
		ctx:         ctx,
		cancel:      cancel,
		config:      config,
		rpcClient:   pb.NewTaskServiceClient(client.Conn()),
		redisClient: redisClient,
		scanners:    make(map[string]scanner.Scanner),
		taskChan:    make(chan *scheduler.TaskInfo, config.Concurrency),
		resultChan:  make(chan *scanner.ScanResult, 100),
		stopChan:    make(chan struct{}),
		logger:      NewWorkerLogger(redisClient, config.Name),
	}

	// 注册扫描器
	w.registerScanners()

	// 加载HTTP服务映射配置
	w.loadHttpServiceMappings()

	return w, nil
}

// registerScanners 注册扫描器
func (w *Worker) registerScanners() {
	w.scanners["portscan"] = scanner.NewPortScanner()
	w.scanners["masscan"] = scanner.NewMasscanScanner()
	w.scanners["nmap"] = scanner.NewNmapScanner()
	w.scanners["naabu"] = scanner.NewNaabuScanner()
	w.scanners["domainscan"] = scanner.NewDomainScanner()
	w.scanners["fingerprint"] = scanner.NewFingerprintScanner()
	w.scanners["nuclei"] = scanner.NewNucleiScanner()
}

// Start 启动Worker
func (w *Worker) Start() {
	w.isRunning = true

	// 启动任务处理协程
	for i := 0; i < w.config.Concurrency; i++ {
		w.wg.Add(1)
		go w.processTask()
	}

	// 启动任务拉取协程
	w.wg.Add(1)
	go w.fetchTasks()

	// 启动结果上报协程
	w.wg.Add(1)
	go w.reportResult()

	// 启动心跳协程
	w.wg.Add(1)
	go w.keepAlive()

	// 启动状态查询订阅协程
	if w.redisClient != nil {
		w.wg.Add(1)
		go w.subscribeStatusQuery()
	}

	w.logger.Info("Worker %s started with %d workers", w.config.Name, w.config.Concurrency)
}

// fetchTasks 从服务端拉取任务
func (w *Worker) fetchTasks() {
	defer w.wg.Done()

	emptyCount := 0
	baseInterval := 1 * time.Second  // 基础间隔改为1秒
	maxInterval := 5 * time.Second   // 最大间隔改�?秒，确保任务能在5秒内被拉取

	for {
		select {
		case <-w.stopChan:
			return
		default:
			hasTask := w.pullTask()
			if hasTask {
				emptyCount = 0
				time.Sleep(100 * time.Millisecond) // 有任务时快速拉�?
			} else {
				emptyCount++
				// 没有任务时逐渐增加等待时间，最多5秒
				interval := baseInterval * time.Duration(emptyCount)
				if interval > maxInterval {
					interval = maxInterval
				}
				time.Sleep(interval)
			}
		}
	}
}

// pullTask 拉取单个任务，返回是否获取到任务
func (w *Worker) pullTask() bool {
	ctx := context.Background()

	// 检查是否有空闲槽位
	if len(w.taskChan) >= w.config.Concurrency {
		return false
	}

	// 通过 RPC 获取任务
	resp, err := w.rpcClient.CheckTask(ctx, &pb.CheckTaskReq{
		TaskId: w.config.Name, // �?worker name 作为标识请求任务
	})
	if err != nil {
		return false
	}

	if resp.IsExist && !resp.IsFinished {
		// 有待执行的任务
		task := &scheduler.TaskInfo{
			TaskId:      resp.TaskId,
			MainTaskId:  resp.TaskId,
			WorkspaceId: resp.WorkspaceId,
			TaskName:    "scan",
			Config:      resp.Config,
		}
		w.taskChan <- task
		return true
	}
	return false
}

// Stop 停止Worker
func (w *Worker) Stop() {
	w.isRunning = false
	w.cancel() // 通知所有 goroutine 停止
	close(w.stopChan)
	w.wg.Wait()
	w.logger.Info("Worker %s stopped", w.config.Name)
}

// SubmitTask 提交任务
func (w *Worker) SubmitTask(task *scheduler.TaskInfo) {
	w.taskChan <- task
}

// processTask 处理任务
func (w *Worker) processTask() {
	defer w.wg.Done()

	for {
		select {
		case <-w.stopChan:
			return
		case task := <-w.taskChan:
			w.executeTask(task)
		}
	}
}

// checkTaskControl 检查任务控制信号
// 返回: "PAUSE" - 暂停, "STOP" - 停止, "" - 继续执行
func (w *Worker) checkTaskControl(ctx context.Context, taskId string) string {
	if w.redisClient == nil {
		return ""
	}
	ctrlKey := "cscan:task:ctrl:" + taskId
	ctrl, err := w.redisClient.Get(ctx, ctrlKey).Result()
	if err != nil {
		return ""
	}
	return ctrl
}

// saveTaskProgress 保存任务进度（用于暂停后继续�?
func (w *Worker) saveTaskProgress(ctx context.Context, task *scheduler.TaskInfo, completedPhases map[string]bool, assets []*scanner.Asset) {
	// 构建状态
	phases := make([]string, 0)
	for phase, completed := range completedPhases {
		if completed {
			phases = append(phases, phase)
		}
	}
	
	assetsJson, _ := json.Marshal(assets)
	state := map[string]interface{}{
		"completedPhases": phases,
		"assets":          string(assetsJson),
	}
	stateJson, _ := json.Marshal(state)
	
	// 通过RPC保存到数据库
	w.rpcClient.UpdateTask(ctx, &pb.UpdateTaskReq{
		TaskId: task.TaskId,
		State:  "PAUSED",
		Result: string(stateJson),
	})
	w.taskLog(task.TaskId, LevelInfo, "Task %s progress saved: completedPhases=%v, assets=%d", task.TaskId, phases, len(assets))
}

// createTaskContext 创建带有任务控制信号检查的上下文
// 当任务被停止时，上下文会被取消
func (w *Worker) createTaskContext(parentCtx context.Context, taskId string) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(parentCtx)
	
	// 启动一个goroutine定期检查任务控制信号
	go func() {
		ticker := time.NewTicker(1 * time.Second) // 每秒检查一次
		defer ticker.Stop()
		
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if ctrl := w.checkTaskControl(ctx, taskId); ctrl == "STOP" {
					w.taskLog(taskId, LevelInfo, "Task %s received stop signal, cancelling context", taskId)
					cancel()
					return
				}
			}
		}
	}()
	
	return ctx, cancel
}

// executeTask 执行任务
func (w *Worker) executeTask(task *scheduler.TaskInfo) {
	baseCtx := context.Background()
	startTime := time.Now()

	w.mu.Lock()
	w.taskStarted++
	w.mu.Unlock()

	// 检查是否有停止信号（任务可能在队列中被停止�?
	if ctrl := w.checkTaskControl(baseCtx, task.TaskId); ctrl == "STOP" {
		w.taskLog(task.TaskId, LevelInfo, "Task %s was stopped before execution", task.TaskId)
		return
	}

	// 创建带有任务控制信号检查的上下文
	ctx, cancelTask := w.createTaskContext(baseCtx, task.TaskId)
	defer cancelTask()

	// 更新任务状态为开�?
	w.updateTaskStatus(ctx, task.TaskId, scheduler.TaskStatusStarted, "")

	// 解析任务配置
	var taskConfig map[string]interface{}
	if err := json.Unmarshal([]byte(task.Config), &taskConfig); err != nil {
		w.updateTaskStatus(ctx, task.TaskId, scheduler.TaskStatusFailure, "配置解析失败: "+err.Error())
		return
	}

	// 检查任务类型，处理POC验证任务
	taskType, _ := taskConfig["taskType"].(string)
	if taskType == "poc_validate" {
		w.executePocValidateTask(ctx, task, taskConfig, startTime)
		return
	}

	// 获取目标
	target, _ := taskConfig["target"].(string)
	if target == "" {
		w.updateTaskStatus(ctx, task.TaskId, scheduler.TaskStatusFailure, "目标为空")
		return
	}

	var allAssets []*scanner.Asset
	var allVuls []*scanner.Vulnerability

	// 解析扫描配置
	config, _ := scheduler.ParseTaskConfig(task.Config)
	if config == nil {
		config = &scheduler.TaskConfig{
			PortScan: &scheduler.PortScanConfig{Enable: true, Ports: "80,443,8080"},
		}
	}
	// 输出端口阈值配置，方便调试
	if config.PortScan != nil {
		w.taskLog(task.TaskId, LevelInfo, "Port threshold config: %d (0=no filter)", config.PortScan.PortThreshold)
	}

	// 解析恢复状态（如果是继续执行的任务）
	var resumeState map[string]interface{}
	if stateStr, ok := taskConfig["resumeState"].(string); ok && stateStr != "" {
		json.Unmarshal([]byte(stateStr), &resumeState)
		w.taskLog(task.TaskId, LevelInfo, "Resuming task from saved state: %v", resumeState)
	}
	completedPhases := make(map[string]bool)
	if resumeState != nil {
		if phases, ok := resumeState["completedPhases"].([]interface{}); ok {
			for _, p := range phases {
				if ps, ok := p.(string); ok {
					completedPhases[ps] = true
				}
			}
		}
		// 恢复已扫描的资产
		if assetsJson, ok := resumeState["assets"].(string); ok && assetsJson != "" {
			json.Unmarshal([]byte(assetsJson), &allAssets)
			w.taskLog(task.TaskId, LevelInfo, "Restored %d assets from saved state", len(allAssets))
		}
	}

	// 执行端口扫描
	if (config.PortScan == nil || config.PortScan.Enable) && !completedPhases["portscan"] {
		// 检查控制信号
		if ctrl := w.checkTaskControl(ctx, task.TaskId); ctrl == "STOP" {
			w.taskLog(task.TaskId, LevelInfo, "Task %s stopped during port scan phase", task.TaskId)
			return
		} else if ctrl == "PAUSE" {
			w.taskLog(task.TaskId, LevelInfo, "Task %s paused during port scan phase", task.TaskId)
			return
		}

		// 根据配置选择端口发现工具（默认使用Naabu�?
		portDiscoveryTool := "naabu"
		if config.PortScan != nil && config.PortScan.Tool != "" {
			portDiscoveryTool = config.PortScan.Tool
		}

		var openPorts []*scanner.Asset
		
		// 第一步：端口发现（Naabu �?Masscan�?
		switch portDiscoveryTool {
		case "masscan":
			w.taskLog(task.TaskId, LevelInfo, "Phase 1: Running Masscan for fast port discovery on target: %s", target)
			masscanScanner := w.scanners["masscan"]
			masscanResult, err := masscanScanner.Scan(ctx, &scanner.ScanConfig{
				Target:  target,
				Options: config.PortScan,
			})
			if err != nil {
				w.taskLog(task.TaskId, LevelError, "Masscan error: %v", err)
			}
			if masscanResult != nil && len(masscanResult.Assets) > 0 {
				openPorts = filterByPortThreshold(masscanResult.Assets, config.PortScan.PortThreshold)
				w.taskLog(task.TaskId, LevelInfo, "Masscan found %d open ports (filtered from %d)", len(openPorts), len(masscanResult.Assets))
			}
		default: // naabu
			w.taskLog(task.TaskId, LevelInfo, "Phase 1: Running Naabu for fast port discovery on target: %s", target)
			naabuScanner := w.scanners["naabu"]
			naabuResult, err := naabuScanner.Scan(ctx, &scanner.ScanConfig{
				Target:  target,
				Options: config.PortScan,
			})
			if err != nil {
				w.taskLog(task.TaskId, LevelError, "Naabu error: %v", err)
			}
			if naabuResult != nil && len(naabuResult.Assets) > 0 {
				openPorts = filterByPortThreshold(naabuResult.Assets, config.PortScan.PortThreshold)
				w.taskLog(task.TaskId, LevelInfo, "Naabu found %d open ports (filtered from %d)", len(openPorts), len(naabuResult.Assets))
			}
		}
		
		// 第二步：Nmap 对存活端口进行服务识�?
		if len(openPorts) > 0 {
			w.taskLog(task.TaskId, LevelInfo, "Phase 2: Running Nmap for service detection on %d open ports", len(openPorts))
			
			// 按主机分组端�?
			hostPorts := make(map[string][]int)
			for _, asset := range openPorts {
				hostPorts[asset.Host] = append(hostPorts[asset.Host], asset.Port)
			}
			
			nmapScanner := w.scanners["nmap"]
			for host, ports := range hostPorts {
				// 构建端口字符�?
				portStrs := make([]string, len(ports))
				for i, p := range ports {
					portStrs[i] = fmt.Sprintf("%d", p)
				}
				portsStr := strings.Join(portStrs, ",")
				
				w.taskLog(task.TaskId, LevelInfo, "Running Nmap on %s with ports: %s", host, portsStr)
				
				nmapResult, err := nmapScanner.Scan(ctx, &scanner.ScanConfig{
					Target: host,
					Options: &scanner.NmapOptions{
						Ports:   portsStr,
						Timeout: config.PortScan.Timeout,
					},
				})
				
				if err != nil {
					w.taskLog(task.TaskId, LevelError, "Nmap error for %s: %v", host, err)
					// Nmap失败时，使用端口发现阶段的结�?
					for _, asset := range openPorts {
						if asset.Host == host {
							asset.IsHTTP = scanner.IsHTTPService(asset.Service, asset.Port)
							allAssets = append(allAssets, asset)
						}
					}
					continue
				}
				
				if nmapResult != nil && len(nmapResult.Assets) > 0 {
					// 设置 IsHTTP 字段
					for _, asset := range nmapResult.Assets {
						asset.IsHTTP = scanner.IsHTTPService(asset.Service, asset.Port)
					}
					allAssets = append(allAssets, nmapResult.Assets...)
				} else {
					// Nmap没有结果时，使用端口发现阶段的结�?
					for _, asset := range openPorts {
						if asset.Host == host {
							asset.IsHTTP = scanner.IsHTTPService(asset.Service, asset.Port)
							allAssets = append(allAssets, asset)
						}
					}
				}
			}
			
			w.taskLog(task.TaskId, LevelInfo, "Port scan completed: %d assets with service info", len(allAssets))
			
			// 端口扫描完成后立即保存结果
			if len(allAssets) > 0 {
				w.taskLog(task.TaskId, LevelInfo, "Saving port scan results immediately...")
				w.saveAssetResult(ctx, task.WorkspaceId, task.MainTaskId, allAssets)
			}
		} else {
			w.taskLog(task.TaskId, LevelInfo, "No open ports found by %s", portDiscoveryTool)
		}
		
		completedPhases["portscan"] = true
	}

	// 检查控制信号
	if ctrl := w.checkTaskControl(ctx, task.TaskId); ctrl == "STOP" {
		w.taskLog(task.TaskId, LevelInfo, "Task %s stopped after port scan", task.TaskId)
		return
	} else if ctrl == "PAUSE" {
		w.taskLog(task.TaskId, LevelInfo, "Task %s paused after port scan, saving state...", task.TaskId)
		w.saveTaskProgress(ctx, task, completedPhases, allAssets)
		return
	}

	// 执行指纹识别
	if config.Fingerprint != nil && config.Fingerprint.Enable && len(allAssets) > 0 && !completedPhases["fingerprint"] {
		if s, ok := w.scanners["fingerprint"]; ok {
			w.taskLog(task.TaskId, LevelInfo, "Running fingerprint scan on %d assets", len(allAssets))
			
			// 每次扫描前实时加载HTTP服务映射配置（类似POC扫描方式�?
			w.loadHttpServiceMappings()
			
			// 如果启用自定义指纹引擎，加载自定义指�?
			if config.Fingerprint.CustomEngine {
				w.loadCustomFingerprints(ctx, s.(*scanner.FingerprintScanner))
			}
			
			result, err := s.Scan(ctx, &scanner.ScanConfig{
				Assets:  allAssets,
				Options: config.Fingerprint,
			})
			
			// 检查是否被取消
			if ctx.Err() != nil {
				w.taskLog(task.TaskId, LevelInfo, "Task %s stopped during fingerprint scan", task.TaskId)
				return
			}
			
			if err == nil && result != nil {
				// 构建 Host:Port -> Asset 的映射，用于匹配指纹结果
				assetMap := make(map[string]*scanner.Asset)
				for _, asset := range allAssets {
					key := fmt.Sprintf("%s:%d", asset.Host, asset.Port)
					assetMap[key] = asset
				}
				
				// 通过 Host:Port 匹配来更新资产信息，而不是按索引
				for _, fpAsset := range result.Assets {
					key := fmt.Sprintf("%s:%d", fpAsset.Host, fpAsset.Port)
					if originalAsset, ok := assetMap[key]; ok {
						originalAsset.Service = fpAsset.Service
						originalAsset.Title = fpAsset.Title
						originalAsset.App = fpAsset.App
						originalAsset.HttpStatus = fpAsset.HttpStatus
						originalAsset.HttpHeader = fpAsset.HttpHeader
						originalAsset.HttpBody = fpAsset.HttpBody
						originalAsset.Server = fpAsset.Server
						originalAsset.IconHash = fpAsset.IconHash
						originalAsset.Screenshot = fpAsset.Screenshot
					}
				}
				
				// 指纹识别完成后保存更新结果（会以更新方式合并到已有资产）
				w.taskLog(task.TaskId, LevelInfo, "Saving fingerprint results...")
				w.saveAssetResult(ctx, task.WorkspaceId, task.MainTaskId, allAssets)
			}
		}
		completedPhases["fingerprint"] = true
	}

	// 检查控制信号
	if ctrl := w.checkTaskControl(ctx, task.TaskId); ctrl == "STOP" {
		w.taskLog(task.TaskId, LevelInfo, "Task %s stopped after fingerprint scan", task.TaskId)
		return
	} else if ctrl == "PAUSE" {
		w.taskLog(task.TaskId, LevelInfo, "Task %s paused after fingerprint scan, saving state...", task.TaskId)
		w.saveTaskProgress(ctx, task, completedPhases, allAssets)
		return
	}

	// 执行POC扫描 (使用Nuclei引擎)
	if config.PocScan != nil && config.PocScan.Enable && len(allAssets) > 0 && !completedPhases["pocscan"] {
		if s, ok := w.scanners["nuclei"]; ok {
			w.taskLog(task.TaskId, LevelInfo, "Running Nuclei POC scan on %d assets", len(allAssets))

			// 从数据库获取模板（所有模板都存储在数据库中）
			var templates []string
			var autoTags []string

			// 检查是否有模板ID列表（任务创建时已筛选好的模板）
			if len(config.PocScan.NucleiTemplateIds) > 0 || len(config.PocScan.CustomPocIds) > 0 {
				// 通过RPC根据ID获取模板内容（包括默认模板和自定义POC�?
				templates = w.getTemplatesByIds(ctx, config.PocScan.NucleiTemplateIds, config.PocScan.CustomPocIds)
				w.taskLog(task.TaskId, LevelInfo, "Fetched %d templates by IDs from database (nuclei: %d, custom: %d)", 
					len(templates), len(config.PocScan.NucleiTemplateIds), len(config.PocScan.CustomPocIds))
			} else {
				// 没有预设的模板ID，根据自动扫描配置生成标签并获取模板
				if config.PocScan.AutoScan || config.PocScan.AutomaticScan {
					autoTags = w.generateAutoTags(allAssets, config.PocScan)
					w.taskLog(task.TaskId, LevelInfo, "Auto-scan generated tags: %v", autoTags)
				}

				if len(autoTags) > 0 {
					// 有自动生成的标签，通过RPC获取符合标签的模�?
					severities := []string{}
					if config.PocScan.Severity != "" {
						severities = strings.Split(config.PocScan.Severity, ",")
					}
					templates = w.getTemplatesByTags(ctx, autoTags, severities)
					w.taskLog(task.TaskId, LevelInfo, "Fetched %d templates by tags from database", len(templates))
				} else {
					// 没有模板ID也没有自动标签，记录警告
					w.taskLog(task.TaskId, LevelError, "No template IDs or auto-scan tags provided, POC scan will be skipped")
				}
			}

			// 只有在有模板时才执行扫描
			if len(templates) > 0 {
				// 用于统计漏洞数量
				var vulCount int

				// 创建漏洞缓冲区，�?0个漏洞批量保存一次
				vulBuffer := NewVulnerabilityBuffer(10)

				// 启动后台刷新协程
				flushDone := make(chan struct{})
				go func() {
					defer close(flushDone)
					ticker := time.NewTicker(5 * time.Second) // �?秒也刷新一次
					defer ticker.Stop()
					for {
						select {
						case <-ctx.Done():
							return
						case <-flushDone:
							return
						case <-vulBuffer.flushChan:
							vulBuffer.Flush(ctx, func(vuls []*scanner.Vulnerability) {
								w.saveVulResult(ctx, task.WorkspaceId, task.MainTaskId, vuls)
							})
						case <-ticker.C:
							vulBuffer.Flush(ctx, func(vuls []*scanner.Vulnerability) {
								w.saveVulResult(ctx, task.WorkspaceId, task.MainTaskId, vuls)
							})
						}
					}
				}()

				// 构建Nuclei扫描选项，设置回调函数批量保存漏�?
				taskIdForCallback := task.TaskId // 捕获taskId用于回调
				nucleiOpts := &scanner.NucleiOptions{
					Severity:        config.PocScan.Severity,
					Tags:            autoTags,
					ExcludeTags:     config.PocScan.ExcludeTags,
					RateLimit:       config.PocScan.RateLimit,
					Concurrency:     config.PocScan.Concurrency,
					AutoScan:        false, // 标签已在Worker端生成，不需要nuclei再生�?
					AutomaticScan:   false,
					CustomPocOnly:   config.PocScan.CustomPocOnly,
					CustomTemplates: templates,
					TagMappings:     config.PocScan.TagMappings,
					// 设置回调函数，发现漏洞时添加到缓冲区
					OnVulnerabilityFound: func(vul *scanner.Vulnerability) {
						vulCount++
						w.taskLog(taskIdForCallback, LevelInfo, "Found vulnerability #%d: %s on %s", vulCount, vul.PocFile, vul.Url)
						vulBuffer.Add(vul)
					},
				}
				// 设置默认�?
				if nucleiOpts.RateLimit == 0 {
					nucleiOpts.RateLimit = 150
				}
				if nucleiOpts.Concurrency == 0 {
					nucleiOpts.Concurrency = 25
				}
				w.taskLog(task.TaskId, LevelInfo, "Nuclei options: Templates=%d, Tags=%v", len(nucleiOpts.CustomTemplates), nucleiOpts.Tags)

				result, err := s.Scan(ctx, &scanner.ScanConfig{
					Assets:  allAssets,
					Options: nucleiOpts,
				})

				// 扫描完成后，刷新剩余的漏�?
				vulBuffer.Flush(ctx, func(vuls []*scanner.Vulnerability) {
					w.saveVulResult(ctx, task.WorkspaceId, task.MainTaskId, vuls)
				})

				if err != nil {
					w.taskLog(task.TaskId, LevelError, "POC scan error: %v", err)
				}
				if result != nil {
					allVuls = append(allVuls, result.Vulnerabilities...)
					if vulCount > 0 {
						w.taskLog(task.TaskId, LevelInfo, "POC scan completed: %d vulnerabilities found and saved", vulCount)
					} else {
						w.taskLog(task.TaskId, LevelInfo, "POC scan completed, no vulnerabilities found")
					}
				}
			} else {
				w.taskLog(task.TaskId, LevelInfo, "No templates available, skipping POC scan")
			}
		}
	}

	// 更新任务状态为完成
	duration := time.Since(startTime).Seconds()
	result := fmt.Sprintf("资产:%d 漏洞:%d 耗时:%.0fs", len(allAssets), len(allVuls), duration)
	w.updateTaskStatus(ctx, task.TaskId, scheduler.TaskStatusSuccess, result)
	w.taskLog(task.TaskId, LevelInfo, "Task %s completed: %s", task.TaskId, result)

	w.mu.Lock()
	w.taskExecuted++
	w.mu.Unlock()
}

// updateTaskStatus 更新任务状�?
func (w *Worker) updateTaskStatus(ctx context.Context, taskId, status, result string) {
	_, err := w.rpcClient.UpdateTask(ctx, &pb.UpdateTaskReq{
		TaskId: taskId,
		State:  status,
		Worker: w.config.Name,
		Result: result,
	})
	if err != nil {
		w.taskLog(taskId, LevelError, "update task status failed: %v", err)
	}
}

// saveAssetResult 保存资产结果
func (w *Worker) saveAssetResult(ctx context.Context, workspaceId, mainTaskId string, assets []*scanner.Asset) {
	if len(assets) == 0 {
		return
	}

	w.taskLog(mainTaskId, LevelInfo, "Saving %d assets to workspace: %s", len(assets), workspaceId)
	

	pbAssets := make([]*pb.AssetDocument, 0, len(assets))
	for _, asset := range assets {
		pbAsset := &pb.AssetDocument{
			Authority:  asset.Authority,
			Host:       asset.Host,
			Port:       int32(asset.Port),
			Category:   asset.Category,
			Service:    asset.Service,
			Title:      asset.Title,
			App:        asset.App,
			HttpStatus: asset.HttpStatus,
			HttpHeader: asset.HttpHeader,
			HttpBody:   asset.HttpBody,
			IconHash:   asset.IconHash,
			Screenshot: asset.Screenshot,
			Server:     asset.Server,
			Banner:     asset.Banner,
			IsHttp:     asset.IsHTTP,
		}
		pbAssets = append(pbAssets, pbAsset)
	}

	resp, err := w.rpcClient.SaveTaskResult(ctx, &pb.SaveTaskResultReq{
		WorkspaceId: workspaceId,
		MainTaskId:  mainTaskId,
		Assets:      pbAssets,
	})
	if err != nil {
		w.taskLog(mainTaskId, LevelError, "save asset result failed: %v", err)
	} else {
		w.taskLog(mainTaskId, LevelInfo, "Save asset result: %s", resp.Message)
	}
}

// saveVulResult 保存漏洞结果（支持去重与聚合�?
func (w *Worker) saveVulResult(ctx context.Context, workspaceId, mainTaskId string, vuls []*scanner.Vulnerability) {
	if len(vuls) == 0 {
		return
	}

	pbVuls := make([]*pb.VulDocument, 0, len(vuls))
	for _, vul := range vuls {
		// Debug: 打印证据链数�?
		w.taskLog(mainTaskId, LevelDebug, "[SaveVul] PocFile=%s, CurlCommand len=%d, Request len=%d, Response len=%d",
			vul.PocFile, len(vul.CurlCommand), len(vul.Request), len(vul.Response))

		pbVul := &pb.VulDocument{
			Authority: vul.Authority,
			Host:      vul.Host,
			Port:      int32(vul.Port),
			Url:       vul.Url,
			PocFile:   vul.PocFile,
			Source:    vul.Source,
			Severity:  vul.Severity,
			Result:    vul.Result,
		}

		// 漏洞知识库关联字�?- 使用proto.Float64/String等辅助函�?
		if vul.CvssScore > 0 {
			pbVul.CvssScore = &vul.CvssScore
		}
		if vul.CveId != "" {
			pbVul.CveId = &vul.CveId
		}
		if vul.CweId != "" {
			pbVul.CweId = &vul.CweId
		}
		if vul.Remediation != "" {
			pbVul.Remediation = &vul.Remediation
		}
		if len(vul.References) > 0 {
			pbVul.References = vul.References
		}

		// 证据链字�?- 使用局部变量避免指针问�?
		if vul.MatcherName != "" {
			matcherName := vul.MatcherName
			pbVul.MatcherName = &matcherName
		}
		if len(vul.ExtractedResults) > 0 {
			pbVul.ExtractedResults = vul.ExtractedResults
		}
		if vul.CurlCommand != "" {
			curlCommand := vul.CurlCommand
			pbVul.CurlCommand = &curlCommand
		}
		if vul.Request != "" {
			request := vul.Request
			pbVul.Request = &request
		}
		if vul.Response != "" {
			response := vul.Response
			pbVul.Response = &response
		}
		if vul.ResponseTruncated {
			responseTruncated := vul.ResponseTruncated
			pbVul.ResponseTruncated = &responseTruncated
		}

		// Debug: 确认pbVul中的证据字段
		w.taskLog(mainTaskId, LevelDebug, "[SaveVul] pbVul.CurlCommand=%v, pbVul.Request=%v, pbVul.Response=%v",
			pbVul.CurlCommand != nil, pbVul.Request != nil, pbVul.Response != nil)

		pbVuls = append(pbVuls, pbVul)
	}

	_, err := w.rpcClient.SaveVulResult(ctx, &pb.SaveVulResultReq{
		WorkspaceId: workspaceId,
		MainTaskId:  mainTaskId,
		Vuls:        pbVuls,
	})
	if err != nil {
		w.taskLog(mainTaskId, LevelError, "save vul result failed: %v", err)
	}
}

// reportResult 上报结果
func (w *Worker) reportResult() {
	defer w.wg.Done()

	for {
		select {
		case <-w.stopChan:
			return
		case result := <-w.resultChan:
			w.handleResult(result)
		}
	}
}

// handleResult 处理结果
func (w *Worker) handleResult(result *scanner.ScanResult) {
	ctx := context.Background()
	w.saveAssetResult(ctx, result.WorkspaceId, result.MainTaskId, result.Assets)
	w.saveVulResult(ctx, result.WorkspaceId, result.MainTaskId, result.Vulnerabilities)
}

// keepAlive 心跳
func (w *Worker) keepAlive() {
	defer w.wg.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-w.stopChan:
			return
		case <-ticker.C:
			w.sendHeartbeat()
		}
	}
}

// sendHeartbeat 发送心�?
func (w *Worker) sendHeartbeat() {
	ctx, cancel := context.WithTimeout(w.ctx, 10*time.Second) // 继承父Context
	defer cancel()

	// 获取系统资源使用情况
	cpuPercent, _ := cpu.Percent(time.Second, false)
	memInfo, _ := mem.VirtualMemory()

	cpuLoad := 0.0
	if len(cpuPercent) > 0 {
		cpuLoad = cpuPercent[0]
	}
	memUsed := 0.0
	if memInfo != nil {
		memUsed = memInfo.UsedPercent
	}

	// 确保数值有�?
	if cpuLoad < 0 || cpuLoad > 100 {
		cpuLoad = 0.0
	}
	if memUsed < 0 || memUsed > 100 {
		memUsed = 0.0
	}

	// 计算正在执行的任务数（已开始但未完成的任务）
	w.mu.Lock()
	runningTasks := w.taskStarted - w.taskExecuted
	if runningTasks < 0 {
		runningTasks = 0
	}
	w.mu.Unlock()

	resp, err := w.rpcClient.KeepAlive(ctx, &pb.KeepAliveReq{
		WorkerName:         w.config.Name,
		CpuLoad:            cpuLoad,
		MemUsed:            memUsed,
		TaskStartedNumber:  int32(w.taskStarted),
		TaskExecutedNumber: int32(w.taskExecuted),
		IsDaemon:           false,
	})
	if err != nil {
		w.logger.Error("keepalive failed: %v", err)
		return
	}

	// 处理控制指令
	if resp.ManualStopFlag {
		w.logger.Info("received stop signal, stopping worker...")
		w.Stop()
		os.Exit(0)
	}
	if resp.ManualReloadFlag {
		w.logger.Info("received reload signal")
		// 重新加载配置
	}
}

// subscribeStatusQuery 订阅状态查询请�?
func (w *Worker) subscribeStatusQuery() {
	defer w.wg.Done()

	ctx := context.Background()
	pubsub := w.redisClient.Subscribe(ctx, "cscan:worker:query")
	defer pubsub.Close()

	ch := pubsub.Channel()
	w.logger.Info("Worker %s subscribed to status query channel", w.config.Name)

	for {
		select {
		case <-w.stopChan:
			return
		case msg := <-ch:
			if msg != nil {
				// 收到查询请求，立即上报状�?
				w.reportStatusToRedis()
			}
		}
	}
}

// reportStatusToRedis 立即上报状态到Redis
func (w *Worker) reportStatusToRedis() {
	if w.redisClient == nil {
		return
	}

	ctx := context.Background()

	// 快速获取CPU使用率（不等�?秒）
	cpuPercent, _ := cpu.Percent(0, false)
	memInfo, _ := mem.VirtualMemory()

	cpuLoad := 0.0
	if len(cpuPercent) > 0 {
		cpuLoad = cpuPercent[0]
	}
	memUsed := 0.0
	if memInfo != nil {
		memUsed = memInfo.UsedPercent
	}

	// 确保数值有�?
	if cpuLoad < 0 || cpuLoad > 100 {
		cpuLoad = 0.0
	}
	if memUsed < 0 || memUsed > 100 {
		memUsed = 0.0
	}

	w.mu.Lock()
	taskStarted := w.taskStarted
	taskExecuted := w.taskExecuted
	w.mu.Unlock()

	// 保存状态到Redis
	key := fmt.Sprintf("worker:%s", w.config.Name)
	status := map[string]interface{}{
		"workerName":         w.config.Name,
		"cpuLoad":            cpuLoad,
		"memUsed":            memUsed,
		"taskStartedNumber":  taskStarted,
		"taskExecutedNumber": taskExecuted,
		"isDaemon":           false,
		"updateTime":         time.Now().Format("2006-01-02 15:04:05"),
	}

	data, _ := json.Marshal(status)
	w.redisClient.Set(ctx, key, data, 10*time.Minute)
}

// GetWorkerName 获取Worker名称
func GetWorkerName() string {
	hostname, _ := os.Hostname()
	return fmt.Sprintf("%s-%d", hostname, os.Getpid())
}

// GetSystemInfo 获取系统信息
func GetSystemInfo() map[string]interface{} {
	return map[string]interface{}{
		"os":       runtime.GOOS,
		"arch":     runtime.GOARCH,
		"cpus":     runtime.NumCPU(),
		"hostname": func() string { h, _ := os.Hostname(); return h }(),
	}
}

// generateAutoTags 根据资产的应用信息生成Nuclei标签
func (w *Worker) generateAutoTags(assets []*scanner.Asset, pocConfig *scheduler.PocScanConfig) []string {
	tagSet := make(map[string]bool)

	for _, asset := range assets {
		for _, app := range asset.App {
			appName := parseAppName(app)
			appNameLower := strings.ToLower(appName)

			// 模式1: 基于自定义标签映�?
			if pocConfig.AutoScan && pocConfig.TagMappings != nil {
				for mappedApp, tags := range pocConfig.TagMappings {
					if strings.ToLower(mappedApp) == appNameLower {
						for _, tag := range tags {
							tagSet[tag] = true
						}
						break
					}
				}
			}

			// 模式2: 基于Wappalyzer内置映射（类似nuclei -as�?
			if pocConfig.AutomaticScan {
				if tags, ok := mapping.WappalyzerNucleiMapping[appNameLower]; ok {
					for _, tag := range tags {
						tagSet[tag] = true
					}
				}
			}
		}
	}

	tags := make([]string, 0, len(tagSet))
	for tag := range tagSet {
		tags = append(tags, tag)
	}
	return tags
}

// getTemplatesByTags 通过RPC从数据库获取符合标签的模�?
func (w *Worker) getTemplatesByTags(ctx context.Context, tags []string, severities []string) []string {
	if len(tags) == 0 {
		return nil
	}

	resp, err := w.rpcClient.GetTemplatesByTags(ctx, &pb.GetTemplatesByTagsReq{
		Tags:       tags,
		Severities: severities,
	})
	if err != nil {
		w.logger.Error("GetTemplatesByTags RPC failed: %v", err)
		return nil
	}

	if !resp.Success {
		w.logger.Error("GetTemplatesByTags failed: %s", resp.Message)
		return nil
	}

	w.logger.Info("GetTemplatesByTags: fetched %d templates for tags %v", resp.Count, tags)
	return resp.Templates
}

// getTemplatesByIds 通过RPC根据ID列表获取模板内容
func (w *Worker) getTemplatesByIds(ctx context.Context, nucleiTemplateIds, customPocIds []string) []string {
	if len(nucleiTemplateIds) == 0 && len(customPocIds) == 0 {
		return nil
	}

	resp, err := w.rpcClient.GetTemplatesByIds(ctx, &pb.GetTemplatesByIdsReq{
		NucleiTemplateIds: nucleiTemplateIds,
		CustomPocIds:      customPocIds,
	})
	if err != nil {
		w.logger.Error("GetTemplatesByIds RPC failed: %v", err)
		return nil
	}

	if !resp.Success {
		w.logger.Error("GetTemplatesByIds failed: %s", resp.Message)
		return nil
	}

	w.logger.Info("GetTemplatesByIds: fetched %d templates", resp.Count)
	return resp.Templates
}

// parseAppName 解析应用名称，去除版本号和来源标�?
func parseAppName(app string) string {
	appName := app
	// 先去�?[source] 后缀
	if idx := strings.Index(appName, "["); idx > 0 {
		appName = appName[:idx]
	}
	// 再去�?:version 后缀
	if idx := strings.Index(appName, ":"); idx > 0 {
		appName = appName[:idx]
	}
	return strings.TrimSpace(appName)
}

// loadCustomFingerprints 加载自定义指纹到指纹扫描�?
func (w *Worker) loadCustomFingerprints(ctx context.Context, fpScanner *scanner.FingerprintScanner) {
	resp, err := w.rpcClient.GetCustomFingerprints(ctx, &pb.GetCustomFingerprintsReq{
		EnabledOnly: true,
	})
	if err != nil {
		w.logger.Error("GetCustomFingerprints RPC failed: %v", err)
		return
	}

	if !resp.Success {
		w.logger.Error("GetCustomFingerprints failed: %s", resp.Message)
		return
	}

	if len(resp.Fingerprints) == 0 {
		w.logger.Info("No custom fingerprints found")
		return
	}

	// 转换为model.Fingerprint
	var fingerprints []*model.Fingerprint
	for _, fp := range resp.Fingerprints {
		mfp := &model.Fingerprint{
			Name:      fp.Name,
			Category:  fp.Category,
			Rule:      fp.Rule,
			Source:    fp.Source,
			Headers:   fp.Headers,
			Cookies:   fp.Cookies,
			HTML:      fp.Html,
			Scripts:   fp.Scripts,
			ScriptSrc: fp.ScriptSrc,
			Meta:      fp.Meta,
			CSS:       fp.Css,
			URL:       fp.Url,
			IsBuiltin: fp.IsBuiltin,
			Enabled:   fp.Enabled,
		}
		// 解析ID
		if fp.Id != "" {
			if oid, err := primitive.ObjectIDFromHex(fp.Id); err == nil {
				mfp.Id = oid
			}
		}
		fingerprints = append(fingerprints, mfp)
	}

	// 创建自定义指纹引擎并设置到扫描器
	customEngine := scanner.NewCustomFingerprintEngine(fingerprints)
	fpScanner.SetCustomFingerprintEngine(customEngine)
	w.logger.Info("Loaded %d fingerprints (builtin + custom) into fingerprint scanner", len(fingerprints))
}

// filterByPortThreshold 根据端口阈值过滤资�?
// 如果某个主机开放的端口数量超过阈值，则过滤掉该主机的所有资产（可能是防火墙或蜜罐）
func filterByPortThreshold(assets []*scanner.Asset, threshold int) []*scanner.Asset {
	if threshold <= 0 {
		return assets // 阈值为0或负数表示不过滤
	}

	// 统计每个主机的开放端口数�?
	hostPortCount := make(map[string]int)
	for _, asset := range assets {
		hostPortCount[asset.Host]++
	}

	// 找出需要过滤的主机
	filteredHosts := make(map[string]bool)
	for host, count := range hostPortCount {
		if count > threshold {
			filteredHosts[host] = true
			// 这里使用 logx 因为没有 Worker 上下文
			logx.Infof("Host %s has %d open ports (threshold: %d), filtered as potential honeypot/firewall", host, count, threshold)
		}
	}

	// 过滤资产
	if len(filteredHosts) == 0 {
		return assets
	}

	result := make([]*scanner.Asset, 0, len(assets))
	for _, asset := range assets {
		if !filteredHosts[asset.Host] {
			result = append(result, asset)
		}
	}
	return result
}

// executePocValidateTask 执行POC验证任务
func (w *Worker) executePocValidateTask(ctx context.Context, task *scheduler.TaskInfo, taskConfig map[string]interface{}, startTime time.Time) {
	// 解析配置
	url, _ := taskConfig["url"].(string)
	pocId, _ := taskConfig["pocId"].(string)
	pocType, _ := taskConfig["pocType"].(string)
	timeout, _ := taskConfig["timeout"].(float64)
	batchId, _ := taskConfig["batchId"].(string)

	// 立即输出任务接收日志
	w.taskLog(task.TaskId, LevelInfo, "[%s] 收到POC验证任务, 目标: %s", task.TaskId, url)

	if url == "" {
		w.taskLog(task.TaskId, LevelError, "[%s] POC验证失败: URL为空", task.TaskId)
		w.updateTaskStatus(ctx, task.TaskId, scheduler.TaskStatusFailure, "URL为空")
		w.savePocValidationResult(ctx, task.TaskId, batchId, nil, "URL为空")
		return
	}

	if timeout == 0 {
		timeout = 30
	}

	// 获取Nuclei扫描器
	nucleiScanner, ok := w.scanners["nuclei"]
	if !ok {
		w.taskLog(task.TaskId, LevelError, "[%s] POC验证失败: Nuclei扫描器未初始化", task.TaskId)
		w.updateTaskStatus(ctx, task.TaskId, scheduler.TaskStatusFailure, "Nuclei扫描器未初始化")
		w.savePocValidationResult(ctx, task.TaskId, batchId, nil, "Nuclei扫描器未初始化")
		return
	}

	// 获取POC模板
	var templates []string
	var pocName string
	var pocSeverity string

	// 如果指定了pocId，通过RPC获取POC内容
	if pocId != "" {
		w.taskLog(task.TaskId, LevelInfo, "[%s] 正在加载POC模板...", task.TaskId)
		resp, err := w.rpcClient.GetPocById(ctx, &pb.GetPocByIdReq{
			PocId:   pocId,
			PocType: pocType,
		})
		if err != nil {
			w.taskLog(task.TaskId, LevelError, "[%s] POC验证失败: 获取POC失败 - %v", task.TaskId, err)
			w.updateTaskStatus(ctx, task.TaskId, scheduler.TaskStatusFailure, "获取POC失败: "+err.Error())
			w.savePocValidationResult(ctx, task.TaskId, batchId, nil, "获取POC失败: "+err.Error())
			return
		}
		if !resp.Success {
			w.taskLog(task.TaskId, LevelError, "[%s] POC验证失败: POC不存�?- %s", task.TaskId, resp.Message)
			w.updateTaskStatus(ctx, task.TaskId, scheduler.TaskStatusFailure, "POC不存�? "+resp.Message)
			w.savePocValidationResult(ctx, task.TaskId, batchId, nil, "POC不存�? "+resp.Message)
			return
		}
		if resp.Content == "" {
			w.taskLog(task.TaskId, LevelError, "[%s] POC验证失败: POC内容为空", task.TaskId)
			w.updateTaskStatus(ctx, task.TaskId, scheduler.TaskStatusFailure, "POC内容为空")
			w.savePocValidationResult(ctx, task.TaskId, batchId, nil, "POC内容为空")
			return
		}
		templates = []string{resp.Content}
		pocName = resp.Name
		pocSeverity = resp.Severity
		pocType = resp.PocType
		w.taskLog(task.TaskId, LevelInfo, "[%s] POC模板加载完成: %s", task.TaskId, pocName)
	} else {
		// 没有指定pocId，尝试通过标签获取模板
		var severities []string
		var tags []string

		// 解析严重级别
		if sevList, ok := taskConfig["severities"].([]interface{}); ok {
			for _, s := range sevList {
				if str, ok := s.(string); ok {
					severities = append(severities, str)
				}
			}
		}

		// 解析标签
		if tagList, ok := taskConfig["tags"].([]interface{}); ok {
			for _, t := range tagList {
				if str, ok := t.(string); ok {
					tags = append(tags, str)
				}
			}
		}

		// 根据标签获取模板
		if len(tags) > 0 {
			templates = w.getTemplatesByTags(ctx, tags, severities)
		}

		if len(templates) == 0 {
			w.taskLog(task.TaskId, LevelError, "[%s] POC验证失败: 未找到POC模板", task.TaskId)
			w.updateTaskStatus(ctx, task.TaskId, scheduler.TaskStatusFailure, "未找到POC模板")
			w.savePocValidationResult(ctx, task.TaskId, batchId, nil, "未找到POC模板")
			return
		}
	}

	// 输出开始扫描日�?
	w.taskLog(task.TaskId, LevelInfo, "[%s] 正在初始化Nuclei扫描引擎...", task.TaskId)

	// 构建Nuclei扫描选项
	nucleiOpts := &scanner.NucleiOptions{
		RateLimit:       50,
		Concurrency:     10,
		CustomTemplates: templates,
		CustomPocOnly:   true, // 只使用自定义POC
	}

	w.taskLog(task.TaskId, LevelInfo, "[%s] 开始扫描目�? %s", task.TaskId, url)

	// 执行扫描 - 直接传递URL作为目标，不通过Asset构建
	result, err := nucleiScanner.Scan(ctx, &scanner.ScanConfig{
		Targets: []string{url}, // 直接使用URL作为目标
		Options: nucleiOpts,
	})

	duration := time.Since(startTime).Seconds()

	if err != nil {
		w.taskLog(task.TaskId, LevelError, "[%s] POC验证失败: %v", task.TaskId, err)
		w.updateTaskStatus(ctx, task.TaskId, scheduler.TaskStatusFailure, fmt.Sprintf("扫描失败: %v", err))
		w.savePocValidationResult(ctx, task.TaskId, batchId, nil, fmt.Sprintf("扫描失败: %v", err))
		return
	}

	// 构建验证结果
	var validationResults []*PocValidationResult
	matched := false
	vulCount := 0
	if result != nil {
		vulCount = len(result.Vulnerabilities)
	}

	w.taskLog(task.TaskId, LevelInfo, "[%s] 扫描完成, 耗时: %.2fs", task.TaskId, duration)

	if result != nil && len(result.Vulnerabilities) > 0 {
		matched = true
		for _, vul := range result.Vulnerabilities {
			// 优先使用配置中的POC信息
			resultPocName := pocName
			resultSeverity := pocSeverity
			if resultPocName == "" {
				resultPocName = vul.PocFile
			}
			if resultSeverity == "" {
				resultSeverity = vul.Severity
			}
			validationResults = append(validationResults, &PocValidationResult{
				PocId:      pocId,
				PocName:    resultPocName,
				TemplateId: pocId,
				Severity:   resultSeverity,
				Matched:    true,
				MatchedUrl: vul.Url,
				Details:    vul.Result,
				Output:     vul.Extra,
				PocType:    pocType,
			})
			logx.Infof("[%s] 发现漏洞! 匹配URL: %s", task.TaskId, vul.Url)
			w.taskLog(task.TaskId, LevelInfo, "[%s] 发现漏洞! 匹配URL: %s", task.TaskId, vul.Url)
		}
	} else {
		// 没有发现漏洞，添加一个未匹配的结�?
		resultPocName := pocName
		if resultPocName == "" {
			resultPocName = pocId
		}
		validationResults = append(validationResults, &PocValidationResult{
			PocId:      pocId,
			PocName:    resultPocName,
			Severity:   pocSeverity,
			Matched:    false,
			MatchedUrl: url,
			Details:    "未发现漏洞",
			PocType:    pocType,
		})
		w.taskLog(task.TaskId, LevelInfo, "[%s] 未发现漏洞", task.TaskId)
	}

	// 保存结果到Redis
	w.savePocValidationResult(ctx, task.TaskId, batchId, validationResults, "")

	// 更新任务状态
	resultMsg := fmt.Sprintf("验证完成: 匹配=%v, 漏洞=%d, 耗时=%.2fs", matched, vulCount, duration)
	w.updateTaskStatus(ctx, task.TaskId, scheduler.TaskStatusSuccess, resultMsg)

	w.mu.Lock()
	w.taskExecuted++
	w.mu.Unlock()
}

// PocValidationResult POC验证结果
type PocValidationResult struct {
	PocId      string   `json:"pocId"`
	PocName    string   `json:"pocName"`
	TemplateId string   `json:"templateId"`
	Severity   string   `json:"severity"`
	Matched    bool     `json:"matched"`
	MatchedUrl string   `json:"matchedUrl"`
	Details    string   `json:"details"`
	Output     string   `json:"output"`
	PocType    string   `json:"pocType"`
	Tags       []string `json:"tags"`
}

// savePocValidationResult 保存POC验证结果到Redis
func (w *Worker) savePocValidationResult(ctx context.Context, taskId, batchId string, results []*PocValidationResult, errorMsg string) {
	if w.redisClient == nil {
		w.logger.Error("Redis client not available, cannot save POC validation result")
		return
	}

	// 构建结果数据
	resultData := map[string]interface{}{
		"taskId":     taskId,
		"batchId":    batchId,
		"status":     "SUCCESS",
		"results":    results,
		"updateTime": time.Now().Format("2006-01-02 15:04:05"),
	}

	if errorMsg != "" {
		resultData["status"] = "FAILURE"
		resultData["error"] = errorMsg
	}

	resultJson, err := json.Marshal(resultData)
	if err != nil {
		w.taskLog(taskId, LevelError, "Failed to marshal POC validation result: %v", err)
		return
	}

	// 保存到Redis
	resultKey := fmt.Sprintf("cscan:task:result:%s", taskId)
	err = w.redisClient.Set(ctx, resultKey, resultJson, 24*time.Hour).Err()
	if err != nil {
		w.taskLog(taskId, LevelError, "Failed to save POC validation result to Redis: %v", err)
		return
	}

	// 更新任务信息状�?
	taskInfoKey := fmt.Sprintf("cscan:task:info:%s", taskId)
	taskInfoData, err := w.redisClient.Get(ctx, taskInfoKey).Result()
	if err == nil && taskInfoData != "" {
		var taskInfo map[string]string
		if json.Unmarshal([]byte(taskInfoData), &taskInfo) == nil {
			if errorMsg != "" {
				taskInfo["status"] = "FAILURE"
			} else {
				taskInfo["status"] = "SUCCESS"
			}
			taskInfo["updateTime"] = time.Now().Format("2006-01-02 15:04:05")
			updatedInfo, _ := json.Marshal(taskInfo)
			w.redisClient.Set(ctx, taskInfoKey, updatedInfo, 24*time.Hour)
		}
	}
}

// WorkerHttpServiceChecker Worker端的HTTP服务检查器实现
type WorkerHttpServiceChecker struct {
	cache map[string]bool // serviceName -> isHttp
	mu    sync.RWMutex
}

// NewWorkerHttpServiceChecker 创建HTTP服务检查器
func NewWorkerHttpServiceChecker() *WorkerHttpServiceChecker {
	return &WorkerHttpServiceChecker{
		cache: make(map[string]bool),
	}
}

// IsHttpService 判断服务是否为HTTP服务
func (c *WorkerHttpServiceChecker) IsHttpService(serviceName string) (isHttp bool, found bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	isHttp, found = c.cache[serviceName]
	return
}

// SetMapping 设置服务映射
func (c *WorkerHttpServiceChecker) SetMapping(serviceName string, isHttp bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache[serviceName] = isHttp
}

// loadHttpServiceMappings 从RPC服务加载HTTP服务映射配置
func (w *Worker) loadHttpServiceMappings() {
	ctx := context.Background()

	resp, err := w.rpcClient.GetHttpServiceMappings(ctx, &pb.GetHttpServiceMappingsReq{
		EnabledOnly: true,
	})
	if err != nil {
		w.logger.Error("GetHttpServiceMappings RPC failed: %v, using default mappings", err)
		return
	}

	if !resp.Success {
		w.logger.Error("GetHttpServiceMappings failed: %s, using default mappings", resp.Message)
		return
	}

	if len(resp.Mappings) == 0 {
		w.logger.Info("No HTTP service mappings found, using default mappings")
		return
	}

	// 创建检查器并设置映�?
	checker := NewWorkerHttpServiceChecker()
	for _, mapping := range resp.Mappings {
		checker.SetMapping(mapping.ServiceName, mapping.IsHttp)
	}

	// 设置全局检查器
	scanner.SetHttpServiceChecker(checker)
	w.logger.Info("Loaded %d HTTP service mappings from database", len(resp.Mappings))
}
