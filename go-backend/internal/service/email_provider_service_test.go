package service

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"commerce-platform/internal/domain/notification"
	emailpkg "commerce-platform/internal/pkg/email"
	"commerce-platform/internal/pkg/secretbox"
	"commerce-platform/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func newEmailProviderServiceForTest(t *testing.T) *EmailProviderService {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&notification.EmailProviderConfig{}))
	t.Cleanup(func() {
		if sqlDB, dbErr := db.DB(); dbErr == nil {
			_ = sqlDB.Close()
		}
	})
	return NewEmailProviderService(repository.NewEmailProviderRepository(db))
}

func validEmailProviderInput(code string) SaveEmailProviderInput {
	return SaveEmailProviderInput{
		Code: code, Name: code, Host: "smtp.example.com", Port: 587,
		Username: "smtp-user", Password: "smtp-secret", FromName: "Store",
		FromEmail: "noreply@example.com", EncryptionType: notification.EmailProviderEncryptionSTARTTLS,
		IsActive: emailProviderBoolPtr(true),
	}
}

func emailProviderBoolPtr(value bool) *bool { return &value }

func TestEmailProviderServiceEncryptsPasswordAndReturnsMaskedView(t *testing.T) {
	t.Setenv(EmailProviderMasterKeyEnv, "email-provider-test-master-key")
	service := newEmailProviderServiceForTest(t)

	view, err := service.Create(validEmailProviderInput("primary"))
	require.NoError(t, err)
	require.NotNil(t, view)
	assert.True(t, view.HasPassword)
	assert.NotContains(t, view.FromEmail, "smtp-secret")

	provider, err := service.repo.FindByID(view.ID)
	require.NoError(t, err)
	assert.NotEqual(t, "smtp-secret", provider.PasswordEncrypted)
	decrypted, err := secretbox.DecryptString(provider.PasswordEncrypted, EmailProviderMasterKey())
	require.NoError(t, err)
	assert.Equal(t, "smtp-secret", decrypted)
}

func TestEmailProviderServiceUpdateWithoutPasswordPreservesExistingSecret(t *testing.T) {
	t.Setenv(EmailProviderMasterKeyEnv, "email-provider-test-master-key")
	service := newEmailProviderServiceForTest(t)
	created, err := service.Create(validEmailProviderInput("primary"))
	require.NoError(t, err)
	before, err := service.repo.FindByID(created.ID)
	require.NoError(t, err)

	input := validEmailProviderInput("primary")
	input.Password = ""
	input.Name = "renamed"
	updated, err := service.Update(created.ID, input)
	require.NoError(t, err)
	assert.Equal(t, "renamed", updated.Name)
	after, err := service.repo.FindByID(created.ID)
	require.NoError(t, err)
	assert.Equal(t, before.PasswordEncrypted, after.PasswordEncrypted)
}

func TestEmailProviderServiceCanReplaceDefaultWithoutUniqueIndexConflict(t *testing.T) {
	t.Setenv(EmailProviderMasterKeyEnv, "email-provider-test-master-key")
	service := newEmailProviderServiceForTest(t)

	first := validEmailProviderInput("first")
	first.IsDefault = emailProviderBoolPtr(true)
	firstView, err := service.Create(first)
	require.NoError(t, err)
	second := validEmailProviderInput("second")
	secondView, err := service.Create(second)
	require.NoError(t, err)

	updated, err := service.SetDefault(secondView.ID)
	require.NoError(t, err)
	assert.True(t, updated.IsDefault)

	old, err := service.Get(firstView.ID)
	require.NoError(t, err)
	assert.False(t, old.IsDefault)
}

func TestEmailProviderServiceRejectsInvalidProviderAsClientError(t *testing.T) {
	service := newEmailProviderServiceForTest(t)
	input := validEmailProviderInput("invalid")
	input.FromEmail = "not-an-email"

	_, err := service.Create(input)
	assert.ErrorIs(t, err, ErrInvalidEmailProvider)
	assert.ErrorIs(t, err, notification.ErrEmailProviderFromEmailInvalid)
}

func TestEmailProviderServiceDefaultSMTPConfigDecryptsPassword(t *testing.T) {
	t.Setenv(EmailProviderMasterKeyEnv, "email-provider-test-master-key")
	service := newEmailProviderServiceForTest(t)
	input := validEmailProviderInput("primary")
	input.IsDefault = emailProviderBoolPtr(true)
	_, err := service.Create(input)
	require.NoError(t, err)

	config, configured, err := service.DefaultSMTPConfig()
	require.NoError(t, err)
	require.True(t, configured)
	require.NotNil(t, config)
	assert.Equal(t, "smtp.example.com", config.Host)
	assert.Equal(t, "smtp-secret", config.Password)
	assert.Equal(t, "noreply@example.com", config.From)
}

