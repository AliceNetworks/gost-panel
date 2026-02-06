package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// NodeTemplate 节点预配置模板
type NodeTemplate struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Category    string            `json:"category"` // basic, secure, tunnel, advanced
	Icon        string            `json:"icon"`
	Defaults    NodeTemplateDefaults `json:"defaults"`
}

// NodeTemplateDefaults 模板默认值
type NodeTemplateDefaults struct {
	Protocol       string `json:"protocol"`
	Transport      string `json:"transport"`
	Port           int    `json:"port"`
	TLSEnabled     bool   `json:"tls_enabled"`
	WSPath         string `json:"ws_path,omitempty"`
	SSMethod       string `json:"ss_method,omitempty"`
	RequireAuth    bool   `json:"require_auth"`
	RequireTLS     bool   `json:"require_tls"`
	RequireUUID    bool   `json:"require_uuid"`
	RequirePassword bool  `json:"require_password"`
}

// 预置模板列表
var nodeTemplates = []NodeTemplate{
	// ===== 基础代理 =====
	{
		ID:          "socks5-basic",
		Name:        "SOCKS5 基础代理",
		Description: "最简单的 SOCKS5 代理，适合内网或可信网络",
		Category:    "basic",
		Icon:        "server",
		Defaults: NodeTemplateDefaults{
			Protocol:    "socks5",
			Transport:   "tcp",
			Port:        1080,
			RequireAuth: true,
		},
	},
	{
		ID:          "http-basic",
		Name:        "HTTP 代理",
		Description: "HTTP/HTTPS 代理，浏览器直接可用",
		Category:    "basic",
		Icon:        "globe",
		Defaults: NodeTemplateDefaults{
			Protocol:    "http",
			Transport:   "tcp",
			Port:        8080,
			RequireAuth: true,
		},
	},

	// ===== 加密代理 =====
	{
		ID:          "socks5-tls",
		Name:        "SOCKS5 + TLS",
		Description: "TLS 加密的 SOCKS5 代理，安全可靠",
		Category:    "secure",
		Icon:        "lock",
		Defaults: NodeTemplateDefaults{
			Protocol:    "socks5",
			Transport:   "tls",
			Port:        443,
			TLSEnabled:  true,
			RequireAuth: true,
			RequireTLS:  true,
		},
	},
	{
		ID:          "socks5-ws-tls",
		Name:        "SOCKS5 + WebSocket + TLS",
		Description: "WebSocket 传输，可过 CDN，伪装性强",
		Category:    "secure",
		Icon:        "shield",
		Defaults: NodeTemplateDefaults{
			Protocol:    "socks5",
			Transport:   "wss",
			Port:        443,
			TLSEnabled:  true,
			WSPath:      "/ws",
			RequireAuth: true,
			RequireTLS:  true,
		},
	},
	{
		ID:          "shadowsocks",
		Name:        "Shadowsocks",
		Description: "经典 SS 协议，兼容各种客户端",
		Category:    "secure",
		Icon:        "flash",
		Defaults: NodeTemplateDefaults{
			Protocol:        "ss",
			Transport:       "tcp",
			Port:            8388,
			SSMethod:        "aes-256-gcm",
			RequirePassword: true,
		},
	},
	{
		ID:          "trojan",
		Name:        "Trojan",
		Description: "Trojan 协议，TLS 加密，难以检测",
		Category:    "secure",
		Icon:        "shield-checkmark",
		Defaults: NodeTemplateDefaults{
			Protocol:        "trojan",
			Transport:       "tls",
			Port:            443,
			TLSEnabled:      true,
			RequireTLS:      true,
			RequirePassword: true,
		},
	},
	{
		ID:          "trojan-ws",
		Name:        "Trojan + WebSocket",
		Description: "Trojan over WebSocket，可过 CDN",
		Category:    "secure",
		Icon:        "shield-checkmark",
		Defaults: NodeTemplateDefaults{
			Protocol:        "trojan",
			Transport:       "wss",
			Port:            443,
			TLSEnabled:      true,
			WSPath:          "/trojan",
			RequireTLS:      true,
			RequirePassword: true,
		},
	},
	{
		ID:          "vmess",
		Name:        "VMess",
		Description: "V2Ray VMess 协议",
		Category:    "secure",
		Icon:        "rocket",
		Defaults: NodeTemplateDefaults{
			Protocol:    "vmess",
			Transport:   "tcp",
			Port:        10086,
			RequireUUID: true,
		},
	},
	{
		ID:          "vmess-ws-tls",
		Name:        "VMess + WebSocket + TLS",
		Description: "V2Ray 经典配置，可过 CDN",
		Category:    "secure",
		Icon:        "rocket",
		Defaults: NodeTemplateDefaults{
			Protocol:    "vmess",
			Transport:   "wss",
			Port:        443,
			TLSEnabled:  true,
			WSPath:      "/vmess",
			RequireTLS:  true,
			RequireUUID: true,
		},
	},

	// ===== 隧道转发 =====
	{
		ID:          "relay-tcp",
		Name:        "TCP 端口转发",
		Description: "将本地端口转发到远程地址",
		Category:    "tunnel",
		Icon:        "swap-horizontal",
		Defaults: NodeTemplateDefaults{
			Protocol:  "relay",
			Transport: "tcp",
			Port:      12345,
		},
	},
	{
		ID:          "relay-udp",
		Name:        "UDP 端口转发",
		Description: "UDP 端口转发，适合游戏加速",
		Category:    "tunnel",
		Icon:        "swap-horizontal",
		Defaults: NodeTemplateDefaults{
			Protocol:  "relay",
			Transport: "udp",
			Port:      12345,
		},
	},
	{
		ID:          "rtcp-tunnel",
		Name:        "反向 TCP 隧道",
		Description: "内网穿透，将内网服务暴露到公网",
		Category:    "tunnel",
		Icon:        "git-compare",
		Defaults: NodeTemplateDefaults{
			Protocol:    "socks5",
			Transport:   "tcp",
			Port:        38567,
			RequireAuth: true,
		},
	},

	// ===== 高级配置 =====
	{
		ID:          "socks5-quic",
		Name:        "SOCKS5 + QUIC",
		Description: "QUIC 传输，低延迟，抗丢包",
		Category:    "advanced",
		Icon:        "speedometer",
		Defaults: NodeTemplateDefaults{
			Protocol:    "socks5",
			Transport:   "quic",
			Port:        443,
			TLSEnabled:  true,
			RequireAuth: true,
			RequireTLS:  true,
		},
	},
	{
		ID:          "socks5-kcp",
		Name:        "SOCKS5 + KCP",
		Description: "KCP 传输，高丢包网络优化",
		Category:    "advanced",
		Icon:        "pulse",
		Defaults: NodeTemplateDefaults{
			Protocol:    "socks5",
			Transport:   "kcp",
			Port:        29900,
			RequireAuth: true,
		},
	},
	{
		ID:          "socks5-h2",
		Name:        "SOCKS5 + HTTP/2",
		Description: "HTTP/2 传输，多路复用",
		Category:    "advanced",
		Icon:        "layers",
		Defaults: NodeTemplateDefaults{
			Protocol:    "socks5",
			Transport:   "h2",
			Port:        443,
			TLSEnabled:  true,
			RequireAuth: true,
			RequireTLS:  true,
		},
	},
	{
		ID:          "ss-relay",
		Name:        "Shadowsocks 中转",
		Description: "SS 协议 + 转发链，多跳代理",
		Category:    "advanced",
		Icon:        "git-network",
		Defaults: NodeTemplateDefaults{
			Protocol:        "ss",
			Transport:       "tcp",
			Port:            8388,
			SSMethod:        "chacha20-ietf-poly1305",
			RequirePassword: true,
		},
	},
}

