package onlineapi

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// HunterClient Hunter API客户端
type HunterClient struct {
	apiKey string
	client *http.Client
}

// NewHunterClient 创建Hunter客户端
func NewHunterClient(apiKey string) *HunterClient {
	return &HunterClient{
		apiKey: apiKey,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// HunterResponse Hunter响应
type HunterResponse struct {
	Code    int        `json:"code"`
	Message string     `json:"message"`
	Data    HunterData `json:"data"`
}

// HunterData Hunter数据
type HunterData struct {
	AccountType string        `json:"account_type"`
	Total       int           `json:"total"`
	Time        int           `json:"time"`
	Arr         []HunterAsset `json:"arr"`
	ConsumeQuota string       `json:"consume_quota"`
	RestQuota    string       `json:"rest_quota"`
	SyntaxPrompt string       `json:"syntax_prompt"`
}

// HunterAsset Hunter资产
type HunterAsset struct {
	IsRisk         string `json:"is_risk"`
	URL            string `json:"url"`
	IP             string `json:"ip"`
	Port           int    `json:"port"`
	WebTitle       string `json:"web_title"`
	Domain         string `json:"domain"`
	IsRiskProtocol string `json:"is_risk_protocol"`
	Protocol       string `json:"protocol"`
	BaseProtocol   string `json:"base_protocol"`
	StatusCode     int    `json:"status_code"`
	Component      []struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	} `json:"component"`
	OS          string `json:"os"`
	Company     string `json:"company"`
	Number      string `json:"number"`
	Country     string `json:"country"`
	Province    string `json:"province"`
	City        string `json:"city"`
	UpdatedAt   string `json:"updated_at"`
	IsWeb       string `json:"is_web"`
	AsOrg       string `json:"as_org"`
	ISP         string `json:"isp"`
	Banner      string `json:"banner"`
}

// Search 搜索
func (c *HunterClient) Search(ctx context.Context, query string, page, size int, startTime, endTime string) (*HunterResponse, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("hunter api key is empty")
	}

	// Base64编码查询语句
	queryBase64 := base64.URLEncoding.EncodeToString([]byte(query))

	// 构建URL
	apiURL := fmt.Sprintf(
		"https://hunter.qianxin.com/openApi/search?api-key=%s&search=%s&page=%d&page_size=%d&is_web=3",
		url.QueryEscape(c.apiKey),
		url.QueryEscape(queryBase64),
		page,
		size,
	)

	if startTime != "" {
		apiURL += "&start_time=" + url.QueryEscape(startTime)
	}
	if endTime != "" {
		apiURL += "&end_time=" + url.QueryEscape(endTime)
	}

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

	var result HunterResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	if result.Code != 200 {
		return nil, fmt.Errorf("hunter error: %s", result.Message)
	}

	return &result, nil
}

// SearchByIP 按IP搜索
func (c *HunterClient) SearchByIP(ctx context.Context, ip string, page, size int) ([]HunterAsset, error) {
	query := fmt.Sprintf(`ip="%s"`, ip)
	result, err := c.Search(ctx, query, page, size, "", "")
	if err != nil {
		return nil, err
	}
	return result.Data.Arr, nil
}

// SearchByDomain 按域名搜索
func (c *HunterClient) SearchByDomain(ctx context.Context, domain string, page, size int) ([]HunterAsset, error) {
	query := fmt.Sprintf(`domain.suffix="%s"`, domain)
	result, err := c.Search(ctx, query, page, size, "", "")
	if err != nil {
		return nil, err
	}
	return result.Data.Arr, nil
}

// SearchByCompany 按公司搜索
func (c *HunterClient) SearchByCompany(ctx context.Context, company string, page, size int) ([]HunterAsset, error) {
	query := fmt.Sprintf(`icp.name="%s"`, company)
	result, err := c.Search(ctx, query, page, size, "", "")
	if err != nil {
		return nil, err
	}
	return result.Data.Arr, nil
}

// SearchByICPNumber 按ICP备案号搜索
func (c *HunterClient) SearchByICPNumber(ctx context.Context, icpNumber string, page, size int) ([]HunterAsset, error) {
	query := fmt.Sprintf(`icp.number="%s"`, icpNumber)
	result, err := c.Search(ctx, query, page, size, "", "")
	if err != nil {
		return nil, err
	}
	return result.Data.Arr, nil
}
