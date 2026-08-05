package service

import (
	"testing"

	"tanzanite/internal/domain/faq"
	"tanzanite/internal/domain/ticket"
	"tanzanite/internal/domain/user"
	"tanzanite/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestAutoReplyWelcomeUsesConversationStaffIdentityAndCooldown(t *testing.T) {
	db := newAutoReplyTestDB(t)
	agent := createAutoReplyTestUser(t, db, "auto-agent@example.test", "auto-agent")
	customer := createAutoReplyTestUser(t, db, "auto-customer@example.test", "auto-customer")
	customer.Role = "user"
	require.NoError(t, db.Save(&customer).Error)
	conversationID := "auto-reply-welcome"
	conversation := ticket.Ticket{
		TicketNumber:       "TK-AUTO-REPLY-WELCOME",
		UserID:             customer.ID,
		CustomerUserID:     &customer.ID,
		ConversationID:     &conversationID,
		VisitorSessionHash: "auto-reply-visitor",
		Category:           customerServiceTicketCategory,
		Status:             "open",
		AssignedTo:         agent.ID,
		Subject:            "Customer service chat",
	}
	require.NoError(t, db.Create(&conversation).Error)
	require.NoError(t, db.Create(&ticket.AutoReplyRule{
		Type:            "welcome",
		ReplyMessage:    "Welcome to Tanzanite.",
		Locale:          "en",
		MessageType:     "text",
		IsActive:        true,
		CooldownSeconds: 86400,
	}).Error)

	service := NewTicketService(repository.NewTicketRepository(db), repository.NewUserRepository(db))
	owner := CustomerServiceOwner{VisitorSessionHash: "auto-reply-visitor"}

	_, alreadySent, first, err := service.GetWelcomeMessage(conversationID, owner, agent.ID, "en")
	require.NoError(t, err)
	require.False(t, alreadySent)
	require.NotNil(t, first)
	require.Equal(t, agent.ID, first.UserID)
	require.Equal(t, "text", first.MessageType)
	require.Contains(t, first.Metadata, `"auto_reply"`)

	_, alreadySent, second, err := service.GetWelcomeMessage(conversationID, owner, agent.ID, "en")
	require.NoError(t, err)
	require.True(t, alreadySent)
	require.Nil(t, second)

	var count int64
	require.NoError(t, db.Model(&ticket.TicketMessage{}).Where("ticket_id = ?", conversation.ID).Count(&count).Error)
	require.EqualValues(t, 1, count)
}

func TestAutoReplyRuleValidationRejectsUnsafeLink(t *testing.T) {
	_, err := normalizeAutoReplyRuleInput(AutoReplyRuleInput{
		Type:           "keyword",
		TriggerKeyword: "shipping",
		ReplyMessage:   "Open the guide",
		Locale:         "en",
		MessageType:    "link",
		Metadata:       `{"url":"javascript:alert(1)","title":"Guide"}`,
		IsActive:       true,
		MatchType:      "contains",
	}, nil)
	require.ErrorIs(t, err, ErrInvalidAutoReplyRule)

	_, err = normalizeAutoReplyRuleInput(AutoReplyRuleInput{
		Type:           "keyword",
		TriggerKeyword: "shipping",
		ReplyMessage:   "Open the guide",
		Locale:         "en",
		MessageType:    "link",
		Metadata:       `{"url":"//evil.example/redirect","title":"Guide"}`,
		IsActive:       true,
		MatchType:      "contains",
	}, nil)
	require.ErrorIs(t, err, ErrInvalidAutoReplyRule)
}

func TestAutoReplyRuleValidationAllowsAppRelativeLink(t *testing.T) {
	rule, err := normalizeAutoReplyRuleInput(AutoReplyRuleInput{
		Type:           "keyword",
		TriggerKeyword: "faq",
		ReplyMessage:   "Open the FAQ",
		Locale:         "en",
		MessageType:    "link",
		Metadata:       `{"url":"/support/faqs","title":"FAQ"}`,
		IsActive:       true,
		MatchType:      "contains",
	}, nil)
	require.NoError(t, err)
	require.Equal(t, "link", rule.MessageType)
}

func TestAutoReplyRuleValidationAllowsStructuredFAQ(t *testing.T) {
	rule, err := normalizeAutoReplyRuleInput(AutoReplyRuleInput{
		Type:           "keyword",
		TriggerKeyword: "payment",
		ReplyMessage:   "This FAQ should help.",
		Locale:         "en",
		MessageType:    "faq",
		Metadata:       `{"faq_id":123,"page_id":"payment-security","category":"payments","question":"Is my payment secure?","answer_excerpt":"Payments are securely processed.","url":"/support/faqs?page=payment-security&faq=123","answer_image_url":"/uploads/faq/payment.webp"}`,
		IsActive:       true,
		MatchType:      "contains",
	}, nil)
	require.NoError(t, err)
	require.Equal(t, "faq", rule.MessageType)
}