func TestRuntimeEmailServiceFallsBackWhenNoDefaultProviderExists(t *testing.T) {
	providers := newEmailProviderServiceForTest(t)
	fallback := &recordingRuntimeEmailSender{}
	runtime := NewRuntimeEmailService(providers, fallback)
	runtime.factory = func(config *emailpkg.SMTPConfig) (emailpkg.EmailService, error) {
		t.Fatalf("database provider factory called for an empty provider registry")
		return nil, nil
	}

	require.NoError(t, runtime.SendEmail([]string{"customer@example.com"}, "subject", "body"))
	assert.Equal(t, 1, fallback.sendCalls)
}

func TestRuntimeEmailServicePrefersDatabaseDefaultProvider(t *testing.T) {
	t.Setenv(EmailProviderMasterKeyEnv, "email-provider-test-master-key")
	providers := newEmailProviderServiceForTest(t)
	input := validEmailProviderInput("primary")
	input.IsDefault = emailProviderBoolPtr(true)
	_, err := providers.Create(input)
	require.NoError(t, err)

	fallback := &recordingRuntimeEmailSender{}
	databaseSender := &recordingRuntimeEmailSender{}
	runtime := NewRuntimeEmailService(providers, fallback)
	var selected *emailpkg.SMTPConfig
	runtime.factory = func(config *emailpkg.SMTPConfig) (emailpkg.EmailService, error) {
		selected = config
		return databaseSender, nil
	}

	require.NoError(t, runtime.SendEmail([]string{"customer@example.com"}, "subject", "body"))
	assert.Equal(t, 0, fallback.sendCalls)
	assert.Equal(t, 1, databaseSender.sendCalls)
	require.NotNil(t, selected)
	assert.Equal(t, "smtp-secret", selected.Password)
}

func TestRuntimeEmailServiceFailsClosedWhenDatabaseSecretCannotDecrypt(t *testing.T) {
	t.Setenv(EmailProviderMasterKeyEnv, "email-provider-test-master-key")
	providers := newEmailProviderServiceForTest(t)
	input := validEmailProviderInput("primary")
	input.IsDefault = emailProviderBoolPtr(true)
	_, err := providers.Create(input)
	require.NoError(t, err)
	t.Setenv(EmailProviderMasterKeyEnv, "different-master-key")

	fallback := &recordingRuntimeEmailSender{}
	runtime := NewRuntimeEmailService(providers, fallback)
	err = runtime.SendEmail([]string{"customer@example.com"}, "subject", "body")
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrEmailProviderPasswordUnavailable)
	assert.Equal(t, 0, fallback.sendCalls)
}

func TestEmailProviderServiceTestPersistsHealthyResultWithoutReturningPassword(t *testing.T) {
	t.Setenv(EmailProviderMasterKeyEnv, "email-provider-test-master-key")
	service := newEmailProviderServiceForTest(t)
	input := validEmailProviderInput("primary")
	view, err := service.Create(input)
	require.NoError(t, err)
	sender := &recordingRuntimeEmailSender{}
	service.factory = func(config *emailpkg.SMTPConfig) (emailpkg.EmailService, error) {
		assert.Equal(t, "smtp-secret", config.Password)
		return sender, nil
	}

	result, err := service.Test(view.ID, "ops@example.com")
	require.NoError(t, err)
	assert.Equal(t, notification.EmailProviderTestStatusHealthy, result.LastTestStatus)
	assert.Empty(t, result.LastTestError)
	assert.Equal(t, 1, sender.sendCalls)
	assert.False(t, strings.Contains(fmt.Sprint(result), "smtp-secret"))
}

func TestEmailProviderServiceTestPersistsFailedResult(t *testing.T) {
	t.Setenv(EmailProviderMasterKeyEnv, "email-provider-test-master-key")
	service := newEmailProviderServiceForTest(t)
	view, err := service.Create(validEmailProviderInput("primary"))
	require.NoError(t, err)
	service.factory = func(*emailpkg.SMTPConfig) (emailpkg.EmailService, error) {
		return nil, errors.New("simulated smtp failure")
	}

	result, err := service.Test(view.ID, "ops@example.com")
	require.ErrorIs(t, err, ErrEmailProviderTestFailed)
	require.NotNil(t, result)
	assert.Equal(t, notification.EmailProviderTestStatusFailed, result.LastTestStatus)
	assert.Contains(t, result.LastTestError, "simulated smtp failure")
}

func TestEmailProviderServiceTestRequiresExplicitTargetEmail(t *testing.T) {
	t.Setenv(EmailProviderMasterKeyEnv, "email-provider-test-master-key")
	service := newEmailProviderServiceForTest(t)
	view, err := service.Create(validEmailProviderInput("primary"))
	require.NoError(t, err)
	_, err = service.Test(view.ID, "")
	assert.ErrorIs(t, err, ErrInvalidEmailProvider)
}

type recordingRuntimeEmailSender struct {
	sendCalls int
}

func (s *recordingRuntimeEmailSender) SendEmail([]string, string, string) error {
	s.sendCalls++
	return nil
}

func (s *recordingRuntimeEmailSender) SendHTMLEmail([]string, string, string, interface{}) error {
	return nil
}

func (s *recordingRuntimeEmailSender) SendPasswordReset(string, interface{}) error { return nil }

func (s *recordingRuntimeEmailSender) SendWelcomeEmail(string, interface{}) error { return nil }
