package main

import (
	"log/slog"
	"net/http"
)

func errorResponse(w http.ResponseWriter, status int, message any) {
	err := writeJSON(w, status, message)
	if err != nil {
		slog.Error("failed to write error response", "error", err)
	}
}

func serverErrorResponse(w http.ResponseWriter, err error) {
	slog.Error("request failed", "error", err)
	message := "the server encountered a problem and could not process your request"
	errorResponse(w, http.StatusInternalServerError, message)
}

func notFoundResponse(w http.ResponseWriter, err error) {
	slog.Warn("resource not found", "error", err)
	errorResponse(w, http.StatusNotFound, "resource not found")
}

func forbiddenResponse(w http.ResponseWriter, err error) {
	slog.Warn("forbidden request", "error", err)
	errorResponse(w, http.StatusForbidden, "forbidden")
}

func badRequestResponse(w http.ResponseWriter, err error) {
	slog.Warn("invalid request", "error", err)
	errorResponse(w, http.StatusBadRequest, "invalid request")
}
