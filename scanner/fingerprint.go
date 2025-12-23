package scanner

import (
	"context"
	"crypto/md5"
	"crypto/tls"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
	wappalyzer "github.com/projectdiscovery/wappalyzergo"
	"github.com/zeromicro/go-zero/core/logx"
)

// FingerprintScanner 指纹扫描器
type FingerprintScanner struct {
	BaseScanner
	client               *http.Client
	wappalyzerClient     *wappalyzer.Wappalyze
	customFingerprintEngine *CustomFingerprintEngine
}

// AppDetectionResult 应用检测结果，用于合并多个来源的识别结果
type AppDetectionResult struct {
	Name         string   // 应用名称
	OriginalName string   // 原始名称（可能包含版本号）
	Sources      []string // 检测来源：httpx, wappalyzer, custom
	CustomIDs    []string // 自定义指纹的ID列表
}

// NewFingerprintScanner 创建指纹扫描器
func NewFingerprintScanner() *FingerprintScanner {
	wappalyzerClient, _ := wappalyzer.New()
	return &FingerprintScanner{
		BaseScanner: BaseScanner{name: "fingerprint"},
		client: &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= 3 {
					return http.ErrUseLastResponse
				}
				return nil
			},
		},
		wappalyzerClient: wappalyzerClient,
	}
}

// SetCustomFingerprintEngine 设置自定义指纹引擎
func (s *FingerprintScanner) SetCustomFingerprintEngine(engine *CustomFingerprintEngine) {
	s.customFingerprintEngine = engine
}

// FingerprintOptions 指纹识别选项
type FingerprintOptions struct {
	Enable       bool `json:"enable"`
	Httpx        bool `json:"httpx"`
	Screenshot   bool `json:"screenshot"`
	IconHash     bool `json:"iconHash"`
	Wappalyzer   bool `json:"wappalyzer"`
	CustomEngine bool `json:"customEngine"` // 使用自定义指纹引擎
	Timeout      int  `json:"timeout"`
	Concurrency  int  `json:"concurrency"`
}


// Scan 执行指纹识别
func (s *FingerprintScanner) Scan(ctx context.Context, config *ScanConfig) (*ScanResult, error) {
	// 解析配置
	opts := &FingerprintOptions{
		Enable:       true,
		Httpx:        true,
		IconHash:     true,
		Wappalyzer:   true,
		CustomEngine: true, // 默认启用自定义指纹引擎
		Screenshot:   false,
		Timeout:      10,
	}
	if config.Options != nil {
		switch v := config.Options.(type) {
		case *FingerprintOptions:
			opts = v
		default:
			if data, err := json.Marshal(config.Options); err == nil {
				json.Unmarshal(data, opts)
			}
		}
	}

	result := &ScanResult{
		WorkspaceId: config.WorkspaceId,
		MainTaskId:  config.MainTaskId,
		Assets:      make([]*Asset, 0),
	}

	// 过滤出HTTP/HTTPS相关的资产
	httpAssets := filterHttpAssets(config.Assets)
	if len(httpAssets) == 0 {
		logx.Info("No HTTP/HTTPS assets found, skipping fingerprint detection")
		// 返回所有原始资产，但不进行指纹识别
		result.Assets = config.Assets
		return result, nil
	}
	
	logx.Infof("Filtered %d HTTP/HTTPS assets from %d total assets for fingerprint detection", len(httpAssets), len(config.Assets))

	// 检查httpx是否可用
	httpxInstalled := checkHttpxInstalled()
	useHttpx := opts.Httpx && httpxInstalled

	if useHttpx {
		// httpx支持截图功能，如果同时启用httpx和screenshot，使用httpx的-screenshot参数
		if opts.Screenshot {
			logx.Info("Using httpx for fingerprint detection with screenshot enabled")
		} else {
			logx.Info("Using httpx for fingerprint detection")
		}
		s.runHttpx(ctx, httpAssets, opts)
	} else if !opts.Httpx && opts.Screenshot {
		logx.Info("Screenshot enabled without httpx, using builtin fingerprint method with chromedp")
	} else if opts.Httpx && !httpxInstalled {
		logx.Info("httpx not installed, using builtin fingerprint method")
	}

	for _, asset := range httpAssets {
		select {
		case <-ctx.Done():
			logx.Info("Fingerprint scan cancelled by context")
			return result, ctx.Err()
		default:
			// 如果没有使用httpx，或者httpx没有获取到信息，使用内置方法
			if !useHttpx || (asset.Title == "" && asset.HttpStatus == "") {
				logx.Infof("Using builtin fingerprint for %s:%d", asset.Host, asset.Port)
				s.fingerprint(ctx, asset, opts)
			} else {
				// 使用了httpx，但仍需执行勾选的其他功能（不包括截图，因为httpx已处理）
				s.runAdditionalFingerprint(ctx, asset, opts)
			}
			result.Assets = append(result.Assets, asset)
		}
	}

	// 添加非HTTP资产到结果中（不进行指纹识别）
	for _, asset := range config.Assets {
		if !isHttpAsset(asset) {
			result.Assets = append(result.Assets, asset)
		}
	}

	return result, nil
}

