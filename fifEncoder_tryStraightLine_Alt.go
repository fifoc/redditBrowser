package main

func tryStraightLineAlt(x int, y int, data [][]*FifSegment, mask [][]bool) (int, []byte, []int){ // Strategy 1: Straight line
	baseline := data[x][y]
	covered := 1
	cX := x + 1

	opCode := []byte{0x10, byte(x), byte(y), 0x1, data[x][y].ToByte()}
	good := make([]int, 1)
	good[0] = x

	for {
		if cX < len(data) {
			if mask[cX][y] == true {
				covered++
				opCode[3] = byte(covered)
				opCode = append(opCode, data[cX][y].ToByte())
				if data[cX][y].bg == baseline.bg && data[cX][y].fg == baseline.fg {
					good = append(good, cX)
				}
				cX++
			} else {
				break
			}
		} else {
			break
		}
	}

	// Cut shit off
	maxLen := good[len(good) - 1] - x + 1
	opCode = opCode[:4 + maxLen]
	opCode[3] = byte(maxLen)

	return len(good), opCode, good
}

func updateMaskStraightLineAlt(packet []byte, mask [][]bool, good []int) ( [][]bool, int ) {
	_, baseY, size := int(packet[1]), int(packet[2]), int(packet[3])

	for i := 0; i < len(good); i++ {
		mask[good[i]][baseY] = false
	}

	return mask, size
}