package main

import (
	"net/http"
	"strings"
)

func extractClerkJWT(r *http.Request) string {
	if authorization := strings.TrimSpace(r.Header.Get("Authorization")); authorization != "" {
		return strings.TrimPrefix(authorization, "Bearer ")
	}

	return strings.TrimSpace(r.Header.Get("Sec-WebSocket-Protocol"))
}
