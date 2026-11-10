//go:build !windows

// Status: implemented, not yet tested on real hardware.

package giowin

import (
	"os/exec"
	"runtime"
	"strings"

	"gioui.org/io/event"
)

type platformState struct{}

// The IME problem and the title bar setting are Windows-specific; on other
// systems the window manager draws the title bar.
func (w *Window) handleEvent(event.Event) {}
func (w *Window) applyTitleBar()          {}

func systemPrefersDark() (dark, ok bool) {
	switch runtime.GOOS {
	case "darwin":
		// Prints "Dark" in dark mode; the key does not exist in light mode.
		out, err := exec.Command("defaults", "read", "-g", "AppleInterfaceStyle").Output()
		if err != nil {
			return false, true
		}
		return strings.TrimSpace(string(out)) == "Dark", true
	case "linux", "freebsd", "openbsd", "netbsd":
		out, err := exec.Command("gsettings", "get", "org.gnome.desktop.interface", "color-scheme").Output()
		if err != nil {
			return false, false
		}
		return strings.Contains(string(out), "prefer-dark"), true
	}
	return false, false
}
