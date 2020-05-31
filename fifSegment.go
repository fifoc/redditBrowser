package main

import (
	"math"
)

type FifSegment struct {
	bg uint32
	fg uint32
	data [8]bool
}

func boolToByte(a bool) byte {
	if a == true {
		return 1
	} else {
		return 0
	}
}

func (f *FifSegment) ToByte() byte {
	a := byte(0)
	if f.data[0] == true {a += 0x01}
	if f.data[1] == true {a += 0x02}
	if f.data[2] == true {a += 0x04}
	if f.data[3] == true {a += 0x08}
	if f.data[4] == true {a += 0x10}
	if f.data[5] == true {a += 0x20}
	if f.data[6] == true {a += 0x40}
	if f.data[7] == true {a += 0x80}

	return a
}

type NonFinalFifSegment struct {
	data [8]uint32
}

// 0 3
// 1 4
// 2 5
// 6 7

type ColorImportance struct {
	luma uint64
	occurences byte
}

func (n *NonFinalFifSegment) Set(x int, y int, col uint32) {
	pos := 0
	switch y {
	case 0:
		pos = 0
		if x == 1 { pos = 3 }
		break
	case 1:
		pos = 1
		if x == 1 { pos = 4 }
		break
	case 2:
		pos = 2
		if x == 1 { pos = 5 }
		break
	case 3:
		pos = 6
		if x == 1 { pos = 7 }
	}
	n.data[pos] = col
}

func intAbs(a uint64, b uint64) uint64 {
	if a > b {
		return a - b
	} else {
		return b - a
	}
}

func (n *NonFinalFifSegment) ToFinalFifSegment() *FifSegment {
	a := new(FifSegment)

	// Go over every color in the current segment and calculate its significance.
	colorsInSegment := 0
	colorSet := make(map[uint32]bool)
	colorA := make([]uint32, 0)

	for i := 0; i < 8; i++ {
		col := n.data[i]
		if _, ok := colorSet[col]; ok != true {
			colorSet[col] = true
			colorsInSegment++
			colorA = append(colorA, col)
		}
	}

	if colorsInSegment < 1 || colorsInSegment > 8 {
		panic("A CATASTROPHIC error has occured.")
	}

	// Special cases
	if colorsInSegment == 1 {
		a.bg = colorA[0]
		a.fg = colorA[0]
		a.data = [8]bool{false, false, false, false, false, false, false, false}
	}
	if colorsInSegment == 2 {
		a.bg = colorA[0]
		a.fg = colorA[1]
		if a.bg > a.fg {
			a.bg, a.fg = a.fg, a.bg
		}
		for i := 0; i < 8; i++ {
			if n.data[i] == a.fg {
				a.data[i] = true
			}
		}
	}
	if colorsInSegment > 2 {
		// Oh man, the actually difficult part of my life.
		colorOccurences := make(map[uint32]uint64)
		for i := 0; i < 8; i++ {
			if _, ok := colorOccurences[n.data[i]]; ok != true {
				colorOccurences[n.data[i]] = 1
			} else {
				colorOccurences[n.data[i]]++
			}
		}

		// Now we find the smallest luma
		smallestLuma, slc := uint64(math.MaxUint64), uint32(0)
		largestLuma, llc := uint64(0), uint32(0)

		for color /*occurences */:= range colorOccurences {
			slum := lumaCache[color] /*/ occurences*/
			llum := lumaCache[color] /** occurences*/

			if slum < smallestLuma {
				smallestLuma, slc = slum, color
			}

			if llum > largestLuma {
				largestLuma, llc = llum, color
			}
		}

		a.bg = slc
		a.fg = llc

		for i := 0; i < 8; i++ {
			col := n.data[i]
			if col == slc {
				// noop
			} else if col == llc {
				a.data[i] = true
			} else {
				sdelta := intAbs(lumaCache[col], smallestLuma)
				ldelta := intAbs(lumaCache[col], largestLuma)

				if ldelta < sdelta {
					a.data[i] = true
				}
			}
		}
	}

	return a
}