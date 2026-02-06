package api

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/AliceNetworks/gost-panel/internal/config"
	"github.com/AliceNetworks/gost-panel/internal/model"
	"github.com/AliceNetworks/gost-panel/internal/notify"
	"github.com/AliceNetworks/gost-panel/internal/service"
	"github.com/golang-jwt/jwt/v5"
)

//go:embed all:dist
var staticFS embed.FS

type Server struct {
	svc          *service.Service
	cfg          *config.Config
	router       *gin.Engine
	loginLimiter *RateLimiter
	audit        *AuditLogger
	wsHub        *WSHub
}

func NewServer(svc *service.Service, cfg *config.Config) *Server {
	if !cfg.Debug {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	// 禁用自动重定向 (防止 SPA 路由循环)
	r.RedirectTrailingSlash = false
	r.RedirectFixedPath = false

	// CORS 配置 - 根据环境和配置决定策略
	corsConfig := cors.Config{
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}

	if len(cfg.AllowedOrigins) > 0 {
		// 使用配置的允许来源
		corsConfig.AllowOrigins = cfg.AllowedOrigins
	} else if cfg.Debug {
		// 调试模式：允许 localhost
		corsConfig.AllowOriginFunc = func(origin string) bool {
			return strings.HasPrefix(origin, "http://localhost") ||
				strings.HasPrefix(origin, "http://127.0.0.1") ||
				strings.HasPrefix(origin, "https://localhost") ||
				strings.HasPrefix(origin, "https://127.0.0.1")
		}
	} else {
		// 生产模式：仅允许同源请求
		corsConfig.AllowOriginFunc = func(origin string) bool {
			// 生产环境默认拒绝跨域请求
			// 用户应通过 ALLOWED_ORIGINS 环境变量配置允许的域名
			return false
		}
	}

	r.Use(cors.New(corsConfig))

	// 设置 WebSocket 允许的来源
	SetWSOrigins(cfg.AllowedOrigins, cfg.Debug)

	s := &Server{
		svc:          svc,
		cfg:          cfg,
		router:       r,
		loginLimiter: NewRateLimiter(5, time.Minute, 5*time.Minute), // 每分钟5次，封锁5分钟
		audit:        NewAuditLogger(svc),
		wsHub:        NewWSHub(),
	}

	// Start WebSocket hub
	go s.wsHub.Run()

	// 初始化默认网站配置
	s.svc.InitDefaultSiteConfigs()

	s.setupRoutes()

	return s
}

func (s *Server) setupRoutes() {
	// Prometheus 指标中间件
	s.router.Use(PrometheusMiddleware())

	// Prometheus 指标端点 (公开)
	s.router.GET("/metrics", MetricsHandler())

	// API 路由
	api := s.router.Group("/api")
	{
		// 健康检查 (公开)
		api.GET("/health", s.healthCheck)

		// 公开接口 (带限流)
		api.POST("/login", RateLimitMiddleware(s.loginLimiter), s.login)
		api.GET("/site-config", s.getPublicSiteConfig) // 公开的网站配置

		// 用户注册和验证 (公开，带限流)
		api.POST("/register", RateLimitMiddleware(s.loginLimiter), s.register)
		api.POST("/verify-email", s.verifyEmail)
		api.POST("/forgot-password", RateLimitMiddleware(s.loginLimiter), s.forgotPassword)
		api.POST("/reset-password", s.resetPassword)
		api.GET("/registration-status", s.getRegistrationStatus)

		// 需要认证的接口
		auth := api.Group("")
		auth.Use(s.authMiddleware())
		{
			// 统计
			auth.GET("/stats", s.getStats)

			// 全局搜索
			auth.GET("/search", s.globalSearch)

			// 节点管理
			auth.GET("/nodes", s.listNodes)
			auth.GET("/nodes/paginated", s.listNodesPaginated)
			auth.POST("/nodes", s.createNode)
			auth.GET("/nodes/:id", s.getNode)
			auth.PUT("/nodes/:id", s.updateNode)
			auth.DELETE("/nodes/:id", s.deleteNode)
			auth.POST("/nodes/:id/apply", s.applyNodeConfig)
			auth.POST("/nodes/:id/sync", s.syncNodeConfig)
			auth.GET("/nodes/:id/gost-config", s.getNodeGostConfig)
			auth.GET("/nodes/:id/proxy-uri", s.getNodeProxyURI)
			auth.GET("/nodes/:id/install-script", s.getNodeInstallScript)
			auth.GET("/nodes/:id/ping", s.pingNode)
			auth.GET("/nodes/ping", s.pingAllNodes)

			// 节点批量操作
			auth.POST("/nodes/batch-delete", s.batchDeleteNodes)
			auth.POST("/nodes/batch-sync", s.batchSyncNodes)

			// 客户端管理
			auth.GET("/clients", s.listClients)
			auth.GET("/clients/paginated", s.listClientsPaginated)
			auth.POST("/clients", s.createClient)
			auth.GET("/clients/:id", s.getClient)
			auth.PUT("/clients/:id", s.updateClient)
			auth.DELETE("/clients/:id", s.deleteClient)
			auth.GET("/clients/:id/install-script", s.getClientInstallScript)
			auth.GET("/clients/:id/gost-config", s.getClientGostConfig)
			auth.GET("/clients/:id/proxy-uri", s.getClientProxyURI)

			// 客户端批量操作
			auth.POST("/clients/batch-delete", s.batchDeleteClients)

			// 用户管理
			auth.GET("/users", s.listUsers)
			auth.POST("/users", s.createUser)
			auth.GET("/users/:id", s.getUser)
			auth.PUT("/users/:id", s.updateUser)
			auth.DELETE("/users/:id", s.deleteUser)
			auth.POST("/change-password", s.changePassword)

			// 个人账户设置
			auth.GET("/profile", s.getProfile)
			auth.PUT("/profile", s.updateProfile)

			// 流量历史
			auth.GET("/traffic-history", s.getTrafficHistory)

			// 通知渠道管理
			auth.GET("/notify-channels", s.listNotifyChannels)
			auth.POST("/notify-channels", s.createNotifyChannel)
			auth.GET("/notify-channels/:id", s.getNotifyChannel)
			auth.PUT("/notify-channels/:id", s.updateNotifyChannel)
			auth.DELETE("/notify-channels/:id", s.deleteNotifyChannel)
			auth.POST("/notify-channels/:id/test", s.testNotifyChannel)

			// 告警规则管理
			auth.GET("/alert-rules", s.listAlertRules)
			auth.POST("/alert-rules", s.createAlertRule)
			auth.GET("/alert-rules/:id", s.getAlertRule)
			auth.PUT("/alert-rules/:id", s.updateAlertRule)
			auth.DELETE("/alert-rules/:id", s.deleteAlertRule)

			// 告警日志
			auth.GET("/alert-logs", s.getAlertLogs)

			// 操作日志
			auth.GET("/operation-logs", s.getOperationLogs)

			// 数据导出/导入
			auth.GET("/export", s.exportData)
			auth.POST("/import", s.importData)

			// 数据库备份/恢复
			auth.GET("/backup", s.backupDatabase)
			auth.POST("/restore", s.restoreDatabase)

			// 端口转发
			auth.GET("/port-forwards", s.listPortForwards)
			auth.POST("/port-forwards", s.createPortForward)
			auth.GET("/port-forwards/:id", s.getPortForward)
			auth.PUT("/port-forwards/:id", s.updatePortForward)
			auth.DELETE("/port-forwards/:id", s.deletePortForward)

			// 节点组 (负载均衡)
			auth.GET("/node-groups", s.listNodeGroups)
			auth.POST("/node-groups", s.createNodeGroup)
			auth.GET("/node-groups/:id", s.getNodeGroup)
			auth.PUT("/node-groups/:id", s.updateNodeGroup)
			auth.DELETE("/node-groups/:id", s.deleteNodeGroup)
			auth.GET("/node-groups/:id/members", s.listNodeGroupMembers)
			auth.POST("/node-groups/:id/members", s.addNodeGroupMember)
			auth.DELETE("/node-groups/:id/members/:memberId", s.removeNodeGroupMember)
			auth.GET("/node-groups/:id/config", s.getNodeGroupConfig)

			// 代理链/隧道转发
			auth.GET("/proxy-chains", s.listProxyChains)
			auth.POST("/proxy-chains", s.createProxyChain)
			auth.GET("/proxy-chains/:id", s.getProxyChain)
			auth.PUT("/proxy-chains/:id", s.updateProxyChain)
			auth.DELETE("/proxy-chains/:id", s.deleteProxyChain)
			auth.GET("/proxy-chains/:id/hops", s.listProxyChainHops)
			auth.POST("/proxy-chains/:id/hops", s.addProxyChainHop)
			auth.PUT("/proxy-chains/:id/hops/:hopId", s.updateProxyChainHop)
			auth.DELETE("/proxy-chains/:id/hops/:hopId", s.removeProxyChainHop)
			auth.GET("/proxy-chains/:id/config", s.getProxyChainConfig)

			// 隧道转发 (入口-出口模式)
			auth.GET("/tunnels", s.listTunnels)
			auth.POST("/tunnels", s.createTunnel)
			auth.GET("/tunnels/:id", s.getTunnel)
			auth.PUT("/tunnels/:id", s.updateTunnel)
			auth.DELETE("/tunnels/:id", s.deleteTunnel)
			auth.POST("/tunnels/:id/sync", s.syncTunnel)
			auth.GET("/tunnels/:id/entry-config", s.getTunnelEntryConfig)
			auth.GET("/tunnels/:id/exit-config", s.getTunnelExitConfig)

			// 预配置模板
			auth.GET("/templates", s.listTemplates)
			auth.GET("/templates/categories", s.getTemplateCategories)
			auth.GET("/templates/:id", s.getTemplate)

			// 客户端模板
			auth.GET("/client-templates", s.listClientTemplates)
			auth.GET("/client-templates/categories", s.getClientTemplateCategories)
			auth.GET("/client-templates/:id", s.getClientTemplate)

			// 网站配置 (仅管理员)
			auth.GET("/site-configs", s.getSiteConfigs)
			auth.PUT("/site-configs", s.updateSiteConfigs)

			// 节点标签管理
			auth.GET("/tags", s.listTags)
			auth.GET("/tags/:id", s.getTag)
			auth.POST("/tags", s.createTag)
			auth.PUT("/tags/:id", s.updateTag)
			auth.DELETE("/tags/:id", s.deleteTag)
			auth.GET("/tags/:id/nodes", s.getNodesByTag)

			// 节点的标签操作
			auth.GET("/nodes/:id/tags", s.getNodeTags)
			auth.POST("/nodes/:id/tags", s.addNodeTag)
			auth.PUT("/nodes/:id/tags", s.setNodeTags)
			auth.DELETE("/nodes/:id/tags/:tagId", s.removeNodeTag)

			// 管理员用户操作
			auth.POST("/users/:id/verify-email", s.adminVerifyUserEmail)
			auth.POST("/users/:id/resend-verification", s.resendVerificationEmail)
		}
	}

	// Agent 接口 (使用 Token 认证)
	agent := s.router.Group("/agent")
	{
		agent.POST("/register", s.agentRegister)
		agent.POST("/heartbeat", s.agentHeartbeat)
		agent.GET("/config/:token", s.agentGetConfig)
		agent.GET("/version", s.agentGetVersion)
		agent.GET("/check-update", s.agentCheckUpdate)
		agent.GET("/download/:os/:arch", s.agentDownload)
		// 客户端心跳 (通过 token 认证)
		agent.POST("/client-heartbeat/:token", s.clientHeartbeat)
	}

	// WebSocket 接口
	s.router.GET("/ws", s.handleWebSocket)

	// 安装脚本接口 (公开)
	scripts := s.router.Group("/scripts")
	{
		scripts.GET("/install-node.sh", s.serveInstallScript("install-node.sh"))
		scripts.GET("/install-client.sh", s.serveInstallScript("install-client.sh"))
		scripts.GET("/install-node.ps1", s.serveInstallScript("install-node.ps1"))
		scripts.GET("/install-client.ps1", s.serveInstallScript("install-client.ps1"))
		// 动态生成的安装脚本 (通过 token 认证)
		scripts.GET("/client/:token", s.serveClientScript)
	}

	// 静态文件 (前端)
	s.setupStaticFiles()
}

func (s *Server) setupStaticFiles() {
	// 尝试加载嵌入的静态文件
	subFS, err := fs.Sub(staticFS, "dist")
	if err != nil {
		return
	}

	// 静态资源文件 - 手动处理以确保正确的 MIME 类型
	s.router.GET("/assets/*filepath", func(c *gin.Context) {
		fp := c.Param("filepath")
		path := "assets" + fp
		data, err := fs.ReadFile(subFS, path)
		if err != nil {
			c.Status(http.StatusNotFound)
			return
		}

		// 根据扩展名设置 MIME 类型
		contentType := "application/octet-stream"
		switch {
		case strings.HasSuffix(fp, ".js"):
			contentType = "application/javascript; charset=utf-8"
		case strings.HasSuffix(fp, ".css"):
			contentType = "text/css; charset=utf-8"
		case strings.HasSuffix(fp, ".svg"):
			contentType = "image/svg+xml"
		case strings.HasSuffix(fp, ".png"):
			contentType = "image/png"
		case strings.HasSuffix(fp, ".jpg"), strings.HasSuffix(fp, ".jpeg"):
			contentType = "image/jpeg"
		case strings.HasSuffix(fp, ".woff2"):
			contentType = "font/woff2"
		case strings.HasSuffix(fp, ".woff"):
			contentType = "font/woff"
		}

		c.Data(http.StatusOK, contentType, data)
	})

	// vite.svg
	s.router.GET("/vite.svg", func(c *gin.Context) {
		data, _ := fs.ReadFile(subFS, "vite.svg")
		c.Data(http.StatusOK, "image/svg+xml", data)
	})

	// 首页
	s.router.GET("/", func(c *gin.Context) {
		data, _ := fs.ReadFile(subFS, "index.html")
		c.Data(http.StatusOK, "text/html; charset=utf-8", data)
	})

	// SPA 路由回退
	s.router.NoRoute(func(c *gin.Context) {
		data, _ := fs.ReadFile(subFS, "index.html")
		c.Data(http.StatusOK, "text/html; charset=utf-8", data)
	})
}

func (s *Server) Run() error {
	return s.router.Run(s.cfg.ListenAddr)
}

// ==================== 中间件 ====================

func (s *Server) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := c.GetHeader("Authorization")
		if tokenStr == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
			c.Abort()
			return
		}

		// 移除 "Bearer " 前缀
		if len(tokenStr) > 7 && tokenStr[:7] == "Bearer " {
			tokenStr = tokenStr[7:]
		}

		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			// 验证签名方法，防止算法替换攻击
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(s.cfg.JWTSecret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		claims := token.Claims.(jwt.MapClaims)
		c.Set("user_id", claims["user_id"])
		c.Set("username", claims["username"])
		c.Set("role", claims["role"])
		c.Next()
	}
}

