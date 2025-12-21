package scanner

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"cscan/model"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

// CustomFingerprintEngine 自定义指纹识别引擎
// 支持ARL finger.json/webapp.json格式的规则语法
// 参考ARL项目的指纹识别方式进行优化
type CustomFingerprintEngine struct {
	fingerprints []*model.Fingerprint
}

// NewCustomFingerprintEngine 创建自定义指纹引擎
func NewCustomFingerprintEngine(fingerprints []*model.Fingerprint) *CustomFingerprintEngine {
	return &CustomFingerprintEngine{
		fingerprints: fingerprints,
	}
}

// FingerprintData 用于指纹匹配的数据
// 参考ARL的fetch_fingerprint函数设计
type FingerprintData struct {
	Title        string      // 网页标题
	Body         string      // 网页内容（UTF-8）
	BodyBytes    []byte      // 网页原始字节（用于GBK编码匹配）
	Headers      http.Header // HTTP响应头（标准格式）
	HeaderString string      // HTTP响应头原始字符串（用于httpx等非标准格式匹配）
	Server       string      // Server头
	URL          string      // 请求URL
	FaviconHash  string      // favicon的MMH3 hash（Shodan风格）
	Cookies      string      // Set-Cookie头内容
}

// MatchedFingerprint 匹配到的指纹结果
type MatchedFingerprint struct {
	Name string // 指纹名称
	Id   string // 指纹ID（MongoDB ObjectID）
}

// GetFingerprintCount 返回已加载的指纹数量
func (e *CustomFingerprintEngine) GetFingerprintCount() int {
	if e == nil {
		return 0
	}
	return len(e.fingerprints)
}

// Match 执行指纹匹配，返回匹配到的应用名称列表（兼容旧接口）
func (e *CustomFingerprintEngine) Match(data *FingerprintData) []string {
	results := e.MatchWithId(data)
	names := make([]string, len(results))
	for i, r := range results {
		names[i] = r.Name
	}
	return names
}

// MatchWithId 执行指纹匹配，返回匹配到的应用名称和ID列表
func (e *CustomFingerprintEngine) MatchWithId(data *FingerprintData) []MatchedFingerprint {
	var matched []MatchedFingerprint
	seen := make(map[string]bool)

	// 检查指纹数量
	if len(e.fingerprints) == 0 {
		return matched
	}

	for _, fp := range e.fingerprints {
		if !fp.Enabled {
			continue
		}

		// 优先使用Rule字段（ARL格式规则语法）
		if fp.Rule != "" {
			if e.matchRule(fp.Rule, data) {
				if !seen[fp.Name] {
					matched = append(matched, MatchedFingerprint{
						Name: fp.Name,
						Id:   fp.Id.Hex(),
					})
					seen[fp.Name] = true
				}
			}
			continue
		}

		// 使用ARL webapp.json格式规则（html/title/headers数组）
		if e.matchARLWebappRules(fp, data) {
			if !seen[fp.Name] {
				matched = append(matched, MatchedFingerprint{
					Name: fp.Name,
					Id:   fp.Id.Hex(),
				})
				seen[fp.Name] = true
			}
			continue
		}

		// 使用Wappalyzer格式规则
		if e.matchWappalyzerRules(fp, data) {
			if !seen[fp.Name] {
				matched = append(matched, MatchedFingerprint{
					Name: fp.Name,
					Id:   fp.Id.Hex(),
				})
				seen[fp.Name] = true
			}
		}
	}

	return matched
}

