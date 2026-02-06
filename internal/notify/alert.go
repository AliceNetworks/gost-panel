package notify

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/AliceNetworks/gost-panel/internal/model"
	"gorm.io/gorm"
)

// AlertService 告警服务
type AlertService struct {
	db *gorm.DB
}

func NewAlertService(db *gorm.DB) *AlertService {
	return &AlertService{db: db}
}

// CheckNodeQuota 检查节点流量配额
func (a *AlertService) CheckNodeQuota(node *model.Node) {
	if node.TrafficQuota <= 0 {
		return // 无限制
	}

	totalUsed := node.QuotaUsed
	usagePercent := float64(totalUsed) / float64(node.TrafficQuota) * 100

	// 检查预警阈值 (80%, 90%)
	a.checkQuotaWarning(node.ID, "node", node.Name, totalUsed, node.TrafficQuota, usagePercent)

	if totalUsed >= node.TrafficQuota && !node.QuotaExceeded {
		// 标记超限
		a.db.Model(node).Update("quota_exceeded", true)
		// 触发告警
		a.TriggerAlert("quota_exceeded", "node", node.ID, node.Name,
			fmt.Sprintf("节点 %s 流量已超限\n已用: %s / 配额: %s",
				node.Name,
				formatBytes(totalUsed),
				formatBytes(node.TrafficQuota)))
	}
}

// checkQuotaWarning 检查配额预警 (80%, 90%)
func (a *AlertService) checkQuotaWarning(targetID uint, targetType, targetName string, used, quota int64, percent float64) {
	// 查找预警规则
	var rules []model.AlertRule
	a.db.Where("type = ? AND enabled = ?", "quota_warning", true).Find(&rules)

	for _, rule := range rules {
		// 解析条件获取阈值
		condition, err := ParseCondition(rule.Condition)
		if err != nil {
			continue
		}

		threshold := condition.Threshold
		if threshold <= 0 {
			threshold = 80 // 默认 80%
		}

		// 检查是否达到阈值
		if percent < float64(threshold) {
			continue
		}

		// 检查冷却时间
		if time.Since(rule.LastAlertAt) < time.Duration(rule.CooldownMin)*time.Minute {
			continue
		}

		// 检查是否已经发送过该阈值的预警 (避免重复告警)
		warningKey := fmt.Sprintf("quota_warning_%s_%d_%d", targetType, targetID, threshold)
		if a.hasRecentWarning(warningKey, 24*time.Hour) {
			continue
		}

		// 发送预警
		message := fmt.Sprintf("%s %s 流量使用已达 %.1f%%\n已用: %s / 配额: %s\n请及时关注流量使用情况",
			targetTypeToName(targetType),
			targetName,
			percent,
			formatBytes(used),
			formatBytes(quota))

		a.TriggerAlertWithKey("quota_warning", targetType, targetID, targetName, message, warningKey)
	}
}

// hasRecentWarning 检查是否有最近的预警记录
func (a *AlertService) hasRecentWarning(key string, duration time.Duration) bool {
	var count int64
	threshold := time.Now().Add(-duration)
	a.db.Model(&model.AlertLog{}).
		Where("type = ? AND message LIKE ? AND created_at > ?", "quota_warning", "%"+key+"%", threshold).
		Count(&count)
	return count > 0
}

// TriggerAlertWithKey 触发带标识的告警
func (a *AlertService) TriggerAlertWithKey(alertType, targetType string, targetID uint, targetName, message, key string) {
	// 在消息中添加隐藏标识用于去重
	messageWithKey := message + "\n<!-- " + key + " -->"
	a.TriggerAlert(alertType, targetType, targetID, targetName, messageWithKey)
}

func targetTypeToName(targetType string) string {
	switch targetType {
	case "node":
		return "节点"
	case "client":
		return "客户端"
	default:
		return targetType
	}
}

// CheckClientQuota 检查客户端流量配额
func (a *AlertService) CheckClientQuota(client *model.Client) {
	if client.TrafficQuota <= 0 {
		return
	}

	totalUsed := client.QuotaUsed
	usagePercent := float64(totalUsed) / float64(client.TrafficQuota) * 100

	// 检查预警阈值
	a.checkQuotaWarning(client.ID, "client", client.Name, totalUsed, client.TrafficQuota, usagePercent)

	if totalUsed >= client.TrafficQuota && !client.QuotaExceeded {
		a.db.Model(client).Update("quota_exceeded", true)
		a.TriggerAlert("quota_exceeded", "client", client.ID, client.Name,
			fmt.Sprintf("客户端 %s 流量已超限\n已用: %s / 配额: %s",
				client.Name,
				formatBytes(totalUsed),
				formatBytes(client.TrafficQuota)))
	}
}

