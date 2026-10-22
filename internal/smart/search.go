package smart

import (
	"strings"
	"unicode"
)

// Match scores how well a query matches s, its letters in order: higher
// is better, -1 no match. Starts of words, runs of letters and a match at
// the very start score more; spaces in the query separate words matched
// in turn. Case is ignored.
func Match(query, s string) int {
	query = strings.TrimSpace(query)
	if query == "" {
		return 0
	}
	q := []rune(strings.ToLower(query))
	r := []rune(strings.ToLower(s))
	// A plain substring beats letters strewn about.
	if i := strings.Index(strings.ToLower(s), strings.ToLower(query)); i >= 0 {
		bonus := 400
		if i == 0 {
			bonus = 600
		} else if pr := []rune(strings.ToLower(s)[:i]); len(pr) > 0 && isBoundary(pr[len(pr)-1]) {
			bonus = 500
		}
		return bonus + len(q)*20 - len(r)/4
	}
	score, qi, run := 0, 0, 0
	for i := 0; i < len(r) && qi < len(q); i++ {
		if q[qi] == ' ' {
			qi++
			run = 0
			if qi == len(q) {
				break
			}
		}
		if r[i] == q[qi] {
			bonus := 1
			if i == 0 || isBoundary(r[i-1]) {
				bonus += 6
			}
			run++
			score += bonus + run*2
			qi++
		} else {
			run = 0
		}
	}
	if qi < len(q) {
		return -1
	}
	return score*10/len(q) - len(r)/4
}

func isBoundary(r rune) bool {
	return !unicode.IsLetter(r) && !unicode.IsDigit(r) || unicode.Is(unicode.Han, r)
}

// Result is a window matched by a query.
type Result struct {
	Item  Item
	Score float64
}

// Rank scores items for a query: by how well their titles and apps match
// and, when personal, by how often and lately the user goes to them, so
// that the same letters bring the user's usual window first. Items that
// do not match are left out; the rest keep their order among equals.
func Rank(m *Model, query string, items []Item, personal bool, at func() int64) []Result {
	var out []Result
	for _, it := range items {
		t := Match(query, it.Title)
		a := Match(query, it.Name)
		both := Match(query, it.Name+" "+it.Title)
		best := max(t, a-50, both-80)
		if best < 0 {
			continue
		}
		score := float64(best)
		if personal && m != nil {
			// A boost that orders close matches, not one that lets a poor
			// match beat a good one.
			f := m.frecencyAt(it, at())
			score += 120 * f / (f + 4)
		}
		out = append(out, Result{Item: it, Score: score})
	}
	sortStable(out)
	return out
}

func sortStable(rs []Result) {
	for i := 1; i < len(rs); i++ {
		for j := i; j > 0 && rs[j].Score > rs[j-1].Score; j-- {
			rs[j], rs[j-1] = rs[j-1], rs[j]
		}
	}
}

func (m *Model) frecencyAt(i Item, now int64) float64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	f := 0.0
	if s := m.Apps[i.App]; s != nil {
		f += s.Value(now)
	}
	if s := m.Titles[i.titleKey()]; s != nil {
		f += 2 * s.Value(now)
	}
	return f
}

// Positions are the runes of s that a query matches, as Match finds
// them, to show in bold: the substring when there is one, else the
// letters in turn. Nil when it does not match.
func Positions(query, s string) []int {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil
	}
	q := []rune(strings.ToLower(query))
	r := []rune(strings.ToLower(s))
	if len(r) != len([]rune(s)) {
		return nil // a case mapping that changes lengths: no bold
	}
	if i := runeIndex(r, q); i >= 0 {
		out := make([]int, len(q))
		for k := range q {
			out[k] = i + k
		}
		return out
	}
	var out []int
	qi := 0
	for i := 0; i < len(r) && qi < len(q); i++ {
		if q[qi] == ' ' {
			qi++
			if qi == len(q) {
				break
			}
		}
		if r[i] == q[qi] {
			out = append(out, i)
			qi++
		}
	}
	if qi < len(q) {
		return nil
	}
	return out
}

func runeIndex(s, sub []rune) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if string(s[i:i+len(sub)]) == string(sub) {
			return i
		}
	}
	return -1
}
