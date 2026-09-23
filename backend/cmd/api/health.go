package main

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

type databasePinger interface {
	Ping(context.Context) error
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]string{
		"status": "ok",
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(data)
}

func (app *application) readinessCheckHandler(w http.ResponseWriter, r *http.Request) {
	if app.db == nil {
		http.Error(w, `{"status":"not_ready"}`, http.StatusServiceUnavailable)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	if err := app.db.Ping(ctx); err != nil {
		http.Error(w, `{"status":"not_ready"}`, http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ready"})
}