// filterHttpAssets 过滤出HTTP/HTTPS相关的资产
func filterHttpAssets(assets []*Asset) []*Asset {
	httpAssets := make([]*Asset, 0)
	for _, asset := range assets {
		if isHttpAsset(asset) {
			httpAssets = append(httpAssets, asset)
		}
	}
	return httpAssets
}

// isHttpAsset 判断资产是否为HTTP/HTTPS服务
func isHttpAsset(asset *Asset) bool {
	// 1. 优先根据IsHTTP字段判断（端口扫描阶段已设置）
	if asset.IsHTTP {
		return true
	}
	
	// 2. 根据Service字段判断
	service := strings.ToLower(asset.Service)
	
	// 使用全局HTTP服务检查器（如果已设置）
	if globalHttpServiceChecker != nil {
		isHttp, found := globalHttpServiceChecker.IsHttpService(service)
		if found {
			return isHttp
		}
	} else {
		// 回退到默认的HTTP服务列表
		defaultHttpServices := map[string]bool{
			"http": true, "https": true, "http-proxy": true,
			"https-alt": true, "http-alt": true, "ajp12": true, "esmagent": true,
		}
		if defaultHttpServices[service] {
			return true
		}
	}
	
	// 3. 明确的非HTTP服务，直接排除
	nonHttpServices := map[string]bool{
		"ssh": true, "ftp": true, "smtp": true, "pop3": true, "imap": true,
		"mysql": true, "mssql": true, "oracle": true, "postgresql": true, "redis": true,
		"mongodb": true, "memcached": true, "elasticsearch": true,
		"dns": true, "snmp": true, "ldap": true, "smb": true, "netbios": true,
		"rdp": true, "vnc": true, "telnet": true, "rpc": true,
		"ntp": true, "tftp": true, "sip": true, "rtsp": true,
	}
	
	// 使用全局检查器判断非HTTP服务
	if globalHttpServiceChecker != nil {
		isHttp, found := globalHttpServiceChecker.IsHttpService(service)
		if found && !isHttp {
			return false
		}
	} else if nonHttpServices[service] {
		return false
	}
	
	// 4. 常见HTTP端口（高置信度）
	commonHttpPorts := map[int]bool{
		80: true, 443: true, 8080: true, 8443: true, 8000: true, 8888: true,
		8081: true, 8082: true, 8083: true, 8084: true, 8085: true,
		9000: true, 9001: true, 9090: true, 9443: true,
		3000: true, 3001: true, 4000: true, 5000: true, 5001: true,
		7001: true, 7002: true, // WebLogic
		8180: true,             // Tomcat
		8280: true, 8380: true, 8480: true, 8580: true,
		10000: true, 10001: true, 10080: true, 10443: true,
		8800: true, 8880: true, 8881: true,
		18080: true, 28080: true,
	}
	if commonHttpPorts[asset.Port] {
		return true
	}
	
	// 5. 如果Service为空且端口未知，标记为需要探测
	// 这些资产会在后续通过实际HTTP请求来验证
	if service == "" {
		return true // 让fingerprint函数去实际探测
	}
	
	return false
}

// HttpServiceChecker HTTP服务检查器接口
type HttpServiceChecker interface {
	IsHttpService(serviceName string) (isHttp bool, found bool)
}

// 全局HTTP服务检查器
var globalHttpServiceChecker HttpServiceChecker

// SetHttpServiceChecker 设置全局HTTP服务检查器
func SetHttpServiceChecker(checker HttpServiceChecker) {
	globalHttpServiceChecker = checker
}

