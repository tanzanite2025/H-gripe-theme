package notification

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
)

func TestEmailTemplateValidateRequiresContentAndKnownCategory(t *testing.T) {
	template := &EmailTemplate{
		Code:            "order_confirmation",
		Locale:          "en",
		Category:        TemplateCategoryOrder,
		Name:            "Order confirmation",
		SubjectTemplate: "Order {{order_number}}",
		BodyText:        "Thanks",
		Version:         1,
	}
	require.NoError(t, template.Validate())

	template.BodyText = ""
	template.BodyHTML = ""
	assert.ErrorIs(t, template.Validate(), ErrTemplateBodyRequired)

	template.BodyText = "Thanks"
	template.Category = "unknown"
	assert.ErrorIs(t, template.Validate(), ErrTemplateCategoryInvalid)
}

func TestEmailTemplateVersionUsesImmutableSnapshotFields(t *testing.T) {
	version := EmailTemplateVersion{
		TemplateID:        7,
		Code:              "order_confirmation",
		Locale:            "en",
		Version:           2,
		Name:              "Order confirmation",
		SubjectTemplate:   "Order",
		BodyText:          "Thanks",
		AllowedVariables:  datatypes.JSON(`[]`),
		RequiredVariables: datatypes.JSON(`[]`),
	}
	assert.Equal(t, uint(7), version.TemplateID)
	assert.Equal(t, 2, version.Version)
}
