//go:build windows

package giowin

import (
	"sync"
	"unsafe"

	"gioui.org/app"
	"gioui.org/io/event"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

var (
	user32                    = windows.NewLazySystemDLL("user32.dll")
	procSetWindowLongPtrW     = user32.NewProc("SetWindowLongPtrW")
	procCallWindowProcW       = user32.NewProc("CallWindowProcW")
	dwmapi                    = windows.NewLazySystemDLL("dwmapi.dll")
	procDwmSetWindowAttribute = dwmapi.NewProc("DwmSetWindowAttribute")
)

const (
	wmIMEComposition = 0x010F
	wmNCDestroy      = 0x0082
	gcsDeltaStart    = 0x0100
	gwlpWndProc      = -4

	dwmwaUseImmersiveDarkMode = 20
)

type platformState struct{}

func (w *Window) handleEvent(e event.Event) {
	ve, ok := e.(app.Win32ViewEvent)
	if !ok {
		return
	}
	w.handle = ve.HWND
	if ve.HWND == 0 {
		return
	}
	installIMEFix(ve.HWND)
	w.applyTitleBar()
}

func (w *Window) applyTitleBar() {
	if w.handle == 0 || !w.darkSet {
		return
	}
	var v int32
	if w.dark {
		v = 1
	}
	procDwmSetWindowAttribute.Call(w.handle, dwmwaUseImmersiveDarkMode, uintptr(unsafe.Pointer(&v)), unsafe.Sizeof(v))
}

// The IME fix subclasses each window once. A single callback serves all
// windows because Windows limits the number of callbacks a process can make.
var (
	imeMu       sync.Mutex
	imeOrigProc = map[uintptr]uintptr{}
	imeCallback = windows.NewCallback(imeWndProc)
)

func installIMEFix(hwnd uintptr) {
	imeMu.Lock()
	defer imeMu.Unlock()
	if _, done := imeOrigProc[hwnd]; done {
		return
	}
	idx := gwlpWndProc
	orig, _, _ := procSetWindowLongPtrW.Call(hwnd, uintptr(idx), imeCallback)
	if orig != 0 {
		imeOrigProc[hwnd] = orig
	}
}

func imeWndProc(hwnd, msg, wParam, lParam uintptr) uintptr {
	imeMu.Lock()
	orig := imeOrigProc[hwnd]
	if msg == wmNCDestroy {
		// The handle may be reused by a new window; forget this one.
		delete(imeOrigProc, hwnd)
	}
	imeMu.Unlock()
	if msg == wmIMEComposition {
		// Gio v0.10.3 treats GCS_DELTASTART as the composition start and
		// shrinks the composing region; without the flag it uses all of
		// GCS_COMPSTR, which is correct.
		lParam &^= gcsDeltaStart
	}
	r, _, _ := procCallWindowProcW.Call(orig, hwnd, msg, wParam, lParam)
	return r
}

func systemPrefersDark() (dark, ok bool) {
	k, err := registry.OpenKey(registry.CURRENT_USER,
		`Software\Microsoft\Windows\CurrentVersion\Themes\Personalize`, registry.QUERY_VALUE)
	if err != nil {
		return false, false
	}
	defer k.Close()
	v, _, err := k.GetIntegerValue("AppsUseLightTheme")
	if err != nil {
		return false, false
	}
	return v == 0, true
}
