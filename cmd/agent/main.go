package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const AgentVersion = "1.3.0"

var (
	panelURL    = flag.String("panel", "", "Panel URL (e.g., http://panel.example.com:8080)")
	token       = flag.String("token", "", "Agent token")
	configPath  = flag.String("config", "/etc/gost/gost.yml", "GOST config path")
	gostPath    = flag.String("gost", "/usr/local/bin/gost", "GOST binary path")
	gostAPI     = flag.String("gost-api", "http://127.0.0.1:18080", "GOST API address")
	gostUser    = flag.String("gost-user", "", "GOST API username")
	gostPass    = flag.String("gost-pass", "", "GOST API password")
	autoUpdate  = flag.Bool("auto-update", true, "Enable auto update")
	showVersion = flag.Bool("version", false, "Show version")
)

type Agent struct {
	panelURL   string
	token      string
	configPath string
	gostPath   string
	gostAPI    string
	gostUser   string
	gostPass   string
	autoUpdate bool
	gostCmd    *exec.Cmd
	client     *http.Client
	// 用于计算增量流量
	lastTrafficIn    int64
	lastTrafficOut   int64
	lastServiceStats map[string]ServiceStats // 按服务名记录上次统计
}

// ServiceStats 单个服务的统计
type ServiceStats struct {
	TrafficIn   int64
	TrafficOut  int64
	Connections int
}

func NewAgent(panelURL, token, configPath, gostPath, gostAPI, gostUser, gostPass string, autoUpdate bool) *Agent {
	return &Agent{
		panelURL:         panelURL,
		token:            token,
		configPath:       configPath,
		gostPath:         gostPath,
		gostAPI:          gostAPI,
		gostUser:         gostUser,
		gostPass:         gostPass,
		autoUpdate:       autoUpdate,
		lastServiceStats: make(map[string]ServiceStats),
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (a *Agent) Run() error {
	// 检查更新
	if a.autoUpdate {
		if updated, err := a.checkAndUpdate(); err != nil {
			log.Printf("Update check failed: %v", err)
		} else if updated {
			log.Println("Agent updated, restarting...")
			return a.restartSelf()
		}
	}

	// 注册到面板
	if err := a.register(); err != nil {
		return fmt.Errorf("register failed: %w", err)
	}
	log.Println("Registered to panel successfully")

	// 下载配置
	if err := a.downloadConfig(); err != nil {
		return fmt.Errorf("download config failed: %w", err)
	}
	log.Println("Config downloaded")

	// 启动 GOST
	if err := a.startGost(); err != nil {
		return fmt.Errorf("start gost failed: %w", err)
	}
	log.Println("GOST started")

	// 启动心跳
	go a.heartbeatLoop()

	// 启动更新检查
	if a.autoUpdate {
		go a.updateCheckLoop()
	}

	// 等待信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down...")
	a.stopGost()

	return nil
}

func (a *Agent) register() error {
	data := map[string]string{
		"token":   a.token,
		"version": AgentVersion,
	}

	body, _ := json.Marshal(data)
	resp, err := a.client.Post(a.panelURL+"/agent/register", "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("register failed: %s", string(respBody))
	}

	return nil
}

func (a *Agent) downloadConfig() error {
	resp, err := a.client.Get(a.panelURL + "/agent/config/" + a.token)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download config failed: status %d", resp.StatusCode)
	}

	// 确保目录存在
	if err := os.MkdirAll("/etc/gost", 0755); err != nil {
		return err
	}

	// 写入配置文件
	configData, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return os.WriteFile(a.configPath, configData, 0644)
}

func (a *Agent) startGost() error {
	a.gostCmd = exec.Command(a.gostPath, "-C", a.configPath)
	a.gostCmd.Stdout = os.Stdout
	a.gostCmd.Stderr = os.Stderr

	if err := a.gostCmd.Start(); err != nil {
		return err
	}

	// 监控进程
	go func() {
		err := a.gostCmd.Wait()
		if err != nil {
			log.Printf("GOST exited with error: %v", err)
		}
	}()

	return nil
}

func (a *Agent) stopGost() {
	if a.gostCmd != nil && a.gostCmd.Process != nil {
		a.gostCmd.Process.Signal(syscall.SIGTERM)
		time.Sleep(2 * time.Second)
		a.gostCmd.Process.Kill()
	}
}

func (a *Agent) heartbeatLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if err := a.sendHeartbeat(); err != nil {
			log.Printf("Heartbeat failed: %v", err)
		}
	}
}

