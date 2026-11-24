package main

import (
	"syscall"
	"unsafe"
)

var (
	modUser32   = syscall.NewLazyDLL("user32.dll")
	modGdi32    = syscall.NewLazyDLL("gdi32.dll")
	modKernel32 = syscall.NewLazyDLL("kernel32.dll")
	modShell32  = syscall.NewLazyDLL("shell32.dll")
	modDwmapi   = syscall.NewLazyDLL("dwmapi.dll")
	modOle32    = syscall.NewLazyDLL("ole32.dll")
	modAdvapi32 = syscall.NewLazyDLL("advapi32.dll")
	modUxTheme  = syscall.NewLazyDLL("uxtheme.dll")

	// user32
	procRegisterClassExW           = modUser32.NewProc("RegisterClassExW")
	procCreateWindowExW            = modUser32.NewProc("CreateWindowExW")
	procDefWindowProcW             = modUser32.NewProc("DefWindowProcW")
	procDestroyWindow              = modUser32.NewProc("DestroyWindow")
	procShowWindow                 = modUser32.NewProc("ShowWindow")
	procUpdateWindow               = modUser32.NewProc("UpdateWindow")
	procGetMessageW                = modUser32.NewProc("GetMessageW")
	procTranslateMessage           = modUser32.NewProc("TranslateMessage")
	procDispatchMessageW           = modUser32.NewProc("DispatchMessageW")
	procPostQuitMessage            = modUser32.NewProc("PostQuitMessage")
	procSendMessageW               = modUser32.NewProc("SendMessageW")
	procPostMessageW               = modUser32.NewProc("PostMessageW")
	procInvalidateRect             = modUser32.NewProc("InvalidateRect")
	procBeginPaint                 = modUser32.NewProc("BeginPaint")
	procEndPaint                   = modUser32.NewProc("EndPaint")
	procGetClientRect              = modUser32.NewProc("GetClientRect")
	procGetWindowRect              = modUser32.NewProc("GetWindowRect")
	procMoveWindow                 = modUser32.NewProc("MoveWindow")
	procSetWindowPos               = modUser32.NewProc("SetWindowPos")
	procLoadCursorW                = modUser32.NewProc("LoadCursorW")
	procLoadImageW                 = modUser32.NewProc("LoadImageW")
	procSetCursor                  = modUser32.NewProc("SetCursor")
	procSetFocus                   = modUser32.NewProc("SetFocus")
	procGetFocus                   = modUser32.NewProc("GetFocus")
	procEnableWindow               = modUser32.NewProc("EnableWindow")
	procSetCapture                 = modUser32.NewProc("SetCapture")
	procReleaseCapture             = modUser32.NewProc("ReleaseCapture")
	procTrackMouseEvent            = modUser32.NewProc("TrackMouseEvent")
	procGetWindowTextLengthW       = modUser32.NewProc("GetWindowTextLengthW")
	procGetWindowTextW             = modUser32.NewProc("GetWindowTextW")
	procSetWindowTextW             = modUser32.NewProc("SetWindowTextW")
	procMessageBoxW                = modUser32.NewProc("MessageBoxW")
	procCreatePopupMenu            = modUser32.NewProc("CreatePopupMenu")
	procAppendMenuW                = modUser32.NewProc("AppendMenuW")
	procTrackPopupMenu             = modUser32.NewProc("TrackPopupMenu")
	procDestroyMenu                = modUser32.NewProc("DestroyMenu")
	procClientToScreen             = modUser32.NewProc("ClientToScreen")
	procFillRect                   = modUser32.NewProc("FillRect")
	procDrawTextW                  = modUser32.NewProc("DrawTextW")
	procGetSystemMetrics           = modUser32.NewProc("GetSystemMetrics")
	procSetProcessDpiAwarenessCtx  = modUser32.NewProc("SetProcessDpiAwarenessContext")
	procGetDpiForWindow            = modUser32.NewProc("GetDpiForWindow")
	procOpenClipboard              = modUser32.NewProc("OpenClipboard")
	procEmptyClipboard             = modUser32.NewProc("EmptyClipboard")
	procSetClipboardData           = modUser32.NewProc("SetClipboardData")
	procCloseClipboard             = modUser32.NewProc("CloseClipboard")
	procGetKeyState                = modUser32.NewProc("GetKeyState")
	procSetWindowLongPtrW          = modUser32.NewProc("SetWindowLongPtrW")
	procGetWindowLongPtrW          = modUser32.NewProc("GetWindowLongPtrW")
	procCallWindowProcW            = modUser32.NewProc("CallWindowProcW")
	procCreateIconIndirect         = modUser32.NewProc("CreateIconIndirect")
	procDestroyIcon                = modUser32.NewProc("DestroyIcon")
	procSetTimer                   = modUser32.NewProc("SetTimer")
	procKillTimer                  = modUser32.NewProc("KillTimer")
	procIsWindow                   = modUser32.NewProc("IsWindow")

	// gdi32
	procCreateCompatibleDC     = modGdi32.NewProc("CreateCompatibleDC")
	procCreateCompatibleBitmap = modGdi32.NewProc("CreateCompatibleBitmap")
	procCreateDIBSection       = modGdi32.NewProc("CreateDIBSection")
	procCreateBitmap           = modGdi32.NewProc("CreateBitmap")
	procSelectObject           = modGdi32.NewProc("SelectObject")
	procDeleteObject           = modGdi32.NewProc("DeleteObject")
	procDeleteDC               = modGdi32.NewProc("DeleteDC")
	procBitBlt                 = modGdi32.NewProc("BitBlt")
	procCreateSolidBrush       = modGdi32.NewProc("CreateSolidBrush")
	procCreatePen              = modGdi32.NewProc("CreatePen")
	procCreateFontW            = modGdi32.NewProc("CreateFontW")
	procSetBkMode              = modGdi32.NewProc("SetBkMode")
	procSetBkColor             = modGdi32.NewProc("SetBkColor")
	procSetTextColor           = modGdi32.NewProc("SetTextColor")
	procRoundRect              = modGdi32.NewProc("RoundRect")
	procRectangle              = modGdi32.NewProc("Rectangle")
	procEllipse                = modGdi32.NewProc("Ellipse")
	procMoveToEx               = modGdi32.NewProc("MoveToEx")
	procLineTo                 = modGdi32.NewProc("LineTo")
	procPolygon                = modGdi32.NewProc("Polygon")
	procGetStockObject         = modGdi32.NewProc("GetStockObject")
	procCreateRectRgn          = modGdi32.NewProc("CreateRectRgn")
	procSelectClipRgn          = modGdi32.NewProc("SelectClipRgn")

	// kernel32
	procGetModuleHandleW = modKernel32.NewProc("GetModuleHandleW")
	procGlobalAlloc      = modKernel32.NewProc("GlobalAlloc")
	procGlobalLock       = modKernel32.NewProc("GlobalLock")
	procGlobalUnlock     = modKernel32.NewProc("GlobalUnlock")
	procGlobalFree       = modKernel32.NewProc("GlobalFree")
	procRtlMoveMemory    = modKernel32.NewProc("RtlMoveMemory")

	// shell32
	procShellExecuteW               = modShell32.NewProc("ShellExecuteW")
	procDragAcceptFiles             = modShell32.NewProc("DragAcceptFiles")
	procDragQueryFileW              = modShell32.NewProc("DragQueryFileW")
	procDragFinish                  = modShell32.NewProc("DragFinish")
	procSHCreateItemFromParsingName = modShell32.NewProc("SHCreateItemFromParsingName")

	// dwmapi
	procDwmSetWindowAttribute = modDwmapi.NewProc("DwmSetWindowAttribute")

	// ole32
	procCoInitializeEx   = modOle32.NewProc("CoInitializeEx")
	procCoUninitialize   = modOle32.NewProc("CoUninitialize")
	procCoCreateInstance = modOle32.NewProc("CoCreateInstance")
	procCoTaskMemFree    = modOle32.NewProc("CoTaskMemFree")

	// advapi32
	procRegOpenKeyExW    = modAdvapi32.NewProc("RegOpenKeyExW")
	procRegQueryValueExW = modAdvapi32.NewProc("RegQueryValueExW")
	procRegCloseKey      = modAdvapi32.NewProc("RegCloseKey")

	// uxtheme
	procSetWindowTheme = modUxTheme.NewProc("SetWindowTheme")
)

