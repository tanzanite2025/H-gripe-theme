package admin

import (
	"net/http/httptest"
	"strconv"
	"testing"

	paymentdomain "commerce-platform/internal/domain/payment"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestPaymentTransactionResponsesDoNotExposeGatewayResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open("file:payment-transaction-handler?mode=memory&cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&paymentdomain.Transaction{}))

	const rawGatewayResponse = "{\"billing_details\":{\"name\":\"Private Buyer\",\"address\":{\"line1\":\"12 Secret Road\",\"postal_code\":\"90210\"},\"phone\":\"+1-555-0100\"},\"ip\":\"203.0.113.9\",\"card\":{\"brand\":\"visa\"}}"
	transaction := paymentdomain.Transaction{
		OrderID: 77, TransactionID: "provider-txn-77", PaymentMethod: "stripe",
		AmountMinor: 12500, Currency: "USD", Status: "completed", GatewayResponse: rawGatewayResponse,
	}
	require.NoError(t, db.Create(&transaction).Error)
	handler := NewPaymentHandler(service.NewPaymentService(nil, repository.NewPaymentRepository(db)), nil)

	t.Run("single transaction", func(t *testing.T) {
		context, recorder := paymentTransactionTestContext("id", transaction.ID)
		handler.GetTransaction(context)
		require.Equal(t, 200, recorder.Code)
		assertTransactionResponseOmitsGatewayPII(t, recorder.Body.String())
		require.Contains(t, recorder.Body.String(), "provider-txn-77")
	})

	t.Run("transactions for order", func(t *testing.T) {
		context, recorder := paymentTransactionTestContext("order_id", transaction.OrderID)
		handler.GetOrderTransactions(context)
		require.Equal(t, 200, recorder.Code)
		assertTransactionResponseOmitsGatewayPII(t, recorder.Body.String())
		require.Contains(t, recorder.Body.String(), "provider-txn-77")
	})
}

func paymentTransactionTestContext(param string, value uint) (*gin.Context, *httptest.ResponseRecorder) {
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Params = gin.Params{{Key: param, Value: strconv.FormatUint(uint64(value), 10)}}
	context.Request = httptest.NewRequest("GET", "/", nil)
	return context, recorder
}

func assertTransactionResponseOmitsGatewayPII(t *testing.T, body string) {
	t.Helper()
	for _, forbidden := range []string{"gateway_response", "Private Buyer", "Secret Road", "90210", "+1-555-0100", "203.0.113.9"} {
		require.NotContains(t, body, forbidden)
	}
	require.Contains(t, body, "amount_minor")
	require.Contains(t, body, "payment_method")
	require.Contains(t, body, "completed")
}
