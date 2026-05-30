package proxy

import (
	"bytes"
	"io"
	"my-cloud/internal/common/response"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type ServiceProxy struct {
	targetURL string
	client    *http.Client
}

func NewServiceProxy(targetURL string) *ServiceProxy {
	return &ServiceProxy{
		targetURL: targetURL,
		client:    &http.Client{},
	}
}

func (p *ServiceProxy) Handle(c *gin.Context) {
	// 获取完整路径并构建目标URL
	// 例如: /api/v1/auth/login -> path param = /login
	// 需要保留 /auth 前缀，所以从原始URL中提取
	fullPath := c.Request.URL.Path
	targetURL := p.targetURL + fullPath
	
	// 添加查询参数
	if c.Request.URL.RawQuery != "" {
		targetURL += "?" + c.Request.URL.RawQuery
	}

	// 读取请求体
	var bodyBytes []byte
	if c.Request.Body != nil {
		bodyBytes, _ = io.ReadAll(c.Request.Body)
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	}

	// 创建新请求
	req, err := http.NewRequest(c.Request.Method, targetURL, bytes.NewReader(bodyBytes))
	if err != nil {
		response.InternalError(c, "failed to create request")
		return
	}

	// 复制请求头
	for key, values := range c.Request.Header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	// 发送请求
	resp, err := p.client.Do(req)
	if err != nil {
		response.InternalError(c, "failed to proxy request: "+err.Error())
		return
	}
	defer resp.Body.Close()

	// 复制响应头
	for key, values := range resp.Header {
		for _, value := range values {
			c.Writer.Header().Add(key, value)
		}
	}

	// 读取响应体
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		response.InternalError(c, "failed to read response")
		return
	}

	// 返回响应
	c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), respBody)
}

// ProxyWithPath 带路径前缀的代理
func (p *ServiceProxy) ProxyWithPath(pathPrefix string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 去除路径前缀
		originalPath := c.Request.URL.Path
		c.Request.URL.Path = strings.TrimPrefix(originalPath, pathPrefix)
		
		p.Handle(c)
	}
}
