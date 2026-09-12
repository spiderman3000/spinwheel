package middleware

import (
	"context"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

// Limiter is a sliding-window in-memory rate limiter keyed by an opaque
// client key (see ClientKey). One instance is shared by all requests;
// limits are per-key, not global. Not distributed: each Cloud Run
// instance tracks its own windows, so the effective limit scales with
// instance count (fine at min 0 max 3).
type Limiter struct {
	requests map[string][]time.Time
	mu       sync.Mutex
	limit    int
	window   time.Duration
}

// NewLimiter creates a limiter allowing limit requests per window per key.
func NewLimiter(limit int, window time.Duration) *Limiter {
	return &Limiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
}

// Allow reports whether a request for key may proceed, recording it.
func (l *Limiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-l.window)

	// Clean old requests
	requests := l.requests[key]
	validRequests := make([]time.Time, 0, len(requests))
	for _, t := range requests {
		if t.After(windowStart) {
			validRequests = append(validRequests, t)
		}
	}

	if len(validRequests) >= l.limit {
		l.requests[key] = validRequests
		return false
	}

	l.requests[key] = append(validRequests, now)
	return true
}

// ClientKey builds the rate-limit key from the client IP and anon session.
// Both parts matter: IP alone over-blocks NATs/CGNAT, session alone is
// forgeable (a client can mint sessions). Either may be empty.
func ClientKey(ip, session string) string {
	if session == "" {
		session = "-"
	}
	return ip + "|" + session
}

// RateLimitInterceptor enforces the limit keyed off the peer address plus
// the optional user ID — never on the user ID alone, and anonymous
// callers are limited (not rejected). Chain it after UserIDInterceptor so
// the user-ID half of the key is populated.
func RateLimitInterceptor(limit int, window time.Duration) grpc.UnaryServerInterceptor {
	limiter := NewLimiter(limit, window)

	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		key := "unknown"
		if p, ok := peer.FromContext(ctx); ok && p.Addr != nil {
			key = p.Addr.String()
		}
		if userID, ok := UserIDFromContext(ctx); ok {
			key += "|" + userID
		}

		if !limiter.Allow(key) {
			return nil, status.Error(codes.ResourceExhausted, "rate limit exceeded")
		}

		return handler(ctx, req)
	}
}
