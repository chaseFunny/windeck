package snapshot

import (
	"image"
	"testing"
	"time"
)

func img(w, h int) *image.RGBA { return image.NewRGBA(image.Rect(0, 0, w, h)) }

func TestBudgetEvictsLeastRecentlyUsed(t *testing.T) {
	p := New(3 * 100 * 100 * 4)
	for id := uint32(1); id <= 3; id++ {
		p.Put(id, img(100, 100), time.Now())
	}
	p.Get(1) // 1 is used: 2 is now the oldest
	p.Put(4, img(100, 100), time.Now())
	if _, ok := p.Get(2); ok {
		t.Fatal("the least recently used picture stayed")
	}
	for _, id := range []uint32{1, 3, 4} {
		if _, ok := p.Get(id); !ok {
			t.Fatalf("picture %d went", id)
		}
	}
	if p.Used() != 3*100*100*4 {
		t.Fatalf("used %d", p.Used())
	}
}

func TestReplaceKeepAndClear(t *testing.T) {
	p := New(DefaultBudget)
	p.Put(1, img(10, 10), time.Now())
	p.Put(1, img(20, 20), time.Now())
	if p.Len() != 1 || p.Used() != 20*20*4 {
		t.Fatalf("len %d used %d", p.Len(), p.Used())
	}
	p.Put(2, img(10, 10), time.Now())
	p.Keep(func(id uint32) bool { return id == 2 })
	if _, ok := p.Get(1); ok || p.Len() != 1 {
		t.Fatal("Keep kept a gone window's picture")
	}
	p.Clear()
	if p.Len() != 0 || p.Used() != 0 {
		t.Fatal("Clear")
	}
}
