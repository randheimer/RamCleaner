# RamCleaner

A lightweight Windows RAM optimization tool that helps free up memory by trimming working sets, purging standby lists, and flushing modified memory pages.

![Platform](https://img.shields.io/badge/platform-Windows-blue)
![Language](https://img.shields.io/badge/language-Go-00ADD8)
![License](https://img.shields.io/badge/license-MIT-green)

## Features

- **Real-time Memory Monitoring**: Visual display of current RAM usage with color-coded indicators
- **Process Working Set Trimming**: Reduces memory footprint of running processes
- **Standby List Purge**: Clears cached memory that can be reclaimed
- **Modified List Flush**: Writes modified pages back to disk
- **Top Memory Consumers**: Shows processes using the most RAM
- **Detailed Statistics**: View comprehensive memory breakdown including kernel pools and system memory
- **Customizable Cleaning Steps**: Enable/disable individual cleaning operations
- **Transparent Window Mode**: Optional semi-transparent interface

## Requirements

- Windows 10/11 (x64)
- Administrator privileges (required for memory operations)

## Installation

### Download Binary

Download the latest `ramcleaner.exe` from the [Releases](https://github.com/yourusername/RamCleaner/releases) page.

### Build from Source

Requires Go 1.23 or later:

```bash
go build -o ramcleaner.exe .
```

For optimized build with obfuscation (requires [garble](https://github.com/burrowers/garble)):

```bash
build.bat
```

## Usage

1. Run `ramcleaner.exe` as Administrator (right-click → Run as administrator)
2. The main menu shows:
   - Current memory usage with visual bar
   - Top memory-consuming processes
   - Available actions

### Main Menu Options

- **[1] Clean Memory**: Execute enabled cleaning steps
- **[2] Details**: View detailed memory breakdown
- **[3] Settings**: Configure cleaning steps and transparency
- **[0] Exit**: Close the application

### Settings

Toggle individual cleaning operations:
- **Trim Working Sets**: Reduces RAM used by processes
- **Purge Standby List**: Clears cached memory
- **Flush Modified List**: Writes modified pages to disk
- **Transparency**: Toggle window transparency

## How It Works

### Trim Working Sets
Forces processes to release unused pages from their working sets. Windows will automatically reload pages as needed.

### Purge Standby List
Clears the standby list, which contains pages removed from process working sets but still cached in RAM. This memory is technically available but shows as "in use" in Task Manager.

### Flush Modified List
Writes modified (dirty) pages back to disk, converting them to standby pages that can be freed.

## Technical Details

- Uses Windows API calls including `EmptyWorkingSet`, `SetSystemFileCacheSize`, and memory management functions
- Requires `SeDebugPrivilege` and `SeProfileSingleProcessPrivilege` for process access
- Safe to use - Windows manages memory reclamation automatically as needed

## Project Structure

```
RamCleaner/
├── main.go              # Entry point
├── go.mod               # Go module definition
├── build.bat            # Build script
└── src/
    ├── app/             # Main application logic and UI flow
    ├── cleaner/         # Memory cleaning operations
    ├── sysinfo/         # System information and memory stats
    ├── ui/              # Terminal UI components and rendering
    └── winapi/          # Windows API bindings
```

## Safety

This tool uses documented Windows API functions and does not modify system files or registry. All operations are reversible - Windows will reallocate memory as applications need it.

## License

MIT License - see LICENSE file for details

## Contributing

Contributions welcome! Please feel free to submit a Pull Request.

## Acknowledgments

Built with Go and the [golang.org/x/sys/windows](https://pkg.go.dev/golang.org/x/sys/windows) package.
