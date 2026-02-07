package model

import (
	"os"
	"path/filepath"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Node 节点 (VPS)
type Node struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:100;not null" json:"name"`
	Host        string    `gorm:"size:255;not null" json:"host"`        // 公网地址/域名
	Port        int       `gorm:"default:38567" json:"port"`            // GOST 端口
	APIPort     int       `gorm:"default:18080" json:"api_port"`        // GOST API 端口
	APIUser     string    `gorm:"size:100" json:"-"`                    // API 认证用户 (隐藏)
	APIPass     string    `gorm:"size:100" json:"-"`                    // API 认证密码 (隐藏)
	ProxyUser   string    `gorm:"size:100" json:"proxy_user"`           // 代理认证用户
	ProxyPass   string    `gorm:"size:100" json:"-"`                    // 代理认证密码 (隐藏)
	AgentToken  string    `gorm:"size:100;uniqueIndex" json:"-"`        // Agent 认证令牌
	Status      string    `gorm:"size:20;default:offline" json:"status"` // online/offline
	TrafficIn   int64     `gorm:"default:0" json:"traffic_in"`          // 入站流量 (bytes)
	TrafficOut  int64     `gorm:"default:0" json:"traffic_out"`         // 出站流量 (bytes)
	Connections int       `gorm:"default:0" json:"connections"`         // 当前连接数
	// 协议配置
	Protocol      string `gorm:"size:50;default:socks5" json:"protocol"`    // socks5/http/ss/trojan/vmess/relay
	Transport     string `gorm:"size:50;default:tcp" json:"transport"`      // tcp/tls/ws/wss/h2/h2c/quic/kcp
	TransportOpts string `gorm:"type:text" json:"transport_opts"`           // 传输层配置 JSON
	// Shadowsocks 配置
	SSMethod   string `gorm:"size:50" json:"ss_method"`                     // aes-256-gcm/chacha20-ietf-poly1305 等
	SSPassword string `gorm:"size:100" json:"-"`                            // SS 密码 (隐藏)
	// Trojan 配置
	TrojanPassword string `gorm:"size:100" json:"-"`                        // Trojan 密码 (隐藏)
	// VMess 配置
	VMessUUID    string `gorm:"size:100" json:"-"`                          // VMess UUID (隐藏)
	VMessAlterID int    `gorm:"default:0" json:"vmess_alter_id"`
	// TLS 配置
	TLSEnabled  bool   `gorm:"default:false" json:"tls_enabled"`
	TLSCertFile string `gorm:"size:255" json:"tls_cert_file"`
	TLSKeyFile  string `gorm:"size:255" json:"tls_key_file"`
	TLSSNI      string `gorm:"size:255" json:"tls_sni"`
	TLSALPN     string `gorm:"size:255" json:"tls_alpn"`                     // TLS ALPN 协议列表 (逗号分隔)
	// WebSocket 配置
	WSPath string `gorm:"size:255" json:"ws_path"`
	WSHost string `gorm:"size:255" json:"ws_host"`
	// 限速配置 (bytes/s, 0=无限制)
	SpeedLimit    int64 `gorm:"default:0" json:"speed_limit"`
	ConnRateLimit int   `gorm:"default:0" json:"conn_rate_limit"`           // 每秒最大连接数
	// DNS 配置
	DNSServer string `gorm:"size:255" json:"dns_server"`                    // 自定义 DNS
	// 高级功能
	ProxyProtocol    int    `gorm:"default:0" json:"proxy_protocol"`        // PROXY Protocol 版本 (0=关闭, 1=v1, 2=v2)
	ProbeResist      string `gorm:"size:50" json:"probe_resist"`            // 探测抵抗类型: code/web/host/file
	ProbeResistValue string `gorm:"size:255" json:"probe_resist_value"`     // 探测抵抗值 (状态码/URL/主机名/文件路径)
	PluginConfig     string `gorm:"type:text" json:"plugin_config"`         // Plugin 配置 JSON
	// 流量配额
	TrafficQuota   int64  `gorm:"default:0" json:"traffic_quota"`       // 流量配额 (bytes), 0=无限制
	QuotaResetDay  int    `gorm:"default:1" json:"quota_reset_day"`     // 每月重置日 (1-28)
	QuotaUsed      int64  `gorm:"default:0" json:"quota_used"`          // 本周期已用流量
	QuotaResetAt   time.Time `json:"quota_reset_at"`                    // 上次重置时间
	QuotaExceeded  bool   `gorm:"default:false" json:"quota_exceeded"`  // 是否超限
	// 所有者 (权限控制)
	OwnerID     *uint     `gorm:"index" json:"owner_id,omitempty"`      // 所有者用户ID
	LastSeen    time.Time `json:"last_seen"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Client 客户端 (内网设备)
type Client struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:100;not null" json:"name"`
	Token       string    `gorm:"size:100;uniqueIndex" json:"-"`         // 认证令牌 (隐藏)
	NodeID      uint      `gorm:"index" json:"node_id"`                  // 绑定的节点
	Node        *Node     `gorm:"foreignKey:NodeID" json:"node,omitempty"`
	LocalPort   int       `gorm:"default:38777" json:"local_port"`       // 本地 SOCKS5 端口
	RemotePort  int       `json:"remote_port"`                           // 远程映射端口
	ProxyUser   string    `gorm:"size:100" json:"proxy_user"`            // 代理认证
	ProxyPass   string    `gorm:"size:100" json:"-"`                     // 代理密码 (隐藏)
	Status      string    `gorm:"size:20;default:offline" json:"status"` // connected/disconnected
	TrafficIn   int64     `gorm:"default:0" json:"traffic_in"`
	TrafficOut  int64     `gorm:"default:0" json:"traffic_out"`
	// 流量配额
	TrafficQuota   int64  `gorm:"default:0" json:"traffic_quota"`        // 流量配额 (bytes), 0=无限制
	QuotaResetDay  int    `gorm:"default:1" json:"quota_reset_day"`      // 每月重置日 (1-28)
	QuotaUsed      int64  `gorm:"default:0" json:"quota_used"`           // 本周期已用流量
	QuotaResetAt   time.Time `json:"quota_reset_at"`                     // 上次重置时间
	QuotaExceeded  bool   `gorm:"default:false" json:"quota_exceeded"`   // 是否超限
	// 所有者 (权限控制)
	OwnerID     *uint     `gorm:"index" json:"owner_id,omitempty"`       // 所有者用户ID
	LastSeen    time.Time `json:"last_seen"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Service GOST 服务配置
type Service struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	NodeID    uint      `gorm:"index" json:"node_id"`                  // 所属节点
	ClientID  *uint     `gorm:"index" json:"client_id,omitempty"`      // 所属客户端 (可选)
	Name      string    `gorm:"size:100;not null" json:"name"`         // 服务名称
	Type      string    `gorm:"size:50;not null" json:"type"`          // socks5/http/ss/trojan/vmess/relay/tcp/udp/rtcp/rudp
	Listen    string    `gorm:"size:255" json:"listen"`                // 监听地址
	Forward   string    `gorm:"size:255" json:"forward"`               // 转发地址
	Options   string    `gorm:"type:text" json:"options"`              // JSON 选项
	Enabled   bool      `gorm:"default:true" json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// PortForward 端口转发规则
type PortForward struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	NodeID      uint      `gorm:"index" json:"node_id"`
	Name        string    `gorm:"size:100;not null" json:"name"`
	Type        string    `gorm:"size:20;not null" json:"type"`           // tcp/udp/rtcp/rudp
	LocalAddr   string    `gorm:"size:255" json:"local_addr"`             // 本地监听地址
	RemoteAddr  string    `gorm:"size:255" json:"remote_addr"`            // 远程目标地址
	ChainID     *uint     `gorm:"index" json:"chain_id,omitempty"`        // 使用的转发链
	Enabled     bool      `gorm:"default:true" json:"enabled"`
	OwnerID     *uint     `gorm:"index" json:"owner_id,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// NodeGroup 节点组 (用于负载均衡)
type NodeGroup struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	Name          string    `gorm:"size:100;not null" json:"name"`
	Strategy      string    `gorm:"size:50;default:round" json:"strategy"` // round/random/fifo/hash
	Selector      string    `gorm:"size:255" json:"selector"`              // 选择器配置 JSON
	FailTimeout   int       `gorm:"default:30" json:"fail_timeout"`        // 故障超时时间(秒)
	MaxFails      int       `gorm:"default:3" json:"max_fails"`            // 最大失败次数
	HealthCheck   bool      `gorm:"default:true" json:"health_check"`      // 是否启用健康检查
	CheckInterval int       `gorm:"default:30" json:"check_interval"`      // 健康检查间隔(秒)
	OwnerID       *uint     `gorm:"index" json:"owner_id,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// NodeGroupMember 节点组成员
type NodeGroupMember struct {
	ID        uint `gorm:"primaryKey" json:"id"`
	GroupID   uint `gorm:"index" json:"group_id"`
	NodeID    uint `gorm:"index" json:"node_id"`
	Weight    int  `gorm:"default:1" json:"weight"`                      // 权重
	Priority  int  `gorm:"default:0" json:"priority"`                    // 优先级 (故障转移用)
	Enabled   bool `gorm:"default:true" json:"enabled"`
}

// Tunnel 隧道转发 (入口端-出口端模式)
type Tunnel struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:100;not null" json:"name"`
	Description string    `gorm:"size:255" json:"description"`
	// 入口端配置
	EntryNodeID uint      `gorm:"index" json:"entry_node_id"`              // 入口节点ID
	EntryNode   *Node     `gorm:"foreignKey:EntryNodeID" json:"entry_node,omitempty"`
	EntryPort   int       `gorm:"default:10000" json:"entry_port"`         // 入口监听端口
	Protocol    string    `gorm:"size:20;default:tcp+udp" json:"protocol"` // tcp/udp/tcp+udp (端口复用)
	// 出口端配置
	ExitNodeID  uint      `gorm:"index" json:"exit_node_id"`               // 出口节点ID
	ExitNode    *Node     `gorm:"foreignKey:ExitNodeID" json:"exit_node,omitempty"`
	TargetAddr  string    `gorm:"size:255" json:"target_addr"`             // 目标地址 (如 google.com:443)
	// 状态
	Enabled     bool      `gorm:"default:true" json:"enabled"`
	TrafficIn   int64     `gorm:"default:0" json:"traffic_in"`
	TrafficOut  int64     `gorm:"default:0" json:"traffic_out"`
	// 流量配额
	TrafficQuota  int64   `gorm:"default:0" json:"traffic_quota"`          // 流量配额 (bytes), 0=无限制
	QuotaResetDay int     `gorm:"default:1" json:"quota_reset_day"`
	// 限速
	SpeedLimit    int64   `gorm:"default:0" json:"speed_limit"`            // 限速 (bytes/s), 0=不限
	// 所有者
	OwnerID     *uint     `gorm:"index" json:"owner_id,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ProxyChain 代理链 (多跳顺序转发，保留用于高级场景)
type ProxyChain struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:100;not null" json:"name"`
	Description string    `gorm:"size:255" json:"description"`
	ListenAddr  string    `gorm:"size:255" json:"listen_addr"`           // 监听地址 :port
	ListenType  string    `gorm:"size:50;default:socks5" json:"listen_type"` // socks5/http/tcp/udp
	TargetAddr  string    `gorm:"size:255" json:"target_addr"`           // 最终目标地址 (可选，用于端口转发)
	Enabled     bool      `gorm:"default:true" json:"enabled"`
	OwnerID     *uint     `gorm:"index" json:"owner_id,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ProxyChainHop 代理链跳点 (多跳节点)
type ProxyChainHop struct {
	ID        uint  `gorm:"primaryKey" json:"id"`
	ChainID   uint  `gorm:"index" json:"chain_id"`
	NodeID    uint  `gorm:"index" json:"node_id"`
	Node      *Node `gorm:"foreignKey:NodeID" json:"node,omitempty"`
	HopOrder  int   `gorm:"default:0" json:"hop_order"`                  // 跳点顺序 (0=第一跳, 1=第二跳...)
	Enabled   bool  `gorm:"default:true" json:"enabled"`
}

// DNSConfig DNS 配置
type DNSConfig struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	NodeID      uint      `gorm:"index" json:"node_id"`
	Nameservers string    `gorm:"type:text" json:"nameservers"`           // DNS 服务器列表 JSON
	TTL         int       `gorm:"default:60" json:"ttl"`                  // 缓存 TTL
	Async       bool      `gorm:"default:false" json:"async"`             // 异步解析
	Enabled     bool      `gorm:"default:true" json:"enabled"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Tag 标签
type Tag struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:50;uniqueIndex;not null" json:"name"`
	Color     string    `gorm:"size:20;default:#3b82f6" json:"color"`   // 标签颜色 (hex)
	CreatedAt time.Time `json:"created_at"`
}

