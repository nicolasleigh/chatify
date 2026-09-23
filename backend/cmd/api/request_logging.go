package main

import (
	"log/slog"
	"net/http"
	"time"
)

func withRequestLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startedAt := time.Now()
		next.ServeHTTP(w, r)

		slog.Info("http request completed",
			"method", r.Method,
			"path", r.URL.Path,
			"request_id", requestIDFromContext(r.Context()),
			"remote_addr", r.RemoteAddr,
			"duration_ms", time.Since(startedAt).Milliseconds(),
		)
	})
}
