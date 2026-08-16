package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	"commerce-platform/internal/domain/ops"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

const (
	OpsDeploymentRollbackEnabledEnv         = "OPS_DEPLOY_ROLLBACK_ENABLED"
	OpsDeploymentRollbackSSHHostEnv         = "OPS_DEPLOY_ROLLBACK_SSH_HOST"
	OpsDeploymentRollbackSSHPortEnv         = "OPS_DEPLOY_ROLLBACK_SSH_PORT"
	OpsDeploymentRollbackSSHUserEnv         = "OPS_DEPLOY_ROLLBACK_SSH_USER"
	OpsDeploymentRollbackSSHPrivateKeyEnv   = "OPS_DEPLOY_ROLLBACK_SSH_PRIVATE_KEY_PATH"
	OpsDeploymentRollbackSSHKnownHostsEnv   = "OPS_DEPLOY_ROLLBACK_SSH_KNOWN_HOSTS_PATH"
	OpsDeploymentRollbackSSHWorkdirEnv      = "OPS_DEPLOY_ROLLBACK_SSH_WORKDIR"
	OpsDeploymentRollbackSSHTimeoutEnv      = "OPS_DEPLOY_ROLLBACK_SSH_TIMEOUT_SECONDS"
	opsDeploymentRollbackDefaultPort        = 22
	opsDeploymentRollbackDefaultTimeout     = 15 * time.Minute
	opsDeploymentRollbackDefaultOutputLimit = 24 * 1024
)

var (
	ErrOpsDeploymentRollbackDisabled       = errors.New("operations SSH rollback is disabled")
	ErrOpsDeploymentRollbackInvalidConfig  = errors.New("operations SSH rollback configuration is invalid")
	ErrOpsDeploymentRollbackInvalidTarget  = errors.New("operations SSH rollback target is not bound to the workflow VPS")
	ErrOpsDeploymentRollbackInvalidRef     = errors.New("operations SSH rollback ref must be a full commit SHA")
	ErrOpsDeploymentRollbackUnsupportedEnv = errors.New("operations SSH rollback is limited to production workflows")
)

type OpsDeploymentRollbackExecutor interface {
	ExecuteRollback(context.Context, OpsDeploymentRollbackExecutionInput) (*OpsDeploymentRollbackExecutionResult, error)
}

type OpsDeploymentRollbackExecutionInput struct {
	WorkflowID  uint
	ProjectID   uint
	Environment string
	RollbackRef string
	VPS         *ops.VPSBinding
}

type OpsDeploymentRollbackExecutionResult struct {
	OperationID   string
	Target        string
	OutputSummary string
	StartedAt     time.Time
	CompletedAt   time.Time
}

type OpsDeploymentRollbackSSHConfig struct {
	Enabled          bool
	Host             string
	Port             int
	User             string
	PrivateKeyPath   string
	KnownHostsPath   string
	Workdir          string
	Timeout          time.Duration
	MaxOutputBytes   int
	configurationErr error
}

type OpsDeploymentSSHRollbackExecutor struct {
	config   OpsDeploymentRollbackSSHConfig
	dial     func(context.Context, string, string) (net.Conn, error)
	readFile func(string) ([]byte, error)
	now      func() time.Time
}

func NewOpsDeploymentSSHRollbackExecutor(config OpsDeploymentRollbackSSHConfig) *OpsDeploymentSSHRollbackExecutor {
	return &OpsDeploymentSSHRollbackExecutor{
		config:   config,
		dial:     (&net.Dialer{}).DialContext,
		readFile: os.ReadFile,
		now: func() time.Time {
			return time.Now().UTC()
		},
	}
}

func NewOpsDeploymentSSHRollbackExecutorFromEnv() *OpsDeploymentSSHRollbackExecutor {
	config, err := loadOpsDeploymentRollbackSSHConfigFromEnv()
	config.configurationErr = err
	return NewOpsDeploymentSSHRollbackExecutor(config)
}

