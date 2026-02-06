package service

import (
	"log"
	"sync"
	"time"

	"github.com/AliceNetworks/gost-panel/internal/gost"
	"github.com/AliceNetworks/gost-panel/internal/model"
	"gorm.io/gorm"
)

// HealthChecker 节点健康检查器
type HealthChecker struct {
	db           *gorm.DB
	alertService interface {
		TriggerAlert(alertType, targetType string, targetID uint, targetName, message string)
	}
	interval time.Duration
	stopCh   chan struct{}
	wg       sync.WaitGroup
}

// NewHealthChecker 创建健康检查器
func NewHealthChecker(db *gorm.DB, alertService interface {
	TriggerAlert(alertType, targetType string, targetID uint, targetName, message string)
}, interval time.Duration) *HealthChecker {
	return &HealthChecker{
		db:           db,
		alertService: alertService,
		interval:     interval,
		stopCh:       make(chan struct{}),
	}
}

// Start 启动健康检查
func (h *HealthChecker) Start() {
	h.wg.Add(1)
	go h.run()
	log.Printf("Health checker started (interval: %v)", h.interval)
}

// Stop 停止健康检查
func (h *HealthChecker) Stop() {
	close(h.stopCh)
	h.wg.Wait()
	log.Println("Health checker stopped")
}

func (h *HealthChecker) run() {
	defer h.wg.Done()

	// 延迟 30 秒再执行首次检测，给 Agent 时间发送心跳
	log.Println("Health checker: waiting 30s before first check...")
	select {
	case <-time.After(30 * time.Second):
		h.checkAll()
	case <-h.stopCh:
		return
	}

	ticker := time.NewTicker(h.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			h.checkAll()
		case <-h.stopCh:
			return
		}
	}
}

func (h *HealthChecker) checkAll() {
	var nodes []model.Node
	if err := h.db.Find(&nodes).Error; err != nil {
		log.Printf("Health check: failed to get nodes: %v", err)
		return
	}

	log.Printf("Health check: checking %d nodes...", len(nodes))

	for _, node := range nodes {
		go h.checkNode(node)
	}

	// 检查客户端超时 (2分钟无心跳则标记离线)
	h.checkClientTimeout()
}

func (h *HealthChecker) checkClientTimeout() {
	timeout := time.Now().Add(-2 * time.Minute)
	h.db.Model(&model.Client{}).
		Where("status = ? AND last_seen < ?", "online", timeout).
		Update("status", "offline")
}

func (h *HealthChecker) checkNode(node model.Node) {
	client := gost.NewClient(node.Host, node.APIPort, node.APIUser, node.APIPass)

	err := client.Ping()
	newStatus := "online"
	if err != nil {
		newStatus = "offline"
	}

	// 状态变更
	if node.Status != newStatus {
		log.Printf("Health check: node %s status changed: %s -> %s", node.Name, node.Status, newStatus)

		// 更新数据库状态
		h.db.Model(&model.Node{}).Where("id = ?", node.ID).Updates(map[string]interface{}{
			"status":    newStatus,
			"last_seen": time.Now(),
		})

		// 触发告警
		if newStatus == "offline" && h.alertService != nil {
			h.alertService.TriggerAlert("node_offline", "node", node.ID, node.Name, "Node is offline")
		}
	} else if newStatus == "online" {
		// 在线时更新 last_seen
		h.db.Model(&model.Node{}).Where("id = ?", node.ID).Update("last_seen", time.Now())
	}
}
