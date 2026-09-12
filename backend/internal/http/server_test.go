package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"spinwheel/backend/internal/repository"
	"spinwheel/backend/internal/service"
	"spinwheel/backend/pkg/models"
)

func newTestServer(cfg Config) (*Server, *service.WheelService, *models.Wheel) {
	repo := repository.NewInMemoryWheelRepository()
	svc := service.NewWheelService(repo, "test-hmac-secret")
	wheel, err := svc.CreateWheel(&models.Wheel{
		Name: "http-test",
		Items: []models.WheelItem{
			{Option: "A", Color: "red", Weight: 2.0},
			{Option: "B", Color: "blue", Weight: 1.0},
		},
	})
	if err != nil {
		panic(err)
	}
	cfg.Service = svc
	return NewServer(cfg), &svc, wheel
}

func doRequest(t *testing.T, h http.Handler, method, target string, body any, headers map[string]string, cookies []*http.Cookie) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		switch b := body.(type) {
		case string:
			buf.WriteString(b)
		default:
			if err := json.NewEncoder(&buf).Encode(body); err != nil {
				t.Fatal(err)
			}
		}
	}
	req := httptest.NewRequest(method, target, &buf)
	req.RemoteAddr = "203.0.113.7:1234"
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	for _, c := range cookies {
		req.AddCookie(c)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func swCookie(rec *httptest.ResponseRecorder) *http.Cookie {
	for _, c := range rec.Result().Cookies() {
		if c.Name == sessionCookieName {
			return c
		}
	}
	return nil
}

func TestHealthz(t *testing.T) {
	srv, _, _ := newTestServer(Config{})
	rec := doRequest(t, srv.Handler(), http.MethodGet, "/healthz", nil, nil, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("healthz = %d, want 200", rec.Code)
	}
	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil || body["status"] != "ok" {
		t.Fatalf("healthz body = %q", rec.Body.String())
	}
	if rec := doRequest(t, srv.Handler(), http.MethodPost, "/healthz", nil, nil, nil); rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("POST healthz = %d, want 405", rec.Code)
	}
}

func TestSpinRoundTripAndSessionCookie(t *testing.T) {
	srv, svc, wheel := newTestServer(Config{})
	h := srv.Handler()
	hash := service.ComputeItemsHash(wheel.Items)

	rec := doRequest(t, h, http.MethodPost, "/v1/wheels/"+wheel.ID+"/spin",
		map[string]string{"client_seed": "s1", "items_hash": hash},
		map[string]string{"Content-Type": "application/json", "Cf-Ray": "ray-1", "User-Agent": "test/1.0"},
		nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("spin = %d (%s), want 200", rec.Code, rec.Body.String())
	}
	var resp spinResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("spin body: %v", err)
	}
	if !(*svc).VerifySpin(resp.SpinID, wheel.ID, resp.WinnerIndex, hash, resp.Nonce, resp.Sig) {
		t.Fatal("spin sig does not verify")
	}
	if resp.WinnerItem.Weight != wheel.Items[resp.WinnerIndex].Weight {
		t.Fatal("winner weight not mapped")
	}

	// New client gets a sw_sid cookie: HttpOnly, Lax, 1yr.
	c := swCookie(rec)
	if c == nil {
		t.Fatal("no sw_sid cookie set for new client")
	}
	if !c.HttpOnly || c.SameSite != http.SameSiteLaxMode || c.MaxAge != sessionMaxAge {
		t.Fatalf("bad cookie attrs: %+v", c)
	}

	// Returning client keeps its session: no rotation.
	rec2 := doRequest(t, h, http.MethodPost, "/v1/wheels/"+wheel.ID+"/spin",
		map[string]string{"client_seed": "s2", "items_hash": hash},
		map[string]string{"Content-Type": "application/json"},
		[]*http.Cookie{c})
	if rec2.Code != http.StatusOK {
		t.Fatalf("spin 2 = %d, want 200", rec2.Code)
	}
	if swCookie(rec2) != nil {
		t.Fatal("server rotated a valid session cookie")
	}
}

