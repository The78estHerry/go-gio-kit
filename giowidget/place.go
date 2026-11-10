package giowidget

import (
	"image"
	"sync"
)

// screen is what the placement logic needs from the operating system. All
// coordinates are physical pixels in the virtual desktop.
type screen interface {
	// windowRect returns the widget window's frame.
	windowRect() (image.Rectangle, bool)
	// move puts the window's top-left corner at p without activating it.
	move(p image.Point)
	// workAreas returns every monitor's work area (without taskbars).
	workAreas() []image.Rectangle
	// primaryWork returns the primary monitor's work area.
	primaryWork() image.Rectangle
	// cursor returns the mouse pointer position.
	cursor() (image.Point, bool)
}

// Distances in pixels.
const (
	cornerMargin = 24 // from the work area corner to the default position
	cursorOffset = 16 // between the pointer and the widget
)

// defaultPos is the bottom-right corner of work, inset by cornerMargin.
func defaultPos(work image.Rectangle, size image.Point) image.Point {
	return work.Max.Sub(size).Sub(image.Pt(cornerMargin, cornerMargin))
}

// besideCursor places a window of the given size below and to the right of
// the pointer, flipping to the left or above where it does not fit, then
// clamps it into work.
func besideCursor(c, size image.Point, work image.Rectangle) image.Point {
	p := c.Add(image.Pt(cursorOffset, cursorOffset))
	if p.X+size.X > work.Max.X {
		p.X = c.X - cursorOffset - size.X
	}
	if p.Y+size.Y > work.Max.Y {
		p.Y = c.Y - cursorOffset - size.Y
	}
	p.X = max(work.Min.X, min(p.X, work.Max.X-size.X))
	p.Y = max(work.Min.Y, min(p.Y, work.Max.Y-size.Y))
	return p
}

// visibleEnough reports whether at least half of r lies in one work area.
func visibleEnough(r image.Rectangle, works []image.Rectangle) bool {
	area := int64(r.Dx()) * int64(r.Dy())
	if area <= 0 {
		return false
	}
	for _, w := range works {
		in := r.Intersect(w)
		if 2*int64(in.Dx())*int64(in.Dy()) >= area {
			return true
		}
	}
	return false
}

// placer moves the widget to the primary monitor while the monitor it was
// on is gone, and back when that monitor returns. If the user drags it in
// the meantime, the new place wins.
type placer struct {
	mu        sync.Mutex
	displaced bool
	prefer    image.Point // where the widget belongs
	fallback  image.Point // where it was moved meanwhile
}

func (p *placer) reset() {
	p.mu.Lock()
	p.displaced = false
	p.mu.Unlock()
}

// check is called periodically.
func (p *placer) check(s screen) {
	r, ok := s.windowRect()
	if !ok {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if !p.displaced {
		if visibleEnough(r, s.workAreas()) {
			return
		}
		p.prefer = r.Min
		p.fallback = defaultPos(s.primaryWork(), r.Size())
		p.displaced = true
		s.move(p.fallback)
		return
	}
	if r.Min != p.fallback {
		p.displaced = false // dragged by the user: keep the new place
		return
	}
	if visibleEnough(r.Add(p.prefer.Sub(r.Min)), s.workAreas()) {
		p.displaced = false
		s.move(p.prefer)
	}
}

// saved returns the position to remember: while displaced, the original
// one, so the widget returns there next time if that monitor is present.
func (p *placer) saved(s screen) (image.Point, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.displaced {
		return p.prefer, true
	}
	r, ok := s.windowRect()
	return r.Min, ok
}
