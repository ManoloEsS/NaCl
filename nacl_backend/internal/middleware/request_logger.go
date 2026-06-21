package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

func RequestLogger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rw := newResponseWriter(w)

			next.ServeHTTP(rw, r)

			duration := time.Since(start)

			logger.Log(r.Context(), slog.LevelInfo, "Served request",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", rw.status),
				slog.String("client_ip", r.RemoteAddr),
				slog.Duration("duration", duration),
			)
		})
	}
}

// func determineLogLevel(status int) slog.Level {
// 	if status >= 500 {
// 		return slog.LevelError
// 	}
// 	if status >= 400 {
// 		return slog.LevelWarn
// 	}
// 	return slog.LevelInfo
// }
