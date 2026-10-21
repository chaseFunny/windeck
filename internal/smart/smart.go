// Package smart is what Windeck learns of the user's habits, on this Mac
// only: how often and how lately each app and window is used (frecency),
// which ranks search results, and which window tends to follow which (a
// Markov chain of transitions), which predicts the cycle's second place.
//
// Nothing here runs as the shortcut is pressed: the model learns as
// windows are focused, and the prediction is a lookup.
package smart

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// halfLife is how long a visit takes to count half as much.
const halfLife = 72 * time.Hour

// Stat is a decaying count: Score as of At.
type Stat struct {
	Score float64 `json:"s"`
	At    int64   `json:"t"` // Unix seconds
}

func decay(score float64, from, to int64) float64 {
	if to <= from {
		return score
	}
	return score * math.Exp2(-float64(to-from)/halfLife.Seconds())
}

// Value is the stat's score at a time.
func (s Stat) Value(now int64) float64 { return decay(s.Score, s.At, now) }

func (s *Stat) add(now int64, n float64) {
	s.Score = decay(s.Score, s.At, now) + n
	s.At = now
}

// Item is a window as the model knows it: by its app and its title, which
// outlive the window's number across launches, and by that number within
// a run.
type Item struct {
	ID    uint32
	App   string // the app's bundle ID
	Name  string // the app's name, which search matches
	Title string
}

func (i Item) titleKey() string { return i.App + "\x1f" + i.Title }

// Metrics are counts of how the switcher served the user, kept on this
// Mac, for the Settings window to show.
type Metrics struct {
	Day       string `json:"day"`       // the day Switches counts
	Switches  int    `json:"switches"`  // switches that day
	Shown     int    `json:"shown"`     // switches made with the panel shown
	TopTwo    int    `json:"topTwo"`    // of those, to the cycle's first two places
	Predicted int    `json:"predicted"` // of those, to the predicted window
	Keys      int    `json:"keys"`      // keys pressed in those switches
	Total     int    `json:"total"`     // switches ever
}

// Model is what was learned. Its methods are safe for concurrent use.
type Model struct {
	mu    sync.Mutex
	path  string
	dirty bool

	Apps   map[string]*Stat            `json:"apps"`
	Titles map[string]*Stat            `json:"titles"`
	Trans  map[string]map[string]*Stat `json:"transitions"` // app → app
	M      Metrics                     `json:"metrics"`

	// Transitions between windows of this run, by their numbers, which do
	// not outlive it.
	winTrans map[uint32]map[uint32]float64
}

// New returns an empty model saved at path ("" for none).
func New(path string) *Model {
	return &Model{path: path, Apps: map[string]*Stat{}, Titles: map[string]*Stat{}, Trans: map[string]map[string]*Stat{},
		winTrans: map[uint32]map[uint32]float64{}}
}

// Load reads the model saved at path, or returns an empty one.
func Load(path string) *Model {
	m := New(path)
	if b, err := os.ReadFile(path); err == nil {
		json.Unmarshal(b, m)
	}
	if m.Apps == nil {
		m.Apps = map[string]*Stat{}
	}
	if m.Titles == nil {
		m.Titles = map[string]*Stat{}
	}
	if m.Trans == nil {
		m.Trans = map[string]map[string]*Stat{}
	}
	return m
}

// Save writes the model, if it changed.
func (m *Model) Save() error {
	m.mu.Lock()
	if !m.dirty || m.path == "" {
		m.mu.Unlock()
		return nil
	}
	m.prune(time.Now().Unix())
	b, err := json.Marshal(m)
	m.dirty = false
	m.mu.Unlock()
	if err != nil {
		return err
	}
	os.MkdirAll(filepath.Dir(m.path), 0o700)
	tmp := m.path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, m.path)
}

// prune forgets what decayed to nothing, so the file stays small.
func (m *Model) prune(now int64) {
	const floor = 0.01
	for k, s := range m.Titles {
		if s.Value(now) < floor {
			delete(m.Titles, k)
		}
	}
	for k, s := range m.Apps {
		if s.Value(now) < floor {
			delete(m.Apps, k)
		}
	}
	for from, tos := range m.Trans {
		for to, s := range tos {
			if s.Value(now) < floor {
				delete(tos, to)
			}
		}
		if len(tos) == 0 {
			delete(m.Trans, from)
		}
	}
}