const (
	// Class styles
	CS_HREDRAW = 0x0002
	CS_VREDRAW = 0x0001

	// Window styles
	WS_OVERLAPPED       = 0x00000000
	WS_POPUP            = 0x80000000
	WS_CHILD            = 0x40000000
	WS_VISIBLE          = 0x10000000
	WS_CLIPSIBLINGS     = 0x04000000
	WS_CLIPCHILDREN     = 0x02000000
	WS_CAPTION          = 0x00C00000
	WS_BORDER           = 0x00800000
	WS_SYSMENU          = 0x00080000
	WS_THICKFRAME       = 0x00040000
	WS_MINIMIZEBOX      = 0x00020000
	WS_MAXIMIZEBOX      = 0x00010000
	WS_TABSTOP          = 0x00010000
	WS_OVERLAPPEDWINDOW = WS_OVERLAPPED | WS_CAPTION | WS_SYSMENU | WS_THICKFRAME | WS_MINIMIZEBOX | WS_MAXIMIZEBOX

	WS_EX_DLGMODALFRAME = 0x00000001
	WS_EX_ACCEPTFILES   = 0x00000010

	// Edit styles
	ES_LEFT        = 0x0000
	ES_AUTOHSCROLL = 0x0080

	// Messages
	WM_CREATE         = 0x0001
	WM_DESTROY        = 0x0002
	WM_SIZE           = 0x0005
	WM_SETFOCUS       = 0x0007
	WM_PAINT          = 0x000F
	WM_CLOSE          = 0x0010
	WM_ERASEBKGND     = 0x0014
	WM_SETTINGCHANGE  = 0x001A
	WM_SETCURSOR      = 0x0020
	WM_GETMINMAXINFO  = 0x0024
	WM_SETFONT        = 0x0030
	WM_SETICON        = 0x0080
	WM_KEYDOWN        = 0x0100
	WM_CHAR           = 0x0102
	WM_COMMAND        = 0x0111
	WM_TIMER          = 0x0113
	WM_CTLCOLOREDIT   = 0x0133
	WM_CTLCOLORSTATIC = 0x0138
	WM_MOUSEMOVE      = 0x0200
	WM_LBUTTONDOWN    = 0x0201
	WM_LBUTTONUP      = 0x0202
	WM_RBUTTONUP      = 0x0205
	WM_MOUSEWHEEL     = 0x020A
	WM_DROPFILES      = 0x0233
	WM_MOUSELEAVE     = 0x02A3
	WM_DPICHANGED     = 0x02E0

	EM_SETMARGINS    = 0x00D3
	EM_SETCUEBANNER  = 0x1501
	EM_SETSEL        = 0x00B1
	EC_LEFTMARGIN    = 0x0001
	EC_RIGHTMARGIN   = 0x0002
	EN_CHANGE        = 0x0300
	EN_SETFOCUS      = 0x0100
	EN_KILLFOCUS     = 0x0200

	// Virtual keys
	VK_TAB     = 0x09
	VK_RETURN  = 0x0D
	VK_CONTROL = 0x11
	VK_MENU    = 0x12 // Alt
	VK_ESCAPE  = 0x1B
	VK_PRIOR   = 0x21 // PageUp
	VK_NEXT    = 0x22 // PageDown
	VK_END     = 0x23
	VK_HOME    = 0x24
	VK_UP      = 0x26
	VK_DOWN    = 0x28
	VK_DELETE  = 0x2E

	// Cursors
	IDC_ARROW = 32512
	IDC_HAND  = 32649
	IDC_IBEAM = 32513

	// ShowWindow
	SW_HIDE       = 0
	SW_SHOWNORMAL = 1
	SW_SHOW       = 5

	// SystemMetrics
	SM_CXSCREEN = 0
	SM_CYSCREEN = 1
	SM_CXICON   = 11
	SM_CYICON   = 12
	SM_CXSMICON = 49
	SM_CYSMICON = 50

	// LoadImage
	IMAGE_ICON      = 1
	LR_DEFAULTCOLOR = 0x00000000

	// DrawText
	DT_LEFT         = 0x00000000
	DT_CENTER       = 0x00000001
	DT_RIGHT        = 0x00000002
	DT_VCENTER      = 0x00000004
	DT_SINGLELINE   = 0x00000020
	DT_NOPREFIX     = 0x00000800
	DT_PATH_ELLIPSIS = 0x00004000
	DT_END_ELLIPSIS = 0x00008000

	// GDI
	SRCCOPY          = 0x00CC0020
	TRANSPARENT      = 1
	OPAQUE           = 2
	PS_SOLID         = 0
	PS_NULL          = 5
	NULL_BRUSH       = 5
	NULL_PEN         = 8
	CLEARTYPE_QUALITY = 5
	FW_NORMAL        = 400
	FW_MEDIUM        = 500
	FW_SEMIBOLD      = 600
	FW_BOLD          = 700
	DEFAULT_CHARSET  = 1
	OUT_DEFAULT_PRECIS = 0
	CLIP_DEFAULT_PRECIS = 0
	DEFAULT_PITCH    = 0
	FF_SWISS         = 2 << 4
	DIB_RGB_COLORS   = 0
	BI_RGB           = 0

	// Icons
	ICON_SMALL = 0
	ICON_BIG   = 1

	// Window Long
	GWLP_WNDPROC = -4

	// Menu
	MF_STRING    = 0x00000000
	MF_GRAYED    = 0x00000001
	MF_SEPARATOR = 0x00000800
	TPM_LEFTALIGN   = 0x0000
	TPM_TOPALIGN    = 0x0000
	TPM_RETURNCMD   = 0x0100
	TPM_RIGHTBUTTON = 0x0002

	// MessageBox
	MB_OK              = 0x00000000
	MB_YESNO           = 0x00000004
	MB_ICONERROR       = 0x00000010
	MB_ICONQUESTION    = 0x00000020
	MB_ICONWARNING     = 0x00000030
	MB_ICONINFORMATION = 0x00000040
	IDYES              = 6

	// TrackMouseEvent
	TME_LEAVE = 0x00000002

	// Hit test
	HTCLIENT = 1

	// Clipboard
	CF_UNICODETEXT = 13
	GMEM_MOVEABLE  = 0x0002

	// DWM
	DWMWA_USE_IMMERSIVE_DARK_MODE    = 20
	DWMWA_USE_IMMERSIVE_DARK_MODE_19 = 19

	// Registry
	HKEY_CURRENT_USER = 0x80000001
	KEY_READ          = 0x20019

	// COM
	COINIT_APARTMENTTHREADED = 0x2
	COINIT_DISABLE_OLE1DDE   = 0x4
	CLSCTX_INPROC_SERVER     = 0x1

	// SetWindowPos
	SWP_NOZORDER = 0x0004
	SWP_NOACTIVATE = 0x0010
)

