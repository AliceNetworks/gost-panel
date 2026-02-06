package api

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// HTTP 请求计数器
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gost_panel_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	// HTTP 请求延迟
	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "gost_panel_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1, 5},
		},
		[]string{"method", "path"},
	)

	// 节点状态
	nodesTotal = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "gost_panel_nodes_total",
			Help: "Total number of nodes",
		},
	)

	nodesOnline = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "gost_panel_nodes_online",
			Help: "Number of online nodes",
		},
	)

	// 客户端状态
	clientsTotal = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "gost_panel_clients_total",
			Help: "Total number of clients",
		},
	)

	clientsOnline = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "gost_panel_clients_online",
			Help: "Number of online clients",
		},
	)

	// 用户数量
	usersTotal = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "gost_panel_users_total",
			Help: "Total number of users",
		},
	)

	// 流量统计
	trafficBytesTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gost_panel_traffic_bytes_total",
			Help: "Total traffic in bytes",
		},
		[]string{"direction"}, // upload, download
	)

	// 登录尝试
	loginAttemptsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gost_panel_login_attempts_total",
			Help: "Total number of login attempts",
		},
		[]string{"status"}, // success, failed
	)

	// 数据库连接状态
	dbStatus = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "gost_panel_db_status",
			Help: "Database connection status (1=ok, 0=error)",
		},
	)
)

func init() {
	// 注册所有指标
	prometheus.MustRegister(
		httpRequestsTotal,
		httpRequestDuration,
		nodesTotal,
		nodesOnline,
		clientsTotal,
		clientsOnline,
		usersTotal,
		trafficBytesTotal,
		loginAttemptsTotal,
		dbStatus,
	)
}

// PrometheusMiddleware 记录 HTTP 请求指标
func PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		c.Next()

		// 记录请求
		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Writer.Status())

		httpRequestsTotal.WithLabelValues(c.Request.Method, path, status).Inc()
		httpRequestDuration.WithLabelValues(c.Request.Method, path).Observe(duration)
	}
}

// MetricsHandler 返回 Prometheus 指标处理器
func MetricsHandler() gin.HandlerFunc {
	h := promhttp.Handler()
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

// UpdateNodeMetrics 更新节点指标
func UpdateNodeMetrics(total, online int) {
	nodesTotal.Set(float64(total))
	nodesOnline.Set(float64(online))
}

// UpdateClientMetrics 更新客户端指标
func UpdateClientMetrics(total, online int) {
	clientsTotal.Set(float64(total))
	clientsOnline.Set(float64(online))
}

// UpdateUserMetrics 更新用户指标
func UpdateUserMetrics(total int) {
	usersTotal.Set(float64(total))
}

// RecordTraffic 记录流量
func RecordTraffic(upload, download int64) {
	trafficBytesTotal.WithLabelValues("upload").Add(float64(upload))
	trafficBytesTotal.WithLabelValues("download").Add(float64(download))
}

// RecordLoginAttempt 记录登录尝试
func RecordLoginAttempt(success bool) {
	if success {
		loginAttemptsTotal.WithLabelValues("success").Inc()
	} else {
		loginAttemptsTotal.WithLabelValues("failed").Inc()
	}
}

// UpdateDBStatus 更新数据库状态
func UpdateDBStatus(ok bool) {
	if ok {
		dbStatus.Set(1)
	} else {
		dbStatus.Set(0)
	}
}