// runAdditionalFingerprint 执行额外的指纹识别功能（httpx已执行后）
// 注意：截图功能由httpx的-screenshot参数处理，这里不再重复执行
func (s *FingerprintScanner) runAdditionalFingerprint(ctx context.Context, asset *Asset, opts *FingerprintOptions) {
	targetUrl := fmt.Sprintf("%s://%s:%d", asset.Service, asset.Host, asset.Port)
	if asset.Service == "" {
		if asset.Port == 443 || asset.Port == 8443 {
			targetUrl = fmt.Sprintf("https://%s:%d", asset.Host, asset.Port)
		} else {
			targetUrl = fmt.Sprintf("http://%s:%d", asset.Host, asset.Port)
		}
	}

	// 解析HTTP headers用于指纹识别
	var headers http.Header
	if asset.HttpHeader != "" {
		headers = parseHttpHeaders(asset.HttpHeader)
	}

	// 如果 httpx 没有获取到 body（可能是重定向等原因），使用内置方法重新获取
	var bodyBytes []byte
	if asset.HttpBody == "" || asset.Title == "" {
		logx.Debugf("httpx didn't get body/title for %s:%d, fetching with builtin client", asset.Host, asset.Port)
		resp, err := s.client.Get(targetUrl)
		if err == nil {
			defer resp.Body.Close()
			body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024*1024))
			bodyBytes = body
			if asset.HttpBody == "" {
				if len(body) > 50*1024 {
					asset.HttpBody = string(body[:50*1024]) + "\n...[truncated]"
				} else {
					asset.HttpBody = string(body)
				}
			}
			if asset.Title == "" {
				asset.Title = extractTitle(string(body))
			}
			if asset.Server == "" {
				asset.Server = resp.Header.Get("Server")
			}
			if headers == nil {
				headers = resp.Header
			}
			// 更新状态码为最终响应的状态码
			asset.HttpStatus = fmt.Sprintf("%d", resp.StatusCode)
		}
	} else {
		bodyBytes = []byte(asset.HttpBody)
	}

	// 收集所有指纹识别结果，用于智能合并
	appResults := make(map[string]*AppDetectionResult)

	// 解析现有的httpx结果
	for _, app := range asset.App {
		if strings.Contains(app, "[httpx]") {
			appName := extractAppName(app)
			// 移除版本号（格式如 Nginx:1.24.0）
			baseAppName := appName
			if colonIdx := strings.Index(appName, ":"); colonIdx > 0 {
				baseAppName = appName[:colonIdx]
			}
			baseAppNameLower := strings.ToLower(baseAppName)
			if appResults[baseAppNameLower] == nil {
				appResults[baseAppNameLower] = &AppDetectionResult{Name: baseAppName}
			}
			appResults[baseAppNameLower].Sources = append(appResults[baseAppNameLower].Sources, "httpx")
			appResults[baseAppNameLower].OriginalName = app // 保留原始名称（可能包含版本号）
		}
	}

	// 如果启用Wappalyzer，进行检测
	if opts.Wappalyzer && s.wappalyzerClient != nil {
		apps := s.wappalyzerClient.Fingerprint(headers, []byte(asset.HttpBody))
		logx.Debugf("Wappalyzer detected apps for %s:%d: %v", asset.Host, asset.Port, apps)
		
		for app := range apps {
			appNameLower := strings.ToLower(app)
			// 查找是否已有相同应用（使用小写key匹配）
			if result, exists := appResults[appNameLower]; exists {
				result.Sources = append(result.Sources, "wappalyzer")
			} else {
				appResults[appNameLower] = &AppDetectionResult{
					Name:         app,
					OriginalName: app,
					Sources:      []string{"wappalyzer"},
				}
			}
		}
	}

	// 如果启用IconHash且httpx没有获取到，单独获取MMH3 hash
	var faviconMMH3Hash string
	if opts.IconHash && asset.IconHash == "" {
		iconHash, iconData := s.getIconHashWithData(targetUrl)
		if iconHash != "" {
			asset.IconHash = iconHash
		}
		// 计算MMH3 hash用于自定义指纹匹配（ARL格式）
		if len(iconData) > 0 {
			faviconMMH3Hash = CalculateMMH3Hash(iconData)
		}
	} else if asset.IconHash != "" {
		// httpx已获取到hash，尝试获取favicon数据计算MMH3
		_, iconData := s.getIconHashWithData(targetUrl)
		if len(iconData) > 0 {
			faviconMMH3Hash = CalculateMMH3Hash(iconData)
		}
	}

	// 如果启用自定义指纹引擎，使用自定义格式的规则进行识别
	if opts.CustomEngine && s.customFingerprintEngine != nil {
		fpCount := s.customFingerprintEngine.GetFingerprintCount()
		// 使用原始字节数据进行GBK编码匹配
		if len(bodyBytes) == 0 {
			bodyBytes = []byte(asset.HttpBody)
		}
		// 从header字符串中提取所有Set-Cookie值
		var cookies string
		if headers != nil {
			// 尝试多种可能的key格式
			cookies = headers.Get("Set-Cookie")
			if cookies == "" {
				cookies = headers.Get("set-cookie")
			}
			if cookies == "" {
				cookies = headers.Get("set_cookie")
			}
		}
		// 如果从http.Header获取失败，直接从原始字符串中提取
		if cookies == "" && asset.HttpHeader != "" {
			cookies = extractCookiesFromHeader(asset.HttpHeader)
		}
		fpData := &FingerprintData{
			Title:        asset.Title,
			Body:         asset.HttpBody,
			BodyBytes:    bodyBytes, // 用于GBK编码匹配
			Headers:      headers,
			HeaderString: asset.HttpHeader, // 原始header字符串，用于httpx非标准格式匹配
			Server:       asset.Server,
			URL:          targetUrl,
			FaviconHash:  faviconMMH3Hash,
			Cookies:      cookies,
		}
		customApps := s.customFingerprintEngine.MatchWithId(fpData)
		logx.Debugf("Custom fingerprint engine (loaded %d fingerprints) detected apps for %s:%d: %v", fpCount, asset.Host, asset.Port, customApps)

		for _, customApp := range customApps {
			appNameLower := strings.ToLower(customApp.Name)
			// 查找是否已有相同应用（使用小写key匹配）
			if result, exists := appResults[appNameLower]; exists {
				result.Sources = append(result.Sources, "custom")
				result.CustomIDs = append(result.CustomIDs, customApp.Id)
			} else {
				appResults[appNameLower] = &AppDetectionResult{
					Name:         customApp.Name,
					OriginalName: customApp.Name,
					Sources:      []string{"custom"},
					CustomIDs:    []string{customApp.Id},
				}
			}
		}
	}

	// 重新构建asset.App列表，使用智能合并的结果
	asset.App = make([]string, 0, len(appResults))
	for _, result := range appResults {
		formattedApp := formatAppWithSources(result)
		asset.App = append(asset.App, formattedApp)
	}
}