func (a *Agent) sendHeartbeat() error {
	// 从 GOST API 获取统计数据
	stats := a.getGostStats()
	serviceStats := a.getServiceStats()

	// 计算当前配置的哈希值
	configHash := a.getConfigHash()

	data := map[string]interface{}{
		"token":          a.token,
		"connections":    stats.Connections,
		"traffic_in":     stats.TrafficIn,
		"traffic_out":    stats.TrafficOut,
		"config_hash":    configHash,
		"agent_version":  AgentVersion,
		"service_stats":  serviceStats, // 按服务名分类的统计
	}

	body, _ := json.Marshal(data)
	resp, err := a.client.Post(a.panelURL+"/agent/heartbeat", "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 解析响应
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	// 检查是否需要卸载
	if uninstall, ok := result["uninstall"].(bool); ok && uninstall {
		log.Println("Received uninstall command from panel, uninstalling...")
		go a.uninstall()
		return nil
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("heartbeat failed: status %d", resp.StatusCode)
	}

	// 检查是否需要重载配置
	if reload, ok := result["reload_config"].(bool); ok && reload {
		log.Println("Config update detected, reloading...")
		go a.reloadConfig()
	}

	// 检查是否需要更新 Agent (服务端推送)
	if a.autoUpdate {
		forceUpdate, _ := result["force_update"].(bool)
		needsUpdate, _ := result["needs_update"].(bool)

		if forceUpdate {
			log.Println("Force update command received from panel, updating...")
			go a.performUpdate()
		} else if needsUpdate {
			log.Println("Update available, will update on next restart")
		}
	}

	return nil
}

// performUpdate 执行更新
func (a *Agent) performUpdate() {
	if updated, err := a.checkAndUpdate(); err != nil {
		log.Printf("Update failed: %v", err)
	} else if updated {
		log.Println("Agent updated, restarting...")
		a.restartSelf()
	}
}

// uninstall 卸载 Agent 和 GOST
func (a *Agent) uninstall() {
	log.Println("Stopping GOST...")
	a.stopGost()

	log.Println("Stopping and disabling service...")
	exec.Command("systemctl", "stop", "gost-node").Run()
	exec.Command("systemctl", "disable", "gost-node").Run()

	log.Println("Removing files...")
	os.Remove("/etc/systemd/system/gost-node.service")
	os.RemoveAll("/opt/gost-panel")
	os.RemoveAll("/etc/gost")

	exec.Command("systemctl", "daemon-reload").Run()

	log.Println("Uninstall complete. Exiting...")
	os.Exit(0)
}

// getConfigHash 获取当前配置文件的修改时间作为版本号
func (a *Agent) getConfigHash() string {
	info, err := os.Stat(a.configPath)
	if err != nil {
		return ""
	}
	return strconv.FormatInt(info.ModTime().Unix(), 10)
}

// reloadConfig 重新下载并应用配置
func (a *Agent) reloadConfig() {
	if err := a.downloadConfig(); err != nil {
		log.Printf("Failed to download config: %v", err)
		return
	}
	log.Println("Config downloaded, restarting GOST...")

	// 重启 GOST
	a.stopGost()
	time.Sleep(time.Second)
	if err := a.startGost(); err != nil {
		log.Printf("Failed to restart GOST: %v", err)
	} else {
		log.Println("GOST restarted successfully")
	}
}

// GostStats GOST 统计数据
type GostStats struct {
	Connections int
	TrafficIn   int64
	TrafficOut  int64
}

// getGostStats 从 GOST API 获取统计数据
func (a *Agent) getGostStats() GostStats {
	stats := GostStats{}

	// 尝试方法1: 从 observer 获取统计
	if observerStats, err := a.fetchObserverStats(); err == nil && observerStats != nil {
		return a.processObserverStats(observerStats)
	}

	// 尝试方法2: 从 Prometheus metrics 获取
	if metrics, err := a.fetchPrometheusMetrics(); err == nil {
		return a.parsePrometheusMetrics(metrics)
	}

	// 尝试方法3: 从服务列表获取 (可能不包含统计)
	services, err := a.fetchGostServices()
	if err != nil {
		log.Printf("Failed to fetch GOST stats: %v", err)
		return stats
	}

	var totalIn, totalOut int64
	var totalConns int

	for _, svc := range services {
		if svcStats, ok := svc["stats"].(map[string]interface{}); ok {
			if inputBytes, ok := svcStats["inputBytes"].(float64); ok {
				totalIn += int64(inputBytes)
			}
			if outputBytes, ok := svcStats["outputBytes"].(float64); ok {
				totalOut += int64(outputBytes)
			}
			if currentConns, ok := svcStats["currentConns"].(float64); ok {
				totalConns += int(currentConns)
			}
		}
	}

	return a.calculateDelta(totalConns, totalIn, totalOut)
}

// fetchObserverStats 从 observer 获取统计
func (a *Agent) fetchObserverStats() (map[string]interface{}, error) {
	req, err := http.NewRequest("GET", a.gostAPI+"/config/observers/stats-observer/stats", nil)
	if err != nil {
		return nil, err
	}

	if a.gostUser != "" {
		req.SetBasicAuth(a.gostUser, a.gostPass)
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("observer API error: %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)
	return result, nil
}

// processObserverStats 处理 observer 统计数据
func (a *Agent) processObserverStats(data map[string]interface{}) GostStats {
	var totalIn, totalOut int64
	var totalConns int

	if services, ok := data["services"].([]interface{}); ok {
		for _, svc := range services {
			if svcMap, ok := svc.(map[string]interface{}); ok {
				if inputBytes, ok := svcMap["inputBytes"].(float64); ok {
					totalIn += int64(inputBytes)
				}
				if outputBytes, ok := svcMap["outputBytes"].(float64); ok {
					totalOut += int64(outputBytes)
				}
				if conns, ok := svcMap["currentConns"].(float64); ok {
					totalConns += int(conns)
				}
			}
		}
	}

	return a.calculateDelta(totalConns, totalIn, totalOut)
}

// fetchPrometheusMetrics 从 Prometheus 端点获取 metrics
func (a *Agent) fetchPrometheusMetrics() (string, error) {
	resp, err := a.client.Get("http://127.0.0.1:9000/metrics")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("metrics error: %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	return string(body), nil
}

// parsePrometheusMetrics 解析 Prometheus metrics
func (a *Agent) parsePrometheusMetrics(metrics string) GostStats {
	var totalIn, totalOut int64
	var totalConns int

	lines := strings.Split(metrics, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}

		// 解析 gost_service_transfer_input_bytes_total
		if strings.Contains(line, "gost_service_transfer_input_bytes_total") {
			if val := extractMetricValue(line); val > 0 {
				totalIn += int64(val)
			}
		}
		// 解析 gost_service_transfer_output_bytes_total
		if strings.Contains(line, "gost_service_transfer_output_bytes_total") {
			if val := extractMetricValue(line); val > 0 {
				totalOut += int64(val)
			}
		}
		// 解析 gost_services
		if strings.Contains(line, "gost_service_requests_in_flight") {
			if val := extractMetricValue(line); val > 0 {
				totalConns += int(val)
			}
		}
	}

	return a.calculateDelta(totalConns, totalIn, totalOut)
}

// extractMetricValue 从 Prometheus 格式行中提取值
func extractMetricValue(line string) float64 {
	parts := strings.Fields(line)
	if len(parts) >= 2 {
		val, _ := strconv.ParseFloat(parts[len(parts)-1], 64)
		return val
	}
	return 0
}

// calculateDelta 计算增量流量
func (a *Agent) calculateDelta(conns int, totalIn, totalOut int64) GostStats {
	stats := GostStats{Connections: conns}

	// 计算增量流量 (面板存储累计流量，这里发送增量)
	deltaIn := totalIn - a.lastTrafficIn
	deltaOut := totalOut - a.lastTrafficOut

	// 首次运行或重启后，不发送历史数据
	if a.lastTrafficIn == 0 && a.lastTrafficOut == 0 {
		deltaIn = 0
		deltaOut = 0
	}

	// 处理 GOST 重启导致的计数器重置
	if deltaIn < 0 {
		deltaIn = totalIn
	}
	if deltaOut < 0 {
		deltaOut = totalOut
	}

	a.lastTrafficIn = totalIn
	a.lastTrafficOut = totalOut

	stats.TrafficIn = deltaIn
	stats.TrafficOut = deltaOut

	return stats
}

// fetchGostServices 从 GOST API 获取服务列表
func (a *Agent) fetchGostServices() ([]map[string]interface{}, error) {
	req, err := http.NewRequest("GET", a.gostAPI+"/config/services", nil)
	if err != nil {
		return nil, err
	}

	if a.gostUser != "" {
		req.SetBasicAuth(a.gostUser, a.gostPass)
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GOST API error: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// GOST API 响应格式: {"code":0,"data":[...]}
	var apiResp struct {
		Code int                      `json:"code"`
		Data []map[string]interface{} `json:"data"`
	}
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, err
	}

	if apiResp.Code != 0 {
		return nil, fmt.Errorf("GOST API returned code: %d", apiResp.Code)
	}

	return apiResp.Data, nil
}

// getServiceStats 获取按服务名分类的流量统计 (增量)
func (a *Agent) getServiceStats() map[string]map[string]int64 {
	result := make(map[string]map[string]int64)

	// 尝试从 observer 获取
	if observerStats, err := a.fetchObserverStats(); err == nil && observerStats != nil {
		if services, ok := observerStats["services"].([]interface{}); ok {
			for _, svc := range services {
				if svcMap, ok := svc.(map[string]interface{}); ok {
					name, _ := svcMap["service"].(string)
					if name == "" || name == "main-service" {
						continue // 跳过主服务，已在总流量中统计
					}

					var totalIn, totalOut int64
					var conns int
					if inputBytes, ok := svcMap["inputBytes"].(float64); ok {
						totalIn = int64(inputBytes)
					}
					if outputBytes, ok := svcMap["outputBytes"].(float64); ok {
						totalOut = int64(outputBytes)
					}
					if currentConns, ok := svcMap["currentConns"].(float64); ok {
						conns = int(currentConns)
					}

					// 计算增量
					lastStats := a.lastServiceStats[name]
					deltaIn := totalIn - lastStats.TrafficIn
					deltaOut := totalOut - lastStats.TrafficOut

					// 处理重置
					if deltaIn < 0 {
						deltaIn = totalIn
					}
					if deltaOut < 0 {
						deltaOut = totalOut
					}

					// 更新上次记录
					a.lastServiceStats[name] = ServiceStats{
						TrafficIn:   totalIn,
						TrafficOut:  totalOut,
						Connections: conns,
					}

					// 只上报有流量的服务
					if deltaIn > 0 || deltaOut > 0 || conns > 0 {
						result[name] = map[string]int64{
							"traffic_in":  deltaIn,
							"traffic_out": deltaOut,
							"connections": int64(conns),
						}
					}
				}
			}
		}
	}

	return result
}

// ==================== 自动更新相关 ====================

// UpdateInfo represents update check response
type UpdateInfo struct {
	CurrentVersion string `json:"current_version"`
	LatestVersion  string `json:"latest_version"`
	NeedsUpdate    bool   `json:"needs_update"`
	DownloadURL    string `json:"download_url"`
	Checksum       string `json:"checksum"`
}

// updateCheckLoop periodically checks for updates
func (a *Agent) updateCheckLoop() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		if updated, err := a.checkAndUpdate(); err != nil {
			log.Printf("Update check failed: %v", err)
		} else if updated {
			log.Println("Agent updated, restarting...")
			a.stopGost()
			a.restartSelf()
		}
	}
}

// checkAndUpdate checks for updates and downloads if available
func (a *Agent) checkAndUpdate() (bool, error) {
	url := fmt.Sprintf("%s/agent/check-update?version=%s&os=%s&arch=%s",
		a.panelURL, AgentVersion, runtime.GOOS, runtime.GOARCH)

	resp, err := a.client.Get(url)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("check update failed: status %d", resp.StatusCode)
	}

	var info UpdateInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return false, err
	}

	if !info.NeedsUpdate {
		log.Printf("Agent is up to date (version %s)", AgentVersion)
		return false, nil
	}

	log.Printf("Update available: %s -> %s", AgentVersion, info.LatestVersion)

	if info.DownloadURL == "" {
		return false, fmt.Errorf("no download URL provided")
	}

	// Download the update
	if err := a.downloadUpdate(info.DownloadURL, info.Checksum); err != nil {
		return false, fmt.Errorf("download update failed: %w", err)
	}

	return true, nil
}

