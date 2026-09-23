package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/clerk/clerk-sdk-go/v2"
	clerkhttp "github.com/clerk/clerk-sdk-go/v2/http"
)

func (app *application) NewServer() *http.Server {
	mux := app.NewRouter()
	errLog := slog.NewLogLogger(NewLog.Handler(), slog.LevelError)
	clerk.SetKey(os.Getenv("CLERK_KEY"))

	jwtExtractor := clerkhttp.AuthorizationJWTExtractor(extractClerkJWT)
	// wrap middlewares
	wrappedMux := withRequestID(withRequestLogging(app.enableCORS(clerkhttp.WithHeaderAuthorization(jwtExtractor)(mux))))
	// wrappedMux := app.enableCORS(mux)

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", app.config.port),
		Handler:           wrappedMux,
		ErrorLog:          errLog,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}
	return srv
}
