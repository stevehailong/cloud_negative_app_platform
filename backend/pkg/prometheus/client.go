package prometheus

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Client wraps the Prometheus HTTP API.
type Client struct {
	baseURL string
	client  *http.Client
}

// QueryResponse 是 Prometheus /api/v1/query 接口的响应结构
type QueryResponse struct {
	Status string    `json:"status"`
	Data   QueryData `json:"data"`
	Error  string    `json:"error,omitempty"`
}

// QueryData 包含查询结果
type QueryData struct {
	ResultType string   `json:"resultType"`
	Result     []Result `json:"result"`
}

// Result 表示单个时间序列结果
type Result struct {
	Metric map[string]string `json:"metric"`
	// Value: [timestamp, "value"]
	Value  []interface{}   `json:"value,omitempty"`
	Values [][]interface{} `json:"values,omitempty"`
}

// NewClient 创建一个新的 Prometheus API 客户端
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		client: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// BaseURL 返回基础地址（用于诊断/外部展示）
func (c *Client) BaseURL() string {
	return c.baseURL
}

// Ping 测试 Prometheus 连通性
func (c *Client) Ping() error {
	resp, err := c.client.Get(c.baseURL + "/-/healthy")
	if err != nil {
		// 部分 Prometheus 关闭了 /-/healthy，回退测一个简单查询
		return c.fallbackPing()
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return c.fallbackPing()
	}
	return nil
}

func (c *Client) fallbackPing() error {
	_, err := c.Query("up", time.Time{})
	return err
}

// Query 执行即时查询
// ts 为零值时表示当前时刻
func (c *Client) Query(promQL string, ts time.Time) (*QueryResponse, error) {
	params := url.Values{}
	params.Set("query", promQL)
	if !ts.IsZero() {
		params.Set("time", strconv.FormatInt(ts.Unix(), 10))
	}
	return c.doQuery("/api/v1/query", params)
}

// QueryScalar 执行查询并返回单一标量值；若结果为空返回 (0, false, nil)
func (c *Client) QueryScalar(promQL string) (float64, bool, error) {
	resp, err := c.Query(promQL, time.Time{})
	if err != nil {
		return 0, false, err
	}
	if resp.Status != "success" {
		return 0, false, fmt.Errorf("prometheus query failed: %s", resp.Error)
	}
	if len(resp.Data.Result) == 0 {
		return 0, false, nil
	}
	val, ok := extractValue(resp.Data.Result[0].Value)
	return val, ok, nil
}

// extractValue 解析 Prometheus 返回的 [timestamp, "value"] 元组
func extractValue(v []interface{}) (float64, bool) {
	if len(v) < 2 {
		return 0, false
	}
	switch s := v[1].(type) {
	case string:
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return 0, false
		}
		return f, true
	case float64:
		return s, true
	}
	return 0, false
}

func (c *Client) doQuery(path string, params url.Values) (*QueryResponse, error) {
	if c.baseURL == "" {
		return nil, fmt.Errorf("prometheus base URL is empty")
	}
	u := c.baseURL + path + "?" + params.Encode()
	resp, err := c.client.Get(u)
	if err != nil {
		return nil, fmt.Errorf("prometheus request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read prometheus response: %w", err)
	}

	if resp.StatusCode >= 500 {
		return nil, fmt.Errorf("prometheus server error (status %d): %s", resp.StatusCode, string(body))
	}

	var qr QueryResponse
	if err := json.Unmarshal(body, &qr); err != nil {
		return nil, fmt.Errorf("parse prometheus response: %w (body=%s)", err, string(body))
	}
	if qr.Status != "success" {
		return &qr, fmt.Errorf("prometheus returned error: %s", qr.Error)
	}
	return &qr, nil
}