func TestAutoReplyRuleValidationRejectsInvalidFAQMetadata(t *testing.T) {
	_, err := normalizeAutoReplyRuleInput(AutoReplyRuleInput{
		Type:           "keyword",
		TriggerKeyword: "payment",
		ReplyMessage:   "This FAQ should help.",
		Locale:         "en",
		MessageType:    "faq",
		Metadata:       `{"faq_id":123,"question":"Is my payment secure?","url":"javascript:alert(1)"}`,
		IsActive:       true,
		MatchType:      "contains",
	}, nil)
	require.ErrorIs(t, err, ErrInvalidAutoReplyRule)

	_, err = normalizeAutoReplyRuleInput(AutoReplyRuleInput{
		Type:           "keyword",
		TriggerKeyword: "payment",
		ReplyMessage:   "This FAQ should help.",
		Locale:         "en",
		MessageType:    "faq",
		Metadata:       `{"page_id":"payment-security","question":"Is my payment secure?","url":"/support/faqs?page=payment-security&faq=123"}`,
		IsActive:       true,
		MatchType:      "contains",
	}, nil)
	require.ErrorIs(t, err, ErrInvalidAutoReplyRule)
}

func TestAutoReplyRuleValidationRequiresStructuredContent(t *testing.T) {
	_, err := normalizeAutoReplyRuleInput(AutoReplyRuleInput{
		Type:         "welcome",
		ReplyMessage: "See the wheel image",
		MessageType:  "image",
		IsActive:     true,
	}, nil)
	require.ErrorIs(t, err, ErrInvalidAutoReplyRule)

	_, err = normalizeAutoReplyRuleInput(AutoReplyRuleInput{
		Type:         "welcome",
		ReplyMessage: "Open the product",
		MessageType:  "product",
		IsActive:     true,
		Metadata:     `{"url":"javascript:alert(1)"}`,
	}, nil)
	require.ErrorIs(t, err, ErrInvalidAutoReplyRule)
}

func TestAutoReplyCooldownUsesRuleIdentityAcrossLocales(t *testing.T) {
	db := newAutoReplyTestDB(t)
	agent := createAutoReplyTestUser(t, db, "auto-rule-agent@example.test", "auto-rule-agent")
	conversationID := "auto-reply-locale-cooldown"
	conversation := ticket.Ticket{
		TicketNumber:       "TK-AUTO-REPLY-LOCALE",
		UserID:             agent.ID,
		ConversationID:     &conversationID,
		VisitorSessionHash: "auto-reply-locale-visitor",
		Category:           customerServiceTicketCategory,
		Status:             "open",
		AssignedTo:         agent.ID,
		Subject:            "Customer service chat",
	}
	require.NoError(t, db.Create(&conversation).Error)
	require.NoError(t, db.Create(&ticket.AutoReplyRule{
		Type:            "welcome",
		ReplyMessage:    "Welcome.",
		Locale:          "en",
		MessageType:     "text",
		IsActive:        true,
		CooldownSeconds: 86400,
	}).Error)

	service := NewTicketService(repository.NewTicketRepository(db), repository.NewUserRepository(db))
	owner := CustomerServiceOwner{VisitorSessionHash: "auto-reply-locale-visitor"}

	_, alreadySent, first, err := service.GetWelcomeMessage(conversationID, owner, agent.ID, "en-US")
	require.NoError(t, err)
	require.False(t, alreadySent)
	require.NotNil(t, first)
	require.Contains(t, first.Metadata, `"_dedupe_key":"autoreply:rule:`)

	_, alreadySent, second, err := service.GetWelcomeMessage(conversationID, owner, agent.ID, "en-GB")
	require.NoError(t, err)
	require.True(t, alreadySent)
	require.Nil(t, second)

	var count int64
	require.NoError(t, db.Model(&ticket.TicketMessage{}).Where("ticket_id = ?", conversation.ID).Count(&count).Error)
	require.EqualValues(t, 1, count)
}

func TestAutoReplyRuleSelectionPrefersLocaleAndAgentScope(t *testing.T) {
	rules := []ticket.AutoReplyRule{
		{ID: 1, Type: "keyword", Priority: 20, Locale: "en", AgentID: ""},
		{ID: 2, Type: "keyword", Priority: 1, Locale: "en", AgentID: ""},
		{ID: 3, Type: "keyword", Priority: 1, Locale: "en", AgentID: "7"},
	}

	selected := selectAutoReplyRule(rules, "en-US", 7)
	require.NotNil(t, selected)
	require.Equal(t, uint(3), selected.ID)
}