// getIconHashWithData 获取favicon的hash值和原始数据
func (s *FingerprintScanner) getIconHashWithData(baseUrl string) (string, []byte) {
	// 尝试常见的favicon路径
	faviconPaths := []string{
		"/favicon.ico",
		"/favicon.png",
		"/static/favicon.ico",
		"/assets/favicon.ico",
	}

	for _, path := range faviconPaths {
		iconUrl := baseUrl + path
		resp, err := s.client.Get(iconUrl)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			continue
		}

		// 读取icon内容
		iconData, err := io.ReadAll(io.LimitReader(resp.Body, 1024*1024))
		if err != nil || len(iconData) == 0 {
			continue
		}

		// 计算MD5 hash（用于显示）
		hash := calculateIconHash(iconData)
		return hash, iconData
	}

	return "", nil
}

// fingerprint 识别单个资产指纹
func (s *FingerprintScanner) fingerprint(ctx context.Context, asset *Asset, opts *FingerprintOptions) {
	// 检查上下文是否已取消
	if ctx.Err() != nil {
		return
	}

	// 尝试HTTP和HTTPS
	schemes := []string{"http", "https"}
	if asset.Port == 443 || asset.Port == 8443 || asset.Port == 9443 {
		schemes = []string{"https", "http"}
	}

	var httpDetected bool
	for _, scheme := range schemes {
		// 检查上下文是否已取消
		if ctx.Err() != nil {
			return
		}

		targetUrl := fmt.Sprintf("%s://%s:%d", scheme, asset.Host, asset.Port)
		resp, err := s.client.Get(targetUrl)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		// 验证是否为有效的HTTP响应
		if !isValidHttpResponse(resp) {
			continue
		}

		httpDetected = true

		// 读取响应体（保留原始字节用于GBK编码匹配）
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024*1024)) // 限制1MB

		// 提取信息
		asset.HttpStatus = fmt.Sprintf("%d", resp.StatusCode)
		asset.HttpHeader = formatHeaders(resp.Header)
		// 限制HttpBody大小为50KB
		if len(body) > 50*1024 {
			asset.HttpBody = string(body[:50*1024]) + "\n...[truncated]"
		} else {
			asset.HttpBody = string(body)
		}
		asset.Title = extractTitle(string(body))
		asset.Server = resp.Header.Get("Server")
		asset.Service = scheme

		// 获取Icon Hash和原始数据
		var faviconMMH3Hash string
		if opts.IconHash {
			iconHash, iconData := s.getIconHashWithData(targetUrl)
			if iconHash != "" {
				asset.IconHash = iconHash
			}
			// 计算MMH3 hash用于ARL格式指纹匹配
			if len(iconData) > 0 {
				faviconMMH3Hash = CalculateMMH3Hash(iconData)
			}
		}

		// 使用wappalyzergo识别应用指纹
		var appResults = make(map[string]*AppDetectionResult)
		if opts.Wappalyzer && s.wappalyzerClient != nil {
			apps := s.identifyWithWappalyzer(resp.Header, body)
			for _, app := range apps {
				appNameLower := strings.ToLower(app)
				appResults[appNameLower] = &AppDetectionResult{
					Name:         app,
					OriginalName: app,
					Sources:      []string{"wappalyzer"},
				}
			}
		}

		// 使用自定义指纹引擎
		if opts.CustomEngine && s.customFingerprintEngine != nil {
			fpCount := s.customFingerprintEngine.GetFingerprintCount()
			fpData := &FingerprintData{
				Title:        asset.Title,
				Body:         asset.HttpBody,
				BodyBytes:    body, // 原始字节用于GBK编码匹配
				Headers:      resp.Header,
				HeaderString: asset.HttpHeader, // 原始header字符串
				Server:       asset.Server,
				URL:          targetUrl,
				FaviconHash:  faviconMMH3Hash,
				Cookies:      resp.Header.Get("Set-Cookie"),
			}
			customApps := s.customFingerprintEngine.MatchWithId(fpData)
			logx.Debugf("Custom fingerprint engine (loaded %d fingerprints) detected apps for %s:%d: %v", fpCount, asset.Host, asset.Port, customApps)
			
			for _, customApp := range customApps {
				appNameLower := strings.ToLower(customApp.Name)
				// 查找是否已有相同应用（使用小写key匹配）
				if result, exists := appResults[appNameLower]; exists {
					result.Sources = append(result.Sources, "custom")
					result.CustomIDs = append(result.CustomIDs, customApp.Id)
				} else {
					appResults[appNameLower] = &AppDetectionResult{
						Name:         customApp.Name,
						OriginalName: customApp.Name,
						Sources:      []string{"custom"},
						CustomIDs:    []string{customApp.Id},
					}
				}
			}
		}

		// 构建最终的应用列表
		for _, result := range appResults {
			formattedApp := formatAppWithSources(result)
			asset.App = append(asset.App, formattedApp)
		}

		// 截图
		if opts.Screenshot {
			screenshot := s.takeScreenshot(ctx, targetUrl)
			logx.Infof("takeScreenshot截图: targetUrl:%s ->screenshot)", targetUrl)
			if screenshot != "" {
				asset.Screenshot = screenshot
			}
		}

		break
	}

	// 如果HTTP探测失败，标记为非HTTP服务
	if !httpDetected && asset.Service == "" {
		logx.Debugf("HTTP probe failed for %s:%d, marking as non-http", asset.Host, asset.Port)
		asset.Service = "unknown"
	}
}

