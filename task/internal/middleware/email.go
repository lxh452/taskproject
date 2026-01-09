package middleware

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

// EmailConfig 邮件配置
type EmailConfig struct {
	Enabled  bool   `json:"enabled"`  // 是否启用邮件发送
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
	// 检查邮件功能是否启用
	if !e.config.Enabled {
		logx.Infof("[EmailMiddleware] Email sending is disabled, skipping: subject=%s, to=%v", msg.Subject, msg.To)
		return nil
	}

	if len(msg.To) == 0 || msg.To[0] == "" {
		return fmt.Errorf("recipient email is empty")
	}
	// 构建邮件头
	headers := make(map[string]string)
	headers["From"] = fmt.Sprintf("%s <%s>", e.config.From, e.config.Username)
	headers["To"] = msg.To[0]
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

	var err error
	switch {
	case e.config.UseTLS && e.config.Port == 465:
		// SMTPS implicit TLS
		logx.Infof("[EmailMiddleware] SMTP send mode=SMTPS addr=%s from=%s to=%v subject=%s",
			addr, e.config.Username, msg.To, msg.Subject)
		err = e.sendMailWithTLS(addr, auth, e.config.Username, msg.To, []byte(emailBody))
		if err != nil {
			logx.Errorf("[EmailMiddleware] SMTPS send failed: error=%v, addr=%s, from=%s, to=%v",
				err, addr, e.config.Username, msg.To)
		}
	case e.config.UseTLS && e.config.Port == 587:
		// STARTTLS on submission port
		logx.Infof("[EmailMiddleware] SMTP send mode=STARTTLS addr=%s from=%s to=%v subject=%s",
			addr, e.config.Username, msg.To, msg.Subject)
		err = e.sendMailWithSTARTTLS(addr, auth, e.config.Host, e.config.Username, msg.To, []byte(emailBody))
		if err != nil {
			logx.Errorf("[EmailMiddleware] STARTTLS send failed: error=%v, addr=%s, from=%s, to=%v",
				err, addr, e.config.Username, msg.To)
		}
	default:
		// Plain SMTP (for debug environments)
		logx.Infof("[EmailMiddleware] SMTP send mode=PLAIN addr=%s from=%s to=%v subject=%s",
			addr, e.config.Username, msg.To, msg.Subject)
		err = e.sendMailPlain(addr, auth, e.config.Username, msg.To, []byte(emailBody))
		if err != nil {
			logx.Errorf("[EmailMiddleware] Plain SMTP send failed: error=%v, addr=%s, from=%s, to=%v",
				err, addr, e.config.Username, msg.To)
		}
	}
	if err != nil {
		logx.Errorf("[EmailMiddleware] 发送邮件失败: error=%v, subject=%s, to=%v", err, msg.Subject, msg.To)
		return err
	}

	logx.Infof("[EmailMiddleware] 邮件发送成功: subject=%s, from=%s, to=%v", msg.Subject, e.config.Username, msg.To)
	return nil
}