func TestAutoReplyDoesNotUseWildcardOrOtherLocaleFallback(t *testing.T) {
	db := newAutoReplyTestDB(t)
	agent := createAutoReplyTestUser(t, db, "auto-locale-agent@example.test", "auto-locale-agent")
	conversationID := "auto-reply-strict-locale"
	conversation := ticket.Ticket{
		TicketNumber:       "TK-AUTO-REPLY-STRICT-LOCALE",
		UserID:             agent.ID,
		ConversationID:     &conversationID,
		VisitorSessionHash: "auto-reply-strict-locale-visitor",
		Category:           customerServiceTicketCategory,
		Status:             "open",
		AssignedTo:         agent.ID,
		Subject:            "Customer service chat",
	}
	require.NoError(t, db.Create(&conversation).Error)
	require.NoError(t, db.Create(&ticket.AutoReplyRule{
		Type:            "welcome",
		ReplyMessage:    "English only.",
		Locale:          "en",
		MessageType:     "text",
		IsActive:        true,
		CooldownSeconds: 86400,
	}).Error)

	service := NewTicketService(repository.NewTicketRepository(db), repository.NewUserRepository(db))
	owner := CustomerServiceOwner{VisitorSessionHash: "auto-reply-strict-locale-visitor"}

	_, alreadySent, message, err := service.GetWelcomeMessage(conversationID, owner, agent.ID, "zh-CN")
	require.NoError(t, err)
	require.False(t, alreadySent)
	require.Nil(t, message)

	var count int64
	require.NoError(t, db.Model(&ticket.TicketMessage{}).Where("ticket_id = ?", conversation.ID).Count(&count).Error)
	require.Zero(t, count)
}

func TestAutoReplyRuleValidationRequiresSupportedLocale(t *testing.T) {
	for _, locale := range []string{"", "*", "zz", "en-US"} {
		_, err := normalizeAutoReplyRuleInput(AutoReplyRuleInput{
			Type:         "welcome",
			ReplyMessage: "Hello.",
			Locale:       locale,
			MessageType:  "text",
			IsActive:     true,
		}, nil)
		if locale == "en-US" {
			require.NoError(t, err)
			continue
		}
		require.ErrorIs(t, err, ErrInvalidAutoReplyRule)
	}
}

func TestAutoReplyFAQReferenceRequiresPublishedSameLocale(t *testing.T) {
	db := newAutoReplyTestDB(t)
	require.NoError(t, db.AutoMigrate(&faq.FAQ{}))

	item := faq.FAQ{
		Question: "Are payments secure?",
		Answer:   "Yes.",
		PageID:   "support-payment",
		Category: "payment-security",
		Locale:   "en",
		Status:   "published",
	}
	require.NoError(t, db.Create(&item).Error)

	service := NewTicketService(
		repository.NewTicketRepository(db),
		repository.NewUserRepository(db),
		repository.NewFAQRepository(db),
	)
	input := AutoReplyRuleInput{
		Type:           "keyword",
		TriggerKeyword: "payment",
		ReplyMessage:   "See the payment FAQ.",
		Locale:         "en-US",
		MessageType:    "faq",
		Metadata:       `{"faq_id":1,"page_id":"support-payment","category":"payment-security","locale":"en","question":"Are payments secure?","url":"/support/faqs?page=support-payment&faq=1"}`,
		IsActive:       true,
		MatchType:      "contains",
	}

	rule, err := service.CreateAutoReplyRule(input)
	require.NoError(t, err)
	require.Equal(t, "en", rule.Locale)

	item.Status = "draft"
	require.NoError(t, db.Save(&item).Error)
	_, err = service.CreateAutoReplyRule(input)
	require.ErrorIs(t, err, ErrInvalidAutoReplyRule)

	item.Status = "published"
	item.Locale = "zh_cn"
	require.NoError(t, db.Save(&item).Error)
	_, err = service.CreateAutoReplyRule(input)
	require.ErrorIs(t, err, ErrInvalidAutoReplyRule)
}

func TestAutoReplyFAQReferenceFailsClosedWithoutRepository(t *testing.T) {
	service := NewTicketService(nil, nil)
	rule := ticket.AutoReplyRule{
		MessageType: "faq",
		Locale:      "en",
		Metadata:    `{"faq_id":1,"question":"Are payments secure?","url":"/support/faqs?faq=1"}`,
	}

	require.ErrorIs(t, service.validateAutoReplyFAQReference(rule), ErrInvalidAutoReplyRule)
}

func newAutoReplyTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	require.NoError(t, db.AutoMigrate(
		&user.User{},
		&user.AgentGroup{},
		&user.AgentProfile{},
		&user.AgentGroupMember{},
		&ticket.Ticket{},
		&ticket.TicketMessage{},
		&ticket.AutoReplyRule{},
	))
	return db
}

func createAutoReplyTestUser(t *testing.T, db *gorm.DB, email, username string) user.User {
	t.Helper()

	item := user.User{
		Email:    email,
		Username: username,
		Password: "test-password",
		Role:     "support",
		Status:   "active",
	}
	require.NoError(t, db.Create(&item).Error)
	return item
}
