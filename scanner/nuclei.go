package scanner

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"cscan/pkg/mapping"

	nuclei "github.com/projectdiscovery/nuclei/v3/lib"
	"github.com/projectdiscovery/nuclei/v3/pkg/output"
	"github.com/zeromicro/go-zero/core/logx"
)

// NucleiScanner Nuclei扫描器 (使用SDK模式)
type NucleiScanner struct {
	BaseScanner
}

// NewNucleiScanner 创建Nuclei扫描器
func NewNucleiScanner() *NucleiScanner {
	return &NucleiScanner{
		BaseScanner: BaseScanner{name: "nuclei"},
	}
}

// NucleiOptions Nuclei扫描选项
type NucleiOptions struct {
	Templates            []string                      `json:"templates"`            // 模板路径
	Tags                 []string                      `json:"tags"`                 // 标签过滤
	Severity             string                        `json:"severity"`             // 严重级别: critical,high,medium,low,info (CSV格式)
	ExcludeTags          []string                      `json:"excludeTags"`          // 排除标签
	ExcludeTemplates     []string                      `json:"excludeTemplates"`     // 排除模板
	RateLimit            int                           `json:"rateLimit"`            // 速率限制
	Concurrency          int                           `json:"concurrency"`          // 并发数
	Timeout              int                           `json:"timeout"`              // 超时时间(秒)
	Retries              int                           `json:"retries"`              // 重试次数
	AutoScan             bool                          `json:"autoScan"`             // 基于自定义标签映射自动扫描
	AutomaticScan        bool                          `json:"automaticScan"`        // 基于Wappalyzer技术的自动扫描（nuclei -as）
	TagMappings          map[string][]string           `json:"tagMappings"`          // 应用名称到Nuclei标签的映射
	CustomTemplates      []string                      `json:"customTemplates"`      // 自定义模板内容(YAML)
	CustomPocOnly        bool                          `json:"customPocOnly"`        // 只使用自定义POC
	NucleiTemplates      []string                      `json:"nucleiTemplates"`      // 从数据库加载的Nuclei模板内容
	OnVulnerabilityFound func(vul *Vulnerability)      `json:"-"`                    // 发现漏洞时的回调函数
}

// Scan 执行Nuclei扫描
func (s *NucleiScanner) Scan(ctx context.Context, config *ScanConfig) (*ScanResult, error) {
	result := &ScanResult{
		WorkspaceId:     config.WorkspaceId,
		MainTaskId:      config.MainTaskId,
		Vulnerabilities: make([]*Vulnerability, 0),
	}

	// 解析选项
	opts := &NucleiOptions{
		Severity:    "critical,high,medium",
		RateLimit:   150,
		Concurrency: 25,
		Timeout:     10,
		Retries:     1,
	}
	if config.Options != nil {
		if o, ok := config.Options.(*NucleiOptions); ok {
			opts = o
		}
	}

	// 自动扫描模式1: 基于自定义标签映射
	if opts.AutoScan && opts.TagMappings != nil {
		autoTags := s.generateAutoTags(config.Assets, opts.TagMappings)
		if len(autoTags) > 0 {
			logx.Debugf("Auto-scan (custom mapping) generated tags: %v", autoTags)
			opts.Tags = append(opts.Tags, autoTags...)
		}
	}

	// 自动扫描模式2: 基于Wappalyzer内置映射（类似nuclei -as）
	if opts.AutomaticScan {
		wappalyzerTags := s.generateWappalyzerAutoTags(config.Assets)
		if len(wappalyzerTags) > 0 {
			logx.Debugf("Auto-scan (Wappalyzer) generated tags: %v", wappalyzerTags)
			opts.Tags = append(opts.Tags, wappalyzerTags...)
		}
	}

	// 去重标签
	if len(opts.Tags) > 0 {
		opts.Tags = uniqueStrings(opts.Tags)
	}

	// 准备目标列表
	var targets []string
	if len(config.Targets) > 0 {
		// 直接使用配置中的目标URL（用于POC验证等场景）
		targets = config.Targets
	} else {
		// 从资产列表构建目标URL
		targets = s.prepareTargets(config.Assets)
	}
	if len(targets) == 0 {
		logx.Info("No targets for nuclei scan")
		return result, nil
	}

	logx.Debugf("Targets (%d): %v", len(targets), targets)

	// 处理自定义POC - 写入临时文件
	var customTemplatePaths []string
	var tempDir string
	if len(opts.CustomTemplates) > 0 {
		var err error
		tempDir, err = os.MkdirTemp("", "nuclei-custom-*")
		if err != nil {
			logx.Errorf("Failed to create temp dir for custom templates: %v", err)
		} else {
			for i, content := range opts.CustomTemplates {
				templatePath := filepath.Join(tempDir, fmt.Sprintf("custom-%d.yaml", i))
				if err := os.WriteFile(templatePath, []byte(content), 0644); err != nil {
					logx.Errorf("Failed to write custom template %d: %v", i, err)
					continue
				}
				customTemplatePaths = append(customTemplatePaths, templatePath)
			}
		}
		// 清理临时目录
		defer func() {
			if tempDir != "" {
				os.RemoveAll(tempDir)
			}
		}()
	}

	// 构建Nuclei SDK选项
	nucleiOpts := s.buildNucleiOptions(opts, customTemplatePaths)

	// 创建Nuclei引擎
	ne, err := nuclei.NewNucleiEngineCtx(ctx, nucleiOpts...)
	if err != nil {
		logx.Errorf("Failed to create nuclei engine: %v", err)
		return result, fmt.Errorf("create nuclei engine failed: %v", err)
	}
	defer ne.Close()

	// 加载模板
	if err := ne.LoadAllTemplates(); err != nil {
		logx.Errorf("Failed to load templates: %v", err)
	}

	// 获取加载的模板数量
	templates := ne.GetTemplates()
	logx.Debugf("Loaded %d templates", len(templates))

	// 加载目标
	ne.LoadTargets(targets, false)

	// 收集结果（使用map去重）
	var vuls []*Vulnerability
	seen := make(map[string]bool)

	// 执行扫描并通过回调收集结果
	err = ne.ExecuteCallbackWithCtx(ctx, func(event *output.ResultEvent) {
		logx.Debugf("Nuclei result: TemplateID=%s, Host=%s, Matched=%s",
			event.TemplateID, event.Host, event.Matched)

		vul := s.convertResult(event)
		if vul != nil {
			// 基于 host+port+templateId+url 去重
			key := fmt.Sprintf("%s:%d:%s:%s", vul.Host, vul.Port, vul.PocFile, vul.Url)
			if !seen[key] {
				seen[key] = true
				vuls = append(vuls, vul)
				// 如果有回调函数，实时通知
				if opts.OnVulnerabilityFound != nil {
					opts.OnVulnerabilityFound(vul)
				}
			}
		}
	})

	if err != nil {
		logx.Errorf("Nuclei scan execution error: %v", err)
	}

	result.Vulnerabilities = vuls

	return result, nil
}

