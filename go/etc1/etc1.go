package etc1

const (
	EncodedBlockSize = 8
	DecodedBlockSize = 48
	PKMHeaderSize    = 16
)

const (
	RGB8OES = 0x8D64
)

var (
	modifierTable = [][]int{
		{2, 8, -2, -8},
		{5, 17, -5, -17},
		{9, 29, -9, -29},
		{13, 42, -13, -42},
		{18, 60, -18, -60},
		{24, 80, -24, -80},
		{33, 106, -33, -106},
		{47, 183, -47, -183},
	}

	lookup = []int{0, 1, 2, 3, -4, -3, -2, -1}
)

func clamp(x int) byte {
	if x < 0 {
		return 0
	}
	if x > 255 {
		return 255
	}
	return byte(x)
}

func convert4To8(b int) int {
	c := b & 0xf
	return (c << 4) | c
}

func convert5To8(b int) int {
	c := b & 0x1f
	return (c << 3) | (c >> 2)
}

func convert6To8(b int) int {
	c := b & 0x3f
	return (c << 2) | (c >> 4)
}

func divideBy255(d int) int {
	return (d + 128 + (d >> 8)) >> 8
}

func convert8To4(b int) int {
	c := b & 0xff
	return divideBy255(c * 15)
}

func convert8To5(b int) int {
	c := b & 0xff
	return divideBy255(c * 31)
}

func convertDiff(base, diff int) int {
	return convert5To8((0x1f & base) + lookup[0x7&diff])
}

func decodeSubblock(pOut []byte, r, g, b int, table []int, low uint32, second, flipped bool) {
	baseX, baseY := 0, 0
	if second {
		if flipped {
			baseY = 2
		} else {
			baseX = 2
		}
	}

	for i := 0; i < 8; i++ {
		var x, y int
		if flipped {
			x = baseX + (i >> 1)
			y = baseY + (i & 1)
		} else {
			x = baseX + (i >> 2)
			y = baseY + (i & 3)
		}

		k := y + (x * 4)
		offset := ((low >> k) & 1) | ((low >> (k + 15)) & 2)
		delta := table[offset]

		index := 3 * (x + 4*y)
		pOut[index] = clamp(r + delta)
		pOut[index+1] = clamp(g + delta)
		pOut[index+2] = clamp(b + delta)
	}
}

func DecodeBlock(pIn, pOut []byte) {
	high := uint32(pIn[0])<<24 | uint32(pIn[1])<<16 | uint32(pIn[2])<<8 | uint32(pIn[3])
	low := uint32(pIn[4])<<24 | uint32(pIn[5])<<16 | uint32(pIn[6])<<8 | uint32(pIn[7])

	var r1, r2, g1, g2, b1, b2 int
	if high&2 != 0 {
		rBase := high >> 27
		gBase := high >> 19
		bBase := high >> 11

		r1 = convert5To8(int(rBase))
		r2 = convertDiff(int(rBase), int(high>>24))
		g1 = convert5To8(int(gBase))
		g2 = convertDiff(int(gBase), int(high>>16))
		b1 = convert5To8(int(bBase))
		b2 = convertDiff(int(bBase), int(high>>8))
	} else {
		r1 = convert4To8(int(high >> 28))
		r2 = convert4To8(int(high >> 24))
		g1 = convert4To8(int(high >> 20))
		g2 = convert4To8(int(high >> 16))
		b1 = convert4To8(int(high >> 12))
		b2 = convert4To8(int(high >> 8))
	}

	tableIndexA := 7 & (high >> 5)
	tableIndexB := 7 & (high >> 2)
	tableA := modifierTable[tableIndexA]
	tableB := modifierTable[tableIndexB]
	flipped := (high & 1) != 0

	decodeSubblock(pOut, r1, g1, b1, tableA, low, false, flipped)
	decodeSubblock(pOut, r2, g2, b2, tableB, low, true, flipped)
}