type WNDCLASSEXW struct {
	CbSize        uint32
	Style         uint32
	LpfnWndProc   uintptr
	CbClsExtra    int32
	CbWndExtra    int32
	HInstance     syscall.Handle
	HIcon         syscall.Handle
	HCursor       syscall.Handle
	HbrBackground syscall.Handle
	LpszMenuName  *uint16
	LpszClassName *uint16
	HIconSm       syscall.Handle
}

type POINT struct {
	X, Y int32
}

type RECT struct {
	Left, Top, Right, Bottom int32
}

type MSG struct {
	Hwnd    uintptr
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      POINT
}

type PAINTSTRUCT struct {
	Hdc         uintptr
	FErase      int32
	RcPaint     RECT
	FRestore    int32
	FIncUpdate  int32
	RgbReserved [32]byte
}

type MINMAXINFO struct {
	PtReserved     POINT
	PtMaxSize      POINT
	PtMaxPosition  POINT
	PtMinTrackSize POINT
	PtMaxTrackSize POINT
}

type TRACKMOUSEEVENT struct {
	CbSize      uint32
	DwFlags     uint32
	HwndTrack   uintptr
	DwHoverTime uint32
}

type ICONINFO struct {
	FIcon    int32
	XHotspot uint32
	YHotspot uint32
	HbmMask  uintptr
	HbmColor uintptr
}