// listTemplates 获取所有预配置模板
func (s *Server) listTemplates(c *gin.Context) {
	category := c.Query("category")

	if category == "" {
		c.JSON(http.StatusOK, nodeTemplates)
		return
	}

	// 按分类过滤
	filtered := make([]NodeTemplate, 0)
	for _, t := range nodeTemplates {
		if t.Category == category {
			filtered = append(filtered, t)
		}
	}

	c.JSON(http.StatusOK, filtered)
}

// getTemplate 获取单个模板详情
func (s *Server) getTemplate(c *gin.Context) {
	id := c.Param("id")

	for _, t := range nodeTemplates {
		if t.ID == id {
			c.JSON(http.StatusOK, t)
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "template not found"})
}

// getTemplateCategories 获取模板分类
func (s *Server) getTemplateCategories(c *gin.Context) {
	categories := []map[string]string{
		{"id": "basic", "name": "基础代理", "description": "简单易用的代理配置"},
		{"id": "secure", "name": "加密代理", "description": "安全加密的代理协议"},
		{"id": "tunnel", "name": "隧道转发", "description": "端口转发和内网穿透"},
		{"id": "advanced", "name": "高级配置", "description": "高级传输协议"},
	}
	c.JSON(http.StatusOK, categories)
}

// ==================== 客户端模板 ====================