// Reset forgets everything learned.
func (m *Model) Reset() {
	m.mu.Lock()
	m.Apps, m.Titles, m.Trans = map[string]*Stat{}, map[string]*Stat{}, map[string]map[string]*Stat{}
	m.winTrans = map[uint32]map[uint32]float64{}
	m.M = Metrics{}
	m.dirty = true
	m.mu.Unlock()
	if m.path != "" {
		os.Remove(m.path)
	}
}

// Visit learns that the user went from one window to another.
func (m *Model) Visit(from, to Item, at time.Time) {
	now := at.Unix()
	m.mu.Lock()
	defer m.mu.Unlock()
	m.dirty = true
	stat := func(t map[string]*Stat, k string) *Stat {
		s := t[k]
		if s == nil {
			s = &Stat{}
			t[k] = s
		}
		return s
	}
	stat(m.Apps, to.App).add(now, 1)
	if to.Title != "" {
		stat(m.Titles, to.titleKey()).add(now, 1)
	}
	if from.App == "" || from.ID == to.ID {
		return
	}
	tos := m.Trans[from.App]
	if tos == nil {
		tos = map[string]*Stat{}
		m.Trans[from.App] = tos
	}
	stat(tos, to.App).add(now, 1)
	if from.ID != 0 && to.ID != 0 {
		w := m.winTrans[from.ID]
		if w == nil {
			w = map[uint32]float64{}
			m.winTrans[from.ID] = w
		}
		w[to.ID]++
	}
}

// Frecency is how often and lately the user went to a window, its title
// counting more than its app.
func (m *Model) Frecency(i Item, at time.Time) float64 { return m.frecencyAt(i, at.Unix()) }

// Predict returns the window to put in the cycle's second place: the one
// the user most likely goes to from the current window, past the previous
// one. It returns 0 when the model knows too little to beat plain
// recency. candidates are the cycle in its plain order: the first is the
// previous window, whose place the prediction never takes.
func (m *Model) Predict(cur Item, candidates []Item, at time.Time) uint32 {
	now := at.Unix()
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(candidates) < 2 {
		return 0
	}
	var winTotal float64
	for _, n := range m.winTrans[cur.ID] {
		winTotal += n
	}
	var appTotal float64
	for _, s := range m.Trans[cur.App] {
		appTotal += s.Value(now)
	}
	// Too few transitions from here to say anything.
	if winTotal < 2 && appTotal < 3 {
		return 0
	}
	score := func(c Item) float64 {
		s := 0.0
		if winTotal > 0 {
			s += 3 * m.winTrans[cur.ID][c.ID] / winTotal
		}
		if appTotal > 0 {
			if t := m.Trans[cur.App][c.App]; t != nil {
				s += 1.5 * t.Value(now) / appTotal
			}
		}
		return s
	}
	best, bestScore := uint32(0), 0.0
	for _, c := range candidates[1:] {
		if c.ID == cur.ID {
			continue
		}
		if s := score(c); s > bestScore {
			best, bestScore = c.ID, s
		}
	}
	// It must stand out, and beat the window recency already puts second
	// by a margin: otherwise the plain order stays.
	if best == candidates[1].ID || bestScore < 0.45 || bestScore < score(candidates[1])*1.25 {
		return 0
	}
	return best
}

// Switched records a switch for the metrics: shown tells that the panel
// showed, place where the window was in the cycle (0 the first),
// predicted that it was the predicted one, keys how many keys it took.
func (m *Model) Switched(at time.Time, shown bool, place int, predicted bool, keys int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.dirty = true
	day := at.Format("2006-01-02")
	if m.M.Day != day {
		m.M.Day, m.M.Switches = day, 0
	}
	m.M.Switches++
	m.M.Total++
	if !shown {
		return
	}
	m.M.Shown++
	m.M.Keys += keys
	if place >= 0 && place < 2 {
		m.M.TopTwo++
	}
	if predicted {
		m.M.Predicted++
	}
}

// Metrics returns the counts.
func (m *Model) Metrics() Metrics {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.M
}
