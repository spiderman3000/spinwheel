package models

import "time"

type Wheel struct {
	ID    string      `json:"id"`
	Name  string      `json:"name"`
	Items []WheelItem `json:"items"`
}

type WheelItem struct {
	ID     string  `json:"id"`
	Option string  `json:"option"`
	Color  string  `json:"color"`
	Weight float64 `json:"weight"`
}

// SpinParams carries the inputs for one persisted server spin decision.
type SpinParams struct {
	WheelID    string
	SeedClient string
	ItemsHash  string
	SessionID  string
	IPHash     string
	UA         string
	CFRay      string
	ExpiresAt  time.Time
}

// SpinRecord is the persisted server decision audit log (spins table).
// Sig is filled in by the service layer (HMAC); the repository stores it.
type SpinRecord struct {
	ID           string
	WheelID      string
	WinnerIdx    int
	WinnerItemID string
	SeedServer   []byte
	SeedClient   string
	ItemsHash    string
	Nonce        string
	Sig          string
	SessionID    string
	IPHash       string
	UA           string
	CFRay        string
	ExpiresAt    time.Time
	CreatedAt    time.Time
}

// EventParams carries one analytics event (pageviews table doubles as the
// RecordEvent store: PAGEVIEW rows carry Path, SPIN_START/SPIN_END rows
// carry WheelID/SpinID/Sig).
type EventParams struct {
	SessionID string
	Type      string // PAGEVIEW | SPIN_START | SPIN_END
	WheelID   string // optional
	SpinID    string // optional
	Path      string
	Sig       string
	ClientTS  time.Time
	IPHash    string
	UA        string
	CFRay     string
}

// EventRecord is one persisted analytics event.
type EventRecord struct {
	ID        string
	SessionID string
	Type      string
	WheelID   string
	SpinID    string
	Path      string
	Sig       string
	ClientTS  time.Time
	IPHash    string
	UA        string
	CFRay     string
	CreatedAt time.Time
}
