package scanner

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"
)

// MasscanScanner Masscan扫描器
type MasscanScanner struct {
	BaseScanner
}

// NewMasscanScanner 创建Masscan扫描器
func NewMasscanScanner() *MasscanScanner {
	return &MasscanScanner{
		BaseScanner: BaseScanner{name: "masscan"},
	}
}

// MasscanOptions Masscan扫描选项
type MasscanOptions struct {
	Ports         string `json:"ports"`
	Rate          int    `json:"rate"`
	Timeout       int    `json:"timeout"`
	PortThreshold int    `json:"portThreshold"` // 端口阈值，实时检测
}

// MasscanResult Masscan输出结果
type MasscanResult struct {
	IP    string `json:"ip"`
	Ports []struct {
		Port   int    `json:"port"`
		Proto  string `json:"proto"`
		Status string `json:"status"`
	} `json:"ports"`
}

// Scan 执行Masscan扫描
func (s *MasscanScanner) Scan(ctx context.Context, config *ScanConfig) (*ScanResult, error) {
	// 默认配置
	opts := &MasscanOptions{
		Ports:         "21,22,23,25,80,443,3306,3389,6379,8080",
		Rate:          1000,
		Timeout:       3,
		PortThreshold: 0, // 默认不限制
	}

	// 尝试从不同类型的Options中提取配置
	if config.Options != nil {
		switch v := config.Options.(type) {
		case *MasscanOptions:
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
			// 尝试通过JSON转换
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

	logx.Infof("Masscan Scan config - Ports: %s, Rate: %d, Timeout: %d, PortThreshold: %d", opts.Ports, opts.Rate, opts.Timeout, opts.PortThreshold)

	logx.Infof("Masscan Scan config - Ports: %s, Rate: %d, Timeout: %d, PortThreshold: %d", opts.Ports, opts.Rate, opts.Timeout, opts.PortThreshold)

	// 检查masscan是否安装
	if !checkMasscanInstalled() {
		logx.Error("masscan not installed, falling back to tcp scan")
		// 回退到TCP扫描
		tcpScanner := NewPortScanner()
		return tcpScanner.Scan(ctx, config)
	}

	// 解析目标
	targets := parseTargets(config.Target)
	if len(config.Targets) > 0 {
		targets = append(targets, config.Targets...)
	}

	// 执行masscan扫描（传入阈值参数）
	assets := s.runMasscan(ctx, targets, opts)

	return &ScanResult{
		WorkspaceId: config.WorkspaceId,
		MainTaskId:  config.MainTaskId,
		Assets:      assets,
	}, nil
}

// runMasscan 运行masscan
func (s *MasscanScanner) runMasscan(ctx context.Context, targets []string, opts *MasscanOptions) []*Asset {
	var assets []*Asset

	// 实时端口阈值检测：记录每个主机的开放端口数量和是否已超过阈值
	hostPortCount := make(map[string]int)
	skippedHosts := make(map[string]bool)

	// 查找域名目标（masscan会将域名解析为IP）
	var domainTarget string
	for _, target := range targets {
		if getCategory(target) == "domain" {
			domainTarget = target
			break
		}
	}

	// 使用 parsePorts 解析端口，统一处理 top100/top1000 和其他格式
	ports := parsePorts(opts.Ports)
	portsStr := portsToString(ports)

	// 构建masscan命令
	// masscan -p ports targets --rate=rate -oJ -
	args := []string{
		"-p", portsStr,
		"--rate", strconv.Itoa(opts.Rate),
		"-oJ", "-", // JSON输出到stdout
	}
	args = append(args, targets...)

	// 输出执行命令到日志
	logx.Infof("Executing command: masscan %s", strings.Join(args, " "))

	cmd := exec.CommandContext(ctx, "masscan", args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		logx.Errorf("masscan stdout pipe error: %v", err)
		return assets
	}

	if err := cmd.Start(); err != nil {
		logx.Errorf("masscan start error: %v", err)
		return assets
	}

	// 解析JSON输出
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || line == "[" || line == "]" {
			continue
		}
		// 去除行尾逗号
		line = strings.TrimSuffix(line, ",")

		var result MasscanResult
		if err := json.Unmarshal([]byte(line), &result); err != nil {
			continue
		}

		// 确定主机标识（用于阈值检测）
		hostKey := result.IP
		if domainTarget != "" {
			hostKey = domainTarget
		}

		// 如果该主机已被标记为跳过，直接忽略后续结果
		if skippedHosts[hostKey] {
			continue
		}

		for _, port := range result.Ports {
			if port.Status == "open" {
				// 实时检测端口阈值
				hostPortCount[hostKey]++
				if opts.PortThreshold > 0 && hostPortCount[hostKey] > opts.PortThreshold {
					// 第一次超过阈值时记录日志
					if !skippedHosts[hostKey] {
						skippedHosts[hostKey] = true
						logx.Infof("Host %s exceeded port threshold (%d > %d), skipping as potential honeypot/firewall",
							hostKey, hostPortCount[hostKey], opts.PortThreshold)
						// 移除该主机已收集的所有端口
						newAssets := make([]*Asset, 0, len(assets))
						for _, a := range assets {
							if a.Host != hostKey {
								newAssets = append(newAssets, a)
							}
						}
						assets = newAssets
					}
					continue
				}

				// 如果原始目标是域名，使用域名作为Authority和Host
				host := result.IP
				authority := fmt.Sprintf("%s:%d", result.IP, port.Port)
				category := getCategory(result.IP)

				if domainTarget != "" {
					host = domainTarget
					authority = fmt.Sprintf("%s:%d", domainTarget, port.Port)
					category = "domain"
				}

				asset := &Asset{
					Authority: authority,
					Host:      host,
					Port:      port.Port,
					Category:  category,
				}
				assets = append(assets, asset)
			}
		}
	}

	cmd.Wait()

	// 输出跳过的主机统计
	if len(skippedHosts) > 0 {
		logx.Infof("Masscan scan: skipped %d hosts due to port threshold", len(skippedHosts))
	}

	return assets
}

// checkMasscanInstalled 检查masscan是否安装
func checkMasscanInstalled() bool {
	cmd := exec.Command("masscan", "--version")
	output, _ := cmd.CombinedOutput()
	// 通过检查输出内容来判断是否安装
	return strings.Contains(string(output), "Masscan version")
}