func (e *OpsDeploymentSSHRollbackExecutor) ExecuteRollback(
	ctx context.Context,
	input OpsDeploymentRollbackExecutionInput,
) (*OpsDeploymentRollbackExecutionResult, error) {
	if e == nil {
		return nil, errors.New("operations SSH rollback executor is not configured")
	}
	if e.config.configurationErr != nil {
		return nil, fmt.Errorf("%w: %v", ErrOpsDeploymentRollbackInvalidConfig, e.config.configurationErr)
	}
	if !e.config.Enabled {
		return nil, ErrOpsDeploymentRollbackDisabled
	}
	if err := validateOpsDeploymentRollbackInput(e.config, input); err != nil {
		return nil, err
	}
	if ctx == nil {
		ctx = context.Background()
	}
	timeout := e.config.Timeout
	if timeout <= 0 {
		timeout = opsDeploymentRollbackDefaultTimeout
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	hostKeyCallback, err := knownhosts.New(e.config.KnownHostsPath)
	if err != nil {
		return nil, fmt.Errorf("%w: load known hosts: %v", ErrOpsDeploymentRollbackInvalidConfig, err)
	}
	privateKey, err := e.readFile(e.config.PrivateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("%w: read private key: %v", ErrOpsDeploymentRollbackInvalidConfig, err)
	}
	signer, err := ssh.ParsePrivateKey(privateKey)
	if err != nil {
		return nil, fmt.Errorf("%w: parse private key: %v", ErrOpsDeploymentRollbackInvalidConfig, err)
	}

	address := net.JoinHostPort(strings.TrimSpace(e.config.Host), strconv.Itoa(e.config.Port))
	connection, err := e.dial(ctx, "tcp", address)
	if err != nil {
		return nil, fmt.Errorf("dial rollback target %s: %w", address, err)
	}
	sshConfig := &ssh.ClientConfig{
		User:            strings.TrimSpace(e.config.User),
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback: hostKeyCallback,
		Timeout:         timeout,
	}
	clientConnection, channels, requests, err := ssh.NewClientConn(connection, address, sshConfig)
	if err != nil {
		_ = connection.Close()
		return nil, fmt.Errorf("authenticate rollback target %s: %w", address, err)
	}
	client := ssh.NewClient(clientConnection, channels, requests)
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return nil, fmt.Errorf("open rollback SSH session: %w", err)
	}
	defer session.Close()

	outputLimit := e.config.MaxOutputBytes
	if outputLimit <= 0 {
		outputLimit = opsDeploymentRollbackDefaultOutputLimit
	}
	stdout := newOpsDeploymentBoundedOutput(outputLimit)
	stderr := newOpsDeploymentBoundedOutput(outputLimit)
	session.Stdout = stdout
	session.Stderr = stderr

	startedAt := e.now().UTC()
	result := &OpsDeploymentRollbackExecutionResult{
		OperationID: fmt.Sprintf("ssh-rollback-%d-%s", input.WorkflowID, strings.ToLower(strings.TrimSpace(input.RollbackRef))[:12]),
		Target:      address,
		StartedAt:   startedAt,
	}
	command := opsDeploymentRollbackCommand(input.RollbackRef, e.config.Workdir)
	done := make(chan error, 1)
	go func() {
		done <- session.Run(command)
	}()

	select {
	case err = <-done:
	case <-ctx.Done():
		_ = session.Close()
		_ = client.Close()
		select {
		case <-done:
		case <-time.After(time.Second):
		}
		err = ctx.Err()
	}
	result.CompletedAt = e.now().UTC()
	result.OutputSummary = opsDeploymentRollbackOutputSummary(stdout.String(), stderr.String(), outputLimit)
	if err != nil {
		return result, fmt.Errorf("execute rollback command on %s: %w", address, err)
	}
	return result, nil
}