// ==================== 认证接口 ====================

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (s *Server) login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := s.svc.ValidateUser(req.Username, req.Password)
	if err != nil {
		// 记录登录失败
		s.svc.LogOperation(0, req.Username, "login", "user", 0, "login failed", c.ClientIP(), c.GetHeader("User-Agent"), "failed")
		RecordLoginAttempt(false)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	// 检查账户是否启用
	if !user.Enabled {
		s.svc.LogOperation(user.ID, user.Username, "login", "user", user.ID, "account disabled", c.ClientIP(), c.GetHeader("User-Agent"), "failed")
		c.JSON(http.StatusForbidden, gin.H{"error": "account is disabled"})
		return
	}

	// 检查邮箱是否已验证（如果需要）
	if s.svc.IsEmailVerificationRequired() && !user.EmailVerified && user.Email != nil && *user.Email != "" {
		c.JSON(http.StatusForbidden, gin.H{"error": "email not verified", "code": "EMAIL_NOT_VERIFIED"})
		return
	}

	// 登录成功，重置限流计数
	s.loginLimiter.Reset(c.ClientIP())
	RecordLoginAttempt(true)

	// 更新登录信息
	s.svc.UpdateUserLoginInfo(user.ID, c.ClientIP())

	// 记录登录成功
	s.svc.LogOperation(user.ID, user.Username, "login", "user", user.ID, "login success", c.ClientIP(), c.GetHeader("User-Agent"), "success")

	// 生成 JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"role":     user.Role,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	})

	tokenString, err := token.SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": tokenString,
		"user": gin.H{
			"id":               user.ID,
			"username":         user.Username,
			"email":            user.Email,
			"role":             user.Role,
			"email_verified":   user.EmailVerified,
			"password_changed": user.PasswordChanged,
		},
	})
}

