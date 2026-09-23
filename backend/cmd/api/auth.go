package main

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/clerk/clerk-sdk-go/v2"
)

func extractClerkJWT(r *http.Request) string {
	if authorization := strings.TrimSpace(r.Header.Get("Authorization")); authorization != "" {
		return strings.TrimPrefix(authorization, "Bearer ")
	}

	return strings.TrimSpace(r.Header.Get("Sec-WebSocket-Protocol"))
}

func authenticatedClerkID(ctx context.Context) (string, error) {
	claims, ok := clerk.SessionClaimsFromContext(ctx)
	if !ok || claims == nil || claims.Subject == "" {
		return "", errors.New("unauthorized")
	}

	return claims.Subject, nil
}
