<div align="center">

# 🚀 RamCleaner

### Lightning-fast Windows RAM optimization tool

[![Platform](https://img.shields.io/badge/platform-Windows%2010%2F11-0078D4?style=for-the-badge&logo=windows)](https://www.microsoft.com/windows)
[![Language](https://img.shields.io/badge/Go-1.23+-00ADD8?style=for-the-badge&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-green?style=for-the-badge)](LICENSE)
[![Release](https://img.shields.io/github/v/release/randheimer/RamCleaner?style=for-the-badge)](https://github.com/randheimer/RamCleaner/releases)

**Free up memory instantly** • **Monitor RAM usage** • **Optimize performance**

[Download Latest Release](https://github.com/randheimer/RamCleaner/releases) • [Report Bug](https://github.com/randheimer/RamCleaner/issues) • [Request Feature](https://github.com/randheimer/RamCleaner/issues)

</div>

---

## ✨ Features

<table>
<tr>
<td width="50%">

### 📊 Real-time Monitoring
- Visual RAM usage display with color-coded indicators
- Top memory-consuming processes at a glance
- Detailed memory breakdown (kernel, drivers, system)

</td>
<td width="50%">

### 🧹 Smart Cleaning
- **Trim Working Sets** - Reduce process memory footprint
- **Purge Standby List** - Clear cached memory
- **Flush Modified Pages** - Write dirty pages to disk

</td>
</tr>
<tr>
<td width="50%">

### ⚙️ Customizable
- Enable/disable individual cleaning steps
- Configure which operations to run
- Optional transparent window mode

</td>
<td width="50%">

### ⚡ Lightweight & Fast
- Native Windows application
- Minimal resource usage
- No background services

</td>
</tr>
</table>

---

## 🎯 Quick Start

### Prerequisites
- Windows 10 or Windows 11 (64-bit)
- Administrator privileges

### Installation

#### Option 1: Download Binary (Recommended)
1. Go to [Releases](https://github.com/randheimer/RamCleaner/releases/latest)
2. Download `ramcleaner.exe`
3. Right-click → **Run as administrator**

#### Option 2: Build from Source
```bash
# Clone the repository
git clone https://github.com/randheimer/RamCleaner.git
cd RamCleaner

# Build
go build -o ramcleaner.exe .

# Or use the build script for optimized binary
build.bat
```

---

## 📖 Usage Guide

### Main Menu

When you launch RamCleaner, you'll see:

- **Header:** Application title and version
- **Memory Stats:** Current usage (e.g., 8,234 / 16,384 MB [50%])
- **Visual Bar:** Color-coded progress indicator
- **Available RAM:** Free memory display
- **Menu Options:**
  - `[1]` Clean memory
  - `[2]` Details
  - `[3]` Settings
  - `[0]` Exit

The interface uses color coding:
- 🟢 **Green** (< 55%) - Healthy
- 🟡 **Yellow** (55-79%) - Moderate
- 🔴 **Red** (≥ 80%) - High usage

### Commands

| Key | Action | Description |
|-----|--------|-------------|
| `1` | **Clean Memory** | Run enabled cleaning operations |
| `2` | **Details** | View comprehensive memory breakdown |
| `3` | **Settings** | Configure cleaning steps and appearance |
| `0` | **Exit** | Close application |

### Settings Menu

Toggle cleaning operations:
- ✅ **Trim Working Sets** - Forces processes to release unused memory pages
- ✅ **Purge Standby List** - Clears standby memory cache
- ✅ **Flush Modified List** - Writes modified pages to disk
- 🎨 **Transparency** - Toggle window transparency

---

## 🔧 How It Works

### Trim Working Sets
Forces processes to release unused pages from their working sets. Windows automatically reloads pages as needed, making this a safe operation that can free significant RAM.

**Impact:** Immediate reduction in process memory usage

### Purge Standby List
Clears the standby list containing pages removed from working sets but still cached in RAM. This memory shows as "in use" in Task Manager but is technically available.

**Impact:** Frees cached memory for new allocations

### Flush Modified List
Writes modified (dirty) pages back to disk, converting them to standby pages that can be freed.

**Impact:** Reduces active memory pressure

---

## 🏗️ Project Structure

```
RamCleaner/
├── 📄 main.go                 # Application entry point
├── 📦 go.mod                  # Go module definition
├── 🔨 build.bat              # Optimized build script
│
├── 📁 src/
│   ├── app/                  # Main application logic & UI flow
│   ├── cleaner/              # Memory cleaning operations
│   ├── sysinfo/              # System info & memory statistics
│   ├── ui/                   # Terminal UI components
│   └── winapi/               # Windows API bindings
│
├── 📁 .github/
│   └── workflows/            # GitHub Actions CI/CD
│
└── 📄 README.md              # You are here
```

---

## 🛡️ Safety & Security

<table>
<tr>
<td>

### ✅ Safe Operations
- Uses documented Windows APIs
- No system file modifications
- No registry changes
- Fully reversible operations

</td>
<td>

### 🔒 Security
- Requires admin for memory access only
- No network connections
- No data collection
- Open source & auditable

</td>
</tr>
</table>

**Note:** Windows automatically manages memory allocation. This tool helps reclaim memory faster, but Windows will reallocate as applications need it.

---

## 💻 Technical Details

### System Requirements
- **OS:** Windows 10 20H1+ or Windows 11
- **Architecture:** x64
- **Privileges:** Administrator (for process memory access)
- **RAM:** 4GB minimum, 8GB+ recommended

### Windows API Functions Used
- `EmptyWorkingSet` / `SetProcessWorkingSetSize` - Trim working sets
- `SetSystemFileCacheSize` - Purge standby list
- Memory management APIs - Flush modified pages

### Required Privileges
- `SeDebugPrivilege` - Process access for enumeration
- `SeProfileSingleProcessPrivilege` - Memory statistics

---

## 🤝 Contributing

Contributions are welcome! Here's how you can help:

1. 🍴 Fork the repository
2. 🔨 Create your feature branch (`git checkout -b feature/amazing-feature`)
3. ✅ Commit your changes (`git commit -m 'Add amazing feature'`)
4. 📤 Push to the branch (`git push origin feature/amazing-feature`)
5. 🎉 Open a Pull Request

### Development Setup
```bash
# Clone the repo
git clone https://github.com/randheimer/RamCleaner.git
cd RamCleaner

# Install dependencies
go mod download

# Build
go build -v .

# Run
./ramcleaner.exe
```

---

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## 🙏 Acknowledgments

- Built with [Go](https://go.dev/) programming language
- Uses [golang.org/x/sys/windows](https://pkg.go.dev/golang.org/x/sys/windows) for Windows API access
- Inspired by the need for simple, effective RAM management tools

---

<div align="center">

### ⭐ Star this repository if you find it useful!

Made with ❤️ for the Windows community

[⬆ Back to Top](#-ramcleaner)

</div>
