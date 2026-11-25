package main

import (
	"path/filepath"
	"syscall"
	"unsafe"
)

const aboutClassName = "MyFinderAboutDialog"

type aboutDialog struct {
	parent    uintptr
	hwnd      uintptr
	theme     *ThemeResources
	storePath string
	icon      uintptr
	done      bool

	rectBtnData  RECT
	rectBtnClose RECT

	hoverBtn   int // 0=none, 1=openData, 2=close
	pressedBtn int
	tracking   bool
}

var (
	activeAbout       *aboutDialog
	aboutClassRegDone bool
)

func showAboutDialog(parent uintptr, theme *ThemeResources, storePath string) {
	if !aboutClassRegDone {
		wc := WNDCLASSEXW{
			CbSize:        uint32(unsafe.Sizeof(WNDCLASSEXW{})),
			Style:         CS_HREDRAW | CS_VREDRAW,
			LpfnWndProc:   syscall.NewCallback(aboutWndProc),
			HInstance:     getModuleHandle(),
			HCursor:       loadCursor(IDC_ARROW),
			LpszClassName: utf16Ptr(aboutClassName),
		}
		procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))
		aboutClassRegDone = true
	}

	d := &aboutDialog{
		parent:    parent,
		theme:     theme,
		storePath: storePath,
		icon:      loadAppIconSize(theme.S(48)),
	}
	activeAbout = d
	defer func() {
		if d.icon != 0 {
			procDestroyIcon.Call(d.icon)
			d.icon = 0
		}
		activeAbout = nil
	}()

	w := theme.S(440)
	h := theme.S(346)

	var pr RECT
	procGetWindowRect.Call(parent, uintptr(unsafe.Pointer(&pr)))
	x := pr.Left + ((pr.Right-pr.Left)-w)/2
	y := pr.Top + ((pr.Bottom-pr.Top)-h)/2
	if x < 20 {
		x = 20
	}
	if y < 20 {
		y = 20
	}

	hwnd, _, _ := procCreateWindowExW.Call(
		WS_EX_DLGMODALFRAME,
		uintptr(unsafe.Pointer(utf16Ptr(aboutClassName))),
		uintptr(unsafe.Pointer(utf16Ptr("Tentang "+appName))),
		WS_POPUP|WS_CAPTION|WS_SYSMENU,
		uintptr(x), uintptr(y), uintptr(w), uintptr(h),
		parent,
		0,
		uintptr(getModuleHandle()),
		0,
	)
	if hwnd == 0 {
		return
	}
	d.hwnd = hwnd
	setDarkTitleBar(hwnd, theme.Palette.Dark)
	d.layout()

	procEnableWindow.Call(parent, 0)
	procShowWindow.Call(hwnd, SW_SHOW)
	procUpdateWindow.Call(hwnd)
	procSetFocus.Call(hwnd)

	var msg MSG
	for !d.done {
		ret, _, _ := procGetMessageW.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
		if int32(ret) <= 0 {
			if int32(ret) == 0 {
				procPostQuitMessage.Call(msg.WParam)
			}
			break
		}
		if msg.Message == WM_KEYDOWN && (msg.WParam == VK_ESCAPE || msg.WParam == VK_RETURN) {
			d.done = true
			continue
		}
		procTranslateMessage.Call(uintptr(unsafe.Pointer(&msg)))
		procDispatchMessageW.Call(uintptr(unsafe.Pointer(&msg)))
	}

	procEnableWindow.Call(parent, 1)
	if isWindowValid(hwnd) {
		procDestroyWindow.Call(hwnd)
	}
	procSetWindowPos.Call(parent, 0, 0, 0, 0, 0, SWP_NOZORDER|0x0001|0x0002)
	procSetFocus.Call(parent)
}

func (d *aboutDialog) layout() {
	var rc RECT
	procGetClientRect.Call(d.hwnd, uintptr(unsafe.Pointer(&rc)))
	s := d.theme.S
	w := rc.Right - rc.Left
	h := rc.Bottom - rc.Top
	pad := s(20)
	btnH := s(34)
	btnBottom := h - s(16)

	closeW := s(92)
	dataW := s(134)
	d.rectBtnClose = RECT{
		Left:   w - pad - closeW,
		Top:    btnBottom - btnH,
		Right:  w - pad,
		Bottom: btnBottom,
	}
	d.rectBtnData = RECT{
		Left:   pad,
		Top:    btnBottom - btnH,
		Right:  pad + dataW,
		Bottom: btnBottom,
	}
}

func (d *aboutDialog) hitButton(x, y int32) int {
	if ptInRect(d.rectBtnData, x, y) {
		return 1
	}
	if ptInRect(d.rectBtnClose, x, y) {
		return 2
	}
	return 0
}