// matchARLWebappRules 匹配ARL webapp.json格式规则
// 参考ARL的fetch_fingerprint函数：单个规则字段内是OR关系
func (e *CustomFingerprintEngine) matchARLWebappRules(fp *model.Fingerprint, data *FingerprintData) bool {
	// 检查是否有ARL格式的规则
	hasRule := false

	// HTML/Body匹配 - 支持UTF-8和GBK双编码匹配
	if len(fp.HTML) > 0 {
		hasRule = true
		for _, keyword := range fp.HTML {
			if e.matchBodyWithEncoding(data, keyword) {
				return true // OR关系，匹配一个即可
			}
		}
	}

	// Headers匹配
	if len(fp.Headers) > 0 {
		hasRule = true
		for key, pattern := range fp.Headers {
			// ARL格式：headers数组中的值直接在整个header中搜索
			if pattern == "" {
				// 只检查key是否存在
				if data.Headers.Get(key) != "" {
					return true
				}
			} else {
				// 检查header值是否包含pattern
				headerStr := formatHeadersToString(data.Headers)
				if containsIgnoreCase(headerStr, pattern) {
					return true
				}
			}
		}
	}

	// 如果没有任何规则，返回false
	return hasRule && false
}

// matchRule 匹配ARL格式规则
// 支持: body="xxx", title="xxx", header="xxx", server="xxx"
// 逻辑: && (AND), || (OR)
func (e *CustomFingerprintEngine) matchRule(rule string, data *FingerprintData) bool {
	rule = strings.TrimSpace(rule)
	if rule == "" {
		return false
	}

	// 处理OR逻辑 (||) - 优先级低于AND
	parts := splitByOperator(rule, "||")
	if len(parts) > 1 {
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			if e.matchRuleAnd(part, data) {
				return true
			}
		}
		return false
	}

	// 没有OR，处理AND逻辑
	return e.matchRuleAnd(rule, data)
}

// matchRuleAnd 处理AND逻辑
func (e *CustomFingerprintEngine) matchRuleAnd(rule string, data *FingerprintData) bool {
	rule = strings.TrimSpace(rule)
	if rule == "" {
		return false
	}

	parts := splitByOperator(rule, "&&")
	if len(parts) > 1 {
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			if !e.matchSingleCondition(part, data) {
				return false
			}
		}
		return true
	}

	// 单个条件匹配
	return e.matchSingleCondition(rule, data)
}

// splitByOperator 按操作符分割，考虑引号内的内容
func splitByOperator(rule, op string) []string {
	var parts []string
	var current strings.Builder
	inQuote := false
	quoteChar := byte(0)

	for i := 0; i < len(rule); i++ {
		c := rule[i]

		if (c == '"' || c == '\'') && (i == 0 || rule[i-1] != '\\') {
			if !inQuote {
				inQuote = true
				quoteChar = c
			} else if c == quoteChar {
				inQuote = false
			}
		}

		if !inQuote && i+len(op) <= len(rule) && rule[i:i+len(op)] == op {
			parts = append(parts, strings.TrimSpace(current.String()))
			current.Reset()
			i += len(op) - 1
			continue
		}

		current.WriteByte(c)
	}

	if current.Len() > 0 {
		parts = append(parts, strings.TrimSpace(current.String()))
	}

	return parts
}

