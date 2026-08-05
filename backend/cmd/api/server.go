package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/clerk/clerk-sdk-go/v2"
	clerkhttp "github.com/clerk/clerk-sdk-go/v2/http"
)

func (app *application) NewServer() *http.Server {
	mux := app.NewRouter()
	errLog := slog.NewLogLogger(NewLog.Handler(), slog.LevelError)
	clerk.SetKey(os.Getenv("CLERK_KEY"))

	jwtExtractor := clerkhttp.AuthorizationJWTExtractor(func(r *http.Request) string {
		// WebSocket connections send the JWT as the Sec-WebSocket-Protocol value.
		if ws := r.Header.Get("Sec-WebSocket-Protocol"); ws != "" {
			return ws
		}
		// Standard requests use Authorization: Bearer <token>.
		return strings.TrimPrefix(strings.TrimSpace(r.Header.Get("Authorization")), "Bearer ")
	})
	// wrap middlewares
	wrappedMux := app.enableCORS(clerkhttp.WithHeaderAuthorization(jwtExtractor)(mux))
	// wrappedMux := app.enableCORS(mux)

	srv := &http.Server{
		Addr:     fmt.Sprintf(":%d", app.config.port),
		Handler:  wrappedMux,
		ErrorLog: errLog,
	}
	return srv
}
