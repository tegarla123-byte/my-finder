# My Finder — Agent Guidelines

## Core Architecture Constraints
- **Stack:** Pure Go standard library + native Windows Win32/GDI/DWM/COM APIs via `syscall.NewLazyDLL`.
- **Zero external dependencies & zero CGO:** Keep `go.mod` free of third-party modules and never require GCC/MSVC/WebView2/Electron.
- **Target footprint:** Single portable `.exe` (~2–3 MB) and low RAM (~3–8 MB Task Manager Private Working Set).

## Critical Win32 & Go Invariants
- **Thread affinity:** Keep `runtime.LockOSThread()` on the main goroutine for Win32 message loops and COM STA (`CoInitializeEx`).
- **`go vet` safety (`unsafeptr`):** Never cast Win32 `uintptr` (`lParam`, `GlobalLock` pointers) directly to `unsafe.Pointer`. Copy structs/buffers using `procRtlMoveMemory.Call(...)` in `win32.go`.
- **GDI discipline:** Never leak GDI objects. Always `SelectObject` old handles back and `DeleteObject`/`DeleteDC` temporary objects; release cached fonts/brushes in `ThemeResources.Destroy()`.
- **Non-blocking paint:** Never perform disk/network I/O (such as `os.Stat`) inside `WM_PAINT`.
- **HiDPI & theme:** Scale all pixel metrics via `theme.S(px)`. Support both Dark and Light palettes in `theme.go` and keep native title bars synced via `setDarkTitleBar`.
- **Explorer launch:** Always use `ShellExecuteW` with raw parameters (`/n,"<path>"` for directories, `/n,/select,"<path>"` for files) in `explorer.go` rather than `os/exec` quoting.
- **Callback table safety:** Register `syscall.NewCallback` once per window class / subclass, never inside repeated loops or per-dialog instances.

## Build & Verification
- Release build: `go build -ldflags="-H=windowsgui -s -w" -o my-finder.exe .`
- Full verification (vet, unit tests, release build, live window & RAM smoke test):
  `powershell -ExecutionPolicy Bypass -File .agents/skills/my-finder-win32-dev/scripts/verify.ps1`
- Technical & user docs live in `docs/` (`docs/dokumentasi-teknis.md`, `docs/panduan-pengguna.md`). Keep them updated when adding user-facing features or changing architecture.