// ClientTemplate 客户端预配置模板
type ClientTemplate struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Category    string                 `json:"category"` // reverse-tunnel, forward, access
	Icon        string                 `json:"icon"`
	Defaults    ClientTemplateDefaults `json:"defaults"`
}

// ClientTemplateDefaults 客户端模板默认值
type ClientTemplateDefaults struct {
	LocalPort    int    `json:"local_port"`
	RemotePort   int    `json:"remote_port"`
	Protocol     string `json:"protocol"`      // socks5, http
	Transport    string `json:"transport"`     // tcp, tls, ws, wss
	EnableTCP    bool   `json:"enable_tcp"`    // 启用 TCP (RTCP)
	EnableUDP    bool   `json:"enable_udp"`    // 启用 UDP (RUDP)
	RequireAuth  bool   `json:"require_auth"`
	TLSEnabled   bool   `json:"tls_enabled"`
	KeepAlive    bool   `json:"keepalive"`
}

// 客户端预置模板列表
var clientTemplates = []ClientTemplate{
	// ===== 反向隧道 (内网穿透) =====
	{
		ID:          "reverse-socks5-tcp-udp",
		Name:        "反向 SOCKS5 (TCP+UDP)",
		Description: "内网穿透，同时支持 TCP 和 UDP 代理，适合游戏和视频",
		Category:    "reverse-tunnel",
		Icon:        "git-compare",
		Defaults: ClientTemplateDefaults{
			LocalPort:   38777,
			RemotePort:  38777,
			Protocol:    "socks5",
			Transport:   "tcp",
			EnableTCP:   true,
			EnableUDP:   true,
			RequireAuth: true,
			KeepAlive:   true,
		},
	},
	{
		ID:          "reverse-socks5-tcp",
		Name:        "反向 SOCKS5 (仅TCP)",
		Description: "内网穿透，仅 TCP 代理，轻量级",
		Category:    "reverse-tunnel",
		Icon:        "swap-horizontal",
		Defaults: ClientTemplateDefaults{
			LocalPort:   38777,
			RemotePort:  38777,
			Protocol:    "socks5",
			Transport:   "tcp",
			EnableTCP:   true,
			EnableUDP:   false,
			RequireAuth: true,
			KeepAlive:   true,
		},
	},
	{
		ID:          "reverse-http",
		Name:        "反向 HTTP 代理",
		Description: "内网穿透，HTTP 代理，浏览器直接可用",
		Category:    "reverse-tunnel",
		Icon:        "globe",
		Defaults: ClientTemplateDefaults{
			LocalPort:   38080,
			RemotePort:  38080,
			Protocol:    "http",
			Transport:   "tcp",
			EnableTCP:   true,
			EnableUDP:   false,
			RequireAuth: true,
			KeepAlive:   true,
		},
	},
	{
		ID:          "reverse-socks5-tls",
		Name:        "反向 SOCKS5 + TLS",
		Description: "TLS 加密的反向隧道，安全可靠",
		Category:    "reverse-tunnel",
		Icon:        "lock",
		Defaults: ClientTemplateDefaults{
			LocalPort:   38777,
			RemotePort:  38777,
			Protocol:    "socks5",
			Transport:   "tls",
			EnableTCP:   true,
			EnableUDP:   true,
			RequireAuth: true,
			TLSEnabled:  true,
			KeepAlive:   true,
		},
	},

	// ===== 端口转发 =====
	{
		ID:          "forward-tcp",
		Name:        "TCP 端口转发",
		Description: "将本地端口转发到远程地址",
		Category:    "forward",
		Icon:        "arrow-forward",
		Defaults: ClientTemplateDefaults{
			LocalPort:  12345,
			RemotePort: 12345,
			Protocol:   "tcp",
			Transport:  "tcp",
			EnableTCP:  true,
			EnableUDP:  false,
			KeepAlive:  true,
		},
	},
	{
		ID:          "forward-udp",
		Name:        "UDP 端口转发",
		Description: "UDP 端口转发，适合游戏加速",
		Category:    "forward",
		Icon:        "game-controller",
		Defaults: ClientTemplateDefaults{
			LocalPort:  12345,
			RemotePort: 12345,
			Protocol:   "udp",
			Transport:  "udp",
			EnableTCP:  false,
			EnableUDP:  true,
			KeepAlive:  true,
		},
	},
	{
		ID:          "forward-tcp-udp",
		Name:        "TCP+UDP 端口转发",
		Description: "同时转发 TCP 和 UDP 端口",
		Category:    "forward",
		Icon:        "swap-horizontal",
		Defaults: ClientTemplateDefaults{
			LocalPort:  12345,
			RemotePort: 12345,
			Protocol:   "relay",
			Transport:  "tcp",
			EnableTCP:  true,
			EnableUDP:  true,
			KeepAlive:  true,
		},
	},

	// ===== 访问控制 =====
	{
		ID:          "access-local-socks5",
		Name:        "本地 SOCKS5 入口",
		Description: "在本地开启 SOCKS5 代理入口，用于连接远程节点",
		Category:    "access",
		Icon:        "enter",
		Defaults: ClientTemplateDefaults{
			LocalPort:   1080,
			Protocol:    "socks5",
			Transport:   "tcp",
			RequireAuth: false,
		},
	},
	{
		ID:          "access-local-http",
		Name:        "本地 HTTP 入口",
		Description: "在本地开启 HTTP 代理入口，浏览器直接可用",
		Category:    "access",
		Icon:        "globe",
		Defaults: ClientTemplateDefaults{
			LocalPort:   8080,
			Protocol:    "http",
			Transport:   "tcp",
			RequireAuth: false,
		},
	},
}

// listClientTemplates 获取客户端模板列表
func (s *Server) listClientTemplates(c *gin.Context) {
	category := c.Query("category")

	if category == "" {
		c.JSON(http.StatusOK, clientTemplates)
		return
	}

	// 按分类过滤
	filtered := make([]ClientTemplate, 0)
	for _, t := range clientTemplates {
		if t.Category == category {
			filtered = append(filtered, t)
		}
	}

	c.JSON(http.StatusOK, filtered)
}

// getClientTemplate 获取单个客户端模板详情
func (s *Server) getClientTemplate(c *gin.Context) {
	id := c.Param("id")

	for _, t := range clientTemplates {
		if t.ID == id {
			c.JSON(http.StatusOK, t)
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "template not found"})
}

// getClientTemplateCategories 获取客户端模板分类
func (s *Server) getClientTemplateCategories(c *gin.Context) {
	categories := []map[string]string{
		{"id": "reverse-tunnel", "name": "反向隧道", "description": "内网穿透，将内网服务暴露到公网"},
		{"id": "forward", "name": "端口转发", "description": "TCP/UDP 端口转发"},
		{"id": "access", "name": "访问入口", "description": "本地代理入口配置"},
	}
	c.JSON(http.StatusOK, categories)
}
