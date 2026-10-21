// Package snapshot keeps windows' pictures in memory: the last frame of a
// window as the user left it, which shows for windows the window server
// can no longer capture (minimized before, hidden, on another Space), and
// live captures, which show at once the next time.
//
// The pool has a budget of bytes: past it, the pictures used least
// recently go first, so Windeck stays within its memory target however
// many windows the user has.
package snapshot

import (
	"container/list"
	"image"
	"sync"
	"time"
)

// DefaultBudget is the pool's size: what keeps Windeck under 150 MB with
// everything else it holds.
const DefaultBudget = 96 << 20

// Shot is a window's picture and when it was taken.
type Shot struct {
	Image *image.RGBA
	At    time.Time
}

type entry struct {
	id   uint32
	shot Shot
}

// Pool is the pictures, by window. It is safe for concurrent use.
type Pool struct {
	mu     sync.Mutex
	budget int
	used   int
	lru    *list.List // of *entry, the most recently used first
	byID   map[uint32]*list.Element
}

// New returns an empty pool of budget bytes.
func New(budget int) *Pool {
	return &Pool{budget: budget, lru: list.New(), byID: map[uint32]*list.Element{}}
}

func size(img *image.RGBA) int {
	if img == nil {
		return 0
	}
	return len(img.Pix)
}

// Put keeps a window's picture, replacing the one it had.
func (p *Pool) Put(id uint32, img *image.RGBA, at time.Time) {
	if img == nil {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if el, ok := p.byID[id]; ok {
		e := el.Value.(*entry)
		p.used -= size(e.shot.Image)
		e.shot = Shot{img, at}
		p.used += size(img)
		p.lru.MoveToFront(el)
	} else {
		p.byID[id] = p.lru.PushFront(&entry{id, Shot{img, at}})
		p.used += size(img)
	}
	p.evict()
}

func (p *Pool) evict() {
	for p.used > p.budget && p.lru.Len() > 1 {
		el := p.lru.Back()
		e := el.Value.(*entry)
		p.lru.Remove(el)
		delete(p.byID, e.id)
		p.used -= size(e.shot.Image)
	}
}

// Get returns a window's picture, which counts as a use.
func (p *Pool) Get(id uint32) (Shot, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	el, ok := p.byID[id]
	if !ok {
		return Shot{}, false
	}
	p.lru.MoveToFront(el)
	return el.Value.(*entry).shot, true
}

// Remove forgets a window's picture.
func (p *Pool) Remove(id uint32) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if el, ok := p.byID[id]; ok {
		p.used -= size(el.Value.(*entry).shot.Image)
		p.lru.Remove(el)
		delete(p.byID, id)
	}
}

// Keep forgets the pictures of windows that are gone.
func (p *Pool) Keep(live func(id uint32) bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	for id, el := range p.byID {
		if !live(id) {
			p.used -= size(el.Value.(*entry).shot.Image)
			p.lru.Remove(el)
			delete(p.byID, id)
		}
	}
}

// Clear forgets every picture.
func (p *Pool) Clear() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.lru.Init()
	p.byID = map[uint32]*list.Element{}
	p.used = 0
}

// Used is the bytes the pictures take, and Len how many there are.
func (p *Pool) Used() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.used
}

func (p *Pool) Len() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.lru.Len()
}