// sendMailWithTLS 使用TLS发送邮件
func (e *EmailMiddleware) sendMailWithTLS(addr string, auth smtp.Auth, from string, to []string, msg []byte) error {
	logx.Infof("[EmailMiddleware] Establishing TLS connection to %s (host=%s)", addr, e.config.Host)

	// 增加超时时间到 60 秒，QQ 邮箱可能响应较慢
	dialTimeout := 60 * time.Second

	// 方法1: 尝试直接使用 tls.Dial（QQ 邮箱可能更适合这种方式）
	tlsConfig := &tls.Config{
		ServerName:         e.config.Host,
		InsecureSkipVerify: true, // QQ 邮箱可能需要跳过证书验证
		MinVersion:         tls.VersionTLS10,
		MaxVersion:         tls.VersionTLS13,
	}

	logx.Infof("[EmailMiddleware] Attempting direct TLS dial with timeout=%v", dialTimeout)
	logx.Infof("[EmailMiddleware] NOTE: If you're using VPN, it may interfere with TLS connections. Try disabling VPN.")

	startTime := time.Now()
	conn, err := tls.DialWithDialer(&net.Dialer{Timeout: dialTimeout}, "tcp", addr, tlsConfig)
	dialDuration := time.Since(startTime)

	if err != nil {
		logx.Errorf("[EmailMiddleware] Direct TLS dial failed: error=%v, addr=%s, host=%s, timeout=%v, duration=%v",
			err, addr, e.config.Host, dialTimeout, dialDuration)
		logx.Errorf("[EmailMiddleware] TROUBLESHOOTING: This error might be caused by:")
		logx.Errorf("[EmailMiddleware]   1. VPN interference - Try disabling VPN or adding SMTP exception")
		logx.Errorf("[EmailMiddleware]   2. Firewall blocking port 465")
		logx.Errorf("[EmailMiddleware]   3. Network connectivity issues")
		logx.Errorf("[EmailMiddleware]   4. TLS handshake timeout (server not responding)")

		// 方法2: 尝试先建立 TCP 连接再升级（分步进行，更容易定位问题）
		logx.Infof("[EmailMiddleware] Trying TCP connection then TLS upgrade with timeout=%v", dialTimeout)
		dialer := &net.Dialer{
			Timeout:   dialTimeout,
			KeepAlive: 30 * time.Second, // 保持连接活跃
		}
		tcpConn, err2 := dialer.Dial("tcp", addr)
		if err2 != nil {
			logx.Errorf("[EmailMiddleware] TCP dial failed: error=%v, addr=%s, timeout=%v", err2, addr, dialTimeout)
			// 方法3: 尝试使用标准库（可能内部有更好的处理）
			logx.Infof("[EmailMiddleware] Trying standard library smtp.SendMail as fallback")
			return e.sendMailWithStandardLibrary(addr, auth, from, to, msg)
		}
		logx.Infof("[EmailMiddleware] TCP connection established, upgrading to TLS")

		// 设置 TCP 连接的读写超时
		if tcpConn != nil {
			if tcp, ok := tcpConn.(*net.TCPConn); ok {
				tcp.SetKeepAlive(true)
				tcp.SetKeepAlivePeriod(30 * time.Second)
			}
		}

		conn = tls.Client(tcpConn, tlsConfig)

		// 设置 TLS 握手超时
		handshakeTimeout := 30 * time.Second
		logx.Infof("[EmailMiddleware] Performing TLS handshake with timeout=%v", handshakeTimeout)
		handshakeDone := make(chan error, 1)
		go func() {
			handshakeDone <- conn.Handshake()
		}()

		select {
		case err := <-handshakeDone:
			if err != nil {
				tcpConn.Close()
				logx.Errorf("[EmailMiddleware] TLS handshake failed: error=%v, timeout=%v", err, handshakeTimeout)
				// 最后尝试标准库
				logx.Infof("[EmailMiddleware] Trying standard library smtp.SendMail as last resort")
				return e.sendMailWithStandardLibrary(addr, auth, from, to, msg)
			}
			logx.Infof("[EmailMiddleware] TLS handshake successful")
		case <-time.After(handshakeTimeout):
			tcpConn.Close()
			logx.Errorf("[EmailMiddleware] TLS handshake timeout after %v", handshakeTimeout)
			// 最后尝试标准库
			logx.Infof("[EmailMiddleware] Trying standard library smtp.SendMail as last resort")
			return e.sendMailWithStandardLibrary(addr, auth, from, to, msg)
		}
	} else {
		logx.Infof("[EmailMiddleware] Direct TLS dial successful")
	}
	defer conn.Close()

	// 创建SMTP客户端
	logx.Infof("[EmailMiddleware] Creating SMTP client for host=%s", e.config.Host)
	client, err := smtp.NewClient(conn, e.config.Host)
	if err != nil {
		logx.Errorf("[EmailMiddleware] Failed to create SMTP client: error=%v, host=%s", err, e.config.Host)
		return fmt.Errorf("create SMTP client failed: %w", err)
	}
	defer client.Quit()
	logx.Infof("[EmailMiddleware] SMTP client created successfully")

	// 认证
	logx.Infof("[EmailMiddleware] Authenticating with SMTP server, username=%s", e.config.Username)
	if err = client.Auth(auth); err != nil {
		logx.Errorf("[EmailMiddleware] SMTP authentication failed: error=%v, username=%s, host=%s",
			err, e.config.Username, e.config.Host)
		return fmt.Errorf("SMTP authentication failed: %w", err)
	}
	logx.Infof("[EmailMiddleware] SMTP authentication successful")

	// 设置发送者
	logx.Infof("[EmailMiddleware] Setting mail from=%s", from)
	if err = client.Mail(from); err != nil {
		logx.Errorf("[EmailMiddleware] Failed to set mail from: error=%v, from=%s", err, from)
		return fmt.Errorf("set mail from failed: %w", err)
	}
	logx.Infof("[EmailMiddleware] Mail from set successfully")

	// 设置收件人
	logx.Infof("[EmailMiddleware] Setting recipients: %v", to)
	for _, addr := range to {
		if err = client.Rcpt(addr); err != nil {
			logx.Errorf("[EmailMiddleware] Failed to set recipient: error=%v, recipient=%s", err, addr)
			return fmt.Errorf("set recipient failed for %s: %w", addr, err)
		}
		logx.Infof("[EmailMiddleware] Recipient set successfully: %s", addr)
	}

	// 发送邮件内容
	logx.Infof("[EmailMiddleware] Sending email data, message size=%d bytes", len(msg))
	w, err := client.Data()
	if err != nil {
		logx.Errorf("[EmailMiddleware] Failed to open data connection: error=%v", err)
		return fmt.Errorf("open data connection failed: %w", err)
	}
	defer w.Close()
	logx.Infof("[EmailMiddleware] Data connection opened, writing message")

	_, err = w.Write(msg)
	if err != nil {
		logx.Errorf("[EmailMiddleware] Failed to write message: error=%v, messageSize=%d", err, len(msg))
		return fmt.Errorf("write message failed: %w", err)
	}
	logx.Infof("[EmailMiddleware] Message written successfully, closing data connection")
	return nil
}

