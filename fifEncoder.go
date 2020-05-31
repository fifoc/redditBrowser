package main

import (
	"image"
	"math"
	"sort"
	"sync"
)

var simpCache = make(map[uint32]uint32)
var simpMutex sync.Mutex

func simplifyColor(palette []uint32, color uint32) uint32 {
	simpMutex.Lock()
	if _, ok := simpCache[color]; ok == true {
		simpMutex.Unlock()
		return simpCache[color]
	}
	simpMutex.Unlock()
	var closestDelta = int64(math.MaxInt64)
	var pick uint32
	r, g, b := splitColor(color)

	for i := 0; i < len(palette); i++ {
		col := palette[i]
		if color == col {
			simpMutex.Lock()
			simpCache[color] = col
			simpMutex.Unlock()
			return col
		} else {
			pR, pG, pB := splitColor(col)
			factorR := int64(int(pR) - int(r))
			factorG := int64(int(pG) - int(g))
			factorB := int64(int(pB) - int(b))
			delta := (factorR * factorR) + (factorG * factorG) + (factorB * factorB)
			if delta < closestDelta {
				closestDelta = delta
				pick = col
			}
		}
	}
	simpMutex.Lock()
	simpCache[color] = pick
	simpMutex.Unlock()
	return pick
}

type CAEntry struct {
	color   uint64
	entries int
}

