package auth

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"
)

func TestSMTPEmailSenderSendsResetMessage(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	messageCh := make(chan string, 1)
	serverErrCh := make(chan error, 1)
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			serverErrCh <- err
			return
		}
		defer conn.Close()
		_ = conn.SetDeadline(time.Now().Add(5 * time.Second))
		reader := bufio.NewReader(conn)
		if _, err := fmt.Fprint(conn, "220 localhost ESMTP\r\n"); err != nil {
			serverErrCh <- err
			return
		}
		var message strings.Builder
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				serverErrCh <- err
				return
			}
			switch {
			case strings.HasPrefix(line, "EHLO "):
				_, err = fmt.Fprint(conn, "250-localhost\r\n250 SIZE 100000\r\n")
			case strings.HasPrefix(line, "MAIL FROM:"):
				_, err = fmt.Fprint(conn, "250 sender accepted\r\n")
			case strings.HasPrefix(line, "RCPT TO:"):
				_, err = fmt.Fprint(conn, "250 recipient accepted\r\n")
			case strings.TrimSpace(line) == "DATA":
				_, err = fmt.Fprint(conn, "354 send message\r\n")
				if err != nil {
					serverErrCh <- err
					return
				}
				for {
					line, err = reader.ReadString('\n')
					if err != nil {
						serverErrCh <- err
						return
					}
					if line == ".\r\n" {
						break
					}
					message.WriteString(line)
				}
				messageCh <- message.String()
				_, err = fmt.Fprint(conn, "250 message accepted\r\n")
			case strings.TrimSpace(line) == "QUIT":
				_, err = fmt.Fprint(conn, "221 bye\r\n")
				if err != nil {
					serverErrCh <- err
				}
				return
			default:
				_, err = fmt.Fprint(conn, "250 ok\r\n")
			}
			if err != nil {
				serverErrCh <- err
				return
			}
		}
	}()

	_, port, err := net.SplitHostPort(listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	link := "https://planner.example.com/#/reset-password?token=sample"
	sender := NewSMTPEmailSender(SMTPConfig{
		Host: "127.0.0.1",
		Port: port,
		From: "Year Planner <no-reply@example.com>",
	})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := sender.SendPasswordReset(ctx, "user@example.com", link); err != nil {
		t.Fatalf("SendPasswordReset() error = %v", err)
	}

	select {
	case message := <-messageCh:
		for _, expected := range []string{
			"To: <user@example.com>",
			"From: \"Year Planner\" <no-reply@example.com>",
			"Subject: Reset your Year Planner password",
			link,
		} {
			if !strings.Contains(message, expected) {
				t.Fatalf("SMTP message missing %q: %q", expected, message)
			}
		}
	case err := <-serverErrCh:
		t.Fatalf("SMTP test server failed: %v", err)
	case <-ctx.Done():
		t.Fatal("SMTP test server did not receive a message")
	}
}
