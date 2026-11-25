package main

import (
	"fmt"
	"strings"
	"syscall"
	"unsafe"
)

const (
	mainClassName = "MyFinderMainWindow"

	idSearchEdit = 101

	// Context menu command IDs
	cmdOpenExplorer = 1001
	cmdCopyPath     = 1002
	cmdEditItem     = 1003
	cmdMoveUp       = 1004
	cmdMoveDown     = 1005
	cmdDeleteItem   = 1006

	timerToastClear = 1
)

type hitTarget struct {
	kind  int // 0=none, 1=addBtn, 2=clearSearch, 3=card, 4=cardEdit, 5=cardDelete, 6=scrollThumb, 7=scrollTrack, 8=searchBox
	index int // index into filtered slice for card hits
}

const (
	hitNone = iota
	hitAddBtn
	hitAboutBtn
	hitClearSearch
	hitCard
	hitCardEdit
	hitCardDelete
	hitScrollThumb
	hitScrollTrack
	hitSearchBox
)

type App struct {
	hwnd        uintptr
	searchEdit  uintptr
	origEditWP  uintptr
	iconBig     uintptr
	iconSmall   uintptr
	theme       *ThemeResources
	store       *Store
	filtered    []int
	query       string
	selectedIdx int // index into filtered
	scrollY     int32
	maxScrollY  int32

	searchFocused bool
	hover         hitTarget
	pressed       hitTarget
	trackingMouse bool

	draggingScroll    bool
	dragStartMouseY   int32
	dragStartScrollY  int32

	toastText string

	// Cached layout rects
	clientW        int32
	clientH        int32
	rectSearchBox  RECT
	rectClearBtn   RECT
	rectAddBtn     RECT
	rectAboutBtn   RECT
	rectListArea   RECT
	rectScrollBar  RECT
	rectScrollThumb RECT
	rectFooter     RECT
}

var (
	globalApp    *App
	cursorArrow  syscall.Handle
	cursorHand   syscall.Handle
)

func loadCursor(id uintptr) syscall.Handle {
	if id == IDC_ARROW && cursorArrow != 0 {
		return cursorArrow
	}
	if id == IDC_HAND && cursorHand != 0 {
		return cursorHand
	}
	h, _, _ := procLoadCursorW.Call(0, id)
	if id == IDC_ARROW {
		cursorArrow = syscall.Handle(h)
	} else if id == IDC_HAND {
		cursorHand = syscall.Handle(h)
	}
	return syscall.Handle(h)
}

func newApp(store *Store) *App {
	a := &App{
		store:       store,
		selectedIdx: 0,
	}
	a.refreshFilter()
	return a
}

func (a *App) Run() int {
	globalApp = a
	defer func() {
		a.cleanup()
		globalApp = nil
	}()

	a.iconBig = loadAppIcon(false)
	a.iconSmall = loadAppIcon(true)

	wc := WNDCLASSEXW{
		CbSize:        uint32(unsafe.Sizeof(WNDCLASSEXW{})),
		Style:         CS_HREDRAW | CS_VREDRAW,
		LpfnWndProc:   syscall.NewCallback(mainWndProc),
		HInstance:     getModuleHandle(),
		HIcon:         syscall.Handle(a.iconBig),
		HIconSm:       syscall.Handle(a.iconSmall),
		HCursor:       loadCursor(IDC_ARROW),
		LpszClassName: utf16Ptr(mainClassName),
	}
	atom, _, _ := procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))
	if atom == 0 {
		return 1
	}

	// Create initial window centered on primary monitor
	screenW, _, _ := procGetSystemMetrics.Call(SM_CXSCREEN)
	screenH, _, _ := procGetSystemMetrics.Call(SM_CYSCREEN)

	initW := int32(480)
	initH := int32(580)
	x := (int32(screenW) - initW) / 2
	y := (int32(screenH) - initH) / 2

	hwnd, _, _ := procCreateWindowExW.Call(
		WS_EX_ACCEPTFILES,
		uintptr(unsafe.Pointer(utf16Ptr(mainClassName))),
		uintptr(unsafe.Pointer(utf16Ptr("My Finder"))),
		WS_OVERLAPPEDWINDOW|WS_CLIPCHILDREN,
		uintptr(x), uintptr(y), uintptr(initW), uintptr(initH),
		0, 0,
		uintptr(getModuleHandle()),
		0,
	)
	if hwnd == 0 {
		return 1
	}
	a.hwnd = hwnd

	// Resize to DPI-aware dimensions if monitor DPI != 96
	dpi := getWindowDPI(hwnd)
	if dpi != 96 {
		scaledW := (initW * dpi) / 96
		scaledH := (initH * dpi) / 96
		sx := (int32(screenW) - scaledW) / 2
		sy := (int32(screenH) - scaledH) / 2
		procMoveWindow.Call(hwnd, uintptr(sx), uintptr(sy), uintptr(scaledW), uintptr(scaledH), 0)
	}

	procShowWindow.Call(hwnd, SW_SHOWNORMAL)
	procUpdateWindow.Call(hwnd)
	procSetFocus.Call(a.searchEdit)

	var msg MSG
	for {
		ret, _, _ := procGetMessageW.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
		if int32(ret) <= 0 {
			break
		}
		procTranslateMessage.Call(uintptr(unsafe.Pointer(&msg)))
		procDispatchMessageW.Call(uintptr(unsafe.Pointer(&msg)))
	}
	return int(msg.WParam)
}

func (a *App) cleanup() {
	if a.theme != nil {
		a.theme.Destroy()
		a.theme = nil
	}
	if a.iconBig != 0 {
		procDestroyIcon.Call(a.iconBig)
		a.iconBig = 0
	}
	if a.iconSmall != 0 {
		procDestroyIcon.Call(a.iconSmall)
		a.iconSmall = 0
	}
}

func (a *App) applyTheme(dark bool, dpi int32) {
	if a.theme != nil {
		a.theme.Destroy()
	}
	a.theme = newThemeResources(dark, dpi)
	setDarkTitleBar(a.hwnd, dark)
	if a.searchEdit != 0 {
		procSendMessageW.Call(a.searchEdit, WM_SETFONT, a.theme.FontRegular, 1)
	}
	a.layout()
	procInvalidateRect.Call(a.hwnd, 0, 0)
}

func (a *App) refreshFilter() {
	a.filtered = a.store.Filter(a.query)
	if len(a.filtered) == 0 {
		a.selectedIdx = -1
	} else {
		if a.selectedIdx < 0 {
			a.selectedIdx = 0
		}
		if a.selectedIdx >= len(a.filtered) {
			a.selectedIdx = len(a.filtered) - 1
		}
	}
	a.updateScrollBounds()
}