func loadOpsDeploymentRollbackSSHConfigFromEnv() (OpsDeploymentRollbackSSHConfig, error) {
	timeoutSeconds, err := opsDeploymentRollbackPositiveEnv(
		OpsDeploymentRollbackSSHTimeoutEnv,
		int(opsDeploymentRollbackDefaultTimeout/time.Second),
	)
	if err != nil {
		return OpsDeploymentRollbackSSHConfig{}, err
	}
	port, err := opsDeploymentRollbackPositiveEnv(OpsDeploymentRollbackSSHPortEnv, opsDeploymentRollbackDefaultPort)
	if err != nil {
		return OpsDeploymentRollbackSSHConfig{}, err
	}
	return OpsDeploymentRollbackSSHConfig{
		Enabled:        opsDeploymentRollbackBoolEnv(OpsDeploymentRollbackEnabledEnv),
		Host:           strings.TrimSpace(os.Getenv(OpsDeploymentRollbackSSHHostEnv)),
		Port:           port,
		User:           strings.TrimSpace(os.Getenv(OpsDeploymentRollbackSSHUserEnv)),
		PrivateKeyPath: strings.TrimSpace(os.Getenv(OpsDeploymentRollbackSSHPrivateKeyEnv)),
		KnownHostsPath: strings.TrimSpace(os.Getenv(OpsDeploymentRollbackSSHKnownHostsEnv)),
		Workdir:        strings.TrimSpace(os.Getenv(OpsDeploymentRollbackSSHWorkdirEnv)),
		Timeout:        time.Duration(timeoutSeconds) * time.Second,
		MaxOutputBytes: opsDeploymentRollbackDefaultOutputLimit,
	}, nil
}

func validateOpsDeploymentRollbackInput(
	config OpsDeploymentRollbackSSHConfig,
	input OpsDeploymentRollbackExecutionInput,
) error {
	if strings.TrimSpace(input.Environment) != ops.ProjectEnvironmentProduction {
		return ErrOpsDeploymentRollbackUnsupportedEnv
	}
	ref := strings.ToLower(strings.TrimSpace(input.RollbackRef))
	if len(ref) != 40 || !isHexString(ref) {
		return ErrOpsDeploymentRollbackInvalidRef
	}
	if input.VPS == nil || input.VPS.ID == 0 || !opsDeploymentRollbackHostMatches(config.Host, input.VPS) {
		return ErrOpsDeploymentRollbackInvalidTarget
	}
	if config.Port < 1 || config.Port > 65535 {
		return fmt.Errorf("%w: SSH port must be between 1 and 65535", ErrOpsDeploymentRollbackInvalidConfig)
	}
	if !isOpsDeploymentSSHUser(config.User) {
		return fmt.Errorf("%w: SSH user is required and must be a Linux account name", ErrOpsDeploymentRollbackInvalidConfig)
	}
	if !isOpsDeploymentRollbackSafePath(config.PrivateKeyPath) {
		return fmt.Errorf("%w: private key path must be an absolute safe path", ErrOpsDeploymentRollbackInvalidConfig)
	}
	if !isOpsDeploymentRollbackSafePath(config.KnownHostsPath) {
		return fmt.Errorf("%w: known hosts path must be an absolute safe path", ErrOpsDeploymentRollbackInvalidConfig)
	}
	if !isOpsDeploymentRollbackSafePath(config.Workdir) {
		return fmt.Errorf("%w: workdir must be an absolute safe path", ErrOpsDeploymentRollbackInvalidConfig)
	}
	return nil
}

func opsDeploymentRollbackHostMatches(host string, vps *ops.VPSBinding) bool {
	target := normalizeOpsDeploymentRollbackHost(host)
	if target == "" || vps == nil {
		return false
	}
	for _, candidate := range []string{
		vps.Hostname,
		vps.ObservedHostname,
		vps.IPv4,
		vps.ObservedIPv4,
	} {
		if target == normalizeOpsDeploymentRollbackHost(candidate) {
			return true
		}
	}
	return false
}