// NodeTag 节点标签关联
type NodeTag struct {
	ID     uint `gorm:"primaryKey" json:"id"`
	NodeID uint `gorm:"index;not null" json:"node_id"`
	TagID  uint `gorm:"index;not null" json:"tag_id"`
	Tag    *Tag `gorm:"foreignKey:TagID" json:"tag,omitempty"`
}

// User 用户
type User struct {
	ID                uint       `gorm:"primaryKey" json:"id"`
	Username          string     `gorm:"size:50;uniqueIndex;not null" json:"username"`
	Email             *string    `gorm:"size:100;uniqueIndex" json:"email"`
	Password          string     `gorm:"size:100;not null" json:"-"`
	Role              string     `gorm:"size:20;default:user" json:"role"`    // admin/user/viewer
	Enabled           bool       `gorm:"default:true" json:"enabled"`         // 账户是否启用
	PasswordChanged   bool       `gorm:"default:false" json:"password_changed"` // 是否已修改初始密码
	EmailVerified     bool       `gorm:"default:false" json:"email_verified"` // 邮箱是否已验证
	VerificationToken string     `gorm:"size:100" json:"-"`                   // 邮箱验证令牌
	ResetToken        string     `gorm:"size:100" json:"-"`                   // 密码重置令牌
	ResetTokenExpiry  *time.Time `json:"-"`                                   // 重置令牌过期时间
	LastLoginAt       *time.Time `json:"last_login_at,omitempty"`             // 上次登录时间
	LastLoginIP       string     `gorm:"size:50" json:"last_login_ip,omitempty"` // 上次登录 IP
	// 用户套餐
	PlanID         *uint      `gorm:"index" json:"plan_id,omitempty"`        // 当前套餐ID
	Plan           *Plan      `gorm:"foreignKey:PlanID" json:"plan,omitempty"`
	PlanStartAt    *time.Time `json:"plan_start_at,omitempty"`               // 套餐开始时间
	PlanExpireAt   *time.Time `json:"plan_expire_at,omitempty"`              // 套餐到期时间
	PlanTrafficUsed int64     `gorm:"default:0" json:"plan_traffic_used"`    // 套餐周期内已用流量
	// 用户流量配额 (聚合用户拥有的所有资源流量，可被套餐覆盖)
	TrafficQuota   int64     `gorm:"default:0" json:"traffic_quota"`       // 流量配额 (bytes), 0=无限制
	QuotaUsed      int64     `gorm:"default:0" json:"quota_used"`          // 本周期已用流量
	QuotaResetDay  int       `gorm:"default:1" json:"quota_reset_day"`     // 每月重置日 (1-28)
	QuotaResetAt   time.Time `json:"quota_reset_at"`                       // 上次重置时间
	QuotaExceeded  bool      `gorm:"default:false" json:"quota_exceeded"`  // 是否超限
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

// Plan 套餐
type Plan struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	Name          string    `gorm:"size:100;not null" json:"name"`           // 套餐名称
	Description   string    `gorm:"size:255" json:"description"`             // 套餐描述
	TrafficQuota  int64     `gorm:"default:0" json:"traffic_quota"`          // 流量配额 (bytes), 0=无限制
	SpeedLimit    int64     `gorm:"default:0" json:"speed_limit"`            // 速度限制 (bytes/s), 0=不限速
	Duration      int       `gorm:"default:30" json:"duration"`              // 有效期 (天), 0=永久
	MaxNodes      int       `gorm:"default:0" json:"max_nodes"`              // 最大节点数, 0=无限制
	MaxClients    int       `gorm:"default:0" json:"max_clients"`            // 最大客户端数, 0=无限制
	Enabled       bool      `gorm:"default:true" json:"enabled"`             // 是否启用
	SortOrder     int       `gorm:"default:0" json:"sort_order"`             // 排序顺序
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// TrafficHistory 流量历史记录
type TrafficHistory struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	NodeID     *uint     `gorm:"index" json:"node_id,omitempty"`
	TrafficIn  int64     `gorm:"default:0" json:"traffic_in"`
	TrafficOut int64     `gorm:"default:0" json:"traffic_out"`
	Connections int      `gorm:"default:0" json:"connections"`
	RecordedAt time.Time `gorm:"index" json:"recorded_at"`
}