// matchSingleCondition 匹配单个条件
// 参考ARL的规则格式：body="xxx", title="xxx", header="xxx", icon_hash="xxx"
func (e *CustomFingerprintEngine) matchSingleCondition(condition string, data *FingerprintData) bool {
	condition = strings.TrimSpace(condition)

	// 解析条件: type="value" 或 type!="value"
	var condType, value string
	var negate bool

	if idx := strings.Index(condition, "!=\""); idx > 0 {
		condType = strings.TrimSpace(condition[:idx])
		rawValue := condition[idx+3:]
		value = extractQuotedValue(rawValue)
		negate = true
	} else if idx := strings.Index(condition, "=\""); idx > 0 {
		condType = strings.TrimSpace(condition[:idx])
		rawValue := condition[idx+2:]
		value = extractQuotedValue(rawValue)
		negate = false
	} else {
		// 尝试简单的 = 分割（不带引号的格式）
		if idx := strings.Index(condition, "="); idx > 0 {
			condType = strings.TrimSpace(condition[:idx])
			value = strings.TrimSpace(condition[idx+1:])
			// 去掉可能的引号
			value = strings.Trim(value, "\"'")
			negate = false
		} else {
			return false
		}
	}

	var result bool
	switch strings.ToLower(condType) {
	case "body":
		// 参考ARL：同时支持UTF-8和GBK编码匹配
		result = e.matchBodyWithEncoding(data, value)
	case "title":
		result = containsIgnoreCase(data.Title, value)
	case "header":
		// 同时检查标准http.Header和原始header字符串
		// 因为httpx返回的header格式可能是非标准的（如set_cookie而不是Set-Cookie）
		headerResult := containsInHeaders(data.Headers, value)
		headerStrResult := containsInHeaderString(data.HeaderString, value)
		result = headerResult || headerStrResult
	case "server":
		result = containsIgnoreCase(data.Server, value)
	case "url":
		result = containsIgnoreCase(data.URL, value)
	case "body_regex", "body_re":
		result = matchRegex(data.Body, value)
	case "title_regex", "title_re":
		result = matchRegex(data.Title, value)
	// ARL特有的icon_hash匹配
	case "icon_hash", "favicon_hash":
		result = e.matchIconHash(data, value)
	// Cookie匹配（用于识别Shiro等框架）
	case "cookie":
		// 同时检查Cookies字段和header字符串中的cookie
		result = containsIgnoreCase(data.Cookies, value) || containsInHeaderString(data.HeaderString, value)
	default:
		return false
	}

	if negate {
		return !result
	}
	return result
}

// matchBodyWithEncoding 同时支持UTF-8和GBK编码匹配
// 参考ARL的fetch_fingerprint函数实现
func (e *CustomFingerprintEngine) matchBodyWithEncoding(data *FingerprintData, keyword string) bool {
	// 1. 先尝试UTF-8匹配
	if containsIgnoreCase(data.Body, keyword) {
		return true
	}

	// 2. 尝试将keyword转换为GBK编码后在原始字节中匹配
	// 这是ARL的做法：html.encode("gbk") in content
	if len(data.BodyBytes) > 0 {
		gbkKeyword, err := encodeToGBK(keyword)
		if err == nil && len(gbkKeyword) > 0 {
			if bytes.Contains(data.BodyBytes, gbkKeyword) {
				return true
			}
		}
	}

	return false
}

// matchIconHash 匹配favicon hash
// 参考ARL的icon_hash匹配方式
func (e *CustomFingerprintEngine) matchIconHash(data *FingerprintData, expectedHash string) bool {
	if data.FaviconHash == "" || expectedHash == "" {
		return false
	}
	// 支持直接比较或包含比较
	return data.FaviconHash == expectedHash || strings.Contains(data.FaviconHash, expectedHash)
}

