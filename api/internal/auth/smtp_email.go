package auth

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/mail"
	"net/smtp"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type SMTPConfig struct {
	Host       string
	Port       string
	Username   string
	Password   string
	From       string
	RequireTLS bool
}

type SMTPEmailSender struct {
	cfg SMTPConfig
}

func NewSMTPEmailSender(cfg SMTPConfig) *SMTPEmailSender {
	return &SMTPEmailSender{cfg: cfg}
}

func (s *SMTPEmailSender) SendPasswordReset(ctx context.Context, email string, link string) error {
	to, err := mail.ParseAddress(email)
	if err != nil || to.Address != email {
		return errors.New("invalid password recovery recipient")
	}
	from, err := mail.ParseAddress(s.cfg.From)
	if err != nil {
		return fmt.Errorf("invalid SMTP sender address: %w", err)
	}
	resetURL, err := url.ParseRequestURI(link)
	if err != nil || (resetURL.Scheme != "https" && resetURL.Scheme != "http") || resetURL.Host == "" ||
		strings.ContainsAny(link, "\r\n") {
		return errors.New("invalid password reset link")
	}

	port, err := strconv.Atoi(s.cfg.Port)
	if err != nil || port < 1 || port > 65535 {
		return errors.New("invalid SMTP port")
	}
	address := net.JoinHostPort(s.cfg.Host, strconv.Itoa(port))
	dialer := net.Dialer{Timeout: 10 * time.Second}
	conn, err := dialer.DialContext(ctx, "tcp", address)
	if err != nil {
		return fmt.Errorf("connect to SMTP server: %w", err)
	}
	defer func() {
		_ = conn.Close()
	}()

	deadline := time.Now().Add(10 * time.Second)
	if ctxDeadline, ok := ctx.Deadline(); ok && ctxDeadline.Before(deadline) {
		deadline = ctxDeadline
	}
	if err := conn.SetDeadline(deadline); err != nil {
		return fmt.Errorf("set SMTP connection deadline: %w", err)
	}
	stopCancel := context.AfterFunc(ctx, func() {
		_ = conn.SetDeadline(time.Now())
	})
	defer stopCancel()

	if port == 465 {
		tlsConn := tls.Client(conn, &tls.Config{
			ServerName: s.cfg.Host,
			MinVersion: tls.VersionTLS12,
		})
		if err := tlsConn.HandshakeContext(ctx); err != nil {
			return fmt.Errorf("start SMTP TLS: %w", err)
		}
		conn = tlsConn
	}

	client, err := smtp.NewClient(conn, s.cfg.Host)
	if err != nil {
		return fmt.Errorf("initialize SMTP client: %w", err)
	}
	defer client.Close()

	if port != 465 {
		if ok, _ := client.Extension("STARTTLS"); ok {
			if err := client.StartTLS(&tls.Config{
				ServerName: s.cfg.Host,
				MinVersion: tls.VersionTLS12,
			}); err != nil {
				return fmt.Errorf("start SMTP TLS: %w", err)
			}
		} else if s.cfg.RequireTLS || s.cfg.Username != "" {
			return errors.New("SMTP server does not support required STARTTLS")
		}
	}
	if s.cfg.Username != "" {
		if err := client.Auth(smtp.PlainAuth("", s.cfg.Username, s.cfg.Password, s.cfg.Host)); err != nil {
			return fmt.Errorf("authenticate with SMTP server: %w", err)
		}
	}

	if err := client.Mail(from.Address); err != nil {
		return fmt.Errorf("set SMTP sender: %w", err)
	}
	if err := client.Rcpt(to.Address); err != nil {
		return fmt.Errorf("set SMTP recipient: %w", err)
	}
	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("start SMTP message: %w", err)
	}
	message := "From: " + from.String() + "\r\n" +
		"To: " + to.String() + "\r\n" +
		"Subject: Reset your Year Planner password\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/plain; charset=UTF-8\r\n" +
		"\r\n" +
		"Use the following link to reset your password. This link expires in one hour and can only be used once:\r\n" +
		link + "\r\n"
	if _, err := writer.Write([]byte(message)); err != nil {
		_ = writer.Close()
		return fmt.Errorf("write SMTP message: %w", err)
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("send SMTP message: %w", err)
	}
	if err := client.Quit(); err != nil {
		return fmt.Errorf("close SMTP session: %w", err)
	}
	return nil
}

var _ PasswordResetEmailSender = (*SMTPEmailSender)(nil)
