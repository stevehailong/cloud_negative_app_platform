package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// 响应结构体
type Response struct {
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	RequestID string      `json:"requestId,omitempty"`
}

// 分页响应
type PageResponse struct {
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"pageSize"`
	List     interface{} `json:"list"` // 统一使用list字段
}

// 错误码定义
const (
	CodeSuccess           = 0
	CodeInvalidParams     = 40001
	CodeUnauthorized      = 40101
	CodeForbidden         = 40301
	CodeNotFound          = 40401
	CodeConflict          = 40901
	CodeInternalError     = 50001
	CodeDatabaseError     = 50002
	CodeExternalAPIError  = 50003
)

// 错误消息映射
var codeMessages = map[int]string{
	CodeSuccess:          "success",
	CodeInvalidParams:    "invalid parameters",
	CodeUnauthorized:     "unauthorized",
	CodeForbidden:        "forbidden",
	CodeNotFound:         "resource not found",
	CodeConflict:         "resource conflict",
	CodeInternalError:    "internal server error",
	CodeDatabaseError:    "database error",
	CodeExternalAPIError: "external api error",
}

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	response := Response{
		Code:      CodeSuccess,
		Message:   "success",
		Data:      data,
		RequestID: getRequestID(c),
	}
	// 使用 PureJSON 避免 HTML 转义，保留中文
	c.PureJSON(http.StatusOK, response)
}

// SuccessWithPage 成功响应（分页）
func SuccessWithPage(c *gin.Context, total int64, page, pageSize int, items interface{}) {
	pageData := PageResponse{
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		List:     items, // 使用List字段
	}
	Success(c, pageData)
}

// Error 错误响应
func Error(c *gin.Context, code int, message string) {
	if message == "" {
		message = codeMessages[code]
	}
	response := Response{
		Code:      code,
		Message:   message,
		RequestID: getRequestID(c),
	}
	
	// 根据错误码确定 HTTP 状态码
	httpStatus := getHTTPStatus(code)
	c.PureJSON(httpStatus, response)
}

// ErrorWithData 带数据的错误响应
func ErrorWithData(c *gin.Context, code int, message string, data interface{}) {
	if message == "" {
		message = codeMessages[code]
	}
	response := Response{
		Code:      code,
		Message:   message,
		Data:      data,
		RequestID: getRequestID(c),
	}
	
	httpStatus := getHTTPStatus(code)
	c.PureJSON(httpStatus, response)
}

// 便捷方法
func InvalidParams(c *gin.Context, message string) {
	Error(c, CodeInvalidParams, message)
}

func Unauthorized(c *gin.Context, message string) {
	Error(c, CodeUnauthorized, message)
}

func Forbidden(c *gin.Context, message string) {
	Error(c, CodeForbidden, message)
}

func NotFound(c *gin.Context, message string) {
	Error(c, CodeNotFound, message)
}

func Conflict(c *gin.Context, message string) {
	Error(c, CodeConflict, message)
}

func InternalError(c *gin.Context, message string) {
	Error(c, CodeInternalError, message)
}

func DatabaseError(c *gin.Context, message string) {
	Error(c, CodeDatabaseError, message)
}

// getHTTPStatus 根据业务错误码返回 HTTP 状态码
func getHTTPStatus(code int) int {
	switch {
	case code == CodeSuccess:
		return http.StatusOK
	case code >= 40000 && code < 40100:
		return http.StatusBadRequest
	case code >= 40100 && code < 40200:
		return http.StatusUnauthorized
	case code >= 40300 && code < 40400:
		return http.StatusForbidden
	case code >= 40400 && code < 40500:
		return http.StatusNotFound
	case code >= 40900 && code < 41000:
		return http.StatusConflict
	case code >= 50000:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

// getRequestID 从上下文获取请求ID
func getRequestID(c *gin.Context) string {
	if requestID, exists := c.Get("requestId"); exists {
		return requestID.(string)
	}
	return ""
}
