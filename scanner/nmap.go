package scanner

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"
)

// NmapScanner Nmap扫描器
type NmapScanner struct {
	BaseScanner
}

// NewNmapScanner 创建Nmap扫描器
func NewNmapScanner() *NmapScanner {
	return &NmapScanner{
		BaseScanner: BaseScanner{name: "nmap"},
	}
}

// NmapOptions Nmap扫描选项
type NmapOptions struct {
	Ports   string `json:"ports"`
	Rate    int    `json:"rate"`
	Timeout int    `json:"timeout"`
	Args    string `json:"args"` // 额外参数
}

// NmapRun Nmap XML输出结构
type NmapRun struct {
	XMLName xml.Name   `xml:"nmaprun"`
	Hosts   []NmapHost `xml:"host"`
}

type NmapHost struct {
	Addresses []NmapAddress `xml:"address"`
	Ports     NmapPorts     `xml:"ports"`
}

type NmapAddress struct {
	Addr     string `xml:"addr,attr"`
	AddrType string `xml:"addrtype,attr"`
}

// GetIPv4Address 获取IPv4地址（忽略MAC地址）
func (h *NmapHost) GetIPv4Address() string {
	for _, addr := range h.Addresses {
		if addr.AddrType == "ipv4" {
			return addr.Addr
		}
	}
	// 如果没有ipv4，尝试ipv6
	for _, addr := range h.Addresses {
		if addr.AddrType == "ipv6" {
			return addr.Addr
		}
	}
	// 最后返回第一个非mac地址
	for _, addr := range h.Addresses {
		if addr.AddrType != "mac" {
			return addr.Addr
		}
	}
	return ""
}

type NmapPorts struct {
	Ports []NmapPort `xml:"port"`
}

type NmapPort struct {
	Protocol string      `xml:"protocol,attr"`
	PortID   int         `xml:"portid,attr"`
	State    NmapState   `xml:"state"`
	Service  NmapService `xml:"service"`
}

type NmapState struct {
	State string `xml:"state,attr"`
}

type NmapService struct {
	Name    string `xml:"name,attr"`
	Product string `xml:"product,attr"`
	Version string `xml:"version,attr"`
}

// Scan 执行Nmap扫描
func (s *NmapScanner) Scan(ctx context.Context, config *ScanConfig) (*ScanResult, error) {
	// 默认配置
	opts := &NmapOptions{
		Ports:   "21,22,23,25,80,443,3306,3389,6379,8080",
		Timeout: 3,
	}

	// 尝试从不同类型的Options中提取配置
	if config.Options != nil {
		switch v := config.Options.(type) {
		case *NmapOptions:
			opts = v
		case *PortScanOptions:
			if v.Ports != "" {
				opts.Ports = v.Ports
			}
			if v.Timeout > 0 {
				opts.Timeout = v.Timeout
			}
		default:
			// 尝试通过JSON转换
			if data, err := json.Marshal(config.Options); err == nil {
				json.Unmarshal(data, opts)
			}
		}
	}

	// 检查nmap是否安装
	if !checkNmapInstalled() {
		logx.Error("nmap not installed, falling back to tcp scan")
		// 回退到TCP扫描
		tcpScanner := NewPortScanner()
		return tcpScanner.Scan(ctx, config)
	}

	// 解析目标
	targets := parseTargets(config.Target)
	if len(config.Targets) > 0 {
		targets = append(targets, config.Targets...)
	}

	// 执行nmap扫描
	assets := s.runNmap(ctx, targets, opts)

	return &ScanResult{
		WorkspaceId: config.WorkspaceId,
		MainTaskId:  config.MainTaskId,
		Assets:      assets,
	}, nil
}

// runNmap 运行nmap
func (s *NmapScanner) runNmap(ctx context.Context, targets []string, opts *NmapOptions) []*Asset {
	var assets []*Asset

	// 构建IP到原始目标的映射（用于域名解析后还原）
	ipToTarget := make(map[string]string)
	for _, target := range targets {
		// 如果是域名，记录映射关系
		if getCategory(target) == "domain" {
			ipToTarget[target] = target
		}
	}

	// 解析端口字符串，支持 top100/top1000
	portsStr := expandPortsString(opts.Ports)

	// 构建nmap命令
	// nmap -sS -Pn -p ports targets -oX -
	args := []string{
		"-sS",                                       // SYN扫描
		"-Pn",                                       // 跳过主机发现
		"-p", portsStr,                              // 端口
		"--host-timeout", strconv.Itoa(opts.Timeout) + "s",
		"-oX", "-", // XML输出到stdout
	}

	// 添加额外参数
	if opts.Args != "" {
		extraArgs := strings.Fields(opts.Args)
		args = append(args, extraArgs...)
	}

	args = append(args, targets...)

	// 输出执行命令到日志
	logx.Infof("Executing command: nmap %s", strings.Join(args, " "))

	cmd := exec.CommandContext(ctx, "nmap", args...)
	output, err := cmd.Output()
	if err != nil {
		logx.Errorf("nmap error: %v", err)
		return assets
	}

	// 解析XML输出
	var nmapRun NmapRun
	if err := xml.Unmarshal(output, &nmapRun); err != nil {
		logx.Errorf("nmap xml parse error: %v", err)
		return assets
	}

	for _, host := range nmapRun.Hosts {
		// 获取IPv4地址（忽略MAC地址）
		ip := host.GetIPv4Address()
		if ip == "" {
			logx.Infof("No valid IP address found for host, skipping")
			continue
		}
		
		// 查找原始目标（可能是域名）
		originalTarget := ip
		for _, target := range targets {
			// 如果目标是域名，使用域名作为Authority
			if getCategory(target) == "domain" {
				// nmap会将域名解析为IP，这里尝试匹配
				originalTarget = target
				break
			}
		}
		
		for _, port := range host.Ports.Ports {
			if port.State.State == "open" {
				// 如果原始目标是域名，Authority使用域名；否则使用IP
				authority := fmt.Sprintf("%s:%d", originalTarget, port.PortID)
				hostStr := originalTarget
				category := getCategory(originalTarget)
				
				asset := &Asset{
					Authority: authority,
					Host:      hostStr,
					Port:      port.PortID,
					Category:  category,
					Service:   port.Service.Name,
				}
				// 如果有产品信息，添加到App
				if port.Service.Product != "" {
					asset.App = []string{port.Service.Product}
				}
				assets = append(assets, asset)
			}
		}
	}

	return assets
}

// checkNmapInstalled 检查nmap是否安装
func checkNmapInstalled() bool {
	cmd := exec.Command("nmap", "--version")
	err := cmd.Run()
	return err == nil
}
