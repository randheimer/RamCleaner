package ui

import (
	"unsafe"

	"ramcleaner/src/winapi"
)

// CenterWindow moves the console window to the middle of its
// monitor (work area, so above the taskbar). Best effort:
// silently does nothing where it is not supported.
func CenterWindow() bool {
	if winapi.ProcGetConsoleWindow.Find() != nil ||
		winapi.ProcMonitorFromWindow.Find() != nil ||
		winapi.ProcGetMonitorInfoW.Find() != nil ||
		winapi.ProcGetWindowRect.Find() != nil ||
		winapi.ProcSetWindowPos.Find() != nil {
		return false
	}
	hwnd, _, _ := winapi.ProcGetConsoleWindow.Call()
	if hwnd == 0 {
		return false
	}
	mon, _, _ := winapi.ProcMonitorFromWindow.Call(hwnd, 2) // nearest
	if mon == 0 {
		return false
	}
	var mi winapi.MonitorInfo
	mi.CbSize = uint32(unsafe.Sizeof(mi))
	r1, _, _ := winapi.ProcGetMonitorInfoW.Call(mon, uintptr(unsafe.Pointer(&mi)))
	if r1 == 0 {
		return false
	}
	var rc winapi.Rect
	r1, _, _ = winapi.ProcGetWindowRect.Call(hwnd, uintptr(unsafe.Pointer(&rc)))
	if r1 == 0 {
		return false
	}
	ww := rc.Right - rc.Left
	wh := rc.Bottom - rc.Top
	x := mi.RcWork.Left + (mi.RcWork.Right-mi.RcWork.Left-ww)/2
	y := mi.RcWork.Top + (mi.RcWork.Bottom-mi.RcWork.Top-wh)/2
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}
	r1, _, _ = winapi.ProcSetWindowPos.Call(hwnd, 0,
		uintptr(int64(x)), uintptr(int64(y)), 0, 0, 0x1|0x4|0x10)
	return r1 != 0
}

// HideConsoleWindow makes our own console window invisible.
// Used before the UAC relaunch so no un-elevated window flashes.
func HideConsoleWindow() bool {
	return showConsole(0) // SW_HIDE
}

// ShowConsoleWindow brings our console window back
// (only if the elevated relaunch failed).
func ShowConsoleWindow() bool {
	return showConsole(9) // SW_RESTORE
}

func showConsole(cmd uintptr) bool {
	if winapi.ProcGetConsoleWindow.Find() != nil ||
		winapi.ProcShowWindow.Find() != nil {
		return false
	}
	hwnd, _, _ := winapi.ProcGetConsoleWindow.Call()
	if hwnd == 0 {
		return false
	}
	winapi.ProcShowWindow.Call(hwnd, cmd)
	return true
}
