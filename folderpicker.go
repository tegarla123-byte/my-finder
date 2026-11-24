package main

import (
	"fmt"
	"syscall"
	"unsafe"
)

var (
	// CLSID_FileOpenDialog = {DC1C5A9C-E88A-4dde-A5A1-60F82A20AEF7}
	clsidFileOpenDialog = GUID{
		Data1: 0xDC1C5A9C,
		Data2: 0xE88A,
		Data3: 0x4DDE,
		Data4: [8]byte{0xA5, 0xA1, 0x60, 0xF8, 0x2A, 0x20, 0xAE, 0xF7},
	}
	// IID_IFileOpenDialog = {d57c7288-d4ad-4768-be02-9d969532d960}
	iidIFileOpenDialog = GUID{
		Data1: 0xD57C7288,
		Data2: 0xD4AD,
		Data3: 0x4768,
		Data4: [8]byte{0xBE, 0x02, 0x9D, 0x96, 0x95, 0x32, 0xD9, 0x60},
	}
	// IID_IShellItem = {43826d1e-e718-42ee-bc55-a1e261c37bfe}
	iidIShellItem = GUID{
		Data1: 0x43826D1E,
		Data2: 0xE718,
		Data3: 0x42EE,
		Data4: [8]byte{0xBC, 0x55, 0xA1, 0xE2, 0x61, 0xC3, 0x7B, 0xFE},
	}
)

const (
	fosPickFolders     = 0x00000020
	fosForceFileSystem = 0x00000040
	fosPathMustExist   = 0x00000800
	sigdnFileSysPath   = 0x80058000
	hrErrorCancelled   = 0x800704C7
)

type comObject struct {
	vtbl *[32]uintptr
}

func (o *comObject) release() {
	if o != nil {
		syscall.SyscallN(o.vtbl[2], uintptr(unsafe.Pointer(o)))
	}
}

// pickFolderDialog opens the modern Windows Explorer folder picker dialog via COM IFileOpenDialog.
func pickFolderDialog(hwndOwner uintptr, initialDir string) (string, bool, error) {
	var dlg *comObject
	hr, _, _ := procCoCreateInstance.Call(
		uintptr(unsafe.Pointer(&clsidFileOpenDialog)),
		0,
		CLSCTX_INPROC_SERVER,
		uintptr(unsafe.Pointer(&iidIFileOpenDialog)),
		uintptr(unsafe.Pointer(&dlg)),
	)
	if hr != 0 || dlg == nil {
		return "", false, fmt.Errorf("CoCreateInstance(FileOpenDialog) gagal: 0x%X", hr)
	}
	defer dlg.release()

	// GetOptions (vtable index 10)
	var opts uint32
	hr, _, _ = syscall.SyscallN(dlg.vtbl[10], uintptr(unsafe.Pointer(dlg)), uintptr(unsafe.Pointer(&opts)))
	if hr == 0 {
		opts |= fosPickFolders | fosForceFileSystem | fosPathMustExist
		// SetOptions (vtable index 9)
		syscall.SyscallN(dlg.vtbl[9], uintptr(unsafe.Pointer(dlg)), uintptr(opts))
	}

	// SetTitle (vtable index 17)
	syscall.SyscallN(dlg.vtbl[17], uintptr(unsafe.Pointer(dlg)), uintptr(unsafe.Pointer(utf16Ptr("Pilih Folder Shortcut"))))

	// Optional initial folder via SHCreateItemFromParsingName + SetFolder (vtable index 12)
	if initPath := resolvePath(initialDir); initPath != "" && pathExists(initPath) {
		var initItem *comObject
		hrInit, _, _ := procSHCreateItemFromParsingName.Call(
			uintptr(unsafe.Pointer(utf16Ptr(initPath))),
			0,
			uintptr(unsafe.Pointer(&iidIShellItem)),
			uintptr(unsafe.Pointer(&initItem)),
		)
		if hrInit == 0 && initItem != nil {
			syscall.SyscallN(dlg.vtbl[12], uintptr(unsafe.Pointer(dlg)), uintptr(unsafe.Pointer(initItem)))
			initItem.release()
		}
	}

	// Show (vtable index 3)
	hr, _, _ = syscall.SyscallN(dlg.vtbl[3], uintptr(unsafe.Pointer(dlg)), hwndOwner)
	if uint32(hr) == hrErrorCancelled {
		return "", false, nil
	}
	if hr != 0 {
		return "", false, fmt.Errorf("dialog folder gagal ditampilkan: 0x%X", hr)
	}

	// GetResult (vtable index 20)
	var item *comObject
	hr, _, _ = syscall.SyscallN(dlg.vtbl[20], uintptr(unsafe.Pointer(dlg)), uintptr(unsafe.Pointer(&item)))
	if hr != 0 || item == nil {
		return "", false, fmt.Errorf("gagal mengambil hasil folder: 0x%X", hr)
	}
	defer item.release()

	// IShellItem::GetDisplayName (vtable index 5)
	var pszPath *uint16
	hr, _, _ = syscall.SyscallN(item.vtbl[5], uintptr(unsafe.Pointer(item)), sigdnFileSysPath, uintptr(unsafe.Pointer(&pszPath)))
	if hr != 0 || pszPath == nil {
		return "", false, fmt.Errorf("gagal membaca path folder: 0x%X", hr)
	}
	defer procCoTaskMemFree.Call(uintptr(unsafe.Pointer(pszPath)))

	return utf16PtrToString(pszPath), true, nil
}

func utf16PtrToString(p *uint16) string {
	if p == nil {
		return ""
	}
	var length int
	for ptr := p; *ptr != 0; ptr = (*uint16)(unsafe.Add(unsafe.Pointer(ptr), 2)) {
		length++
		if length > 32768 {
			break
		}
	}
	return syscall.UTF16ToString(unsafe.Slice(p, length))
}
