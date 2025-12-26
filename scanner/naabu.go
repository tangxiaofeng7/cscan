package scanner

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/projectdiscovery/goflags"
	"github.com/projectdiscovery/naabu/v2/pkg/result"
	"github.com/projectdiscovery/naabu/v2/pkg/runner"
	"github.com/zeromicro/go-zero/core/logx"
)

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
	Ports         string `json:"ports"`
	Rate          int    `json:"rate"`
	Timeout       int    `json:"timeout"`
	ScanType      string `json:"scanType"`      // s=SYN, c=CONNECT
	PortThreshold int    `json:"portThreshold"` // 端口阈值，实时检测
}

// Scan 执行Naabu扫描
func (s *NaabuScanner) Scan(ctx context.Context, config *ScanConfig) (*ScanResult, error) {
	// 默认配置
	opts := &NaabuOptions{
		Ports:         "80,443,8080",
		Rate:          1000,
		Timeout:       60, // 单个目标扫描超时，默认60秒
		ScanType:      "s", // SYN扫描
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
					Ports         string `json:"ports"`
					Rate          int    `json:"rate"`
					Timeout       int    `json:"timeout"`
					PortThreshold int    `json:"portThreshold"`
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
	assets := s.runNaabuWithLogger(ctx, targets, opts, logInfo, logWarn)

	return &ScanResult{
		WorkspaceId: config.WorkspaceId,
		MainTaskId:  config.MainTaskId,
		Assets:      assets,
	}, nil
}

// logFunc 日志函数类型
type logFunc func(format string, args ...interface{})

// runNaabuWithLogger 运行Naabu扫描（带日志回调）
// 按单个目标拆分，串行执行，每个目标独立超时控制
func (s *NaabuScanner) runNaabuWithLogger(ctx context.Context, targets []string, opts *NaabuOptions, logInfo, logWarn logFunc) []*Asset {
	var allAssets []*Asset

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

	logInfo("Naabu: scanning %d targets, timeout %ds/target", len(targets), timeout)

	// 串行扫描每个目标
	for i, target := range targets {
		// 检查父context是否已取消
		select {
		case <-ctx.Done():
			logInfo("Naabu: cancelled at %d/%d targets", i, len(targets))
			return allAssets
		default:
		}

		assets := s.scanSingleTargetWithLogger(ctx, target, portsStr, topPorts, opts, logInfo, logWarn)
		allAssets = append(allAssets, assets...)
	}

	logInfo("Naabu: completed, found %d open ports", len(allAssets))
	return allAssets
}

// scanSingleTargetWithLogger 扫描单个目标（带日志回调）
func (s *NaabuScanner) scanSingleTargetWithLogger(ctx context.Context, target, portsStr, topPorts string, opts *NaabuOptions, logInfo, logWarn logFunc) []*Asset {
	var assets []*Asset
	var mu sync.Mutex
	var foundPorts []string // 收集发现的端口

	// 单个目标超时
	timeout := opts.Timeout
	if timeout <= 0 {
		timeout = 60
	}
	targetCtx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()

	// 端口计数，用于阈值检测
	portCount := 0
	skipped := false

	options := runner.Options{
		Host:     goflags.StringSlice([]string{target}),
		Ports:    portsStr,
		TopPorts: topPorts,
		Rate:     opts.Rate,
		Timeout:  5, // 单个端口连接超时
		ScanType: opts.ScanType,
		Silent:   true,
		OnResult: func(hr *result.HostResult) {
			mu.Lock()
			defer mu.Unlock()

			if skipped {
				return
			}

			host := hr.Host
			for _, port := range hr.Ports {
				portCount++
				if opts.PortThreshold > 0 && portCount > opts.PortThreshold {
					if !skipped {
						skipped = true
						logWarn("Naabu: %s exceeded port threshold (%d), skipping", host, opts.PortThreshold)
						assets = nil
						foundPorts = nil
					}
					return
				}

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
		return assets
	}
	defer naabuRunner.Close()

	if err := naabuRunner.RunEnumeration(targetCtx); err != nil {
		if targetCtx.Err() == context.DeadlineExceeded {
			logWarn("Naabu: %s timeout after %ds", target, timeout)
		} else if ctx.Err() == nil {
			logWarn("Naabu: %s error: %v", target, err)
		}
	}

	if len(foundPorts) > 0 {
		logInfo("Naabu: %s -> %s", target, strings.Join(foundPorts, ","))
	}

	return assets
}
