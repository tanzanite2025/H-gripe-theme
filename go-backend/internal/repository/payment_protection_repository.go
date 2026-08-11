package repository

import (
	"encoding/json"
	"time"

	"commerce-platform/internal/domain/audit"
	paymentdomain "commerce-platform/internal/domain/payment"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const paymentProtectionAuditResource = "payment_protection_control"

type PaymentProtectionAuditContext struct {
	UserID    uint
	Username  string
	IPAddress string
	UserAgent string
	Method    string
	Path      string
}

type PaymentProtectionRepository struct {
	db *gorm.DB
}

func NewPaymentProtectionRepository(db *gorm.DB) *PaymentProtectionRepository {
	return &PaymentProtectionRepository{db: db}
}

func (r *PaymentProtectionRepository) WithTx(tx *gorm.DB) *PaymentProtectionRepository {
	return &PaymentProtectionRepository{db: tx}
}

func (r *PaymentProtectionRepository) CreateControlWithAudit(
	control *paymentdomain.PaymentProtectionControl,
	auditContext PaymentProtectionAuditContext,
) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(control).Error; err != nil {
			return err
		}
		auditLog, err := newPaymentProtectionAuditLog(
			auditContext,
			"create",
			control.ID,
			nil,
			control,
			time.Now().UTC(),
		)
		if err != nil {
			return err
		}
		return tx.Create(auditLog).Error
	})
}

func (r *PaymentProtectionRepository) ListControls(
	now time.Time,
	includeExpired bool,
) ([]paymentdomain.PaymentProtectionControl, error) {
	if now.IsZero() {
		now = time.Now().UTC()
	} else {
		now = now.UTC()
	}
	query := r.db.Model(&paymentdomain.PaymentProtectionControl{}).
		Order("enabled DESC, expires_at ASC, id DESC")
	if !includeExpired {
		query = query.Where("enabled = ? AND expires_at > ?", true, now)
	}
	var controls []paymentdomain.PaymentProtectionControl
	if err := query.Find(&controls).Error; err != nil {
		return nil, err
	}
	return controls, nil
}

func (r *PaymentProtectionRepository) FindControlByID(id uint) (*paymentdomain.PaymentProtectionControl, error) {
	var control paymentdomain.PaymentProtectionControl
	if err := r.db.First(&control, id).Error; err != nil {
		return nil, err
	}
	return &control, nil
}

func (r *PaymentProtectionRepository) RevokeControlWithAudit(
	id uint,
	updatedBy uint,
	auditContext PaymentProtectionAuditContext,
) (*paymentdomain.PaymentProtectionControl, error) {
	var revoked *paymentdomain.PaymentProtectionControl
	err := r.db.Transaction(func(tx *gorm.DB) error {
		query := tx.Where("id = ?", id)
		switch tx.Dialector.Name() {
		case "postgres", "mysql", "sqlserver":
			query = query.Clauses(clause.Locking{Strength: "UPDATE"})
		}

		var control paymentdomain.PaymentProtectionControl
		if err := query.First(&control).Error; err != nil {
			return err
		}
		if !control.Enabled {
			revoked = &control
			return nil
		}

		before := control
		now := time.Now().UTC()
		if err := tx.Model(&paymentdomain.PaymentProtectionControl{}).
			Where("id = ? AND enabled = ?", id, true).
			Updates(map[string]interface{}{
				"enabled":    false,
				"updated_by": updatedBy,
				"updated_at": now,
			}).Error; err != nil {
			return err
		}
		control.Enabled = false
		control.UpdatedBy = updatedBy
		control.UpdatedAt = now
		revoked = &control

		auditLog, err := newPaymentProtectionAuditLog(
			auditContext,
			"revoke",
			control.ID,
			&before,
			&control,
			now,
		)
		if err != nil {
			return err
		}
		return tx.Create(auditLog).Error
	})
	return revoked, err
}

func (r *PaymentProtectionRepository) ListControlAuditLogs(
	controlID uint,
	page int,
	pageSize int,
) ([]audit.AuditLog, int64, error) {
	var logs []audit.AuditLog
	var total int64
	query := r.db.Model(&audit.AuditLog{}).
		Where("resource = ? AND resource_id = ?", paymentProtectionAuditResource, controlID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC, id DESC").Offset(offset).Limit(pageSize).Find(&logs).Error; err != nil {
		return nil, 0, err
	}
	return logs, total, nil
}

func (r *PaymentProtectionRepository) FindActiveControlsForEvaluation(
	now time.Time,
) ([]paymentdomain.PaymentProtectionControl, error) {
	return r.ListControls(now, false)
}

func newPaymentProtectionAuditLog(
	context PaymentProtectionAuditContext,
	action string,
	resourceID uint,
	oldValue interface{},
	newValue interface{},
	createdAt time.Time,
) (*audit.AuditLog, error) {
	oldJSON, err := marshalPaymentProtectionAuditValue(oldValue)
	if err != nil {
		return nil, err
	}
	newJSON, err := marshalPaymentProtectionAuditValue(newValue)
	if err != nil {
		return nil, err
	}
	return &audit.AuditLog{
		UserID:     context.UserID,
		Username:   context.Username,
		Action:     action,
		Resource:   paymentProtectionAuditResource,
		ResourceID: resourceID,
		Method:     context.Method,
		Path:       context.Path,
		IPAddress:  context.IPAddress,
		UserAgent:  context.UserAgent,
		OldValue:   oldJSON,
		NewValue:   newJSON,
		Status:     "success",
		CreatedAt:  createdAt,
	}, nil
}

func marshalPaymentProtectionAuditValue(value interface{}) (string, error) {
	if value == nil {
		return "", nil
	}
	payload, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return string(payload), nil
}
