package api

import (
	"encoding/json"

	"github.com/gin-gonic/gin"
)

// AuditLogger 审计日志记录器
type AuditLogger struct {
	svc interface {
		LogOperation(userID uint, username, action, resource string, resourceID uint, detail, ip, userAgent, status string)
	}
}

// NewAuditLogger 创建审计日志记录器
func NewAuditLogger(svc interface {
	LogOperation(userID uint, username, action, resource string, resourceID uint, detail, ip, userAgent, status string)
}) *AuditLogger {
	return &AuditLogger{svc: svc}
}

// Log 记录操作日志
func (a *AuditLogger) Log(c *gin.Context, action, resource string, resourceID uint, detail interface{}, status string) {
	userID, _ := c.Get("user_id")
	username, _ := c.Get("username")

	var uid uint
	if userID != nil {
		if id, ok := userID.(float64); ok {
			uid = uint(id)
		}
	}

	var uname string
	if username != nil {
		uname, _ = username.(string)
	}

	var detailStr string
	if detail != nil {
		if s, ok := detail.(string); ok {
			detailStr = s
		} else {
			bytes, _ := json.Marshal(detail)
			detailStr = string(bytes)
		}
	}

	a.svc.LogOperation(
		uid,
		uname,
		action,
		resource,
		resourceID,
		detailStr,
		c.ClientIP(),
		c.GetHeader("User-Agent"),
		status,
	)
}

// LogSuccess 记录成功的操作
func (a *AuditLogger) LogSuccess(c *gin.Context, action, resource string, resourceID uint, detail interface{}) {
	a.Log(c, action, resource, resourceID, detail, "success")
}

// LogFailed 记录失败的操作
func (a *AuditLogger) LogFailed(c *gin.Context, action, resource string, resourceID uint, detail interface{}) {
	a.Log(c, action, resource, resourceID, detail, "failed")
}