// NotifyChannel 通知渠道配置
type NotifyChannel struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:100;not null" json:"name"`          // 渠道名称
	Type      string    `gorm:"size:20;not null" json:"type"`           // telegram/webhook/smtp
	Config    string    `gorm:"type:text" json:"config"`                // JSON 配置
	Enabled   bool      `gorm:"default:true" json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TelegramConfig Telegram 配置
type TelegramConfig struct {
	BotToken string `json:"bot_token"`
	ChatID   string `json:"chat_id"`
}

// WebhookConfig Webhook 配置
type WebhookConfig struct {
	URL     string            `json:"url"`
	Method  string            `json:"method"`  // POST/GET
	Headers map[string]string `json:"headers"` // 自定义头
}

// SMTPConfig SMTP 邮件配置
type SMTPConfig struct {
	Host     string `json:"smtp_host"`
	Port     int    `json:"smtp_port"`
	Username string `json:"username"`
	Password string `json:"password"`
	From     string `json:"from"`
	To       string `json:"to"` // 多个邮箱用逗号分隔
	UseTLS   bool   `json:"use_tls"`
}

// AlertRule 告警规则
type AlertRule struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:100;not null" json:"name"`
	Type        string    `gorm:"size:50;not null" json:"type"`          // node_offline/quota_exceeded/traffic_spike
	Condition   string    `gorm:"type:text" json:"condition"`            // JSON 条件配置
	ChannelIDs  string    `gorm:"size:255" json:"channel_ids"`           // 通知渠道 ID，逗号分隔
	Enabled     bool      `gorm:"default:true" json:"enabled"`
	CooldownMin int       `gorm:"default:30" json:"cooldown_min"`        // 告警冷却时间（分钟）
	LastAlertAt time.Time `json:"last_alert_at"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// AlertLog 告警记录
type AlertLog struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	RuleID     uint      `gorm:"index" json:"rule_id"`
	RuleName   string    `gorm:"size:100" json:"rule_name"`
	Type       string    `gorm:"size:50" json:"type"`
	Message    string    `gorm:"type:text" json:"message"`
	TargetType string    `gorm:"size:20" json:"target_type"`             // node/client
	TargetID   uint      `json:"target_id"`
	TargetName string    `gorm:"size:100" json:"target_name"`
	Status     string    `gorm:"size:20;default:sent" json:"status"`     // sent/failed
	CreatedAt  time.Time `gorm:"index" json:"created_at"`
}

// OperationLog 操作日志
type OperationLog struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	UserID     uint      `gorm:"index" json:"user_id"`
	Username   string    `gorm:"size:100" json:"username"`
	Action     string    `gorm:"size:50;index" json:"action"`           // login/create/update/delete
	Resource   string    `gorm:"size:50;index" json:"resource"`         // node/client/user/port_forward/etc
	ResourceID uint      `json:"resource_id"`
	Detail     string    `gorm:"type:text" json:"detail"`               // 操作详情 JSON
	IP         string    `gorm:"size:50" json:"ip"`                     // 客户端 IP
	UserAgent  string    `gorm:"size:255" json:"user_agent"`
	Status     string    `gorm:"size:20;default:success" json:"status"` // success/failed
	CreatedAt  time.Time `gorm:"index" json:"created_at"`
}

// Bypass 分流规则 (域名/IP 白名单或黑名单)
type Bypass struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:100;not null" json:"name"`
	Whitelist bool      `gorm:"default:false" json:"whitelist"` // true=白名单模式, false=黑名单模式
	Matchers  string    `gorm:"type:text" json:"matchers"`      // JSON 数组: ["*.google.com", "10.0.0.0/8"]
	NodeID    *uint     `gorm:"index" json:"node_id,omitempty"` // 关联节点 (可选，nil=全局)
	OwnerID   *uint     `gorm:"index" json:"owner_id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Admission 准入控制 (连接 IP 白名单或黑名单)
type Admission struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:100;not null" json:"name"`
	Whitelist bool      `gorm:"default:false" json:"whitelist"` // true=白名单模式, false=黑名单模式
	Matchers  string    `gorm:"type:text" json:"matchers"`      // JSON 数组: ["192.168.0.0/16"]
	NodeID    *uint     `gorm:"index" json:"node_id,omitempty"`
	OwnerID   *uint     `gorm:"index" json:"owner_id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// HostMapping 自定义主机映射 (类似 /etc/hosts)
type HostMapping struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:100;not null" json:"name"`
	Mappings  string    `gorm:"type:text" json:"mappings"` // JSON 数组: [{"hostname":"example.com","ip":"1.2.3.4"}]
	NodeID    *uint     `gorm:"index" json:"node_id,omitempty"`
	OwnerID   *uint     `gorm:"index" json:"owner_id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Ingress 反向代理域名路由
type Ingress struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:100;not null" json:"name"`
	Rules     string    `gorm:"type:text" json:"rules"` // JSON: [{"hostname":"example.com","endpoint":"192.168.1.1:8080"}]
	NodeID    *uint     `gorm:"index" json:"node_id,omitempty"`
	OwnerID   *uint     `gorm:"index" json:"owner_id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Recorder 流量记录器
type Recorder struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:100;not null" json:"name"`
	Type      string    `gorm:"size:20;default:file" json:"type"` // file, redis, http
	Config    string    `gorm:"type:text" json:"config"`          // JSON 配置
	NodeID    *uint     `gorm:"index" json:"node_id,omitempty"`
	OwnerID   *uint     `gorm:"index" json:"owner_id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Router 路由管理
type Router struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:100;not null" json:"name"`
	Routes    string    `gorm:"type:text" json:"routes"` // JSON: [{"net":"192.168.0.0/16","gateway":"192.168.0.1"}]
	NodeID    *uint     `gorm:"index" json:"node_id,omitempty"`
	OwnerID   *uint     `gorm:"index" json:"owner_id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// SD 服务发现
type SD struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:100;not null" json:"name"`
	Type      string    `gorm:"size:20;default:http" json:"type"` // http, consul, etcd, redis
	Config    string    `gorm:"type:text" json:"config"`          // JSON 配置
	NodeID    *uint     `gorm:"index" json:"node_id,omitempty"`
	OwnerID   *uint     `gorm:"index" json:"owner_id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ConfigVersion 配置版本历史
type ConfigVersion struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	NodeID    uint      `gorm:"index;not null" json:"node_id"`
	Config    string    `gorm:"type:text;not null" json:"config"` // YAML 配置快照
	Comment   string    `gorm:"size:255" json:"comment"`          // 版本说明
	CreatedAt time.Time `json:"created_at"`
}

// SiteConfig 网站配置
type SiteConfig struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Key       string    `gorm:"size:100;uniqueIndex;not null" json:"key"`
	Value     string    `gorm:"type:text" json:"value"`
	UpdatedAt time.Time `json:"updated_at"`
}

// InitDB 初始化数据库
func InitDB(dbPath string) (*gorm.DB, error) {
	// 确保目录存在
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, err
	}

	// 自动迁移
	if err := db.AutoMigrate(&Node{}, &Client{}, &Service{}, &User{}, &Plan{}, &TrafficHistory{}, &NotifyChannel{}, &AlertRule{}, &AlertLog{}, &PortForward{}, &NodeGroup{}, &NodeGroupMember{}, &DNSConfig{}, &OperationLog{}, &ProxyChain{}, &ProxyChainHop{}, &Tunnel{}, &SiteConfig{}, &Tag{}, &NodeTag{}, &Bypass{}, &Admission{}, &HostMapping{}, &Ingress{}, &Recorder{}, &Router{}, &SD{}, &ConfigVersion{}); err != nil {
		return nil, err
	}

	// 创建默认管理员
	var count int64
	db.Model(&User{}).Count(&count)
	if count == 0 {
		db.Create(&User{
			Username:      "admin",
			Email:         nil, // 空邮箱使用 nil
			Password:      hashPassword("admin123"), // 默认密码
			Role:          "admin",
			Enabled:       true,
			EmailVerified: true, // 默认管理员自动验证
		})
	}

	// 初始化默认系统配置
	initDefaultSiteConfigs(db)

	return db, nil
}

func hashPassword(password string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return ""
	}
	return string(hash)
}

// HashPassword 导出的密码哈希函数，供其他包使用
func HashPassword(password string) string {
	return hashPassword(password)
}

// CheckPassword 验证密码是否匹配
func CheckPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

// ValidatePasswordStrength 验证密码强度
// 要求: 至少8个字符，包含大小写字母、数字和特殊字符
func ValidatePasswordStrength(password string) error {
	if len(password) < 8 {
		return PasswordTooShortError
	}

	// 检查常见弱密码
	if isCommonPassword(password) {
		return PasswordTooCommonError
	}

	var hasUpper, hasLower, hasDigit, hasSpecial bool
	for _, c := range password {
		switch {
		case c >= 'A' && c <= 'Z':
			hasUpper = true
		case c >= 'a' && c <= 'z':
			hasLower = true
		case c >= '0' && c <= '9':
			hasDigit = true
		case isSpecialChar(c):
			hasSpecial = true
		}
	}

	if !hasUpper || !hasLower || !hasDigit {
		return PasswordTooWeakError
	}

	// 特殊字符是可选的，但如果密码少于12位则需要
	if len(password) < 12 && !hasSpecial {
		return PasswordNeedsSpecialError
	}

	return nil
}

// isSpecialChar 检查是否为特殊字符
func isSpecialChar(c rune) bool {
	specialChars := "!@#$%^&*()_+-=[]{}|;':\",./<>?`~"
	for _, sc := range specialChars {
		if c == sc {
			return true
		}
	}
	return false
}

