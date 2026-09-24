package repository

import (
	"encoding/json"
	"strings"
	"time"

	"commerce-platform/internal/domain/audit"
	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/domain/ticket"
	"gorm.io/gorm"
)

const customerServiceTicketCategory = "customer_service"

// CustomerServiceAttachmentReferenceScan keeps the cleanup references and
// the number of non-empty legacy payloads that could not be parsed. The
// reference list remains best-effort so one malformed message cannot block an
// otherwise eligible retention purge.
type CustomerServiceAttachmentReferenceScan struct {
	References        []string
	MalformedPayloads int
}

// FindCustomerServiceRetentionTickets reads the explicit administrator target
// set, including soft-deleted rows for purge eligibility checks.
func (r *TicketRepository) FindCustomerServiceRetentionTickets(ids []uint) ([]ticket.Ticket, error) {
	var records []ticket.Ticket
	if len(ids) == 0 {
		return records, nil
	}
	if r == nil || r.db == nil {
		return nil, gorm.ErrInvalidDB
	}
	err := r.db.Unscoped().Where("id IN ? AND category = ?", ids, customerServiceTicketCategory).
		Order("id ASC").Find(&records).Error
	return records, err
}

// FindCustomerServiceConversationAttachmentReferenceScan returns the raw
// attachment references owned by a conversation and counts malformed,
// non-empty historical payloads for operational reporting.
func (r *TicketRepository) FindCustomerServiceConversationAttachmentReferenceScan(id uint) (CustomerServiceAttachmentReferenceScan, error) {
	scan := CustomerServiceAttachmentReferenceScan{References: []string{}}
	if r == nil || r.db == nil {
		return scan, gorm.ErrInvalidDB
	}
	if id == 0 {
		return scan, nil
	}
	var payloads []string
	if err := r.db.Unscoped().Model(&ticket.TicketMessage{}).
		Where("ticket_id = ?", id).
		Pluck("attachments", &payloads).Error; err != nil {
		return scan, err
	}
	seen := make(map[string]struct{})
	for _, payload := range payloads {
		if strings.TrimSpace(payload) == "" {
			continue
		}
		var values []string
		if err := json.Unmarshal([]byte(payload), &values); err != nil {
			scan.MalformedPayloads++
			continue
		}
		for _, value := range values {
			value = strings.TrimSpace(value)
			if value == "" {
				continue
			}
			if _, ok := seen[value]; ok {
				continue
			}
			seen[value] = struct{}{}
			scan.References = append(scan.References, value)
		}
	}
	return scan, nil
}

// FindCustomerServiceRetentionCandidates returns only conversations whose
// administrator soft-delete recovery window has elapsed. The caller still
// performs dependency/hold checks before purging each row.
func (r *TicketRepository) FindCustomerServiceRetentionCandidates(purgeBefore time.Time, limit int) ([]ticket.Ticket, error) {
	if r == nil || r.db == nil {
		return nil, gorm.ErrInvalidDB
	}
	if purgeBefore.IsZero() || limit <= 0 {
		return []ticket.Ticket{}, nil
	}
	var records []ticket.Ticket
	err := r.db.Unscoped().Where(
		"category = ? AND deleted_at IS NOT NULL AND deleted_at <= ?",
		customerServiceTicketCategory,
		purgeBefore.UTC(),
	).Order("deleted_at ASC").Order("id ASC").Limit(limit).Find(&records).Error
	return records, err
}

func (r *TicketRepository) SoftDeleteCustomerServiceConversation(id uint, deletedAt time.Time) error {
	if r == nil || r.db == nil {
		return gorm.ErrInvalidDB
	}
	if id == 0 {
		return gorm.ErrInvalidData
	}
	if deletedAt.IsZero() {
		deletedAt = time.Now().UTC()
	}
	result := r.db.Model(&ticket.Ticket{}).Where("id = ? AND category = ? AND deleted_at IS NULL", id, customerServiceTicketCategory).
		Update("deleted_at", deletedAt.UTC())
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// PurgeCustomerServiceConversation removes only the ticket-owned database
// records after the retention service has completed all dependency and
// recovery-window checks. Object references are cleaned by the service before
// this method is called. It deliberately does not cascade into orders,
// visitors, or audit.
func (r *TicketRepository) PurgeCustomerServiceConversation(id uint) error {
	return r.purgeCustomerServiceConversation(id, nil)
}

// PurgeCustomerServiceConversationWithAudit removes ticket-owned rows and
// persists the corresponding retention audit record in the same transaction.
// External attachment/search cleanup is intentionally performed by the service
// before this method; those integrations must remain idempotent because they
// cannot participate in the SQL transaction.
func (r *TicketRepository) PurgeCustomerServiceConversationWithAudit(id uint, entry *audit.AuditLog) error {
	if entry == nil {
		return gorm.ErrInvalidData
	}
	return r.purgeCustomerServiceConversation(id, entry)
}

// PurgeCustomerServiceConversationWithAuditAndCleanup commits the ticket-owned
// delete, success audit, and post-commit external-cleanup event together.
func (r *TicketRepository) PurgeCustomerServiceConversationWithAuditAndCleanup(
	id uint,
	entry *audit.AuditLog,
	cleanupEvent *outbox.Event,
) error {
	if entry == nil || cleanupEvent == nil {
		return gorm.ErrInvalidData
	}
	return r.purgeCustomerServiceConversationWithOutbox(id, entry, cleanupEvent)
}

func (r *TicketRepository) purgeCustomerServiceConversation(id uint, entry *audit.AuditLog) error {
	return r.purgeCustomerServiceConversationWithOutbox(id, entry, nil)
}

func (r *TicketRepository) purgeCustomerServiceConversationWithOutbox(
	id uint,
	entry *audit.AuditLog,
	cleanupEvent *outbox.Event,
) error {
	if r == nil || r.db == nil {
		return gorm.ErrInvalidDB
	}
	if id == 0 {
		return gorm.ErrInvalidData
	}
	return r.db.Transaction(func(tx *gorm.DB) error {
		var target ticket.Ticket
		if err := tx.Unscoped().Where("id = ? AND category = ?", id, customerServiceTicketCategory).First(&target).Error; err != nil {
			return err
		}
		if !target.DeletedAt.Valid {
			return gorm.ErrRecordNotFound
		}
		if err := tx.Unscoped().Where("ticket_id = ?", id).Delete(&ticket.TicketMessage{}).Error; err != nil {
			return err
		}
		if err := tx.Unscoped().Where("ticket_id = ?", id).Delete(&ticket.CustomerServiceInboxState{}).Error; err != nil {
			return err
		}
		result := tx.Unscoped().Where("id = ? AND category = ? AND deleted_at IS NOT NULL", id, customerServiceTicketCategory).Delete(&ticket.Ticket{})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		if entry != nil {
			if err := tx.Create(entry).Error; err != nil {
				return err
			}
		}
		if cleanupEvent != nil {
			if err := tx.Create(cleanupEvent).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
