package admin

import (
	"strconv"
	"time"

	paymentdomain "commerce-platform/internal/domain/payment"
	"commerce-platform/internal/pkg/apierror"
	"commerce-platform/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

func (h *PaymentHandler) GetTransaction(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "invalid transaction id")
		return
	}

	transaction, err := h.paymentService.GetTransaction(uint(id))
	if err != nil {
		apierror.RespondNotFound(c, "Transaction")
		return
	}

	response.Success(c, adminPaymentTransactionResponseFromDomain(*transaction))
}

func (h *PaymentHandler) GetOrderTransactions(c *gin.Context) {
	orderID, err := strconv.ParseUint(c.Param("order_id"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "invalid order id")
		return
	}

	transactions, err := h.paymentService.GetOrderTransactions(uint(orderID))
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	items := make([]adminPaymentTransactionResponse, 0, len(transactions))
	for _, transaction := range transactions {
		items = append(items, adminPaymentTransactionResponseFromDomain(transaction))
	}
	response.Success(c, gin.H{"data": items})
}

type adminPaymentTransactionResponse struct {
	ID               uint       `json:"id"`
	OrderID          uint       `json:"order_id"`
	TransactionID    string     `json:"transaction_id"`
	AttemptKey       string     `json:"attempt_key,omitempty"`
	PaymentMethod    string     `json:"payment_method"`
	AmountMinor      int64      `json:"amount_minor"`
	Currency         string     `json:"currency"`
	Status           string     `json:"status"`
	LiabilityShifted *bool      `json:"liability_shifted,omitempty"`
	ErrorMessage     string     `json:"error_message"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	CompletedAt      *time.Time `json:"completed_at"`
}

func adminPaymentTransactionResponseFromDomain(transaction paymentdomain.Transaction) adminPaymentTransactionResponse {
	return adminPaymentTransactionResponse{
		ID:               transaction.ID,
		OrderID:          transaction.OrderID,
		TransactionID:    transaction.TransactionID,
		AttemptKey:       transaction.AttemptKey,
		PaymentMethod:    transaction.PaymentMethod,
		AmountMinor:      transaction.AmountMinor,
		Currency:         transaction.Currency,
		Status:           transaction.Status,
		LiabilityShifted: transaction.LiabilityShifted,
		ErrorMessage:     transaction.ErrorMessage,
		CreatedAt:        transaction.CreatedAt,
		UpdatedAt:        transaction.UpdatedAt,
		CompletedAt:      transaction.CompletedAt,
	}
}
