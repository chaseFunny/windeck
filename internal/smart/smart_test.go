package smart

import (
	"path/filepath"
	"testing"
	"time"
)

var t0 = time.Date(2026, 10, 1, 9, 0, 0, 0, time.UTC)

func item(id uint32, app, title string) Item { return Item{ID: id, App: app, Name: app, Title: title} }

func TestPredictLearnsTransitions(t *testing.T) {
	m := New("")
	code, docs, chat, term := item(1, "code", "main.go"), item(2, "chrome", "Go docs"), item(3, "slack", "general"), item(4, "iterm", "zsh")
	// The user goes code → term again and again.
	for i := range 6 {
		at := t0.Add(time.Duration(i) * time.Minute)
		m.Visit(code, term, at)
		m.Visit(term, code, at)
	}
	// The cycle's plain order, from code: docs (previous), chat, term.
	got := m.Predict(code, []Item{docs, chat, term}, t0.Add(time.Hour))
	if got != term.ID {
		t.Fatalf("predicted %d, want %d", got, term.ID)
	}
	// The previous window's place is never taken.
	if got := m.Predict(code, []Item{term, chat, docs}, t0.Add(time.Hour)); got != 0 {
		t.Fatalf("predicted %d where the previous window is the likely one", got)
	}
}

func TestPredictWaitsForData(t *testing.T) {
	m := New("")
	a, b, c := item(1, "a", ""), item(2, "b", ""), item(3, "c", "")
	m.Visit(a, c, t0)
	if got := m.Predict(a, []Item{b, item(4, "d", ""), c}, t0); got != 0 {
		t.Fatalf("predicted %d from a single visit", got)
	}
}

func TestFrecencyDecays(t *testing.T) {
	m := New("")
	w := item(1, "app", "doc")
	m.Visit(Item{}, w, t0)
	now := m.Frecency(w, t0)
	later := m.Frecency(w, t0.Add(halfLife))
	if now <= 0 || later > now*0.51 || later < now*0.49 {
		t.Fatalf("frecency %v then %v, want half after a half-life", now, later)
	}
}

func TestRankPrefersHabit(t *testing.T) {
	m := New("")
	a := item(1, "chrome", "Report — Q3")
	b := item(2, "chrome", "Report — Q4")
	for range 5 {
		m.Visit(Item{}, b, t0)
	}
	at := func() int64 { return t0.Unix() }
	rs := Rank(m, "report", []Item{a, b}, true, at)
	if len(rs) != 2 || rs[0].Item.ID != 2 {
		t.Fatalf("personal ranking %+v", rs)
	}
	rs = Rank(m, "report", []Item{a, b}, false, at)
	if rs[0].Item.ID != 1 {
		t.Fatalf("plain ranking %+v, want the list's order among equals", rs)
	}
	// A poor match does not beat a good one on habit alone.
	c := item(3, "notes", "rxexpxoxrxt")
	for range 50 {
		m.Visit(Item{}, c, t0)
	}
	rs = Rank(m, "report", []Item{c, a}, true, at)
	if rs[0].Item.ID != 1 {
		t.Fatalf("habit beat the match: %+v", rs)
	}
}

func TestMatch(t *testing.T) {
	for _, tc := range []struct {
		q, s string
		ok   bool
	}{
		{"rep", "Weekly report", true},
		{"wr", "Weekly report", true},
		{"报销", "差旅报销单 - Excel", true},
		{"xyz", "Weekly report", false},
		{"", "anything", true},
	} {
		if got := Match(tc.q, tc.s) >= 0; got != tc.ok {
			t.Errorf("Match(%q, %q) = %v", tc.q, tc.s, got)
		}
	}
	if Match("rep", "report") <= Match("rep", "a reply in a post") {
		t.Error("a match at the start should score more")
	}
}

func TestSaveLoadReset(t *testing.T) {
	path := filepath.Join(t.TempDir(), "learned.json")
	m := New(path)
	m.Visit(item(1, "a", "x"), item(2, "b", "y"), time.Now())
	m.Switched(time.Now(), true, 1, false, 2)
	if err := m.Save(); err != nil {
		t.Fatal(err)
	}
	l := Load(path)
	if l.Apps["b"] == nil || l.Trans["a"]["b"] == nil || l.Metrics().TopTwo != 1 {
		t.Fatalf("loaded %+v", l)
	}
	l.Reset()
	if len(l.Apps) != 0 || l.Metrics().Total != 0 {
		t.Fatal("reset kept data")
	}
}
