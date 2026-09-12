package http

import (
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"

	"spinwheel/backend/internal/middleware"
)

// statusRecorder captures the status code for access logs.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// withCORS enforces the origin allowlist. Allowed origins get an explicit
// echo + credentials (the sw_sid cookie is credentialed); anything else
// gets no CORS headers so browsers block the read. Preflights short-circuit
// here with 204, before rate limiting and logging.
func (s *Server) withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" {
			if _, ok := s.cors[strings.ToLower(strings.TrimSpace(origin))]; ok {
				w.Header().Set("Access-Control-Allow-Origin", strings.TrimSpace(origin))
				w.Header().Set("Vary", "Origin")
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
				w.Header().Set("Access-Control-Max-Age", "86400")
			}
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// withRateLimit enforces RequestsPerMin per IP+session key. The session is
// peeked, never set, here — cookie issuance stays in the handlers so
// limited responses never mint sessions.
func (s *Server) withRateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := middleware.ClientKey(clientIP(r), sessionFromCookie(r))
		if !s.limiter.Allow(key) {
			w.Header().Set("Retry-After", "60")
			writeError(w, http.StatusTooManyRequests, "rate limit exceeded")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// withLogging emits one access line per request. Only the IP hash is
// logged — raw IPs never reach the logs, matching the analytics store.
// The spin handler adds its own line with spin_id/cf_ray for correlation.
func (s *Server) withLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		s.log.Info("http request",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.Int("status", rec.status),
			zap.Duration("latency", time.Since(start)),
			zap.String("ip_hash", ipHash(clientIP(r))),
			zap.String("ua", r.UserAgent()),
			zap.String("cf_ray", cfRay(r)),
		)
	})
}