// prepareTargets 准备目标URL列表（跳过非HTTP资产）
func (s *NucleiScanner) prepareTargets(assets []*Asset) []string {
	targets := make([]string, 0, len(assets))
	seen := make(map[string]bool)
	skipped := 0

	for _, asset := range assets {
		// 使用 IsHTTP 字段判断（端口扫描阶段已设置）
		if !asset.IsHTTP {
			skipped++
			logx.Debugf("Skipping non-HTTP asset: %s:%d (service: %s, isHttp: %v)", asset.Host, asset.Port, asset.Service, asset.IsHTTP)
			continue
		}

		scheme := "http"
		if asset.Service == "https" || asset.Port == 443 || asset.Port == 8443 {
			scheme = "https"
		}

		target := fmt.Sprintf("%s://%s:%d", scheme, asset.Host, asset.Port)

		if !seen[target] {
			seen[target] = true
			targets = append(targets, target)
		}
	}

	if skipped > 0 {
		logx.Infof("Nuclei: skipped %d non-HTTP assets, scanning %d HTTP targets", skipped, len(targets))
	}

	return targets
}

// buildNucleiOptions 构建Nuclei SDK选项
// 所有模板都应该从数据库获取，不使用本地模板目录
func (s *NucleiScanner) buildNucleiOptions(opts *NucleiOptions, customTemplatePaths []string) []nuclei.NucleiSDKOptions {
	var nucleiOpts []nuclei.NucleiSDKOptions

	// 判断是否有模板（从数据库获取的模板）
	hasTemplates := len(customTemplatePaths) > 0

	if hasTemplates {
		// 使用从数据库获取的模板
		nucleiOpts = append(nucleiOpts, nuclei.WithTemplatesOrWorkflows(nuclei.TemplateSources{
			Templates: customTemplatePaths,
		}))
		logx.Infof("Using %d templates from database", len(customTemplatePaths))
	} else {
		// 没有模板，记录警告
		logx.Errorf("No templates provided! POC scan requires templates from database.")
	}

	// 模板过滤器 - 当使用数据库模板时，模板已经是筛选过的，跳过tag过滤
	filters := nuclei.TemplateFilters{}
	hasFilters := false

	// severity过滤仍然生效
	if opts.Severity != "" {
		filters.Severity = opts.Severity
		hasFilters = true
	}

	if hasFilters {
		nucleiOpts = append(nucleiOpts, nuclei.WithTemplateFilters(filters))
	}

	// 并发配置
	if opts.Concurrency > 0 {
		nucleiOpts = append(nucleiOpts, nuclei.WithConcurrency(nuclei.Concurrency{
			TemplateConcurrency:           opts.Concurrency,
			HostConcurrency:               opts.Concurrency,
			HeadlessHostConcurrency:       5,
			HeadlessTemplateConcurrency:   5,
			JavascriptTemplateConcurrency: 5,
			TemplatePayloadConcurrency:    10,
			ProbeConcurrency:              50,
		}))
	}

	// 速率限制
	if opts.RateLimit > 0 {
		nucleiOpts = append(nucleiOpts, nuclei.WithGlobalRateLimit(opts.RateLimit, 1))
	}

	// 禁用更新检查
	nucleiOpts = append(nucleiOpts, nuclei.DisableUpdateCheck())

	return nucleiOpts
}