// CheckNodeOffline 检查节点离线
func (a *AlertService) CheckNodeOffline(node *model.Node, previousStatus string) {
	if previousStatus == "online" && node.Status == "offline" {
		a.TriggerAlert("node_offline", "node", node.ID, node.Name,
			fmt.Sprintf("节点 %s 已离线\n最后在线: %s",
				node.Name,
				node.LastSeen.Format("2006-01-02 15:04:05")))
	}
}

// TriggerAlert 触发告警
func (a *AlertService) TriggerAlert(alertType, targetType string, targetID uint, targetName, message string) {
	// 查找匹配的告警规则
	var rules []model.AlertRule
	a.db.Where("type = ? AND enabled = ?", alertType, true).Find(&rules)

	for _, rule := range rules {
		// 检查冷却时间
		if time.Since(rule.LastAlertAt) < time.Duration(rule.CooldownMin)*time.Minute {
			continue
		}

		// 发送通知
		channelIDs := strings.Split(rule.ChannelIDs, ",")
		for _, idStr := range channelIDs {
			idStr = strings.TrimSpace(idStr)
			if idStr == "" {
				continue
			}

			channelID, err := strconv.ParseUint(idStr, 10, 32)
			if err != nil {
				continue
			}

			var channel model.NotifyChannel
			if err := a.db.First(&channel, channelID).Error; err != nil {
				continue
			}

			if !channel.Enabled {
				continue
			}

			notifier, err := CreateNotifier(&channel)
			if err != nil {
				log.Printf("Create notifier failed: %v", err)
				continue
			}

			status := "sent"
			title := fmt.Sprintf("[%s] %s", alertTypeToTitle(alertType), targetName)
			if err := notifier.Send(title, message); err != nil {
				log.Printf("Send notification failed: %v", err)
				status = "failed"
			}

			// 记录告警日志
			a.db.Create(&model.AlertLog{
				RuleID:     rule.ID,
				RuleName:   rule.Name,
				Type:       alertType,
				Message:    message,
				TargetType: targetType,
				TargetID:   targetID,
				TargetName: targetName,
				Status:     status,
				CreatedAt:  time.Now(),
			})
		}

		// 更新规则的最后告警时间
		a.db.Model(&rule).Update("last_alert_at", time.Now())
	}
}

// ResetQuotas 重置流量配额（每天检查一次）
func (a *AlertService) ResetQuotas() {
	today := time.Now().Day()

	// 重置节点配额
	var nodes []model.Node
	a.db.Where("quota_reset_day = ? AND (quota_reset_at IS NULL OR quota_reset_at < ?)",
		today, time.Now().AddDate(0, 0, -28)).Find(&nodes)

	for _, node := range nodes {
		a.db.Model(&node).Updates(map[string]interface{}{
			"quota_used":     0,
			"quota_exceeded": false,
			"quota_reset_at": time.Now(),
		})
	}

	// 重置客户端配额
	var clients []model.Client
	a.db.Where("quota_reset_day = ? AND (quota_reset_at IS NULL OR quota_reset_at < ?)",
		today, time.Now().AddDate(0, 0, -28)).Find(&clients)

	for _, client := range clients {
		a.db.Model(&client).Updates(map[string]interface{}{
			"quota_used":     0,
			"quota_exceeded": false,
			"quota_reset_at": time.Now(),
		})
	}
}

// CheckOfflineNodes 检查离线节点（心跳超时）
func (a *AlertService) CheckOfflineNodes(timeoutMinutes int) {
	threshold := time.Now().Add(-time.Duration(timeoutMinutes) * time.Minute)

	var nodes []model.Node
	a.db.Where("status = ? AND last_seen < ?", "online", threshold).Find(&nodes)

	for _, node := range nodes {
		a.db.Model(&node).Update("status", "offline")
		a.TriggerAlert("node_offline", "node", node.ID, node.Name,
			fmt.Sprintf("节点 %s 心跳超时，已标记为离线\n最后心跳: %s",
				node.Name,
				node.LastSeen.Format("2006-01-02 15:04:05")))
	}
}

// CleanupAlertLogs 清理旧的告警日志
func (a *AlertService) CleanupAlertLogs(retentionDays int) {
	threshold := time.Now().AddDate(0, 0, -retentionDays)
	a.db.Where("created_at < ?", threshold).Delete(&model.AlertLog{})
}

// GetAlertLogs 获取告警日志
func (a *AlertService) GetAlertLogs(limit, offset int) ([]model.AlertLog, int64, error) {
	var logs []model.AlertLog
	var total int64

	a.db.Model(&model.AlertLog{}).Count(&total)
	err := a.db.Order("created_at desc").Limit(limit).Offset(offset).Find(&logs).Error

	return logs, total, err
}

// ==================== 辅助函数 ====================

func alertTypeToTitle(alertType string) string {
	switch alertType {
	case "node_offline":
		return "节点离线"
	case "quota_exceeded":
		return "流量超限"
	case "quota_warning":
		return "流量预警"
	case "traffic_spike":
		return "流量异常"
	case "agent_update":
		return "Agent 更新"
	default:
		return "告警"
	}
}