// downloadUpdate downloads and installs the update
func (a *Agent) downloadUpdate(downloadURL, expectedChecksum string) error {
	fullURL := a.panelURL + downloadURL
	log.Printf("Downloading update from %s", fullURL)

	resp, err := a.client.Get(fullURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed: status %d", resp.StatusCode)
	}

	// Get current executable path
	execPath, err := os.Executable()
	if err != nil {
		return err
	}
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return err
	}

	// Create temp file
	tmpFile := execPath + ".new"
	f, err := os.OpenFile(tmpFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}

	// Download and calculate checksum
	hash := sha256.New()
	writer := io.MultiWriter(f, hash)

	_, err = io.Copy(writer, resp.Body)
	f.Close()
	if err != nil {
		os.Remove(tmpFile)
		return err
	}

	// Verify checksum
	if expectedChecksum != "" {
		actualChecksum := fmt.Sprintf("%x", hash.Sum(nil))
		if actualChecksum != expectedChecksum {
			os.Remove(tmpFile)
			return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedChecksum, actualChecksum)
		}
		log.Println("Checksum verified")
	}

	// Backup old binary
	backupPath := execPath + ".bak"
	if err := os.Rename(execPath, backupPath); err != nil {
		os.Remove(tmpFile)
		return fmt.Errorf("backup failed: %w", err)
	}

	// Install new binary
	if err := os.Rename(tmpFile, execPath); err != nil {
		// Restore backup
		os.Rename(backupPath, execPath)
		return fmt.Errorf("install failed: %w", err)
	}

	log.Printf("Update installed successfully")
	return nil
}

