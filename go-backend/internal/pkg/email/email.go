package email

import (
	"bytes"
	"crypto/tls"
	"embed"
	"fmt"
	"html/template"
	"net"
	"net/smtp"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const defaultSMTPTimeout = 10 * time.Second

//go:embed templates/*.html
var embeddedTemplates embed.FS

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
	Timeout  time.Duration
}

// emailService 邮件服务实现
type emailService struct {
	config    *SMTPConfig
	templates *template.Template
}

// NewEmailService 创建邮件服务
func NewEmailService(config *SMTPConfig) (EmailService, error) {
	if config == nil {
		return nil, fmt.Errorf("smtp config is required")
	}

	templates, err := template.New("email").Funcs(template.FuncMap{
		"field": templateField,
	}).ParseFS(embeddedTemplates, "templates/*.html")
	if err != nil {
		return nil, fmt.Errorf("failed to parse embedded email templates: %w", err)
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
	body, isHTML, err := s.renderTemplate(templateName, subject, data)
	if err != nil {
		return err
	}

	// 构建邮件内容
	message := s.buildMessage(to, subject, body, isHTML)

	// 发送邮件
	return s.send(to, message)
}

// SendOrderConfirmation 发送订单确认邮件
func (s *emailService) SendOrderConfirmation(to string, orderData interface{}) error {
	return s.SendHTMLEmail(
		[]string{to},
		"订单确认",
		"order_confirmation.html",
		orderData,
	)
}

// SendShippingNotification 发送发货通知邮件
func (s *emailService) SendShippingNotification(to string, shippingData interface{}) error {
	return s.SendHTMLEmail(
		[]string{to},
		"您的订单已发货",
		"shipping_notification.html",
		shippingData,
	)
}

// SendPasswordReset 发送密码重置邮件
func (s *emailService) SendPasswordReset(to string, resetData interface{}) error {
	return s.SendHTMLEmail(
		[]string{to},
		"重置密码",
		"password_reset.html",
		resetData,
	)
}

// SendWelcomeEmail 发送欢迎邮件
func (s *emailService) SendWelcomeEmail(to string, userData interface{}) error {
	return s.SendHTMLEmail(
		[]string{to},
		"欢迎加入",
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

func (s *emailService) renderTemplate(templateName, subject string, data interface{}) (string, bool, error) {
	if s.templates.Lookup(templateName) == nil {
		return buildPlainTextFallback(subject, templateName, data), false, nil
	}

	var buf bytes.Buffer
	err := s.templates.ExecuteTemplate(&buf, templateName, data)
	if err != nil {
		return "", false, fmt.Errorf("failed to render template: %w", err)
	}

	return buf.String(), true, nil
}

func buildPlainTextFallback(subject, templateName string, data interface{}) string {
	lines := []string{
		subject,
		"",
	}

	switch templateName {
	case "order_confirmation.html":
		lines = append(lines, "Thank you for your order. We have received it and will send another update when it ships.")
	case "shipping_notification.html":
		lines = append(lines, "Your order has shipped. Please check your account or the carrier tracking page for the latest delivery status.")
	case "password_reset.html":
		resetURL := templateField(data, "ResetURL", "ResetLink", "URL", "Link", "reset_url", "reset_link")
		if resetURL != "" {
			lines = append(lines, "Use this link to reset your password:", resetURL)
		} else {
			lines = append(lines, "We received a password reset request. Please return to the store and request a new reset link if needed.")
		}
	case "welcome.html":
		lines = append(lines, "Welcome. Your account is ready, and we are glad to have you with us.")
	default:
		lines = append(lines, "We have an update for you. Please sign in to your account for more details.")
	}

	reference := templateField(data, "OrderNumber", "OrderNo", "Number", "TrackingNumber", "order_number", "tracking_number")
	if reference != "" {
		lines = append(lines, "", "Reference: "+reference)
	}

	return strings.Join(lines, "\r\n")
}

// send 发送邮件
func (s *emailService) send(to []string, message []byte) error {
	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)

	timeout := s.config.Timeout
	if timeout <= 0 {
		timeout = defaultSMTPTimeout
	}

	dialer := &net.Dialer{Timeout: timeout}
	conn, err := dialer.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to connect smtp server: %w", err)
	}
	defer conn.Close()

	if err := setDeadline(conn, timeout); err != nil {
		return fmt.Errorf("failed to set smtp deadline: %w", err)
	}

	clientConn := conn
	if s.config.Port == 465 {
		tlsConn := tls.Client(conn, smtpTLSConfig(s.config.Host))
		if err := tlsConn.Handshake(); err != nil {
			return fmt.Errorf("failed smtp tls handshake: %w", err)
		}
		clientConn = tlsConn
	}

	client, err := smtp.NewClient(clientConn, s.config.Host)
	if err != nil {
		return fmt.Errorf("failed to create smtp client: %w", err)
	}
	defer client.Close()

	if s.config.Port != 465 {
		if ok, _ := client.Extension("STARTTLS"); ok {
			if err := setDeadline(conn, timeout); err != nil {
				return fmt.Errorf("failed to set smtp deadline: %w", err)
			}
			if err := client.StartTLS(smtpTLSConfig(s.config.Host)); err != nil {
				return fmt.Errorf("failed to start smtp tls: %w", err)
			}
		}
	}

	if s.config.Username != "" || s.config.Password != "" {
		auth := smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.Host)
		if err := setDeadline(conn, timeout); err != nil {
			return fmt.Errorf("failed to set smtp deadline: %w", err)
		}
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("failed smtp auth: %w", err)
		}
	}

	if err := setDeadline(conn, timeout); err != nil {
		return fmt.Errorf("failed to set smtp deadline: %w", err)
	}
	if err := client.Mail(s.config.From); err != nil {
		return fmt.Errorf("failed to set smtp sender: %w", err)
	}

	for _, recipient := range to {
		if err := setDeadline(conn, timeout); err != nil {
			return fmt.Errorf("failed to set smtp deadline: %w", err)
		}
		if err := client.Rcpt(recipient); err != nil {
			return fmt.Errorf("failed to set smtp recipient: %w", err)
		}
	}

	if err := setDeadline(conn, timeout); err != nil {
		return fmt.Errorf("failed to set smtp deadline: %w", err)
	}
	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to open smtp data writer: %w", err)
	}

	if err := setDeadline(conn, timeout); err != nil {
		_ = writer.Close()
		return fmt.Errorf("failed to set smtp deadline: %w", err)
	}
	if _, err := writer.Write(message); err != nil {
		_ = writer.Close()
		return fmt.Errorf("failed to write smtp message: %w", err)
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to close smtp data writer: %w", err)
	}

	if err := setDeadline(conn, timeout); err != nil {
		return fmt.Errorf("failed to set smtp deadline: %w", err)
	}
	if err := client.Quit(); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

