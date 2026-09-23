package main

import (
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/nicolasleigh/chat-app/store"
)

const (
	maxPushEndpointLength = 2048
	maxPushKeyLength      = 512
	maxDeviceLabelLength  = 200
)

type pushSubscriptionRequest struct {
	Endpoint string `json:"endpoint"`
	Keys     struct {
		P256dh string `json:"p256dh"`
		Auth   string `json:"auth"`
	} `json:"keys"`
	DeviceLabel *string `json:"device_label"`
}

// createPushSubscription registers the browser subscription for the currently
// authenticated user. The user ID comes from Clerk claims rather than the
// request body, preventing one user from registering a subscription under
// another user's account.
func (app *application) createPushSubscription(w http.ResponseWriter, r *http.Request) {
	clerkID, err := authenticatedClerkID(r.Context())
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := app.query.GetUser(r.Context(), clerkID)
	if err != nil {
		notFoundResponse(w, err)
		return
	}

	var body pushSubscriptionRequest
	if err := readJSON(w, r, &body); err != nil {
		badRequestResponse(w, err)
		return
	}
	if err := validatePushSubscriptionInput(body); err != nil {
		badRequestResponse(w, err)
		return
	}

	result, err := app.query.UpsertPushSubscription(r.Context(), store.UpsertPushSubscriptionParams{
		UserID:      user.ID,
		Endpoint:    strings.TrimSpace(body.Endpoint),
		P256dh:      strings.TrimSpace(body.Keys.P256dh),
		Auth:        strings.TrimSpace(body.Keys.Auth),
		UserAgent:   optionalString(r.UserAgent()),
		DeviceLabel: normalizedOptionalString(body.DeviceLabel),
	})
	if err != nil {
		serverErrorResponse(w, err)
		return
	}

	if err := writeJSON(w, http.StatusOK, result); err != nil {
		serverErrorResponse(w, err)
	}
}

// deletePushSubscription disables a subscription instead of deleting it so a
// support operator can later understand why delivery stopped. The query is
// scoped by the authenticated user's database ID.
func (app *application) deletePushSubscription(w http.ResponseWriter, r *http.Request) {
	clerkID, err := authenticatedClerkID(r.Context())
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	subscriptionID, err := strconv.ParseInt(r.PathValue("subscription_id"), 10, 64)
	if err != nil || subscriptionID <= 0 {
		badRequestResponse(w, errors.New("invalid subscription ID"))
		return
	}

	user, err := app.query.GetUser(r.Context(), clerkID)
	if err != nil {
		notFoundResponse(w, err)
		return
	}

	if err := app.query.DisablePushSubscription(r.Context(), store.DisablePushSubscriptionParams{
		ID:     subscriptionID,
		UserID: user.ID,
	}); err != nil {
		serverErrorResponse(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func validatePushSubscriptionInput(body pushSubscriptionRequest) error {
	endpoint := strings.TrimSpace(body.Endpoint)
	if endpoint == "" || len(endpoint) > maxPushEndpointLength {
		return errors.New("invalid push endpoint")
	}
	parsedEndpoint, err := url.ParseRequestURI(endpoint)
	if err != nil || parsedEndpoint.Scheme != "https" || parsedEndpoint.Host == "" {
		return errors.New("push endpoint must be an HTTPS URL")
	}

	p256dh := strings.TrimSpace(body.Keys.P256dh)
	if p256dh == "" || len(p256dh) > maxPushKeyLength {
		return errors.New("invalid push p256dh key")
	}
	auth := strings.TrimSpace(body.Keys.Auth)
	if auth == "" || len(auth) > maxPushKeyLength {
		return errors.New("invalid push auth key")
	}

	if body.DeviceLabel != nil && len(strings.TrimSpace(*body.DeviceLabel)) > maxDeviceLabelLength {
		return errors.New("device label is too long")
	}

	return nil
}

func optionalString(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return &value
}

func normalizedOptionalString(value *string) *string {
	if value == nil {
		return nil
	}
	return optionalString(*value)
}
