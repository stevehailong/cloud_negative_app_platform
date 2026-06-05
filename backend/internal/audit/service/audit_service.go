package service

import (
	"errors"
	"my-cloud/internal/audit/repository"
	"my-cloud/internal/common/model"
	"time"
)

type AuditService struct {
	auditRepo *repository.AuditRepository
}

func NewAuditService(auditRepo *repository.AuditRepository) *AuditService {
	return &AuditService{
		auditRepo: auditRepo,
	}
}

// CreateAuditLog 创建审计日志
func (s *AuditService) CreateAuditLog(log *model.AuditLog) error {
	if log == nil {
		return errors.New("audit log is nil")
	}
	return s.auditRepo.Create(log)
}

// ListAuditLogs 获取审计日志列表
func (s *AuditService) ListAuditLogs(filters map[string]interface{}, page, pageSize int) ([]*model.AuditLog, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	return s.auditRepo.List(filters, page, pageSize)
}

// GetAuditLog 获取审计日志详情
func (s *AuditService) GetAuditLog(id uint) (*model.AuditLog, error) {
	if id == 0 {
		return nil, errors.New("无效的审计日志ID")
	}

	return s.auditRepo.GetByID(id)
}

// GetAuditLogsByResourceID 根据资源ID获取审计日志
func (s *AuditService) GetAuditLogsByResourceID(resourceType string, resourceID uint, page, pageSize int) ([]*model.AuditLog, int64, error) {
	if resourceType == "" || resourceID == 0 {
		return nil, 0, errors.New("资源类型和资源ID不能为空")
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	return s.auditRepo.GetByResourceID(resourceType, resourceID, page, pageSize)
}

// GetAuditLogsByUserID 根据用户ID获取审计日志
func (s *AuditService) GetAuditLogsByUserID(userID uint, page, pageSize int) ([]*model.AuditLog, int64, error) {
	if userID == 0 {
		return nil, 0, errors.New("用户ID不能为空")
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	return s.auditRepo.GetByUserID(userID, page, pageSize)
}

// GetStatistics 获取审计日志统计信息
func (s *AuditService) GetStatistics(startTimeStr, endTimeStr string) (map[string]interface{}, error) {
	// 解析时间参数
	var startTime, endTime time.Time
	var err error

	if startTimeStr != "" {
		startTime, err = time.Parse("2006-01-02 15:04:05", startTimeStr)
		if err != nil {
			startTime, err = time.Parse("2006-01-02", startTimeStr)
			if err != nil {
				return nil, errors.New("开始时间格式错误，应为：2006-01-02 15:04:05 或 2006-01-02")
			}
		}
	} else {
		// 默认最近7天
		startTime = time.Now().AddDate(0, 0, -7)
	}

	if endTimeStr != "" {
		endTime, err = time.Parse("2006-01-02 15:04:05", endTimeStr)
		if err != nil {
			endTime, err = time.Parse("2006-01-02", endTimeStr)
			if err != nil {
				return nil, errors.New("结束时间格式错误，应为：2006-01-02 15:04:05 或 2006-01-02")
			}
			// 如果只提供日期，设置为当天23:59:59
			endTime = endTime.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
		}
	} else {
		// 默认到现在
		endTime = time.Now()
	}

	if startTime.After(endTime) {
		return nil, errors.New("开始时间不能晚于结束时间")
	}

	stats, err := s.auditRepo.GetStatistics(startTime, endTime)
	if err != nil {
		return nil, err
	}

	// 添加时间范围到统计结果
	stats["start_time"] = startTime.Format("2006-01-02 15:04:05")
	stats["end_time"] = endTime.Format("2006-01-02 15:04:05")

	return stats, nil
}

// CleanOldLogs 清理过期日志
func (s *AuditService) CleanOldLogs(retentionDays int) (int64, error) {
	if retentionDays < 1 {
		return 0, errors.New("保留天数必须大于0")
	}

	beforeTime := time.Now().AddDate(0, 0, -retentionDays)
	return s.auditRepo.DeleteOldLogs(beforeTime)
}

// ExportAuditLogs 导出审计日志(返回CSV格式)
func (s *AuditService) ExportAuditLogs(filters map[string]interface{}) (string, error) {
	// 不分页，获取所有符合条件的日志(限制最多10000条)
	logs, _, err := s.auditRepo.List(filters, 1, 10000)
	if err != nil {
		return "", err
	}

	// 构建CSV内容
	csv := "ID,用户ID,用户名,操作类型,资源类型,资源ID,请求方法,请求路径,IP地址,响应码,耗时(ms),创建时间\n"
	for _, log := range logs {
		resourceID := ""
		if log.ResourceID != nil {
			resourceID = string(rune(*log.ResourceID + '0'))
		}

		csv += formatCSVRow([]string{
			string(rune(log.ID + '0')),
			string(rune(log.UserID + '0')),
			log.Username,
			log.Action,
			log.ResourceType,
			resourceID,
			log.Method,
			log.Path,
			log.IPAddress,
			string(rune(log.ResponseCode + '0')),
			string(rune(log.DurationMs + '0')),
			log.CreateTime.Format("2006-01-02 15:04:05"),
		})
	}

	return csv, nil
}

// formatCSVRow 格式化CSV行
func formatCSVRow(fields []string) string {
	row := ""
	for i, field := range fields {
		if i > 0 {
			row += ","
		}
		// 如果字段包含逗号或换行，需要用引号包裹
		if containsSpecialChar(field) {
			row += `"` + field + `"`
		} else {
			row += field
		}
	}
	return row + "\n"
}

// containsSpecialChar 检查是否包含特殊字符
func containsSpecialChar(s string) bool {
	for _, c := range s {
		if c == ',' || c == '"' || c == '\n' || c == '\r' {
			return true
		}
	}
	return false
}
