package license

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"
)

var t0 = time.Date(2026, 10, 1, 9, 0, 0, 0, time.UTC)

func TestTrial(t *testing.T) {
	path := filepath.Join(t.TempDir(), "license.json")
	s := Load(path, t0)
	if st := s.Status(t0); !st.Trial || st.DaysLeft != TrialDays || !st.Pro() {
		t.Fatalf("first day %+v", st)
	}
	if st := Load(path, t0.Add(13*24*time.Hour)).Status(t0.Add(13*24*time.Hour + time.Hour)); !st.Trial || st.DaysLeft != 1 {
		t.Fatalf("last day %+v", st)
	}
	if st := s.Status(t0.Add(15 * 24 * time.Hour)); st.Pro() {
		t.Fatalf("after the trial %+v", st)
	}
}

func TestActivateAndDeactivate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		switch r.URL.Path {
		case "/activate":
			if r.Form.Get("license_key") != "GOOD" {
				w.WriteHeader(400)
				w.Write([]byte(`{"activated":false,"error":"license_key not found."}`))
				return
			}
			w.Write([]byte(`{"activated":true,"error":null,"instance":{"id":"inst-1"},"meta":{"customer_email":"a@b.c"}}`))
		case "/validate":
			w.Write([]byte(`{"valid":true,"error":null}`))
		case "/deactivate":
			w.Write([]byte(`{"deactivated":true,"error":null}`))
		}
	}))
	defer srv.Close()
	API = srv.URL
	var s State
	s.TrialStart = t0
	ctx := context.Background()
	if err := s.Activate(ctx, "BAD", "Mac", t0); !errors.Is(err, ErrInvalid) {
		t.Fatalf("bad key: %v", err)
	}
	if err := s.Activate(ctx, " GOOD ", "Mac", t0); err != nil {
		t.Fatal(err)
	}
	later := t0.Add(400 * 24 * time.Hour)
	if st := s.Status(later); !st.Licensed || !st.Pro() || s.Email != "a@b.c" {
		t.Fatalf("licensed %+v %+v", st, s)
	}
	if err := s.Validate(ctx, later); err != nil {
		t.Fatal(err)
	}
	if err := s.Deactivate(ctx); err != nil {
		t.Fatal(err)
	}
	if s.Status(later).Pro() {
		t.Fatal("still Pro after deactivating")
	}
}
