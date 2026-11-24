package main

type Palette struct {
	Dark             bool
	WindowBG         uint32
	SurfaceBG        uint32
	CardBG           uint32
	CardHoverBG      uint32
	CardSelectedBG   uint32
	CardBorder       uint32
	CardSelectedBdr  uint32
	InputBG          uint32
	InputBorder      uint32
	InputFocusBorder uint32
	TextPrimary      uint32
	TextSecondary    uint32
	TextMuted        uint32
	Accent           uint32
	AccentHover      uint32
	AccentPressed    uint32
	AccentText       uint32
	ButtonBG         uint32
	ButtonHoverBG    uint32
	ButtonBorder     uint32
	Danger           uint32
	DangerHoverBG    uint32
	FolderBack       uint32
	FolderFront      uint32
	Divider          uint32
	ScrollTrack      uint32
	ScrollThumb      uint32
	ScrollThumbHover uint32
	ToastBG          uint32
	ToastText        uint32
}

func getPalette(dark bool) Palette {
	if dark {
		return Palette{
			Dark:             true,
			WindowBG:         rgb(28, 28, 30),
			SurfaceBG:        rgb(36, 36, 39),
			CardBG:           rgb(40, 40, 44),
			CardHoverBG:      rgb(50, 50, 56),
			CardSelectedBG:   rgb(46, 58, 72),
			CardBorder:       rgb(56, 56, 62),
			CardSelectedBdr:  rgb(76, 158, 235),
			InputBG:          rgb(43, 43, 47),
			InputBorder:      rgb(68, 68, 74),
			InputFocusBorder: rgb(96, 205, 255),
			TextPrimary:      rgb(243, 243, 245),
			TextSecondary:    rgb(166, 166, 174),
			TextMuted:        rgb(122, 122, 130),
			Accent:           rgb(0, 120, 212),
			AccentHover:      rgb(24, 136, 222),
			AccentPressed:    rgb(0, 102, 184),
			AccentText:       rgb(255, 255, 255),
			ButtonBG:         rgb(52, 52, 58),
			ButtonHoverBG:    rgb(66, 66, 74),
			ButtonBorder:     rgb(76, 76, 84),
			Danger:           rgb(244, 118, 108),
			DangerHoverBG:    rgb(78, 38, 38),
			FolderBack:       rgb(214, 158, 30),
			FolderFront:      rgb(244, 198, 68),
			Divider:          rgb(52, 52, 56),
			ScrollTrack:      rgb(34, 34, 37),
			ScrollThumb:      rgb(80, 80, 88),
			ScrollThumbHover: rgb(110, 110, 120),
			ToastBG:          rgb(45, 78, 56),
			ToastText:        rgb(220, 248, 228),
		}
	}
	return Palette{
		Dark:             false,
		WindowBG:         rgb(244, 245, 247),
		SurfaceBG:        rgb(255, 255, 255),
		CardBG:           rgb(255, 255, 255),
		CardHoverBG:      rgb(240, 245, 252),
		CardSelectedBG:   rgb(230, 240, 252),
		CardBorder:       rgb(224, 226, 232),
		CardSelectedBdr:  rgb(0, 103, 192),
		InputBG:          rgb(255, 255, 255),
		InputBorder:      rgb(206, 210, 218),
		InputFocusBorder: rgb(0, 103, 192),
		TextPrimary:      rgb(24, 26, 30),
		TextSecondary:    rgb(94, 98, 108),
		TextMuted:        rgb(138, 142, 152),
		Accent:           rgb(0, 103, 192),
		AccentHover:      rgb(20, 118, 204),
		AccentPressed:    rgb(0, 86, 162),
		AccentText:       rgb(255, 255, 255),
		ButtonBG:         rgb(255, 255, 255),
		ButtonHoverBG:    rgb(238, 240, 244),
		ButtonBorder:     rgb(206, 210, 218),
		Danger:           rgb(196, 43, 28),
		DangerHoverBG:    rgb(253, 236, 234),
		FolderBack:       rgb(216, 160, 24),
		FolderFront:      rgb(245, 196, 56),
		Divider:          rgb(226, 228, 234),
		ScrollTrack:      rgb(236, 238, 242),
		ScrollThumb:      rgb(190, 194, 202),
		ScrollThumbHover: rgb(150, 154, 164),
		ToastBG:          rgb(224, 244, 230),
		ToastText:        rgb(20, 96, 46),
	}
}

type ThemeResources struct {
	Palette        Palette
	DPI            int32
	FontTitle      uintptr
	FontRegular    uintptr
	FontSub        uintptr
	FontSmall      uintptr
	FontButton     uintptr
	BrushInputBG   uintptr
	BrushWindowBG  uintptr
}

func newThemeResources(dark bool, dpi int32) *ThemeResources {
	if dpi <= 0 {
		dpi = 96
	}
	p := getPalette(dark)
	t := &ThemeResources{
		Palette: p,
		DPI:     dpi,
	}
	t.FontTitle = createFont(dpi, 10, FW_SEMIBOLD, "Segoe UI")
	t.FontRegular = createFont(dpi, 10, FW_NORMAL, "Segoe UI")
	t.FontSub = createFont(dpi, 8, FW_NORMAL, "Segoe UI")
	t.FontSmall = createFont(dpi, 8, FW_SEMIBOLD, "Segoe UI")
	t.FontButton = createFont(dpi, 9, FW_SEMIBOLD, "Segoe UI")
	t.BrushInputBG, _, _ = procCreateSolidBrush.Call(uintptr(p.InputBG))
	t.BrushWindowBG, _, _ = procCreateSolidBrush.Call(uintptr(p.WindowBG))
	return t
}

func (t *ThemeResources) Destroy() {
	if t == nil {
		return
	}
	for _, h := range []uintptr{
		t.FontTitle,
		t.FontRegular,
		t.FontSub,
		t.FontSmall,
		t.FontButton,
		t.BrushInputBG,
		t.BrushWindowBG,
	} {
		if h != 0 {
			procDeleteObject.Call(h)
		}
	}
}

func (t *ThemeResources) S(px int32) int32 {
	return (px * t.DPI + 48) / 96
}
