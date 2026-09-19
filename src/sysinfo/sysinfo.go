package sysinfo

import (
	"os"
	"sort"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
	"ramcleaner/src/winapi"
)

type MemInfo struct {
	TotalMB uint64
	AvailMB uint64
	UsedMB  uint64
	LoadPct uint32
}

type GroupUsage struct {
	Name  string
	MB    uint64
	Count int
}

type Snapshot struct {
	Groups      []GroupUsage
	OtherMB     uint64
	OtherCount  int
	SysMB       uint64
	ModifiedMB  uint64
	KernelMB    uint64
	RestMB      uint64
	HasSplit    bool
	TotalMB     uint64
	TotalProcs  int
	Skipped     int
	Unread      int
}

// PerfInformation mirrors Win32 PERFORMANCE_INFORMATION (64-bit).
type PerfInformation struct {
	Cb                uint32
	_                 uint32
	CommitTotal       uintptr
	CommitLimit       uintptr
	CommitPeak        uintptr
	PhysicalTotal     uintptr
	PhysicalAvailable uintptr
	SystemCache       uintptr
	KernelTotal       uintptr
	KernelPaged       uintptr
	KernelNonPaged    uintptr
	PageSize          uintptr
	HandleCount       uint32
	ProcessCount      uint32
	ThreadCount       uint32
	_                 uint32
}

func GetPerfInfo() (PerfInformation, bool) {
	var p PerfInformation
	p.Cb = uint32(unsafe.Sizeof(p))
	if winapi.ProcGetPerformanceInfo.Find() != nil {
		return p, false
	}
	r1, _, _ := winapi.ProcGetPerformanceInfo.Call(uintptr(unsafe.Pointer(&p)), uintptr(p.Cb))
	return p, r1 != 0
}

// QueryModifiedMB reads the modified page list size via
// NtQuerySystemInformation(SystemMemoryListInformation).
// Modified pages count as "used" but belong to no process.
func QueryModifiedMB(pageSize uint64) (uint64, bool) {
	if winapi.ProcNtQuerySystemInformation.Find() != nil || pageSize == 0 {
		return 0, false
	}
	for _, size := range []uintptr{256, 168, 104} {
		buf := make([]byte, size)
		r1, _, _ := winapi.ProcNtQuerySystemInformation.Call(
			uintptr(winapi.SystemMemoryListInfo),
			uintptr(unsafe.Pointer(&buf[0])),
			size, 0)
		if r1 != 0 || size < 40 {
			continue
		}
		modified := *(*uint64)(unsafe.Pointer(&buf[16])) +
			*(*uint64)(unsafe.Pointer(&buf[24]))
		return modified * pageSize / 1024 / 1024, true
	}
	return 0, false
}

func GetMemory() MemInfo {
	var m winapi.MemoryStatusEx
	m.Length = uint32(unsafe.Sizeof(m))
	r1, _, _ := winapi.ProcGlobalMemoryStatusEx.Call(uintptr(unsafe.Pointer(&m)))
	if r1 == 0 {
		return MemInfo{}
	}
	total := m.TotalPhys / 1024 / 1024
	avail := m.AvailPhys / 1024 / 1024
	var used uint64
	if total > avail {
		used = total - avail
	}
	return MemInfo{TotalMB: total, AvailMB: avail, UsedMB: used, LoadPct: m.MemoryLoad}
}

func IsAdmin() bool {
	r1, _, _ := winapi.ProcIsUserAnAdmin.Call()
	return r1 != 0
}

// NeedsElevation reports whether we should relaunch as admin.
// RAMCLEANER_NOELEVATE=1 skips it (dev / test only).
func NeedsElevation() bool {
	return os.Getenv("RAMCLEANER_NOELEVATE") == "" && !IsAdmin()
}

// RelaunchAsAdmin re-opens this exe with UAC prompt and exits.
// Returns false only if the relaunch itself failed.
func RelaunchAsAdmin() bool {
	exe, err := os.Executable()
	if err != nil {
		return false
	}
	verb, err := windows.UTF16PtrFromString("runas")
	if err != nil {
		return false
	}
	file, err := windows.UTF16PtrFromString(exe)
	if err != nil {
		return false
	}
	r1, _, _ := winapi.ProcShellExecuteW.Call(0,
		uintptr(unsafe.Pointer(verb)),
		uintptr(unsafe.Pointer(file)),
		0, 0, uintptr(winapi.SWShownormal))
	if r1 > 32 {
		os.Exit(0)
	}
	return false
}