type BITMAPINFOHEADER struct {
	BiSize          uint32
	BiWidth         int32
	BiHeight        int32
	BiPlanes        uint16
	BiBitCount      uint16
	BiCompression   uint32
	BiSizeImage     uint32
	BiXPelsPerMeter int32
	BiYPelsPerMeter int32
	BiClrUsed       uint32
	BiClrImportant  uint32
}

type BITMAPINFO struct {
	BmiHeader BITMAPINFOHEADER
	BmiColors [1]uint32
}

type GUID struct {
	Data1 uint32
	Data2 uint16
	Data3 uint16
	Data4 [8]byte
}

func rgb(r, g, b byte) uint32 {
	return uint32(r) | (uint32(g) << 8) | (uint32(b) << 16)
}

func utf16Ptr(s string) *uint16 {
	p, _ := syscall.UTF16PtrFromString(s)
	return p
}

func loword(v uintptr) int32 {
	return int32(int16(v & 0xFFFF))
}

func hiword(v uintptr) int32 {
	return int32(int16((v >> 16) & 0xFFFF))
}

func uloword(v uintptr) uint16 {
	return uint16(v & 0xFFFF)
}

func uhiword(v uintptr) uint16 {
	return uint16((v >> 16) & 0xFFFF)
}

func makeLParam(lo, hi uint16) uintptr {
	return uintptr(lo) | (uintptr(hi) << 16)
}