// isCommonPassword 检查是否为常见弱密码
func isCommonPassword(password string) bool {
	commonPasswords := []string{
		"password", "12345678", "123456789", "1234567890",
		"qwerty123", "abc12345", "password1", "password123",
		"admin123", "admin1234", "letmein123", "welcome1",
		"monkey123", "dragon123", "master123", "qwertyui",
		"iloveyou1", "sunshine1", "princess1", "football1",
		"baseball1", "trustno1", "batman123", "shadow123",
	}
	lowerPass := toLowerSimple(password)
	for _, cp := range commonPasswords {
		if lowerPass == cp {
			return true
		}
	}
	return false
}

// toLowerSimple 简单的小写转换
func toLowerSimple(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			result[i] = c + 32
		} else {
			result[i] = c
		}
	}
	return string(result)
}

// 密码验证错误
var (
	PasswordTooShortError      = &ValidationError{Message: "密码长度至少8位"}
	PasswordTooWeakError       = &ValidationError{Message: "密码必须包含大写字母、小写字母和数字"}
	PasswordNeedsSpecialError  = &ValidationError{Message: "密码少于12位时必须包含特殊字符 (!@#$%^&*等)"}
	PasswordTooCommonError     = &ValidationError{Message: "密码过于常见，请使用更复杂的密码"}
)