// convertResult 转换Nuclei结果为漏洞对象
func (s *NucleiScanner) convertResult(event *output.ResultEvent) *Vulnerability {
	if event == nil {
		return nil
	}

	host, port := s.parseHostPort(event.Host)

	resultDesc := event.Info.Name
	if event.Info.Description != "" {
		resultDesc += "\n" + event.Info.Description
	}
	if len(event.ExtractedResults) > 0 {
		resultDesc += "\nExtracted: " + strings.Join(event.ExtractedResults, ", ")
	}

	return &Vulnerability{
		Authority: fmt.Sprintf("%s:%d", host, port),
		Host:      host,
		Port:      port,
		Url:       event.Matched,
		PocFile:   event.TemplateID,
		Source:    "nuclei",
		Severity:  event.Info.SeverityHolder.Severity.String(),
		Result:    resultDesc,
	}
}

// parseHostPort 从URL解析主机和端口
func (s *NucleiScanner) parseHostPort(rawURL string) (string, int) {
	if !strings.Contains(rawURL, "://") {
		rawURL = "http://" + rawURL
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		return s.parseHostPortSimple(rawURL)
	}

	host := u.Hostname()
	port := 80

	if u.Port() != "" {
		if p, err := strconv.Atoi(u.Port()); err == nil {
			port = p
		}
	} else if u.Scheme == "https" {
		port = 443
	}

	return host, port
}

// parseHostPortSimple 简单解析主机和端口
func (s *NucleiScanner) parseHostPortSimple(hostPort string) (string, int) {
	hostPort = strings.TrimPrefix(hostPort, "http://")
	hostPort = strings.TrimPrefix(hostPort, "https://")

	if idx := strings.Index(hostPort, "/"); idx != -1 {
		hostPort = hostPort[:idx]
	}

	if idx := strings.LastIndex(hostPort, ":"); idx != -1 {
		host := hostPort[:idx]
		port := 80
		if p, err := strconv.Atoi(hostPort[idx+1:]); err == nil {
			port = p
		}
		return host, port
	}

	return hostPort, 80
}

// generateAutoTags 根据资产的应用信息生成Nuclei标签（基于自定义标签映射）
func (s *NucleiScanner) generateAutoTags(assets []*Asset, tagMappings map[string][]string) []string {
	tagSet := make(map[string]bool)

	for _, asset := range assets {
		logx.Debugf("Asset %s:%d apps: %v", asset.Host, asset.Port, asset.App)
		for _, app := range asset.App {
			appName := parseAppName(app)
			appNameLower := strings.ToLower(appName)

			logx.Debugf("Parsed app name: '%s' -> '%s'", app, appName)

			for mappedApp, tags := range tagMappings {
				if strings.ToLower(mappedApp) == appNameLower {
					logx.Infof("Matched app '%s' -> tags: %v", appName, tags)
					for _, tag := range tags {
						tagSet[tag] = true
					}
					break
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

// generateWappalyzerAutoTags 根据资产的应用信息生成Nuclei标签（基于Wappalyzer内置映射，类似nuclei -as）
func (s *NucleiScanner) generateWappalyzerAutoTags(assets []*Asset) []string {
	tagSet := make(map[string]bool)

	for _, asset := range assets {
		logx.Debugf("Asset %s:%d apps: %v", asset.Host, asset.Port, asset.App)
		for _, app := range asset.App {
			appName := parseAppName(app)
			appNameLower := strings.ToLower(appName)

			// 使用内置的Wappalyzer到Nuclei标签映射
			if tags, ok := mapping.WappalyzerNucleiMapping[appNameLower]; ok {
				logx.Infof("Wappalyzer auto-scan matched '%s' -> tags: %v", appName, tags)
				for _, tag := range tags {
					tagSet[tag] = true
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

// parseAppName 解析应用名称，去除版本号和来源标识
func parseAppName(app string) string {
	appName := app
	// 先去掉 [source] 后缀
	if idx := strings.Index(appName, "["); idx > 0 {
		appName = appName[:idx]
	}
	// 再去掉 :version 后缀
	if idx := strings.Index(appName, ":"); idx > 0 {
		appName = appName[:idx]
	}
	return strings.TrimSpace(appName)
}
