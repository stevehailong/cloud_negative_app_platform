package model

import "time"

// AuditLog 审计日志模型
type AuditLog struct {
	ID              uint       `gorm:"primaryKey" json:"id"`
	UserID          uint       `gorm:"not null;index" json:"userId"`
	Username        string     `gorm:"size:100;not null;index" json:"username"`
	Action          string     `gorm:"size:50;not null;index" json:"action"` // create/update/delete/view
	ResourceType    string     `gorm:"size:50;not null;index" json:"resourceType"`
	ResourceID      *uint      `gorm:"index" json:"resourceId,omitempty"`
	ResourceName    string     `gorm:"size:255" json:"resourceName,omitempty"`
	Method          string     `gorm:"size:10;not null" json:"method"`
	Path            string     `gorm:"size:500;not null;index:idx_path" json:"path"`
	IPAddress       string     `gorm:"size:50" json:"ipAddress"`
	UserAgent       string     `gorm:"type:text" json:"userAgent"`
	RequestBody     string     `gorm:"type:text" json:"requestBody,omitempty"`
	ResponseCode    int        `json:"responseCode"`
	ResponseMessage string     `gorm:"size:500" json:"responseMessage"`
	DurationMs      int        `json:"durationMs"`
	CreateTime      time.Time  `gorm:"autoCreateTime;index" json:"createTime"`
}

func (AuditLog) TableName() string {
	return "audit_logs"
}