func (a *App) cardHeight() int32 {
	return a.theme.S(58)
}

func (a *App) cardGap() int32 {
	return a.theme.S(7)
}

func (a *App) contentHeight() int32 {
	n := int32(len(a.filtered))
	if n == 0 {
		return 0
	}
	s := a.theme.S
	return s(6) + n*a.cardHeight() + (n-1)*a.cardGap() + s(12)
}

func (a *App) updateScrollBounds() {
	if a.theme == nil {
		return
	}
	viewH := a.rectListArea.Bottom - a.rectListArea.Top
	if viewH < 0 {
		viewH = 0
	}
	totalH := a.contentHeight()
	if totalH > viewH {
		a.maxScrollY = totalH - viewH
	} else {
		a.maxScrollY = 0
	}
	if a.scrollY > a.maxScrollY {
		a.scrollY = a.maxScrollY
	}
	if a.scrollY < 0 {
		a.scrollY = 0
	}
	a.computeScrollbarThumb()
}

func (a *App) computeScrollbarThumb() {
	if a.maxScrollY <= 0 {
		a.rectScrollThumb = RECT{}
		return
	}
	track := a.rectScrollBar
	trackH := track.Bottom - track.Top
	viewH := a.rectListArea.Bottom - a.rectListArea.Top
	totalH := a.contentHeight()
	if trackH <= 0 || totalH <= 0 {
		a.rectScrollThumb = RECT{}
		return
	}
	thumbH := (viewH * trackH) / totalH
	minThumb := a.theme.S(28)
	if thumbH < minThumb {
		thumbH = minThumb
	}
	if thumbH > trackH {
		thumbH = trackH
	}
	avail := trackH - thumbH
	top := track.Top
	if avail > 0 && a.maxScrollY > 0 {
		top += (a.scrollY * avail) / a.maxScrollY
	}
	a.rectScrollThumb = RECT{
		Left:   track.Left,
		Top:    top,
		Right:  track.Right,
		Bottom: top + thumbH,
	}
}

func (a *App) layout() {
	if a.hwnd == 0 || a.theme == nil {
		return
	}
	var rc RECT
	procGetClientRect.Call(a.hwnd, uintptr(unsafe.Pointer(&rc)))
	a.clientW = rc.Right - rc.Left
	a.clientH = rc.Bottom - rc.Top
	s := a.theme.S

	pad := s(14)
	topY := s(12)
	barH := s(36)
	aboutW := barH
	addW := s(92)
	gap := s(8)
	btnGap := s(6)

	a.rectAboutBtn = RECT{
		Left:   a.clientW - pad - aboutW,
		Top:    topY,
		Right:  a.clientW - pad,
		Bottom: topY + barH,
	}
	a.rectAddBtn = RECT{
		Left:   a.rectAboutBtn.Left - btnGap - addW,
		Top:    topY,
		Right:  a.rectAboutBtn.Left - btnGap,
		Bottom: topY + barH,
	}
	a.rectSearchBox = RECT{
		Left:   pad,
		Top:    topY,
		Right:  a.rectAddBtn.Left - gap,
		Bottom: topY + barH,
	}

	clearSize := s(22)
	a.rectClearBtn = RECT{
		Left:   a.rectSearchBox.Right - s(6) - clearSize,
		Top:    a.rectSearchBox.Top + (barH-clearSize)/2,
		Right:  a.rectSearchBox.Right - s(6),
		Bottom: a.rectSearchBox.Top + (barH+clearSize)/2,
	}

	footerH := s(28)
	a.rectFooter = RECT{
		Left:   0,
		Top:    a.clientH - footerH,
		Right:  a.clientW,
		Bottom: a.clientH,
	}

	a.rectListArea = RECT{
		Left:   0,
		Top:    topY + barH + s(8),
		Right:  a.clientW,
		Bottom: a.rectFooter.Top,
	}

	scrollW := s(7)
	a.rectScrollBar = RECT{
		Left:   a.clientW - s(4) - scrollW,
		Top:    a.rectListArea.Top + s(4),
		Right:  a.clientW - s(4),
		Bottom: a.rectListArea.Bottom - s(4),
	}

	if a.searchEdit != 0 {
		iconPadLeft := s(34)
		rightPad := s(10)
		if a.query != "" {
			rightPad = s(32)
		}
		editH := s(20)
		editY := a.rectSearchBox.Top + (barH-editH)/2
		editW := (a.rectSearchBox.Right - rightPad) - (a.rectSearchBox.Left + iconPadLeft)
		if editW < s(40) {
			editW = s(40)
		}
		procMoveWindow.Call(
			a.searchEdit,
			uintptr(a.rectSearchBox.Left+iconPadLeft),
			uintptr(editY),
			uintptr(editW),
			uintptr(editH),
			1,
		)
	}

	a.updateScrollBounds()
}

func (a *App) cardRect(filteredIdx int) RECT {
	s := a.theme.S
	padL := s(14)
	padR := s(14)
	if a.maxScrollY > 0 {
		padR = s(18)
	}
	step := a.cardHeight() + a.cardGap()
	top := a.rectListArea.Top + s(6) + int32(filteredIdx)*step - a.scrollY
	return RECT{
		Left:   padL,
		Top:    top,
		Right:  a.clientW - padR,
		Bottom: top + a.cardHeight(),
	}
}

func (a *App) cardActionRects(card RECT) (editBtn RECT, delBtn RECT) {
	s := a.theme.S
	btnSize := s(28)
	gap := s(4)
	cy := (card.Top + card.Bottom) / 2
	delBtn = RECT{
		Left:   card.Right - s(10) - btnSize,
		Top:    cy - btnSize/2,
		Right:  card.Right - s(10),
		Bottom: cy + btnSize/2,
	}
	editBtn = RECT{
		Left:   delBtn.Left - gap - btnSize,
		Top:    cy - btnSize/2,
		Right:  delBtn.Left - gap,
		Bottom: cy + btnSize/2,
	}
	return editBtn, delBtn
}

