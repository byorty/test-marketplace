package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func Logging(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			
			start := time.Now()

			rw := &responseWriter{
				ResponseWriter: w,
				status: http.StatusOK,
			}

			next.ServeHTTP(rw, r)

			log.Info(
				"http request", 
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", rw.status),
				slog.Duration("latency", time.Since(start)),
				slog.String("remote_addr", r.RemoteAddr),
			)
		})
	}
}