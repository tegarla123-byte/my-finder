---
name: my-finder-win32-dev
description: >-
  Use this skill when modifying, extending, debugging, or verifying the My Finder
  pure-Go Win32 desktop application (adding UI controls, dialogs, Win32/COM bindings,
  Explorer launch behavior, or running the release build and live GUI/RAM smoke test).
---

# My Finder Win32 Development & Verification Runbook

Use this runbook when making code changes to `my-finder` so the app stays zero-dependency, flicker-free, HiDPI-crisp, `go vet`-clean, and under ~8 MB private RAM.

## 1. Where Changes Belong

- **Win32 DLL procs, constants, structs, GDI/clipboard/icon helpers:** `win32.go`
- **Dark/Light colors, fonts, brushes, DPI scaling (`S(px)`):** `theme.go`
- **Main window layout, double-buffered paint, hit-testing, search, scroll, drag-drop, context menu:** `app.go`
- **Add/Edit modal dialog & keyboard loop:** `dialog.go`
- **COM `IFileOpenDialog` folder picker:** `folderpicker.go`
- **Path `%ENV%` expansion & `explorer.exe /n` launching:** `explorer.go`
- **JSON persistence (`shortcuts.json`), path normalization, filtering, reordering:** `store.go` and `store_test.go`

For deep architectural details and COM vtable indices, read [docs/dokumentasi-teknis.md](../../../docs/dokumentasi-teknis.md).

## 2. Implementation Checklist

1. **Add new Win32 APIs cleanly:** Declare lazy procs in `win32.go` using `mod*.NewProc(...)`. Check `proc.Find() == nil` for Windows 10/11 optional APIs (e.g., DWM/DPI).
2. **Avoid `go vet` `unsafeptr` traps:** Never do `(*T)(unsafe.Pointer(lParam))` or convert a `uintptr` returned from Windows into `unsafe.Pointer`. Use `procRtlMoveMemory.Call(...)` to copy between unmanaged memory and Go variables.
3. **Preserve double-buffering & clipping:** Draw onto `memDC` inside `paint(hdc)`, clip scrollable regions with `CreateRectRgn`/`SelectClipRgn`, and keep `WM_ERASEBKGND` returning `1`.
4. **Prevent GDI leaks:** Any `CreateSolidBrush`, `CreatePen`, `CreateCompatibleBitmap`, `CreateCompatibleDC`, `CreateRectRgn`, or `CreateFontW` must be paired with `DeleteObject`/`DeleteDC` (and restore previous `SelectObject` handles first).
5. **Keep `WM_PAINT` pure in-memory:** Do not call `os.Stat` or touch the filesystem during painting.
6. **Update tests & docs:** Add unit tests in `store_test.go` for any storage/path/filter logic, and update `docs/panduan-pengguna.md` or `docs/dokumentasi-teknis.md` if behavior changed.

## 3. Verify & Build

Run the bundled verification script after making edits:

```powershell
powershell -ExecutionPolicy Bypass -File .agents/skills/my-finder-win32-dev/scripts/verify.ps1
```

What [verify.ps1](./scripts/verify.ps1) validates:
1. `go vet ./...` exits with 0 warnings.
2. `go test -v ./...` passes all unit tests.
3. `go build -ldflags="-H=windowsgui -s -w" -o my-finder.exe .` succeeds and reports binary size.
4. Launches `my-finder.exe`, confirms `MyFinderMainWindow` is visible with title `"My Finder"` and non-zero size (note: `.NET Process.MainWindowTitle` can hit hidden COM/IME windows, so the script probes via `EnumWindows`).
5. Measures Task Manager RAM (`Working Set - Private`) and verifies clean shutdown (`ExitCode=0`) after posting `WM_CLOSE`.
