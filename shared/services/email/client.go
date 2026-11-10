package email

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"net/smtp"
	"strconv"
	"strings"
	"time"
)

const sendTimeout = 30 * time.Second

var ErrSend = errors.New("邮件发送失败，请检查SMTP连接及登录凭据")

// Client 为调用方提供独立的SMTP发送连接。
type Client struct {
	config      Config
	dialContext func(context.Context, string, string) (net.Conn, error)
	tlsConfig   *tls.Config
}

// New 创建可并发使用的邮件客户端，初始化不连接SMTP，每次发送使用独立连接。
func New(config Config) (*Client, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}
	dialer := &net.Dialer{Timeout: sendTimeout}
	return &Client{config: config, dialContext: dialer.DialContext, tlsConfig: &tls.Config{ServerName: config.EmailServerHost, MinVersion: tls.VersionTLS12}}, nil
}

// Send 使用经过证书验证的TLS连接发送，不自动重试，不记录邮件及凭据。
func (c *Client) Send(ctx context.Context, message *Message) error {
	if c == nil || c.dialContext == nil || c.tlsConfig == nil {
		return ErrConfig
	}
	if ctx == nil {
		return ErrMessage
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	from, recipients, body, err := c.compose(message)
	if err != nil {
		return err
	}
	sendContext, cancel := context.WithTimeout(ctx, sendTimeout)
	defer cancel()
	address := net.JoinHostPort(c.config.EmailServerHost, strconv.Itoa(c.config.EmailServerPort))
	conn, err := c.dialContext(sendContext, "tcp", address)
	if err != nil {
		return sendError(sendContext)
	}
	defer conn.Close()
	deadline, _ := sendContext.Deadline()
	if conn.SetDeadline(deadline) != nil {
		return sendError(sendContext)
	}
	rawConn := conn
	stop := context.AfterFunc(sendContext, func() { _ = rawConn.Close() })
	defer stop()
	if c.config.EmailServerPort == 465 {
		tlsConn := tls.Client(conn, c.tlsConfig.Clone())
		if tlsConn.HandshakeContext(sendContext) != nil {
			return sendError(sendContext)
		}
		conn = tlsConn
	}
	client, err := smtp.NewClient(conn, c.config.EmailServerHost)
	if err != nil {
		return sendError(sendContext)
	}
	defer client.Close()
	if c.config.EmailServerPort != 465 {
		if supported, _ := client.Extension("STARTTLS"); !supported || client.StartTLS(c.tlsConfig.Clone()) != nil {
			return sendError(sendContext)
		}
	}
	if c.config.EmailUsername != "" {
		supported, mechanisms := client.Extension("AUTH")
		auth := c.authentication(mechanisms)
		if !supported || auth == nil || client.Auth(auth) != nil {
			return sendError(sendContext)
		}
	}
	if client.Mail(from) != nil {
		return sendError(sendContext)
	}
	for _, recipient := range recipients {
		if client.Rcpt(recipient) != nil {
			return sendError(sendContext)
		}
	}
	writer, err := client.Data()
	if err != nil {
		return sendError(sendContext)
	}
	if _, err := writer.Write(body); err != nil {
		return sendError(sendContext)
	}
	if writer.Close() != nil {
		return sendError(sendContext)
	}
	// DATA已获服务器确认，退出连接失败不改报发送失败，避免调用方重复投递。
	_ = client.Quit()
	return nil
}

func sendError(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return ErrSend
}

func (c *Client) authentication(mechanisms string) smtp.Auth {
	for _, mechanism := range strings.Fields(mechanisms) {
		if mechanism == "PLAIN" {
			return smtp.PlainAuth("", c.config.EmailUsername, c.config.EmailPassword, c.config.EmailServerHost)
		}
	}
	for _, mechanism := range strings.Fields(mechanisms) {
		switch mechanism {
		case "LOGIN":
			return &loginAuth{host: c.config.EmailServerHost, username: c.config.EmailUsername, password: c.config.EmailPassword}
		case "CRAM-MD5":
			return smtp.CRAMMD5Auth(c.config.EmailUsername, c.config.EmailPassword)
		}
	}
	return nil
}

// loginAuth 兼容仅提供LOGIN认证的SMTP服务，凭据只在TLS连接中传递。
type loginAuth struct {
	host, username, password string
}

func (a *loginAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	if !server.TLS || server.Name != a.host {
		return "", nil, ErrSend
	}
	return "LOGIN", nil, nil
}

func (a *loginAuth) Next(challenge []byte, more bool) ([]byte, error) {
	if !more {
		return nil, nil
	}
	switch strings.ToLower(strings.TrimSpace(string(challenge))) {
	case "username:":
		return []byte(a.username), nil
	case "password:":
		return []byte(a.password), nil
	default:
		return nil, ErrSend
	}
}
