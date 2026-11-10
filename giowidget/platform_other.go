//go:build !windows

// Status: implemented, not yet tested on real hardware.

package giowidget

import "gioui.org/app"

// Gio cannot read or set window positions on X11, Wayland or macOS, so the
// window manager places the widget; it is still borderless and can stay on
// top.
func newScreen(app.ViewEvent) screen { return nil }
func setupWindow(screen)             {}
func showWindow(screen, bool)        {}
