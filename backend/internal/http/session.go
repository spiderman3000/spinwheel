package http

import (
	"crypto/sha256"
	"encoding/hex"
	"net"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

// sessionCookieName is the anon identity cookie (contract v1:
// HttpOnly, SameSite=Lax, 1yr, UUIDv7).
const sessionCookieName = "sw_sid"

// sessionMaxAge is one year in seconds.
const sessionMaxAge = 365 * 24 * 60 * 60

func validSessionID(s string) bool {
	_, err := uuid.Parse(s)
	return err == nil
}

// sessionFromCookie peeks at the request cookie without setting anything.
// Empty when absent or malformed (malformed is treated as absent so a
// corrupted cookie can never wedge a client).
func sessionFromCookie(r *http.Request) string {
	c, err := r.Cookie(sessionCookieName)
	if err != nil || !validSessionID(c.Value) {
		return ""
	}
	return c.Value
}

// ensureSessionID resolves the anon session: a valid cookie wins, then an
// explicit client echo (lets the FE keep continuity when the cookie is
// unavailable, e.g. cross-site fetch where Lax is not sent), else a fresh
// UUIDv7. When there is no valid cookie one is set.
func (s *Server) ensureSessionID(w http.ResponseWriter, r *http.Request, explicit string) string {
	if id := sessionFromCookie(r); id != "" {
		return id
	}
	if validSessionID(explicit) {
		s.setSessionCookie(w, explicit)
		return explicit
	}
	id, err := uuid.NewV7()
	if err != nil {
		id = uuid.New()
	}
	s.setSessionCookie(w, id.String())
	return id.String()
}

func (s *Server) setSessionCookie(w http.ResponseWriter, id string) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    id,
		Path:     "/",
		MaxAge:   sessionMaxAge,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   s.secureCookies,
	})
}

// clientIP resolves the client address, trusting the CDN headers Cloud Run
// sits behind: CF-Connecting-IP first, then the leftmost X-Forwarded-For,
// then the connection's remote address.
func clientIP(r *http.Request) string {
	if ip := strings.TrimSpace(r.Header.Get("CF-Connecting-IP")); ip != "" {
		return ip
	}
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if i := strings.Index(xff, ","); i >= 0 {
			xff = xff[:i]
		}
		if ip := strings.TrimSpace(xff); ip != "" {
			return ip
		}
	}
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return host
	}
	return r.RemoteAddr
}

// ipHash pseudonymizes the IP for the analytics store (raw IPs are never
// persisted — only this hash lands in spins/pageviews).
func ipHash(ip string) string {
	sum := sha256.Sum256([]byte(ip))
	return hex.EncodeToString(sum[:])
}

// cfRay returns the Cloudflare request ID for log correlation.
func cfRay(r *http.Request) string {
	return r.Header.Get("Cf-Ray")
}
