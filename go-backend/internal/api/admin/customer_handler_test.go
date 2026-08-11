package admin

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"commerce-platform/internal/domain/user"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestCustomerHandlerListCustomersReturnsOnlyStorefrontAccounts(t *testing.T) {
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

	require.NoError(t, db.AutoMigrate(&user.User{}))
	require.NoError(t, db.Create(&user.User{
		Email:    "customer@example.com",
		Username: "customer",
		Password: "password",
		Role:     "user",
		Status:   "active",
	}).Error)
	require.NoError(t, db.Create(&user.User{
		Email:    "support@example.com",
		Username: "support",
		Password: "password",
		Role:     "support",
		Status:   "active",
	}).Error)

	router := gin.New()
	handler := NewCustomerHandler(service.NewUserService(repository.NewUserRepository(db)))
	router.GET("/customers", handler.ListCustomers)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/customers", nil)
	router.ServeHTTP(response, request)

	require.Equal(t, http.StatusOK, response.Code)

	var body struct {
		Customers  []user.UserResponse `json:"customers"`
		Pagination struct {
			Total int `json:"total"`
		} `json:"pagination"`
	}
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
	require.Equal(t, 1, body.Pagination.Total)
	require.Len(t, body.Customers, 1)
	require.Equal(t, "customer@example.com", body.Customers[0].Email)
	require.Equal(t, "user", body.Customers[0].Role)
}