// sendMailWithSTARTTLS 使用 STARTTLS 发送邮件 (常用于 587)
func (e *EmailMiddleware) sendMailWithSTARTTLS(addr string, auth smtp.Auth, serverName, from string, to []string, msg []byte) error {
	logx.Infof("[EmailMiddleware] Dialing SMTP server with STARTTLS: %s (host=%s)", addr, serverName)
	logx.Infof("[EmailMiddleware] NOTE: If you're using VPN, it may interfere with SMTP connections. Try disabling VPN or adding SMTP exception.")

	// 直接使用 smtp.Dial，它会自动建立 TCP 连接并读取服务器欢迎消息
	// 这比手动创建连接再调用 smtp.NewClient 更可靠
	startTime := time.Now()
	logx.Infof("[EmailMiddleware] Dialing SMTP server (will establish TCP and read greeting)")
	c, err := smtp.Dial(addr)
	dialDuration := time.Since(startTime)

	if err != nil {
		logx.Errorf("[EmailMiddleware] Failed to dial SMTP server: error=%v, addr=%s, duration=%v", err, addr, dialDuration)
		logx.Errorf("[EmailMiddleware] TROUBLESHOOTING: EOF error usually means:")
		logx.Errorf("[EmailMiddleware]   1. VPN is interfering with SMTP protocol handshake - TRY DISABLING VPN")
		logx.Errorf("[EmailMiddleware]   2. Server closed connection during greeting")
		logx.Errorf("[EmailMiddleware]   3. Network interruption after TCP connection")
		logx.Errorf("[EmailMiddleware]   4. Firewall blocking SMTP protocol")
		logx.Infof("[EmailMiddleware] Trying standard library smtp.SendMail as fallback")
		// 尝试使用标准库作为备用方案
		return e.sendMailWithStandardLibrary(addr, auth, from, to, msg)
	}
	defer c.Quit()
	logx.Infof("[EmailMiddleware] SMTP client dialed successfully in %v, server greeting received", dialDuration)

	// 检查是否支持 STARTTLS
	if ok, _ := c.Extension("STARTTLS"); ok {
		logx.Infof("[EmailMiddleware] Server supports STARTTLS, upgrading to TLS, serverName=%s", serverName)
		cfg := &tls.Config{
			ServerName:         serverName,
			InsecureSkipVerify: false, // STARTTLS 通常可以验证证书
		}
		if err := c.StartTLS(cfg); err != nil {
			logx.Errorf("[EmailMiddleware] STARTTLS failed: error=%v, serverName=%s", err, serverName)
			// 如果标准配置失败，尝试跳过证书验证
			logx.Infof("[EmailMiddleware] Retrying STARTTLS with relaxed TLS config")
			cfg.InsecureSkipVerify = true
			if err := c.StartTLS(cfg); err != nil {
				logx.Errorf("[EmailMiddleware] STARTTLS failed even with relaxed config: error=%v", err)
				return fmt.Errorf("STARTTLS failed: %w", err)
			}
			logx.Infof("[EmailMiddleware] STARTTLS successful with relaxed config")
		} else {
			logx.Infof("[EmailMiddleware] STARTTLS upgrade successful")
		}
	} else {
		logx.Infof("[EmailMiddleware] STARTTLS not supported by server, continuing without TLS (not recommended)")
	}

	// 认证
	logx.Infof("[EmailMiddleware] Authenticating, username=%s", e.config.Username)
	if err := c.Auth(auth); err != nil {
		logx.Errorf("[EmailMiddleware] Authentication failed: error=%v, username=%s", err, e.config.Username)
		return fmt.Errorf("authentication failed: %w", err)
	}
	logx.Infof("[EmailMiddleware] Authentication successful")

	// 发送者
	logx.Infof("[EmailMiddleware] Setting mail from=%s", from)
	if err := c.Mail(from); err != nil {
		logx.Errorf("[EmailMiddleware] Failed to set mail from: error=%v, from=%s", err, from)
		return fmt.Errorf("set mail from failed: %w", err)
	}
	logx.Infof("[EmailMiddleware] Mail from set successfully")

	// 收件人
	logx.Infof("[EmailMiddleware] Setting recipients: %v", to)
	for _, r := range to {
		if err := c.Rcpt(r); err != nil {
			logx.Errorf("[EmailMiddleware] Failed to set recipient: error=%v, recipient=%s", err, r)
			return fmt.Errorf("set recipient failed for %s: %w", r, err)
		}
		logx.Infof("[EmailMiddleware] Recipient set successfully: %s", r)
	}

	// 数据
	logx.Infof("[EmailMiddleware] Opening data connection, message size=%d bytes", len(msg))
	w, err := c.Data()
	if err != nil {
		logx.Errorf("[EmailMiddleware] Failed to open data connection: error=%v", err)
		return fmt.Errorf("open data connection failed: %w", err)
	}
	logx.Infof("[EmailMiddleware] Data connection opened, writing message")

	if _, err := w.Write(msg); err != nil {
		_ = w.Close()
		logx.Errorf("[EmailMiddleware] Failed to write message: error=%v, messageSize=%d", err, len(msg))
		return fmt.Errorf("write message failed: %w", err)
	}
	logx.Infof("[EmailMiddleware] Message written successfully, closing data connection")
	return w.Close()
}

