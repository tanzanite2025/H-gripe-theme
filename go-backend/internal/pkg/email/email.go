package email

import (
	"bytes"
	"fmt"
	"html/template"
	"net/smtp"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// EmailService 邮件服务接口
type EmailService interface {
	SendEmail(to []string, subject, body string) error
	SendHTMLEmail(to []string, subject, templateName string, data interface{}) error
	SendOrderConfirmation(to string, orderData interface{}) error
	SendShippingNotification(to string, shippingData interface{}) error
	SendPasswordReset(to string, resetData interface{}) error
	SendWelcomeEmail(to string, userData interface{}) error
}

// SMTPConfig SMTP 配置
type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	FromName string
}

// emailService 邮件服务实现
type emailService struct {
	config    *SMTPConfig
	templates *template.Template
}

// NewEmailService 创建邮件服务
func NewEmailService(config *SMTPConfig) (EmailService, error) {
	// 加载邮件模板
	templatesPath := filepath.Join("templates", "email", "*.html")
	templates, err := template.ParseGlob(templatesPath)
	if err != nil {
		// 如果模板不存在，创建空模板
		templates = template.New("email")
	}

	return &emailService{
		config:    config,
		templates: templates,
	}, nil
}

// SendEmail 发送纯文本邮件
func (s *emailService) SendEmail(to []string, subject, body string) error {
	// 验证邮件地址
	if err := validateEmailAddresses(to); err != nil {
		return err
	}

	// 构建邮件内容
	message := s.buildMessage(to, subject, body, false)

	// 发送邮件
	return s.send(to, message)
}

// SendHTMLEmail 发送 HTML 邮件
func (s *emailService) SendHTMLEmail(to []string, subject, templateName string, data interface{}) error {
	// 验证邮件地址
	if err := validateEmailAddresses(to); err != nil {
		return err
	}

	// 渲染模板
	var buf bytes.Buffer
	err := s.templates.ExecuteTemplate(&buf, templateName, data)
	if err != nil {
		return fmt.Errorf("failed to render template: %w", err)
	}

	// 构建邮件内容
	message := s.buildMessage(to, subject, buf.String(), true)

	// 发送邮件
	return s.send(to, message)
}

// SendOrderConfirmation 发送订单确认邮件
func (s *emailService) SendOrderConfirmation(to string, orderData interface{}) error {
	return s.SendHTMLEmail(
		[]string{to},
		"订单确认 - Tanzanite",
		"order_confirmation.html",
		orderData,
	)
}

// SendShippingNotification 发送发货通知邮件
func (s *emailService) SendShippingNotification(to string, shippingData interface{}) error {
	return s.SendHTMLEmail(
		[]string{to},
		"您的订单已发货 - Tanzanite",
		"shipping_notification.html",
		shippingData,
	)
}

// SendPasswordReset 发送密码重置邮件
func (s *emailService) SendPasswordReset(to string, resetData interface{}) error {
	return s.SendHTMLEmail(
		[]string{to},
		"重置密码 - Tanzanite",
		"password_reset.html",
		resetData,
	)
}

// SendWelcomeEmail 发送欢迎邮件
func (s *emailService) SendWelcomeEmail(to string, userData interface{}) error {
	return s.SendHTMLEmail(
		[]string{to},
		"欢迎加入 Tanzanite",
		"welcome.html",
		userData,
	)
}

// buildMessage 构建邮件消息
func (s *emailService) buildMessage(to []string, subject, body string, isHTML bool) []byte {
	var buf bytes.Buffer

	// 邮件头
	fmt.Fprintf(&buf, "From: %s <%s>\r\n", s.config.FromName, s.config.From)
	fmt.Fprintf(&buf, "To: %s\r\n", strings.Join(to, ", "))
	fmt.Fprintf(&buf, "Subject: %s\r\n", subject)
	buf.WriteString("MIME-Version: 1.0\r\n")

	if isHTML {
		buf.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	} else {
		buf.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	}

	buf.WriteString("\r\n")
	buf.WriteString(body)

	return buf.Bytes()
}

// send 发送邮件
func (s *emailService) send(to []string, message []byte) error {
	// SMTP 认证
	auth := smtp.PlainAuth(
		"",
		s.config.Username,
		s.config.Password,
		s.config.Host,
	)

	// 发送邮件
	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	err := smtp.SendMail(addr, auth, s.config.From, to, message)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

// LoadConfigFromEnv 从环境变量加载配置
func LoadConfigFromEnv() *SMTPConfig {
	return &SMTPConfig{
		Host:     getEnv("SMTP_HOST", "smtp.gmail.com"),
		Port:     getEnvInt("SMTP_PORT", 587),
		Username: getEnv("SMTP_USERNAME", ""),
		Password: getEnv("SMTP_PASSWORD", ""),
		From:     getEnv("SMTP_FROM", "noreply@tanzanite.com"),
		FromName: getEnv("SMTP_FROM_NAME", "Tanzanite"),
	}
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// getEnvInt 获取整数环境变量
func getEnvInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return intValue
}

// validateEmailAddresses 验证邮件地址
func validateEmailAddresses(emails []string) error {
	if len(emails) == 0 {
		return fmt.Errorf("no email addresses provided")
	}

	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

	for _, email := range emails {
		if !emailRegex.MatchString(email) {
			return fmt.Errorf("invalid email address: %s", email)
		}
	}

	return nil
}