func (a *App) hitTest(x, y int32) hitTarget {
	if ptInRect(a.rectAboutBtn, x, y) {
		return hitTarget{kind: hitAboutBtn}
	}
	if ptInRect(a.rectAddBtn, x, y) {
		return hitTarget{kind: hitAddBtn}
	}
	if a.query != "" && ptInRect(a.rectClearBtn, x, y) {
		return hitTarget{kind: hitClearSearch}
	}
	if ptInRect(a.rectSearchBox, x, y) {
		return hitTarget{kind: hitSearchBox}
	}
	if a.maxScrollY > 0 {
		// Generous horizontal hit area for slim scrollbar
		sbHit := a.rectScrollBar
		sbHit.Left -= a.theme.S(5)
		sbHit.Right += a.theme.S(4)
		if ptInRect(sbHit, x, y) {
			if y >= a.rectScrollThumb.Top && y < a.rectScrollThumb.Bottom {
				return hitTarget{kind: hitScrollThumb}
			}
			return hitTarget{kind: hitScrollTrack}
		}
	}
	if ptInRect(a.rectListArea, x, y) {
		for i := range a.filtered {
			cr := a.cardRect(i)
			if cr.Bottom < a.rectListArea.Top {
				continue
			}
			if cr.Top > a.rectListArea.Bottom {
				break
			}
			if ptInRect(cr, x, y) {
				editR, delR := a.cardActionRects(cr)
				if ptInRect(delR, x, y) {
					return hitTarget{kind: hitCardDelete, index: i}
				}
				if ptInRect(editR, x, y) {
					return hitTarget{kind: hitCardEdit, index: i}
				}
				return hitTarget{kind: hitCard, index: i}
			}
		}
	}
	return hitTarget{kind: hitNone}
}

func (a *App) ensureSelectedVisible() {
	if a.selectedIdx < 0 || a.selectedIdx >= len(a.filtered) {
		return
	}
	cr := a.cardRect(a.selectedIdx)
	s := a.theme.S
	if cr.Top < a.rectListArea.Top+s(4) {
		a.scrollY -= (a.rectListArea.Top + s(6)) - cr.Top
	} else if cr.Bottom > a.rectListArea.Bottom-s(4) {
		a.scrollY += cr.Bottom - (a.rectListArea.Bottom - s(6))
	}
	if a.scrollY < 0 {
		a.scrollY = 0
	}
	if a.scrollY > a.maxScrollY {
		a.scrollY = a.maxScrollY
	}
	a.computeScrollbarThumb()
}

func (a *App) showToast(msg string) {
	a.toastText = msg
	procKillTimer.Call(a.hwnd, timerToastClear)
	procSetTimer.Call(a.hwnd, timerToastClear, 2200, 0)
	procInvalidateRect.Call(a.hwnd, 0, 0)
}

func (a *App) openFilteredItem(filteredIdx int) {
	if filteredIdx < 0 || filteredIdx >= len(a.filtered) {
		return
	}
	item := a.store.Items[a.filtered[filteredIdx]]
	if err := openInNewExplorer(a.hwnd, item.Path); err != nil {
		messageBox(a.hwnd, err.Error(), "My Finder", MB_OK|MB_ICONERROR)
		return
	}
	a.showToast("Membuka: " + item.Name)
}

func (a *App) addShortcutDialog() {
	name, path, ok := showShortcutDialog(a.hwnd, a.theme, "Tambah Shortcut Baru", "", "")
	if !ok {
		return
	}
	item, err := a.store.Add(name, path)
	if err != nil {
		messageBox(a.hwnd, "Gagal menyimpan shortcut:\n"+err.Error(), "My Finder", MB_OK|MB_ICONERROR)
		return
	}
	a.refreshFilter()
	// Select newly added item if visible
	for i, storeIdx := range a.filtered {
		if a.store.Items[storeIdx].ID == item.ID {
			a.selectedIdx = i
			a.ensureSelectedVisible()
			break
		}
	}
	a.showToast("Ditambahkan: " + item.Name)
}

func (a *App) editFilteredItem(filteredIdx int) {
	if filteredIdx < 0 || filteredIdx >= len(a.filtered) {
		return
	}
	item := a.store.Items[a.filtered[filteredIdx]]
	name, path, ok := showShortcutDialog(a.hwnd, a.theme, "Edit Shortcut", item.Name, item.Path)
	if !ok {
		return
	}
	if err := a.store.Update(item.ID, name, path); err != nil {
		messageBox(a.hwnd, "Gagal memperbarui shortcut:\n"+err.Error(), "My Finder", MB_OK|MB_ICONERROR)
		return
	}
	a.refreshFilter()
	a.showToast("Diperbarui: " + name)
}

func (a *App) deleteFilteredItem(filteredIdx int) {
	if filteredIdx < 0 || filteredIdx >= len(a.filtered) {
		return
	}
	item := a.store.Items[a.filtered[filteredIdx]]
	msg := fmt.Sprintf("Hapus shortcut \"%s\"?\n(%s)", item.Name, item.Path)
	if messageBox(a.hwnd, msg, "Konfirmasi Hapus", MB_YESNO|MB_ICONQUESTION) != IDYES {
		return
	}
	if err := a.store.Delete(item.ID); err != nil {
		messageBox(a.hwnd, "Gagal menghapus shortcut:\n"+err.Error(), "My Finder", MB_OK|MB_ICONERROR)
		return
	}
	a.refreshFilter()
	a.showToast("Dihapus: " + item.Name)
}

func (a *App) moveFilteredItem(filteredIdx int, delta int) {
	if a.query != "" || filteredIdx < 0 || filteredIdx >= len(a.filtered) {
		return
	}
	item := a.store.Items[a.filtered[filteredIdx]]
	if a.store.Move(item.ID, delta) {
		a.selectedIdx = filteredIdx + delta
		a.refreshFilter()
		a.ensureSelectedVisible()
		procInvalidateRect.Call(a.hwnd, 0, 0)
	}
}

