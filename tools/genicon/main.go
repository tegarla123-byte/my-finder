package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"path/filepath"
)

type rgbaF struct {
	r, g, b, a float64
}

func hexColor(r, g, b uint8, a float64) rgbaF {
	return rgbaF{float64(r) / 255.0, float64(g) / 255.0, float64(b) / 255.0, a}
}

func lerp(a, b, t float64) float64 {
	if t < 0 {
		t = 0
	} else if t > 1 {
		t = 1
	}
	return a + (b-a)*t
}

func lerpColor(c1, c2 rgbaF, t float64) rgbaF {
	return rgbaF{
		r: lerp(c1.r, c2.r, t),
		g: lerp(c1.g, c2.g, t),
		b: lerp(c1.b, c2.b, t),
		a: lerp(c1.a, c2.a, t),
	}
}

// over blends src over dst (straight alpha).
func over(dst, src rgbaF) rgbaF {
	outA := src.a + dst.a*(1-src.a)
	if outA <= 1e-9 {
		return rgbaF{}
	}
	return rgbaF{
		r: (src.r*src.a + dst.r*dst.a*(1-src.a)) / outA,
		g: (src.g*src.a + dst.g*dst.a*(1-src.a)) / outA,
		b: (src.b*src.a + dst.b*dst.a*(1-src.a)) / outA,
		a: outA,
	}
}

// sdRoundBox returns signed distance to a rounded rectangle centered at (cx, cy) with half-extents (hx, hy) and radius r.
func sdRoundBox(px, py, cx, cy, hx, hy, r float64) float64 {
	qx := math.Abs(px-cx) - (hx - r)
	qy := math.Abs(py-cy) - (hy - r)
	ox := math.Max(qx, 0)
	oy := math.Max(qy, 0)
	return math.Hypot(ox, oy) + math.Min(math.Max(qx, qy), 0) - r
}

// sdCircle returns signed distance to a circle.
func sdCircle(px, py, cx, cy, r float64) float64 {
	return math.Hypot(px-cx, py-cy) - r
}

// sdSegment returns distance from point (px, py) to line segment (ax, ay)-(bx, by).
func sdSegment(px, py, ax, ay, bx, by float64) float64 {
	pax, pay := px-ax, py-ay
	bax, bay := bx-ax, by-ay
	denom := bax*bax + bay*bay
	h := 0.0
	if denom > 1e-12 {
		h = (pax*bax + pay*bay) / denom
		if h < 0 {
			h = 0
		} else if h > 1 {
			h = 1
		}
	}
	return math.Hypot(pax-bax*h, pay-bay*h)
}

func smoothAlpha(sd, aa float64) float64 {
	if sd <= -aa {
		return 1
	}
	if sd >= aa {
		return 0
	}
	t := (aa - sd) / (2 * aa)
	return t * t * (3 - 2*t)
}