func TestSpinRejects(t *testing.T) {
	srv, _, wheel := newTestServer(Config{})
	h := srv.Handler()
	post := func(body any) *httptest.ResponseRecorder {
		return doRequest(t, h, http.MethodPost, "/v1/wheels/"+wheel.ID+"/spin", body,
			map[string]string{"Content-Type": "application/json"}, nil)
	}

	if rec := post(map[string]string{"items_hash": "stale"}); rec.Code != http.StatusBadRequest ||
		!strings.Contains(rec.Body.String(), "wheel_changed") {
		t.Fatalf("stale hash = %d %q, want 400 wheel_changed", rec.Code, rec.Body.String())
	}
	if rec := post(map[string]string{}); rec.Code != http.StatusBadRequest {
		t.Fatalf("empty hash = %d, want 400", rec.Code)
	}
	if rec := post("not json{{{"); rec.Code != http.StatusBadRequest {
		t.Fatalf("bad json = %d, want 400", rec.Code)
	}
	if rec := doRequest(t, h, http.MethodPost, "/v1/wheels/no-such-wheel/spin",
		map[string]string{"items_hash": "x"}, map[string]string{"Content-Type": "application/json"}, nil); rec.Code != http.StatusNotFound {
		t.Fatalf("missing wheel = %d, want 404", rec.Code)
	}
	if rec := doRequest(t, h, http.MethodGet, "/v1/wheels/"+wheel.ID+"/spin", nil, nil, nil); rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("GET spin = %d, want 405", rec.Code)
	}
	if rec := doRequest(t, h, http.MethodPost, "/v1/wheels/"+wheel.ID+"/history", map[string]string{}, nil, nil); rec.Code != http.StatusNotFound {
		t.Fatalf("unknown subroute = %d, want 404", rec.Code)
	}

	// Oversize body is rejected before parsing.
	big := strings.Repeat("x", maxBodyBytes+100)
	rec := doRequest(t, h, http.MethodPost, "/v1/wheels/"+wheel.ID+"/spin", big,
		map[string]string{"Content-Type": "application/json"}, nil)
	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("oversize = %d, want 413", rec.Code)
	}
}

func TestEvents(t *testing.T) {
	srv, _, wheel := newTestServer(Config{})
	h := srv.Handler()

	rec := doRequest(t, h, http.MethodPost, "/v1/events",
		map[string]string{"type": "PAGEVIEW", "path": "/play/abc"},
		map[string]string{"Content-Type": "application/json"}, nil)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("pageview = %d (%s), want 204", rec.Code, rec.Body.String())
	}
	if rec.Body.Len() != 0 {
		t.Fatal("204 must have empty body")
	}
	if swCookie(rec) == nil {
		t.Fatal("events should also mint the session cookie")
	}

	// Lowercase type is accepted; client_ts must be RFC3339.
	rec = doRequest(t, h, http.MethodPost, "/v1/events",
		map[string]string{"type": "spin_start", "wheel_id": wheel.ID, "client_ts": "2026-09-12T10:00:00Z"},
		map[string]string{"Content-Type": "application/json"}, nil)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("spin_start = %d (%s), want 204", rec.Code, rec.Body.String())
	}

	bad := []map[string]string{
		{"type": "CLICK"},
		{},
		{"type": "PAGEVIEW", "spin_id": "not-a-uuid"},
		{"type": "PAGEVIEW", "client_ts": "yesterday"},
	}
	for i, b := range bad {
		if rec := doRequest(t, h, http.MethodPost, "/v1/events", b,
			map[string]string{"Content-Type": "application/json"}, nil); rec.Code != http.StatusBadRequest {
			t.Fatalf("bad event %d = %d, want 400", i, rec.Code)
		}
	}
}

func TestRateLimit(t *testing.T) {
	srv, _, _ := newTestServer(Config{RequestsPerMin: 1})
	h := srv.Handler()
	if rec := doRequest(t, h, http.MethodGet, "/healthz", nil, nil, nil); rec.Code != http.StatusOK {
		t.Fatalf("first = %d, want 200", rec.Code)
	}
	rec := doRequest(t, h, http.MethodGet, "/healthz", nil, nil, nil)
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("second = %d, want 429", rec.Code)
	}
	if rec.Header().Get("Retry-After") == "" {
		t.Fatal("429 should carry Retry-After")
	}
}

func TestCORS(t *testing.T) {
	srv, _, _ := newTestServer(Config{CORSOrigins: []string{"https://example.com"}})
	h := srv.Handler()

	rec := doRequest(t, h, http.MethodGet, "/healthz", nil,
		map[string]string{"Origin": "https://example.com"}, nil)
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "https://example.com" {
		t.Fatalf("ACAO = %q, want echo", got)
	}
	if got := rec.Header().Get("Access-Control-Allow-Credentials"); got != "true" {
		t.Fatal("credentialed CORS required for the cookie")
	}

	rec = doRequest(t, h, http.MethodGet, "/healthz", nil,
		map[string]string{"Origin": "https://evil.test"}, nil)
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("unlisted origin echoed: %q", got)
	}

	// Preflight short-circuits before routing/limiting.
	rec = doRequest(t, h, http.MethodOptions, "/v1/events", nil,
		map[string]string{"Origin": "https://example.com"}, nil)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("preflight = %d, want 204", rec.Code)
	}
}
