package scanner

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/projectdiscovery/goflags"
	"github.com/projectdiscovery/naabu/v2/pkg/result"
	"github.com/projectdiscovery/naabu/v2/pkg/runner"
	"github.com/zeromicro/go-zero/core/logx"
)

// ErrPortThresholdExceeded 端口阈值超过错误
var ErrPortThresholdExceeded = errors.New("port threshold exceeded")

// NaabuScanner Naabu端口扫描器
type NaabuScanner struct {
	BaseScanner
}

// NewNaabuScanner 创建Naabu扫描器
func NewNaabuScanner() *NaabuScanner {
	return &NaabuScanner{
		BaseScanner: BaseScanner{name: "naabu"},
	}
}

// NaabuOptions Naabu扫描选项
type NaabuOptions struct {
	Ports             string `json:"ports"`
	Rate              int    `json:"rate"`
	Timeout           int    `json:"timeout"`
	ScanType          string `json:"scanType"`          // s=SYN, c=CONNECT，默认 c
	PortThreshold     int    `json:"portThreshold"`     // 端口阈值，使用 naabu 原生 -port-threshold 参数
	SkipHostDiscovery bool   `json:"skipHostDiscovery"` // 跳过主机发现 (-Pn)
}

// Scan 执行Naabu扫描
func (s *NaabuScanner) Scan(ctx context.Context, config *ScanConfig) (*ScanResult, error) {
	// 默认配置
	opts := &NaabuOptions{
		Ports:         "80,443,8080",
		Rate:          1000,
		Timeout:       60,  // 单个目标扫描超时，默认60秒
		ScanType:      "c", // 默认 CONNECT 扫描（无需 root 权限）
		PortThreshold: 0,   // 默认不限制
	}

	// 日志函数，优先使用任务日志回调
	logInfo := func(format string, args ...interface{}) {
		if config.TaskLogger != nil {
			config.TaskLogger("INFO", format, args...)
		}
		logx.Infof(format, args...)
	}
	logWarn := func(format string, args ...interface{}) {
		if config.TaskLogger != nil {
			config.TaskLogger("WARN", format, args...)
		}
		logx.Infof(format, args...)
	}
	
	// 进度回调
	onProgress := config.OnProgress

	// 从配置中提取选项
	if config.Options != nil {
		switch v := config.Options.(type) {
		case *NaabuOptions:
			opts = v
		case *PortScanOptions:
			if v.Ports != "" {
				opts.Ports = v.Ports
			}
			if v.Rate > 0 {
				opts.Rate = v.Rate
			}
			if v.Timeout > 0 {
				opts.Timeout = v.Timeout
			}
			if v.PortThreshold > 0 {
				opts.PortThreshold = v.PortThreshold
			}
		default:
			// 尝试通过JSON转换（支持scheduler.PortScanConfig等其他类型）
			if data, err := json.Marshal(config.Options); err == nil {
				var portConfig struct {
					Ports             string `json:"ports"`
					Rate              int    `json:"rate"`
					Timeout           int    `json:"timeout"`
					PortThreshold     int    `json:"portThreshold"`
					ScanType          string `json:"scanType"`
					SkipHostDiscovery bool   `json:"skipHostDiscovery"`
				}
				if err := json.Unmarshal(data, &portConfig); err == nil {
					if portConfig.Ports != "" {
						opts.Ports = portConfig.Ports
					}
					if portConfig.Rate > 0 {
						opts.Rate = portConfig.Rate
					}
					if portConfig.Timeout > 0 {
						opts.Timeout = portConfig.Timeout
					}
					if portConfig.PortThreshold > 0 {
						opts.PortThreshold = portConfig.PortThreshold
					}
					if portConfig.ScanType != "" {
						opts.ScanType = portConfig.ScanType
					}
					opts.SkipHostDiscovery = portConfig.SkipHostDiscovery
				}
			}
		}
	}

	// 解析目标
	targets := parseTargets(config.Target)
	if len(config.Targets) > 0 {
		targets = append(targets, config.Targets...)
	}

	if len(targets) == 0 {
		return &ScanResult{
			WorkspaceId: config.WorkspaceId,
			MainTaskId:  config.MainTaskId,
			Assets:      []*Asset{},
		}, nil
	}

	// 执行Naabu扫描
	assets, thresholdExceeded := s.runNaabuWithLogger(ctx, targets, opts, logInfo, logWarn, onProgress)

	if thresholdExceeded {
		return &ScanResult{
			WorkspaceId: config.WorkspaceId,
			MainTaskId:  config.MainTaskId,
			Assets:      assets,
		}, ErrPortThresholdExceeded
	}

	return &ScanResult{
		WorkspaceId: config.WorkspaceId,
		MainTaskId:  config.MainTaskId,
		Assets:      assets,
	}, nil
}

// logFunc 日志函数类型
type logFunc func(format string, args ...interface{})

// progressFunc 进度回调函数类型
type progressFunc func(progress int, message string)

