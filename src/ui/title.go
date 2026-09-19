package ui

import (
	"fmt"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
	"ramcleaner/src/winapi"
)

var lastFreedMB int64
var hasFreed bool

func SetTitle(s string) {
	if winapi.ProcSetConsoleTitleW.Find() != nil {
		return
	}
	p, err := windows.UTF16PtrFromString(s)
	if err != nil {
		return
	}
	winapi.ProcSetConsoleTitleW.Call(uintptr(unsafe.Pointer(p)))
}

func num(n uint64) string {
	s := fmt.Sprintf("%d", n)
	out := ""
	c := 0
	for i := len(s) - 1; i >= 0; i-- {
		out = string(s[i]) + out
		c++
		if c%3 == 0 && i != 0 {
			out = "," + out
		}
	}
	return out
}

// RefreshIdleTitle shows live mem, rotating every 3s with
// top hog and last clean result.
func RefreshIdleTitle(usedMB, totalMB uint64, loadPct uint32, topName string, topMB uint64) {
	slots := []string{
		fmt.Sprintf("ramcleaner · %s / %s MB · %d%%", num(usedMB), num(totalMB), loadPct),
	}
	if topName != "" {
		slots = append(slots, fmt.Sprintf("ramcleaner · top: %s %d MB", topName, topMB))
	}
	if hasFreed {
		slots = append(slots, fmt.Sprintf("ramcleaner · freed %+d MB", lastFreedMB))
	}
	SetTitle(slots[int(time.Now().Unix()/3)%len(slots)])
}

func CleaningTitle(pct int, exe string) {
	if len(exe) > 24 {
		exe = exe[:22] + ".."
	}
	SetTitle(fmt.Sprintf("ramcleaner · cleaning %d%% · %s", pct, exe))
}

func PhaseTitle(phase string) {
	SetTitle("ramcleaner · " + phase)
}

func ReportFreed(mb int64) {
	lastFreedMB, hasFreed = mb, true
	SetTitle(fmt.Sprintf("ramcleaner · freed %+d MB", mb))
}