// ValidationError 验证错误
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

// 系统配置键常量
const (
	ConfigRegistrationEnabled    = "registration_enabled"     // 是否开放注册
	ConfigEmailVerificationRequired = "email_verification_required" // 是否需要邮件验证
	ConfigDefaultRole            = "default_role"             // 新用户默认角色
	ConfigSiteName               = "site_name"                // 站点名称
	ConfigSiteURL                = "site_url"                 // 站点 URL（用于邮件链接）
	ConfigAgentAutoUpdate        = "agent_auto_update"        // Agent 自动更新开关
	ConfigAgentForceUpdate       = "agent_force_update"       // 强制所有 Agent 更新
)

// initDefaultSiteConfigs 初始化默认系统配置
func initDefaultSiteConfigs(db *gorm.DB) {
	defaultConfigs := map[string]string{
		ConfigRegistrationEnabled:       "false",
		ConfigEmailVerificationRequired: "true",
		ConfigDefaultRole:               "user",
		ConfigSiteName:                  "GOST Panel",
		ConfigSiteURL:                   "",
		ConfigAgentAutoUpdate:           "true",
		ConfigAgentForceUpdate:          "false",
	}

	for key, value := range defaultConfigs {
		var config SiteConfig
		if db.Where("key = ?", key).First(&config).Error != nil {
			db.Create(&SiteConfig{Key: key, Value: value})
		}
	}
}
