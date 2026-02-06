package gost

import (
	"encoding/base64"
	"fmt"
	"net/url"

	"github.com/AliceNetworks/gost-panel/internal/model"
)

// GenerateProxyURI 生成代理 URI
func GenerateProxyURI(node *model.Node) string {
	switch node.Protocol {
	case "socks5":
		return generateSocks5URI(node)
	case "http":
		return generateHTTPURI(node)
	case "ss":
		return generateShadowsocksURI(node)
	case "trojan":
		return generateTrojanURI(node)
	case "vmess":
		return generateVMessURI(node)
	default:
		return generateSocks5URI(node)
	}
}

func generateSocks5URI(node *model.Node) string {
	// socks5://user:pass@host:port
	uri := "socks5://"
	if node.ProxyUser != "" && node.ProxyPass != "" {
		uri += url.QueryEscape(node.ProxyUser) + ":" + url.QueryEscape(node.ProxyPass) + "@"
	}
	uri += fmt.Sprintf("%s:%d", node.Host, node.Port)
	return uri
}

func generateHTTPURI(node *model.Node) string {
	// http://user:pass@host:port
	uri := "http://"
	if node.ProxyUser != "" && node.ProxyPass != "" {
		uri += url.QueryEscape(node.ProxyUser) + ":" + url.QueryEscape(node.ProxyPass) + "@"
	}
	uri += fmt.Sprintf("%s:%d", node.Host, node.Port)
	return uri
}

func generateShadowsocksURI(node *model.Node) string {
	// ss://base64(method:password)@host:port#name
	method := node.SSMethod
	if method == "" {
		method = "aes-256-gcm"
	}
	userinfo := fmt.Sprintf("%s:%s", method, node.SSPassword)
	encoded := base64.URLEncoding.EncodeToString([]byte(userinfo))
	return fmt.Sprintf("ss://%s@%s:%d#%s", encoded, node.Host, node.Port, url.QueryEscape(node.Name))
}

func generateTrojanURI(node *model.Node) string {
	// trojan://password@host:port?sni=xxx#name
	uri := fmt.Sprintf("trojan://%s@%s:%d", url.QueryEscape(node.TrojanPassword), node.Host, node.Port)
	params := url.Values{}
	if node.TLSSNI != "" {
		params.Set("sni", node.TLSSNI)
	}
	if len(params) > 0 {
		uri += "?" + params.Encode()
	}
	uri += "#" + url.QueryEscape(node.Name)
	return uri
}

func generateVMessURI(node *model.Node) string {
	// vmess://base64(json)
	// 简化的 VMess 链接
	return fmt.Sprintf("vmess://%s:%d (UUID: %s)", node.Host, node.Port, node.VMessUUID)
}
