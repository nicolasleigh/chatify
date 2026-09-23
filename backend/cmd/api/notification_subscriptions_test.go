package main

import "testing"

func validPushSubscriptionRequest() pushSubscriptionRequest {
	request := pushSubscriptionRequest{
		Endpoint:    "https://push.example.test/subscription/abc",
		DeviceLabel: stringPointer("work browser"),
	}
	request.Keys.P256dh = "public-key"
	request.Keys.Auth = "auth-secret"
	return request
}

func TestValidatePushSubscriptionInput(t *testing.T) {
	if err := validatePushSubscriptionInput(validPushSubscriptionRequest()); err != nil {
		t.Fatalf("validatePushSubscriptionInput() error = %v", err)
	}
}

func TestValidatePushSubscriptionInputRejectsNonHTTPSEndpoint(t *testing.T) {
	request := validPushSubscriptionRequest()
	request.Endpoint = "http://push.example.test/subscription/abc"

	if err := validatePushSubscriptionInput(request); err == nil {
		t.Fatal("validatePushSubscriptionInput() error = nil, want HTTPS validation error")
	}
}

func TestValidatePushSubscriptionInputRejectsMissingKeys(t *testing.T) {
	request := validPushSubscriptionRequest()
	request.Keys.Auth = ""

	if err := validatePushSubscriptionInput(request); err == nil {
		t.Fatal("validatePushSubscriptionInput() error = nil, want missing auth key error")
	}
}