type compressed struct {
	high  uint32
	low   uint32
	score uint32
}

func takeBest(a, b *compressed) {
	if a.score > b.score {
		*a = *b
	}
}

func averageColorsBlock(pIn []byte, inMask uint32, pColors []byte) {
	r, g, b := 0, 0, 0

	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			i := x + 4*y
			if inMask&(1<<uint(i)) != 0 {
				index := i * 3
				r += int(pIn[index])
				g += int(pIn[index+1])
				b += int(pIn[index+2])
			}
		}
	}

	pColors[0] = byte((r + 8) >> 4)
	pColors[1] = byte((g + 8) >> 4)
	pColors[2] = byte((b + 8) >> 4)
	pColors[3] = pColors[0]
	pColors[4] = pColors[1]
	pColors[5] = pColors[2]
}

func averageColorsSubblock(pIn []byte, inMask uint32, pColors []byte, flipped, second bool) {
	r, g, b := 0, 0, 0

	if flipped {
		by := 0
		if second {
			by = 2
		}
		for y := 0; y < 2; y++ {
			yy := by + y
			for x := 0; x < 4; x++ {
				i := x + 4*yy
				if inMask&(1<<uint(i)) != 0 {
					index := i * 3
					r += int(pIn[index])
					g += int(pIn[index+1])
					b += int(pIn[index+2])
				}
			}
		}
	} else {
		bx := 0
		if second {
			bx = 2
		}
		for y := 0; y < 4; y++ {
			for x := 0; x < 2; x++ {
				xx := bx + x
				i := xx + 4*y
				if inMask&(1<<uint(i)) != 0 {
					index := i * 3
					r += int(pIn[index])
					g += int(pIn[index+1])
					b += int(pIn[index+2])
				}
			}
		}
	}

	pColors[0] = byte((r + 4) >> 3)
	pColors[1] = byte((g + 4) >> 3)
	pColors[2] = byte((b + 4) >> 3)
}

func square(x int) int {
	return x * x
}

func chooseModifier(pBaseColors, pIn []byte, pLow *uint32, bitIndex int, pModifierTable []int) uint32 {
	bestScore := uint32(^uint32(0))
	bestIndex := 0

	pixelR := int(pIn[0])
	pixelG := int(pIn[1])
	pixelB := int(pIn[2])
	r := int(pBaseColors[0])
	g := int(pBaseColors[1])
	b := int(pBaseColors[2])

	for i := 0; i < 4; i++ {
		modifier := pModifierTable[i]
		decodedG := clamp(g + modifier)
		score := uint32(6 * square(int(decodedG)-pixelG))

		if score >= bestScore {
			continue
		}

		decodedR := clamp(r + modifier)
		score += uint32(3 * square(int(decodedR)-pixelR))

		if score >= bestScore {
			continue
		}

		decodedB := clamp(b + modifier)
		score += uint32(square(int(decodedB) - pixelB))

		if score < bestScore {
			bestScore = score
			bestIndex = i
		}
	}

	lowMask := (uint32((bestIndex>>1)<<16) | uint32(bestIndex&1)) << uint(bitIndex)
	*pLow |= lowMask

	return bestScore
}

