package service

import (
	"commerce-platform/internal/repository"
	"errors"
)

var (
	ErrUnsupportedLocale                    = errors.New("unsupported locale")
	ErrProductOptionRelationInvalid         = errors.New("product option value relation invalid")
	ErrPostLocaleImmutable                  = errors.New("post locale cannot be changed after creation")
	ErrFAQNotFound                          = errors.New("faq not found")
	ErrFAQLocaleImmutable                   = errors.New("FAQ locale cannot be changed after creation")
	ErrGalleryNotFound                      = errors.New("gallery not found")
	ErrPaymentNotFound                      = errors.New("payment resource not found")
	ErrShippingNotFound                     = errors.New("shipping resource not found")
	ErrShippingRateUnavailable              = errors.New("shipping rate is unavailable")
	ErrInvalidShippingDestination           = errors.New("shipping destination is invalid")
	ErrCountryNotSupported                  = errors.New("shipping country is not supported")
	ErrShippingRateConfigurationInvalid     = errors.New("shipping rate configuration is invalid")
	ErrShippingQuoteExpired                 = errors.New("shipping quote has expired")
	ErrShippingQuoteStale                   = errors.New("shipping quote no longer matches the checkout")
	ErrShippingQuotePlanUnavailable         = errors.New("shipping quote plan is unavailable")
	ErrShowcaseUploadOrderRequired          = errors.New("showcase upload order is required")
	ErrShowcaseUploadOrderNotEligible       = errors.New("showcase upload order is not eligible")
	ErrShowcaseUploadEligibilityUnavailable = errors.New("showcase upload eligibility is unavailable")
	ErrShowcaseUploadPendingLimitExceeded   = errors.New("showcase upload pending submission limit exceeded")
)

func IsRecordNotFound(err error) bool {
	return repository.IsRecordNotFound(err)
}
