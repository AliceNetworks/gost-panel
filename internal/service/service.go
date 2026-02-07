package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/AliceNetworks/gost-panel/internal/config"
	"github.com/AliceNetworks/gost-panel/internal/gost"
	"github.com/AliceNetworks/gost-panel/internal/model"
	"github.com/AliceNetworks/gost-panel/internal/notify"
	"gorm.io/gorm"
)

type Service struct {
	db            *gorm.DB
	cfg           *config.Config
	alertService  *notify.AlertService
	healthChecker *HealthChecker
}

func NewService(db *gorm.DB, cfg *config.Config) *Service {
	alertSvc := notify.NewAlertService(db)
	// 创建默认告警规则
	alertSvc.CreateDefaultRules()

	svc := &Service{
		db:           db,
		cfg:          cfg,
		alertService: alertSvc,
	}

	// 启动健康检查 (每30秒检查一次)
	svc.healthChecker = NewHealthChecker(db, alertSvc, 30*time.Second)
	svc.healthChecker.Start()

	return svc
}

// GetAlertService 获取告警服务
func (s *Service) GetAlertService() *notify.AlertService {
	return s.alertService
}

// Close 关闭服务
func (s *Service) Close() {
	if s.healthChecker != nil {
		s.healthChecker.Stop()
	}
}

// Ping 检查数据库连接
func (s *Service) Ping() error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}

// PaginationParams 分页参数
type PaginationParams struct {
	Page     int    `form:"page" json:"page"`
	PageSize int    `form:"page_size" json:"page_size"`
	Search   string `form:"search" json:"search"`
	SortBy   string `form:"sort_by" json:"sort_by"`
	SortDesc bool   `form:"sort_desc" json:"sort_desc"`
}

