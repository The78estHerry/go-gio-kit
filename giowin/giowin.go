// Package giowin applies platform fixes and settings to a Gio window that Gio
// itself does not handle:
//
//   - Windows: works around Gio v0.10.3 treating GCS_DELTASTART as the start
//     of the IME composition, which breaks Japanese romaji input such as
//     "chi" or "kitte" (the flag is hidden from Gio; Chinese IMEs are not
//     affected either way). Also switches the native title bar between dark
//     and light.
//   - SystemPrefersDark reports the OS color scheme on Windows, macOS and
//     Linux (GNOME).
//
// Use one Window per app.Window and pass it every event:
//
//	var gw giowin.Window
//	for {
//		e := w.Event()
//		gw.HandleEvent(e)
//		...
//	}
package giowin

import "gioui.org/io/event"

// Window holds the native handle of one Gio window.
type Window struct {
	handle   uintptr
	dark     bool
	darkSet  bool
	platform platformState
}

// HandleEvent must see every event of the window. It picks up the native
// handle and installs the platform fixes.
func (w *Window) HandleEvent(e event.Event) {
	w.handleEvent(e)
}

// SetDarkTitleBar switches the native title bar between dark and light. It
// can be called before the window exists; the setting is applied as soon as
// the native handle is known.
func (w *Window) SetDarkTitleBar(dark bool) {
	w.dark, w.darkSet = dark, true
	w.applyTitleBar()
}

// SystemPrefersDark reports whether the OS color scheme is dark. ok is false
// when the setting cannot be read; callers should then pick a default.
func SystemPrefersDark() (dark, ok bool) {
	return systemPrefersDark()
}
