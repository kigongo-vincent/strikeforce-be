package mailer

import (
	"fmt"
	"log"
	"net/http"

	"github.com/mailjet/mailjet-apiv3-go/v4"
)

// InterpretMailjetError logs additional context for Mailjet failures and
// returns a wrapped error with more descriptive messaging.
func InterpretMailjetError(err error, context string) error {
	if err == nil {
		return nil
	}

	if apiErr, ok := err.(mailjet.RequestError); ok {
		log.Printf("mailjet error [%s]: status=%d, info=%s, message=%s", context, apiErr.StatusCode, apiErr.ErrorInfo, apiErr.ErrorMessage)

		if apiErr.StatusCode == http.StatusUnauthorized || apiErr.StatusCode == http.StatusForbidden {
			return fmt.Errorf("email provider rejected the API credentials (status %d)", apiErr.StatusCode)
		}

		return fmt.Errorf("email provider error (status %d): %s", apiErr.StatusCode, apiErr.ErrorMessage)
	}

	log.Printf("mailjet error [%s]: %v", context, err)
	return fmt.Errorf("email provider error: %w", err)
}