func normalizeOpsDeploymentRollbackHost(value string) string {
	value = strings.TrimSuffix(strings.ToLower(strings.TrimSpace(value)), ".")
	if value == "" || strings.ContainsAny(value, " \t\r\n/@") {
		return ""
	}
	if ip := net.ParseIP(value); ip != nil {
		return ip.String()
	}
	if strings.Contains(value, ":") {
		return ""
	}
	return value
}

func isOpsDeploymentSSHUser(value string) bool {
	value = strings.TrimSpace(value)
	if len(value) == 0 || len(value) > 32 {
		return false
	}
	for index := 0; index < len(value); index++ {
		character := value[index]
		if index == 0 {
			if (character < 'a' || character > 'z') && character != '_' {
				return false
			}
			continue
		}
		if (character < 'a' || character > 'z') &&
			(character < '0' || character > '9') &&
			character != '_' &&
			character != '-' {
			return false
		}
	}
	return true
}

func isOpsDeploymentRollbackSafePath(value string) bool {
	value = strings.TrimSpace(value)
	if !path.IsAbs(value) || strings.Contains(value, "..") {
		return false
	}
	for index := 0; index < len(value); index++ {
		character := value[index]
		if (character >= 'a' && character <= 'z') ||
			(character >= 'A' && character <= 'Z') ||
			(character >= '0' && character <= '9') ||
			character == '/' ||
			character == '.' ||
			character == '_' ||
			character == '-' {
			continue
		}
		return false
	}
	return true
}

func opsDeploymentRollbackCommand(ref, workdir string) string {
	return fmt.Sprintf(
		"cd -- %s && DEPLOY_REF=%s ./deploy.sh",
		strings.TrimSpace(workdir),
		strings.ToLower(strings.TrimSpace(ref)),
	)
}

func opsDeploymentRollbackOutputSummary(stdout, stderr string, limit int) string {
	parts := make([]string, 0, 2)
	if value := strings.TrimSpace(stdout); value != "" {
		parts = append(parts, "stdout: "+value)
	}
	if value := strings.TrimSpace(stderr); value != "" {
		parts = append(parts, "stderr: "+value)
	}
	if len(parts) == 0 {
		return "SSH rollback command completed without output."
	}
	value := strings.Join(parts, "\n")
	if limit > 0 && len(value) > limit {
		return value[:limit] + "\n[output truncated]"
	}
	return value
}

func opsDeploymentRollbackBoolEnv(key string) bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv(key))) {
	case "1", "true", "yes", "y", "on", "enabled":
		return true
	default:
		return false
	}
}

func opsDeploymentRollbackPositiveEnv(key string, fallback int) (int, error) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return 0, fmt.Errorf("%s must be a positive integer", key)
	}
	return value, nil
}

type opsDeploymentBoundedOutput struct {
	mu        sync.Mutex
	buffer    bytes.Buffer
	limit     int
	truncated bool
}

func newOpsDeploymentBoundedOutput(limit int) *opsDeploymentBoundedOutput {
	return &opsDeploymentBoundedOutput{limit: limit}
}

func (b *opsDeploymentBoundedOutput) Write(value []byte) (int, error) {
	if b == nil {
		return len(value), nil
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.limit <= 0 || b.buffer.Len() >= b.limit {
		b.truncated = b.truncated || len(value) > 0
		return len(value), nil
	}
	remaining := b.limit - b.buffer.Len()
	if len(value) > remaining {
		_, _ = b.buffer.Write(value[:remaining])
		b.truncated = true
		return len(value), nil
	}
	_, _ = b.buffer.Write(value)
	return len(value), nil
}

func (b *opsDeploymentBoundedOutput) String() string {
	if b == nil {
		return ""
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	value := b.buffer.String()
	if b.truncated {
		return value + "\n[stream output truncated]"
	}
	return value
}