func formatBytes(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
		TB = GB * 1024
	)

	switch {
	case bytes >= TB:
		return fmt.Sprintf("%.2f TB", float64(bytes)/TB)
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/GB)
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/MB)
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

// ==================== 通知渠道管理 ====================

// ListChannels 获取通知渠道列表
func (a *AlertService) ListChannels() ([]model.NotifyChannel, error) {
	var channels []model.NotifyChannel
	err := a.db.Order("id asc").Find(&channels).Error
	return channels, err
}

// GetChannel 获取单个通知渠道
func (a *AlertService) GetChannel(id uint) (*model.NotifyChannel, error) {
	var channel model.NotifyChannel
	err := a.db.First(&channel, id).Error
	return &channel, err
}

// CreateChannel 创建通知渠道
func (a *AlertService) CreateChannel(channel *model.NotifyChannel) error {
	channel.CreatedAt = time.Now()
	channel.UpdatedAt = time.Now()
	return a.db.Create(channel).Error
}

// UpdateChannel 更新通知渠道
func (a *AlertService) UpdateChannel(id uint, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()
	return a.db.Model(&model.NotifyChannel{}).Where("id = ?", id).Updates(updates).Error
}

// DeleteChannel 删除通知渠道
func (a *AlertService) DeleteChannel(id uint) error {
	return a.db.Delete(&model.NotifyChannel{}, id).Error
}

// TestChannel 测试通知渠道
func (a *AlertService) TestChannel(id uint) error {
	channel, err := a.GetChannel(id)
	if err != nil {
		return err
	}

	notifier, err := CreateNotifier(channel)
	if err != nil {
		return err
	}

	return notifier.Send("测试通知", "这是一条来自 GOST Panel 的测试通知消息。\n如果您收到此消息，说明通知渠道配置正确。")
}

// ==================== 告警规则管理 ====================

// ListRules 获取告警规则列表
func (a *AlertService) ListRules() ([]model.AlertRule, error) {
	var rules []model.AlertRule
	err := a.db.Order("id asc").Find(&rules).Error
	return rules, err
}

// GetRule 获取单个告警规则
func (a *AlertService) GetRule(id uint) (*model.AlertRule, error) {
	var rule model.AlertRule
	err := a.db.First(&rule, id).Error
	return &rule, err
}

// CreateRule 创建告警规则
func (a *AlertService) CreateRule(rule *model.AlertRule) error {
	rule.CreatedAt = time.Now()
	rule.UpdatedAt = time.Now()
	return a.db.Create(rule).Error
}

// UpdateRule 更新告警规则
func (a *AlertService) UpdateRule(id uint, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()
	return a.db.Model(&model.AlertRule{}).Where("id = ?", id).Updates(updates).Error
}

// DeleteRule 删除告警规则
func (a *AlertService) DeleteRule(id uint) error {
	return a.db.Delete(&model.AlertRule{}, id).Error
}

// CreateDefaultRules 创建默认告警规则
func (a *AlertService) CreateDefaultRules() error {
	var count int64
	a.db.Model(&model.AlertRule{}).Count(&count)
	if count > 0 {
		return nil
	}

	rules := []model.AlertRule{
		{
			Name:        "节点离线告警",
			Type:        "node_offline",
			Condition:   "{}",
			Enabled:     true,
			CooldownMin: 30,
		},
		{
			Name:        "流量超限告警",
			Type:        "quota_exceeded",
			Condition:   "{}",
			Enabled:     true,
			CooldownMin: 60,
		},
		{
			Name:        "流量预警 (80%)",
			Type:        "quota_warning",
			Condition:   "{\"threshold\": 80}",
			Enabled:     true,
			CooldownMin: 60,
		},
		{
			Name:        "流量预警 (90%)",
			Type:        "quota_warning",
			Condition:   "{\"threshold\": 90}",
			Enabled:     true,
			CooldownMin: 30,
		},
	}

	for _, rule := range rules {
		rule.CreatedAt = time.Now()
		rule.UpdatedAt = time.Now()
		if err := a.db.Create(&rule).Error; err != nil {
			return err
		}
	}

	return nil
}

// AlertRuleCondition 告警条件
type AlertRuleCondition struct {
	Threshold int64 `json:"threshold"` // 阈值
	Duration  int   `json:"duration"`  // 持续时间（分钟）
}

// ParseCondition 解析告警条件
func ParseCondition(conditionJSON string) (*AlertRuleCondition, error) {
	var condition AlertRuleCondition
	if conditionJSON == "" || conditionJSON == "{}" {
		return &AlertRuleCondition{}, nil
	}
	err := json.Unmarshal([]byte(conditionJSON), &condition)
	return &condition, err
}