func encodeSubblockHelper(pIn []byte, inMask uint32, pCompressed *compressed, flipped, second, bETC1S bool, pBaseColors []byte, pModifierTable []int) {
	score := pCompressed.score

	if bETC1S {
		for y := 0; y < 4; y++ {
			for x := 0; x < 4; x++ {
				i := x + 4*y
				if inMask&(1<<uint(i)) != 0 {
					index := i * 3
					score += chooseModifier(pBaseColors, pIn[index:index+3], &pCompressed.low, y+x*4, pModifierTable)
				}
			}
		}
	} else if flipped {
		by := 0
		if second {
			by = 2
		}
		for y := 0; y < 2; y++ {
			yy := by + y
			for x := 0; x < 4; x++ {
				i := x + 4*yy
				if inMask&(1<<uint(i)) != 0 {
					index := i * 3
					score += chooseModifier(pBaseColors, pIn[index:index+3], &pCompressed.low, yy+x*4, pModifierTable)
				}
			}
		}
	} else {
		bx := 0
		if second {
			bx = 2
		}
		for y := 0; y < 4; y++ {
			for x := 0; x < 2; x++ {
				xx := bx + x
				i := xx + 4*y
				if inMask&(1<<uint(i)) != 0 {
					index := i * 3
					score += chooseModifier(pBaseColors, pIn[index:index+3], &pCompressed.low, y+xx*4, pModifierTable)
				}
			}
		}
	}

	pCompressed.score = score
}

func inRange4bitSigned(color int) bool {
	return color >= -4 && color <= 3
}

func encodeBaseColors(pBaseColors, pColors []byte, pCompressed *compressed) {
	r51 := convert8To5(int(pColors[0]))
	g51 := convert8To5(int(pColors[1]))
	b51 := convert8To5(int(pColors[2]))
	r52 := convert8To5(int(pColors[3]))
	g52 := convert8To5(int(pColors[4]))
	b52 := convert8To5(int(pColors[5]))

	r1 := convert5To8(r51)
	g1 := convert5To8(g51)
	b1 := convert5To8(b51)
	var r2, g2, b2 int

	dr := r52 - r51
	dg := g52 - g51
	db := b52 - b51

	differential := inRange4bitSigned(dr) && inRange4bitSigned(dg) && inRange4bitSigned(db)

	if differential {
		r2 = convert5To8(r51 + dr)
		g2 = convert5To8(g51 + dg)
		b2 = convert5To8(b51 + db)

		pCompressed.high |= (uint32(r51) << 27) | (uint32(7&dr) << 24) | (uint32(g51) << 19) |
			(uint32(7&dg) << 16) | (uint32(b51) << 11) | (uint32(7&db) << 8) | 2
	} else {
		r41 := convert8To4(int(pColors[0]))
		g41 := convert8To4(int(pColors[1]))
		b41 := convert8To4(int(pColors[2]))
		r42 := convert8To4(int(pColors[3]))
		g42 := convert8To4(int(pColors[4]))
		b42 := convert8To4(int(pColors[5]))

		r1 = convert4To8(r41)
		g1 = convert4To8(g41)
		b1 = convert4To8(b41)
		r2 = convert4To8(r42)
		g2 = convert4To8(g42)
		b2 = convert4To8(b42)

		pCompressed.high |= (uint32(r41) << 28) | (uint32(r42) << 24) | (uint32(g41) << 20) |
			(uint32(g42) << 16) | (uint32(b41) << 12) | (uint32(b42) << 8)
	}

	pBaseColors[0] = byte(r1)
	pBaseColors[1] = byte(g1)
	pBaseColors[2] = byte(b1)
	pBaseColors[3] = byte(r2)
	pBaseColors[4] = byte(g2)
	pBaseColors[5] = byte(b2)
}

func encodeBlockHelper(pIn []byte, inMask uint32, pColors []byte, pCompressed *compressed, flipped, bETC1S bool) {
	pCompressed.score = 0xffffffff
	pCompressed.high = 0
	if flipped {
		pCompressed.high = 1
	}
	pCompressed.low = 0

	pBaseColors := make([]byte, 6)
	encodeBaseColors(pBaseColors, pColors, pCompressed)

	originalHigh := pCompressed.high

	for i := 0; i < 8; i++ {
		var temp compressed
		temp.score = 0
		temp.high = originalHigh | (uint32(i) << 5)
		if bETC1S {
			temp.high |= (uint32(i) << 2)
		}
		temp.low = 0

		encodeSubblockHelper(pIn, inMask, &temp, flipped, false, bETC1S, pBaseColors, modifierTable[i])
		takeBest(pCompressed, &temp)
	}

	if bETC1S {
		return
	}

	firstHalf := *pCompressed
	for i := 0; i < 8; i++ {
		var temp compressed
		temp.score = firstHalf.score
		temp.high = firstHalf.high | (uint32(i) << 2)
		temp.low = firstHalf.low

		encodeSubblockHelper(pIn, inMask, &temp, flipped, true, bETC1S, pBaseColors[3:], modifierTable[i])

		if i == 0 {
			*pCompressed = temp
		} else {
			takeBest(pCompressed, &temp)
		}
	}
}