func enablePriv(name string) {
	var tok windows.Token
	if err := windows.OpenProcessToken(windows.CurrentProcess(),
		windows.TOKEN_ADJUST_PRIVILEGES|windows.TOKEN_QUERY, &tok); err != nil {
		return
	}
	defer tok.Close()
	var luid windows.LUID
	if err := windows.LookupPrivilegeValue(nil, windows.StringToUTF16Ptr(name), &luid); err != nil {
		return
	}
	tp := windows.Tokenprivileges{
		PrivilegeCount: 1,
		Privileges: [1]windows.LUIDAndAttributes{
			{Luid: luid, Attributes: windows.SE_PRIVILEGE_ENABLED},
		},
	}
	_ = windows.AdjustTokenPrivileges(tok, false, &tp, 0, nil, nil)
}

func EnableProfilePrivilege() { enablePriv("SeProfileSingleProcessPrivilege") }
func EnableDebugPrivilege()   { enablePriv("SeDebugPrivilege") }

func getProcessWorkingSetMB(h uintptr) (uint64, bool) {
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

func openProc(pid uint32) uintptr {
	h, _, _ := winapi.ProcOpenProcess.Call(uintptr(winapi.ProcessQueryLimited), 0, uintptr(pid))
	if h != 0 {
		return h
	}
	h, _, _ = winapi.ProcOpenProcess.Call(
		uintptr(winapi.ProcessQueryInfo|winapi.ProcessVMRead), 0, uintptr(pid))
	return h
}

// TakeSnapshot enumerates every process, groups by exe name and
// reconciles against usedMB so the numbers always add up:
// listed groups + other + system/kernel = used.
func TakeSnapshot(n int, usedMB uint64) Snapshot {
	var snap Snapshot
	snapH, _, _ := winapi.ProcCreateToolhelp32Snapshot.Call(uintptr(winapi.TH32CSSnapProcess), 0)
	if snapH == winapi.InvalidHandleValue || snapH == 0 {
		snap.SysMB = usedMB
		return snap
	}
	defer winapi.ProcCloseHandle.Call(snapH)

	var pe winapi.ProcessEntry32
	pe.Size = uint32(unsafe.Sizeof(pe))
	r1, _, _ := winapi.ProcProcess32FirstW.Call(snapH, uintptr(unsafe.Pointer(&pe)))
	if r1 == 0 {
		snap.SysMB = usedMB
		return snap
	}
	byName := map[string]*GroupUsage{}
	var order []*GroupUsage
	for {
		pid := pe.ProcessID
		name := strings.ToLower(windows.UTF16ToString(pe.ExeFile[:]))
		if pid == 0 || name == "" {
			snap.Skipped++
		} else if h := openProc(pid); h == 0 {
			snap.Skipped++
			snap.Unread++
		} else {
			mb, ok := getProcessWorkingSetMB(h)
			winapi.ProcCloseHandle.Call(h)
			if !ok {
				snap.Skipped++
				snap.Unread++
			} else {
				g := byName[name]
				if g == nil {
					g = &GroupUsage{Name: name}
					byName[name] = g
					order = append(order, g)
				}
				g.MB += mb
				g.Count++
				snap.TotalMB += mb
				snap.TotalProcs++
			}
		}
		r1, _, _ := winapi.ProcProcess32NextW.Call(snapH, uintptr(unsafe.Pointer(&pe)))
		if r1 == 0 {
			break
		}
	}
	sort.Slice(order, func(i, j int) bool { return order[i].MB > order[j].MB })
	if len(order) > n {
		for _, g := range order[n:] {
			snap.OtherMB += g.MB
			snap.OtherCount += g.Count
		}
		order = order[:n]
	}
	for _, g := range order {
		snap.Groups = append(snap.Groups, *g)
	}
	if usedMB > snap.TotalMB {
		snap.SysMB = usedMB - snap.TotalMB
	}
	// Split the remainder: modified page list + kernel pools are
	// measurable, the rest is driver-locked / unreadable.
	if perf, ok := GetPerfInfo(); ok && perf.PageSize > 0 {
		// Kernel* pool values are page counts, not bytes.
		snap.KernelMB = (uint64(perf.KernelPaged) + uint64(perf.KernelNonPaged)) *
			uint64(perf.PageSize) / 1024 / 1024
		snap.HasSplit = true
		if modMB, ok := QueryModifiedMB(uint64(perf.PageSize)); ok {
			snap.ModifiedMB = modMB
		}
		rest := int64(snap.SysMB) - int64(snap.ModifiedMB) - int64(snap.KernelMB)
		if rest < 0 {
			rest = 0
		}
		snap.RestMB = uint64(rest)
	}
	return snap
}
