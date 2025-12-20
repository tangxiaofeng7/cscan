package onlineapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// QuakeClient Quake API客户端
type QuakeClient struct {
	apiKey string
	client *http.Client
}

// NewQuakeClient 创建Quake客户端
func NewQuakeClient(apiKey string) *QuakeClient {
	return &QuakeClient{
		apiKey: apiKey,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// QuakeResponse Quake响应
type QuakeResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    []QuakeData `json:"data"`
	Meta    QuakeMeta   `json:"meta"`
}

// QuakeMeta Quake元数据
type QuakeMeta struct {
	Pagination struct {
		Count     int `json:"count"`
		PageIndex int `json:"page_index"`
		PageSize  int `json:"page_size"`
		Total     int `json:"total"`
	} `json:"pagination"`
}

// QuakeData Quake数据
type QuakeData struct {
	IP       string `json:"ip"`
	Port     int    `json:"port"`
	Hostname string `json:"hostname"`
	Service  struct {
		Name     string `json:"name"`
		Product  string `json:"product"`
		Version  string `json:"version"`
		Response string `json:"response"`
		Cert     string `json:"cert"`
		HTTP     struct {
			Title      string `json:"title"`
			StatusCode int    `json:"status_code"`
			Server     string `json:"server"`
			Host       string `json:"host"`
			Path       string `json:"path"`
		} `json:"http"`
	} `json:"service"`
	Location struct {
		CountryCode string  `json:"country_code"`
		CountryCN   string  `json:"country_cn"`
		CountryEN   string  `json:"country_en"`
		ProvinceCN  string  `json:"province_cn"`
		ProvinceEN  string  `json:"province_en"`
		CityCN      string  `json:"city_cn"`
		CityEN      string  `json:"city_en"`
		DistrictCN  string  `json:"district_cn"`
		DistrictEN  string  `json:"district_en"`
		ISP         string  `json:"isp"`
		Lat         float64 `json:"lat"`
		Lon         float64 `json:"lon"`
	} `json:"location"`
	ASN struct {
		Number int    `json:"number"`
		Org    string `json:"org"`
	} `json:"asn"`
	Time      string   `json:"time"`
	Transport string   `json:"transport"`
	Components []struct {
		ProductLevel    []string `json:"product_level"`
		ProductType     []string `json:"product_type"`
		ProductVendor   []string `json:"product_vendor"`
		ProductNameCN   string   `json:"product_name_cn"`
		ProductNameEN   string   `json:"product_name_en"`
		Version         string   `json:"version"`
	} `json:"components"`
}

// Search 搜索
func (c *QuakeClient) Search(ctx context.Context, query string, page, size int) (*QuakeResponse, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("quake api key is empty")
	}

	// 构建请求体
	reqBody := map[string]interface{}{
		"query":      query,
		"start":      (page - 1) * size,
		"size":       size,
		"ignore_cache": false,
		"latest":     true,
	}

	data, _ := json.Marshal(reqBody)

	req, err := http.NewRequestWithContext(ctx, "POST", "https://quake.360.net/api/v3/search/quake_service", bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-QuakeToken", c.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result QuakeResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	if result.Code != 0 {
		return nil, fmt.Errorf("quake error: %s", result.Message)
	}

	return &result, nil
}

// SearchByIP 按IP搜索
func (c *QuakeClient) SearchByIP(ctx context.Context, ip string, page, size int) ([]QuakeData, error) {
	query := fmt.Sprintf(`ip:"%s"`, ip)
	result, err := c.Search(ctx, query, page, size)
	if err != nil {
		return nil, err
	}
	return result.Data, nil
}

// SearchByDomain 按域名搜索
func (c *QuakeClient) SearchByDomain(ctx context.Context, domain string, page, size int) ([]QuakeData, error) {
	query := fmt.Sprintf(`domain:"%s"`, domain)
	result, err := c.Search(ctx, query, page, size)
	if err != nil {
		return nil, err
	}
	return result.Data, nil
}

// SearchByTitle 按标题搜索
func (c *QuakeClient) SearchByTitle(ctx context.Context, title string, page, size int) ([]QuakeData, error) {
	query := fmt.Sprintf(`title:"%s"`, title)
	result, err := c.Search(ctx, query, page, size)
	if err != nil {
		return nil, err
	}
	return result.Data, nil
}

// SearchByService 按服务搜索
func (c *QuakeClient) SearchByService(ctx context.Context, service string, page, size int) ([]QuakeData, error) {
	query := fmt.Sprintf(`service:"%s"`, service)
	result, err := c.Search(ctx, query, page, size)
	if err != nil {
		return nil, err
	}
	return result.Data, nil
}