func writeBigEndian(pOut []byte, d uint32) {
	pOut[0] = byte(d >> 24)
	pOut[1] = byte(d >> 16)
	pOut[2] = byte(d >> 8)
	pOut[3] = byte(d)
}

func EncodeBlock(pIn []byte, inMask uint32, pOut []byte, bETC1S bool) {
	colors := make([]byte, 6)
	flippedColors := make([]byte, 6)
	var a, b compressed

	if bETC1S {
		averageColorsBlock(pIn, inMask, colors)
		encodeBlockHelper(pIn, inMask, colors, &a, false, bETC1S)
	} else {
		averageColorsSubblock(pIn, inMask, colors, false, false)
		averageColorsSubblock(pIn, inMask, colors[3:], false, true)
		averageColorsSubblock(pIn, inMask, flippedColors, true, false)
		averageColorsSubblock(pIn, inMask, flippedColors[3:], true, true)

		encodeBlockHelper(pIn, inMask, colors, &a, false, bETC1S)
		encodeBlockHelper(pIn, inMask, flippedColors, &b, true, bETC1S)

		takeBest(&a, &b)
	}

	writeBigEndian(pOut, a.high)
	writeBigEndian(pOut[4:], a.low)
}

func GetEncodedDataSize(width, height uint32) uint32 {
	return (((width + 3) & ^uint32(3)) * ((height + 3) & ^uint32(3))) >> 1
}

func EncodeImage(pIn []byte, width, height, pixelSize, stride uint32, pOut []byte, bETC1S bool) int {
	if pixelSize < 2 || pixelSize > 3 {
		return -1
	}

	kYMask := []uint16{0x0, 0xf, 0xff, 0xfff, 0xffff}
	kXMask := []uint16{0x0, 0x1111, 0x3333, 0x7777, 0xffff}

	block := make([]byte, DecodedBlockSize)
	encoded := make([]byte, EncodedBlockSize)

	encodedWidth := (width + 3) & ^uint32(3)
	encodedHeight := (height + 3) & ^uint32(3)

	for y := uint32(0); y < encodedHeight; y += 4 {
		yEnd := height - y
		if yEnd > 4 {
			yEnd = 4
		}
		ymask := kYMask[yEnd]

		for x := uint32(0); x < encodedWidth; x += 4 {
			xEnd := width - x
			if xEnd > 4 {
				xEnd = 4
			}
			mask := ymask & kXMask[xEnd]

			for cy := uint32(0); cy < yEnd; cy++ {
				qIndex := (cy * 4) * 3
				pIndex := pixelSize*x + stride*(y+cy)

				if pixelSize == 3 {
					copy(block[qIndex:], pIn[pIndex:pIndex+xEnd*3])
				} else {
					for cx := uint32(0); cx < xEnd; cx++ {
						pixel := (int(pIn[pIndex+1]) << 8) | int(pIn[pIndex])
						block[qIndex] = byte(convert5To8(pixel >> 11))
						block[qIndex+1] = byte(convert6To8(pixel >> 5))
						block[qIndex+2] = byte(convert5To8(pixel))
						qIndex += 3
						pIndex += pixelSize
					}
				}
			}

			EncodeBlock(block, uint32(mask), encoded, bETC1S)
			copy(pOut, encoded)
			pOut = pOut[EncodedBlockSize:]
		}
	}

	return 0
}