func ptInRect(r RECT, x, y int32) bool {
	return x >= r.Left && x < r.Right && y >= r.Top && y < r.Bottom
}

func getModuleHandle() syscall.Handle {
	h, _, _ := procGetModuleHandleW.Call(0)
	return syscall.Handle(h)
}

func enablePerMonitorDPI() {
	// DPI_AWARENESS_CONTEXT_PER_MONITOR_AWARE_V2 = (HANDLE)-4
	if procSetProcessDpiAwarenessCtx.Find() == nil {
		neg4 := ^uintptr(3)
		procSetProcessDpiAwarenessCtx.Call(neg4)
	}
}

func getWindowDPI(hwnd uintptr) int32 {
	if procGetDpiForWindow.Find() == nil && hwnd != 0 {
		dpi, _, _ := procGetDpiForWindow.Call(hwnd)
		if dpi > 0 {
			return int32(dpi)
		}
	}
	return 96
}

func setDarkTitleBar(hwnd uintptr, dark bool) {
	if procDwmSetWindowAttribute.Find() != nil || hwnd == 0 {
		return
	}
	var val int32
	if dark {
		val = 1
	}
	r, _, _ := procDwmSetWindowAttribute.Call(hwnd, DWMWA_USE_IMMERSIVE_DARK_MODE, uintptr(unsafe.Pointer(&val)), unsafe.Sizeof(val))
	if r != 0 {
		procDwmSetWindowAttribute.Call(hwnd, DWMWA_USE_IMMERSIVE_DARK_MODE_19, uintptr(unsafe.Pointer(&val)), unsafe.Sizeof(val))
	}
}

func isWindowsDarkMode() bool {
	var hKey uintptr
	subKey := utf16Ptr(`Software\Microsoft\Windows\CurrentVersion\Themes\Personalize`)
	ret, _, _ := procRegOpenKeyExW.Call(
		HKEY_CURRENT_USER,
		uintptr(unsafe.Pointer(subKey)),
		0,
		KEY_READ,
		uintptr(unsafe.Pointer(&hKey)),
	)
	if ret != 0 {
		return false
	}
	defer procRegCloseKey.Call(hKey)

	valName := utf16Ptr("AppsUseLightTheme")
	var data uint32
	dataSize := uint32(unsafe.Sizeof(data))
	var valType uint32
	ret, _, _ = procRegQueryValueExW.Call(
		hKey,
		uintptr(unsafe.Pointer(valName)),
		0,
		uintptr(unsafe.Pointer(&valType)),
		uintptr(unsafe.Pointer(&data)),
		uintptr(unsafe.Pointer(&dataSize)),
	)
	if ret != 0 {
		return false
	}
	return data == 0
}

func getWindowText(hwnd uintptr) string {
	n, _, _ := procGetWindowTextLengthW.Call(hwnd)
	if n == 0 {
		return ""
	}
	buf := make([]uint16, n+1)
	procGetWindowTextW.Call(hwnd, uintptr(unsafe.Pointer(&buf[0])), n+1)
	return syscall.UTF16ToString(buf)
}

