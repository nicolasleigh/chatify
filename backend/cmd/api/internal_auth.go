package main

import (
	"crypto/subtle"
	"net/http"
)

const internalWebhookSecretHeader = "X-Internal-Webhook-Secret"

func (app *application) requireInternalWebhook(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		providedSecret := []byte(r.Header.Get(internalWebhookSecretHeader))
		expectedSecret := []byte(app.config.internalWebhookSecret)
		if len(expectedSecret) == 0 || len(providedSecret) != len(expectedSecret) ||
			subtle.ConstantTimeCompare(providedSecret, expectedSecret) != 1 {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