// isValidHttpResponse 验证响应是否为有效的HTTP响应
func isValidHttpResponse(resp *http.Response) bool {
	if resp == nil {
		return false
	}
	
	// 检查状态码是否在有效范围内
	if resp.StatusCode < 100 || resp.StatusCode >= 600 {
		return false
	}
	
	// 检查是否有HTTP特征的响应头
	// 有效的HTTP服务通常会返回以下头之一
	httpHeaders := []string{
		"Content-Type",
		"Server",
		"Date",
		"Content-Length",
		"Transfer-Encoding",
		"Connection",
		"Set-Cookie",
		"X-Powered-By",
	}
	
	for _, header := range httpHeaders {
		if resp.Header.Get(header) != "" {
			return true
		}
	}
	
	// 如果状态码是常见的HTTP状态码，也认为是有效的
	validStatusCodes := map[int]bool{
		200: true, 201: true, 204: true, 206: true,
		301: true, 302: true, 303: true, 304: true, 307: true, 308: true,
		400: true, 401: true, 403: true, 404: true, 405: true, 500: true, 502: true, 503: true,
	}
	
	return validStatusCodes[resp.StatusCode]
}


// runHttpx 使用httpx进行批量探测
func (s *FingerprintScanner) runHttpx(ctx context.Context, assets []*Asset, opts *FingerprintOptions) {
	if len(assets) == 0 {
		return
	}

	// 构建目标列表
	var targets []string
	targetMap := make(map[string]*Asset)
	for _, asset := range assets {
		target := fmt.Sprintf("%s:%d", asset.Host, asset.Port)
		targets = append(targets, target)
		targetMap[target] = asset
	}

	// 构建httpx命令
	args := []string{
		"-silent",
		"-json",
		"-title",
		"-status-code",
		"-tech-detect",
		"-favicon",
		"-server",
		"-content-type",
		"-irh",              // include response header
		"-irr",              // include request/response (包含body)
		"-follow-redirects", // 跟随重定向
		"-max-redirects", "5", // 最大重定向次数
	}
	if opts.Screenshot {
		args = append(args, "-screenshot")
	}

	logx.Infof("Executing command: httpx %s", strings.Join(args, " "))

	cmd := exec.CommandContext(ctx, "httpx", args...)
	stdin, _ := cmd.StdinPipe()
	stdout, _ := cmd.StdoutPipe()

	if err := cmd.Start(); err != nil {
		logx.Errorf("httpx start error: %v", err)
		return
	}

	// 写入目标
	go func() {
		for _, target := range targets {
			stdin.Write([]byte(target + "\n"))
		}
		stdin.Close()
	}()

	// 解析输出
	decoder := json.NewDecoder(stdout)
	for decoder.More() {
		var result HttpxResult
		if err := decoder.Decode(&result); err != nil {
			continue
		}

		// 优先使用input字段匹配原始目标，避免重定向导致的匹配问题
		var asset *Asset
		var key string
		
		// 首先尝试使用input字段匹配
		if result.Input != "" {
			key = result.Input
			asset = targetMap[key]
		}
		
		// 如果input匹配失败，尝试从URL解析
		if asset == nil {
			if u, err := url.Parse(result.URL); err == nil {
				host := u.Hostname()
				port := u.Port()
				if port == "" {
					if u.Scheme == "https" {
						port = "443"
					} else {
						port = "80"
					}
				}
				key = fmt.Sprintf("%s:%s", host, port)
				asset = targetMap[key]
			}
		}
		
		if asset != nil {
			// 从URL获取scheme
			scheme := "http"
			if u, err := url.Parse(result.URL); err == nil {
				scheme = u.Scheme
			}
			asset.Title = result.Title
			asset.HttpStatus = fmt.Sprintf("%d", result.StatusCode)
			asset.Service = scheme
			if len(result.Technologies) > 0 {
				// 为httpx识别的应用添加来源标识
				for _, tech := range result.Technologies {
					asset.App = append(asset.App, tech+"[httpx]")
				}
			}
			if result.FaviconHash != "" {
				asset.IconHash = result.FaviconHash
			}
			if result.ScreenshotPath != "" {
				// 读取截图文件并转换为base64
				asset.Screenshot = readScreenshotAsBase64(result.ScreenshotPath)
			}
			// 填充Server字段
			if result.ServerHeader != "" {
				asset.Server = result.ServerHeader
			} else if result.WebServer != "" {
				asset.Server = result.WebServer
			}
			// 填充HttpHeader字段
			// 优先从Response字段提取完整header（包含所有Set-Cookie）
			if result.Response != "" {
				asset.HttpHeader = extractHeadersFromResponse(result.Response)
			}
			// 如果Response为空，使用ResponseHeader map
			if asset.HttpHeader == "" && len(result.ResponseHeader) > 0 {
				asset.HttpHeader = formatHttpxHeaders(result.ResponseHeader)
			}
			// 填充HttpBody字段
			bodyContent := result.ResponseBody
			asset.HttpBody = bodyContent
			logx.Debugf("Matched httpx result for %s: title=%s, status=%d", key, result.Title, result.StatusCode)
		}
	}

	cmd.Wait()
}