// PaginatedResult 分页结果
type PaginatedResult struct {
	Items    interface{} `json:"items"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
	Pages    int         `json:"pages"`
}

// NewPaginationParams 创建分页参数并设置默认值
func NewPaginationParams(page, pageSize int, search string) PaginationParams {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	return PaginationParams{
		Page:     page,
		PageSize: pageSize,
		Search:   search,
	}
}

// allowedSortFields 定义允许的排序字段白名单 (防止 SQL 注入)
var allowedSortFields = map[string]map[string]bool{
	"nodes": {
		"id": true, "name": true, "host": true, "status": true,
		"traffic_in": true, "traffic_out": true, "connections": true,
		"created_at": true, "updated_at": true, "last_seen": true,
	},
	"clients": {
		"id": true, "name": true, "status": true,
		"traffic_in": true, "traffic_out": true,
		"created_at": true, "updated_at": true, "last_seen": true,
	},
	"users": {
		"id": true, "username": true, "email": true, "role": true,
		"enabled": true, "created_at": true, "last_login_at": true,
		"traffic_quota": true, "quota_used": true, "quota_exceeded": true,
	},
	"port_forwards": {
		"id": true, "name": true, "type": true, "enabled": true,
		"created_at": true, "updated_at": true,
	},
	"node_groups": {
		"id": true, "name": true, "strategy": true,
		"created_at": true, "updated_at": true,
	},
}

// validateSortField 验证排序字段是否在白名单中
func validateSortField(table, field string) bool {
	if field == "" {
		return false
	}
	fields, ok := allowedSortFields[table]
	if !ok {
		return false
	}
	return fields[field]
}

// getSafeOrderBy 获取安全的排序语句
func getSafeOrderBy(table string, params PaginationParams, defaultOrder string) string {
	if params.SortBy == "" || !validateSortField(table, params.SortBy) {
		return defaultOrder
	}
	order := "asc"
	if params.SortDesc {
		order = "desc"
	}
	return params.SortBy + " " + order
}

// ==================== Node 操作 ====================

// ListNodes 获取节点列表（支持权限过滤）
func (s *Service) ListNodes() ([]model.Node, error) {
	var nodes []model.Node
	err := s.db.Order("id desc").Find(&nodes).Error
	return nodes, err
}

// ListNodesByOwner 获取指定用户的节点列表
func (s *Service) ListNodesByOwner(userID uint, isAdmin bool) ([]model.Node, error) {
	var nodes []model.Node
	query := s.db.Order("id desc")
	if !isAdmin {
		query = query.Where("owner_id = ? OR owner_id IS NULL", userID)
	}
	err := query.Find(&nodes).Error
	return nodes, err
}

// ListNodesPaginated 分页获取节点列表
func (s *Service) ListNodesPaginated(userID uint, isAdmin bool, params PaginationParams) (*PaginatedResult, error) {
	var nodes []model.Node
	var total int64

	query := s.db.Model(&model.Node{})

	// 权限过滤
	if !isAdmin {
		query = query.Where("owner_id = ? OR owner_id IS NULL", userID)
	}

	// 搜索过滤
	if params.Search != "" {
		search := "%" + params.Search + "%"
		query = query.Where("name LIKE ? OR host LIKE ?", search, search)
	}

	// 计算总数
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	// 排序 (使用白名单验证防止 SQL 注入)
	orderBy := getSafeOrderBy("nodes", params, "id desc")

	// 分页查询
	offset := (params.Page - 1) * params.PageSize
	if err := query.Order(orderBy).Offset(offset).Limit(params.PageSize).Find(&nodes).Error; err != nil {
		return nil, err
	}

	pages := int(total) / params.PageSize
	if int(total)%params.PageSize > 0 {
		pages++
	}

	return &PaginatedResult{
		Items:    nodes,
		Total:    total,
		Page:     params.Page,
		PageSize: params.PageSize,
		Pages:    pages,
	}, nil
}

// GetNode 获取节点（不检查权限）
func (s *Service) GetNode(id uint) (*model.Node, error) {
	var node model.Node
	err := s.db.First(&node, id).Error
	if err != nil {
		return nil, err
	}
	return &node, nil
}

// GetNodeByOwner 获取节点（检查权限）
func (s *Service) GetNodeByOwner(id uint, userID uint, isAdmin bool) (*model.Node, error) {
	var node model.Node
	query := s.db.Where("id = ?", id)
	if !isAdmin {
		query = query.Where("owner_id = ? OR owner_id IS NULL", userID)
	}
	err := query.First(&node).Error
	if err != nil {
		return nil, err
	}
	return &node, nil
}

func (s *Service) CreateNode(node *model.Node) error {
	node.AgentToken = generateToken()
	node.Status = "offline"
	node.CreatedAt = time.Now()
	node.UpdatedAt = time.Now()
	return s.db.Create(node).Error
}

func (s *Service) UpdateNode(id uint, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()
	return s.db.Model(&model.Node{}).Where("id = ?", id).Updates(updates).Error
}

func (s *Service) DeleteNode(id uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 删除关联的客户端
		if err := tx.Where("node_id = ?", id).Delete(&model.Client{}).Error; err != nil {
			return err
		}
		// 删除关联的服务
		if err := tx.Where("node_id = ?", id).Delete(&model.Service{}).Error; err != nil {
			return err
		}
		// 删除节点
		return tx.Delete(&model.Node{}, id).Error
	})
}

// GetNodeByToken 通过 Agent Token 获取节点
func (s *Service) GetNodeByToken(token string) (*model.Node, error) {
	var node model.Node
	err := s.db.Where("agent_token = ?", token).First(&node).Error
	return &node, err
}

// UpdateNodeStatus 更新节点状态
func (s *Service) UpdateNodeStatus(id uint, status string, connections int, trafficIn, trafficOut int64) error {
	// 获取当前节点信息
	node, err := s.GetNode(id)
	if err != nil {
		return err
	}
	previousStatus := node.Status

	// 更新节点状态和流量
	err = s.db.Model(&model.Node{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":      status,
		"connections": connections,
		"traffic_in":  gorm.Expr("traffic_in + ?", trafficIn),
		"traffic_out": gorm.Expr("traffic_out + ?", trafficOut),
		"quota_used":  gorm.Expr("quota_used + ?", trafficIn+trafficOut),
		"last_seen":   time.Now(),
	}).Error

	if err != nil {
		return err
	}

	// 重新获取更新后的节点信息
	node, _ = s.GetNode(id)

	// 检查节点离线
	if previousStatus == "online" && status == "offline" {
		s.alertService.CheckNodeOffline(node, previousStatus)
	}

	// 检查流量配额
	s.alertService.CheckNodeQuota(node)

	return nil
}

// TouchNode 更新节点的 updated_at 时间戳，用于触发配置同步
func (s *Service) TouchNode(id uint) error {
	return s.db.Model(&model.Node{}).Where("id = ?", id).Update("updated_at", time.Now()).Error
}

// GetNodeConfigHash 获取节点配置的哈希值（基于 updated_at）
func (s *Service) GetNodeConfigHash(id uint) string {
	node, err := s.GetNode(id)
	if err != nil {
		return ""
	}
	// 使用 updated_at 的时间戳作为简单的版本号
	return fmt.Sprintf("%d", node.UpdatedAt.Unix())
}

// ==================== Client 操作 ====================

// GetClientConfigHash 获取客户端配置的哈希值（考虑客户端和关联节点的更新时间）
func (s *Service) GetClientConfigHash(id uint) string {
	var client model.Client
	if err := s.db.Preload("Node").First(&client, id).Error; err != nil {
		return ""
	}
	// 使用客户端和节点的 updated_at 的最大值作为版本号
	clientTime := client.UpdatedAt.Unix()
	nodeTime := int64(0)
	if client.Node.ID > 0 {
		nodeTime = client.Node.UpdatedAt.Unix()
	}
	if nodeTime > clientTime {
		return fmt.Sprintf("%d", nodeTime)
	}
	return fmt.Sprintf("%d", clientTime)
}

func (s *Service) ListClients() ([]model.Client, error) {
	var clients []model.Client
	err := s.db.Preload("Node").Order("id desc").Find(&clients).Error
	return clients, err
}

// ListClientsByOwner 获取指定用户的客户端列表
func (s *Service) ListClientsByOwner(userID uint, isAdmin bool) ([]model.Client, error) {
	var clients []model.Client
	query := s.db.Preload("Node").Order("id desc")
	if !isAdmin {
		query = query.Where("owner_id = ? OR owner_id IS NULL", userID)
	}
	err := query.Find(&clients).Error
	return clients, err
}

// ListClientsPaginated 分页获取客户端列表
func (s *Service) ListClientsPaginated(userID uint, isAdmin bool, params PaginationParams) (*PaginatedResult, error) {
	var clients []model.Client
	var total int64

	query := s.db.Model(&model.Client{})

	// 权限过滤
	if !isAdmin {
		query = query.Where("owner_id = ? OR owner_id IS NULL", userID)
	}

	// 搜索过滤
	if params.Search != "" {
		search := "%" + params.Search + "%"
		query = query.Where("name LIKE ?", search)
	}

	// 计算总数
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	// 排序 (使用白名单验证防止 SQL 注入)
	orderBy := getSafeOrderBy("clients", params, "id desc")

	// 分页查询
	offset := (params.Page - 1) * params.PageSize
	if err := query.Preload("Node").Order(orderBy).Offset(offset).Limit(params.PageSize).Find(&clients).Error; err != nil {
		return nil, err
	}

	pages := int(total) / params.PageSize
	if int(total)%params.PageSize > 0 {
		pages++
	}

	return &PaginatedResult{
		Items:    clients,
		Total:    total,
		Page:     params.Page,
		PageSize: params.PageSize,
		Pages:    pages,
	}, nil
}

func (s *Service) GetClient(id uint) (*model.Client, error) {
	var client model.Client
	err := s.db.Preload("Node").First(&client, id).Error
	if err != nil {
		return nil, err
	}
	return &client, nil
}

// GetClientByOwner 获取客户端（检查权限）
func (s *Service) GetClientByOwner(id uint, userID uint, isAdmin bool) (*model.Client, error) {
	var client model.Client
	query := s.db.Preload("Node").Where("id = ?", id)
	if !isAdmin {
		query = query.Where("owner_id = ? OR owner_id IS NULL", userID)
	}
	err := query.First(&client).Error
	if err != nil {
		return nil, err
	}
	return &client, nil
}

func (s *Service) CreateClient(client *model.Client) error {
	// 验证节点存在
	var node model.Node
	if err := s.db.First(&node, client.NodeID).Error; err != nil {
		return errors.New("node not found")
	}

	client.Token = generateToken()
	client.Status = "offline"
	client.CreatedAt = time.Now()
	client.UpdatedAt = time.Now()
	return s.db.Create(client).Error
}

func (s *Service) UpdateClient(id uint, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()
	return s.db.Model(&model.Client{}).Where("id = ?", id).Updates(updates).Error
}

func (s *Service) DeleteClient(id uint) error {
	return s.db.Delete(&model.Client{}, id).Error
}

// GetClientByToken 通过 Token 获取客户端
func (s *Service) GetClientByToken(token string) (*model.Client, error) {
	var client model.Client
	err := s.db.Preload("Node").Where("token = ?", token).First(&client).Error
	return &client, err
}

// UpdateClientHeartbeat 更新客户端心跳
func (s *Service) UpdateClientHeartbeat(token string) error {
	result := s.db.Model(&model.Client{}).Where("token = ?", token).Updates(map[string]interface{}{
		"status":    "online",
		"last_seen": time.Now(),
	})
	if result.RowsAffected == 0 {
		return errors.New("client not found")
	}
	return result.Error
}

// ==================== Service 操作 ====================

func (s *Service) ListServices(nodeID uint) ([]model.Service, error) {
	var services []model.Service
	err := s.db.Where("node_id = ?", nodeID).Find(&services).Error
	return services, err
}

func (s *Service) CreateService(svc *model.Service) error {
	svc.CreatedAt = time.Now()
	svc.UpdatedAt = time.Now()
	return s.db.Create(svc).Error
}

func (s *Service) DeleteService(id uint) error {
	return s.db.Delete(&model.Service{}, id).Error
}

// ==================== GOST 操作 ====================

// GetGostClient 获取节点的 GOST API 客户端
func (s *Service) GetGostClient(nodeID uint) (*gost.Client, error) {
	node, err := s.GetNode(nodeID)
	if err != nil {
		return nil, err
	}
	return gost.NewClient(node.Host, node.APIPort, node.APIUser, node.APIPass), nil
}

// ApplyNodeConfig 应用节点配置到 GOST
func (s *Service) ApplyNodeConfig(nodeID uint) error {
	client, err := s.GetGostClient(nodeID)
	if err != nil {
		return err
	}

	node, err := s.GetNode(nodeID)
	if err != nil {
		return err
	}

	// 创建主 SOCKS5 服务
	return client.CreateSocks5Service(gost.Socks5Config{
		Name:          "main-socks5",
		Port:          node.Port,
		Username:      node.ProxyUser,
		Password:      node.ProxyPass,
		Bind:          true,
		UDP:           true,
		UDPBufferSize: 4096,
	})
}

// ==================== User 操作 ====================

func (s *Service) GetUserByUsername(username string) (*model.User, error) {
	var user model.User
	err := s.db.Where("username = ?", username).First(&user).Error
	return &user, err
}

func (s *Service) ValidateUser(username, password string) (*model.User, error) {
	user, err := s.GetUserByUsername(username)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}
	if !model.CheckPassword(user.Password, password) {
		return nil, errors.New("invalid credentials")
	}
	return user, nil
}

// ListUsers 获取用户列表
func (s *Service) ListUsers() ([]model.User, error) {
	var users []model.User
	err := s.db.Preload("Plan").Order("id asc").Find(&users).Error
	return users, err
}

// GetUser 获取单个用户
func (s *Service) GetUser(id uint) (*model.User, error) {
	var user model.User
	err := s.db.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// CreateUser 创建用户 (简化版本，用于兼容)
func (s *Service) CreateUser(username, password, role string) (*model.User, error) {
	return s.CreateUserFull(username, "", password, role, true, true)
}

// CreateUserFull 创建用户 (完整版本)
func (s *Service) CreateUserFull(username, email, password, role string, enabled, emailVerified bool) (*model.User, error) {
	// 检查用户名是否已存在
	var count int64
	s.db.Model(&model.User{}).Where("username = ?", username).Count(&count)
	if count > 0 {
		return nil, errors.New("username already exists")
	}

	// 检查邮箱是否已存在（如果提供了邮箱）
	if email != "" {
		s.db.Model(&model.User{}).Where("email = ?", email).Count(&count)
		if count > 0 {
			return nil, errors.New("email already exists")
		}
	}

	// 验证密码强度
	if err := model.ValidatePasswordStrength(password); err != nil {
		return nil, err
	}

	var emailPtr *string
	if email != "" {
		emailPtr = &email
	}
	user := &model.User{
		Username:        username,
		Email:           emailPtr,
		Password:        model.HashPassword(password),
		Role:            role,
		Enabled:         enabled,
		PasswordChanged: true, // 管理员创建的账户无需强制改密码
		EmailVerified:   emailVerified,
	}
	if user.Role == "" {
		user.Role = "user"
	}

	err := s.db.Create(user).Error
	return user, err
}

// UpdateUser 更新用户信息
func (s *Service) UpdateUser(id uint, updates map[string]interface{}) error {
	// 如果更新密码，需要验证和哈希
	if password, ok := updates["password"].(string); ok && password != "" {
		if err := model.ValidatePasswordStrength(password); err != nil {
			return err
		}
		updates["password"] = model.HashPassword(password)
	} else {
		delete(updates, "password")
	}

	// 处理 email 字段：空字符串转为 nil
	if email, ok := updates["email"].(string); ok {
		if email == "" {
			updates["email"] = nil
		}
	}

	delete(updates, "id")
	delete(updates, "created_at")

	return s.db.Model(&model.User{}).Where("id = ?", id).Updates(updates).Error
}

// DeleteUser 删除用户
func (s *Service) DeleteUser(id uint) error {
	// 不允许删除最后一个管理员
	var adminCount int64
	s.db.Model(&model.User{}).Where("role = ?", "admin").Count(&adminCount)

	var user model.User
	if err := s.db.First(&user, id).Error; err != nil {
		return err
	}

	if user.Role == "admin" && adminCount <= 1 {
		return errors.New("cannot delete the last admin user")
	}

	return s.db.Delete(&model.User{}, id).Error
}

// ChangePassword 修改密码
func (s *Service) ChangePassword(id uint, oldPassword, newPassword string) error {
	user, err := s.GetUser(id)
	if err != nil {
		return errors.New("user not found")
	}

	if !model.CheckPassword(user.Password, oldPassword) {
		return errors.New("incorrect old password")
	}

	// 验证新密码强度
	if err := model.ValidatePasswordStrength(newPassword); err != nil {
		return err
	}

	// 更新密码并标记已修改
	return s.db.Model(&model.User{}).Where("id = ?", id).Updates(map[string]interface{}{
		"password":         model.HashPassword(newPassword),
		"password_changed": true,
	}).Error
}

// ==================== 用户注册与验证 ====================

// IsRegistrationEnabled 检查是否开放注册
func (s *Service) IsRegistrationEnabled() bool {
	return s.GetSiteConfig(model.ConfigRegistrationEnabled) == "true"
}

// IsEmailVerificationRequired 检查是否需要邮件验证
func (s *Service) IsEmailVerificationRequired() bool {
	return s.GetSiteConfig(model.ConfigEmailVerificationRequired) == "true"
}

// GenerateToken 生成随机令牌
func GenerateToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// RegisterUser 用户注册
func (s *Service) RegisterUser(username, email, password string) (*model.User, error) {
	// 检查是否开放注册
	if !s.IsRegistrationEnabled() {
		return nil, errors.New("registration is disabled")
	}

	// 检查用户名是否已存在
	var count int64
	s.db.Model(&model.User{}).Where("username = ?", username).Count(&count)
	if count > 0 {
		return nil, errors.New("username already exists")
	}

	// 检查邮箱是否已存在
	if email != "" {
		s.db.Model(&model.User{}).Where("email = ?", email).Count(&count)
		if count > 0 {
			return nil, errors.New("email already exists")
		}
	}

	// 验证密码强度
	if err := model.ValidatePasswordStrength(password); err != nil {
		return nil, err
	}

	// 获取默认角色
	defaultRole := s.GetSiteConfig(model.ConfigDefaultRole)
	if defaultRole == "" {
		defaultRole = "user"
	}

	// 确定邮箱验证状态
	emailVerified := !s.IsEmailVerificationRequired()
	var verificationToken string
	if !emailVerified && email != "" {
		verificationToken = GenerateToken()
	}

	var emailPtr *string
	if email != "" {
		emailPtr = &email
	}
	user := &model.User{
		Username:          username,
		Email:             emailPtr,
		Password:          model.HashPassword(password),
		Role:              defaultRole,
		Enabled:           true,
		PasswordChanged:   true, // 用户自己注册的密码无需强制修改
		EmailVerified:     emailVerified,
		VerificationToken: verificationToken,
	}

	if err := s.db.Create(user).Error; err != nil {
		return nil, err
	}

	return user, nil
}

// VerifyEmail 验证邮箱
func (s *Service) VerifyEmail(token string) (*model.User, error) {
	if token == "" {
		return nil, errors.New("invalid token")
	}

	var user model.User
	if err := s.db.Where("verification_token = ?", token).First(&user).Error; err != nil {
		return nil, errors.New("invalid or expired token")
	}

	if user.EmailVerified {
		return nil, errors.New("email already verified")
	}

	user.EmailVerified = true
	user.VerificationToken = ""
	if err := s.db.Save(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

// RequestPasswordReset 请求密码重置
func (s *Service) RequestPasswordReset(email string) (*model.User, string, error) {
	if email == "" {
		return nil, "", errors.New("email is required")
	}

	var user model.User
	if err := s.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, "", errors.New("email not found")
	}

	// 生成重置令牌
	resetToken := GenerateToken()
	expiry := time.Now().Add(1 * time.Hour)

	user.ResetToken = resetToken
	user.ResetTokenExpiry = &expiry
	if err := s.db.Save(&user).Error; err != nil {
		return nil, "", err
	}

	return &user, resetToken, nil
}

// ResetPassword 重置密码
func (s *Service) ResetPassword(token, newPassword string) error {
	if token == "" {
		return errors.New("invalid token")
	}

	// 验证新密码强度
	if err := model.ValidatePasswordStrength(newPassword); err != nil {
		return err
	}

	var user model.User
	if err := s.db.Where("reset_token = ?", token).First(&user).Error; err != nil {
		return errors.New("invalid or expired token")
	}

	// 检查令牌是否过期
	if user.ResetTokenExpiry == nil || time.Now().After(*user.ResetTokenExpiry) {
		return errors.New("token has expired")
	}

	// 更新密码并清除令牌
	user.Password = model.HashPassword(newPassword)
	user.ResetToken = ""
	user.ResetTokenExpiry = nil

	return s.db.Save(&user).Error
}

// GetUserByEmail 通过邮箱获取用户
func (s *Service) GetUserByEmail(email string) (*model.User, error) {
	var user model.User
	err := s.db.Where("email = ?", email).First(&user).Error
	return &user, err
}

// GetUserByVerificationToken 通过验证令牌获取用户
func (s *Service) GetUserByVerificationToken(token string) (*model.User, error) {
	var user model.User
	err := s.db.Where("verification_token = ?", token).First(&user).Error
	return &user, err
}

// UpdateUserLoginInfo 更新用户登录信息
func (s *Service) UpdateUserLoginInfo(userID uint, ip string) error {
	now := time.Now()
	return s.db.Model(&model.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"last_login_at": now,
		"last_login_ip": ip,
	}).Error
}

// ResendVerificationEmail 重新发送验证邮件
func (s *Service) ResendVerificationEmail(userID uint) (string, error) {
	var user model.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return "", errors.New("user not found")
	}

	if user.EmailVerified {
		return "", errors.New("email already verified")
	}

	if user.Email == nil || *user.Email == "" {
		return "", errors.New("no email address")
	}

	// 生成新令牌
	token := GenerateToken()
	user.VerificationToken = token
	if err := s.db.Save(&user).Error; err != nil {
		return "", err
	}

	return token, nil
}

// ==================== 用户流量配额 ====================

// UserTrafficSummary 用户流量汇总
type UserTrafficSummary struct {
	TotalTrafficIn  int64 `json:"total_traffic_in"`
	TotalTrafficOut int64 `json:"total_traffic_out"`
	TotalQuotaUsed  int64 `json:"total_quota_used"`
	NodesCount      int   `json:"nodes_count"`
	ClientsCount    int   `json:"clients_count"`
	TunnelsCount    int   `json:"tunnels_count"`
}

// GetUserTrafficSummary 获取用户流量汇总 (聚合所有拥有的资源)
func (s *Service) GetUserTrafficSummary(userID uint) (*UserTrafficSummary, error) {
	summary := &UserTrafficSummary{}

	// 统计用户拥有的节点流量
	var nodeResult struct {
		TrafficIn  int64
		TrafficOut int64
		QuotaUsed  int64
		Count      int
	}
	s.db.Model(&model.Node{}).
		Where("owner_id = ?", userID).
		Select("COALESCE(SUM(traffic_in), 0) as traffic_in, COALESCE(SUM(traffic_out), 0) as traffic_out, COALESCE(SUM(quota_used), 0) as quota_used, COUNT(*) as count").
		Scan(&nodeResult)

	// 统计用户拥有的客户端流量
	var clientResult struct {
		TrafficIn  int64
		TrafficOut int64
		QuotaUsed  int64
		Count      int
	}
	s.db.Model(&model.Client{}).
		Where("owner_id = ?", userID).
		Select("COALESCE(SUM(traffic_in), 0) as traffic_in, COALESCE(SUM(traffic_out), 0) as traffic_out, COALESCE(SUM(quota_used), 0) as quota_used, COUNT(*) as count").
		Scan(&clientResult)

	// 统计用户拥有的隧道流量
	var tunnelResult struct {
		TrafficIn  int64
		TrafficOut int64
		Count      int
	}
	s.db.Model(&model.Tunnel{}).
		Where("owner_id = ?", userID).
		Select("COALESCE(SUM(traffic_in), 0) as traffic_in, COALESCE(SUM(traffic_out), 0) as traffic_out, COUNT(*) as count").
		Scan(&tunnelResult)

	summary.TotalTrafficIn = nodeResult.TrafficIn + clientResult.TrafficIn + tunnelResult.TrafficIn
	summary.TotalTrafficOut = nodeResult.TrafficOut + clientResult.TrafficOut + tunnelResult.TrafficOut
	summary.TotalQuotaUsed = nodeResult.QuotaUsed + clientResult.QuotaUsed
	summary.NodesCount = nodeResult.Count
	summary.ClientsCount = clientResult.Count
	summary.TunnelsCount = tunnelResult.Count

	return summary, nil
}

// UpdateUserQuotaUsed 更新用户配额使用量 (从拥有的资源聚合)
func (s *Service) UpdateUserQuotaUsed(userID uint) error {
	summary, err := s.GetUserTrafficSummary(userID)
	if err != nil {
		return err
	}

	// 更新用户的 quota_used
	totalUsed := summary.TotalTrafficIn + summary.TotalTrafficOut

	return s.db.Model(&model.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"quota_used": totalUsed,
	}).Error
}

// CheckUserQuota 检查用户配额是否超限
func (s *Service) CheckUserQuota(userID uint) (bool, error) {
	var user model.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return false, err
	}

	// 如果没有设置配额，则不限制
	if user.TrafficQuota <= 0 {
		return false, nil
	}

	// 获取用户流量汇总
	summary, err := s.GetUserTrafficSummary(userID)
	if err != nil {
		return false, err
	}

	totalUsed := summary.TotalTrafficIn + summary.TotalTrafficOut
	exceeded := totalUsed >= user.TrafficQuota

	// 更新用户的 quota_used 和 quota_exceeded
	s.db.Model(&model.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"quota_used":     totalUsed,
		"quota_exceeded": exceeded,
	})

	return exceeded, nil
}

// ResetUserQuota 重置用户配额
func (s *Service) ResetUserQuota(userID uint) error {
	return s.db.Model(&model.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"quota_used":      0,
		"quota_exceeded":  false,
		"quota_reset_at":  time.Now(),
	}).Error
}

// CheckAndResetUserQuotas 检查并重置所有用户的配额 (按月重置日)
func (s *Service) CheckAndResetUserQuotas() error {
	now := time.Now()
	day := now.Day()

	// 获取需要重置的用户
	var users []model.User
	s.db.Where("quota_reset_day = ? AND (quota_reset_at < ? OR quota_reset_at IS NULL)",
		day, now.AddDate(0, 0, -1)).Find(&users)

	for _, user := range users {
		s.ResetUserQuota(user.ID)
	}

	return nil
}

// GetUsersWithTrafficSummary 获取用户列表并附带流量汇总
func (s *Service) GetUsersWithTrafficSummary() ([]map[string]interface{}, error) {
	users, err := s.ListUsers()
	if err != nil {
		return nil, err
	}

	result := make([]map[string]interface{}, len(users))
	for i, user := range users {
		summary, _ := s.GetUserTrafficSummary(user.ID)

		result[i] = map[string]interface{}{
			"id":               user.ID,
			"username":         user.Username,
			"email":            user.Email,
			"role":             user.Role,
			"enabled":          user.Enabled,
			"password_changed": user.PasswordChanged,
			"email_verified":   user.EmailVerified,
			"last_login_at":    user.LastLoginAt,
			"last_login_ip":    user.LastLoginIP,
			"created_at":       user.CreatedAt,
			"updated_at":       user.UpdatedAt,
			// 流量配额相关
			"traffic_quota":    user.TrafficQuota,
			"quota_used":       user.QuotaUsed,
			"quota_reset_day":  user.QuotaResetDay,
			"quota_reset_at":   user.QuotaResetAt,
			"quota_exceeded":   user.QuotaExceeded,
			// 流量汇总
			"traffic_summary":  summary,
			// 套餐相关
			"plan_id":           user.PlanID,
			"plan":              user.Plan,
			"plan_start_at":     user.PlanStartAt,
			"plan_expire_at":    user.PlanExpireAt,
			"plan_traffic_used": user.PlanTrafficUsed,
		}
	}

	return result, nil
}

// ==================== Stats ====================

type Stats struct {
	TotalNodes       int   `json:"total_nodes"`
	OnlineNodes      int   `json:"online_nodes"`
	TotalClients     int   `json:"total_clients"`
	OnlineClients    int   `json:"online_clients"`
	TotalUsers       int   `json:"total_users"`
	TotalTrafficIn   int64 `json:"total_traffic_in"`
	TotalTrafficOut  int64 `json:"total_traffic_out"`
	TotalConnections int   `json:"total_connections"`
}

func (s *Service) GetStats() (*Stats, error) {
	var stats Stats

	var totalNodes, onlineNodes, totalClients, onlineClients, totalUsers int64
	s.db.Model(&model.Node{}).Count(&totalNodes)
	s.db.Model(&model.Node{}).Where("status = ?", "online").Count(&onlineNodes)
	s.db.Model(&model.Client{}).Count(&totalClients)
	s.db.Model(&model.Client{}).Where("status = ?", "online").Count(&onlineClients)
	s.db.Model(&model.User{}).Count(&totalUsers)
	stats.TotalNodes = int(totalNodes)
	stats.OnlineNodes = int(onlineNodes)
	stats.TotalClients = int(totalClients)
	stats.OnlineClients = int(onlineClients)
	stats.TotalUsers = int(totalUsers)

	var result struct {
		TrafficIn   int64
		TrafficOut  int64
		Connections int
	}
	s.db.Model(&model.Node{}).Select("SUM(traffic_in) as traffic_in, SUM(traffic_out) as traffic_out, SUM(connections) as connections").Scan(&result)
	stats.TotalTrafficIn = result.TrafficIn
	stats.TotalTrafficOut = result.TrafficOut
	stats.TotalConnections = result.Connections

	return &stats, nil
}

// ==================== Traffic History ====================

// RecordTrafficHistory 记录流量历史
func (s *Service) RecordTrafficHistory() error {
	now := time.Now()

	// 获取所有节点的当前流量数据
	var nodes []model.Node
	s.db.Find(&nodes)

	for _, node := range nodes {
		history := &model.TrafficHistory{
			NodeID:      &node.ID,
			TrafficIn:   node.TrafficIn,
			TrafficOut:  node.TrafficOut,
			Connections: node.Connections,
			RecordedAt:  now,
		}
		s.db.Create(history)
	}

	// 记录总体流量
	stats, _ := s.GetStats()
	totalHistory := &model.TrafficHistory{
		NodeID:      nil, // nil 表示总体数据
		TrafficIn:   stats.TotalTrafficIn,
		TrafficOut:  stats.TotalTrafficOut,
		Connections: stats.TotalConnections,
		RecordedAt:  now,
	}
	s.db.Create(totalHistory)

	// 清理超过 24 小时的数据
	s.db.Where("recorded_at < ?", now.Add(-24*time.Hour)).Delete(&model.TrafficHistory{})

	return nil
}

// TrafficPoint 流量数据点
type TrafficPoint struct {
	Time        time.Time `json:"time"`
	TrafficIn   int64     `json:"traffic_in"`
	TrafficOut  int64     `json:"traffic_out"`
	Connections int       `json:"connections"`
}

// GetTrafficHistory 获取流量历史数据
func (s *Service) GetTrafficHistory(nodeID *uint, hours int) ([]TrafficPoint, error) {
	if hours <= 0 {
		hours = 1
	}
	if hours > 24 {
		hours = 24
	}

	since := time.Now().Add(-time.Duration(hours) * time.Hour)

	var histories []model.TrafficHistory
	query := s.db.Where("recorded_at > ?", since).Order("recorded_at asc")

	if nodeID != nil {
		query = query.Where("node_id = ?", *nodeID)
	} else {
		query = query.Where("node_id IS NULL")
	}

	if err := query.Find(&histories).Error; err != nil {
		return nil, err
	}

	points := make([]TrafficPoint, len(histories))
	for i, h := range histories {
		points[i] = TrafficPoint{
			Time:        h.RecordedAt,
			TrafficIn:   h.TrafficIn,
			TrafficOut:  h.TrafficOut,
			Connections: h.Connections,
		}
	}

	return points, nil
}

// ==================== 辅助函数 ====================

func generateToken() string {
	b := make([]byte, 32) // 256 bit (更安全)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// ==================== 端口转发 ====================

func (s *Service) ListPortForwards(userID uint, isAdmin bool) ([]model.PortForward, error) {
	var forwards []model.PortForward
	query := s.db.Order("id desc")
	if !isAdmin {
		query = query.Where("owner_id = ? OR owner_id IS NULL", userID)
	}
	err := query.Find(&forwards).Error
	return forwards, err
}

func (s *Service) GetPortForward(id uint) (*model.PortForward, error) {
	var forward model.PortForward
	err := s.db.First(&forward, id).Error
	if err != nil {
		return nil, err
	}
	return &forward, nil
}

func (s *Service) CreatePortForward(forward *model.PortForward) error {
	forward.CreatedAt = time.Now()
	forward.UpdatedAt = time.Now()
	return s.db.Create(forward).Error
}

func (s *Service) UpdatePortForward(id uint, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()
	delete(updates, "id")
	delete(updates, "created_at")
	return s.db.Model(&model.PortForward{}).Where("id = ?", id).Updates(updates).Error
}

func (s *Service) DeletePortForward(id uint) error {
	return s.db.Delete(&model.PortForward{}, id).Error
}

// ==================== 节点组 (负载均衡) ====================

func (s *Service) ListNodeGroups(userID uint, isAdmin bool) ([]model.NodeGroup, error) {
	var groups []model.NodeGroup
	query := s.db.Order("id desc")
	if !isAdmin {
		query = query.Where("owner_id = ? OR owner_id IS NULL", userID)
	}
	err := query.Find(&groups).Error
	return groups, err
}

func (s *Service) GetNodeGroup(id uint) (*model.NodeGroup, error) {
	var group model.NodeGroup
	err := s.db.First(&group, id).Error
	if err != nil {
		return nil, err
	}
	return &group, nil
}

func (s *Service) CreateNodeGroup(group *model.NodeGroup) error {
	group.CreatedAt = time.Now()
	group.UpdatedAt = time.Now()
	return s.db.Create(group).Error
}

func (s *Service) UpdateNodeGroup(id uint, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()
	delete(updates, "id")
	delete(updates, "created_at")
	return s.db.Model(&model.NodeGroup{}).Where("id = ?", id).Updates(updates).Error
}

func (s *Service) DeleteNodeGroup(id uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 删除组成员
		if err := tx.Where("group_id = ?", id).Delete(&model.NodeGroupMember{}).Error; err != nil {
			return err
		}
		// 删除组
		return tx.Delete(&model.NodeGroup{}, id).Error
	})
}

// ==================== 节点组成员 ====================

func (s *Service) ListNodeGroupMembers(groupID uint) ([]model.NodeGroupMember, error) {
	var members []model.NodeGroupMember
	err := s.db.Where("group_id = ?", groupID).Order("priority asc").Find(&members).Error
	return members, err
}

func (s *Service) AddNodeGroupMember(member *model.NodeGroupMember) error {
	// 检查是否已存在
	var count int64
	s.db.Model(&model.NodeGroupMember{}).Where("group_id = ? AND node_id = ?", member.GroupID, member.NodeID).Count(&count)
	if count > 0 {
		return errors.New("node already in group")
	}
	return s.db.Create(member).Error
}

func (s *Service) RemoveNodeGroupMember(id uint) error {
	return s.db.Delete(&model.NodeGroupMember{}, id).Error
}

func (s *Service) GetNodeGroupMembersWithNodes(groupID uint) ([]gost.NodeMemberWithNode, error) {
	members, err := s.ListNodeGroupMembers(groupID)
	if err != nil {
		return nil, err
	}

	result := make([]gost.NodeMemberWithNode, 0, len(members))
	for _, m := range members {
		node, err := s.GetNode(m.NodeID)
		if err != nil {
			continue
		}
		result = append(result, gost.NodeMemberWithNode{
			Member: m,
			Node:   node,
		})
	}

	return result, nil
}

// ==================== 操作日志 ====================

// LogOperation 记录操作日志
func (s *Service) LogOperation(userID uint, username, action, resource string, resourceID uint, detail, ip, userAgent, status string) {
	log := &model.OperationLog{
		UserID:     userID,
		Username:   username,
		Action:     action,
		Resource:   resource,
		ResourceID: resourceID,
		Detail:     detail,
		IP:         ip,
		UserAgent:  userAgent,
		Status:     status,
	}
	s.db.Create(log)
}

// GetOperationLogs 获取操作日志列表
func (s *Service) GetOperationLogs(limit, offset int, action, resource string) ([]model.OperationLog, int64, error) {
	var logs []model.OperationLog
	var total int64

	query := s.db.Model(&model.OperationLog{})

	if action != "" {
		query = query.Where("action = ?", action)
	}
	if resource != "" {
		query = query.Where("resource = ?", resource)
	}

	query.Count(&total)
	err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&logs).Error
	return logs, total, err
}

// ==================== 代理链/隧道转发 ====================

// CreateProxyChain 创建代理链
func (s *Service) CreateProxyChain(chain *model.ProxyChain) error {
	return s.db.Create(chain).Error
}

// GetProxyChain 获取代理链
func (s *Service) GetProxyChain(id uint) (*model.ProxyChain, error) {
	var chain model.ProxyChain
	err := s.db.First(&chain, id).Error
	return &chain, err
}

// UpdateProxyChain 更新代理链
func (s *Service) UpdateProxyChain(chain *model.ProxyChain) error {
	return s.db.Save(chain).Error
}

// DeleteProxyChain 删除代理链
func (s *Service) DeleteProxyChain(id uint) error {
	// 先删除跳点
	s.db.Where("chain_id = ?", id).Delete(&model.ProxyChainHop{})
	return s.db.Delete(&model.ProxyChain{}, id).Error
}

// ListProxyChains 获取代理链列表
func (s *Service) ListProxyChains(ownerID *uint) ([]model.ProxyChain, error) {
	var chains []model.ProxyChain
	query := s.db.Model(&model.ProxyChain{})
	if ownerID != nil {
		query = query.Where("owner_id = ? OR owner_id IS NULL", *ownerID)
	}
	err := query.Order("id ASC").Find(&chains).Error
	return chains, err
}

// AddProxyChainHop 添加代理链跳点
func (s *Service) AddProxyChainHop(hop *model.ProxyChainHop) error {
	return s.db.Create(hop).Error
}

// RemoveProxyChainHop 移除代理链跳点
func (s *Service) RemoveProxyChainHop(hopID uint) error {
	return s.db.Delete(&model.ProxyChainHop{}, hopID).Error
}

// UpdateProxyChainHop 更新代理链跳点
func (s *Service) UpdateProxyChainHop(hop *model.ProxyChainHop) error {
	return s.db.Save(hop).Error
}

// GetProxyChainHops 获取代理链跳点列表 (按顺序)
func (s *Service) GetProxyChainHops(chainID uint) ([]model.ProxyChainHop, error) {
	var hops []model.ProxyChainHop
	err := s.db.Where("chain_id = ?", chainID).Order("hop_order ASC").Find(&hops).Error
	return hops, err
}

// GetProxyChainHopsWithNodes 获取代理链跳点及节点信息
func (s *Service) GetProxyChainHopsWithNodes(chainID uint) ([]model.ProxyChainHop, error) {
	var hops []model.ProxyChainHop
	err := s.db.Preload("Node").Where("chain_id = ?", chainID).Order("hop_order ASC").Find(&hops).Error
	return hops, err
}

// ==================== 隧道转发 (入口-出口模式) ====================

// CreateTunnel 创建隧道
func (s *Service) CreateTunnel(tunnel *model.Tunnel) error {
	return s.db.Create(tunnel).Error
}

// GetTunnel 获取隧道
func (s *Service) GetTunnel(id uint) (*model.Tunnel, error) {
	var tunnel model.Tunnel
	err := s.db.Preload("EntryNode").Preload("ExitNode").First(&tunnel, id).Error
	return &tunnel, err
}

// UpdateTunnel 更新隧道
func (s *Service) UpdateTunnel(tunnel *model.Tunnel) error {
	return s.db.Save(tunnel).Error
}

// DeleteTunnel 删除隧道
func (s *Service) DeleteTunnel(id uint) error {
	return s.db.Delete(&model.Tunnel{}, id).Error
}

// UpdateTunnelTraffic 更新隧道流量统计 (增量)
func (s *Service) UpdateTunnelTraffic(id uint, trafficIn, trafficOut int64) error {
	return s.db.Model(&model.Tunnel{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"traffic_in":  gorm.Expr("traffic_in + ?", trafficIn),
			"traffic_out": gorm.Expr("traffic_out + ?", trafficOut),
		}).Error
}

// UpdateClientTraffic 更新客户端流量统计 (增量)
func (s *Service) UpdateClientTraffic(id uint, trafficIn, trafficOut int64) error {
	return s.db.Model(&model.Client{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"traffic_in":  gorm.Expr("traffic_in + ?", trafficIn),
			"traffic_out": gorm.Expr("traffic_out + ?", trafficOut),
		}).Error
}

// ListTunnels 获取隧道列表
func (s *Service) ListTunnels(ownerID *uint) ([]model.Tunnel, error) {
	var tunnels []model.Tunnel
	query := s.db.Preload("EntryNode").Preload("ExitNode")
	if ownerID != nil {
		query = query.Where("owner_id = ? OR owner_id IS NULL", *ownerID)
	}
	err := query.Order("id ASC").Find(&tunnels).Error
	return tunnels, err
}

// GetTunnelsByEntryNode 获取指定入口节点的所有隧道
func (s *Service) GetTunnelsByEntryNode(nodeID uint) ([]model.Tunnel, error) {
	var tunnels []model.Tunnel
	err := s.db.Preload("ExitNode").Where("entry_node_id = ? AND enabled = ?", nodeID, true).Find(&tunnels).Error
	return tunnels, err
}

// GetTunnelsByExitNode 获取指定出口节点的所有隧道
func (s *Service) GetTunnelsByExitNode(nodeID uint) ([]model.Tunnel, error) {
	var tunnels []model.Tunnel
	err := s.db.Preload("EntryNode").Where("exit_node_id = ? AND enabled = ?", nodeID, true).Find(&tunnels).Error
	return tunnels, err
}

// ==================== 通知渠道 ====================

// ListNotifyChannels 获取所有通知渠道
func (s *Service) ListNotifyChannels() ([]model.NotifyChannel, error) {
	var channels []model.NotifyChannel
	err := s.db.Find(&channels).Error
	return channels, err
}

// GetNotifyChannel 获取单个通知渠道
func (s *Service) GetNotifyChannel(id uint) (*model.NotifyChannel, error) {
	var channel model.NotifyChannel
	err := s.db.First(&channel, id).Error
	return &channel, err
}

// CreateNotifyChannel 创建通知渠道
func (s *Service) CreateNotifyChannel(channel *model.NotifyChannel) error {
	return s.db.Create(channel).Error
}

// UpdateNotifyChannel 更新通知渠道
func (s *Service) UpdateNotifyChannel(channel *model.NotifyChannel) error {
	return s.db.Save(channel).Error
}

// DeleteNotifyChannel 删除通知渠道
func (s *Service) DeleteNotifyChannel(id uint) error {
	return s.db.Delete(&model.NotifyChannel{}, id).Error
}

// ==================== 网站配置 ====================

// GetSiteConfig 获取单个配置
func (s *Service) GetSiteConfig(key string) string {
	var config model.SiteConfig
	if err := s.db.Where("key = ?", key).First(&config).Error; err != nil {
		return ""
	}
	return config.Value
}

// GetSiteConfigs 获取所有配置
func (s *Service) GetSiteConfigs() map[string]string {
	var configs []model.SiteConfig
	s.db.Find(&configs)
	result := make(map[string]string)
	for _, c := range configs {
		result[c.Key] = c.Value
	}
	return result
}

// SetSiteConfig 设置配置
func (s *Service) SetSiteConfig(key, value string) error {
	var config model.SiteConfig
	if err := s.db.Where("key = ?", key).First(&config).Error; err != nil {
		// 不存在则创建
		config = model.SiteConfig{Key: key, Value: value}
		return s.db.Create(&config).Error
	}
	// 存在则更新
	return s.db.Model(&config).Update("value", value).Error
}

// SetSiteConfigs 批量设置配置
func (s *Service) SetSiteConfigs(configs map[string]string) error {
	for key, value := range configs {
		if err := s.SetSiteConfig(key, value); err != nil {
			return err
		}
	}
	return nil
}

// InitDefaultSiteConfigs 初始化默认网站配置
func (s *Service) InitDefaultSiteConfigs() {
	defaults := map[string]string{
		"site_name":        "GOST Panel",
		"site_description": "GOST 代理管理面板",
		"favicon_url":      "/vite.svg",
		"logo_url":         "",
		"footer_text":      "",
		"custom_css":       "",
	}
	for key, value := range defaults {
		if s.GetSiteConfig(key) == "" {
			s.SetSiteConfig(key, value)
		}
	}
}

// ==================== 节点标签 ====================

// ListTags 获取所有标签
func (s *Service) ListTags() ([]model.Tag, error) {
	var tags []model.Tag
	err := s.db.Order("name asc").Find(&tags).Error
	return tags, err
}

// GetTag 获取单个标签
func (s *Service) GetTag(id uint) (*model.Tag, error) {
	var tag model.Tag
	err := s.db.First(&tag, id).Error
	if err != nil {
		return nil, err
	}
	return &tag, nil
}

// CreateTag 创建标签
func (s *Service) CreateTag(tag *model.Tag) error {
	// 检查名称是否已存在
	var count int64
	s.db.Model(&model.Tag{}).Where("name = ?", tag.Name).Count(&count)
	if count > 0 {
		return errors.New("tag name already exists")
	}
	return s.db.Create(tag).Error
}

// UpdateTag 更新标签
func (s *Service) UpdateTag(id uint, updates map[string]interface{}) error {
	// 如果更新名称，检查是否重复
	if name, ok := updates["name"].(string); ok && name != "" {
		var count int64
		s.db.Model(&model.Tag{}).Where("name = ? AND id != ?", name, id).Count(&count)
		if count > 0 {
			return errors.New("tag name already exists")
		}
	}
	delete(updates, "id")
	delete(updates, "created_at")
	return s.db.Model(&model.Tag{}).Where("id = ?", id).Updates(updates).Error
}

// DeleteTag 删除标签
func (s *Service) DeleteTag(id uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 删除节点-标签关联
		if err := tx.Where("tag_id = ?", id).Delete(&model.NodeTag{}).Error; err != nil {
			return err
		}
		// 删除标签
		return tx.Delete(&model.Tag{}, id).Error
	})
}

// GetNodeTags 获取节点的所有标签
func (s *Service) GetNodeTags(nodeID uint) ([]model.Tag, error) {
	var nodeTags []model.NodeTag
	if err := s.db.Preload("Tag").Where("node_id = ?", nodeID).Find(&nodeTags).Error; err != nil {
		return nil, err
	}
	tags := make([]model.Tag, 0, len(nodeTags))
	for _, nt := range nodeTags {
		if nt.Tag != nil {
			tags = append(tags, *nt.Tag)
		}
	}
	return tags, nil
}

// AddNodeTag 给节点添加标签
func (s *Service) AddNodeTag(nodeID, tagID uint) error {
	// 检查节点是否存在
	var nodeCount int64
	s.db.Model(&model.Node{}).Where("id = ?", nodeID).Count(&nodeCount)
	if nodeCount == 0 {
		return errors.New("node not found")
	}

	// 检查标签是否存在
	var tagCount int64
	s.db.Model(&model.Tag{}).Where("id = ?", tagID).Count(&tagCount)
	if tagCount == 0 {
		return errors.New("tag not found")
	}

	// 检查是否已关联
	var count int64
	s.db.Model(&model.NodeTag{}).Where("node_id = ? AND tag_id = ?", nodeID, tagID).Count(&count)
	if count > 0 {
		return errors.New("tag already added to node")
	}

	return s.db.Create(&model.NodeTag{NodeID: nodeID, TagID: tagID}).Error
}

// RemoveNodeTag 从节点移除标签
func (s *Service) RemoveNodeTag(nodeID, tagID uint) error {
	result := s.db.Where("node_id = ? AND tag_id = ?", nodeID, tagID).Delete(&model.NodeTag{})
	if result.RowsAffected == 0 {
		return errors.New("node tag not found")
	}
	return result.Error
}

// GetNodesByTag 获取具有指定标签的所有节点
func (s *Service) GetNodesByTag(tagID uint) ([]model.Node, error) {
	var nodeTags []model.NodeTag
	if err := s.db.Where("tag_id = ?", tagID).Find(&nodeTags).Error; err != nil {
		return nil, err
	}

	if len(nodeTags) == 0 {
		return []model.Node{}, nil
	}

	nodeIDs := make([]uint, len(nodeTags))
	for i, nt := range nodeTags {
		nodeIDs[i] = nt.NodeID
	}

	var nodes []model.Node
	err := s.db.Where("id IN ?", nodeIDs).Find(&nodes).Error
	return nodes, err
}

// SetNodeTags 设置节点的标签（替换现有标签）
func (s *Service) SetNodeTags(nodeID uint, tagIDs []uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 删除现有关联
		if err := tx.Where("node_id = ?", nodeID).Delete(&model.NodeTag{}).Error; err != nil {
			return err
		}
		// 添加新关联
		for _, tagID := range tagIDs {
			if err := tx.Create(&model.NodeTag{NodeID: nodeID, TagID: tagID}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// ==================== 套餐管理 ====================

// ListPlans 获取所有套餐
func (s *Service) ListPlans() ([]model.Plan, error) {
	var plans []model.Plan
	err := s.db.Order("sort_order asc, id asc").Find(&plans).Error
	return plans, err
}

// GetPlan 获取单个套餐
func (s *Service) GetPlan(id uint) (*model.Plan, error) {
	var plan model.Plan
	err := s.db.First(&plan, id).Error
	if err != nil {
		return nil, err
	}
	return &plan, nil
}

// CreatePlan 创建套餐
func (s *Service) CreatePlan(plan *model.Plan) error {
	plan.CreatedAt = time.Now()
	plan.UpdatedAt = time.Now()
	return s.db.Create(plan).Error
}

// UpdatePlan 更新套餐
func (s *Service) UpdatePlan(id uint, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()
	delete(updates, "id")
	delete(updates, "created_at")
	return s.db.Model(&model.Plan{}).Where("id = ?", id).Updates(updates).Error
}

// DeletePlan 删除套餐
func (s *Service) DeletePlan(id uint) error {
	// 检查是否有用户正在使用此套餐
	var count int64
	s.db.Model(&model.User{}).Where("plan_id = ?", id).Count(&count)
	if count > 0 {
		return errors.New("该套餐正在被使用，无法删除")
	}
	return s.db.Delete(&model.Plan{}, id).Error
}

// AssignUserPlan 为用户分配套餐
func (s *Service) AssignUserPlan(userID, planID uint) error {
	plan, err := s.GetPlan(planID)
	if err != nil {
		return errors.New("套餐不存在")
	}

	if !plan.Enabled {
		return errors.New("套餐已禁用")
	}

	now := time.Now()
	var expireAt *time.Time
	if plan.Duration > 0 {
		expire := now.AddDate(0, 0, plan.Duration)
		expireAt = &expire
	}

	return s.db.Model(&model.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"plan_id":           planID,
		"plan_start_at":     now,
		"plan_expire_at":    expireAt,
		"plan_traffic_used": 0,
		// 套餐配额覆盖用户配额
		"traffic_quota":     plan.TrafficQuota,
		"quota_used":        0,
		"quota_exceeded":    false,
	}).Error
}

// RemoveUserPlan 移除用户套餐
func (s *Service) RemoveUserPlan(userID uint) error {
	return s.db.Model(&model.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"plan_id":           nil,
		"plan_start_at":     nil,
		"plan_expire_at":    nil,
		"plan_traffic_used": 0,
	}).Error
}

// RenewUserPlan 续期用户套餐
func (s *Service) RenewUserPlan(userID uint, days int) error {
	var user model.User
	if err := s.db.Preload("Plan").First(&user, userID).Error; err != nil {
		return errors.New("用户不存在")
	}

	if user.PlanID == nil {
		return errors.New("用户没有套餐")
	}

	now := time.Now()
	var baseTime time.Time
	if user.PlanExpireAt != nil && user.PlanExpireAt.After(now) {
		baseTime = *user.PlanExpireAt
	} else {
		baseTime = now
	}

	newExpireAt := baseTime.AddDate(0, 0, days)

	return s.db.Model(&model.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"plan_expire_at":    newExpireAt,
		"plan_traffic_used": 0,
		"quota_used":        0,
		"quota_exceeded":    false,
	}).Error
}

// CheckUserPlanStatus 检查用户套餐状态 (是否过期或超限)
func (s *Service) CheckUserPlanStatus(userID uint) (expired bool, exceeded bool, err error) {
	var user model.User
	if err = s.db.Preload("Plan").First(&user, userID).Error; err != nil {
		return false, false, err
	}

	if user.PlanID == nil {
		return false, false, nil
	}

	now := time.Now()

	// 检查是否过期
	if user.PlanExpireAt != nil && user.PlanExpireAt.Before(now) {
		expired = true
	}

	// 检查是否超限
	if user.Plan != nil && user.Plan.TrafficQuota > 0 {
		if user.PlanTrafficUsed >= user.Plan.TrafficQuota {
			exceeded = true
		}
	}

	return expired, exceeded, nil
}

// GetUsersWithExpiredPlans 获取套餐已过期的用户
func (s *Service) GetUsersWithExpiredPlans() ([]model.User, error) {
	var users []model.User
	now := time.Now()
	err := s.db.Preload("Plan").
		Where("plan_id IS NOT NULL AND plan_expire_at IS NOT NULL AND plan_expire_at < ?", now).
		Find(&users).Error
	return users, err
}

// GetPlanUserCount 获取套餐的用户数量
func (s *Service) GetPlanUserCount(planID uint) (int64, error) {
	var count int64
	err := s.db.Model(&model.User{}).Where("plan_id = ?", planID).Count(&count).Error
	return count, err
}

// ==================== Bypass 分流规则 ====================

func (s *Service) ListBypasses(userID uint, isAdmin bool) ([]model.Bypass, error) {
	var bypasses []model.Bypass
	query := s.db.Order("id desc")
	if !isAdmin {
		query = query.Where("owner_id = ? OR owner_id IS NULL", userID)
	}
	return bypasses, query.Find(&bypasses).Error
}

func (s *Service) GetBypass(id uint) (*model.Bypass, error) {
	var bypass model.Bypass
	err := s.db.First(&bypass, id).Error
	if err != nil {
		return nil, err
	}
	return &bypass, nil
}

func (s *Service) CreateBypass(bypass *model.Bypass) error {
	bypass.CreatedAt = time.Now()
	bypass.UpdatedAt = time.Now()
	return s.db.Create(bypass).Error
}

func (s *Service) UpdateBypass(id uint, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()
	delete(updates, "id")
	delete(updates, "created_at")
	return s.db.Model(&model.Bypass{}).Where("id = ?", id).Updates(updates).Error
}

func (s *Service) DeleteBypass(id uint) error {
	return s.db.Delete(&model.Bypass{}, id).Error
}

func (s *Service) GetBypassesByNode(nodeID uint) ([]model.Bypass, error) {
	var bypasses []model.Bypass
	err := s.db.Where("node_id = ? OR node_id IS NULL", nodeID).Find(&bypasses).Error
	return bypasses, err
}

// ==================== Admission 准入控制 ====================

func (s *Service) ListAdmissions(userID uint, isAdmin bool) ([]model.Admission, error) {
	var admissions []model.Admission
	query := s.db.Order("id desc")
	if !isAdmin {
		query = query.Where("owner_id = ? OR owner_id IS NULL", userID)
	}
	return admissions, query.Find(&admissions).Error
}

func (s *Service) GetAdmission(id uint) (*model.Admission, error) {
	var admission model.Admission
	err := s.db.First(&admission, id).Error
	if err != nil {
		return nil, err
	}
	return &admission, nil
}

func (s *Service) CreateAdmission(admission *model.Admission) error {
	admission.CreatedAt = time.Now()
	admission.UpdatedAt = time.Now()
	return s.db.Create(admission).Error
}

func (s *Service) UpdateAdmission(id uint, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()
	delete(updates, "id")
	delete(updates, "created_at")
	return s.db.Model(&model.Admission{}).Where("id = ?", id).Updates(updates).Error
}

func (s *Service) DeleteAdmission(id uint) error {
	return s.db.Delete(&model.Admission{}, id).Error
}

func (s *Service) GetAdmissionsByNode(nodeID uint) ([]model.Admission, error) {
	var admissions []model.Admission
	err := s.db.Where("node_id = ? OR node_id IS NULL", nodeID).Find(&admissions).Error
	return admissions, err
}

// ==================== HostMapping 主机映射 ====================

func (s *Service) ListHostMappings(userID uint, isAdmin bool) ([]model.HostMapping, error) {
	var mappings []model.HostMapping
	query := s.db.Order("id desc")
	if !isAdmin {
		query = query.Where("owner_id = ? OR owner_id IS NULL", userID)
	}
	return mappings, query.Find(&mappings).Error
}

func (s *Service) GetHostMapping(id uint) (*model.HostMapping, error) {
	var mapping model.HostMapping
	err := s.db.First(&mapping, id).Error
	if err != nil {
		return nil, err
	}
	return &mapping, nil
}

func (s *Service) CreateHostMapping(mapping *model.HostMapping) error {
	mapping.CreatedAt = time.Now()
	mapping.UpdatedAt = time.Now()
	return s.db.Create(mapping).Error
}

func (s *Service) UpdateHostMapping(id uint, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()
	delete(updates, "id")
	delete(updates, "created_at")
	return s.db.Model(&model.HostMapping{}).Where("id = ?", id).Updates(updates).Error
}

func (s *Service) DeleteHostMapping(id uint) error {
	return s.db.Delete(&model.HostMapping{}, id).Error
}

func (s *Service) GetHostMappingsByNode(nodeID uint) ([]model.HostMapping, error) {
	var mappings []model.HostMapping
	err := s.db.Where("node_id = ? OR node_id IS NULL", nodeID).Find(&mappings).Error
	return mappings, err
}

// ==================== Ingress 反向代理 ====================

func (s *Service) ListIngresses(userID uint, isAdmin bool) ([]model.Ingress, error) {
	var ingresses []model.Ingress
	query := s.db.Order("id desc")
	if !isAdmin {
		query = query.Where("owner_id = ? OR owner_id IS NULL", userID)
	}
	return ingresses, query.Find(&ingresses).Error
}

func (s *Service) GetIngress(id uint) (*model.Ingress, error) {
	var ingress model.Ingress
	err := s.db.First(&ingress, id).Error
	if err != nil {
		return nil, err
	}
	return &ingress, nil
}

func (s *Service) CreateIngress(ingress *model.Ingress) error {
	ingress.CreatedAt = time.Now()
	ingress.UpdatedAt = time.Now()
	return s.db.Create(ingress).Error
}

func (s *Service) UpdateIngress(id uint, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()
	delete(updates, "id")
	delete(updates, "created_at")
	return s.db.Model(&model.Ingress{}).Where("id = ?", id).Updates(updates).Error
}

func (s *Service) DeleteIngress(id uint) error {
	return s.db.Delete(&model.Ingress{}, id).Error
}

func (s *Service) GetIngressesByNode(nodeID uint) ([]model.Ingress, error) {
	var ingresses []model.Ingress
	err := s.db.Where("node_id = ? OR node_id IS NULL", nodeID).Find(&ingresses).Error
	return ingresses, err
}

// ==================== Recorder 流量记录 ====================

func (s *Service) ListRecorders(userID uint, isAdmin bool) ([]model.Recorder, error) {
	var recorders []model.Recorder
	query := s.db.Order("id desc")
	if !isAdmin {
		query = query.Where("owner_id = ? OR owner_id IS NULL", userID)
	}
	return recorders, query.Find(&recorders).Error
}

func (s *Service) GetRecorder(id uint) (*model.Recorder, error) {
	var recorder model.Recorder
	err := s.db.First(&recorder, id).Error
	if err != nil {
		return nil, err
	}
	return &recorder, nil
}

func (s *Service) CreateRecorder(recorder *model.Recorder) error {
	recorder.CreatedAt = time.Now()
	recorder.UpdatedAt = time.Now()
	return s.db.Create(recorder).Error
}

func (s *Service) UpdateRecorder(id uint, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()
	delete(updates, "id")
	delete(updates, "created_at")
	return s.db.Model(&model.Recorder{}).Where("id = ?", id).Updates(updates).Error
}

func (s *Service) DeleteRecorder(id uint) error {
	return s.db.Delete(&model.Recorder{}, id).Error
}

func (s *Service) GetRecordersByNode(nodeID uint) ([]model.Recorder, error) {
	var recorders []model.Recorder
	err := s.db.Where("node_id = ? OR node_id IS NULL", nodeID).Find(&recorders).Error
	return recorders, err
}

// ==================== Router 路由管理 ====================

func (s *Service) ListRouters(userID uint, isAdmin bool) ([]model.Router, error) {
	var routers []model.Router
	query := s.db.Order("id desc")
	if !isAdmin {
		query = query.Where("owner_id = ? OR owner_id IS NULL", userID)
	}
	return routers, query.Find(&routers).Error
}

func (s *Service) GetRouter(id uint) (*model.Router, error) {
	var router model.Router
	err := s.db.First(&router, id).Error
	if err != nil {
		return nil, err
	}
	return &router, nil
}

func (s *Service) CreateRouter(router *model.Router) error {
	router.CreatedAt = time.Now()
	router.UpdatedAt = time.Now()
	return s.db.Create(router).Error
}

func (s *Service) UpdateRouter(id uint, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()
	delete(updates, "id")
	delete(updates, "created_at")
	return s.db.Model(&model.Router{}).Where("id = ?", id).Updates(updates).Error
}

func (s *Service) DeleteRouter(id uint) error {
	return s.db.Delete(&model.Router{}, id).Error
}

func (s *Service) GetRoutersByNode(nodeID uint) ([]model.Router, error) {
	var routers []model.Router
	err := s.db.Where("node_id = ? OR node_id IS NULL", nodeID).Find(&routers).Error
	return routers, err
}

// ==================== SD 服务发现 ====================

func (s *Service) ListSDs(userID uint, isAdmin bool) ([]model.SD, error) {
	var sds []model.SD
	query := s.db.Order("id desc")
	if !isAdmin {
		query = query.Where("owner_id = ? OR owner_id IS NULL", userID)
	}
	return sds, query.Find(&sds).Error
}

func (s *Service) GetSD(id uint) (*model.SD, error) {
	var sd model.SD
	err := s.db.First(&sd, id).Error
	if err != nil {
		return nil, err
	}
	return &sd, nil
}

func (s *Service) CreateSD(sd *model.SD) error {
	sd.CreatedAt = time.Now()
	sd.UpdatedAt = time.Now()
	return s.db.Create(sd).Error
}

func (s *Service) UpdateSD(id uint, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()
	delete(updates, "id")
	delete(updates, "created_at")
	return s.db.Model(&model.SD{}).Where("id = ?", id).Updates(updates).Error
}

func (s *Service) DeleteSD(id uint) error {
	return s.db.Delete(&model.SD{}, id).Error
}

func (s *Service) GetSDsByNode(nodeID uint) ([]model.SD, error) {
	var sds []model.SD
	err := s.db.Where("node_id = ? OR node_id IS NULL", nodeID).Find(&sds).Error
	return sds, err
}

// ==================== ConfigVersion 配置版本历史 ====================

// SaveConfigVersion 保存配置版本快照
func (s *Service) SaveConfigVersion(nodeID uint, config string, comment string) error {
	version := &model.ConfigVersion{
		NodeID:    nodeID,
		Config:    config,
		Comment:   comment,
		CreatedAt: time.Now(),
	}
	return s.db.Create(version).Error
}

// GetConfigVersions 获取节点配置版本列表
func (s *Service) GetConfigVersions(nodeID uint) ([]model.ConfigVersion, error) {
	var versions []model.ConfigVersion
	err := s.db.Where("node_id = ?", nodeID).Order("id desc").Find(&versions).Error
	return versions, err
}

// GetConfigVersion 获取单个版本
func (s *Service) GetConfigVersion(id uint) (*model.ConfigVersion, error) {
	var version model.ConfigVersion
	err := s.db.First(&version, id).Error
	if err != nil {
		return nil, err
	}
	return &version, nil
}

// DeleteConfigVersion 删除版本
func (s *Service) DeleteConfigVersion(id uint) error {
	return s.db.Delete(&model.ConfigVersion{}, id).Error
}

// CleanupOldVersions 清理旧版本（保留最新 N 个）
func (s *Service) CleanupOldVersions(nodeID uint, keepCount int) error {
	// 获取所有版本，按 ID 降序
	var versions []model.ConfigVersion
	if err := s.db.Where("node_id = ?", nodeID).Order("id desc").Find(&versions).Error; err != nil {
		return err
	}

	// 如果版本数量不超过保留数量，无需清理
	if len(versions) <= keepCount {
		return nil
	}

	// 删除多余的旧版本
	var idsToDelete []uint
	for i := keepCount; i < len(versions); i++ {
		idsToDelete = append(idsToDelete, versions[i].ID)
	}

	if len(idsToDelete) > 0 {
		return s.db.Delete(&model.ConfigVersion{}, idsToDelete).Error
	}
	return nil
}
