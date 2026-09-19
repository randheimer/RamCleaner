package cleaner

import (
	"sort"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
	"ramcleaner/src/sysinfo"
	"ramcleaner/src/winapi"
)

func emptyOneWorkingSet(h uintptr) (bool, windows.Errno) {
	// NOTE: recent Win11 builds removed kernel32!EmptyWorkingSet and
	// psapi!K32EmptyWorkingSet. kernel32!K32EmptyWorkingSet is the one
	// that still exists everywhere; SetProcessWorkingSetSize(-1,-1)
	// is the documented equivalent fallback.
	var ferr windows.Errno
	setErr := func(err error) {
		if ferr != 0 {
			return
		}
		if e, ok := err.(windows.Errno); ok {
			ferr = e
		} else {
			ferr = 1
		}
	}
	if winapi.ProcK32EmptyWSK32.Find() == nil {
		r1, _, err := winapi.ProcK32EmptyWSK32.Call(h)
		if r1 != 0 {
			return true, 0
		}
		setErr(err)
	}
	if winapi.ProcK32EmptyWorkingSet.Find() == nil { // psapi, older builds
		r1, _, err := winapi.ProcK32EmptyWorkingSet.Call(h)
		if r1 != 0 {
			return true, 0
		}
		setErr(err)
	}
	if winapi.ProcSetWSSize.Find() == nil {
		r1, _, err := winapi.ProcSetWSSize.Call(h, ^uintptr(0), ^uintptr(0))
		if r1 != 0 {
			return true, 0
		}
		setErr(err)
	}
	return false, ferr
}

func TrimWorkingSets() (ok, fail, skip int) {
	ok, fail, skip, _, _ = TrimWorkingSetsLive(nil)
	return ok, fail, skip
}

// TrimTick reports live progress of the trim run.
type TrimTick struct {
	Done       int
	Total      int
	Name       string
	FreedMB    uint64
	TotalFreed uint64
}

// FreedStat aggregates freed MB per exe name.
type FreedStat struct {
	Name  string
	MB    uint64
	Count int
}

func procWorkingSetMB(h uintptr) (uint64, bool) {
	var c winapi.ProcMemCounters
	c.Cb = uint32(unsafe.Sizeof(c))
	if winapi.ProcPsapiGetMemInfo.Find() == nil {
		r1, _, _ := winapi.ProcPsapiGetMemInfo.Call(h, uintptr(unsafe.Pointer(&c)), uintptr(c.Cb))
		if r1 != 0 {
			return uint64(c.WorkingSetSize) / 1024 / 1024, true
		}
	}
	if winapi.ProcK32GetMemInfo.Find() == nil {
		r1, _, _ := winapi.ProcK32GetMemInfo.Call(h, uintptr(unsafe.Pointer(&c)), uintptr(c.Cb))
		if r1 != 0 {
			return uint64(c.WorkingSetSize) / 1024 / 1024, true
		}
	}
	return 0, false
}

// TrimWorkingSetsLive trims every process working set like TrimWorkingSets
// but measures freed MB per process (before/after) and calls cb for each one.
// Returns counts (denied = access-denied subset of fail, i.e. protected
// processes) plus per-exe freed totals, biggest first.
func TrimWorkingSetsLive(cb func(TrimTick)) (ok, fail, skip, denied int, freed []FreedStat) {
	type target struct {
		pid  uint32
		name string
	}
	snap, _, _ := winapi.ProcCreateToolhelp32Snapshot.Call(uintptr(winapi.TH32CSSnapProcess), 0)
	if snap == winapi.InvalidHandleValue || snap == 0 {
		return 0, 0, 0, 0, nil
	}
	var targets []target
	var pe winapi.ProcessEntry32
	pe.Size = uint32(unsafe.Sizeof(pe))
	r1, _, _ := winapi.ProcProcess32FirstW.Call(snap, uintptr(unsafe.Pointer(&pe)))
	if r1 != 0 {
		for {
			pid := pe.ProcessID
			name := strings.ToLower(windows.UTF16ToString(pe.ExeFile[:]))
			if pid != 0 && name != "" {
				targets = append(targets, target{pid, name})
			}
			r1, _, _ := winapi.ProcProcess32NextW.Call(snap, uintptr(unsafe.Pointer(&pe)))
			if r1 == 0 {
				break
			}
		}
	}
	winapi.ProcCloseHandle.Call(snap)

	total := len(targets)
	byName := map[string]*FreedStat{}
	var order []*FreedStat
	var totalFreed uint64
	tick := func(done int, name string, f uint64) {
		if cb != nil {
			cb(TrimTick{done, total, name, f, totalFreed})
		}
	}
	for i, t := range targets {
		h, _, _ := winapi.ProcOpenProcess.Call(
			uintptr(winapi.ProcessQueryInfo|winapi.ProcessSetQuota), 0, uintptr(t.pid))
		if h == 0 {
			skip++
			tick(i+1, t.name, 0)
			continue
		}
		before, _ := procWorkingSetMB(h)
		good, errno := emptyOneWorkingSet(h)
		after, _ := procWorkingSetMB(h)
		winapi.ProcCloseHandle.Call(h)
		var f uint64
		if after < before {
			f = before - after
		}
		totalFreed += f
		if f > 0 {
			st := byName[t.name]
			if st == nil {
				st = &FreedStat{Name: t.name}
				byName[t.name] = st
				order = append(order, st)
			}
			st.MB += f
			st.Count++
		}
		if good {
			ok++
		} else {
			fail++
			if errno == 5 { // ACCESS_DENIED: PPL/system, untouchable by anyone
				denied++
			}
		}
		tick(i+1, t.name, f)
	}
	sort.Slice(order, func(i, j int) bool { return order[i].MB > order[j].MB })
	for _, st := range order {
		freed = append(freed, *st)
	}
	return ok, fail, skip, denied, freed
}

func ntMemoryListCommand(cmd int32) bool {
	r1, _, _ := winapi.ProcNtSetSystemInfo.Call(
		uintptr(winapi.SystemMemoryListInfo),
		uintptr(unsafe.Pointer(&cmd)),
		uintptr(4),
	)
	return r1 == 0
}

func PurgeStandby() bool {
	sysinfo.EnableProfilePrivilege()
	return ntMemoryListCommand(winapi.MemoryPurgeStandbyList)
}

func PurgeLowPriority() bool {
	sysinfo.EnableProfilePrivilege()
	return ntMemoryListCommand(winapi.MemoryPurgeLowPriority)
}

func FlushModified() bool {
	sysinfo.EnableProfilePrivilege()
	return ntMemoryListCommand(winapi.MemoryFlushModifiedList)
}

func EmptyWorkingSetsSystem() bool {
	sysinfo.EnableProfilePrivilege()
	return ntMemoryListCommand(winapi.MemoryEmptyWorkingSets)
}
