package main

import (
	"strings"
	"syscall"
	"unsafe"
)

const (
	dlgClassName = "MyFinderShortcutDialog"

	idDlgEditName = 201
	idDlgEditPath = 202
)

type shortcutDialog struct {
	parent    uintptr
	hwnd      uintptr
	theme     *ThemeResources
	editName  uintptr
	editPath  uintptr
	title     string
	initName  string
	initPath  string
	resName   string
	resPath   string
	accepted  bool
	done      bool
	focusEdit uintptr

	rectNameBox   RECT
	rectPathBox   RECT
	rectBtnBrowse RECT
	rectBtnSave   RECT
	rectBtnCancel RECT

	hoverBtn   int // 0=none, 1=browse, 2=save, 3=cancel
	pressedBtn int
	tracking   bool
}

var (
	activeDialog    *shortcutDialog
	dlgClassRegDone bool
)

func showShortcutDialog(parent uintptr, theme *ThemeResources, title, initName, initPath string) (string, string, bool) {
	if !dlgClassRegDone {
		wc := WNDCLASSEXW{
			CbSize:        uint32(unsafe.Sizeof(WNDCLASSEXW{})),
			Style:         CS_HREDRAW | CS_VREDRAW,
			LpfnWndProc:   syscall.NewCallback(dialogWndProc),
			HInstance:     getModuleHandle(),
			HCursor:       loadCursor(IDC_ARROW),
			LpszClassName: utf16Ptr(dlgClassName),
		}
		procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))
		dlgClassRegDone = true
	}

	d := &shortcutDialog{
		parent:   parent,
		theme:    theme,
		title:    title,
		initName: initName,
		initPath: initPath,
	}
	activeDialog = d
	defer func() {
		activeDialog = nil
	}()

	w := theme.S(480)
	h := theme.S(256)

	// Center over parent window
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
		uintptr(unsafe.Pointer(utf16Ptr(dlgClassName))),
		uintptr(unsafe.Pointer(utf16Ptr(title))),
		WS_POPUP|WS_CAPTION|WS_SYSMENU|WS_CLIPCHILDREN,
		uintptr(x), uintptr(y), uintptr(w), uintptr(h),
		parent,
		0,
		uintptr(getModuleHandle()),
		0,
	)
	if hwnd == 0 {
		return "", "", false
	}
	d.hwnd = hwnd
	setDarkTitleBar(hwnd, theme.Palette.Dark)

	procEnableWindow.Call(parent, 0)
	procShowWindow.Call(hwnd, SW_SHOW)
	procUpdateWindow.Call(hwnd)
	procSetFocus.Call(d.editName)
	procSendMessageW.Call(d.editName, EM_SETSEL, 0, ^uintptr(0))

	var msg MSG
	for !d.done {
		ret, _, _ := procGetMessageW.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
		if int32(ret) <= 0 {
			if int32(ret) == 0 {
				procPostQuitMessage.Call(msg.WParam)
			}
			break
		}
		if msg.Message == WM_KEYDOWN {
			switch msg.WParam {
			case VK_RETURN:
				d.tryAccept()
				continue
			case VK_ESCAPE:
				d.cancel()
				continue
			case VK_TAB:
				focused, _, _ := procGetFocus.Call()
				if focused == d.editName {
					procSetFocus.Call(d.editPath)
					procSendMessageW.Call(d.editPath, EM_SETSEL, 0, ^uintptr(0))
				} else {
					procSetFocus.Call(d.editName)
					procSendMessageW.Call(d.editName, EM_SETSEL, 0, ^uintptr(0))
				}
				continue
			}
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

	return d.resName, d.resPath, d.accepted
}

func isWindowValid(hwnd uintptr) bool {
	if hwnd == 0 {
		return false
	}
	r, _, _ := procIsWindow.Call(hwnd)
	return r != 0
}

func (d *shortcutDialog) layout() {
	var rc RECT
	procGetClientRect.Call(d.hwnd, uintptr(unsafe.Pointer(&rc)))
	s := d.theme.S
	pad := s(20)
	w := rc.Right - rc.Left
	h := rc.Bottom - rc.Top

	inputH := s(34)
	browseW := s(106)
	gap := s(8)

	// Name box
	nameTop := s(40)
	d.rectNameBox = RECT{Left: pad, Top: nameTop, Right: w - pad, Bottom: nameTop + inputH}

	// Path box + Browse button
	pathTop := s(110)
	d.rectBtnBrowse = RECT{Left: w - pad - browseW, Top: pathTop, Right: w - pad, Bottom: pathTop + inputH}
	d.rectPathBox = RECT{Left: pad, Top: pathTop, Right: d.rectBtnBrowse.Left - gap, Bottom: pathTop + inputH}

	// Bottom buttons
	btnH := s(34)
	btnW := s(96)
	btnBottom := h - s(18)
	d.rectBtnSave = RECT{Left: w - pad - btnW, Top: btnBottom - btnH, Right: w - pad, Bottom: btnBottom}
	d.rectBtnCancel = RECT{Left: d.rectBtnSave.Left - gap - btnW, Top: btnBottom - btnH, Right: d.rectBtnSave.Left - gap, Bottom: btnBottom}

	editInsetX := s(10)
	editH := s(20)
	editOffY := (inputH - editH) / 2
	if d.editName != 0 {
		procMoveWindow.Call(
			d.editName,
			uintptr(d.rectNameBox.Left+editInsetX),
			uintptr(d.rectNameBox.Top+editOffY),
			uintptr(d.rectNameBox.Right-d.rectNameBox.Left-editInsetX*2),
			uintptr(editH),
			1,
		)
	}
	if d.editPath != 0 {
		procMoveWindow.Call(
			d.editPath,
			uintptr(d.rectPathBox.Left+editInsetX),
			uintptr(d.rectPathBox.Top+editOffY),
			uintptr(d.rectPathBox.Right-d.rectPathBox.Left-editInsetX*2),
			uintptr(editH),
			1,
		)
	}
}

func (d *shortcutDialog) hitButton(x, y int32) int {
	if ptInRect(d.rectBtnBrowse, x, y) {
		return 1
	}
	if ptInRect(d.rectBtnSave, x, y) {
		return 2
	}
	if ptInRect(d.rectBtnCancel, x, y) {
		return 3
	}
	return 0
}

func (d *shortcutDialog) browseFolder() {
	cur := strings.TrimSpace(getWindowText(d.editPath))
	chosen, ok, err := pickFolderDialog(d.hwnd, cur)
	if err != nil {
		messageBox(d.hwnd, "Gagal membuka pemilih folder:\n"+err.Error(), "My Finder", MB_OK|MB_ICONERROR)
		return
	}
	if !ok || chosen == "" {
		return
	}
	setWindowText(d.editPath, chosen)
	if strings.TrimSpace(getWindowText(d.editName)) == "" {
		setWindowText(d.editName, defaultNameForPath(chosen))
	}
	procSetFocus.Call(d.editName)
}

func (d *shortcutDialog) tryAccept() {
	rawPath := normalizePath(getWindowText(d.editPath))
	if rawPath == "" {
		messageBox(d.hwnd, "Silakan isi atau pilih lokasi folder terlebih dahulu.", "Validasi", MB_OK|MB_ICONWARNING)
		procSetFocus.Call(d.editPath)
		return
	}
	rawName := strings.TrimSpace(getWindowText(d.editName))
	if rawName == "" {
		rawName = defaultNameForPath(rawPath)
	}
	d.resName = rawName
	d.resPath = rawPath
	d.accepted = true
	d.done = true
}

func (d *shortcutDialog) cancel() {
	d.accepted = false
	d.done = true
}

func (d *shortcutDialog) paint(hdc uintptr) {
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
	radius := s(7)

	procFillRect.Call(memDC, uintptr(unsafe.Pointer(&rc)), d.theme.BrushWindowBG)

	pad := s(20)
	// Labels
	drawText(memDC, "Nama Shortcut", RECT{Left: pad, Top: s(16), Right: w - pad, Bottom: s(36)}, p.TextPrimary, d.theme.FontSmall, DT_LEFT|DT_VCENTER|DT_SINGLELINE|DT_NOPREFIX)
	drawText(memDC, "Lokasi Folder (atau File)", RECT{Left: pad, Top: s(86), Right: w - pad, Bottom: s(106)}, p.TextPrimary, d.theme.FontSmall, DT_LEFT|DT_VCENTER|DT_SINGLELINE|DT_NOPREFIX)

	// Input boxes
	nameBorder := p.InputBorder
	if d.focusEdit == d.editName {
		nameBorder = p.InputFocusBorder
	}
	drawRoundRect(memDC, d.rectNameBox, radius, p.InputBG, nameBorder, true)

	pathBorder := p.InputBorder
	if d.focusEdit == d.editPath {
		pathBorder = p.InputFocusBorder
	}
	drawRoundRect(memDC, d.rectPathBox, radius, p.InputBG, pathBorder, true)

	// Browse button
	browseBG := p.ButtonBG
	if d.hoverBtn == 1 {
		browseBG = p.ButtonHoverBG
	}
	drawRoundRect(memDC, d.rectBtnBrowse, radius, browseBG, p.ButtonBorder, true)
	drawText(memDC, "Pilih Folder...", d.rectBtnBrowse, p.TextPrimary, d.theme.FontButton, DT_CENTER|DT_VCENTER|DT_SINGLELINE|DT_NOPREFIX)

	// Helper note
	drawText(memDC, "Tips: kosongkan nama untuk otomatis memakai nama folder.", RECT{Left: pad, Top: d.rectPathBox.Bottom + s(8), Right: w - pad, Bottom: d.rectPathBox.Bottom + s(26)}, p.TextMuted, d.theme.FontSub, DT_LEFT|DT_VCENTER|DT_SINGLELINE|DT_NOPREFIX)

	// Cancel button
	cancelBG := p.ButtonBG
	if d.hoverBtn == 3 {
		cancelBG = p.ButtonHoverBG
	}
	drawRoundRect(memDC, d.rectBtnCancel, radius, cancelBG, p.ButtonBorder, true)
	drawText(memDC, "Batal", d.rectBtnCancel, p.TextPrimary, d.theme.FontButton, DT_CENTER|DT_VCENTER|DT_SINGLELINE|DT_NOPREFIX)

	// Save button
	saveBG := p.Accent
	if d.pressedBtn == 2 {
		saveBG = p.AccentPressed
	} else if d.hoverBtn == 2 {
		saveBG = p.AccentHover
	}
	drawRoundRect(memDC, d.rectBtnSave, radius, saveBG, saveBG, false)
	drawText(memDC, "Simpan", d.rectBtnSave, p.AccentText, d.theme.FontButton, DT_CENTER|DT_VCENTER|DT_SINGLELINE|DT_NOPREFIX)

	procBitBlt.Call(hdc, 0, 0, uintptr(w), uintptr(h), memDC, 0, 0, SRCCOPY)
}

func dialogWndProc(hwnd uintptr, msg uint32, wParam, lParam uintptr) uintptr {
	d := activeDialog
	if d == nil {
		r, _, _ := procDefWindowProcW.Call(hwnd, uintptr(msg), wParam, lParam)
		return r
	}

	switch msg {
	case WM_CREATE:
		d.hwnd = hwnd
		hInst := uintptr(getModuleHandle())
		d.editName, _, _ = procCreateWindowExW.Call(
			0,
			uintptr(unsafe.Pointer(utf16Ptr("EDIT"))),
			uintptr(unsafe.Pointer(utf16Ptr(d.initName))),
			WS_CHILD|WS_VISIBLE|WS_TABSTOP|ES_LEFT|ES_AUTOHSCROLL,
			0, 0, 100, 24,
			hwnd,
			idDlgEditName,
			hInst,
			0,
		)
		d.editPath, _, _ = procCreateWindowExW.Call(
			0,
			uintptr(unsafe.Pointer(utf16Ptr("EDIT"))),
			uintptr(unsafe.Pointer(utf16Ptr(d.initPath))),
			WS_CHILD|WS_VISIBLE|WS_TABSTOP|ES_LEFT|ES_AUTOHSCROLL,
			0, 0, 100, 24,
			hwnd,
			idDlgEditPath,
			hInst,
			0,
		)
		procSendMessageW.Call(d.editName, WM_SETFONT, d.theme.FontRegular, 1)
		procSendMessageW.Call(d.editPath, WM_SETFONT, d.theme.FontRegular, 1)
		d.layout()
		return 0

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

	case WM_CTLCOLOREDIT:
		hdc := wParam
		procSetTextColor.Call(hdc, uintptr(d.theme.Palette.TextPrimary))
		procSetBkColor.Call(hdc, uintptr(d.theme.Palette.InputBG))
		return d.theme.BrushInputBG

	case WM_COMMAND:
		code := uhiword(wParam)
		if code == EN_SETFOCUS {
			d.focusEdit = lParam
			procInvalidateRect.Call(hwnd, 0, 0)
		} else if code == EN_KILLFOCUS {
			if d.focusEdit == lParam {
				d.focusEdit = 0
				procInvalidateRect.Call(hwnd, 0, 0)
			}
		}
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
		x, y := loword(lParam), hiword(lParam)
		if ptInRect(d.rectNameBox, x, y) {
			procSetFocus.Call(d.editName)
			return 0
		}
		if ptInRect(d.rectPathBox, x, y) {
			procSetFocus.Call(d.editPath)
			return 0
		}
		hit := d.hitButton(x, y)
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
					d.browseFolder()
				case 2:
					d.tryAccept()
				case 3:
					d.cancel()
				}
			}
		}
		return 0

	case WM_CLOSE:
		d.cancel()
		return 0
	}

	r, _, _ := procDefWindowProcW.Call(hwnd, uintptr(msg), wParam, lParam)
	return r
}
