package payment

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/plutov/paypal/v4"
)

type payPalAccessTokenRefresher interface {
	GetAccessToken(context.Context) (*paypal.TokenResponse, error)
}

func (g *paypalGatewayImpl) refreshPayPalAccessToken(ctx context.Context) error {
	if g == nil || g.client == nil {
		return fmt.Errorf("paypal client is unavailable")
	}

	refresher, ok := g.client.(payPalAccessTokenRefresher)
	if !ok {
		return fmt.Errorf("paypal client does not support access token refresh")
	}

	g.refreshMu.Lock()
	defer g.refreshMu.Unlock()

	_, err := refresher.GetAccessToken(ctx)
	return err
}

func isPayPalUnauthorizedError(err error) bool {
	var paypalErr *paypal.ErrorResponse
	if !errors.As(err, &paypalErr) || paypalErr == nil || paypalErr.Response == nil {
		return false
	}
	return paypalErr.Response.StatusCode == http.StatusUnauthorized
}

func retryPayPalOperation[T any](g *paypalGatewayImpl, ctx context.Context, call func(context.Context) (T, error)) (T, error) {
	var zero T

	result, err := call(ctx)
	if err == nil || !isPayPalUnauthorizedError(err) {
		return result, err
	}

	if refreshErr := g.refreshPayPalAccessToken(ctx); refreshErr != nil {
		return zero, fmt.Errorf("refresh access token: %w", refreshErr)
	}

	result, err = call(ctx)
	if err != nil {
		return zero, err
	}
	return result, nil
}
