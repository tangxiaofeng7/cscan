package scanner

import (
	"context"
	"net"
	"strings"
	"sync"
)

// DomainScanner 域名扫描器
type DomainScanner struct {
	BaseScanner
}

// NewDomainScanner 创建域名扫描器
func NewDomainScanner() *DomainScanner {
	return &DomainScanner{
		BaseScanner: BaseScanner{name: "domainscan"},
	}
}

// DomainScanOptions 域名扫描选项
type DomainScanOptions struct {
	Subfinder  bool `json:"subfinder"`
	Massdns    bool `json:"massdns"`
	Concurrent int  `json:"concurrent"`
}

// Scan 执行域名扫描
func (s *DomainScanner) Scan(ctx context.Context, config *ScanConfig) (*ScanResult, error) {
	opts, ok := config.Options.(*DomainScanOptions)
	if !ok {
		opts = &DomainScanOptions{
			Concurrent: 50,
		}
	}

	// 解析目标域名
	domains := parseDomains(config.Target)

	// 子域名枚举
	var subdomains []string
	for _, domain := range domains {
		subs := s.enumerateSubdomains(ctx, domain, opts)
		subdomains = append(subdomains, subs...)
	}

	// DNS解析
	assets := s.resolveDomains(ctx, subdomains, opts.Concurrent)

	return &ScanResult{
		WorkspaceId: config.WorkspaceId,
		MainTaskId:  config.MainTaskId,
		Assets:      assets,
	}, nil
}

// parseDomains 解析域名
func parseDomains(target string) []string {
	var domains []string
	lines := strings.Split(target, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// 只保留域名
		if !net.ParseIP(line).IsUnspecified() {
			continue
		}
		domains = append(domains, line)
	}
	return domains
}

// enumerateSubdomains 子域名枚举
func (s *DomainScanner) enumerateSubdomains(ctx context.Context, domain string, opts *DomainScanOptions) []string {
	var subdomains []string

	// 常见子域名前缀
	prefixes := []string{
		"www", "mail", "ftp", "admin", "api", "dev", "test", "staging",
		"blog", "shop", "store", "app", "m", "mobile", "cdn", "static",
		"img", "images", "assets", "media", "video", "download", "docs",
		"portal", "vpn", "remote", "gateway", "proxy", "ns1", "ns2",
		"mx", "smtp", "pop", "imap", "webmail", "owa", "exchange",
		"git", "gitlab", "github", "jenkins", "ci", "cd", "build",
		"monitor", "grafana", "prometheus", "kibana", "elastic",
		"db", "database", "mysql", "postgres", "redis", "mongo",
		"backup", "bak", "old", "new", "beta", "alpha", "demo",
	}

	for _, prefix := range prefixes {
		subdomain := prefix + "." + domain
		subdomains = append(subdomains, subdomain)
	}

	// 添加主域名
	subdomains = append(subdomains, domain)

	return subdomains
}

// resolveDomains DNS解析
func (s *DomainScanner) resolveDomains(ctx context.Context, domains []string, concurrent int) []*Asset {
	var assets []*Asset
	var mu sync.Mutex
	var wg sync.WaitGroup

	// 创建任务通道
	taskChan := make(chan string, concurrent)

	// 启动工作协程
	for i := 0; i < concurrent; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for domain := range taskChan {
				select {
				case <-ctx.Done():
					return
				default:
					ips, err := net.LookupIP(domain)
					if err != nil || len(ips) == 0 {
						continue
					}

					asset := &Asset{
						Authority: domain,
						Host:      domain,
						Category:  "domain",
					}

					// 分类IPv4和IPv6
					for _, ip := range ips {
						if ip4 := ip.To4(); ip4 != nil {
							asset.IPV4 = append(asset.IPV4, IPInfo{IP: ip4.String()})
						} else {
							asset.IPV6 = append(asset.IPV6, IPInfo{IP: ip.String()})
						}
					}

					// 查询CNAME
					cname, err := net.LookupCNAME(domain)
					if err == nil && cname != domain+"." {
						asset.CName = strings.TrimSuffix(cname, ".")
					}

					mu.Lock()
					assets = append(assets, asset)
					mu.Unlock()
				}
			}
		}()
	}

	// 分发任务
	for _, domain := range domains {
		select {
		case <-ctx.Done():
			close(taskChan)
			wg.Wait()
			return assets
		case taskChan <- domain:
		}
	}

	close(taskChan)
	wg.Wait()

	return assets
}
