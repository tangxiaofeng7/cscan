package onlineapi

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// FofaClient Fofa API客户端
type FofaClient struct {
	email  string
	key    string
	client *http.Client
}

// NewFofaClient 创建Fofa客户端
func NewFofaClient(email, key string) *FofaClient {
	return &FofaClient{
		email: email,
		key:   key,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// FofaResult Fofa查询结果
type FofaResult struct {
	Error   bool       `json:"error"`
	ErrMsg  string     `json:"errmsg"`
	Mode    string     `json:"mode"`
	Page    int        `json:"page"`
	Query   string     `json:"query"`
	Size    int        `json:"size"`
	Results [][]string `json:"results"`
}

// FofaAsset Fofa资产
type FofaAsset struct {
	Host       string `json:"host"`
	IP         string `json:"ip"`
	Port       int    `json:"port"`
	Protocol   string `json:"protocol"`
	Domain     string `json:"domain"`
	Title      string `json:"title"`
	Server     string `json:"server"`
	Country    string `json:"country"`
	City       string `json:"city"`
	ASN        string `json:"asn"`
	Banner     string `json:"banner"`
	Cert       string `json:"cert"`
	ICP        string `json:"icp"`
	Product    string `json:"product"`
	OS         string `json:"os"`
}

// Search 搜索
func (c *FofaClient) Search(ctx context.Context, query string, page, size int) (*FofaResult, error) {
	if c.email == "" || c.key == "" {
		return nil, fmt.Errorf("fofa email or key is empty")
	}

	// Base64编码查询语句
	queryBase64 := base64.StdEncoding.EncodeToString([]byte(query))

	// 构建URL
	apiURL := fmt.Sprintf(
		"https://fofa.info/api/v1/search/all?email=%s&key=%s&qbase64=%s&page=%d&size=%d&fields=host,ip,port,protocol,domain,title,server,country,city,as_number,banner,cert,icp,product,os",
		url.QueryEscape(c.email),
		url.QueryEscape(c.key),
		url.QueryEscape(queryBase64),
		page,
		size,
	)

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result FofaResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	if result.Error {
		return nil, fmt.Errorf("fofa error: %s", result.ErrMsg)
	}

	return &result, nil
}

// ParseResults 解析结果
func (c *FofaClient) ParseResults(result *FofaResult) []FofaAsset {
	var assets []FofaAsset

	for _, row := range result.Results {
		if len(row) < 15 {
			continue
		}

		port, _ := strconv.Atoi(row[2])
		asset := FofaAsset{
			Host:     row[0],
			IP:       row[1],
			Port:     port,
			Protocol: row[3],
			Domain:   row[4],
			Title:    row[5],
			Server:   row[6],
			Country:  row[7],
			City:     row[8],
			ASN:      row[9],
			Banner:   row[10],
			Cert:     row[11],
			ICP:      row[12],
			Product:  row[13],
			OS:       row[14],
		}
		assets = append(assets, asset)
	}

	return assets
}

// SearchByIP 按IP搜索
func (c *FofaClient) SearchByIP(ctx context.Context, ip string, page, size int) ([]FofaAsset, error) {
	query := fmt.Sprintf(`ip="%s"`, ip)
	result, err := c.Search(ctx, query, page, size)
	if err != nil {
		return nil, err
	}
	return c.ParseResults(result), nil
}

// SearchByDomain 按域名搜索
func (c *FofaClient) SearchByDomain(ctx context.Context, domain string, page, size int) ([]FofaAsset, error) {
	query := fmt.Sprintf(`domain="%s"`, domain)
	result, err := c.Search(ctx, query, page, size)
	if err != nil {
		return nil, err
	}
	return c.ParseResults(result), nil
}

// SearchByTitle 按标题搜索
func (c *FofaClient) SearchByTitle(ctx context.Context, title string, page, size int) ([]FofaAsset, error) {
	query := fmt.Sprintf(`title="%s"`, title)
	result, err := c.Search(ctx, query, page, size)
	if err != nil {
		return nil, err
	}
	return c.ParseResults(result), nil
}

// SearchByOrg 按组织搜索
func (c *FofaClient) SearchByOrg(ctx context.Context, org string, page, size int) ([]FofaAsset, error) {
	query := fmt.Sprintf(`org="%s"`, org)
	result, err := c.Search(ctx, query, page, size)
	if err != nil {
		return nil, err
	}
	return c.ParseResults(result), nil
}

// SearchByICPBeian 按ICP备案搜索
func (c *FofaClient) SearchByICPBeian(ctx context.Context, icp string, page, size int) ([]FofaAsset, error) {
	query := fmt.Sprintf(`icp="%s"`, icp)
	result, err := c.Search(ctx, query, page, size)
	if err != nil {
		return nil, err
	}
	return c.ParseResults(result), nil
}

// BuildQuery 构建查询语句
func BuildFofaQuery(conditions map[string]string) string {
	var parts []string
	for key, value := range conditions {
		if value != "" {
			parts = append(parts, fmt.Sprintf(`%s="%s"`, key, value))
		}
	}
	return strings.Join(parts, " && ")
}