func setWindowText(hwnd uintptr, s string) {
	procSetWindowTextW.Call(hwnd, uintptr(unsafe.Pointer(utf16Ptr(s))))
}

func setEditCueBanner(hwnd uintptr, text string) {
	// wParam=1 shows cue banner even when edit has focus
	procSendMessageW.Call(hwnd, EM_SETCUEBANNER, 1, uintptr(unsafe.Pointer(utf16Ptr(text))))
}

func messageBox(hwnd uintptr, text, caption string, flags uint32) int32 {
	r, _, _ := procMessageBoxW.Call(
		hwnd,
		uintptr(unsafe.Pointer(utf16Ptr(text))),
		uintptr(unsafe.Pointer(utf16Ptr(caption))),
		uintptr(flags),
	)
	return int32(r)
}

func copyToClipboard(hwnd uintptr, text string) bool {
	u16, err := syscall.UTF16FromString(text)
	if err != nil {
		return false
	}
	size := uintptr(len(u16) * 2)
	hMem, _, _ := procGlobalAlloc.Call(GMEM_MOVEABLE, size)
	if hMem == 0 {
		return false
	}
	ptr, _, _ := procGlobalLock.Call(hMem)
	if ptr == 0 {
		procGlobalFree.Call(hMem)
		return false
	}
	procRtlMoveMemory.Call(ptr, uintptr(unsafe.Pointer(&u16[0])), size)
	procGlobalUnlock.Call(hMem)

	ok, _, _ := procOpenClipboard.Call(hwnd)
	if ok == 0 {
		procGlobalFree.Call(hMem)
		return false
	}
	defer procCloseClipboard.Call()
	procEmptyClipboard.Call()
	r, _, _ := procSetClipboardData.Call(CF_UNICODETEXT, hMem)
	if r == 0 {
		procGlobalFree.Call(hMem)
		return false
	}
	return true
}

func isKeyDown(vk int) bool {
	r, _, _ := procGetKeyState.Call(uintptr(vk))
	return (r & 0x8000) != 0
}

func setWindowLongPtr(hwnd uintptr, idx int32, val uintptr) uintptr {
	r, _, _ := procSetWindowLongPtrW.Call(hwnd, uintptr(idx), val)
	return r
}

func createFont(dpi int32, pt int32, weight int32, face string) uintptr {
	// Negative height specifies character height in pixels
	height := -((pt * dpi) / 72)
	h, _, _ := procCreateFontW.Call(
		uintptr(int32(height)),
		0, 0, 0,
		uintptr(weight),
		0, 0, 0,
		DEFAULT_CHARSET,
		OUT_DEFAULT_PRECIS,
		CLIP_DEFAULT_PRECIS,
		CLEARTYPE_QUALITY,
		DEFAULT_PITCH|FF_SWISS,
		uintptr(unsafe.Pointer(utf16Ptr(face))),
	)
	return h
}

func drawRoundRect(hdc uintptr, r RECT, radius int32, fillColor, borderColor uint32, hasBorder bool) {
	brush, _, _ := procCreateSolidBrush.Call(uintptr(fillColor))
	defer procDeleteObject.Call(brush)
	var pen uintptr
	if hasBorder {
		pen, _, _ = procCreatePen.Call(PS_SOLID, 1, uintptr(borderColor))
		defer procDeleteObject.Call(pen)
	} else {
		pen, _, _ = procGetStockObject.Call(NULL_PEN)
	}
	oldBrush, _, _ := procSelectObject.Call(hdc, brush)
	oldPen, _, _ := procSelectObject.Call(hdc, pen)
	procRoundRect.Call(hdc, uintptr(r.Left), uintptr(r.Top), uintptr(r.Right), uintptr(r.Bottom), uintptr(radius), uintptr(radius))
	procSelectObject.Call(hdc, oldBrush)
	procSelectObject.Call(hdc, oldPen)
}