func (a *App) showContextMenu(filteredIdx int, clientX, clientY int32) {
	if filteredIdx < 0 || filteredIdx >= len(a.filtered) {
		return
	}
	a.selectedIdx = filteredIdx
	procInvalidateRect.Call(a.hwnd, 0, 0)

	menu, _, _ := procCreatePopupMenu.Call()
	if menu == 0 {
		return
	}
	defer procDestroyMenu.Call(menu)

	canReorder := a.query == ""
	upFlags := uint32(MF_STRING)
	if !canReorder || filteredIdx == 0 {
		upFlags |= MF_GRAYED
	}
	downFlags := uint32(MF_STRING)
	if !canReorder || filteredIdx >= len(a.filtered)-1 {
		downFlags |= MF_GRAYED
	}

	procAppendMenuW.Call(menu, MF_STRING, cmdOpenExplorer, uintptr(unsafe.Pointer(utf16Ptr("Buka di Explorer Baru"))))
	procAppendMenuW.Call(menu, MF_STRING, cmdCopyPath, uintptr(unsafe.Pointer(utf16Ptr("Salin Path"))))
	procAppendMenuW.Call(menu, MF_SEPARATOR, 0, 0)
	procAppendMenuW.Call(menu, MF_STRING, cmdEditItem, uintptr(unsafe.Pointer(utf16Ptr("Edit Shortcut..."))))
	procAppendMenuW.Call(menu, uintptr(upFlags), cmdMoveUp, uintptr(unsafe.Pointer(utf16Ptr("Geser ke Atas  (Alt+Up)"))))
	procAppendMenuW.Call(menu, uintptr(downFlags), cmdMoveDown, uintptr(unsafe.Pointer(utf16Ptr("Geser ke Bawah  (Alt+Down)"))))
	procAppendMenuW.Call(menu, MF_SEPARATOR, 0, 0)
	procAppendMenuW.Call(menu, MF_STRING, cmdDeleteItem, uintptr(unsafe.Pointer(utf16Ptr("Hapus"))))

	pt := POINT{X: clientX, Y: clientY}
	procClientToScreen.Call(a.hwnd, uintptr(unsafe.Pointer(&pt)))

	cmd, _, _ := procTrackPopupMenu.Call(
		menu,
		TPM_LEFTALIGN|TPM_TOPALIGN|TPM_RETURNCMD|TPM_RIGHTBUTTON,
		uintptr(pt.X),
		uintptr(pt.Y),
		0,
		a.hwnd,
		0,
	)

	switch cmd {
	case cmdOpenExplorer:
		a.openFilteredItem(filteredIdx)
	case cmdCopyPath:
		item := a.store.Items[a.filtered[filteredIdx]]
		if copyToClipboard(a.hwnd, item.Path) {
			a.showToast("Path disalin ke clipboard")
		}
	case cmdEditItem:
		a.editFilteredItem(filteredIdx)
	case cmdMoveUp:
		a.moveFilteredItem(filteredIdx, -1)
	case cmdMoveDown:
		a.moveFilteredItem(filteredIdx, 1)
	case cmdDeleteItem:
		a.deleteFilteredItem(filteredIdx)
	}
}

func (a *App) handleDropFiles(hDrop uintptr) {
	defer procDragFinish.Call(hDrop)
	count, _, _ := procDragQueryFileW.Call(hDrop, 0xFFFFFFFF, 0, 0)
	if count == 0 {
		return
	}

	existing := make(map[string]bool, len(a.store.Items))
	for _, it := range a.store.Items {
		existing[strings.ToLower(normalizePath(it.Path))] = true
	}

	added := 0
	var lastName string
	for i := uintptr(0); i < count; i++ {
		n, _, _ := procDragQueryFileW.Call(hDrop, i, 0, 0)
		if n == 0 {
			continue
		}
		buf := make([]uint16, n+1)
		procDragQueryFileW.Call(hDrop, i, uintptr(unsafe.Pointer(&buf[0])), n+1)
		droppedPath := normalizePath(syscall.UTF16ToString(buf))
		if droppedPath == "" {
			continue
		}
		key := strings.ToLower(droppedPath)
		if existing[key] {
			continue
		}
		name := defaultNameForPath(droppedPath)
		if _, err := a.store.Add(name, droppedPath); err == nil {
			existing[key] = true
			added++
			lastName = name
		}
	}

	if added > 0 {
		a.refreshFilter()
		if len(a.filtered) > 0 {
			a.selectedIdx = len(a.filtered) - 1
			a.ensureSelectedVisible()
		}
		if added == 1 {
			a.showToast("Ditambahkan: " + lastName)
		} else {
			a.showToast(fmt.Sprintf("%d shortcut ditambahkan", added))
		}
	} else {
		a.showToast("Folder sudah ada di daftar")
	}
}

func (a *App) handleKeyDown(vk uintptr) bool {
	ctrl := isKeyDown(VK_CONTROL)
	alt := isKeyDown(VK_MENU)

	if ctrl && (vk == 'F' || vk == 'f') {
		procSetFocus.Call(a.searchEdit)
		procSendMessageW.Call(a.searchEdit, EM_SETSEL, 0, ^uintptr(0))
		return true
	}
	if ctrl && (vk == 'N' || vk == 'n') {
		a.addShortcutDialog()
		return true
	}
	if alt && vk == VK_UP {
		a.moveFilteredItem(a.selectedIdx, -1)
		return true
	}
	if alt && vk == VK_DOWN {
		a.moveFilteredItem(a.selectedIdx, 1)
		return true
	}

	switch vk {
	case VK_DOWN:
		if len(a.filtered) > 0 {
			if a.selectedIdx < len(a.filtered)-1 {
				a.selectedIdx++
			}
			a.ensureSelectedVisible()
			procInvalidateRect.Call(a.hwnd, 0, 0)
		}
		return true
	case VK_UP:
		if len(a.filtered) > 0 {
			if a.selectedIdx > 0 {
				a.selectedIdx--
			}
			a.ensureSelectedVisible()
			procInvalidateRect.Call(a.hwnd, 0, 0)
		}
		return true
	case VK_RETURN:
		if len(a.filtered) > 0 {
			idx := a.selectedIdx
			if idx < 0 {
				idx = 0
			}
			a.openFilteredItem(idx)
		}
		return true
	case VK_ESCAPE:
		if a.query != "" {
			setWindowText(a.searchEdit, "")
			return true
		}
	case VK_F1:
		a.showAboutDialog()
		return true
	}
	return false
}

func (a *App) showAboutDialog() {
	showAboutDialog(a.hwnd, a.theme, a.store.FilePath())
}

// --- Custom Vector Icon Helpers (crisp at any DPI, zero external assets) ---

func drawSearchIcon(hdc uintptr, cx, cy, r int32, color uint32) {
	pen, _, _ := procCreatePen.Call(PS_SOLID, 2, uintptr(color))
	defer procDeleteObject.Call(pen)
	nullBrush, _, _ := procGetStockObject.Call(NULL_BRUSH)
	oldPen, _, _ := procSelectObject.Call(hdc, pen)
	oldBrush, _, _ := procSelectObject.Call(hdc, nullBrush)

	procEllipse.Call(hdc, uintptr(cx-r), uintptr(cy-r), uintptr(cx+r), uintptr(cy+r))
	// Handle
	hx1 := cx + (r*7)/10
	hy1 := cy + (r*7)/10
	hx2 := hx1 + r/2 + 2
	hy2 := hy1 + r/2 + 2
	procMoveToEx.Call(hdc, uintptr(hx1), uintptr(hy1), 0)
	procLineTo.Call(hdc, uintptr(hx2), uintptr(hy2))

	procSelectObject.Call(hdc, oldPen)
	procSelectObject.Call(hdc, oldBrush)
}

