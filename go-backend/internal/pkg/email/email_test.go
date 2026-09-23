package email

import (
	"net"
	"strings"
	"testing"
	"time"
)

func TestEmbeddedTemplateRendersWithoutPhysicalFiles(t *testing.T) {
	service, err := NewEmailService(testSMTPConfig())
	if err != nil {
		t.Fatalf("NewEmailService() error = %v", err)
	}

	emailService := service.(*emailService)
	body, isHTML, err := emailService.renderTemplate("welcome.html", "Welcome", map[string]string{
		"customer_name": "Taylor",
	})
	if err != nil {
		t.Fatalf("renderTemplate() error = %v", err)
	}
	if !isHTML {
		t.Fatal("renderTemplate() rendered fallback text, want HTML")
	}
	if !strings.Contains(body, "Taylor") {
		t.Fatalf("rendered body does not contain customer name: %s", body)
	}
}

func TestMissingHTMLTemplateFallsBackToPlainText(t *testing.T) {
	service, err := NewEmailService(testSMTPConfig())
	if err != nil {
		t.Fatalf("NewEmailService() error = %v", err)
	}

	emailService := service.(*emailService)
	body, isHTML, err := emailService.renderTemplate("missing.html", "Account update", nil)
	if err != nil {
		t.Fatalf("renderTemplate() error = %v", err)
	}
	if isHTML {
		t.Fatal("renderTemplate() rendered HTML, want plain text fallback")
	}
	if !strings.Contains(body, "Account update") {
		t.Fatalf("fallback body does not contain subject: %s", body)
	}
}

func TestSendEmailTimesOutWhenSMTPServerDoesNotRespond(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("net.Listen() error = %v", err)
	}
	defer listener.Close()

	done := make(chan struct{})
	go func() {
		defer close(done)
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		time.Sleep(500 * time.Millisecond)
	}()

	host, portText, err := net.SplitHostPort(listener.Addr().String())
	if err != nil {
		t.Fatalf("SplitHostPort() error = %v", err)
	}
	port := mustAtoi(t, portText)

	service, err := NewEmailService(&SMTPConfig{
		Host:     host,
		Port:     port,
		From:     "noreply@example.com",
		FromName: "Store Support",
		Timeout:  50 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("NewEmailService() error = %v", err)
	}

	start := time.Now()
	err = service.SendEmail([]string{"customer@example.com"}, "Test", "Hello")
	elapsed := time.Since(start)
	if err == nil {
		t.Fatal("SendEmail() error = nil, want timeout error")
	}
	if elapsed > time.Second {
		t.Fatalf("SendEmail() took %s, want bounded timeout", elapsed)
	}

	<-done
}

func TestLoadConfigFromEnvParsesSMTPTimeout(t *testing.T) {
	t.Setenv("SMTP_TIMEOUT", "250ms")
	t.Setenv("SMTP_REPLY_TO", "support@example.com")
	t.Setenv("SMTP_ENCRYPTION", "tls")

	config := LoadConfigFromEnv()
	if config.Timeout != 250*time.Millisecond {
		t.Fatalf("Timeout = %s, want 250ms", config.Timeout)
	}
	if config.ReplyTo != "support@example.com" {
		t.Fatalf("ReplyTo = %q, want support@example.com", config.ReplyTo)
	}
	if config.EncryptionType != "tls" {
		t.Fatalf("EncryptionType = %q, want tls", config.EncryptionType)
	}
}

func TestBuildMessageIncludesReplyTo(t *testing.T) {
	service, err := NewEmailService(&SMTPConfig{
		Host: "smtp.example.com", Port: 587, From: "noreply@example.com",
		FromName: "Store Support", ReplyTo: "support@example.com",
	})
	if err != nil {
		t.Fatalf("NewEmailService() error = %v", err)
	}
	message := service.(*emailService).buildMessage([]string{"customer@example.com"}, "Subject", "Body", false)
	if !strings.Contains(string(message), "Reply-To: support@example.com\r\n") {
		t.Fatalf("message does not contain Reply-To header: %s", message)
	}
}

func TestSMTPEncryptionTypeNoneDoesNotRequireSTARTTLS(t *testing.T) {
	service, err := NewEmailService(&SMTPConfig{
		Host: "smtp.example.com", Port: 25, From: "noreply@example.com",
		FromName: "Store Support", EncryptionType: "none",
	})
	if err != nil {
		t.Fatalf("NewEmailService() error = %v", err)
	}
	if service.(*emailService).config.EncryptionType != "none" {
		t.Fatalf("EncryptionType = %q, want none", service.(*emailService).config.EncryptionType)
	}
}

func testSMTPConfig() *SMTPConfig {
	return &SMTPConfig{
		Host:     "smtp.example.com",
		Port:     587,
		From:     "noreply@example.com",
		FromName: "Store Support",
		Timeout:  defaultSMTPTimeout,
	}
}

func mustAtoi(t *testing.T, value string) int {
	t.Helper()

	result := 0
	for _, char := range value {
		if char < '0' || char > '9' {
			t.Fatalf("invalid integer %q", value)
		}
		result = result*10 + int(char-'0')
	}
	return result
}