func (d *aboutDialog) openDataFolder() {
	dir := filepath.Dir(d.storePath)
	if err := openInNewExplorer(d.hwnd, dir); err != nil {
		messageBox(d.hwnd, err.Error(), appName, MB_OK|MB_ICONERROR)
	}
}

func (d *aboutDialog) paint(hdc uintptr) {
	var rc RECT
	procGetClientRect.Call(d.hwnd, uintptr(unsafe.Pointer(&rc)))
	w := rc.Right - rc.Left
	h := rc.Bottom - rc.Top
	if w <= 0 || h <= 0 {
		return
	}

	memDC, _, _ := procCreateCompatibleDC.Call(hdc)
	defer procDeleteDC.Call(memDC)
	bmp, _, _ := procCreateCompatibleBitmap.Call(hdc, uintptr(w), uintptr(h))
	defer procDeleteObject.Call(bmp)
	oldBmp, _, _ := procSelectObject.Call(memDC, bmp)
	defer procSelectObject.Call(memDC, oldBmp)

	p := d.theme.Palette
	s := d.theme.S
	pad := s(20)

	procFillRect.Call(memDC, uintptr(unsafe.Pointer(&rc)), d.theme.BrushWindowBG)

	// Header: Icon + Title + Version + Tagline
	iconSize := s(48)
	iconX := pad
	iconY := s(18)
	if d.icon != 0 {
		procDrawIconEx.Call(memDC, uintptr(iconX), uintptr(iconY), d.icon, uintptr(iconSize), uintptr(iconSize), 0, 0, DI_NORMAL)
	}

	textX := iconX + iconSize + s(14)
	drawText(memDC, appName, RECT{Left: textX, Top: iconY + s(1), Right: w - pad - s(66), Bottom: iconY + s(27)}, p.TextPrimary, d.theme.FontHeading, DT_LEFT|DT_VCENTER|DT_SINGLELINE|DT_NOPREFIX)

	// Version badge pill
	verR := RECT{Left: w - pad - s(62), Top: iconY + s(4), Right: w - pad, Bottom: iconY + s(24)}
	drawRoundRect(memDC, verR, s(10), p.CardSelectedBG, p.CardSelectedBdr, true)
	drawText(memDC, appVersion, verR, p.TextPrimary, d.theme.FontSmall, DT_CENTER|DT_VCENTER|DT_SINGLELINE|DT_NOPREFIX)

	drawText(memDC, "Lightweight Windows Folder Shortcut Launcher", RECT{Left: textX, Top: iconY + s(28), Right: w - pad, Bottom: iconY + s(46)}, p.TextSecondary, d.theme.FontSub, DT_LEFT|DT_VCENTER|DT_SINGLELINE|DT_END_ELLIPSIS|DT_NOPREFIX)

	// Divider
	divY := iconY + iconSize + s(14)
	divPen, _, _ := procCreatePen.Call(PS_SOLID, 1, uintptr(p.Divider))
	oldPen, _, _ := procSelectObject.Call(memDC, divPen)
	procMoveToEx.Call(memDC, uintptr(pad), uintptr(divY), 0)
	procLineTo.Call(memDC, uintptr(w-pad), uintptr(divY))
	procSelectObject.Call(memDC, oldPen)
	procDeleteObject.Call(divPen)

	// Info rows
	rows := []struct {
		label string
		value string
		path  bool
	}{
		{"Author", appAuthor, false},
		{"Lisensi", appLicense, false},
		{"Teknologi", "Pure Go + Native Win32 API", false},
		{"Data", d.storePath, true},
	}
	rowY := divY + s(12)
	rowH := s(22)
	labelW := s(78)
	for _, row := range rows {
		drawText(memDC, row.label, RECT{Left: pad, Top: rowY, Right: pad + labelW, Bottom: rowY + rowH}, p.TextMuted, d.theme.FontSmall, DT_LEFT|DT_VCENTER|DT_SINGLELINE|DT_NOPREFIX)
		flags := uint32(DT_LEFT | DT_VCENTER | DT_SINGLELINE | DT_END_ELLIPSIS | DT_NOPREFIX)
		if row.path {
			flags = DT_LEFT | DT_VCENTER | DT_SINGLELINE | DT_PATH_ELLIPSIS | DT_NOPREFIX
		}
		drawText(memDC, row.value, RECT{Left: pad + labelW, Top: rowY, Right: w - pad, Bottom: rowY + rowH}, p.TextPrimary, d.theme.FontSub, flags)
		rowY += rowH
	}

	// License summary box
	licBox := RECT{
		Left:   pad,
		Top:    rowY + s(10),
		Right:  w - pad,
		Bottom: d.rectBtnClose.Top - s(14),
	}
	drawRoundRect(memDC, licBox, s(8), p.CardBG, p.CardBorder, true)

	inPad := s(12)
	drawText(memDC, appCopyright, RECT{Left: licBox.Left + inPad, Top: licBox.Top + s(10), Right: licBox.Right - inPad, Bottom: licBox.Top + s(28)}, p.TextPrimary, d.theme.FontSmall, DT_LEFT|DT_VCENTER|DT_SINGLELINE|DT_END_ELLIPSIS|DT_NOPREFIX)
	drawText(memDC, "Open source di bawah MIT License — bebas digunakan, dimodifikasi, dan didistribusikan ulang, disediakan apa adanya (\"AS IS\") tanpa garansi.", RECT{Left: licBox.Left + inPad, Top: licBox.Top + s(30), Right: licBox.Right - inPad, Bottom: licBox.Bottom - s(8)}, p.TextSecondary, d.theme.FontSub, DT_LEFT|DT_WORDBREAK|DT_NOPREFIX)

	// Open Data Folder button
	dataBG := p.ButtonBG
	if d.hoverBtn == 1 {
		dataBG = p.ButtonHoverBG
	}
	drawRoundRect(memDC, d.rectBtnData, s(7), dataBG, p.ButtonBorder, true)
	drawText(memDC, "Buka Folder Data", d.rectBtnData, p.TextPrimary, d.theme.FontButton, DT_CENTER|DT_VCENTER|DT_SINGLELINE|DT_NOPREFIX)

	// Close button
	closeBG := p.Accent
	if d.pressedBtn == 2 {
		closeBG = p.AccentPressed
	} else if d.hoverBtn == 2 {
		closeBG = p.AccentHover
	}
	drawRoundRect(memDC, d.rectBtnClose, s(7), closeBG, closeBG, false)
	drawText(memDC, "Tutup", d.rectBtnClose, p.AccentText, d.theme.FontButton, DT_CENTER|DT_VCENTER|DT_SINGLELINE|DT_NOPREFIX)

	procBitBlt.Call(hdc, 0, 0, uintptr(w), uintptr(h), memDC, 0, 0, SRCCOPY)
}