func drawFolderIcon(hdc uintptr, x, y, w, h int32, backCol, frontCol uint32) {
	// Back tab + upper body
	tab := RECT{Left: x, Top: y, Right: x + (w*46)/100, Bottom: y + (h*38)/100}
	drawRoundRect(hdc, tab, 3, backCol, backCol, false)

	back := RECT{Left: x, Top: y + (h*18)/100, Right: x + w, Bottom: y + h}
	drawRoundRect(hdc, back, 4, backCol, backCol, false)

	// Bright front flap
	front := RECT{Left: x, Top: y + (h*32)/100, Right: x + w, Bottom: y + h}
	drawRoundRect(hdc, front, 4, frontCol, frontCol, false)
}

func drawEditIcon(hdc uintptr, r RECT, color uint32) {
	pen, _, _ := procCreatePen.Call(PS_SOLID, 2, uintptr(color))
	defer procDeleteObject.Call(pen)
	oldPen, _, _ := procSelectObject.Call(hdc, pen)

	cx := (r.Left + r.Right) / 2
	cy := (r.Top + r.Bottom) / 2
	d := (r.Right - r.Left) / 5
	if d < 4 {
		d = 4
	}
	// Diagonal pencil shaft
	procMoveToEx.Call(hdc, uintptr(cx-d), uintptr(cy+d), 0)
	procLineTo.Call(hdc, uintptr(cx+d), uintptr(cy-d))
	// Pencil tip accent
	procMoveToEx.Call(hdc, uintptr(cx-d), uintptr(cy+d), 0)
	procLineTo.Call(hdc, uintptr(cx-d+3), uintptr(cy+d))
	// Underline
	procMoveToEx.Call(hdc, uintptr(cx-d), uintptr(cy+d+3), 0)
	procLineTo.Call(hdc, uintptr(cx+d), uintptr(cy+d+3))

	procSelectObject.Call(hdc, oldPen)
}

func drawCloseIcon(hdc uintptr, r RECT, color uint32) {
	pen, _, _ := procCreatePen.Call(PS_SOLID, 2, uintptr(color))
	defer procDeleteObject.Call(pen)
	oldPen, _, _ := procSelectObject.Call(hdc, pen)

	cx := (r.Left + r.Right) / 2
	cy := (r.Top + r.Bottom) / 2
	d := (r.Right - r.Left) / 5
	if d < 4 {
		d = 4
	}
	procMoveToEx.Call(hdc, uintptr(cx-d), uintptr(cy-d), 0)
	procLineTo.Call(hdc, uintptr(cx+d), uintptr(cy+d))
	procMoveToEx.Call(hdc, uintptr(cx+d), uintptr(cy-d), 0)
	procLineTo.Call(hdc, uintptr(cx-d), uintptr(cy+d))

	procSelectObject.Call(hdc, oldPen)
}

func drawInfoIcon(hdc uintptr, r RECT, color uint32) {
	cx := (r.Left + r.Right) / 2
	cy := (r.Top + r.Bottom) / 2
	rad := (r.Right - r.Left) / 4
	if rad < 6 {
		rad = 6
	}

	pen, _, _ := procCreatePen.Call(PS_SOLID, 2, uintptr(color))
	defer procDeleteObject.Call(pen)
	nullBrush, _, _ := procGetStockObject.Call(NULL_BRUSH)
	oldPen, _, _ := procSelectObject.Call(hdc, pen)
	oldBrush, _, _ := procSelectObject.Call(hdc, nullBrush)

	// Outer circle
	procEllipse.Call(hdc, uintptr(cx-rad), uintptr(cy-rad), uintptr(cx+rad), uintptr(cy+rad))
	// Stem of "i"
	procMoveToEx.Call(hdc, uintptr(cx), uintptr(cy-1), 0)
	procLineTo.Call(hdc, uintptr(cx), uintptr(cy+rad-3))

	procSelectObject.Call(hdc, oldPen)
	procSelectObject.Call(hdc, oldBrush)

	// Dot of "i"
	dotBrush, _, _ := procCreateSolidBrush.Call(uintptr(color))
	defer procDeleteObject.Call(dotBrush)
	dotR := RECT{Left: cx - 1, Top: cy - rad + 3, Right: cx + 2, Bottom: cy - rad + 6}
	procFillRect.Call(hdc, uintptr(unsafe.Pointer(&dotR)), dotBrush)
}

