package gost

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client GOST API 客户端
type Client struct {
	baseURL  string
	username string
	password string
	client   *http.Client
}

// NewClient 创建 GOST API 客户端
func NewClient(host string, port int, username, password string) *Client {
	return &Client{
		baseURL:  fmt.Sprintf("http://%s:%d", host, port),
		username: username,
		password: password,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Response GOST API 响应
type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

// Socks5Config SOCKS5 服务配置
type Socks5Config struct {
	Name          string
	Port          int
	Username      string
	Password      string
	Bind          bool
	UDP           bool
	UDPBufferSize int
}

// CreateSocks5Service 创建 SOCKS5 服务
// 对应命令: gost -L "socks5://admin:pass@:38567?bind=true&udp=true&udpBufferSize=4096"
func (c *Client) CreateSocks5Service(cfg Socks5Config) error {
	service := map[string]interface{}{
		"name": cfg.Name,
		"addr": fmt.Sprintf(":%d", cfg.Port),
		"handler": map[string]interface{}{
			"type": "socks5",
			"metadata": map[string]interface{}{
				"bind":          cfg.Bind,
				"udp":           cfg.UDP,
				"udpBufferSize": cfg.UDPBufferSize,
			},
		},
		"listener": map[string]interface{}{
			"type": "tcp",
		},
	}

	// 添加认证
	if cfg.Username != "" && cfg.Password != "" {
		service["handler"].(map[string]interface{})["auther"] = cfg.Name + "-auth"

		// 先创建认证器
		auther := map[string]interface{}{
			"name": cfg.Name + "-auth",
			"auths": []map[string]string{
				{"username": cfg.Username, "password": cfg.Password},
			},
		}
		if err := c.post("/config/authers", auther); err != nil {
			return err
		}
	}

	return c.post("/config/services", service)
}

// ReverseTunnelConfig 反向隧道配置
type ReverseTunnelConfig struct {
	Name        string
	LocalPort   int    // 本地 SOCKS5 端口
	RemotePort  int    // 远程映射端口
	NodeHost    string // 节点地址
	NodePort    int    // 节点端口
	Username    string
	Password    string
}

// CreateReverseTunnel 创建反向隧道配置
// 对应命令:
// gost -L "socks5://admin:pass@:38777?udp=true&udpaddr=:38777&ignoreChain=true"
//      -L "rtcp://:38777/:38777?keepalive=true"
//      -L "rudp://:38777/:38777?keepalive=true"
//      -F "socks5://admin:pass@node:38567"
func (c *Client) CreateReverseTunnel(cfg ReverseTunnelConfig) error {
	chainName := cfg.Name + "-chain"

	// 1. 创建转发链
	chain := map[string]interface{}{
		"name": chainName,
		"hops": []map[string]interface{}{
			{
				"name": "hop-0",
				"nodes": []map[string]interface{}{
					{
						"name": "node-0",
						"addr": fmt.Sprintf("%s:%d", cfg.NodeHost, cfg.NodePort),
						"connector": map[string]interface{}{
							"type": "socks5",
							"auth": map[string]string{
								"username": cfg.Username,
								"password": cfg.Password,
							},
						},
						"dialer": map[string]interface{}{
							"type": "tcp",
						},
					},
				},
			},
		},
	}
	if err := c.post("/config/chains", chain); err != nil {
		return fmt.Errorf("create chain failed: %w", err)
	}

	// 2. 创建本地 SOCKS5 (ignoreChain)
	localSocks5 := map[string]interface{}{
		"name": cfg.Name + "-local-socks5",
		"addr": fmt.Sprintf(":%d", cfg.LocalPort),
		"handler": map[string]interface{}{
			"type": "socks5",
			"metadata": map[string]interface{}{
				"udp":         true,
				"udpAddr":     fmt.Sprintf(":%d", cfg.LocalPort),
				"ignoreChain": true,
			},
		},
		"listener": map[string]interface{}{
			"type": "tcp",
		},
	}
	if cfg.Username != "" {
		autherName := cfg.Name + "-local-auth"
		auther := map[string]interface{}{
			"name": autherName,
			"auths": []map[string]string{
				{"username": cfg.Username, "password": cfg.Password},
			},
		}
		c.post("/config/authers", auther)
		localSocks5["handler"].(map[string]interface{})["auther"] = autherName
	}
	if err := c.post("/config/services", localSocks5); err != nil {
		return fmt.Errorf("create local socks5 failed: %w", err)
	}

	// 3. 创建 RTCP 反向隧道
	rtcp := map[string]interface{}{
		"name": cfg.Name + "-rtcp",
		"addr": fmt.Sprintf(":%d", cfg.RemotePort),
		"handler": map[string]interface{}{
			"type": "rtcp",
		},
		"listener": map[string]interface{}{
			"type": "rtcp",
			"chain": chainName,
			"metadata": map[string]interface{}{
				"keepalive": true,
			},
		},
		"forwarder": map[string]interface{}{
			"nodes": []map[string]interface{}{
				{"name": "target", "addr": fmt.Sprintf(":%d", cfg.RemotePort)},
			},
		},
	}
	if err := c.post("/config/services", rtcp); err != nil {
		return fmt.Errorf("create rtcp failed: %w", err)
	}

	// 4. 创建 RUDP 反向隧道
	rudp := map[string]interface{}{
		"name": cfg.Name + "-rudp",
		"addr": fmt.Sprintf(":%d", cfg.RemotePort),
		"handler": map[string]interface{}{
			"type": "rudp",
		},
		"listener": map[string]interface{}{
			"type": "rudp",
			"chain": chainName,
			"metadata": map[string]interface{}{
				"keepalive": true,
			},
		},
		"forwarder": map[string]interface{}{
			"nodes": []map[string]interface{}{
				{"name": "target", "addr": fmt.Sprintf(":%d", cfg.RemotePort)},
			},
		},
	}
	if err := c.post("/config/services", rudp); err != nil {
		return fmt.Errorf("create rudp failed: %w", err)
	}

	return nil
}

// GetConfig 获取当前配置
func (c *Client) GetConfig() (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.get("/config", &result)
	return result, err
}

// GetServices 获取服务列表
func (c *Client) GetServices() ([]map[string]interface{}, error) {
	var result []map[string]interface{}
	err := c.get("/config/services", &result)
	return result, err
}

// DeleteService 删除服务
func (c *Client) DeleteService(name string) error {
	return c.delete("/config/services/" + name)
}

// CreateService 创建服务
func (c *Client) CreateService(config map[string]interface{}) error {
	return c.post("/config/services", config)
}

// GetChains 获取转发链列表
func (c *Client) GetChains() ([]map[string]interface{}, error) {
	var result []map[string]interface{}
	err := c.get("/config/chains", &result)
	return result, err
}

// CreateChain 创建转发链
func (c *Client) CreateChain(config map[string]interface{}) error {
	return c.post("/config/chains", config)
}

// DeleteChain 删除转发链
func (c *Client) DeleteChain(name string) error {
	return c.delete("/config/chains/" + name)
}

// SyncTunnelConfig 同步隧道配置 (增量添加，不影响现有服务)
func (c *Client) SyncTunnelConfig(config map[string]interface{}, tunnelID uint) error {
	// 删除该隧道的旧服务和链
	services, _ := c.GetServices()
	for _, svc := range services {
		if name, ok := svc["name"].(string); ok {
			// 匹配 tunnel-{id}-tcp 或 tunnel-{id}-udp 格式
			if matchTunnelService(name, tunnelID) {
				c.delete("/config/services/" + name)
			}
		}
	}

	chains, _ := c.GetChains()
	for _, chain := range chains {
		if name, ok := chain["name"].(string); ok {
			// 匹配 tunnel-chain-{id} 格式
			if name == fmt.Sprintf("tunnel-chain-%d", tunnelID) {
				c.delete("/config/chains/" + name)
			}
		}
	}

	// 创建新的转发链
	if newChains, ok := config["chains"].([]map[string]interface{}); ok {
		for _, chain := range newChains {
			if err := c.post("/config/chains", chain); err != nil {
				return fmt.Errorf("create chain failed: %w", err)
			}
		}
	}

	// 创建新服务
	if newServices, ok := config["services"].([]map[string]interface{}); ok {
		for _, svc := range newServices {
			if err := c.post("/config/services", svc); err != nil {
				return fmt.Errorf("create service failed: %w", err)
			}
		}
	}

	return nil
}

// matchTunnelService 检查服务名是否属于指定隧道
func matchTunnelService(name string, tunnelID uint) bool {
	// 匹配 tunnel-{id}-tcp, tunnel-{id}-udp, tunnel-{id}
	prefix := fmt.Sprintf("tunnel-%d", tunnelID)
	return name == prefix || name == prefix+"-tcp" || name == prefix+"-udp"
}

// UpdateService 更新服务配置
func (c *Client) UpdateService(name string, config map[string]interface{}) error {
	return c.put("/config/services/"+name, config)
}

// UpdateAuther 更新认证器配置
func (c *Client) UpdateAuther(name string, config map[string]interface{}) error {
	return c.put("/config/authers/"+name, config)
}

// SyncConfig 同步完整配置到 GOST (删除旧配置，创建新配置)
func (c *Client) SyncConfig(config map[string]interface{}) error {
	// 获取当前服务列表
	services, err := c.GetServices()
	if err != nil {
		return fmt.Errorf("get services failed: %w", err)
	}

	// 删除所有现有服务
	for _, svc := range services {
		if name, ok := svc["name"].(string); ok {
			c.delete("/config/services/" + name)
		}
	}

	// 创建新服务
	if newServices, ok := config["services"].([]interface{}); ok {
		for _, svc := range newServices {
			if svcMap, ok := svc.(map[string]interface{}); ok {
				if err := c.post("/config/services", svcMap); err != nil {
					return fmt.Errorf("create service failed: %w", err)
				}
			}
		}
	}

	return nil
}

// Ping 测试连接
func (c *Client) Ping() error {
	_, err := c.GetConfig()
	return err
}

// ==================== HTTP 方法 ====================

func (c *Client) get(path string, result interface{}) error {
	req, err := http.NewRequest("GET", c.baseURL+path, nil)
	if err != nil {
		return err
	}
	return c.doRequest(req, result)
}

func (c *Client) post(path string, body interface{}) error {
	data, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", c.baseURL+path, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	return c.doRequest(req, nil)
}

func (c *Client) delete(path string) error {
	req, err := http.NewRequest("DELETE", c.baseURL+path, nil)
	if err != nil {
		return err
	}
	return c.doRequest(req, nil)
}

func (c *Client) put(path string, body interface{}) error {
	data, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PUT", c.baseURL+path, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	return c.doRequest(req, nil)
}

func (c *Client) doRequest(req *http.Request, result interface{}) error {
	if c.username != "" {
		req.SetBasicAuth(c.username, c.password)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("API error: %s - %s", resp.Status, string(body))
	}

	if result != nil {
		var apiResp Response
		if err := json.Unmarshal(body, &apiResp); err != nil {
			return err
		}
		if apiResp.Code != 0 {
			return fmt.Errorf("GOST error: %s", apiResp.Msg)
		}
		if apiResp.Data != nil {
			dataBytes, _ := json.Marshal(apiResp.Data)
			return json.Unmarshal(dataBytes, result)
		}
	}

	return nil
}
