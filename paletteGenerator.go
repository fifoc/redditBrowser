package main

func generatePalette() ([]uint32, map[uint32]bool) {
	red := []uint32{0x00, 0x33, 0x66, 0x99, 0xCC, 0xFF}
	green := []uint32{0x00, 0x24, 0x49, 0x6D, 0x92, 0xB6, 0xDB, 0xFF}
	blue := []uint32{0x00, 0x40, 0x80, 0xC0, 0xFF}
	gray := []uint32{0x0F, 0x1E, 0x2D, 0x3C, 0x4B, 0x5A, 0x69, 0x78, 0x87, 0x96, 0xA5, 0xB4, 0xC3, 0xD2, 0xE1, 0xF0}

	var palette []uint32
	var palMap = make(map[uint32]bool)

	for r := 0; r < len(red); r++ {
		for g := 0; g < len(green); g++ {
			for b := 0; b < len(blue); b++ {
				color := red[r] * 0x10000 + green[g] * 0x100 + blue[b]
				palette = append(palette, color)
				palMap[color] = true
			}
		}
	}

	for g := 0; g < len(gray); g++ {
		color := gray[g] * 0x10000 + gray[g] * 0x100 + gray[g]
		palette = append(palette, color)
		palMap[color] = true
	}

	return palette, palMap
}