func smtpTLSConfig(host string) *tls.Config {
	return &tls.Config{
		ServerName: host,
		MinVersion: tls.VersionTLS12,
	}
}

func setDeadline(conn net.Conn, timeout time.Duration) error {
	if timeout <= 0 {
		return nil
	}
	return conn.SetDeadline(time.Now().Add(timeout))
}

// LoadConfigFromEnv 从环境变量加载配置
func LoadConfigFromEnv() *SMTPConfig {
	return &SMTPConfig{
		Host:     getEnv("SMTP_HOST", "smtp.gmail.com"),
		Port:     getEnvInt("SMTP_PORT", 587),
		Username: getEnv("SMTP_USERNAME", ""),
		Password: getEnv("SMTP_PASSWORD", ""),
		From:     getEnv("SMTP_FROM", "noreply@example.com"),
		FromName: getEnv("SMTP_FROM_NAME", "Store Support"),
		Timeout:  getEnvDuration("SMTP_TIMEOUT", defaultSMTPTimeout),
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

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	duration, err := time.ParseDuration(value)
	if err == nil {
		return duration
	}

	seconds, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return time.Duration(seconds) * time.Second
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

func templateField(data interface{}, names ...string) string {
	if data == nil {
		return ""
	}

	value := reflect.ValueOf(data)
	for value.Kind() == reflect.Ptr || value.Kind() == reflect.Interface {
		if value.IsNil() {
			return ""
		}
		value = value.Elem()
	}

	for _, name := range names {
		if fieldValue, ok := lookupField(value, name); ok {
			return stringifyTemplateValue(fieldValue)
		}
	}

	return ""
}

func lookupField(value reflect.Value, name string) (reflect.Value, bool) {
	switch value.Kind() {
	case reflect.Map:
		if value.Type().Key().Kind() != reflect.String {
			return reflect.Value{}, false
		}
		for _, candidate := range fieldNameCandidates(name) {
			mapValue := value.MapIndex(reflect.ValueOf(candidate).Convert(value.Type().Key()))
			if mapValue.IsValid() {
				return mapValue, true
			}
		}
	case reflect.Struct:
		for _, candidate := range fieldNameCandidates(name) {
			fieldValue := value.FieldByName(candidate)
			if fieldValue.IsValid() {
				return fieldValue, true
			}
		}

		valueType := value.Type()
		for i := 0; i < valueType.NumField(); i++ {
			field := valueType.Field(i)
			jsonName := strings.Split(field.Tag.Get("json"), ",")[0]
			if jsonName == "-" || jsonName == "" {
				continue
			}
			for _, candidate := range fieldNameCandidates(name) {
				if jsonName == candidate {
					return value.Field(i), true
				}
			}
		}
	}

	return reflect.Value{}, false
}

func fieldNameCandidates(name string) []string {
	if name == "" {
		return nil
	}

	candidates := []string{name}
	if strings.Contains(name, "_") {
		parts := strings.Split(name, "_")
		for index, part := range parts {
			if part == "" {
				continue
			}
			parts[index] = strings.ToUpper(part[:1]) + part[1:]
		}
		candidates = append(candidates, strings.Join(parts, ""))
	}

	return candidates
}

func stringifyTemplateValue(value reflect.Value) string {
	if !value.IsValid() {
		return ""
	}

	for value.Kind() == reflect.Ptr || value.Kind() == reflect.Interface {
		if value.IsNil() {
			return ""
		}
		value = value.Elem()
	}

	if !value.IsValid() || value.IsZero() {
		return ""
	}

	if value.CanInterface() {
		switch typedValue := value.Interface().(type) {
		case time.Time:
			return typedValue.Format("2006-01-02 15:04")
		case fmt.Stringer:
			return typedValue.String()
		}
	}

	if !value.CanInterface() {
		return ""
	}

	return fmt.Sprint(value.Interface())
}