func aboutWndProc(hwnd uintptr, msg uint32, wParam, lParam uintptr) uintptr {
	d := activeAbout
	if d == nil {
		r, _, _ := procDefWindowProcW.Call(hwnd, uintptr(msg), wParam, lParam)
		return r
	}

	switch msg {
	case WM_SIZE:
		d.layout()
		procInvalidateRect.Call(hwnd, 0, 0)
		return 0

	case WM_ERASEBKGND:
		return 1

	case WM_PAINT:
		var ps PAINTSTRUCT
		hdc, _, _ := procBeginPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))
		d.paint(hdc)
		procEndPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))
		return 0

	case WM_MOUSEMOVE:
		if !d.tracking {
			tme := TRACKMOUSEEVENT{
				CbSize:    uint32(unsafe.Sizeof(TRACKMOUSEEVENT{})),
				DwFlags:   TME_LEAVE,
				HwndTrack: hwnd,
			}
			procTrackMouseEvent.Call(uintptr(unsafe.Pointer(&tme)))
			d.tracking = true
		}
		hit := d.hitButton(loword(lParam), hiword(lParam))
		if hit != d.hoverBtn {
			d.hoverBtn = hit
			procInvalidateRect.Call(hwnd, 0, 0)
		}
		return 0

	case WM_MOUSELEAVE:
		d.tracking = false
		if d.hoverBtn != 0 || d.pressedBtn != 0 {
			d.hoverBtn = 0
			d.pressedBtn = 0
			procInvalidateRect.Call(hwnd, 0, 0)
		}
		return 0

	case WM_SETCURSOR:
		if uloword(lParam) == HTCLIENT && d.hoverBtn != 0 {
			procSetCursor.Call(uintptr(loadCursor(IDC_HAND)))
			return 1
		}

	case WM_LBUTTONDOWN:
		hit := d.hitButton(loword(lParam), hiword(lParam))
		d.pressedBtn = hit
		if hit != 0 {
			procSetCapture.Call(hwnd)
			procInvalidateRect.Call(hwnd, 0, 0)
		}
		return 0

	case WM_LBUTTONUP:
		if d.pressedBtn != 0 {
			pressed := d.pressedBtn
			d.pressedBtn = 0
			procReleaseCapture.Call()
			hit := d.hitButton(loword(lParam), hiword(lParam))
			procInvalidateRect.Call(hwnd, 0, 0)
			if hit == pressed {
				switch hit {
				case 1:
					d.openDataFolder()
				case 2:
					d.done = true
				}
			}
		}
		return 0

	case WM_CLOSE:
		d.done = true
		return 0
	}

	r, _, _ := procDefWindowProcW.Call(hwnd, uintptr(msg), wParam, lParam)
	return r
}
