package scanner

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"
)

// PortScanner 端口扫描器
type PortScanner struct {
	BaseScanner
}

// NewPortScanner 创建端口扫描器
func NewPortScanner() *PortScanner {
	return &PortScanner{
		BaseScanner: BaseScanner{name: "portscan"},
	}
}

// PortScanOptions 端口扫描选项
type PortScanOptions struct {
	Tool          string `json:"tool"`          // tcp, masscan, nmap
	Ports         string `json:"ports"`
	Rate          int    `json:"rate"`
	Timeout       int    `json:"timeout"`
	Concurrent    int    `json:"concurrent"`
	PortThreshold int    `json:"portThreshold"` // 开放端口数量阈值，超过则过滤该主机
}

// Scan 执行端口扫描
func (s *PortScanner) Scan(ctx context.Context, config *ScanConfig) (*ScanResult, error) {
	opts, ok := config.Options.(*PortScanOptions)
	if !ok {
		opts = &PortScanOptions{
			Ports:      "21,22,23,25,80,443,3306,3389,6379,8080",
			Timeout:    3,
			Concurrent: 100,
		}
	}

	// 解析目标
	targets := parseTargets(config.Target)
	if len(config.Targets) > 0 {
		targets = append(targets, config.Targets...)
	}

	// 解析端口
	ports := parsePorts(opts.Ports)

	// 执行扫描
	assets := s.scanPorts(ctx, targets, ports, opts)

	return &ScanResult{
		WorkspaceId: config.WorkspaceId,
		MainTaskId:  config.MainTaskId,
		Assets:      assets,
	}, nil
}

// scanPorts 扫描端口
func (s *PortScanner) scanPorts(ctx context.Context, targets []string, ports []int, opts *PortScanOptions) []*Asset {
	var assets []*Asset
	var mu sync.Mutex
	var wg sync.WaitGroup

	// 创建任务通道
	taskChan := make(chan struct {
		target string
		port   int
	}, opts.Concurrent)

	// 启动工作协程
	for i := 0; i < opts.Concurrent; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for task := range taskChan {
				select {
				case <-ctx.Done():
					return
				default:
					if isPortOpen(task.target, task.port, opts.Timeout) {
						asset := &Asset{
							Authority: fmt.Sprintf("%s:%d", task.target, task.port),
							Host:      task.target,
							Port:      task.port,
							Category:  getCategory(task.target),
						}
						mu.Lock()
						assets = append(assets, asset)
						mu.Unlock()
					}
				}
			}
		}()
	}

	// 分发任务
	for _, target := range targets {
		for _, port := range ports {
			select {
			case <-ctx.Done():
				close(taskChan)
				wg.Wait()
				return assets
			case taskChan <- struct {
				target string
				port   int
			}{target, port}:
			}
		}
	}

	close(taskChan)
	wg.Wait()

	return assets
}

// isPortOpen 检查端口是否开放
func isPortOpen(host string, port int, timeout int) bool {
	address := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.DialTimeout("tcp", address, time.Duration(timeout)*time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}
