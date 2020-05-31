package main

/*
	Explaination of this encoding strategy:
		Try how wide you can go with all heights of square that contains the exact same char and exact same color
		First try height 1, then 2 and so on
		Find the most optimal pixel coverage and return

	NOTE: unlike other strategies, this one requires the CHARACTER to be the same in addition to BG and FG!!

	This has the potential to be very fast in images with a lot of whitespace.
 */

func sqPixelEquality(a *FifSegment, b *FifSegment) bool {
	return (a.bg == b.bg) && (a.fg == b.fg) && (a.ToByte() == b.ToByte())
}

func analyseSquare(x int, y int, w int, h int, data [][]*FifSegment, mask [][]bool) bool {
	w = w + 1
	h = h + 1
	for xx := 0; xx < w; xx++ {
		for yy := 0; yy < h; yy++ {
			if sqPixelEquality(data[x][y], data[x + xx][y + yy]) == true && mask[x + xx][y + yy] == true {
				continue
			} else {
				return false
			}
		}
	}

	return true
}

func trySquare(x int, y int, data [][]*FifSegment, mask [][]bool) (int, []byte) { // Strategy 3: Square
	bestCoverage := 0
	bestOpcode := []byte{0x11, byte(x), byte(y), 1, 1, data[x][y].ToByte()}

	baseLine := data[x][y]

	www := len(data)
	hhh := len(data[0])

	for w := 0; w < www; w++ {
		if x + w > (www - 1) {break}
		if sqPixelEquality(baseLine, data[x + w][y]) == false {break}
		if mask[x + w][y] == false { break }

		for h := 0; h < hhh; h++ {
			if y + h > (hhh - 1) {break}
			if sqPixelEquality(baseLine, data[x + w][y + h]) == false {break}
			if mask[x + w][y + h] == false {break}

			if analyseSquare(x, y, w, h, data, mask) == true {
				bestCoverage = (w + 1) * (h + 1)
				bestOpcode[3] = byte(w) + 1
				bestOpcode[4] = byte(h) + 1
			} else {
				break
			}
		}
	}

	return bestCoverage, bestOpcode
}

func updateMaskSquare(opcode []byte, mask [][]bool)  ( [][]bool, int ) {
	baseX, baseY, sizeX, sizeY := int(opcode[1]), int(opcode[2]), int(opcode[3]), int(opcode[4])

	for i := 0; i < sizeX; i++ {
		for j := 0; j < sizeY; j++ {
			mask[baseX+i][baseY+j] = false
		}
	}

	return mask, sizeX * sizeY
}