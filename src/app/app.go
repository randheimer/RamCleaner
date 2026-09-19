package app

import (
	"fmt"
	"strings"
	"time"

	"ramcleaner/src/cleaner"
	"ramcleaner/src/sysinfo"
	"ramcleaner/src/ui"
)

func comma(n uint64) string {
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

func memColor(pct uint32) string {
	if pct >= 80 {
		return ui.Red
	}
	if pct >= 55 {
		return ui.Yellow
	}
	return ui.Green
}

var logoRaw = []string{
	ui.Blue + "  _____                    _____ _" + ui.Reset,
	ui.Blue + " |  __ \\                  / ____| |" + ui.Reset,
	ui.Blue + " | |__) |__ _ _ __ ___   | |    | | ___  __ _ _ __   ___ _ __" + ui.Reset,
	ui.Blue + " |  _  // _` | '_ ` _ \\  | |    | |/ _ \\/ _` | '_ \\ / _ \\ '__|" + ui.Reset,
	ui.Blue + " | | \\ \\ (_| | | | | | | | |____| |  __/ (_| | | | |  __/ |" + ui.Reset,
	ui.Blue + " |_|  \\_\\__,_|_| |_| |_|  \\_____|_|\\___|\\__,_|_| |_|\\___|_|" + ui.Reset,
}

func logoWidth() int {
	w := 0
	for _, l := range logoRaw {
		if n := ui.VisibleLen(l); n > w {
			w = n
		}
	}
	return w
}

func fullRule() string {
	return ui.Dim + strings.Repeat("─", logoWidth()) + ui.Reset
}

func mainLines(m sysinfo.MemInfo, snap sysinfo.Snapshot, msg string) []string {
	mc := memColor(m.LoadPct)
	logo := ui.AlignBlock(logoRaw)
	lines := []string{""}
	lines = append(lines, logo...)
	rule := fullRule()
	lines = append(lines,
		rule,
		"",
		fmt.Sprintf("%s%s%s / %s MB  %s%d%%%s", ui.Bold+ui.White, comma(m.UsedMB), ui.Reset+ui.Dim, comma(m.TotalMB), mc, m.LoadPct, ui.Reset),
		ui.MemBar(m.LoadPct, 28),
		fmt.Sprintf("%sfree %s MB%s", ui.Gray, comma(m.AvailMB), ui.Reset),
		"",
	)
	topHead := ui.Dim + "────────── top ─────────────" + ui.Reset
	lines = append(lines, topHead)
	lines = append(lines, topRows(snap, false)...)
	lines = append(lines,
		"",
		rule,
	)
	menu := ui.AlignBlock([]string{
		fmt.Sprintf("  %s[1]%s  clean memory", ui.Cyan, ui.Reset),
		fmt.Sprintf("  %s[2]%s  details", ui.Cyan, ui.Reset),
		fmt.Sprintf("  %s[3]%s  settings", ui.Cyan, ui.Reset),
		fmt.Sprintf("  %s[0]%s  exit", ui.Cyan, ui.Reset),
	})
	lines = append(lines, menu...)
	if msg != "" {
		lines = append(lines, "", ui.Yellow+"! "+msg+ui.Reset)
	}
	lines = append(lines, "", ui.Cyan+"select > "+ui.Reset)
	return lines
}

// clean steps, toggled in settings. [1] clean memory runs the enabled ones.
var stepTrim = true
var stepPurge = true
var stepFlush = true

func onOff(b bool) string {
	if b {
		return ui.Green + "on" + ui.Reset
	}
	return ui.Gray + "off" + ui.Reset
}

// topRows renders the grouped top list. rest=true appends the full
// breakdown (other, modified, kernel, drivers, unread note).
func topRows(snap sysinfo.Snapshot, rest bool) []string {
	if len(snap.Groups) == 0 {
		return []string{ui.Dim + "  n/a" + ui.Reset}
	}
	w := 16 // "drivers + locked" width, keeps MB column straight
	for _, g := range snap.Groups {
		if len(g.Name) > w {
			w = len(g.Name)
		}
	}
	if w > 22 {
		w = 22
	}
	trim := func(s string) string {
		if len(s) > w {
			return s[:w-2] + ".."
		}
		return s
	}
	grayRow := func(name string, mb uint64) string {
		return fmt.Sprintf("  %s%-*s%s  %s%5d MB%s",
			ui.Gray, w, trim(name), ui.Reset, ui.Gray, mb, ui.Reset)
	}
	rows := make([]string, 0, len(snap.Groups)+5)
	for _, g := range snap.Groups {
		suffix := ""
		if g.Count > 1 {
			suffix = fmt.Sprintf("  %s×%d%s", ui.Gray, g.Count, ui.Reset)
		}
		rows = append(rows, fmt.Sprintf("  %s%-*s%s  %s%5d MB%s%s",
			ui.White, w, trim(g.Name), ui.Reset, ui.Bold, g.MB, ui.Reset, suffix))
	}
	if !rest {
		return ui.AlignBlock(rows)
	}
	rows = append(rows,
		grayRow(fmt.Sprintf("other (%d)", snap.OtherCount+snap.Unread), snap.OtherMB),
	)
	if snap.HasSplit {
		rows = append(rows,
			grayRow("modified list", snap.ModifiedMB),
			grayRow("kernel pools", snap.KernelMB),
			grayRow("drivers + locked", snap.RestMB),
		)
	} else {
		rows = append(rows, grayRow("system + kernel", snap.SysMB))
	}
	if snap.Unread > 0 {
		rows = append(rows, fmt.Sprintf("  %s+%d protected, unreadable%s",
			ui.Dim, snap.Unread, ui.Reset))
	}
	return ui.AlignBlock(rows)
}

func showDetails() {
	m := sysinfo.GetMemory()
	snap := sysinfo.TakeSnapshot(6, m.UsedMB)
	mc := memColor(m.LoadPct)
	lines := []string{
		"",
		ui.White + "details" + ui.Reset,
		ui.Dim + "──────────────────────────────" + ui.Reset,
		"",
		fmt.Sprintf("%s%s%s / %s MB  %s%d%%%s", ui.Bold+ui.White, comma(m.UsedMB), ui.Reset+ui.Dim, comma(m.TotalMB), mc, m.LoadPct, ui.Reset),
		"",
	}
	lines = append(lines, topRows(snap, true)...)
	lines = append(lines, "")
	ui.ShowCentered(lines)
	ui.WaitEnter()
}

func drawMain(m sysinfo.MemInfo, snap sysinfo.Snapshot, msg string) {
	w, h := ui.Size()
	lines := mainLines(m, snap, msg)
	body := lines[:len(lines)-1]
	prompt := lines[len(lines)-1]
	top := (h - len(lines)) / 2
	if top < 0 {
		top = 0
	}
	ui.Clear()
	for i := 0; i < top; i++ {
		fmt.Println()
	}
	for _, l := range body {
		if l == "" {
			fmt.Println()
			continue
		}
		fmt.Println(ui.CenterLine(l, w))
	}
	fmt.Print(ui.CenterLine(prompt, w))
}

func resultLines(title string, before, after sysinfo.MemInfo, dur time.Duration, detail string) []string {
	freed := int64(after.AvailMB) - int64(before.AvailMB)
	freedCol := ui.Gray
	if freed > 0 {
		freedCol = ui.Green
	} else if freed < 0 {
		freedCol = ui.Red
	}
	lines := []string{
		"",
		ui.Bold + ui.White + title + ui.Reset,
		fullRule(),
		"",
		fmt.Sprintf("%s%-6s%s  %s%6s%s / %s MB  %s(%d%%)%s",
			ui.Gray, "before", ui.Reset, ui.White, comma(before.UsedMB), ui.Reset,
			comma(before.TotalMB), memColor(before.LoadPct), before.LoadPct, ui.Reset),
		"          " + ui.MemBar(before.LoadPct, 24),
		fmt.Sprintf("%s%-6s%s  %s%6s%s / %s MB  %s(%d%%)%s",
			ui.Gray, "after", ui.Reset, ui.White, comma(after.UsedMB), ui.Reset,
			comma(after.TotalMB), memColor(after.LoadPct), after.LoadPct, ui.Reset),
		"          " + ui.MemBar(after.LoadPct, 24),
		fmt.Sprintf("%s%-6s%s  %s%s%+d MB%s  %s%.1fs%s",
			ui.Gray, "freed", ui.Reset, ui.Bold, freedCol, freed, ui.Reset,
			ui.Dim, dur.Seconds(), ui.Reset),
		"",
	}
	if detail != "" {
		lines = append(lines, ui.Dim+detail+ui.Reset, "")
	}
	return lines
}

func doTrim() {
	before := sysinfo.GetMemory()
	ui.ShowCentered([]string{"", ui.Bold + ui.White + "trimming working sets" + ui.Reset, "", ui.Dim + "please wait" + ui.Reset})
	start := time.Now()
	ok, fail, skip := cleaner.TrimWorkingSets()
	after := sysinfo.GetMemory()
	ui.ShowCentered(resultLines("trim done", before, after, time.Since(start),
		fmt.Sprintf("ok: %d  fail: %d  skip: %d", ok, fail, skip)))
	ui.WaitEnter()
}

func doPurge() {
	before := sysinfo.GetMemory()
	ui.ShowCentered([]string{"", ui.Bold + ui.White + "purging standby list" + ui.Reset, "", ui.Dim + "please wait" + ui.Reset})
	start := time.Now()
	ok := cleaner.PurgeStandby()
	_ = cleaner.PurgeLowPriority()
	time.Sleep(400 * time.Millisecond)
	after := sysinfo.GetMemory()
	d := "result: ok"
	if !ok {
		d = "result: failed"
	}
	ui.ShowCentered(resultLines("purge done", before, after, time.Since(start), d))
	ui.WaitEnter()
}

func doFlush() {
	before := sysinfo.GetMemory()
	ui.ShowCentered([]string{"", ui.Bold + ui.White + "flushing modified list" + ui.Reset, "", ui.Dim + "please wait" + ui.Reset})
	start := time.Now()
	ok := cleaner.FlushModified()
	time.Sleep(400 * time.Millisecond)
	after := sysinfo.GetMemory()
	d := "result: ok"
	if !ok {
		d = "result: failed"
	}
	ui.ShowCentered(resultLines("flush done", before, after, time.Since(start), d))
	ui.WaitEnter()
}

func doCleanAll() {
	if !stepTrim && !stepPurge && !stepFlush {
		ui.ShowCentered([]string{"", ui.Yellow + "nothing enabled" + ui.Reset, "", ui.Dim + "turn steps on in settings" + ui.Reset, ""})
		ui.WaitEnter()
		return
	}
	before := sysinfo.GetMemory()
	start := time.Now()
	var steps []string
	var wins []string
	var totalFreed uint64
	lastPct := 0
	stepRow := func(name, status string) string {
		return fmt.Sprintf("  %s%-8s%s  %s", ui.Gray, name, ui.Reset, status)
	}
	doneWord := ui.Green + "done" + ui.Reset
	offWord := ui.Gray + "off" + ui.Reset
	failWord := ui.Red + "failed" + ui.Reset

	// one persistent live screen for the whole run:
	// bar + current exe + freed MB stay visible through every phase.
	w, h := ui.Size()
	top := (h - 6) / 2
	if top < 0 {
		top = 0
	}
	ui.Clear()
	for i := 0; i < top; i++ {
		fmt.Println()
	}
	fmt.Println(ui.CenterLine(ui.Bold+ui.White+"cleaning"+ui.Reset, w))
	fmt.Println(ui.CenterLine(fmt.Sprintf("%ssteps  trim %s  %sstandby %s  %smodified %s",
		ui.Dim, onOff(stepTrim), ui.Dim, onOff(stepPurge), ui.Dim, onOff(stepFlush)), w))
	fmt.Println()
	live := func(pct int, info string) {
		lastPct = pct
		ui.LiveLine(fmt.Sprintf("%s  %s%3d%%%s  %s",
			ui.MemBar(uint32(pct), 24), ui.Bold, pct, ui.Reset, info))
	}
	exeInfo := func(name string, freed uint64) string {
		if len(name) > 20 {
			name = name[:18] + ".."
		}
		return fmt.Sprintf("%s%s%s  %s+%d MB%s",
			ui.White, name, ui.Reset, ui.Gray, freed, ui.Reset)
	}
	if stepTrim {
		shownPct := -1
		ok, fail, skip, denied, freed := cleaner.TrimWorkingSetsLive(func(t cleaner.TrimTick) {
			pct := 0
			if t.Total > 0 {
				pct = t.Done * 100 / t.Total
			}
			totalFreed = t.TotalFreed
			live(pct, exeInfo(t.Name, t.TotalFreed))
			if pct != shownPct {
				shownPct = pct
				ui.CleaningTitle(pct, t.Name)
			}
		})
		trimStat := fmt.Sprintf("%s%d trimmed%s · %d skipped", ui.Green, ok, ui.Reset, skip)
		if denied > 0 {
			trimStat += fmt.Sprintf(" · %s%d protected%s", ui.Gray, denied, ui.Reset)
		}
		if otherFails := fail - denied; otherFails > 0 {
			trimStat += fmt.Sprintf(" · %s%d failed%s", ui.Red, otherFails, ui.Reset)
		}
		steps = append(steps, stepRow("trim", trimStat))
		for i, f := range freed {
			if i >= 5 || f.MB == 0 {
				break
			}
			if i == 0 {
				wins = append(wins, ui.Dim+"biggest wins:"+ui.Reset)
			}
			name := f.Name
			if len(name) > 24 {
				name = name[:22] + ".."
			}
			suffix := ""
			if f.Count > 1 {
				suffix = fmt.Sprintf("  %s×%d%s", ui.Gray, f.Count, ui.Reset)
			}
			wins = append(wins, fmt.Sprintf("  %s%s%s  %s+%d MB%s%s",
				ui.White, name, ui.Reset, ui.Green, f.MB, ui.Reset, suffix))
		}
	} else {
		steps = append(steps, stepRow("trim", offWord))
	}
	if stepPurge {
		live(lastPct, exeInfo("purge standby list", totalFreed))
		ui.PhaseTitle("purge standby list")
		_ = cleaner.EmptyWorkingSetsSystem()
		sOk := cleaner.PurgeStandby()
		_ = cleaner.PurgeLowPriority()
		if !sOk {
			steps = append(steps, stepRow("standby", failWord))
		} else {
			steps = append(steps, stepRow("standby", doneWord))
		}
	} else {
		steps = append(steps, stepRow("standby", offWord))
	}
	if stepFlush {
		live(lastPct, exeInfo("flush modified list", totalFreed))
		ui.PhaseTitle("flush modified list")
		fOk := cleaner.FlushModified()
		if !fOk {
			steps = append(steps, stepRow("modified", failWord))
		} else {
			steps = append(steps, stepRow("modified", doneWord))
		}
	} else {
		steps = append(steps, stepRow("modified", offWord))
	}
	fmt.Println()
	fmt.Println()
	time.Sleep(400 * time.Millisecond)
	after := sysinfo.GetMemory()
	ui.ReportFreed(int64(after.AvailMB) - int64(before.AvailMB))
	screen := resultLines("clean done", before, after, time.Since(start), "")
	extra := append([]string{}, steps...)
	if len(wins) > 0 {
		extra = append(extra, append([]string{""}, wins...)...)
	}
	if len(extra) > 0 {
		screen = append(screen[:len(screen)-1], append([]string{""}, extra...)...)
		screen = append(screen, "")
	}
	ui.ShowCentered(screen)
	ui.WaitEnter()
}

func settingsLines() []string {
	lines := []string{
		"",
		ui.White + "settings" + ui.Reset,
		ui.Dim + "──────────────────────────────" + ui.Reset,
		"",
		ui.Dim + "clean steps" + ui.Reset,
	}
	rows := ui.AlignBlock([]string{
		fmt.Sprintf("  %s[1]%s  trim working sets  %s", ui.Cyan, ui.Reset, onOff(stepTrim)),
		fmt.Sprintf("  %s[2]%s  purge standby list  %s", ui.Cyan, ui.Reset, onOff(stepPurge)),
		fmt.Sprintf("  %s[3]%s  flush modified list  %s", ui.Cyan, ui.Reset, onOff(stepFlush)),
		fmt.Sprintf("  %s[4]%s  transparency  %s", ui.Cyan, ui.Reset, onOff(ui.IsTransparent())),
		fmt.Sprintf("  %s[0]%s  back", ui.Cyan, ui.Reset),
	})
	lines = append(lines, rows...)
	lines = append(lines, "", ui.Cyan+"select > "+ui.Reset)
	return lines
}

func drawSettings() {
	w, h := ui.Size()
	lines := settingsLines()
	body := lines[:len(lines)-1]
	prompt := lines[len(lines)-1]
	top := (h - len(lines)) / 2
	if top < 0 {
		top = 0
	}
	ui.Clear()
	for i := 0; i < top; i++ {
		fmt.Println()
	}
	for _, l := range body {
		if l == "" {
			fmt.Println()
			continue
		}
		fmt.Println(ui.CenterLine(l, w))
	}
	fmt.Print(ui.CenterLine(prompt, w))
}

func settingsLoop() {
	ui.PhaseTitle("settings")
	for {
		drawSettings()
		switch strings.ToLower(ui.ReadLine()) {
		case "1":
			stepTrim = !stepTrim
		case "2":
			stepPurge = !stepPurge
		case "3":
			stepFlush = !stepFlush
		case "4", "t", "transparency":
			ui.ToggleTransparency()
		case "0", "q", "back", "exit", "":
			return
		}
	}
}

func Run() {
	if sysinfo.NeedsElevation() {
		_ = ui.HideConsoleWindow()
		sysinfo.RelaunchAsAdmin() // exits on success
		_ = ui.ShowConsoleWindow()
		ui.EnableVT()
		ui.ShowCentered([]string{"", ui.Yellow + "admin required" + ui.Reset, "", ui.Dim + "right click -> run as administrator" + ui.Reset, ""})
		ui.WaitEnter()
		return
	}
	ui.EnableVT()
	_ = ui.CenterWindow()
	_ = ui.MakeTransparent(215)
	sysinfo.EnableDebugPrivilege()
	sysinfo.EnableProfilePrivilege()
	msg := ""
	for {
		m := sysinfo.GetMemory()
		snap := sysinfo.TakeSnapshot(6, m.UsedMB)
		topName, topMB := "", uint64(0)
		if len(snap.Groups) > 0 {
			topName, topMB = snap.Groups[0].Name, snap.Groups[0].MB
		}
		ui.RefreshIdleTitle(m.UsedMB, m.TotalMB, m.LoadPct, topName, topMB)
		drawMain(m, snap, msg)
		msg = ""
		in := strings.ToLower(ui.ReadLine())
		switch in {
		case "1", "clean":
			doCleanAll()
		case "2", "d", "details":
			showDetails()
		case "3", "s", "settings":
			settingsLoop()
		case "r", "":
			continue
		case "0", "q", "quit", "exit":
			ui.ShowCentered([]string{"", ui.Dim + "bye" + ui.Reset, ""})
			time.Sleep(600 * time.Millisecond)
			ui.Clear()
			ui.SetTitle("ramcleaner")
			return
		default:
			msg = "unknown option: " + in
		}
	}
}