// runNaabuWithLogger 运行Naabu扫描（带日志回调）
// 按单个目标拆分，串行执行，每个目标独立超时控制
// 返回值: assets - 发现的资产, thresholdExceeded - 是否有任何目标超过端口阈值
func (s *NaabuScanner) runNaabuWithLogger(ctx context.Context, targets []string, opts *NaabuOptions, logInfo, logWarn logFunc, onProgress progressFunc) ([]*Asset, bool) {
	var allAssets []*Asset
	anyThresholdExceeded := false // 记录是否有任何目标超过阈值

	// 处理端口配置
	var portsStr string
	var topPorts string

	switch opts.Ports {
	case "top100":
		topPorts = "100"
	case "top1000":
		topPorts = "1000"
	default:
		ports := parsePorts(opts.Ports)
		portsStr = portsToString(ports)
	}

	timeout := opts.Timeout
	if timeout <= 0 {
		timeout = 60
	}

	totalTargets := len(targets)
	logInfo("Naabu: scanning %d targets, timeout %ds/target", totalTargets, timeout)

	// 串行扫描每个目标
	for i, target := range targets {
		// 检查父context是否已取消（任务被停止）
		select {
		case <-ctx.Done():
			logInfo("Naabu: cancelled at %d/%d targets", i, totalTargets)
			return allAssets, anyThresholdExceeded
		default:
		}

		// 报告进度 (端口扫描占总进度的0-30%)
		if onProgress != nil {
			progress := (i * 30) / totalTargets
			onProgress(progress, fmt.Sprintf("Port scan: %d/%d", i, totalTargets))
		}

		assets, thresholdExceeded := s.scanSingleTargetWithLogger(ctx, target, portsStr, topPorts, opts, logInfo, logWarn)
		
		if thresholdExceeded {
			// 单个目标超过阈值，记录并跳过该目标，继续扫描其他目标
			anyThresholdExceeded = true
			logWarn("Naabu: %s skipped due to port threshold, continuing with next target", target)
			continue
		}
		
		allAssets = append(allAssets, assets...)
	}

	// 端口扫描完成，进度到30%
	if onProgress != nil {
		onProgress(30, fmt.Sprintf("Port scan completed: %d ports", len(allAssets)))
	}

	logInfo("Naabu: completed, found %d open ports", len(allAssets))
	return allAssets, anyThresholdExceeded
}

// 使用 Naabu 原生的 PortThreshold 参数实现端口阈值检测
// 当某个主机的开放端口数超过阈值时，Naabu 会自动跳过该主机
func (s *NaabuScanner) scanSingleTargetWithLogger(ctx context.Context, target, portsStr, topPorts string, opts *NaabuOptions, logInfo, logWarn logFunc) ([]*Asset, bool) {
	var assets []*Asset
	var mu sync.Mutex
	var foundPorts []string // 收集发现的端口
	thresholdExceeded := false

	// 单个目标超时
	timeout := opts.Timeout
	if timeout <= 0 {
		timeout = 60
	}
	targetCtx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()

	options := runner.Options{
		Host:              goflags.StringSlice([]string{target}),
		Ports:             portsStr,
		TopPorts:          topPorts,
		Rate:              opts.Rate,
		Timeout:           5, // 单个端口连接超时
		ScanType:          opts.ScanType,
		Silent:            true,
		PortThreshold:     opts.PortThreshold,     // 使用 Naabu 原生端口阈值参数
		SkipHostDiscovery: opts.SkipHostDiscovery, // 跳过主机发现 (-Pn)
		OnResult: func(hr *result.HostResult) {
			mu.Lock()
			defer mu.Unlock()

			host := hr.Host
			for _, port := range hr.Ports {
				asset := &Asset{
					Authority: fmt.Sprintf("%s:%d", host, port.Port),
					Host:      host,
					Port:      port.Port,
					Category:  getCategory(host),
				}
				assets = append(assets, asset)
				foundPorts = append(foundPorts, fmt.Sprintf("%d", port.Port))
			}
		},
	}

	naabuRunner, err := runner.NewRunner(&options)
	if err != nil {
		logWarn("Naabu: failed to scan %s: %v", target, err)
		return assets, false
	}
	defer naabuRunner.Close()

	// 运行扫描
	err = naabuRunner.RunEnumeration(targetCtx)

	// 检查扫描结果
	if err != nil {
		errStr := err.Error()
		// 检查是否是端口阈值超过导致的跳过
		if strings.Contains(errStr, "threshold") || strings.Contains(errStr, "skipping") {
			thresholdExceeded = true
			logWarn("Naabu: %s exceeded port threshold (%d), results discarded", target, opts.PortThreshold)
			// 清空结果，不保存到数据库
			mu.Lock()
			assets = nil
			foundPorts = nil
			mu.Unlock()
		} else if targetCtx.Err() == context.DeadlineExceeded {
			// 超时：保留已识别的结果
			logWarn("Naabu: %s timeout after %ds, keeping %d ports found", target, timeout, len(assets))
		} else if ctx.Err() == nil {
			logWarn("Naabu: %s error: %v", target, err)
		}
	}

	// 额外检查：如果端口数超过阈值，清空结果（仅针对阈值，不影响超时）
	// 这是为了处理 Naabu 没有返回错误但实际上超过阈值的情况
	mu.Lock()
	if !thresholdExceeded && opts.PortThreshold > 0 && len(assets) > opts.PortThreshold {
		thresholdExceeded = true
		logWarn("Naabu: %s exceeded port threshold (%d > %d), results discarded", target, len(assets), opts.PortThreshold)
		assets = nil
		foundPorts = nil
	}
	mu.Unlock()

	// 输出扫描结果日志
	if len(foundPorts) > 0 {
		logInfo("Naabu: %s -> %s", target, strings.Join(foundPorts, ","))
	}

	return assets, thresholdExceeded
}