// HttpxResult httpx JSON输出结构
type HttpxResult struct {
	URL            string            `json:"url"`
	Input          string            `json:"input"`
	Title          string            `json:"title"`
	StatusCode     int               `json:"status_code"`
	Technologies   []string          `json:"tech"`
	FaviconHash    string            `json:"favicon_hash"`
	ScreenshotPath string            `json:"screenshot_path"`
	WebServer      string            `json:"webserver"`
	ContentLength  int               `json:"content_length"`
	ResponseHeader map[string]string `json:"header"`
	ServerHeader   string            `json:"server"`
	ContentType    string            `json:"content_type"`
	ResponseBody   string            `json:"body"`
	Response       string            `json:"response"` // httpx -include-response 输出的字段名
}

// checkHttpxInstalled 检查httpx是否安装
func checkHttpxInstalled() bool {
	cmd := exec.Command("httpx", "-version")
	output, _ := cmd.CombinedOutput()
	return strings.Contains(string(output), "Version")
}


// identifyWithWappalyzer 使用wappalyzergo识别应用
func (s *FingerprintScanner) identifyWithWappalyzer(headers http.Header, body []byte) []string {
	if s.wappalyzerClient == nil {
		return nil
	}

	fingerprints := s.wappalyzerClient.Fingerprint(headers, body)
	apps := make([]string, 0, len(fingerprints))
	for app := range fingerprints {
		apps = append(apps, app)
	}
	return apps
}

