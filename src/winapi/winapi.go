package winapi

import "golang.org/x/sys/windows"

var (
	ModKernel32 = windows.NewLazySystemDLL("kernel32.dll")
	ModPsapi    = windows.NewLazySystemDLL("psapi.dll")
	ModNtdll    = windows.NewLazySystemDLL("ntdll.dll")
	ModShell32  = windows.NewLazySystemDLL("shell32.dll")
	ModUser32   = windows.NewLazySystemDLL("user32.dll")

	ProcGlobalMemoryStatusEx       = ModKernel32.NewProc("GlobalMemoryStatusEx")
	ProcCreateToolhelp32Snapshot   = ModKernel32.NewProc("CreateToolhelp32Snapshot")
	ProcProcess32FirstW            = ModKernel32.NewProc("Process32FirstW")
	ProcProcess32NextW             = ModKernel32.NewProc("Process32NextW")
	ProcOpenProcess                = ModKernel32.NewProc("OpenProcess")
	ProcCloseHandle                = ModKernel32.NewProc("CloseHandle")
	ProcK32EmptyWSK32              = ModKernel32.NewProc("K32EmptyWorkingSet")
	ProcSetWSSize                  = ModKernel32.NewProc("SetProcessWorkingSetSize")
	ProcK32GetMemInfo              = ModKernel32.NewProc("K32GetProcessMemoryInfo")
	ProcGetConsoleScreenBufferInfo = ModKernel32.NewProc("GetConsoleScreenBufferInfo")
	ProcGetStdHandle               = ModKernel32.NewProc("GetStdHandle")
	ProcGetConsoleMode             = ModKernel32.NewProc("GetConsoleMode")
	ProcSetConsoleMode             = ModKernel32.NewProc("SetConsoleMode")
	ProcGetConsoleWindow           = ModKernel32.NewProc("GetConsoleWindow")
	ProcSetConsoleTitleW            = ModKernel32.NewProc("SetConsoleTitleW")

	ProcK32EmptyWorkingSet = ModPsapi.NewProc("K32EmptyWorkingSet")
	ProcPsapiGetMemInfo    = ModPsapi.NewProc("GetProcessMemoryInfo")
	ProcGetPerformanceInfo = ModPsapi.NewProc("GetPerformanceInfo")

	ProcNtSetSystemInfo          = ModNtdll.NewProc("NtSetSystemInformation")
	ProcNtQuerySystemInformation = ModNtdll.NewProc("NtQuerySystemInformation")

	ProcIsUserAnAdmin   = ModShell32.NewProc("IsUserAnAdmin")
	ProcShellExecuteW   = ModShell32.NewProc("ShellExecuteW")

	ProcGetWindowLongW             = ModUser32.NewProc("GetWindowLongW")
	ProcSetWindowLongW             = ModUser32.NewProc("SetWindowLongW")
	ProcSetLayeredWindowAttributes = ModUser32.NewProc("SetLayeredWindowAttributes")
	ProcMonitorFromWindow          = ModUser32.NewProc("MonitorFromWindow")
	ProcGetMonitorInfoW            = ModUser32.NewProc("GetMonitorInfoW")
	ProcGetWindowRect              = ModUser32.NewProc("GetWindowRect")
	ProcSetWindowPos               = ModUser32.NewProc("SetWindowPos")
	ProcShowWindow                 = ModUser32.NewProc("ShowWindow")
)

const (
	TH32CSSnapProcess       = 0x00000002
	ProcessQueryInfo        = 0x0400
	ProcessVMRead           = 0x0010
	ProcessSetQuota         = 0x0100
	ProcessQueryLimited     = 0x1000
	SWShownormal            = 1
	InvalidHandleValue      = ^uintptr(0)
	StdOutputHandle         = ^uintptr(10) // -11
	EnableVTProcessing      = 0x0004
	SystemMemoryListInfo    = 80
	MemoryEmptyWorkingSets  = 2
	MemoryFlushModifiedList = 3
	MemoryPurgeStandbyList  = 4
	MemoryPurgeLowPriority  = 5

	GWLExStyle  = -20
	WSExLayered = 0x00080000
	LWAAlpha    = 0x2
)

type MemoryStatusEx struct {
	Length               uint32
	MemoryLoad           uint32
	TotalPhys            uint64
	AvailPhys            uint64
	TotalPageFile        uint64
	AvailPageFile        uint64
	TotalVirtual         uint64
	AvailVirtual         uint64
	AvailExtendedVirtual uint64
}

type ProcessEntry32 struct {
	Size            uint32
	Usage           uint32
	ProcessID       uint32
	DefaultHeapID   uintptr
	ModuleID        uint32
	Threads         uint32
	ParentProcessID uint32
	PriClassBase    int32
	Flags           uint32
	ExeFile         [260]uint16
}

type Coord struct{ X, Y int16 }
type SmallRect struct{ Left, Top, Right, Bottom int16 }
type ConsoleBufInfo struct {
	Size              Coord
	CursorPosition    Coord
	Attributes        uint16
	Window            SmallRect
	MaximumWindowSize Coord
}

type Rect struct{ Left, Top, Right, Bottom int32 }

type MonitorInfo struct {
	CbSize    uint32
	RcMonitor Rect
	RcWork    Rect
	Flags     uint32
}

type ProcMemCounters struct {
	Cb                         uint32
	PageFaultCount             uint32
	PeakWorkingSetSize         uintptr
	WorkingSetSize             uintptr
	QuotaPeakPagedPoolUsage    uintptr
	QuotaPagedPoolUsage        uintptr
	QuotaPeakNonPagedPoolUsage uintptr
	QuotaNonPagedPoolUsage     uintptr
	PagefileUsage              uintptr
	PeakPagefileUsage          uintptr
}
