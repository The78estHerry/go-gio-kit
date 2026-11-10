// Package giowidget turns a Gio window into a small desktop widget:
// borderless, optionally always on top, dragged by its content, kept out of
// the taskbar, and placed at a remembered position, the bottom-right corner
// or next to the mouse pointer.
//
//	w := giowidget.New(giowidget.Options{Title: "My widget", Width: 260, Height: 64, TopMost: true})
//	for {
//		e := w.Window().Event()
//		w.HandleEvent(e)
//		switch e := e.(type) {
//		case app.DestroyEvent:
//			savePosition(w.Position()) // remember for next time
//			return
//		case app.FrameEvent:
//			gtx := app.NewContext(&ops, e)
//			giowidget.DragArea(gtx) // the background drags the window
//			drawContent(gtx)
//			e.Frame(gtx.Ops)
//		}
//	}
//
// Placement works on Windows. Elsewhere the window manager places the
// window: Gio cannot set window positions on X11 or Wayland.
package giowidget

import (
	"image"
	"sync"
	"time"

	"gioui.org/app"
	"gioui.org/io/event"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/unit"
)

// Options configure a widget.
type Options struct {
	Title         string
	Width, Height unit.Dp
	TopMost       bool
	// Position is the remembered top-left corner in screen pixels, as
	// returned by Widget.Position. Nil means the bottom-right corner of
	// the primary monitor.
	Position *image.Point
	// NearCursor opens the widget next to the mouse pointer instead.
	NearCursor bool
}

// checkInterval is how often the placement is checked against the current
// monitor layout.
const checkInterval = time.Second

// Widget is one widget window. Create a new Widget to show it again after
// it was closed.
type Widget struct {
	win    *app.Window
	opts   Options
	placer placer

	mu     sync.Mutex
	scr    screen // nil until the native window exists
	pos    image.Point
	hasPos bool
	stop   chan struct{}
}

// New creates the window; it appears once its events are processed.
func New(o Options) *Widget {
	w := &Widget{win: new(app.Window), opts: o}
	if o.Position != nil {
		w.pos, w.hasPos = *o.Position, true
	}
	w.win.Option(app.Title(o.Title), app.Decorated(false), app.TopMost(o.TopMost))
	w.win.Option(sizeOptions(o.Width, o.Height)...)
	return w
}

func sizeOptions(wd, ht unit.Dp) []app.Option {
	return []app.Option{app.Size(wd, ht), app.MinSize(wd, ht), app.MaxSize(wd, ht)}
}

// Window returns the Gio window; read its events and pass each one to
// HandleEvent.
func (w *Widget) Window() *app.Window { return w.win }

// HandleEvent must see every event of the window.
func (w *Widget) HandleEvent(e event.Event) {
	switch e := e.(type) {
	case app.ViewEvent:
		s := newScreen(e)
		if s == nil {
			return
		}
		w.mu.Lock()
		if w.stop != nil {
			close(w.stop)
		}
		stop := make(chan struct{})
		w.scr, w.stop = s, stop
		w.mu.Unlock()
		// The window thread is waiting for this event to be handled;
		// calling into the native window synchronously here deadlocks.
		go w.place(s, stop)
	case app.DestroyEvent:
		w.mu.Lock()
		if s := w.scr; s != nil {
			if p, ok := w.placer.saved(s); ok {
				w.pos, w.hasPos = p, true
			}
		}
		if w.stop != nil {
			close(w.stop)
			w.stop = nil
		}
		w.scr = nil
		w.mu.Unlock()
	}
}

func (w *Widget) place(s screen, stop <-chan struct{}) {
	setupWindow(s)
	w.mu.Lock()
	pos, hasPos := w.pos, w.hasPos
	w.mu.Unlock()
	r, ok := s.windowRect()
	if w.opts.NearCursor && ok {
		if c, ok := s.cursor(); ok {
			pos, hasPos = besideCursor(c, r.Size(), workAt(s, c)), true
		}
	}
	if !hasPos && ok {
		pos, hasPos = defaultPos(s.primaryWork(), r.Size()), true
	}
	if hasPos {
		s.move(pos)
	}
	w.placer.reset()
	w.placer.check(s)
	showWindow(s, w.opts.NearCursor)

	t := time.NewTicker(checkInterval)
	defer t.Stop()
	for {
		select {
		case <-stop:
			return
		case <-t.C:
			w.placer.check(s)
		}
	}
}

// workAt returns the work area of the monitor containing p, or the primary
// one.
func workAt(s screen, p image.Point) image.Rectangle {
	for _, wa := range s.workAreas() {
		if p.In(wa) {
			return wa
		}
	}
	return s.primaryWork()
}

// Position returns the top-left corner to remember for next time. While
// the widget's monitor is unplugged it returns the original place. ok is
// false where positions are not supported or nothing is known yet.
func (w *Widget) Position() (p image.Point, ok bool) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.scr != nil {
		return w.placer.saved(w.scr)
	}
	return w.pos, w.hasPos
}

// MoveToCursor moves the widget next to the mouse pointer and brings it to
// the front.
func (w *Widget) MoveToCursor() {
	w.mu.Lock()
	s := w.scr
	w.mu.Unlock()
	if s == nil {
		return
	}
	go func() {
		r, ok := s.windowRect()
		c, ok2 := s.cursor()
		if !ok || !ok2 {
			return
		}
		w.placer.reset()
		s.move(besideCursor(c, r.Size(), workAt(s, c)))
		showWindow(s, true)
	}()
}

// BringToFront shows the widget above other windows and activates it.
func (w *Widget) BringToFront() {
	w.mu.Lock()
	s := w.scr
	w.mu.Unlock()
	if s != nil {
		go showWindow(s, true)
	}
}

// SetTopMost keeps the widget above other windows or not. Like SetSize, it
// is safe to call from the window's event loop.
func (w *Widget) SetTopMost(on bool) { go w.win.Option(app.TopMost(on)) }

// SetSize changes the fixed size.
func (w *Widget) SetSize(wd, ht unit.Dp) { go w.win.Option(sizeOptions(wd, ht)...) }

// DragArea makes the whole current constraint area drag the window. Lay it
// out before the content, so buttons on top still receive clicks.
func DragArea(gtx layout.Context) {
	defer clip.Rect{Max: gtx.Constraints.Max}.Push(gtx.Ops).Pop()
	system.ActionInputOp(system.ActionMove).Add(gtx.Ops)
}