// sampleIcon evaluates the My Finder Fluent-style icon at normalized coordinates (x, y) in [0, 1].
func sampleIcon(x, y, pxSize float64) rgbaF {
	aa := 0.75 * pxSize
	col := rgbaF{}

	// 1. Rounded-square dark slate tile with soft elevation
	tileSD := sdRoundBox(x, y, 0.5, 0.5, 0.44, 0.44, 0.19)
	if tileA := smoothAlpha(tileSD, aa); tileA > 0 {
		gradT := (x*0.35 + y*0.65)
		tileCol := lerpColor(hexColor(36, 52, 74, 1), hexColor(15, 23, 38, 1), gradT)
		// Smooth top-to-mid glass highlight rim (no sharp cutoff)
		if tileSD > -0.024 {
			rimT := 1.0 - math.Min(math.Abs(tileSD+0.011)/0.014, 1.0)
			fadeY := lerp(1.0, 0.0, (y-0.08)/0.58)
			if fadeY > 0 {
				rimCol := hexColor(168, 198, 234, 0.30*rimT*fadeY)
				tileCol = over(tileCol, rimCol)
			}
		}
		tileCol.a *= tileA
		col = over(col, tileCol)
	}

	// 2. Soft drop shadow beneath the folder
	shadowSD := sdRoundBox(x, y, 0.50, 0.56, 0.285, 0.20, 0.065)
	if shadowSD < 0.08 {
		s := math.Max(0, 1.0-(shadowSD+0.02)/0.10)
		s = s * s * 0.42
		if tileA := smoothAlpha(tileSD, aa); tileA > 0 {
			col = over(col, rgbaF{0, 0, 0, s * tileA})
		}
	}

	// 3. Folder back plate + top-left tab (rich warm amber)
	tabSD := sdRoundBox(x, y, 0.34, 0.345, 0.13, 0.085, 0.042)
	backBodySD := sdRoundBox(x, y, 0.49, 0.515, 0.28, 0.195, 0.058)
	backSD := math.Min(tabSD, backBodySD)
	if backA := smoothAlpha(backSD, aa); backA > 0 {
		t := (y - 0.26) / 0.45
		c := lerpColor(hexColor(232, 166, 28, 1), hexColor(196, 128, 14, 1), t)
		c.a *= backA
		col = over(col, c)
	}

	// 4. Folder front flap (bright golden yellow with top sheen)
	frontSD := sdRoundBox(x, y, 0.49, 0.555, 0.28, 0.165, 0.056)
	if frontA := smoothAlpha(frontSD, aa); frontA > 0 {
		t := (y - 0.39) / 0.33
		c := lerpColor(hexColor(255, 217, 92, 1), hexColor(242, 175, 34, 1), t)
		// Crisp top lip sheen
		if y >= 0.39 && y <= 0.425 {
			sheen := 1.0 - math.Abs(y-0.402)/0.023
			if sheen > 0 {
				c = over(c, hexColor(255, 246, 200, 0.45*sheen))
			}
		}
		c.a *= frontA
		col = over(col, c)
	}

	// 5. Quick-launch sky-blue circular badge at bottom-right
	bx, by, br := 0.70, 0.69, 0.165
	// Subtle dark halo separating badge from folder
	haloSD := sdCircle(x, y, bx, by, br+0.026)
	if haloA := smoothAlpha(haloSD, aa); haloA > 0 {
		col = over(col, hexColor(14, 22, 36, 0.72*haloA))
	}

	badgeSD := sdCircle(x, y, bx, by, br)
	if badgeA := smoothAlpha(badgeSD, aa); badgeA > 0 {
		bt := ((x - (bx - br)) + (y - (by - br))) / (4 * br)
		bc := lerpColor(hexColor(56, 189, 248, 1), hexColor(2, 132, 199, 1), bt)
		bc.a *= badgeA
		col = over(col, bc)
	}

	// 6. Crisp white diagonal launch arrow (↗) inside badge
	stroke := 0.021
	ax1, ay1 := bx-0.056, by+0.056
	ax2, ay2 := bx+0.054, by-0.054
	dShaft := sdSegment(x, y, ax1, ay1, ax2, ay2) - stroke
	dHead1 := sdSegment(x, y, ax2, ay2, ax2-0.074, ay2) - stroke
	dHead2 := sdSegment(x, y, ax2, ay2, ax2, ay2+0.074) - stroke
	arrowSD := math.Min(dShaft, math.Min(dHead1, dHead2))
	if arrowA := smoothAlpha(arrowSD, aa); arrowA > 0 {
		col = over(col, hexColor(255, 255, 255, arrowA))
	}

	return col
}

func renderIcon(size int) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, size, size))
	// 4x4 supersampling per pixel for silky smooth edges at all sizes
	const ss = 4
	invSize := 1.0 / float64(size)
	invSS := 1.0 / float64(ss)
	pxSize := invSize

	for py := 0; py < size; py++ {
		for px := 0; px < size; px++ {
			var sumR, sumG, sumB, sumA float64
			for sy := 0; sy < ss; sy++ {
				y := (float64(py) + (float64(sy)+0.5)*invSS) * invSize
				for sx := 0; sx < ss; sx++ {
					x := (float64(px) + (float64(sx)+0.5)*invSS) * invSize
					c := sampleIcon(x, y, pxSize)
					// Accumulate premultiplied for accurate anti-aliasing
					sumR += c.r * c.a
					sumG += c.g * c.a
					sumB += c.b * c.a
					sumA += c.a
				}
			}
			n := float64(ss * ss)
			a := sumA / n
			var r, g, b float64
			if a > 1e-6 {
				r = (sumR / n) / a
				g = (sumG / n) / a
				b = (sumB / n) / a
			}
			img.SetNRGBA(px, py, color.NRGBA{
				R: clampByte(r),
				G: clampByte(g),
				B: clampByte(b),
				A: clampByte(a),
			})
		}
	}
	return img
}