// ==================== 用户注册与验证 ====================

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

func (s *Server) register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := s.svc.RegisterUser(req.Username, req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 如果需要邮件验证，发送验证邮件
	if !user.EmailVerified && user.VerificationToken != "" {
		go s.sendVerificationEmail(user)
	}

	// 记录注册操作
	s.svc.LogOperation(user.ID, user.Username, "register", "user", user.ID, "user registered", c.ClientIP(), c.GetHeader("User-Agent"), "success")

	c.JSON(http.StatusOK, gin.H{
		"message":            "Registration successful",
		"email_verification": !user.EmailVerified,
	})
}

func (s *Server) verifyEmail(c *gin.Context) {
	var req struct {
		Token string `json:"token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := s.svc.VerifyEmail(req.Token)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 发送欢迎邮件
	go s.sendWelcomeEmail(user)

	c.JSON(http.StatusOK, gin.H{"message": "Email verified successfully"})
}

// ForgotPasswordRequest 忘记密码请求
type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

func (s *Server) forgotPassword(c *gin.Context) {
	var req ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, token, err := s.svc.RequestPasswordReset(req.Email)
	if err != nil {
		// 为了安全，不暴露邮箱是否存在
		c.JSON(http.StatusOK, gin.H{"message": "If the email exists, a reset link has been sent"})
		return
	}

	// 发送密码重置邮件
	go s.sendPasswordResetEmail(user, token)

	c.JSON(http.StatusOK, gin.H{"message": "If the email exists, a reset link has been sent"})
}

// ResetPasswordRequest 重置密码请求
type ResetPasswordRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

func (s *Server) resetPassword(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := s.svc.ResetPassword(req.Token, req.NewPassword); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password reset successfully"})
}

func (s *Server) getRegistrationStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"enabled":            s.svc.IsRegistrationEnabled(),
		"email_verification": s.svc.IsEmailVerificationRequired(),
	})
}

// 管理员手动验证用户邮箱
func (s *Server) adminVerifyUserEmail(c *gin.Context) {
	_, isAdmin := getUserInfo(c)
	if !isAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "admin only"})
		return
	}

	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	if err := s.svc.UpdateUser(uint(id), map[string]interface{}{
		"email_verified":     true,
		"verification_token": "",
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// 重新发送验证邮件
func (s *Server) resendVerificationEmail(c *gin.Context) {
	_, isAdmin := getUserInfo(c)
	if !isAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "admin only"})
		return
	}

	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	token, err := s.svc.ResendVerificationEmail(uint(id))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, _ := s.svc.GetUser(uint(id))
	if user != nil {
		user.VerificationToken = token
		go s.sendVerificationEmail(user)
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// sendVerificationEmail 发送验证邮件
func (s *Server) sendVerificationEmail(user *model.User) {
	if user.Email == nil || *user.Email == "" {
		return
	}
	emailSender := s.getEmailSender()
	if emailSender == nil {
		return
	}
	emailSender.SendVerificationEmail(*user.Email, user.Username, user.VerificationToken)
}

// sendPasswordResetEmail 发送密码重置邮件
func (s *Server) sendPasswordResetEmail(user *model.User, token string) {
	if user.Email == nil || *user.Email == "" {
		return
	}
	emailSender := s.getEmailSender()
	if emailSender == nil {
		return
	}
	emailSender.SendPasswordResetEmail(*user.Email, user.Username, token)
}

// sendWelcomeEmail 发送欢迎邮件
func (s *Server) sendWelcomeEmail(user *model.User) {
	if user.Email == nil || *user.Email == "" {
		return
	}
	emailSender := s.getEmailSender()
	if emailSender == nil {
		return
	}
	emailSender.SendWelcomeEmail(*user.Email, user.Username)
}

// getEmailSender 获取邮件发送器
func (s *Server) getEmailSender() *notify.EmailSender {
	// 查找 SMTP 类型的通知渠道
	channels, err := s.svc.ListNotifyChannels()
	if err != nil {
		return nil
	}

	for _, ch := range channels {
		if (ch.Type == "smtp" || ch.Type == "email") && ch.Enabled {
			var smtpConfig model.SMTPConfig
			if err := json.Unmarshal([]byte(ch.Config), &smtpConfig); err != nil {
				continue
			}
			siteName := s.svc.GetSiteConfig(model.ConfigSiteName)
			siteURL := s.svc.GetSiteConfig(model.ConfigSiteURL)
			if siteName == "" {
				siteName = "GOST Panel"
			}
			return notify.NewEmailSender(&smtpConfig, siteName, siteURL)
		}
	}
	return nil
}

func (s *Server) getStats(c *gin.Context) {
	stats, err := s.svc.GetStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, stats)
}

// healthCheck 健康检查端点
func (s *Server) healthCheck(c *gin.Context) {
	// 检查数据库连接
	dbOk := true
	dbStatusStr := "ok"
	if err := s.svc.Ping(); err != nil {
		dbStatusStr = "error"
		dbOk = false
	}
	UpdateDBStatus(dbOk)

	// 获取基本统计
	stats, _ := s.svc.GetStats()
	nodeCount := 0
	onlineNodes := 0
	clientCount := 0
	onlineClients := 0
	userCount := 0
	if stats != nil {
		nodeCount = stats.TotalNodes
		onlineNodes = stats.OnlineNodes
		clientCount = stats.TotalClients
		onlineClients = stats.OnlineClients
		userCount = stats.TotalUsers
	}

	// 更新 Prometheus 指标
	UpdateNodeMetrics(nodeCount, onlineNodes)
	UpdateClientMetrics(clientCount, onlineClients)
	UpdateUserMetrics(userCount)

	c.JSON(http.StatusOK, gin.H{
		"status":         "ok",
		"database":       dbStatusStr,
		"version":        "1.4.0",
		"nodes":          nodeCount,
		"online_nodes":   onlineNodes,
		"clients":        clientCount,
		"online_clients": onlineClients,
		"users":          userCount,
	})
}

// allowedScripts 允许的脚本白名单 (防止路径遍历)
var allowedScripts = map[string]bool{
	"install-node.sh":     true,
	"install-client.sh":   true,
	"install-node.ps1":    true,
	"install-client.ps1":  true,
}

// serveInstallScript returns a handler that serves install scripts
func (s *Server) serveInstallScript(scriptName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 验证脚本名在白名单中 (防止路径遍历)
		if !allowedScripts[scriptName] {
			c.JSON(http.StatusForbidden, gin.H{"error": "script not allowed"})
			return
		}

		// Try multiple paths
		paths := []string{
			filepath.Join("scripts", scriptName),
			filepath.Join("/root/gost-panel/scripts", scriptName),
			filepath.Join(".", "scripts", scriptName),
		}

		var scriptPath string
		for _, p := range paths {
			if _, err := os.Stat(p); err == nil {
				scriptPath = p
				break
			}
		}

		if scriptPath == "" {
			c.JSON(http.StatusNotFound, gin.H{"error": "script not found"})
			return
		}

		content, err := os.ReadFile(scriptPath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Set content type based on extension
		if filepath.Ext(scriptName) == ".ps1" {
			c.Header("Content-Type", "text/plain; charset=utf-8")
		} else {
			c.Header("Content-Type", "text/x-shellscript; charset=utf-8")
		}
		c.Header("Content-Disposition", "inline; filename="+scriptName)

		c.String(http.StatusOK, string(content))
	}
}

// serveClientScript 通过 token 提供客户端安装脚本 (公开接口)
func (s *Server) serveClientScript(c *gin.Context) {
	token := c.Param("token")

	client, err := s.svc.GetClientByToken(token)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "client not found"})
		return
	}

	// 获取正确的 Panel URL
	panelURL := getPanelURL(c)

	script := fmt.Sprintf(`#!/bin/bash
# GOST 客户端安装脚本
# 客户端: %s

set -e

PANEL_URL="%s"
CLIENT_TOKEN="%s"

# 安装 GOST
echo "Installing GOST..."
bash <(curl -fsSL https://github.com/go-gost/gost/raw/master/install.sh) --install

# 创建配置目录
mkdir -p /etc/gost

# 下载配置
echo "Downloading config..."
curl -fsSL "${PANEL_URL}/agent/config/${CLIENT_TOKEN}" -o /etc/gost/gost.yml

# 创建 systemd 服务
cat > /etc/systemd/system/gost.service << 'EOF'
[Unit]
Description=GOST Tunnel Service
After=network.target

[Service]
Type=simple
ExecStart=/usr/local/bin/gost -C /etc/gost/gost.yml
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

# 创建心跳脚本
cat > /etc/gost/heartbeat.sh << HEARTBEAT
#!/bin/bash
curl -fsSL -X POST "${PANEL_URL}/agent/client-heartbeat/${CLIENT_TOKEN}" > /dev/null 2>&1
HEARTBEAT
chmod +x /etc/gost/heartbeat.sh

# 添加 cron job (每分钟心跳)
(crontab -l 2>/dev/null | grep -v "gost/heartbeat"; echo "* * * * * /etc/gost/heartbeat.sh") | crontab -

# 创建心跳 systemd timer (备用方案)
cat > /etc/systemd/system/gost-heartbeat.service << 'EOF'
[Unit]
Description=GOST Client Heartbeat

[Service]
Type=oneshot
ExecStart=/etc/gost/heartbeat.sh
EOF

cat > /etc/systemd/system/gost-heartbeat.timer << 'EOF'
[Unit]
Description=GOST Client Heartbeat Timer

[Timer]
OnBootSec=10s
OnUnitActiveSec=1m

[Install]
WantedBy=timers.target
EOF

# 启动服务
systemctl daemon-reload
systemctl enable gost gost-heartbeat.timer
systemctl start gost gost-heartbeat.timer

# 发送首次心跳
/etc/gost/heartbeat.sh

echo "GOST client installed successfully!"
echo "Local SOCKS5 port: %d (credentials configured in panel)"
`, client.Name, panelURL, client.Token, client.LocalPort)

	c.Header("Content-Type", "text/x-shellscript; charset=utf-8")
	c.String(http.StatusOK, script)
}