// encodeToGBK 将UTF-8字符串转换为GBK编码
func encodeToGBK(s string) ([]byte, error) {
	reader := transform.NewReader(strings.NewReader(s), simplifiedchinese.GBK.NewEncoder())
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(reader)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// formatHeadersToString 将http.Header格式化为字符串
func formatHeadersToString(headers http.Header) string {
	var sb strings.Builder
	for key, values := range headers {
		for _, value := range values {
			sb.WriteString(key)
			sb.WriteString(": ")
			sb.WriteString(value)
			sb.WriteString("\n")
		}
	}
	return sb.String()
}

// extractQuotedValue 提取引号内的值
func extractQuotedValue(s string) string {
	s = strings.TrimSpace(s)
	if len(s) == 0 {
		return s
	}

	// 检查是否以引号开头
	if s[0] == '"' || s[0] == '\'' {
		quoteChar := s[0]
		// 从第二个字符开始找结束引号
		for i := 1; i < len(s); i++ {
			if s[i] == quoteChar {
				// 检查是否是转义的引号
				if i > 1 && s[i-1] == '\\' {
					continue
				}
				return s[1:i]
			}
		}
		// 没有找到结束引号，返回去掉开头引号的内容
		return s[1:]
	}

	// 不以引号开头，检查是否以引号结尾（处理格式错误的规则）
	if s[len(s)-1] == '"' || s[len(s)-1] == '\'' {
		return s[:len(s)-1]
	}

	return s
}

// containsIgnoreCase 不区分大小写的包含检查
func containsIgnoreCase(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

// containsInHeaders 检查headers中是否包含指定值
func containsInHeaders(headers http.Header, value string) bool {
	if headers == nil {
		return false
	}
	valueLower := strings.ToLower(value)
	for key, values := range headers {
		if strings.Contains(strings.ToLower(key), valueLower) {
			return true
		}
		for _, v := range values {
			if strings.Contains(strings.ToLower(v), valueLower) {
				return true
			}
		}
	}
	return false
}

// containsInHeaderString 检查header字符串中是否包含指定值
// 用于处理httpx返回的非标准header格式
func containsInHeaderString(headerStr, value string) bool {
	if headerStr == "" || value == "" {
		return false
	}
	return strings.Contains(strings.ToLower(headerStr), strings.ToLower(value))
}

// matchRegex 正则匹配
func matchRegex(s, pattern string) bool {
	re, err := regexp.Compile("(?i)" + pattern)
	if err != nil {
		return false
	}
	return re.MatchString(s)
}

// matchWappalyzerRules 匹配Wappalyzer格式规则
// Wappalyzer的html、scripts、css等字段是正则表达式
func (e *CustomFingerprintEngine) matchWappalyzerRules(fp *model.Fingerprint, data *FingerprintData) bool {
	hasRule := false
	allMatch := true

	// Headers匹配 - key需要大小写不敏感匹配
	if len(fp.Headers) > 0 {
		hasRule = true
		headerMatch := false
		for key, pattern := range fp.Headers {
			// 遍历响应头，大小写不敏感匹配key
			for hKey, hVal := range data.Headers {
				if strings.EqualFold(hKey, key) {
					headerValue := strings.Join(hVal, " ")
					if pattern == "" {
						// 只要header存在就匹配
						headerMatch = true
						break
					}
					// pattern是正则表达式
					if matchRegexOrContains(headerValue, pattern) {
						headerMatch = true
						break
					}
				}
			}
			if headerMatch {
				break
			}
		}
		if !headerMatch {
			allMatch = false
		}
	}

	// HTML匹配 - pattern是正则表达式
	if len(fp.HTML) > 0 && allMatch {
		hasRule = true
		htmlMatch := false
		for _, pattern := range fp.HTML {
			if matchRegexOrContains(data.Body, pattern) {
				htmlMatch = true
				break
			}
		}
		if !htmlMatch {
			allMatch = false
		}
	}

	// Meta匹配
	if len(fp.Meta) > 0 && allMatch {
		hasRule = true
		metaMatch := false
		for name, pattern := range fp.Meta {
			if matchMetaTag(data.Body, name, pattern) {
				metaMatch = true
				break
			}
		}
		if !metaMatch {
			allMatch = false
		}
	}

	// Scripts匹配 - pattern是正则表达式，匹配script标签的src属性
	if len(fp.Scripts) > 0 && allMatch {
		hasRule = true
		scriptMatch := false
		// 提取所有script src
		scriptSrcRe := regexp.MustCompile(`(?i)<script[^>]*src=["']([^"']+)["']`)
		scriptSrcs := scriptSrcRe.FindAllStringSubmatch(data.Body, -1)
		for _, pattern := range fp.Scripts {
			for _, src := range scriptSrcs {
				if len(src) > 1 && matchRegexOrContains(src[1], pattern) {
					scriptMatch = true
					break
				}
			}
			if scriptMatch {
				break
			}
		}
		if !scriptMatch {
			allMatch = false
		}
	}

	// ScriptSrc匹配
	if len(fp.ScriptSrc) > 0 && allMatch {
		hasRule = true
		scriptSrcMatch := false
		scriptSrcRe := regexp.MustCompile(`(?i)<script[^>]*src=["']([^"']+)["']`)
		scriptSrcs := scriptSrcRe.FindAllStringSubmatch(data.Body, -1)
		for _, pattern := range fp.ScriptSrc {
			for _, src := range scriptSrcs {
				if len(src) > 1 && matchRegexOrContains(src[1], pattern) {
					scriptSrcMatch = true
					break
				}
			}
			if scriptSrcMatch {
				break
			}
		}
		if !scriptSrcMatch {
			allMatch = false
		}
	}

	// CSS匹配
	if len(fp.CSS) > 0 && allMatch {
		hasRule = true
		cssMatch := false
		for _, pattern := range fp.CSS {
			if matchRegexOrContains(data.Body, pattern) {
				cssMatch = true
				break
			}
		}
		if !cssMatch {
			allMatch = false
		}
	}

	// Cookies匹配
	if len(fp.Cookies) > 0 && allMatch {
		hasRule = true
		cookieMatch := false
		// 同时检查Cookies字段和header中的Set-Cookie
		cookieStr := data.Cookies
		if cookieStr == "" && data.Headers != nil {
			cookieStr = data.Headers.Get("Set-Cookie")
		}
		for name, pattern := range fp.Cookies {
			if containsIgnoreCase(cookieStr, name) {
				if pattern == "" || matchRegexOrContains(cookieStr, pattern) {
					cookieMatch = true
					break
				}
			}
		}
		if !cookieMatch {
			allMatch = false
		}
	}

	// URL匹配
	if len(fp.URL) > 0 && allMatch {
		hasRule = true
		urlMatch := false
		for _, pattern := range fp.URL {
			if matchRegexOrContains(data.URL, pattern) {
				urlMatch = true
				break
			}
		}
		if !urlMatch {
			allMatch = false
		}
	}

	return hasRule && allMatch
}

// matchRegexOrContains 尝试正则匹配，如果正则无效则回退到字符串包含匹配
func matchRegexOrContains(text, pattern string) bool {
	if pattern == "" {
		return true
	}
	// 尝试编译为正则表达式
	re, err := regexp.Compile("(?i)" + pattern)
	if err != nil {
		// 正则无效，回退到字符串包含匹配
		return containsIgnoreCase(text, pattern)
	}
	return re.MatchString(text)
}

// matchMetaTag 匹配Meta标签
func matchMetaTag(body, name, pattern string) bool {
	// 简单的meta标签匹配
	metaPattern := `(?i)<meta[^>]*name\s*=\s*["']?` + regexp.QuoteMeta(name) + `["']?[^>]*content\s*=\s*["']([^"']+)["']`
	re, err := regexp.Compile(metaPattern)
	if err != nil {
		return false
	}
	matches := re.FindStringSubmatch(body)
	if len(matches) < 2 {
		return false
	}
	if pattern == "" {
		return true
	}
	return matchRegex(matches[1], pattern)
}

// ParseARLFingerYAML 解析ARL finger.yml格式
type ARLFingerprint struct {
	Name string `yaml:"name"`
	Rule string `yaml:"rule"`
}

// ConvertARLToFingerprint 将ARL格式转换为Fingerprint
func ConvertARLToFingerprint(arl *ARLFingerprint) *model.Fingerprint {
	return &model.Fingerprint{
		Name:      arl.Name,
		Rule:      arl.Rule,
		Source:    "arl",
		IsBuiltin: false,
		Enabled:   true,
		Category:  guessCategory(arl.Name, arl.Rule),
	}
}

// guessCategory 根据名称和规则猜测分类
func guessCategory(name, rule string) string {
	nameLower := strings.ToLower(name)
	ruleLower := strings.ToLower(rule)

	if strings.Contains(nameLower, "oa") || strings.Contains(nameLower, "办公") {
		return "OA"
	}
	if strings.Contains(nameLower, "cms") {
		return "CMS"
	}
	if strings.Contains(nameLower, "erp") {
		return "ERP"
	}
	if strings.Contains(nameLower, "crm") {
		return "CRM"
	}
	if strings.Contains(nameLower, "vpn") || strings.Contains(ruleLower, "vpn") {
		return "VPN"
	}
	if strings.Contains(nameLower, "weblogic") || strings.Contains(nameLower, "tomcat") ||
		strings.Contains(nameLower, "nginx") || strings.Contains(nameLower, "apache") {
		return "Web servers"
	}
	if strings.Contains(nameLower, "spring") || strings.Contains(nameLower, "struts") {
		return "Web frameworks"
	}
	if strings.Contains(nameLower, "kibana") || strings.Contains(nameLower, "grafana") {
		return "Analytics"
	}
	if strings.Contains(nameLower, "jenkins") || strings.Contains(nameLower, "gitlab") {
		return "Dev tools"
	}

	return "Miscellaneous"
}

// CalculateMMH3Hash 计算Shodan风格的MMH3 favicon hash
// 这是ARL和Shodan使用的icon_hash算法
func CalculateMMH3Hash(data []byte) string {
	if len(data) == 0 {
		return ""
	}
	// Shodan的favicon hash计算方式：
	// 1. Base64编码
	// 2. 计算MMH3 hash
	b64 := base64.StdEncoding.EncodeToString(data)
	hash := mmh3Hash32([]byte(b64))
	return fmt.Sprintf("%d", int32(hash))
}

// mmh3Hash32 MurmurHash3 32位实现
// 参考Shodan的favicon hash算法
func mmh3Hash32(data []byte) uint32 {
	const (
		c1 = 0xcc9e2d51
		c2 = 0x1b873593
		r1 = 15
		r2 = 13
		m  = 5
		n  = 0xe6546b64
	)

	length := len(data)
	h := uint32(0) // seed = 0

	// 处理4字节块
	nblocks := length / 4
	for i := 0; i < nblocks; i++ {
		k := binary.LittleEndian.Uint32(data[i*4:])
		k *= c1
		k = (k << r1) | (k >> (32 - r1))
		k *= c2

		h ^= k
		h = (h << r2) | (h >> (32 - r2))
		h = h*m + n
	}

	// 处理剩余字节
	tail := data[nblocks*4:]
	var k uint32
	switch len(tail) {
	case 3:
		k ^= uint32(tail[2]) << 16
		fallthrough
	case 2:
		k ^= uint32(tail[1]) << 8
		fallthrough
	case 1:
		k ^= uint32(tail[0])
		k *= c1
		k = (k << r1) | (k >> (32 - r1))
		k *= c2
		h ^= k
	}

	// 最终混合
	h ^= uint32(length)
	h ^= h >> 16
	h *= 0x85ebca6b
	h ^= h >> 13
	h *= 0xc2b2ae35
	h ^= h >> 16

	return h
}

// CalculateMD5Hash 计算MD5 hash（备用方案）
func CalculateMD5Hash(data []byte) string {
	if len(data) == 0 {
		return ""
	}
	sum := md5.Sum(data)
	return fmt.Sprintf("%x", sum)
}


// ARLFingerJSON ARL finger.json格式的指纹规则
type ARLFingerJSON struct {
	CMS      string   `json:"cms"`
	Method   string   `json:"method"`
	Location string   `json:"location"`
	Keyword  []string `json:"keyword"`
}

// ARLWebappJSON ARL webapp.json格式的指纹规则
type ARLWebappJSON struct {
	Cats     []int    `json:"cats"`
	Headers  []string `json:"headers"`
	HTML     []string `json:"html"`
	Title    []string `json:"title"`
	Icon     string   `json:"icon"`
	Website  string   `json:"website"`
	FofaRule string   `json:"fofa_rule"`
}

// ConvertARLFingerJSONToFingerprint 将ARL finger.json格式转换为Fingerprint
// 参考ARL的finger.json规则格式
func ConvertARLFingerJSONToFingerprint(arl *ARLFingerJSON) *model.Fingerprint {
	fp := &model.Fingerprint{
		Name:      arl.CMS,
		Source:    "arl-finger",
		IsBuiltin: false,
		Enabled:   true,
		Category:  guessCategory(arl.CMS, ""),
	}

	// 解析location字段，确定匹配类型
	// 格式: "rule: body", "rule: title", "rule: icon_hash"
	location := strings.ToLower(arl.Location)
	
	// 构建ARL格式的规则字符串
	var rules []string
	for _, kw := range arl.Keyword {
		if strings.Contains(location, "body") {
			rules = append(rules, fmt.Sprintf(`body="%s"`, kw))
		} else if strings.Contains(location, "title") {
			rules = append(rules, fmt.Sprintf(`title="%s"`, kw))
		} else if strings.Contains(location, "icon_hash") || strings.Contains(location, "header") {
			// icon_hash在ARL中有时放在header位置，实际是cookie或header匹配
			// 检查是否是数字（icon_hash）
			if isNumeric(kw) || strings.HasPrefix(kw, "-") {
				rules = append(rules, fmt.Sprintf(`icon_hash="%s"`, kw))
			} else {
				// 可能是cookie或header中的关键字
				rules = append(rules, fmt.Sprintf(`header="%s"`, kw))
			}
		}
	}
	
	// 多个关键字之间是OR关系
	if len(rules) > 0 {
		fp.Rule = strings.Join(rules, " || ")
	}

	return fp
}

// ConvertARLWebappJSONToFingerprint 将ARL webapp.json格式转换为Fingerprint
// 参考ARL的webapp.json规则格式
func ConvertARLWebappJSONToFingerprint(name string, arl *ARLWebappJSON) *model.Fingerprint {
	fp := &model.Fingerprint{
		Name:      name,
		Website:   arl.Website,
		Source:    "arl-webapp",
		IsBuiltin: false,
		Enabled:   true,
		Category:  guessCategory(name, ""),
	}

	// 构建规则字符串，各字段内是OR关系
	var allRules []string

	// HTML/Body规则
	for _, html := range arl.HTML {
		allRules = append(allRules, fmt.Sprintf(`body="%s"`, html))
	}

	// Title规则
	for _, title := range arl.Title {
		allRules = append(allRules, fmt.Sprintf(`title="%s"`, title))
	}

	// Headers规则
	for _, header := range arl.Headers {
		allRules = append(allRules, fmt.Sprintf(`header="%s"`, header))
	}

	// 所有规则之间是OR关系
	if len(allRules) > 0 {
		fp.Rule = strings.Join(allRules, " || ")
	}

	return fp
}

// isNumeric 检查字符串是否为数字（包括负数）
func isNumeric(s string) bool {
	if s == "" {
		return false
	}
	start := 0
	if s[0] == '-' || s[0] == '+' {
		start = 1
	}
	for i := start; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return start < len(s)
}

// BatchConvertARLFingerJSON 批量转换ARL finger.json格式规则
func BatchConvertARLFingerJSON(rules []ARLFingerJSON) []*model.Fingerprint {
	var fps []*model.Fingerprint
	seen := make(map[string]bool)
	
	for _, rule := range rules {
		fp := ConvertARLFingerJSONToFingerprint(&rule)
		// 合并同名规则（ARL中同一个CMS可能有多条规则）
		if existing, ok := seen[fp.Name]; ok && existing {
			// 找到已存在的规则，合并Rule字段
			for i, existingFp := range fps {
				if existingFp.Name == fp.Name && fp.Rule != "" {
					if existingFp.Rule != "" {
						fps[i].Rule = existingFp.Rule + " || " + fp.Rule
					} else {
						fps[i].Rule = fp.Rule
					}
					break
				}
			}
		} else {
			fps = append(fps, fp)
			seen[fp.Name] = true
		}
	}
	
	return fps
}
