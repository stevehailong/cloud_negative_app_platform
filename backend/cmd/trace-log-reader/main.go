package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/google/uuid"
)

// NginxLogEntry nginx JSON 日志格式
type NginxLogEntry struct {
	Time        string  `json:"time"`
	Host        string  `json:"host"`
	Method      string  `json:"method"`
	Path        string  `json:"path"`
	Status      int     `json:"status"`
	Duration    float64 `json:"duration"`
	RequestID   string  `json:"request_id"`
	RemoteAddr  string  `json:"remote_addr"`
	UserAgent   string  `json:"user_agent"`
}

// SpanReport 上报的 Span 结构
type SpanReport struct {
	TraceID       string `json:"traceId"`
	SpanID        string `json:"spanId"`
	ParentSpanID  string `json:"parentSpanId,omitempty"`
	ServiceName   string `json:"serviceName"`
	OperationName string `json:"operationName"`
	Method        string `json:"method"`
	DurationMs    uint32 `json:"durationMs"`
	StartTime     string `json:"startTime"`
	EndTime       string `json:"endTime"`
	StatusCode    int    `json:"statusCode"`
	HasError      int    `json:"hasError"`
}

func main() {
	logPath := os.Getenv("NGINX_LOG_PATH")
	if logPath == "" {
		logPath = "/var/log/nginx/access_trace.log"
	}
	monitorURL := os.Getenv("MONITOR_SERVICE_URL")
	if monitorURL == "" {
		monitorURL = "http://monitor-service:8090"
	}
	serviceName := os.Getenv("TRACE_SERVICE_NAME")
	if serviceName == "" {
		serviceName = "frontend-nginx"
	}

	log.Printf("Trace log reader started, watching: %s, monitor: %s", logPath, monitorURL)

	// 等待日志文件创建
	waitForFile(logPath)

	file, err := os.Open(logPath)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	defer file.Close()

	// 跳到文件末尾
	file.Seek(0, 2)

	reader := bufio.NewReader(file)
	client := &http.Client{Timeout: 3 * time.Second}

	// 优雅退出
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				time.Sleep(100 * time.Millisecond)
				// 文件被轮转，重新打开
				if newFile, err2 := os.Open(logPath); err2 == nil {
					file.Close()
					file = newFile
					file.Seek(0, 2)
					reader = bufio.NewReader(file)
				}
				continue
			}

			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			var entry NginxLogEntry
			if err := json.Unmarshal([]byte(line), &entry); err != nil {
				log.Printf("Failed to parse log entry: %v", err)
				continue
			}

			// 跳过非应用请求
			if entry.Host == "localhost" || entry.Host == "" {
				continue
			}

			// 报告 Span
			reportSpan(client, monitorURL, entry, serviceName)
		}
	}()

	<-sigCh
	log.Println("Trace log reader stopped")
}

func waitForFile(path string) {
	for i := 0; i < 30; i++ {
		if _, err := os.Stat(path); err == nil {
			return
		}
		time.Sleep(1 * time.Second)
	}
	log.Printf("Warning: log file %s not found after 30s, will retry", path)
}

func reportSpan(client *http.Client, monitorURL string, entry NginxLogEntry, serviceName string) {
	startTime, _ := time.Parse(time.RFC3339, entry.Time)
	if startTime.IsZero() {
		startTime = time.Now().Add(-time.Duration(entry.Duration) * time.Second)
	}
	endTime := startTime.Add(time.Duration(entry.Duration) * time.Second)

	traceID := entry.RequestID
	if traceID == "" {
		traceID = uuid.New().String()
	}

	span := SpanReport{
		TraceID:       traceID,
		SpanID:        uuid.New().String(),
		ServiceName:   extractServiceName(entry.Host),
		OperationName: entry.Path,
		Method:        entry.Method,
		DurationMs:    uint32(entry.Duration * 1000),
		StartTime:     startTime.Format(time.RFC3339Nano),
		EndTime:       endTime.Format(time.RFC3339Nano),
		StatusCode:    entry.Status,
		HasError:      boolToInt(entry.Status >= 500),
	}

	body, _ := json.Marshal(span)
	req, err := http.NewRequest("POST", monitorURL+"/internal/v1/traces/spans", bytes.NewReader(body))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return
	}
	resp.Body.Close()
}

func extractServiceName(host string) string {
	// 去掉端口
	if idx := strings.Index(host, ":"); idx != -1 {
		host = host[:idx]
	}
	// 去掉 .local 后缀
	host = strings.TrimSuffix(host, ".local")
	return host
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// 每 10 秒输出一次统计
func init() {
	go func() {
		for {
			time.Sleep(10 * time.Second)
			// 心跳日志
			fmt.Fprintf(os.Stderr, "trace-log-reader: alive\n")
		}
	}()
}