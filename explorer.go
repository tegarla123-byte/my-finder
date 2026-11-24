package main

import (
	"fmt"
	"os"
	"strings"
	"unsafe"
)

// expandWindowsEnv expands %VAR% style environment variables commonly used in Windows paths.
func expandWindowsEnv(p string) string {
	if !strings.Contains(p, "%") {
		return p
	}
	var b strings.Builder
	b.Grow(len(p))
	for i := 0; i < len(p); {
		if p[i] == '%' {
			end := strings.IndexByte(p[i+1:], '%')
			if end > 0 {
				varName := p[i+1 : i+1+end]
				if val, ok := os.LookupEnv(varName); ok {
					b.WriteString(val)
					i += end + 2
					continue
				}
			}
		}
		b.WriteByte(p[i])
		i++
	}
	return b.String()
}

func resolvePath(rawPath string) string {
	return normalizePath(expandWindowsEnv(rawPath))
}

func pathExists(rawPath string) bool {
	resolved := resolvePath(rawPath)
	if resolved == "" {
		return false
	}
	_, err := os.Stat(resolved)
	return err == nil
}

// openInNewExplorer opens a brand-new Windows File Explorer window directly at target path.
func openInNewExplorer(hwnd uintptr, rawPath string) error {
	target := resolvePath(rawPath)
	if target == "" {
		return fmt.Errorf("lokasi path kosong")
	}

	info, err := os.Stat(target)
	if err != nil {
		return fmt.Errorf("lokasi tidak ditemukan:\n%s", target)
	}

	// /n forces Explorer to open a new single-pane window directly inside the folder.
	// If the registered path is a file instead of a folder, reveal/select it in a new window.
	var params string
	if info.IsDir() {
		params = fmt.Sprintf(`/n,"%s"`, target)
	} else {
		params = fmt.Sprintf(`/n,/select,"%s"`, target)
	}

	ret, _, _ := procShellExecuteW.Call(
		hwnd,
		uintptr(unsafe.Pointer(utf16Ptr("open"))),
		uintptr(unsafe.Pointer(utf16Ptr("explorer.exe"))),
		uintptr(unsafe.Pointer(utf16Ptr(params))),
		0,
		SW_SHOWNORMAL,
	)
	if ret <= 32 {
		return fmt.Errorf("gagal membuka File Explorer (kode %d)", ret)
	}
	return nil
}