func clampByte(v float64) uint8 {
	if v <= 0 {
		return 0
	}
	if v >= 1 {
		return 255
	}
	return uint8(math.Round(v * 255.0))
}

func encodePNG(img image.Image) ([]byte, error) {
	var buf bytes.Buffer
	enc := png.Encoder{CompressionLevel: png.BestCompression}
	if err := enc.Encode(&buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

type iconFrame struct {
	size int
	png  []byte
}

// buildICO builds a standard Windows multi-resolution ICO file with PNG-encoded frames.
func buildICO(frames []iconFrame) []byte {
	var out bytes.Buffer
	// ICONDIR (6 bytes)
	_ = binary.Write(&out, binary.LittleEndian, uint16(0)) // reserved
	_ = binary.Write(&out, binary.LittleEndian, uint16(1)) // type = ICO
	_ = binary.Write(&out, binary.LittleEndian, uint16(len(frames)))

	offset := uint32(6 + 16*len(frames))
	for _, f := range frames {
		dim := uint8(f.size)
		if f.size >= 256 {
			dim = 0
		}
		out.WriteByte(dim)                                          // width
		out.WriteByte(dim)                                          // height
		out.WriteByte(0)                                            // color count
		out.WriteByte(0)                                            // reserved
		_ = binary.Write(&out, binary.LittleEndian, uint16(1))      // planes
		_ = binary.Write(&out, binary.LittleEndian, uint16(32))     // bpp
		_ = binary.Write(&out, binary.LittleEndian, uint32(len(f.png)))
		_ = binary.Write(&out, binary.LittleEndian, offset)
		offset += uint32(len(f.png))
	}
	for _, f := range frames {
		out.Write(f.png)
	}
	return out.Bytes()
}

// buildGroupIconData builds the RT_GROUP_ICON payload referencing RT_ICON IDs 1..N.
func buildGroupIconData(frames []iconFrame) []byte {
	var out bytes.Buffer
	// GRPICONDIR (6 bytes)
	_ = binary.Write(&out, binary.LittleEndian, uint16(0))
	_ = binary.Write(&out, binary.LittleEndian, uint16(1))
	_ = binary.Write(&out, binary.LittleEndian, uint16(len(frames)))
	// GRPICONDIRENTRY (14 bytes each)
	for i, f := range frames {
		dim := uint8(f.size)
		if f.size >= 256 {
			dim = 0
		}
		out.WriteByte(dim)
		out.WriteByte(dim)
		out.WriteByte(0)
		out.WriteByte(0)
		_ = binary.Write(&out, binary.LittleEndian, uint16(1))
		_ = binary.Write(&out, binary.LittleEndian, uint16(32))
		_ = binary.Write(&out, binary.LittleEndian, uint32(len(f.png)))
		_ = binary.Write(&out, binary.LittleEndian, uint16(i+1)) // RT_ICON resource ID
	}
	return out.Bytes()
}

// buildCOFFSyso builds a Windows x64 COFF object file (.syso) containing .rsrc with RT_ICON + RT_GROUP_ICON.
func buildCOFFSyso(frames []iconFrame) []byte {
	type resLeaf struct {
		id   uint32
		data []byte
	}
	type resType struct {
		typeID uint32
		leaves []resLeaf
	}

	iconLeaves := make([]resLeaf, len(frames))
	for i, f := range frames {
		iconLeaves[i] = resLeaf{id: uint32(i + 1), data: f.png}
	}
	types := []resType{
		{typeID: 3, leaves: iconLeaves}, // RT_ICON
		{typeID: 14, leaves: []resLeaf{{id: 1, data: buildGroupIconData(frames)}}}, // RT_GROUP_ICON
	}

	// Build .rsrc section layout in two passes (directories -> data entries -> raw blobs).
	var rsrc bytes.Buffer
	writeDirHeader := func(numID uint16) {
		_ = binary.Write(&rsrc, binary.LittleEndian, uint32(0)) // Characteristics
		_ = binary.Write(&rsrc, binary.LittleEndian, uint32(0)) // TimeDateStamp
		_ = binary.Write(&rsrc, binary.LittleEndian, uint16(0)) // MajorVersion
		_ = binary.Write(&rsrc, binary.LittleEndian, uint16(0)) // MinorVersion
		_ = binary.Write(&rsrc, binary.LittleEndian, uint16(0)) // NumberOfNamedEntries
		_ = binary.Write(&rsrc, binary.LittleEndian, numID)     // NumberOfIdEntries
	}

	const dirHdrSize = 16
	const dirEntSize = 8
	const dataEntSize = 16

	// Calculate directory offsets
	rootSize := dirHdrSize + len(types)*dirEntSize
	typeDirOffsets := make([]uint32, len(types))
	cur := uint32(rootSize)
	for i, rt := range types {
		typeDirOffsets[i] = cur
		cur += uint32(dirHdrSize + len(rt.leaves)*dirEntSize)
	}

	langDirOffsets := make([][]uint32, len(types))
	totalLeaves := 0
	for i, rt := range types {
		langDirOffsets[i] = make([]uint32, len(rt.leaves))
		for j := range rt.leaves {
			langDirOffsets[i][j] = cur
			cur += uint32(dirHdrSize + dirEntSize)
			totalLeaves++
		}
	}

	dataEntryOffsets := make([][]uint32, len(types))
	for i, rt := range types {
		dataEntryOffsets[i] = make([]uint32, len(rt.leaves))
		for j := range rt.leaves {
			dataEntryOffsets[i][j] = cur
			cur += dataEntSize
		}
	}

	// Align payload start to 8 bytes
	for cur%8 != 0 {
		cur++
	}
	payloadOffsets := make([][]uint32, len(types))
	for i, rt := range types {
		payloadOffsets[i] = make([]uint32, len(rt.leaves))
		for j, lf := range rt.leaves {
			payloadOffsets[i][j] = cur
			cur += uint32(len(lf.data))
			for cur%4 != 0 {
				cur++
			}
		}
	}

	// Emit Level 1 (Root)
	writeDirHeader(uint16(len(types)))
	for i, rt := range types {
		_ = binary.Write(&rsrc, binary.LittleEndian, rt.typeID)
		_ = binary.Write(&rsrc, binary.LittleEndian, uint32(0x80000000)|typeDirOffsets[i])
	}

	// Emit Level 2 (IDs per type)
	for i, rt := range types {
		writeDirHeader(uint16(len(rt.leaves)))
		for j, lf := range rt.leaves {
			_ = binary.Write(&rsrc, binary.LittleEndian, lf.id)
			_ = binary.Write(&rsrc, binary.LittleEndian, uint32(0x80000000)|langDirOffsets[i][j])
		}
	}

	// Emit Level 3 (Language = 0x0409 en-US)
	for i, rt := range types {
		for j := range rt.leaves {
			writeDirHeader(1)
			_ = binary.Write(&rsrc, binary.LittleEndian, uint32(0x0409))
			_ = binary.Write(&rsrc, binary.LittleEndian, dataEntryOffsets[i][j])
		}
	}

	// Emit IMAGE_RESOURCE_DATA_ENTRY and collect relocations
	relocOffsets := make([]uint32, 0, totalLeaves)
	for i, rt := range types {
		for j, lf := range rt.leaves {
			relocOffsets = append(relocOffsets, uint32(rsrc.Len()))
			_ = binary.Write(&rsrc, binary.LittleEndian, payloadOffsets[i][j]) // relocated via ADDR32NB
			_ = binary.Write(&rsrc, binary.LittleEndian, uint32(len(lf.data)))
			_ = binary.Write(&rsrc, binary.LittleEndian, uint32(0)) // CodePage
			_ = binary.Write(&rsrc, binary.LittleEndian, uint32(0)) // Reserved
		}
	}

	// Pad to 8 bytes before payloads
	for rsrc.Len()%8 != 0 {
		rsrc.WriteByte(0)
	}
	for _, rt := range types {
		for _, lf := range rt.leaves {
			rsrc.Write(lf.data)
			for rsrc.Len()%4 != 0 {
				rsrc.WriteByte(0)
			}
		}
	}

	rsrcBytes := rsrc.Bytes()

	// Assemble COFF object file
	const (
		coffHdrSize = 20
		secHdrSize  = 40
		relocSize   = 10
	)
	rawDataOffset := uint32(coffHdrSize + secHdrSize)
	relocTableOffset := rawDataOffset + uint32(len(rsrcBytes))
	symTableOffset := relocTableOffset + uint32(len(relocOffsets)*relocSize)

	var obj bytes.Buffer
	// IMAGE_FILE_HEADER (20 bytes)
	_ = binary.Write(&obj, binary.LittleEndian, uint16(0x8664)) // IMAGE_FILE_MACHINE_AMD64
	_ = binary.Write(&obj, binary.LittleEndian, uint16(1))      // NumberOfSections
	_ = binary.Write(&obj, binary.LittleEndian, uint32(0))      // TimeDateStamp
	_ = binary.Write(&obj, binary.LittleEndian, symTableOffset) // PointerToSymbolTable
	_ = binary.Write(&obj, binary.LittleEndian, uint32(1))      // NumberOfSymbols
	_ = binary.Write(&obj, binary.LittleEndian, uint16(0))      // SizeOfOptionalHeader
	_ = binary.Write(&obj, binary.LittleEndian, uint16(0x0104)) // Characteristics

	// IMAGE_SECTION_HEADER (40 bytes)
	var secName [8]byte
	copy(secName[:], ".rsrc")
	obj.Write(secName[:])
	_ = binary.Write(&obj, binary.LittleEndian, uint32(0))                     // VirtualSize
	_ = binary.Write(&obj, binary.LittleEndian, uint32(0))                     // VirtualAddress
	_ = binary.Write(&obj, binary.LittleEndian, uint32(len(rsrcBytes)))        // SizeOfRawData
	_ = binary.Write(&obj, binary.LittleEndian, rawDataOffset)                 // PointerToRawData
	_ = binary.Write(&obj, binary.LittleEndian, relocTableOffset)              // PointerToRelocations
	_ = binary.Write(&obj, binary.LittleEndian, uint32(0))                     // PointerToLinenumbers
	_ = binary.Write(&obj, binary.LittleEndian, uint16(len(relocOffsets)))     // NumberOfRelocations
	_ = binary.Write(&obj, binary.LittleEndian, uint16(0))                     // NumberOfLinenumbers
	_ = binary.Write(&obj, binary.LittleEndian, uint32(0x40000040))            // INITIALIZED_DATA | MEM_READ

	// Section raw data
	obj.Write(rsrcBytes)

	// IMAGE_RELOCATION entries (10 bytes each)
	for _, off := range relocOffsets {
		_ = binary.Write(&obj, binary.LittleEndian, off)            // VirtualAddress
		_ = binary.Write(&obj, binary.LittleEndian, uint32(0))      // SymbolTableIndex (.rsrc)
		_ = binary.Write(&obj, binary.LittleEndian, uint16(0x0003)) // IMAGE_REL_AMD64_ADDR32NB
	}

	// IMAGE_SYMBOL (18 bytes) for ".rsrc"
	obj.Write(secName[:])
	_ = binary.Write(&obj, binary.LittleEndian, uint32(0)) // Value
	_ = binary.Write(&obj, binary.LittleEndian, int16(1))  // SectionNumber (1-based)
	_ = binary.Write(&obj, binary.LittleEndian, uint16(0)) // Type
	obj.WriteByte(3)                                       // StorageClass = IMAGE_SYM_CLASS_STATIC
	obj.WriteByte(0)                                       // NumberOfAuxSymbols

	// String table (empty -> 4-byte length field = 4)
	_ = binary.Write(&obj, binary.LittleEndian, uint32(4))

	return obj.Bytes()
}

func main() {
	sizes := []int{256, 64, 48, 32, 16}
	frames := make([]iconFrame, 0, len(sizes))

	if err := os.MkdirAll("assets", 0o755); err != nil {
		panic(err)
	}

	for _, sz := range sizes {
		img := renderIcon(sz)
		data, err := encodePNG(img)
		if err != nil {
			panic(err)
		}
		if sz == 256 {
			if err := os.WriteFile(filepath.Join("assets", "icon.png"), data, 0o644); err != nil {
				panic(err)
			}
		}
		frames = append(frames, iconFrame{size: sz, png: data})
	}

	ico := buildICO(frames)
	if err := os.WriteFile(filepath.Join("assets", "icon.ico"), ico, 0o644); err != nil {
		panic(err)
	}

	syso := buildCOFFSyso(frames)
	if err := os.WriteFile("rsrc_windows_amd64.syso", syso, 0o644); err != nil {
		panic(err)
	}

	fmt.Printf("Generated assets/icon.png, assets/icon.ico (%d bytes), rsrc_windows_amd64.syso (%d bytes)\n", len(ico), len(syso))
}