func (a *App) paint(hdc uintptr) {
	if a.clientW <= 0 || a.clientH <= 0 || a.theme == nil {
		return
	}

	memDC, _, _ := procCreateCompatibleDC.Call(hdc)
	defer procDeleteDC.Call(memDC)
	bmp, _, _ := procCreateCompatibleBitmap.Call(hdc, uintptr(a.clientW), uintptr(a.clientH))
	defer procDeleteObject.Call(bmp)
	oldBmp, _, _ := procSelectObject.Call(memDC, bmp)
	defer procSelectObject.Call(memDC, oldBmp)

	p := a.theme.Palette
	s := a.theme.S

	fullRC := RECT{Right: a.clientW, Bottom: a.clientH}
	procFillRect.Call(memDC, uintptr(unsafe.Pointer(&fullRC)), a.theme.BrushWindowBG)

	// --- Top Bar: Search Box ---
	searchBorder := p.InputBorder
	if a.searchFocused {
		searchBorder = p.InputFocusBorder
	}
	drawRoundRect(memDC, a.rectSearchBox, s(8), p.InputBG, searchBorder, true)
	drawSearchIcon(
		memDC,
		a.rectSearchBox.Left+s(16),
		(a.rectSearchBox.Top+a.rectSearchBox.Bottom)/2-s(1),
		s(5),
		p.TextMuted,
	)
	if a.query != "" {
		if a.hover.kind == hitClearSearch {
			drawRoundRect(memDC, a.rectClearBtn, s(5), p.ButtonHoverBG, p.ButtonHoverBG, false)
		}
		drawCloseIcon(memDC, a.rectClearBtn, p.TextSecondary)
	}

	// --- Top Bar: "+ Tambah" Button ---
	addBG := p.Accent
	if a.pressed.kind == hitAddBtn {
		addBG = p.AccentPressed
	} else if a.hover.kind == hitAddBtn {
		addBG = p.AccentHover
	}
	drawRoundRect(memDC, a.rectAddBtn, s(8), addBG, addBG, false)
	drawText(memDC, "+ Tambah", a.rectAddBtn, p.AccentText, a.theme.FontButton, DT_CENTER|DT_VCENTER|DT_SINGLELINE|DT_NOPREFIX)

	// --- Top Bar: About ("i") Button ---
	aboutBG := p.ButtonBG
	if a.pressed.kind == hitAboutBtn || a.hover.kind == hitAboutBtn {
		aboutBG = p.ButtonHoverBG
	}
	drawRoundRect(memDC, a.rectAboutBtn, s(8), aboutBG, p.ButtonBorder, true)
	drawInfoIcon(memDC, a.rectAboutBtn, p.TextSecondary)

	// --- Scrollable List Area (clipped) ---
	clipRgn, _, _ := procCreateRectRgn.Call(
		uintptr(a.rectListArea.Left),
		uintptr(a.rectListArea.Top),
		uintptr(a.rectListArea.Right),
		uintptr(a.rectListArea.Bottom),
	)
	procSelectClipRgn.Call(memDC, clipRgn)

	if len(a.filtered) == 0 {
		a.paintEmptyState(memDC)
	} else {
		for i, storeIdx := range a.filtered {
			cr := a.cardRect(i)
			if cr.Bottom < a.rectListArea.Top {
				continue
			}
			if cr.Top > a.rectListArea.Bottom {
				break
			}
			item := a.store.Items[storeIdx]
			isHovered := (a.hover.kind == hitCard || a.hover.kind == hitCardEdit || a.hover.kind == hitCardDelete) && a.hover.index == i
			isSelected := i == a.selectedIdx

			bg := p.CardBG
			bdr := p.CardBorder
			if isSelected && isHovered {
				bg = p.CardSelectedBG
				bdr = p.CardSelectedBdr
			} else if isHovered {
				bg = p.CardHoverBG
			} else if isSelected {
				bg = p.CardSelectedBG
				bdr = p.CardSelectedBdr
			}
			drawRoundRect(memDC, cr, s(8), bg, bdr, true)

			// Folder icon on left
			iconW := s(26)
			iconH := s(21)
			iconX := cr.Left + s(13)
			iconY := (cr.Top+cr.Bottom-iconH)/2
			drawFolderIcon(memDC, iconX, iconY, iconW, iconH, p.FolderBack, p.FolderFront)

			// Action buttons (shown when hovered or selected)
			editR, delR := a.cardActionRects(cr)
			textRight := cr.Right - s(14)
			if isHovered || isSelected {
				textRight = editR.Left - s(8)

				// Edit button
				if a.hover.kind == hitCardEdit && a.hover.index == i {
					drawRoundRect(memDC, editR, s(6), p.ButtonHoverBG, p.ButtonHoverBG, false)
					drawEditIcon(memDC, editR, p.TextPrimary)
				} else {
					drawEditIcon(memDC, editR, p.TextMuted)
				}

				// Delete button
				if a.hover.kind == hitCardDelete && a.hover.index == i {
					drawRoundRect(memDC, delR, s(6), p.DangerHoverBG, p.DangerHoverBG, false)
					drawCloseIcon(memDC, delR, p.Danger)
				} else {
					drawCloseIcon(memDC, delR, p.TextMuted)
				}
			}

			// Text block: Name + Path
			textLeft := iconX + iconW + s(12)
			nameR := RECT{
				Left:   textLeft,
				Top:    cr.Top + s(9),
				Right:  textRight,
				Bottom: cr.Top + s(30),
			}
			pathR := RECT{
				Left:   textLeft,
				Top:    cr.Top + s(31),
				Right:  textRight,
				Bottom: cr.Bottom - s(8),
			}
			drawText(memDC, item.Name, nameR, p.TextPrimary, a.theme.FontTitle, DT_LEFT|DT_VCENTER|DT_SINGLELINE|DT_END_ELLIPSIS|DT_NOPREFIX)
			drawText(memDC, item.Path, pathR, p.TextSecondary, a.theme.FontSub, DT_LEFT|DT_VCENTER|DT_SINGLELINE|DT_PATH_ELLIPSIS|DT_NOPREFIX)
		}

		// Slim scrollbar
		if a.maxScrollY > 0 {
			thumbCol := p.ScrollThumb
			if a.draggingScroll || a.hover.kind == hitScrollThumb {
				thumbCol = p.ScrollThumbHover
			}
			drawRoundRect(memDC, a.rectScrollThumb, s(6), thumbCol, thumbCol, false)
		}
	}

	procSelectClipRgn.Call(memDC, 0)
	procDeleteObject.Call(clipRgn)

	// --- Footer Bar ---
	divPen, _, _ := procCreatePen.Call(PS_SOLID, 1, uintptr(p.Divider))
	oldPen, _, _ := procSelectObject.Call(memDC, divPen)
	procMoveToEx.Call(memDC, 0, uintptr(a.rectFooter.Top), 0)
	procLineTo.Call(memDC, uintptr(a.clientW), uintptr(a.rectFooter.Top))
	procSelectObject.Call(memDC, oldPen)
	procDeleteObject.Call(divPen)

	if a.toastText != "" {
		toastR := RECT{
			Left:   s(12),
			Top:    a.rectFooter.Top + s(4),
			Right:  a.clientW - s(12),
			Bottom: a.rectFooter.Bottom - s(4),
		}
		drawRoundRect(memDC, toastR, s(5), p.ToastBG, p.ToastBG, false)
		toastTextR := toastR
		toastTextR.Left += s(8)
		toastTextR.Right -= s(8)
		drawText(memDC, a.toastText, toastTextR, p.ToastText, a.theme.FontSmall, DT_LEFT|DT_VCENTER|DT_SINGLELINE|DT_END_ELLIPSIS|DT_NOPREFIX)
	} else {
		var status string
		if a.query != "" {
			status = fmt.Sprintf("%d dari %d shortcut", len(a.filtered), len(a.store.Items))
		} else {
			status = fmt.Sprintf("%d shortcut", len(a.store.Items))
		}
		leftR := RECT{Left: s(14), Top: a.rectFooter.Top, Right: a.clientW / 2, Bottom: a.rectFooter.Bottom}
		rightR := RECT{Left: a.clientW / 2, Top: a.rectFooter.Top, Right: a.clientW - s(14), Bottom: a.rectFooter.Bottom}
		drawText(memDC, status, leftR, p.TextMuted, a.theme.FontSub, DT_LEFT|DT_VCENTER|DT_SINGLELINE|DT_NOPREFIX)
		drawText(memDC, "Klik untuk buka • Drag folder ke sini", rightR, p.TextMuted, a.theme.FontSub, DT_RIGHT|DT_VCENTER|DT_SINGLELINE|DT_NOPREFIX)
	}

	procBitBlt.Call(hdc, 0, 0, uintptr(a.clientW), uintptr(a.clientH), memDC, 0, 0, SRCCOPY)
}

