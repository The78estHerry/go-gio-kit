package giowidget

import (
	"image"
	"testing"
)

// fakeScreen is a monitor layout with one window.
type fakeScreen struct {
	win     image.Rectangle
	works   []image.Rectangle
	pointer image.Point
}

func (f *fakeScreen) windowRect() (image.Rectangle, bool) { return f.win, true }
func (f *fakeScreen) move(p image.Point)                  { f.win = f.win.Add(p.Sub(f.win.Min)) }
func (f *fakeScreen) workAreas() []image.Rectangle        { return f.works }
func (f *fakeScreen) primaryWork() image.Rectangle        { return f.works[0] }
func (f *fakeScreen) cursor() (image.Point, bool)         { return f.pointer, true }

var (
	primary = image.Rect(0, 0, 1920, 1040)
	second  = image.Rect(1920, 0, 3840, 1080)
	size    = image.Pt(260, 64)
)

func TestDefaultPos(t *testing.T) {
	if got, want := defaultPos(primary, size), image.Pt(1920-260-24, 1040-64-24); got != want {
		t.Errorf("defaultPos = %v, want %v", got, want)
	}
}

func TestBesideCursor(t *testing.T) {
	for _, c := range []struct {
		cursor, want image.Point
	}{
		{image.Pt(100, 100), image.Pt(116, 116)},                     // below right
		{image.Pt(1900, 100), image.Pt(1900-16-260, 116)},            // flips left
		{image.Pt(100, 1030), image.Pt(116, 1030-16-64)},             // flips up
		{image.Pt(5, 1035), image.Pt(21, 1035-16-64)},                // corner
		{image.Pt(200, 30), image.Pt(216, 46)},                       // fits
		{image.Pt(1910, 1035), image.Pt(1910-16-260, 1035-16-64)},    // both
		{image.Pt(-50, -50), image.Pt(0, 0)},                         // clamped into the work area
		{image.Pt(1919, 1039), image.Pt(1919-16-260, 1039-16-64)},    // last pixel
		{image.Pt(260/2, 64/2), image.Pt(260/2+16, 64/2+16)},         // near origin
		{image.Pt(1920-100, 500), image.Pt(1920-100-16-260, 500+16)}, // near right edge
	} {
		if got := besideCursor(c.cursor, size, primary); got != c.want {
			t.Errorf("besideCursor(%v) = %v, want %v", c.cursor, got, c.want)
		}
	}
}

func TestVisibleEnough(t *testing.T) {
	works := []image.Rectangle{primary, second}
	for _, c := range []struct {
		r    image.Rectangle
		want bool
	}{
		{image.Rect(100, 100, 360, 164), true},
		{image.Rect(100, 1020, 360, 1084), false}, // mostly below the bottom edge
		{image.Rect(1800, 100, 2060, 164), true},  // split over two monitors, half on one
		{image.Rect(1700, 100, 1960, 164), true},  // most on the primary monitor
		{image.Rect(4000, 100, 4260, 164), false}, // off every monitor
		{image.Rect(0, 0, 0, 0), false},
	} {
		if got := visibleEnough(c.r, works); got != c.want {
			t.Errorf("visibleEnough(%v) = %v, want %v", c.r, got, c.want)
		}
	}
}

func TestPlacerMonitorUnplugged(t *testing.T) {
	onSecond := image.Rect(2500, 500, 2760, 564)
	s := &fakeScreen{win: onSecond, works: []image.Rectangle{primary, second}}
	var p placer

	p.check(s) // both monitors present: nothing happens
	if s.win != onSecond {
		t.Fatalf("moved while visible: %v", s.win)
	}

	s.works = []image.Rectangle{primary} // second monitor unplugged
	p.check(s)
	if want := defaultPos(primary, size); s.win.Min != want {
		t.Fatalf("not moved to the primary monitor: %v, want %v", s.win.Min, want)
	}
	if got, _ := p.saved(s); got != onSecond.Min {
		t.Errorf("saved = %v, want the original %v", got, onSecond.Min)
	}

	p.check(s) // still unplugged: stays
	s.works = []image.Rectangle{primary, second}
	p.check(s) // plugged back: returns
	if s.win != onSecond {
		t.Errorf("not moved back: %v", s.win)
	}
}

func TestPlacerUserDragged(t *testing.T) {
	onSecond := image.Rect(2500, 500, 2760, 564)
	s := &fakeScreen{win: onSecond, works: []image.Rectangle{primary}}
	var p placer
	p.check(s) // displaced to the primary monitor
	s.move(image.Pt(300, 300))
	p.check(s) // the user dragged it: that place wins
	s.works = []image.Rectangle{primary, second}
	p.check(s)
	if s.win.Min != image.Pt(300, 300) {
		t.Errorf("dragged widget was moved: %v", s.win.Min)
	}
	if got, _ := p.saved(s); got != image.Pt(300, 300) {
		t.Errorf("saved = %v, want the dragged place", got)
	}
}