func drawText(hdc uintptr, text string, r RECT, color uint32, font uintptr, flags uint32) {
	if text == "" {
		return
	}
	u16, err := syscall.UTF16FromString(text)
	if err != nil || len(u16) <= 1 {
		return
	}
	oldFont, _, _ := procSelectObject.Call(hdc, font)
	procSetBkMode.Call(hdc, TRANSPARENT)
	procSetTextColor.Call(hdc, uintptr(color))
	rc := r
	procDrawTextW.Call(hdc, uintptr(unsafe.Pointer(&u16[0])), uintptr(len(u16)-1), uintptr(unsafe.Pointer(&rc)), uintptr(flags))
	procSelectObject.Call(hdc, oldFont)
}

// loadAppIcon loads the embedded RT_GROUP_ICON (ID=1) from the executable at the exact system icon size,
// falling back to an in-memory generated icon if no embedded resource is present.
func loadAppIcon(small bool) uintptr {
	var cx, cy uintptr
	if small {
		cx, _, _ = procGetSystemMetrics.Call(SM_CXSMICON)
		cy, _, _ = procGetSystemMetrics.Call(SM_CYSMICON)
		if cx == 0 {
			cx, cy = 16, 16
		}
	} else {
		cx, _, _ = procGetSystemMetrics.Call(SM_CXICON)
		cy, _, _ = procGetSystemMetrics.Call(SM_CYICON)
		if cx == 0 {
			cx, cy = 32, 32
		}
	}
	h, _, _ := procLoadImageW.Call(
		uintptr(getModuleHandle()),
		1, // MAKEINTRESOURCE(1)
		IMAGE_ICON,
		cx,
		cy,
		LR_DEFAULTCOLOR,
	)
	if h != 0 {
		return h
	}
	return createAppIcon(int32(cx))
}

// createAppIcon creates a crisp 32x32 golden folder icon in memory for the window & taskbar.
func createAppIcon(size int32) uintptr {
	bmi := BITMAPINFO{
		BmiHeader: BITMAPINFOHEADER{
			BiSize:        uint32(unsafe.Sizeof(BITMAPINFOHEADER{})),
			BiWidth:       size,
			BiHeight:      -size, // top-down
			BiPlanes:      1,
			BiBitCount:    32,
			BiCompression: BI_RGB,
		},
	}
	var bits unsafe.Pointer
	hColor, _, _ := procCreateDIBSection.Call(0, uintptr(unsafe.Pointer(&bmi)), DIB_RGB_COLORS, uintptr(unsafe.Pointer(&bits)), 0, 0)
	if hColor == 0 || bits == nil {
		return 0
	}
	defer procDeleteObject.Call(hColor)

	pixels := unsafe.Slice((*uint32)(bits), int(size*size))
	s := float64(size)
	// BGRA format
	const (
		colBack   = 0xFF_D99B00 // darker gold tab/back
		colFront  = 0xFF_F5C542 // bright warm gold front
		colHigh   = 0xFF_FCD96B // subtle top highlight
		colBorder = 0xFF_B57D00 // border
	)
	for y := int32(0); y < size; y++ {
		ny := float64(y) / s
		for x := int32(0); x < size; x++ {
			nx := float64(x) / s
			var c uint32
			// Tab region: top-left
			inTab := nx >= 0.08 && nx <= 0.46 && ny >= 0.16 && ny <= 0.34
			// Main back body
			inBack := nx >= 0.08 && nx <= 0.92 && ny >= 0.25 && ny <= 0.84
			// Front flap
			inFront := nx >= 0.08 && nx <= 0.92 && ny >= 0.36 && ny <= 0.84
			if inTab || inBack {
				c = colBack
				if inFront {
					c = colFront
					if ny >= 0.36 && ny <= 0.41 {
						c = colHigh
					}
				}
				// Round outer corners slightly
				if (nx < 0.13 || nx > 0.87) && (ny < 0.20 || ny > 0.80) {
					c = 0
				}
				// Bottom edge shadow
				if c != 0 && ny >= 0.80 {
					c = colBorder
				}
			}
			pixels[int(y*size+x)] = c
		}
	}

	hMask, _, _ := procCreateBitmap.Call(uintptr(size), uintptr(size), 1, 1, 0)
	if hMask == 0 {
		return 0
	}
	defer procDeleteObject.Call(hMask)

	ii := ICONINFO{
		FIcon:    1,
		HbmMask:  hMask,
		HbmColor: hColor,
	}
	icon, _, _ := procCreateIconIndirect.Call(uintptr(unsafe.Pointer(&ii)))
	return icon
}