func (a *App) paintEmptyState(hdc uintptr) {
	p := a.theme.Palette
	s := a.theme.S
	cx := (a.rectListArea.Left + a.rectListArea.Right) / 2
	cy := (a.rectListArea.Top + a.rectListArea.Bottom) / 2 - s(16)

	iconW := s(44)
	iconH := s(35)
	drawFolderIcon(hdc, cx-iconW/2, cy-s(38), iconW, iconH, p.FolderBack, p.FolderFront)

	var title, sub string
	if len(a.store.Items) == 0 {
		title = "Belum ada shortcut folder"
		sub = "Klik '+ Tambah' atau drag & drop folder ke jendela ini"
	} else {
		title = "Tidak ditemukan"
		sub = fmt.Sprintf("Tidak ada shortcut cocok dengan \"%s\"", a.query)
	}

	r1 := RECT{Left: s(20), Top: cy + s(6), Right: a.clientW - s(20), Bottom: cy + s(30)}
	r2 := RECT{Left: s(20), Top: cy + s(30), Right: a.clientW - s(20), Bottom: cy + s(52)}
	drawText(hdc, title, r1, p.TextPrimary, a.theme.FontTitle, DT_CENTER|DT_VCENTER|DT_SINGLELINE|DT_END_ELLIPSIS|DT_NOPREFIX)
	drawText(hdc, sub, r2, p.TextMuted, a.theme.FontSub, DT_CENTER|DT_VCENTER|DT_SINGLELINE|DT_END_ELLIPSIS|DT_NOPREFIX)
}

// searchEditWndProc subclasses the search EDIT control so arrow keys, Enter, Esc, Ctrl+N,
// and placeholder drawing work seamlessly without losing typing focus.
func searchEditWndProc(hwnd uintptr, msg uint32, wParam, lParam uintptr) uintptr {
	a := globalApp
	if a == nil {
		return 0
	}
	switch msg {
	case WM_KEYDOWN:
		if a.handleKeyDown(wParam) {
			return 0
		}
	case WM_CHAR:
		// Suppress beep on Enter / Escape / Ctrl+F / Ctrl+N
		if wParam == VK_RETURN || wParam == VK_ESCAPE || wParam == 6 || wParam == 14 {
			return 0
		}
	case WM_PAINT:
		ret, _, _ := procCallWindowProcW.Call(a.origEditWP, hwnd, uintptr(msg), wParam, lParam)
		n, _, _ := procGetWindowTextLengthW.Call(hwnd)
		if n == 0 && a.theme != nil {
			drawEditPlaceholder(hwnd, a.theme, "Cari shortcut... (Ctrl+F)")
		}
		return ret
	}
	r, _, _ := procCallWindowProcW.Call(a.origEditWP, hwnd, uintptr(msg), wParam, lParam)
	return r
}

var (
	procGetDC     = modUser32.NewProc("GetDC")
	procReleaseDC = modUser32.NewProc("ReleaseDC")
)

func drawEditPlaceholder(hwnd uintptr, theme *ThemeResources, placeholder string) {
	hdc, _, _ := procGetDC.Call(hwnd)
	if hdc == 0 {
		return
	}
	defer procReleaseDC.Call(hwnd, hdc)
	var rc RECT
	procGetClientRect.Call(hwnd, uintptr(unsafe.Pointer(&rc)))
	rc.Left += theme.S(3)
	drawText(hdc, placeholder, rc, theme.Palette.TextMuted, theme.FontRegular, DT_LEFT|DT_VCENTER|DT_SINGLELINE|DT_END_ELLIPSIS|DT_NOPREFIX)
}

