// Package license is Windeck Pro: a 14-day trial from the first launch,
// then a license key, bought once and activated on up to three Macs, which
// Lemon Squeezy's license API checks and frees for another Mac.
//
// The free version keeps everything the switcher needs; Pro adds what
// learns from the user (the predicted window, search ranked by habit) and
// what is tuned to taste.
package license

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// TrialDays is how long Pro is free to try.
const TrialDays = 14

// State is what is kept of the trial and the license.
type State struct {
	TrialStart time.Time `json:"trialStart"`
	Key        string    `json:"key,omitempty"`
	Instance   string    `json:"instance,omitempty"` // Lemon Squeezy's id of this Mac's activation
	Email      string    `json:"email,omitempty"`
	Activated  time.Time `json:"activated,omitzero"`
	Checked    time.Time `json:"checked,omitzero"`
}

// Status is where the user stands.
type Status struct {
	Licensed bool
	Trial    bool // Pro while the trial lasts
	DaysLeft int
}

// Pro reports whether Pro's features are on.
func (s Status) Pro() bool { return s.Licensed || s.Trial }

// Load reads the state at path, starting the trial on the first launch.
func Load(path string, now time.Time) State {
	var s State
	if b, err := os.ReadFile(path); err == nil {
		json.Unmarshal(b, &s)
	}
	if s.TrialStart.IsZero() || s.TrialStart.After(now) {
		s.TrialStart = now
		s.Save(path)
	}
	return s
}

// Save writes the state.
func (s State) Save(path string) error {
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	os.MkdirAll(filepath.Dir(path), 0o700)
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// Status returns where the user stands at a time.
func (s State) Status(now time.Time) Status {
	if s.Key != "" && s.Instance != "" {
		return Status{Licensed: true}
	}
	end := s.TrialStart.Add(TrialDays * 24 * time.Hour)
	if now.Before(end) {
		left := int(end.Sub(now).Hours()/24) + 1
		return Status{Trial: true, DaysLeft: min(left, TrialDays)}
	}
	return Status{}
}

// API is Lemon Squeezy's license API.
var API = "https://api.lemonsqueezy.com/v1/licenses"

// ErrInvalid is a key Lemon Squeezy does not know, or one already active
// on as many Macs as it may be.
var ErrInvalid = errors.New("license: invalid key")

type response struct {
	Activated   bool    `json:"activated"`
	Deactivated bool    `json:"deactivated"`
	Valid       bool    `json:"valid"`
	Error       *string `json:"error"`
	LicenseKey  struct {
		Status string `json:"status"`
	} `json:"license_key"`
	Instance *struct {
		ID string `json:"id"`
	} `json:"instance"`
	Meta struct {
		CustomerEmail string `json:"customer_email"`
	} `json:"meta"`
}

func post(ctx context.Context, action string, form url.Values) (*response, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", API+"/"+action, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	var r response
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		return nil, fmt.Errorf("license: %s: %w", res.Status, err)
	}
	if r.Error != nil && *r.Error != "" {
		return &r, fmt.Errorf("%w: %s", ErrInvalid, *r.Error)
	}
	return &r, nil
}

// Activate activates a key on this Mac, named name in the customer's
// list of activations.
func (s *State) Activate(ctx context.Context, key, name string, now time.Time) error {
	key = strings.TrimSpace(key)
	if key == "" {
		return ErrInvalid
	}
	r, err := post(ctx, "activate", url.Values{"license_key": {key}, "instance_name": {name}})
	if err != nil {
		return err
	}
	if !r.Activated || r.Instance == nil {
		return ErrInvalid
	}
	s.Key, s.Instance, s.Email = key, r.Instance.ID, r.Meta.CustomerEmail
	s.Activated, s.Checked = now, now
	return nil
}

// Deactivate frees this Mac's activation, for the key to serve another.
func (s *State) Deactivate(ctx context.Context) error {
	if s.Key == "" {
		return nil
	}
	if s.Instance != "" {
		if _, err := post(ctx, "deactivate", url.Values{"license_key": {s.Key}, "instance_id": {s.Instance}}); err != nil && !errors.Is(err, ErrInvalid) {
			return err
		}
	}
	s.Key, s.Instance, s.Email = "", "", ""
	return nil
}

// Validate asks whether the activation still holds, as after a refund. A
// network failure keeps the license: Windeck works offline.
func (s *State) Validate(ctx context.Context, now time.Time) error {
	if s.Key == "" || s.Instance == "" {
		return nil
	}
	r, err := post(ctx, "validate", url.Values{"license_key": {s.Key}, "instance_id": {s.Instance}})
	if errors.Is(err, ErrInvalid) || err == nil && !r.Valid {
		s.Key, s.Instance = "", ""
		return ErrInvalid
	}
	if err != nil {
		return err
	}
	s.Checked = now
	return nil
}