func encodeFif(w int, h int, src image.Image, dontSimplify bool) []byte {
	mimage := make([][]uint32, w)
	imask := make([][]bool, w) // Pixels set to false in the imask will no longer be rendered.
	// Later on when converting to characters, when even 1 character exists
	// The imask for it is true.

	fif := []byte("FastIF")
	fif = append(fif, byte(w/2), byte(h/4))

	for x := 0; x < w; x++ {
		mimage[x] = make([]uint32, h)
		imask[x] = make([]bool, h)
		for y := 0; y < h; y++ {
			r, g, b, a := src.At(x, y).RGBA()
			r = r / 257
			g = g / 257
			b = b / 257

			if a > 8 {
				zcolor := ((r & 0xFF) << 16) + ((g & 0xFF) << 8) + b&0xFF
				if dontSimplify == false {
					mimage[x][y] = simplifyColor(palette, zcolor)
				} else {
					mimage[x][y] = zcolor
				}
				imask[x][y] = true
			} else {
				imask[x][y] = false
			}
		}
	}

	// Encode the whole thing as fif segments
	fifParts := make([][]NonFinalFifSegment, w/2)
	for i := 0; i < (w / 2); i++ {
		fifParts[i] = make([]NonFinalFifSegment, h/4)
	}

	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {
			fifParts[x/2][y/4].Set(x%2, y%4, mimage[x][y])
		}
	}

	//	fmt.Println(fifParts)

	// Simplify the fif segments
	// TODO: This can be multithreaded!
	sFifParts := make([][]*FifSegment, w/2)
	for i := 0; i < (w / 2); i++ {
		sFifParts[i] = make([]*FifSegment, h/4)
		for j := 0; j < (h / 4); j++ {
			sFifParts[i][j] = fifParts[i][j].ToFinalFifSegment()
		}
	}

	maskParts := make([][]bool, w/2)
	maskCount := 0
	for i := 0; i < (w / 2); i++ {
		maskParts[i] = make([]bool, h/4)

		for j := 0; j < (h / 4); j++ {
			maskParts[i][j] = false
			for x := 0; x < 2; x++ {
				for y := 0; y < 4; y++ {
					if imask[(i*2)+x][(j*4)+y] == true {
						maskParts[i][j] = true
					}
				}
			}
			if maskParts[i][j] == true {
				maskCount++
			}
		}
	}

	colorCombos := make(map[uint64]int)

	for i := 0; i < (w / 2); i++ {
		for j := 0; j < (h / 4); j++ {
			combo := uint64(sFifParts[i][j].fg) << 24
			combo = combo + uint64(sFifParts[i][j].bg)
			if _, ok := colorCombos[combo]; ok == false {
				colorCombos[combo] = 0
			}
			colorCombos[combo] = colorCombos[combo] + 1
		}
	}

	caArray := make([]CAEntry, 0)

	for color, occurences := range colorCombos {
		caArray = append(caArray, CAEntry{
			color:   color,
			entries: occurences,
		})
	}

	sort.Slice(caArray, func(i, j int) bool {
		return caArray[i].entries > caArray[j].entries
	})

	oldbg := uint32(math.MaxUint32)
	oldfg := uint32(math.MaxUint32)

	for c := 0; c < len(caArray); c++ {
		color := caArray[c].color
		bg := uint32(color & 0xFFFFFF)
		fg := uint32((color >> 24) & 0xFFFFFF)

		// Write set bg and set fg opcodes
		if bg != oldbg {
			fif = append(fif, writeSetBg(bg)...)
			oldbg = bg
		}
		if fg != oldfg {
			fif = append(fif, writeSetFg(fg)...)
			oldfg = fg
		}

		for y := 0; y < (h / 4); y++ {
			for x := 0; x < (w / 2); x++ {
				// First, check if not masked away.
				if maskParts[x][y] == false {
					continue
				}

				// Compare colors
				if sFifParts[x][y].bg != bg || sFifParts[x][y].fg != fg {
					continue
				}

				if FIF_altMode == false {
					// Test all 3 cases
					coveredSl, packetSl := tryStraightLine(x, y, sFifParts, maskParts)
					coveredVl, packetVl := tryVerticalLine(x, y, sFifParts, maskParts)
					coveredSq, packetSq := trySquare(x, y, sFifParts, maskParts)

					greatest, gV := 0, 0

					if coveredSl > gV {
						greatest, gV = 0, coveredSl
					}
					if coveredVl > gV {
						greatest, gV = 1, coveredVl
					}
					if coveredSq > gV {
						greatest, gV = 2, coveredSq
					}

					if greatest == 0 {
						fif = append(fif, packetSl...)
						maskParts, _ = updateMaskStraightLine(packetSl, maskParts)
					}

					if greatest == 1 {
						fif = append(fif, packetVl...)
						maskParts, _ = updateMaskVerticalLine(packetVl, maskParts)
					}

					if greatest == 2 {
						fif = append(fif, packetSq...)
						maskParts, _ = updateMaskSquare(packetSq, maskParts)
					}
				} else {
					// Test all 3 cases
					coveredASl, packetASl, goodSl := tryStraightLineAlt(x, y, sFifParts, maskParts)
					coveredAVl, packetAVl, goodVl := tryVerticalLineAlt(x, y, sFifParts, maskParts)
					coveredSl, packetSl := tryStraightLine(x, y, sFifParts, maskParts)
					coveredVl, packetVl := tryVerticalLine(x, y, sFifParts, maskParts)
					coveredSq, packetSq := trySquare(x, y, sFifParts, maskParts)
					greatest, gV := 0, 0

					if coveredSl > gV {
						greatest, gV = 0, coveredSl
					}
					if coveredVl > gV {
						greatest, gV = 1, coveredVl
					}
					if coveredSq > gV {
						greatest, gV = 2, coveredSq
					}
					if coveredASl > gV {
						greatest, gV = 3, coveredASl
					}
					if coveredAVl > gV {
						greatest, gV = 4, coveredAVl
					}

					if greatest == 0 {
						fif = append(fif, packetSl...)
						maskParts, _ = updateMaskStraightLine(packetSl, maskParts)
					}

					if greatest == 1 {
						fif = append(fif, packetVl...)
						maskParts, _ = updateMaskVerticalLine(packetVl, maskParts)
					}

					if greatest == 2 {
						fif = append(fif, packetSq...)
						maskParts, _ = updateMaskSquare(packetSq, maskParts)
					}
					if greatest == 3 {
						fif = append(fif, packetASl...)
						maskParts, _ = updateMaskStraightLineAlt(packetASl, maskParts, goodSl)
					}
					if greatest == 4 {
						fif = append(fif, packetAVl...)
						maskParts, _ = updateMaskVerticalLineAlt(packetAVl, maskParts, goodVl)
					}
				}
			}
		}
	}

	fif = append(fif, 0x20)

	return fif
}