func mainWndProc(hwnd uintptr, msg uint32, wParam, lParam uintptr) uintptr {
	a := globalApp
	if a == nil {
		r, _, _ := procDefWindowProcW.Call(hwnd, uintptr(msg), wParam, lParam)
		return r
	}

	switch msg {
	case WM_CREATE:
		a.hwnd = hwnd
		dpi := getWindowDPI(hwnd)
		a.theme = newThemeResources(isWindowsDarkMode(), dpi)
		setDarkTitleBar(hwnd, a.theme.Palette.Dark)
		procDragAcceptFiles.Call(hwnd, 1)

		a.searchEdit, _, _ = procCreateWindowExW.Call(
			0,
			uintptr(unsafe.Pointer(utf16Ptr("EDIT"))),
			0,
			WS_CHILD|WS_VISIBLE|ES_LEFT|ES_AUTOHSCROLL,
			0, 0, 100, 24,
			hwnd,
			idSearchEdit,
			uintptr(getModuleHandle()),
			0,
		)
		procSendMessageW.Call(a.searchEdit, WM_SETFONT, a.theme.FontRegular, 1)
		a.origEditWP = setWindowLongPtr(a.searchEdit, GWLP_WNDPROC, syscall.NewCallback(searchEditWndProc))

		a.layout()
		return 0

	case WM_GETMINMAXINFO:
		if a.theme != nil && lParam != 0 {
			var mmi MINMAXINFO
			procRtlMoveMemory.Call(uintptr(unsafe.Pointer(&mmi)), lParam, unsafe.Sizeof(mmi))
			mmi.PtMinTrackSize.X = a.theme.S(360)
			mmi.PtMinTrackSize.Y = a.theme.S(300)
			procRtlMoveMemory.Call(lParam, uintptr(unsafe.Pointer(&mmi)), unsafe.Sizeof(mmi))
			return 0
		}

	case WM_SIZE:
		a.layout()
		procInvalidateRect.Call(hwnd, 0, 0)
		return 0

	case WM_DPICHANGED:
		newDPI := int32(uloword(wParam))
		var suggested RECT
		if lParam != 0 {
			procRtlMoveMemory.Call(uintptr(unsafe.Pointer(&suggested)), lParam, unsafe.Sizeof(suggested))
		}
		a.applyTheme(a.theme.Palette.Dark, newDPI)
		if lParam != 0 {
			procSetWindowPos.Call(
				hwnd, 0,
				uintptr(suggested.Left), uintptr(suggested.Top),
				uintptr(suggested.Right-suggested.Left), uintptr(suggested.Bottom-suggested.Top),
				SWP_NOZORDER|SWP_NOACTIVATE,
			)
		}
		return 0

	case WM_SETTINGCHANGE:
		dark := isWindowsDarkMode()
		if a.theme != nil && dark != a.theme.Palette.Dark {
			a.applyTheme(dark, a.theme.DPI)
		}
		return 0

	case WM_ERASEBKGND:
		return 1

	case WM_PAINT:
		var ps PAINTSTRUCT
		hdc, _, _ := procBeginPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))
		a.paint(hdc)
		procEndPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))
		return 0

	case WM_CTLCOLOREDIT:
		if a.theme != nil {
			hdc := wParam
			procSetTextColor.Call(hdc, uintptr(a.theme.Palette.TextPrimary))
			procSetBkColor.Call(hdc, uintptr(a.theme.Palette.InputBG))
			return a.theme.BrushInputBG
		}

	case WM_COMMAND:
		id := uloword(wParam)
		code := uhiword(wParam)
		if id == idSearchEdit {
			switch code {
			case EN_CHANGE:
				newQ := getWindowText(a.searchEdit)
				if newQ != a.query {
					wasEmpty := a.query == ""
					isEmpty := newQ == ""
					a.query = newQ
					a.scrollY = 0
					a.selectedIdx = 0
					a.refreshFilter()
					if wasEmpty != isEmpty {
						a.layout()
					}
					procInvalidateRect.Call(hwnd, 0, 0)
				}
			case EN_SETFOCUS:
				a.searchFocused = true
				procInvalidateRect.Call(hwnd, 0, 0)
			case EN_KILLFOCUS:
				a.searchFocused = false
				procInvalidateRect.Call(hwnd, 0, 0)
			}
		}
		return 0

	case WM_KEYDOWN:
		if a.handleKeyDown(wParam) {
			return 0
		}
		if wParam == VK_DELETE && len(a.filtered) > 0 && a.selectedIdx >= 0 {
			a.deleteFilteredItem(a.selectedIdx)
			return 0
		}

	case WM_MOUSEMOVE:
		x, y := loword(lParam), hiword(lParam)
		if a.draggingScroll && a.maxScrollY > 0 {
			trackH := (a.rectScrollBar.Bottom - a.rectScrollBar.Top) - (a.rectScrollThumb.Bottom - a.rectScrollThumb.Top)
			if trackH > 0 {
				dy := y - a.dragStartMouseY
				a.scrollY = a.dragStartScrollY + (dy*a.maxScrollY)/trackH
				if a.scrollY < 0 {
					a.scrollY = 0
				}
				if a.scrollY > a.maxScrollY {
					a.scrollY = a.maxScrollY
				}
				a.computeScrollbarThumb()
				procInvalidateRect.Call(hwnd, 0, 0)
			}
			return 0
		}

		if !a.trackingMouse {
			tme := TRACKMOUSEEVENT{
				CbSize:    uint32(unsafe.Sizeof(TRACKMOUSEEVENT{})),
				DwFlags:   TME_LEAVE,
				HwndTrack: hwnd,
			}
			procTrackMouseEvent.Call(uintptr(unsafe.Pointer(&tme)))
			a.trackingMouse = true
		}

		ht := a.hitTest(x, y)
		if ht != a.hover {
			a.hover = ht
			procInvalidateRect.Call(hwnd, 0, 0)
		}
		return 0

	case WM_MOUSELEAVE:
		a.trackingMouse = false
		if a.hover.kind != hitNone {
			a.hover = hitTarget{}
			procInvalidateRect.Call(hwnd, 0, 0)
		}
		return 0

	case WM_SETCURSOR:
		if uloword(lParam) == HTCLIENT {
			switch a.hover.kind {
			case hitAddBtn, hitAboutBtn, hitClearSearch, hitCard, hitCardEdit, hitCardDelete, hitScrollThumb:
				procSetCursor.Call(uintptr(loadCursor(IDC_HAND)))
				return 1
			}
		}

	case WM_LBUTTONDOWN:
		x, y := loword(lParam), hiword(lParam)
		ht := a.hitTest(x, y)
		a.pressed = ht
		switch ht.kind {
		case hitSearchBox:
			procSetFocus.Call(a.searchEdit)
		case hitScrollThumb:
			a.draggingScroll = true
			a.dragStartMouseY = y
			a.dragStartScrollY = a.scrollY
			procSetCapture.Call(hwnd)
		case hitScrollTrack:
			// Jump scroll page
			viewH := a.rectListArea.Bottom - a.rectListArea.Top
			if y < a.rectScrollThumb.Top {
				a.scrollY -= viewH * 3 / 4
			} else {
				a.scrollY += viewH * 3 / 4
			}
			a.updateScrollBounds()
			procInvalidateRect.Call(hwnd, 0, 0)
		case hitAddBtn, hitAboutBtn, hitClearSearch, hitCard, hitCardEdit, hitCardDelete:
			if ht.kind == hitCard || ht.kind == hitCardEdit || ht.kind == hitCardDelete {
				a.selectedIdx = ht.index
			}
			procSetCapture.Call(hwnd)
			procInvalidateRect.Call(hwnd, 0, 0)
		}
		return 0

	case WM_LBUTTONUP:
		x, y := loword(lParam), hiword(lParam)
		if a.draggingScroll {
			a.draggingScroll = false
			procReleaseCapture.Call()
			procInvalidateRect.Call(hwnd, 0, 0)
			return 0
		}
		if a.pressed.kind != hitNone {
			pressed := a.pressed
			a.pressed = hitTarget{}
			procReleaseCapture.Call()
			ht := a.hitTest(x, y)
			procInvalidateRect.Call(hwnd, 0, 0)

			if ht == pressed {
				switch ht.kind {
				case hitAddBtn:
					a.addShortcutDialog()
				case hitAboutBtn:
					a.showAboutDialog()
				case hitClearSearch:
					setWindowText(a.searchEdit, "")
					procSetFocus.Call(a.searchEdit)
				case hitCard:
					a.openFilteredItem(ht.index)
				case hitCardEdit:
					a.editFilteredItem(ht.index)
				case hitCardDelete:
					a.deleteFilteredItem(ht.index)
				}
			}
		}
		return 0

	case WM_RBUTTONUP:
		x, y := loword(lParam), hiword(lParam)
		ht := a.hitTest(x, y)
		if ht.kind == hitCard || ht.kind == hitCardEdit || ht.kind == hitCardDelete {
			a.showContextMenu(ht.index, x, y)
		}
		return 0

	case WM_MOUSEWHEEL:
		if a.maxScrollY > 0 {
			delta := hiword(wParam)
			step := a.theme.S(44)
			a.scrollY -= (delta / 120) * step
			a.updateScrollBounds()
			procInvalidateRect.Call(hwnd, 0, 0)
		}
		return 0

	case WM_DROPFILES:
		a.handleDropFiles(wParam)
		return 0

	case WM_TIMER:
		if wParam == timerToastClear {
			procKillTimer.Call(hwnd, timerToastClear)
			a.toastText = ""
			procInvalidateRect.Call(hwnd, 0, 0)
		}
		return 0

	case WM_DESTROY:
		procPostQuitMessage.Call(0)
		return 0
	}

	r, _, _ := procDefWindowProcW.Call(hwnd, uintptr(msg), wParam, lParam)
	return r
}
