package notify

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/smtp"
	"strings"
	"time"

	"github.com/AliceNetworks/gost-panel/internal/model"
)

// Notifier 通知发送接口
type Notifier interface {
	Send(title, message string) error
}

// TelegramNotifier Telegram 通知
type TelegramNotifier struct {
	BotToken string
	ChatID   string
}

func NewTelegramNotifier(config *model.TelegramConfig) *TelegramNotifier {
	return &TelegramNotifier{
		BotToken: config.BotToken,
		ChatID:   config.ChatID,
	}
}

func (t *TelegramNotifier) Send(title, message string) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", t.BotToken)

	text := fmt.Sprintf("*%s*\n\n%s", escapeMarkdown(title), escapeMarkdown(message))

	payload := map[string]interface{}{
		"chat_id":    t.ChatID,
		"text":       text,
		"parse_mode": "MarkdownV2",
	}

	body, _ := json.Marshal(payload)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("telegram request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("telegram API error: %s", string(respBody))
	}

	return nil
}

func escapeMarkdown(text string) string {
	chars := []string{"_", "*", "[", "]", "(", ")", "~", "`", ">", "#", "+", "-", "=", "|", "{", "}", ".", "!"}
	for _, char := range chars {
		text = strings.ReplaceAll(text, char, "\\"+char)
	}
	return text
}

// WebhookNotifier Webhook 通知
type WebhookNotifier struct {
	URL     string
	Method  string
	Headers map[string]string
}

func NewWebhookNotifier(config *model.WebhookConfig) *WebhookNotifier {
	method := config.Method
	if method == "" {
		method = "POST"
	}
	return &WebhookNotifier{
		URL:     config.URL,
		Method:  method,
		Headers: config.Headers,
	}
}

func (w *WebhookNotifier) Send(title, message string) error {
	payload := map[string]interface{}{
		"title":     title,
		"message":   message,
		"timestamp": time.Now().Unix(),
	}

	body, _ := json.Marshal(payload)

	req, err := http.NewRequest(w.Method, w.URL, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("create request failed: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	for k, v := range w.Headers {
		req.Header.Set(k, v)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("webhook request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("webhook error (status %d): %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// SMTPNotifier SMTP 邮件通知
type SMTPNotifier struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	To       []string
	UseTLS   bool
}

func NewSMTPNotifier(config *model.SMTPConfig) *SMTPNotifier {
	to := strings.Split(config.To, ",")
	for i := range to {
		to[i] = strings.TrimSpace(to[i])
	}
	return &SMTPNotifier{
		Host:     config.Host,
		Port:     config.Port,
		Username: config.Username,
		Password: config.Password,
		From:     config.From,
		To:       to,
		UseTLS:   config.UseTLS,
	}
}

func (s *SMTPNotifier) Send(title, message string) error {
	addr := fmt.Sprintf("%s:%d", s.Host, s.Port)

	// 构建邮件内容
	subject := fmt.Sprintf("Subject: [GOST Panel] %s\r\n", title)
	mime := "MIME-version: 1.0;\r\nContent-Type: text/plain; charset=\"UTF-8\";\r\n\r\n"
	body := []byte(subject + mime + message)

	auth := smtp.PlainAuth("", s.Username, s.Password, s.Host)

	if s.UseTLS {
		// TLS 连接
		tlsConfig := &tls.Config{
			InsecureSkipVerify: false,
			ServerName:         s.Host,
		}

		conn, err := tls.Dial("tcp", addr, tlsConfig)
		if err != nil {
			return fmt.Errorf("TLS dial failed: %w", err)
		}
		defer conn.Close()

		client, err := smtp.NewClient(conn, s.Host)
		if err != nil {
			return fmt.Errorf("SMTP client failed: %w", err)
		}
		defer client.Close()

		if err = client.Auth(auth); err != nil {
			return fmt.Errorf("SMTP auth failed: %w", err)
		}

		if err = client.Mail(s.From); err != nil {
			return fmt.Errorf("SMTP mail failed: %w", err)
		}

		for _, to := range s.To {
			if err = client.Rcpt(to); err != nil {
				return fmt.Errorf("SMTP rcpt failed: %w", err)
			}
		}

		w, err := client.Data()
		if err != nil {
			return fmt.Errorf("SMTP data failed: %w", err)
		}

		_, err = w.Write(body)
		if err != nil {
			return fmt.Errorf("SMTP write failed: %w", err)
		}

		err = w.Close()
		if err != nil {
			return fmt.Errorf("SMTP close failed: %w", err)
		}

		return client.Quit()
	}

	// 非 TLS
	return smtp.SendMail(addr, auth, s.From, s.To, body)
}

// CreateNotifier 根据渠道配置创建通知器
func CreateNotifier(channel *model.NotifyChannel) (Notifier, error) {
	switch channel.Type {
	case "telegram":
		var config model.TelegramConfig
		if err := json.Unmarshal([]byte(channel.Config), &config); err != nil {
			return nil, fmt.Errorf("parse telegram config failed: %w", err)
		}
		return NewTelegramNotifier(&config), nil

	case "webhook":
		var config model.WebhookConfig
		if err := json.Unmarshal([]byte(channel.Config), &config); err != nil {
			return nil, fmt.Errorf("parse webhook config failed: %w", err)
		}
		return NewWebhookNotifier(&config), nil

	case "smtp":
		var config model.SMTPConfig
		if err := json.Unmarshal([]byte(channel.Config), &config); err != nil {
			return nil, fmt.Errorf("parse smtp config failed: %w", err)
		}
		return NewSMTPNotifier(&config), nil

	default:
		return nil, fmt.Errorf("unknown channel type: %s", channel.Type)
	}
}
