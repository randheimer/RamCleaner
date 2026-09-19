package ui

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"unsafe"

	"ramcleaner/src/winapi"
)

var stdin = bufio.NewReader(os.Stdin)

var transparent = false

func EnableVT() {
	h, _, _ := winapi.ProcGetStdHandle.Call(winapi.StdOutputHandle)
	if h == 0 || h == winapi.InvalidHandleValue {
		return
	}
	var mode uint32
	r1, _, _ := winapi.ProcGetConsoleMode.Call(h, uintptr(unsafe.Pointer(&mode)))
	if r1 == 0 {
		return
	}
	mode |= winapi.EnableVTProcessing
	winapi.ProcSetConsoleMode.Call(h, uintptr(mode))
}

func Clear() {
	fmt.Print("\x1b[2J\x1b[H")
}

func Size() (w, h int) {
	w, h = 80, 25
	hOut, _, _ := winapi.ProcGetStdHandle.Call(winapi.StdOutputHandle)
	if hOut == 0 || hOut == winapi.InvalidHandleValue {
		return w, h
	}
	var info winapi.ConsoleBufInfo
	r1, _, _ := winapi.ProcGetConsoleScreenBufferInfo.Call(hOut, uintptr(unsafe.Pointer(&info)))
	if r1 == 0 {
		return w, h
	}
	cw := int(info.Window.Right-info.Window.Left) + 1
	ch := int(info.Window.Bottom-info.Window.Top) + 1
	if cw >= 40 && cw <= 300 {
		w = cw
	}
	if ch >= 10 && ch <= 100 {
		h = ch
	}
	return w, h
}

const (
	Reset = "\x1b[0m"
	Bold  = "\x1b[1m"
	Dim   = "\x1b[2m"
	White = "\x1b[97m"
	Gray  = "\x1b[90m"
	Cyan  = "\x1b[36m"
	Blue  = "\x1b[94m"
	Green = "\x1b[32m"
	Yellow = "\x1b[33m"
	Red   = "\x1b[31m"
)

func VisibleLen(s string) int {
	n := 0
	inEsc := false
	for i := 0; i < len(s); {
		if !inEsc && s[i] == 0x1b && i+1 < len(s) && s[i+1] == '[' {
			inEsc = true
			i += 2
			continue
		}
		if inEsc {
			// end of CSI on a-z / A-Z
			c := s[i]
			i++
			if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') {
				inEsc = false
			}
			continue
		}
		// utf8 rune
		_, size := decodeRune(s[i:])
		if size < 1 {
			size = 1
		}
		i += size
		n++
	}
	return n
}

func decodeRune(s string) (rune, int) {
	if len(s) == 0 {
		return 0, 0
	}
	b := s[0]
	if b < 0x80 {
		return rune(b), 1
	}
	// minimal utf8 len detect, enough for width calc
	if b>>5 == 0x6 && len(s) >= 2 {
		return 0, 2
	}
	if b>>4 == 0xE && len(s) >= 3 {
		return 0, 3
	}
	if b>>3 == 0x1E && len(s) >= 4 {
		return 0, 4
	}
	return 0, 1
}

func CenterLine(s string, w int) string {
	n := VisibleLen(s)
	if n >= w {
		return s
	}
	pad := (w - n) / 2
	if pad < 0 {
		pad = 0
	}
	return strings.Repeat(" ", pad) + s
}

// AlignBlock pads rows to equal visible width so they form one
// straight column once each line gets centered.
func AlignBlock(rows []string) []string {
	max := 0
	for _, r := range rows {
		if n := VisibleLen(r); n > max {
			max = n
		}
	}
	out := make([]string, len(rows))
	for i, r := range rows {
		out[i] = r + strings.Repeat(" ", max-VisibleLen(r))
	}
	return out
}

func ShowCentered(lines []string) {
	w, h := Size()
	top := (h - len(lines)) / 2
	if top < 0 {
		top = 0
	}
	Clear()
	for i := 0; i < top; i++ {
		fmt.Println()
	}
	for _, l := range lines {
		if l == "" {
			fmt.Println()
			continue
		}
		fmt.Println(CenterLine(l, w))
	}
}

func MemBar(pct uint32, width int) string {
	if width < 10 {
		width = 10
	}
	filled := int(pct) * width / 100
	if filled < 0 {
		filled = 0
	}
	if filled > width {
		filled = width
	}
	color := Green
	if pct >= 80 {
		color = Red
	} else if pct >= 55 {
		color = Yellow
	}
	bar := color + strings.Repeat("\u2588", filled) + Gray + strings.Repeat("\u2591", width-filled) + Reset
	return Dim + "[" + Reset + bar + Dim + "]" + Reset
}

func ReadLine() string {
	s, _ := stdin.ReadString('\n')
	return strings.TrimSpace(s)
}

func WaitEnter() {
	w, _ := Size()
	fmt.Println()
	fmt.Print(CenterLine(Dim+"press enter to continue"+Reset, w))
	ReadLine()
}

// LiveLine redraws one centered line in place (progress bars).
// Print the static header first, call LiveLine per update, then fmt.Println().
func LiveLine(s string) {
	w, _ := Size()
	n := VisibleLen(s)
	pad := 0
	if n < w {
		pad = (w - n) / 2
	}
	fmt.Printf("\r%s%s\x1b[K", strings.Repeat(" ", pad), s)
}

// ---------- transparency ----------
// Makes the console window itself see-through (conhost).
// Keeps text with no background fill so terminal acrylic still shows.
// Best effort: silently does nothing on Windows Terminal / unsupported hosts.

func MakeTransparent(alpha uint8) bool {
	if winapi.ProcGetConsoleWindow.Find() != nil {
		return false
	}
	hwnd, _, _ := winapi.ProcGetConsoleWindow.Call()
	if hwnd == 0 {
		return false
	}
	if winapi.ProcGetWindowLongW.Find() != nil ||
		winapi.ProcSetWindowLongW.Find() != nil ||
		winapi.ProcSetLayeredWindowAttributes.Find() != nil {
		return false
	}
	gwl := winapi.GWLExStyle
	idx := uintptr(int32(gwl))
	ex, _, _ := winapi.ProcGetWindowLongW.Call(hwnd, idx)
	_, _, _ = winapi.ProcSetWindowLongW.Call(hwnd, idx, ex|winapi.WSExLayered)
	r1, _, _ := winapi.ProcSetLayeredWindowAttributes.Call(hwnd, 0, uintptr(alpha), winapi.LWAAlpha)
	if r1 == 0 {
		return false
	}
	transparent = true
	return true
}

func DisableTransparency() bool {
	if winapi.ProcGetConsoleWindow.Find() != nil {
		return false
	}
	hwnd, _, _ := winapi.ProcGetConsoleWindow.Call()
	if hwnd == 0 {
		return false
	}
	if winapi.ProcSetLayeredWindowAttributes.Find() != nil {
		return false
	}
	r1, _, _ := winapi.ProcSetLayeredWindowAttributes.Call(hwnd, 0, 255, winapi.LWAAlpha)
	if r1 == 0 {
		return false
	}
	transparent = false
	return true
}

func IsTransparent() bool { return transparent }

func ToggleTransparency() bool {
	if transparent {
		DisableTransparency()
		return false
	}
	if !MakeTransparent(215) {
		transparent = false
		return false
	}
	return true
}