func DecodeImage(pIn []byte, pOut []byte, width, height, pixelSize, stride uint32) int {
	if pixelSize < 2 || pixelSize > 3 {
		return -1
	}

	block := make([]byte, DecodedBlockSize)

	encodedWidth := (width + 3) & ^uint32(3)
	encodedHeight := (height + 3) & ^uint32(3)

	for y := uint32(0); y < encodedHeight; y += 4 {
		yEnd := height - y
		if yEnd > 4 {
			yEnd = 4
		}

		for x := uint32(0); x < encodedWidth; x += 4 {
			xEnd := width - x
			if xEnd > 4 {
				xEnd = 4
			}

			DecodeBlock(pIn, block)
			pIn = pIn[EncodedBlockSize:]

			for cy := uint32(0); cy < yEnd; cy++ {
				qIndex := (cy * 4) * 3
				pIndex := pixelSize*x + stride*(y+cy)

				if pixelSize == 3 {
					copy(pOut[pIndex:pIndex+xEnd*3], block[qIndex:qIndex+xEnd*3])
				} else {
					for cx := uint32(0); cx < xEnd; cx++ {
						r := block[qIndex]
						g := block[qIndex+1]
						b := block[qIndex+2]
						qIndex += 3

						pixel := ((int(r) >> 3) << 11) | ((int(g) >> 2) << 5) | (int(b) >> 3)
						pOut[pIndex] = byte(pixel)
						pOut[pIndex+1] = byte(pixel >> 8)
						pIndex += pixelSize
					}
				}
			}
		}
	}

	return 0
}

const (
	PKMFormatOffset        = 6
	PKMEncodedWidthOffset  = 8
	PKMEncodedHeightOffset = 10
	PKMWidthOffset         = 12
	PKMHeightOffset        = 14
)

const RGBNoMipmaps = 0

var magic = []byte{'P', 'K', 'M', ' ', '1', '0'}

func writeBEUint16(pOut []byte, data uint32) {
	pOut[0] = byte(data >> 8)
	pOut[1] = byte(data)
}

func readBEUint16(pIn []byte) uint32 {
	return uint32(pIn[0])<<8 | uint32(pIn[1])
}

func PKMFormatHeader(pHeader []byte, width, height uint32) {
	copy(pHeader, magic)
	encodedWidth := (width + 3) & ^uint32(3)
	encodedHeight := (height + 3) & ^uint32(3)

	writeBEUint16(pHeader[PKMFormatOffset:], RGBNoMipmaps)
	writeBEUint16(pHeader[PKMEncodedWidthOffset:], encodedWidth)
	writeBEUint16(pHeader[PKMEncodedHeightOffset:], encodedHeight)
	writeBEUint16(pHeader[PKMWidthOffset:], width)
	writeBEUint16(pHeader[PKMHeightOffset:], height)
}

func PKMIisValid(pHeader []byte) bool {
	if string(pHeader[:len(magic)]) != string(magic) {
		return false
	}

	format := readBEUint16(pHeader[PKMFormatOffset:])
	encodedWidth := readBEUint16(pHeader[PKMEncodedWidthOffset:])
	encodedHeight := readBEUint16(pHeader[PKMEncodedHeightOffset:])
	width := readBEUint16(pHeader[PKMWidthOffset:])
	height := readBEUint16(pHeader[PKMHeightOffset:])

	return format == RGBNoMipmaps &&
		encodedWidth >= width && encodedWidth-width < 4 &&
		encodedHeight >= height && encodedHeight-height < 4
}

func PKMGetWidth(pHeader []byte) uint32 {
	return readBEUint16(pHeader[PKMWidthOffset:])
}

func PKMGetHeight(pHeader []byte) uint32 {
	return readBEUint16(pHeader[PKMHeightOffset:])
}