// getIconHash 获取favicon的hash值
func (s *FingerprintScanner) getIconHash(baseUrl string) string {
	// 尝试常见的favicon路径
	faviconPaths := []string{
		"/favicon.ico",
		"/favicon.png",
		"/static/favicon.ico",
		"/assets/favicon.ico",
	}

	for _, path := range faviconPaths {
		iconUrl := baseUrl + path
		resp, err := s.client.Get(iconUrl)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			continue
		}

		// 读取icon内容
		iconData, err := io.ReadAll(io.LimitReader(resp.Body, 1024*1024))
		if err != nil || len(iconData) == 0 {
			continue
		}

		// 计算MMH3 hash (Shodan风格)
		hash := calculateIconHash(iconData)
		return hash
	}

	return ""
}

// calculateIconHash 计算icon hash
func calculateIconHash(data []byte) string {
	sum := md5.Sum(data)
	return hex.EncodeToString(sum[:])
}

// takeScreenshot 使用chromedp截图
func (s *FingerprintScanner) takeScreenshot(ctx context.Context, targetUrl string) string {
	// 创建chromedp上下文，设置超时
	screenshotCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// 配置chromedp选项，支持 Docker 环境中的 Chromium
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("ignore-certificate-errors", true),
		chromedp.Flag("disable-software-rasterizer", true),
		chromedp.WindowSize(1920, 1080),
	)

	// 检查环境变量中是否指定了 Chrome 路径
	if chromePath := os.Getenv("CHROME_BIN"); chromePath != "" {
		opts = append(opts, chromedp.ExecPath(chromePath))
	}

	allocCtx, allocCancel := chromedp.NewExecAllocator(screenshotCtx, opts...)
	defer allocCancel()

	taskCtx, taskCancel := chromedp.NewContext(allocCtx)
	defer taskCancel()

	var buf []byte
	err := chromedp.Run(taskCtx,
		chromedp.Navigate(targetUrl),
		chromedp.Sleep(2*time.Second),
		chromedp.FullScreenshot(&buf, 90),
	)

	if err != nil {
		logx.Errorf("Screenshot failed for %s: %v", targetUrl, err)
		return ""
	}
	logx.Infof("完成使用chromedp截图: %s",targetUrl)
	// 返回base64编码的截图
	if len(buf) > 0 {
		return base64.StdEncoding.EncodeToString(buf)
	}
	return ""
}


// extractTitle 提取网页标题
func extractTitle(body string) string {
	re := regexp.MustCompile(`(?i)<title[^>]*>([^<]+)</title>`)
	matches := re.FindStringSubmatch(body)
	if len(matches) > 1 {
		title := strings.TrimSpace(matches[1])
		// 限制长度
		if len(title) > 100 {
			title = title[:100]
		}
		return title
	}
	return ""
}

// formatHeaders 格式化响应头
func formatHeaders(headers http.Header) string {
	var sb strings.Builder
	for key, values := range headers {
		for _, value := range values {
			sb.WriteString(fmt.Sprintf("%s: %s\n", key, value))
		}
	}
	return sb.String()
}

// formatHttpxHeaders 格式化httpx返回的响应头
func formatHttpxHeaders(headers map[string]string) string {
	var sb strings.Builder
	for key, value := range headers {
		sb.WriteString(fmt.Sprintf("%s: %s\n", key, value))
	}
	return sb.String()
}

// extractHeadersFromResponse 从完整HTTP响应中提取header部分
func extractHeadersFromResponse(response string) string {
	// HTTP响应格式: 状态行 + headers + 空行 + body
	// 找到header和body的分隔（空行）
	idx := strings.Index(response, "\r\n\r\n")
	if idx == -1 {
		idx = strings.Index(response, "\n\n")
	}
	if idx == -1 {
		// 没有找到分隔，可能整个都是header
		return response
	}
	return response[:idx]
}

// parseHttpHeaders 解析HTTP headers字符串为http.Header
func parseHttpHeaders(headerStr string) http.Header {
	headers := make(http.Header)
	lines := strings.Split(headerStr, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			headers.Add(key, value)
		}
	}
	return headers
}

// extractCookiesFromHeader 从header字符串中提取所有Set-Cookie值
func extractCookiesFromHeader(headerStr string) string {
	var cookies []string
	lines := strings.Split(headerStr, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// 不区分大小写匹配 Set-Cookie, set-cookie, set_cookie
		lowerLine := strings.ToLower(line)
		if strings.HasPrefix(lowerLine, "set-cookie:") || strings.HasPrefix(lowerLine, "set_cookie:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				cookies = append(cookies, strings.TrimSpace(parts[1]))
			}
		}
	}
	return strings.Join(cookies, "; ")
}