// restartSelf restarts the agent process
func (a *Agent) restartSelf() error {
	execPath, err := os.Executable()
	if err != nil {
		return err
	}

	// Re-exec with same arguments
	args := os.Args
	log.Printf("Restarting: %s %v", execPath, args[1:])

	cmd := exec.Command(execPath, args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Start(); err != nil {
		return err
	}

	// Exit current process
	os.Exit(0)
	return nil
}

func main() {
	flag.Parse()

	if *showVersion {
		fmt.Printf("gost-agent version %s (%s/%s)\n", AgentVersion, runtime.GOOS, runtime.GOARCH)
		os.Exit(0)
	}

	if *panelURL == "" || *token == "" {
		fmt.Println("Usage: gost-agent -panel <panel_url> -token <token>")
		fmt.Println("  -panel       Panel URL (e.g., http://panel.example.com:8080)")
		fmt.Println("  -token       Agent token from panel")
		fmt.Println("  -config      GOST config path (default: /etc/gost/gost.yml)")
		fmt.Println("  -gost        GOST binary path (default: /usr/local/bin/gost)")
		fmt.Println("  -gost-api    GOST API address (default: http://127.0.0.1:18080)")
		fmt.Println("  -gost-user   GOST API username (optional)")
		fmt.Println("  -gost-pass   GOST API password (optional)")
		fmt.Println("  -auto-update Enable auto update (default: true)")
		fmt.Println("  -version     Show version")
		os.Exit(1)
	}

	log.Printf("Starting gost-agent %s (%s/%s)", AgentVersion, runtime.GOOS, runtime.GOARCH)

	agent := NewAgent(*panelURL, *token, *configPath, *gostPath, *gostAPI, *gostUser, *gostPass, *autoUpdate)
	if err := agent.Run(); err != nil {
		log.Fatalf("Agent error: %v", err)
	}
}
