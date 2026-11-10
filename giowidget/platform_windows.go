//go:build windows

package giowidget

import (
	"image"
	"sync"
	"unsafe"

	"gioui.org/app"
	"golang.org/x/sys/windows"
)

var (
	user32                  = windows.NewLazySystemDLL("user32.dll")
	dwmapi                  = windows.NewLazySystemDLL("dwmapi.dll")
	procGetWindowLongPtr    = user32.NewProc("GetWindowLongPtrW")
	procSetWindowLongPtr    = user32.NewProc("SetWindowLongPtrW")
	procSetWindowPos        = user32.NewProc("SetWindowPos")
	procGetWindowRect       = user32.NewProc("GetWindowRect")
	procShowWindow          = user32.NewProc("ShowWindow")
	procSetForeground       = user32.NewProc("SetForegroundWindow")
	procSystemParamsInfo    = user32.NewProc("SystemParametersInfoW")
	procEnumDisplayMonitors = user32.NewProc("EnumDisplayMonitors")
	procGetMonitorInfo      = user32.NewProc("GetMonitorInfoW")
	procGetCursorPos        = user32.NewProc("GetCursorPos")
	procDwmSetWindowAttr    = dwmapi.NewProc("DwmSetWindowAttribute")
)

const (
	gwlExStyle       = -20
	wsExToolWindow   = 0x00000080
	wsExAppWindow    = 0x00040000
	swpNoSize        = 0x0001
	swpNoZOrder      = 0x0004
	swpNoActivate    = 0x0010
	swpFrameChanged  = 0x0020
	swHide           = 0
	swShowNoActivate = 4
	spiGetWorkArea   = 0x0030
	dwmwaCornerPref  = 33
	dwmwcpRound      = 2
)

type rect struct{ Left, Top, Right, Bottom int32 }

func (r rect) image() image.Rectangle {
	return image.Rect(int(r.Left), int(r.Top), int(r.Right), int(r.Bottom))
}

type monitorInfo struct {
	CbSize  uint32
	Monitor rect
	Work    rect
	Flags   uint32
}

type winScreen struct{ hwnd uintptr }

func newScreen(e app.ViewEvent) screen {
	if v, ok := e.(app.Win32ViewEvent); ok && v.HWND != 0 {
		return &winScreen{hwnd: v.HWND}
	}
	return nil
}

// setupWindow keeps the widget out of the taskbar and Alt+Tab and asks
// Windows 11 for rounded corners. The style only takes effect while the
// window is hidden.
func setupWindow(s screen) {
	h := s.(*winScreen).hwnd
	idx := gwlExStyle
	ex, _, _ := procGetWindowLongPtr.Call(h, uintptr(idx))
	ex = ex&^wsExAppWindow | wsExToolWindow
	procShowWindow.Call(h, swHide)
	procSetWindowLongPtr.Call(h, uintptr(idx), ex)
	pref := uint32(dwmwcpRound)
	procDwmSetWindowAttr.Call(h, dwmwaCornerPref, uintptr(unsafe.Pointer(&pref)), unsafe.Sizeof(pref))
	procSetWindowPos.Call(h, 0, 0, 0, 0, 0, swpNoSize|swpNoZOrder|swpNoActivate|swpFrameChanged|0x0002 /* SWP_NOMOVE */)
}

// showWindow shows the window; with activate it also takes the focus.
func showWindow(s screen, activate bool) {
	h := s.(*winScreen).hwnd
	procShowWindow.Call(h, swShowNoActivate)
	if activate {
		procSetForeground.Call(h)
	}
}

func (s *winScreen) windowRect() (image.Rectangle, bool) {
	var r rect
	if ok, _, _ := procGetWindowRect.Call(s.hwnd, uintptr(unsafe.Pointer(&r))); ok == 0 {
		return image.Rectangle{}, false
	}
	return r.image(), true
}

func (s *winScreen) move(p image.Point) {
	procSetWindowPos.Call(s.hwnd, 0, uintptr(p.X), uintptr(p.Y), 0, 0, swpNoSize|swpNoZOrder|swpNoActivate)
}

func (s *winScreen) primaryWork() image.Rectangle {
	var r rect
	procSystemParamsInfo.Call(spiGetWorkArea, 0, uintptr(unsafe.Pointer(&r)), 0)
	return r.image()
}

func (s *winScreen) cursor() (image.Point, bool) {
	var pt struct{ X, Y int32 }
	if ok, _, _ := procGetCursorPos.Call(uintptr(unsafe.Pointer(&pt))); ok == 0 {
		return image.Point{}, false
	}
	return image.Pt(int(pt.X), int(pt.Y)), true
}

// Monitor enumeration uses one callback for the whole process, because
// Windows limits how many callbacks a program can create.
var (
	enumMu    sync.Mutex
	enumAreas []image.Rectangle
	enumCB    = windows.NewCallback(func(mon, hdc, clip, data uintptr) uintptr {
		mi := monitorInfo{CbSize: uint32(unsafe.Sizeof(monitorInfo{}))}
		if ok, _, _ := procGetMonitorInfo.Call(mon, uintptr(unsafe.Pointer(&mi))); ok != 0 {
			enumAreas = append(enumAreas, mi.Work.image())
		}
		return 1 // continue
	})
)

func (s *winScreen) workAreas() []image.Rectangle {
	enumMu.Lock()
	defer enumMu.Unlock()
	enumAreas = nil
	procEnumDisplayMonitors.Call(0, 0, enumCB, 0)
	return enumAreas
}
