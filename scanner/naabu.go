package scanner

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

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
		Timeout:       5,
		ScanType:      "s", // SYN扫描
		PortThreshold: 0,   // 默认不限制
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
						logx.Infof("Naabu: parsed ports from config: %s", portConfig.Ports)
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

	logx.Infof("Naabu Scan config - Ports: %s, Rate: %d, Timeout: %d, PortThreshold: %d", opts.Ports, opts.Rate, opts.Timeout, opts.PortThreshold)

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
	assets := s.runNaabu(ctx, targets, opts)

	return &ScanResult{
		WorkspaceId: config.WorkspaceId,
		MainTaskId:  config.MainTaskId,
		Assets:      assets,
	}, nil
}

// runNaabu 运行Naabu扫描
func (s *NaabuScanner) runNaabu(ctx context.Context, targets []string, opts *NaabuOptions) []*Asset {
	var assets []*Asset
	var mu sync.Mutex

	// 实时端口阈值检测：记录每个主机的开放端口数量和是否已超过阈值
	hostPortCount := make(map[string]int)
	skippedHosts := make(map[string]bool)

	// 处理端口配置
	var portsStr string
	var topPorts string

	// 检查是否使用预定义端口集（使用Naabu内置的TopPorts功能）
	switch opts.Ports {
	case "top100":
		topPorts = "100"
		logx.Infof("Running Naabu scan on %d targets with top 100 ports, rate: %d", len(targets), opts.Rate)
	case "top1000":
		topPorts = "1000"
		logx.Infof("Running Naabu scan on %d targets with top 1000 ports, rate: %d", len(targets), opts.Rate)
	default:
		// 其他情况使用自定义端口列表
		ports := parsePorts(opts.Ports)
		portsStr = portsToString(ports)
		logx.Infof("Running Naabu scan on %d targets with %d ports, rate: %d", len(targets), len(ports), opts.Rate)
	}

	// 构建Naabu选项
	options := runner.Options{
		Host:     goflags.StringSlice(targets),
		Ports:    portsStr,
		TopPorts: topPorts,
		Rate:     opts.Rate,
		ScanType: opts.ScanType,
		Silent:   true,
		OnResult: func(hr *result.HostResult) {
			mu.Lock()
			defer mu.Unlock()

			host := hr.Host

			// 如果该主机已被标记为跳过，直接忽略后续结果
			if skippedHosts[host] {
				return
			}

			for _, port := range hr.Ports {
				// 实时检测端口阈值
				hostPortCount[host]++
				if opts.PortThreshold > 0 && hostPortCount[host] > opts.PortThreshold {
					// 第一次超过阈值时记录日志
					if !skippedHosts[host] {
						skippedHosts[host] = true
						logx.Infof("Host %s exceeded port threshold (%d > %d), skipping as potential honeypot/firewall",
							host, hostPortCount[host], opts.PortThreshold)
						// 移除该主机已收集的所有端口
						newAssets := make([]*Asset, 0, len(assets))
						for _, a := range assets {
							if a.Host != host {
								newAssets = append(newAssets, a)
							}
						}
						assets = newAssets
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
				logx.Debugf("Naabu found: %s:%d", host, port.Port)
			}
		},
	}

	// 创建Naabu runner
	naabuRunner, err := runner.NewRunner(&options)
	if err != nil {
		logx.Errorf("Failed to create Naabu runner: %v", err)
		return assets
	}
	defer naabuRunner.Close()

	// 执行扫描
	if err := naabuRunner.RunEnumeration(ctx); err != nil {
		logx.Errorf("Naabu scan error: %v", err)
	}

	// 输出跳过的主机统计
	if len(skippedHosts) > 0 {
		logx.Infof("Naabu scan: skipped %d hosts due to port threshold", len(skippedHosts))
	}

	logx.Infof("Naabu scan completed, found %d open ports", len(assets))
	return assets
}