// sendMailWithStandardLibrary 使用标准库的 smtp.SendMail 发送邮件（备用方案）
func (e *EmailMiddleware) sendMailWithStandardLibrary(addr string, auth smtp.Auth, from string, to []string, msg []byte) error {
	logx.Infof("[EmailMiddleware] Using standard library smtp.SendMail as fallback")

	// 构建邮件头
	headers := make(map[string]string)
	headers["From"] = fmt.Sprintf("%s <%s>", e.config.From, e.config.Username)
	headers["To"] = to[0]
	headers["Subject"] = ""
	headers["Date"] = time.Now().Format(time.RFC1123Z)
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=UTF-8"

	// 解析消息体以提取主题
	msgStr := string(msg)
	lines := strings.Split(msgStr, "\r\n")
	for _, line := range lines {
		if strings.HasPrefix(strings.ToLower(line), "subject:") {
			headers["Subject"] = strings.TrimPrefix(line, "Subject:")
			headers["Subject"] = strings.TrimSpace(headers["Subject"])
			break
		}
	}

	// 构建邮件内容
	emailBody := ""
	for k, v := range headers {
		if k != "Subject" || headers["Subject"] != "" {
			emailBody += fmt.Sprintf("%s: %s\r\n", k, v)
		}
	}
	// 找到消息体的开始（空行后）
	bodyStart := strings.Index(msgStr, "\r\n\r\n")
	if bodyStart > 0 {
		emailBody += msgStr[bodyStart+4:]
	} else {
		emailBody += msgStr
	}

	// 使用标准库发送
	err := smtp.SendMail(addr, auth, from, to, []byte(emailBody))
	if err != nil {
		logx.Errorf("[EmailMiddleware] Standard library smtp.SendMail failed: error=%v", err)
		return fmt.Errorf("smtp.SendMail failed: %w", err)
	}

	logx.Infof("[EmailMiddleware] Email sent successfully using standard library")
	return nil
}

// sendMailPlain 使用明文 SMTP 发送（不推荐生产，仅用于本地调试）
func (e *EmailMiddleware) sendMailPlain(addr string, auth smtp.Auth, from string, to []string, msg []byte) error {
	c, err := smtp.Dial(addr)
	if err != nil {
		return err
	}
	defer c.Quit()

	// 某些服务器要求先 EHLO 后 AUTH，smtp.Dial 已处理 EHLO
	if err := c.Auth(auth); err != nil {
		return err
	}
	if err := c.Mail(from); err != nil {
		return err
	}
	for _, r := range to {
		if err := c.Rcpt(r); err != nil {
			return err
		}
	}
	w, err := c.Data()
	if err != nil {
		return err
	}
	if _, err := w.Write(msg); err != nil {
		_ = w.Close()
		return err
	}
	return w.Close()
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
