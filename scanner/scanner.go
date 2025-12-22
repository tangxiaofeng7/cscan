package scanner

import (
	"context"
)

// Scanner 扫描器接口
type Scanner interface {
	// Name 扫描器名称
	Name() string
	// Scan 执行扫描
	Scan(ctx context.Context, config *ScanConfig) (*ScanResult, error)
}

// ScanConfig 扫描配置
type ScanConfig struct {
	Target      string      `json:"target"`
	Targets     []string    `json:"targets"`
	Assets      []*Asset    `json:"assets"`
	Options     interface{} `json:"options"`
	WorkspaceId string      `json:"workspaceId"`
	MainTaskId  string      `json:"mainTaskId"`
}

// ScanResult 扫描结果
type ScanResult struct {
	WorkspaceId     string           `json:"workspaceId"`
	MainTaskId      string           `json:"mainTaskId"`
	Assets          []*Asset         `json:"assets"`
	Vulnerabilities []*Vulnerability `json:"vulnerabilities"`
}

// Asset 资产
type Asset struct {
	Authority  string   `json:"authority"`
	Host       string   `json:"host"`
	Port       int      `json:"port"`
	Category   string   `json:"category"` // ipv4/ipv6/domain
	Service    string   `json:"service"`
	Server     string   `json:"server"`
	Banner     string   `json:"banner"`
	Title      string   `json:"title"`
	App        []string `json:"app"`
	HttpStatus string   `json:"httpStatus"`
	HttpHeader string   `json:"httpHeader"`
	HttpBody   string   `json:"httpBody"`
	Cert       string   `json:"cert"`
	IconHash   string   `json:"iconHash"`
	Screenshot string   `json:"screenshot"`
	IsCDN      bool     `json:"isCdn"`
	CName      string   `json:"cname"`
	IsCloud    bool     `json:"isCloud"`
	IsHTTP     bool     `json:"isHttp"` // 是否为HTTP服务
	IPV4       []IPInfo `json:"ipv4"`
	IPV6       []IPInfo `json:"ipv6"`
}

// IPInfo IP信息
type IPInfo struct {
	IP       string `json:"ip"`
	Location string `json:"location"`
}

// Vulnerability 漏洞
type Vulnerability struct {
	Authority string `json:"authority"`
	Host      string `json:"host"`
	Port      int    `json:"port"`
	Url       string `json:"url"`
	PocFile   string `json:"pocFile"`
	Source    string `json:"source"`
	Severity  string `json:"severity"`
	Extra     string `json:"extra"`
	Result    string `json:"result"`

	// 漏洞知识库关联字段
	CvssScore   float64  `json:"cvssScore,omitempty"`
	CveId       string   `json:"cveId,omitempty"`
	CweId       string   `json:"cweId,omitempty"`
	Remediation string   `json:"remediation,omitempty"`
	References  []string `json:"references,omitempty"`

	// 证据链字段
	MatcherName       string   `json:"matcherName,omitempty"`
	ExtractedResults  []string `json:"extractedResults,omitempty"`
	CurlCommand       string   `json:"curlCommand,omitempty"`
	Request           string   `json:"request,omitempty"`
	Response          string   `json:"response,omitempty"`
	ResponseTruncated bool     `json:"responseTruncated,omitempty"`
}

// BaseScanner 基础扫描器
type BaseScanner struct {
	name string
}

func (s *BaseScanner) Name() string {
	return s.name
}
