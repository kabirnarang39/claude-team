package stripeerr

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/stripe/stripe-go/v76"
)

// ToHTTPStatus maps a Stripe error to an appropriate HTTP status code.
// It returns 500 for non-Stripe errors.
func ToHTTPStatus(err error) int {
	if err == nil {
		return http.StatusOK
	}
	var stripeErr *stripe.Error
	if !errors.As(err, &stripeErr) {
		return http.StatusInternalServerError
	}
	switch stripeErr.HTTPStatusCode {
	case http.StatusBadRequest:
		return http.StatusBadRequest
	case http.StatusUnauthorized:
		return http.StatusUnauthorized
	case http.StatusPaymentRequired:
		return http.StatusPaymentRequired
	case http.StatusForbidden:
		return http.StatusForbidden
	case http.StatusNotFound:
		return http.StatusNotFound
	case http.StatusConflict:
		return http.StatusConflict
	case http.StatusTooManyRequests:
		return http.StatusTooManyRequests
	default:
		if stripeErr.HTTPStatusCode >= 400 && stripeErr.HTTPStatusCode < 500 {
			return stripeErr.HTTPStatusCode
		}
		return http.StatusInternalServerError
	}
}

// Scrub returns a safe-to-log error message. It replaces any API key fragments
// and sensitive card data with redacted placeholders.
func Scrub(err error) string {
	if err == nil {
		return ""
	}
	var stripeErr *stripe.Error
	if errors.As(err, &stripeErr) {
		// Return only the Stripe error type and code — never the raw message
		// which could contain user-supplied card data.
		return fmt.Sprintf("stripe error type=%s code=%s status=%d",
			stripeErr.Type, stripeErr.Code, stripeErr.HTTPStatusCode)
	}
	// For non-Stripe errors scrub any obvious key patterns.
	msg := err.Error()
	msg = redactPattern(msg, "sk_test_", 8)
	msg = redactPattern(msg, "sk_live_", 8)
	msg = redactPattern(msg, "rk_test_", 8)
	msg = redactPattern(msg, "rk_live_", 8)
	return msg
}

// redactPattern replaces everything after prefix+keep characters with "***".
func redactPattern(s, prefix string, keep int) string {
	idx := strings.Index(s, prefix)
	if idx < 0 {
		return s
	}
	end := idx + len(prefix) + keep
	if end > len(s) {
		end = len(s)
	}
	return s[:end] + "***"
}