// readScreenshotAsBase64 读取截图文件并返回base64编码
func readScreenshotAsBase64(filePath string) string {
	if filePath == "" {
		return ""
	}
	data, err := os.ReadFile(filePath)
	if err != nil {
		logx.Errorf("Failed to read screenshot file %s: %v", filePath, err)
		return ""
	}
	if len(data) == 0 {
		return ""
	}
	return base64.StdEncoding.EncodeToString(data)
}

// uniqueStrings 去重字符串切片
func uniqueStrings(slice []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0)
	for _, s := range slice {
		if !seen[s] {
			seen[s] = true
			result = append(result, s)
		}
	}
	return result
}

// containsAppName 检查应用列表中是否包含指定应用名（忽略来源标识）
func containsAppName(apps []string, appName string) bool {
	appNameLower := strings.ToLower(appName)
	for _, app := range apps {
		// 移除来源标识后比较
		name := app
		if idx := strings.Index(app, "["); idx > 0 {
			name = app[:idx]
		}
		if strings.ToLower(name) == appNameLower {
			return true
		}
	}
	return false
}

// findAppIndex 查找应用在列表中的索引，匹配应用名和指定来源标识
// 支持匹配带版本号的应用名，如 "Nginx:1.24.0" 匹配 "nginx"
func findAppIndex(apps []string, appName string, source string) int {
	appNameLower := strings.ToLower(appName)
	for i, app := range apps {
		// 检查是否包含指定来源标识
		if !strings.Contains(app, source) {
			continue
		}
		// 移除来源标识后获取应用名
		name := app
		if idx := strings.Index(app, "["); idx > 0 {
			name = app[:idx]
		}
		// 移除版本号（格式如 Nginx:1.24.0）
		if colonIdx := strings.Index(name, ":"); colonIdx > 0 {
			name = name[:colonIdx]
		}
		if strings.ToLower(name) == appNameLower {
			return i
		}
	}
	return -1
}

// extractAppName 从应用字符串中提取应用名称（移除来源标识）
func extractAppName(app string) string {
	if idx := strings.Index(app, "["); idx > 0 {
		return strings.TrimSpace(app[:idx])
	}
	return app
}

// formatAppWithSources 根据检测来源格式化应用名称
func formatAppWithSources(result *AppDetectionResult) string {
	if len(result.Sources) == 0 {
		return result.OriginalName
	}

	// 使用原始名称（可能包含版本号）
	appName := result.OriginalName
	if appName == "" {
		appName = result.Name
	}

	// 如果应用名称仍为空，跳过
	if appName == "" {
		return ""
	}

	// 移除现有的来源标识
	if idx := strings.Index(appName, "["); idx > 0 {
		appName = strings.TrimSpace(appName[:idx])
	}

	// 去重并排序来源
	sources := uniqueStrings(result.Sources)
	
	// 构建来源标识
	var sourceStr string
	if len(sources) == 1 {
		switch sources[0] {
		case "custom":
			if len(result.CustomIDs) > 0 {
				sourceStr = fmt.Sprintf("[custom(%s)]", strings.Join(result.CustomIDs, ","))
			} else {
				sourceStr = "[custom]"
			}
		default:
			sourceStr = fmt.Sprintf("[%s]", sources[0])
		}
	} else {
		// 多个来源，按优先级排序：httpx > wappalyzer > custom
		var orderedSources []string
		for _, source := range []string{"httpx", "wappalyzer", "custom"} {
			for _, s := range sources {
				if s == source {
					orderedSources = append(orderedSources, s)
					break
				}
			}
		}
		
		// 构建合并的来源标识
		if containsString(orderedSources, "custom") && len(result.CustomIDs) > 0 {
			// 包含自定义指纹，需要特殊处理
			var nonCustomSources []string
			for _, s := range orderedSources {
				if s != "custom" {
					nonCustomSources = append(nonCustomSources, s)
				}
			}
			if len(nonCustomSources) > 0 {
				sourceStr = fmt.Sprintf("[%s+custom(%s)]", strings.Join(nonCustomSources, "+"), strings.Join(result.CustomIDs, ","))
			} else {
				sourceStr = fmt.Sprintf("[custom(%s)]", strings.Join(result.CustomIDs, ","))
			}
		} else {
			sourceStr = fmt.Sprintf("[%s]", strings.Join(orderedSources, "+"))
		}
	}

	return appName + sourceStr
}

// containsString 检查字符串切片是否包含指定字符串
func containsString(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}
