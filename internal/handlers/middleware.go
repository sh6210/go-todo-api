package handlers

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type statusRecorder struct {
	http.ResponseWriter
	status int
}
type contextKey string

const requestIDKey contextKey = "requestID"

func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestId := uuid.New().String()
		ctx := context.WithValue(r.Context(), requestIDKey, requestId)
		w.Header().Set("X-Request-Id", requestId)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func LoggingMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			ww := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

			next.ServeHTTP(ww, r)

			logger.Info("request handler",
				"request_id", GetRequestId(r.Context()),
				"method", r.Method,
				"path", r.URL.Path,
				"host", r.Host,
				"status", ww.status,
				"duration", time.Since(start),
			)
		})
	}
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func GetRequestId(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDKey).(string); ok {
		return id
	}
	return ""
}
