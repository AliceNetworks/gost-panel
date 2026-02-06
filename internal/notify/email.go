package notify

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"

	"github.com/AliceNetworks/gost-panel/internal/model"
)

// EmailSender 用户邮件发送服务
type EmailSender struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	UseTLS   bool
	SiteName string
	SiteURL  string
}

// NewEmailSender 创建邮件发送器
func NewEmailSender(config *model.SMTPConfig, siteName, siteURL string) *EmailSender {
	return &EmailSender{
		Host:     config.Host,
		Port:     config.Port,
		Username: config.Username,
		Password: config.Password,
		From:     config.From,
		UseTLS:   config.UseTLS,
		SiteName: siteName,
		SiteURL:  siteURL,
	}
}

// SendVerificationEmail 发送邮箱验证邮件
func (e *EmailSender) SendVerificationEmail(to, username, token string) error {
	subject := fmt.Sprintf("验证您的 %s 账户", e.SiteName)

	verifyURL := fmt.Sprintf("%s/verify-email?token=%s", strings.TrimSuffix(e.SiteURL, "/"), token)

	htmlBody := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #18a058; color: white; padding: 20px; text-align: center; border-radius: 8px 8px 0 0; }
        .content { background: #f9f9f9; padding: 30px; border-radius: 0 0 8px 8px; }
        .button { display: inline-block; background: #18a058; color: white; padding: 12px 24px; text-decoration: none; border-radius: 6px; margin: 20px 0; }
        .footer { text-align: center; color: #666; font-size: 12px; margin-top: 20px; }
        .code { background: #e8e8e8; padding: 10px 15px; border-radius: 4px; font-family: monospace; font-size: 14px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>%s</h1>
        </div>
        <div class="content">
            <h2>欢迎，%s！</h2>
            <p>感谢您注册 %s。请点击下面的按钮验证您的邮箱地址：</p>
            <p style="text-align: center;">
                <a href="%s" class="button">验证邮箱</a>
            </p>
            <p>或者复制以下链接到浏览器：</p>
            <p class="code">%s</p>
            <p>如果您没有注册账户，请忽略此邮件。</p>
        </div>
        <div class="footer">
            <p>此邮件由 %s 自动发送，请勿直接回复。</p>
        </div>
    </div>
</body>
</html>`, e.SiteName, username, e.SiteName, verifyURL, verifyURL, e.SiteName)

	return e.sendHTMLEmail(to, subject, htmlBody)
}

// SendPasswordResetEmail 发送密码重置邮件
func (e *EmailSender) SendPasswordResetEmail(to, username, token string) error {
	subject := fmt.Sprintf("重置您的 %s 密码", e.SiteName)

	resetURL := fmt.Sprintf("%s/reset-password?token=%s", strings.TrimSuffix(e.SiteURL, "/"), token)

	htmlBody := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #f0a020; color: white; padding: 20px; text-align: center; border-radius: 8px 8px 0 0; }
        .content { background: #f9f9f9; padding: 30px; border-radius: 0 0 8px 8px; }
        .button { display: inline-block; background: #f0a020; color: white; padding: 12px 24px; text-decoration: none; border-radius: 6px; margin: 20px 0; }
        .footer { text-align: center; color: #666; font-size: 12px; margin-top: 20px; }
        .code { background: #e8e8e8; padding: 10px 15px; border-radius: 4px; font-family: monospace; font-size: 14px; }
        .warning { color: #d03050; font-weight: bold; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>密码重置</h1>
        </div>
        <div class="content">
            <h2>您好，%s</h2>
            <p>我们收到了重置您 %s 账户密码的请求。</p>
            <p>请点击下面的按钮重置密码：</p>
            <p style="text-align: center;">
                <a href="%s" class="button">重置密码</a>
            </p>
            <p>或者复制以下链接到浏览器：</p>
            <p class="code">%s</p>
            <p class="warning">此链接将在 1 小时后失效。</p>
            <p>如果您没有请求重置密码，请忽略此邮件，您的密码不会被更改。</p>
        </div>
        <div class="footer">
            <p>此邮件由 %s 自动发送，请勿直接回复。</p>
        </div>
    </div>
</body>
</html>`, username, e.SiteName, resetURL, resetURL, e.SiteName)

	return e.sendHTMLEmail(to, subject, htmlBody)
}

// SendWelcomeEmail 发送欢迎邮件（验证成功后）
func (e *EmailSender) SendWelcomeEmail(to, username string) error {
	subject := fmt.Sprintf("欢迎加入 %s", e.SiteName)

	loginURL := fmt.Sprintf("%s/login", strings.TrimSuffix(e.SiteURL, "/"))

	htmlBody := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #18a058; color: white; padding: 20px; text-align: center; border-radius: 8px 8px 0 0; }
        .content { background: #f9f9f9; padding: 30px; border-radius: 0 0 8px 8px; }
        .button { display: inline-block; background: #18a058; color: white; padding: 12px 24px; text-decoration: none; border-radius: 6px; margin: 20px 0; }
        .footer { text-align: center; color: #666; font-size: 12px; margin-top: 20px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🎉 欢迎加入！</h1>
        </div>
        <div class="content">
            <h2>您好，%s！</h2>
            <p>您的邮箱已成功验证，账户现已激活。</p>
            <p>现在您可以登录并开始使用 %s 的所有功能：</p>
            <p style="text-align: center;">
                <a href="%s" class="button">立即登录</a>
            </p>
        </div>
        <div class="footer">
            <p>此邮件由 %s 自动发送，请勿直接回复。</p>
        </div>
    </div>
</body>
</html>`, username, e.SiteName, loginURL, e.SiteName)

	return e.sendHTMLEmail(to, subject, htmlBody)
}

// sendHTMLEmail 发送 HTML 邮件
func (e *EmailSender) sendHTMLEmail(to, subject, htmlBody string) error {
	addr := fmt.Sprintf("%s:%d", e.Host, e.Port)

	// 构建邮件头
	headers := make(map[string]string)
	headers["From"] = e.From
	headers["To"] = to
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=UTF-8"

	var message strings.Builder
	for k, v := range headers {
		message.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	message.WriteString("\r\n")
	message.WriteString(htmlBody)

	body := []byte(message.String())
	auth := smtp.PlainAuth("", e.Username, e.Password, e.Host)

	if e.UseTLS {
		// TLS 连接
		tlsConfig := &tls.Config{
			InsecureSkipVerify: false,
			ServerName:         e.Host,
		}

		conn, err := tls.Dial("tcp", addr, tlsConfig)
		if err != nil {
			return fmt.Errorf("TLS dial failed: %w", err)
		}
		defer conn.Close()

		client, err := smtp.NewClient(conn, e.Host)
		if err != nil {
			return fmt.Errorf("SMTP client failed: %w", err)
		}
		defer client.Close()

		if err = client.Auth(auth); err != nil {
			return fmt.Errorf("SMTP auth failed: %w", err)
		}

		if err = client.Mail(e.From); err != nil {
			return fmt.Errorf("SMTP mail failed: %w", err)
		}

		if err = client.Rcpt(to); err != nil {
			return fmt.Errorf("SMTP rcpt failed: %w", err)
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
	return smtp.SendMail(addr, auth, e.From, []string{to}, body)
}
