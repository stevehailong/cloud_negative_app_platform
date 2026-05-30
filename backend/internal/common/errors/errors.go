package errors

// 错误码定义
const (
	// 成功
	CodeSuccess = 0

	// 通用错误 (100xx)
	CodeBadRequest     = 10001
	CodeUnauthorized   = 10002
	CodeForbidden      = 10003
	CodeNotFound       = 10004
	CodeInternalError  = 10005
	CodeValidationFail = 10006

	// 认证错误 (401xx)
	CodeAuthFailed          = 40101
	CodeTokenInvalid        = 40102
	CodeTokenExpired        = 40103
	CodePermissionDenied    = 40104
	CodeUserNotFound        = 40105
	CodeUserDisabled        = 40106
	CodePasswordIncorrect   = 40107
	CodeUserAlreadyExists   = 40108

	// 业务错误 (500xx)
	CodeTenantNotFound    = 50001
	CodeProjectNotFound   = 50002
	CodeApplicationNotFound = 50003
	CodeDuplicateCode     = 50004
)

// 错误信息映射
var errorMessages = map[int]string{
	CodeSuccess:            "操作成功",
	CodeBadRequest:         "请求参数错误",
	CodeUnauthorized:       "未授权，请先登录",
	CodeForbidden:          "无权限访问",
	CodeNotFound:           "资源不存在",
	CodeInternalError:      "服务器内部错误",
	CodeValidationFail:     "数据验证失败",
	
	CodeAuthFailed:         "认证失败",
	CodeTokenInvalid:       "Token无效",
	CodeTokenExpired:       "Token已过期，请重新登录",
	CodePermissionDenied:   "您没有权限执行此操作",
	CodeUserNotFound:       "用户不存在",
	CodeUserDisabled:       "用户已被禁用",
	CodePasswordIncorrect:  "用户名或密码错误",
	CodeUserAlreadyExists:  "用户名已存在",
	
	CodeTenantNotFound:     "租户不存在",
	CodeProjectNotFound:    "项目不存在",
	CodeApplicationNotFound: "应用不存在",
	CodeDuplicateCode:      "编码已存在，请使用其他编码",
}

// GetMessage 获取错误信息
func GetMessage(code int) string {
	if msg, ok := errorMessages[code]; ok {
		return msg
	}
	return "未知错误"
}

// Error 自定义错误类型
type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *Error) Error() string {
	return e.Message
}

// New 创建新错误
func New(code int, message ...string) *Error {
	msg := GetMessage(code)
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}
	return &Error{
		Code:    code,
		Message: msg,
	}
}

// NewWithDetail 创建带详细信息的错误
func NewWithDetail(code int, detail string) *Error {
	return &Error{
		Code:    code,
		Message: GetMessage(code) + ": " + detail,
	}
}
