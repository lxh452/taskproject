package middleware

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/smtp"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

// EmailConfig 邮件配置
type EmailConfig struct {
	Host     string `json:"host"`     // SMTP服务器地址
	Port     int    `json:"port"`     // SMTP端口
	Username string `json:"username"` // 发送者邮箱
	Password string `json:"password"` // 邮箱密码或授权码
	From     string `json:"from"`     // 发送者名称
	UseTLS   bool   `json:"useTLS"`   // 是否使用TLS
}

// EmailMessage 邮件消息
type EmailMessage struct {
	To      []string `json:"to"`      // 收件人列表
	Subject string   `json:"subject"` // 邮件主题
	Body    string   `json:"body"`    // 邮件内容
	IsHTML  bool     `json:"isHtml"`  // 是否为HTML格式
}

// EmailMiddleware 邮件中间件
type EmailMiddleware struct {
	config EmailConfig
}

// NewEmailMiddleware 创建邮件中间件
func NewEmailMiddleware(config EmailConfig) *EmailMiddleware {
	return &EmailMiddleware{
		config: config,
	}
}

// SendEmail 发送邮件
func (e *EmailMiddleware) SendEmail(ctx context.Context, msg EmailMessage) error {
	// 构建邮件头
	headers := make(map[string]string)
	headers["From"] = fmt.Sprintf("%s <%s>", e.config.From, e.config.Username)
	headers["To"] = fmt.Sprintf("%s", msg.To[0])
	headers["Subject"] = msg.Subject
	headers["Date"] = time.Now().Format(time.RFC1123Z)
	headers["MIME-Version"] = "1.0"
	
	if msg.IsHTML {
		headers["Content-Type"] = "text/html; charset=UTF-8"
	} else {
		headers["Content-Type"] = "text/plain; charset=UTF-8"
	}

	// 构建邮件内容
	emailBody := ""
	for k, v := range headers {
		emailBody += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	emailBody += "\r\n" + msg.Body

	// 设置SMTP认证
	auth := smtp.PlainAuth("", e.config.Username, e.config.Password, e.config.Host)

	// 构建SMTP地址
	addr := fmt.Sprintf("%s:%d", e.config.Host, e.config.Port)

	// 发送邮件
	err := e.sendMailWithTLS(addr, auth, e.config.Username, msg.To, []byte(emailBody))
	if err != nil {
		logx.Errorf("发送邮件失败: %v", err)
		return err
	}

	logx.Infof("邮件发送成功: %s -> %v", msg.Subject, msg.To)
	return nil
}

// sendMailWithTLS 使用TLS发送邮件
func (e *EmailMiddleware) sendMailWithTLS(addr string, auth smtp.Auth, from string, to []string, msg []byte) error {
	// 建立TLS连接
	conn, err := tls.Dial("tcp", addr, &tls.Config{
		InsecureSkipVerify: false,
		ServerName:         e.config.Host,
	})
	if err != nil {
		return err
	}
	defer conn.Close()

	// 创建SMTP客户端
	client, err := smtp.NewClient(conn, e.config.Host)
	if err != nil {
		return err
	}
	defer client.Quit()

	// 认证
	if err = client.Auth(auth); err != nil {
		return err
	}

	// 设置发送者
	if err = client.Mail(from); err != nil {
		return err
	}

	// 设置收件人
	for _, addr := range to {
		if err = client.Rcpt(addr); err != nil {
			return err
		}
	}

	// 发送邮件内容
	w, err := client.Data()
	if err != nil {
		return err
	}
	defer w.Close()

	_, err = w.Write(msg)
	return err
}

// SendBatchEmail 批量发送邮件
func (e *EmailMiddleware) SendBatchEmail(ctx context.Context, messages []EmailMessage) error {
	for _, msg := range messages {
		if err := e.SendEmail(ctx, msg); err != nil {
			logx.Errorf("批量发送邮件失败: %v", err)
			return err
		}
		// 避免发送过快
		time.Sleep(100 * time.Millisecond)
	}
	return nil
}

// SendTemplateEmail 发送模板邮件
func (e *EmailMiddleware) SendTemplateEmail(ctx context.Context, to []string, subject, template string, data map[string]interface{}) error {
	// 这里可以集成模板引擎，如html/template
	// 简化实现，直接使用模板字符串
	body := template
	for k, v := range data {
		body = fmt.Sprintf(body, k, v)
	}

	msg := EmailMessage{
		To:      to,
		Subject: subject,
		Body:    body,
		IsHTML:  true,
	}

	return e.SendEmail(ctx, msg)
}
