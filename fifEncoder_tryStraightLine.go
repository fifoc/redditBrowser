package main

func tryStraightLine(x int, y int, data [][]*FifSegment, mask [][]bool) (int, []byte){ // Strategy 1: Straight line
	baseline := data[x][y]
	covered := 1
	cX := x + 1

	opCode := []byte{0x10, byte(x), byte(y), 0x1, data[x][y].ToByte()}

	for {
		if cX < len(data) {
			if data[cX][y].bg == baseline.bg && data[cX][y].fg == baseline.fg {
				if mask[cX][y] == true {
					covered++
					opCode[3] = byte(covered)
					opCode = append(opCode, data[cX][y].ToByte())
					cX++
				} else {
					break
				}
			} else {
				break
			}
		} else {
			break
		}
	}

	return covered, opCode
}

func updateMaskStraightLine(packet []byte, mask [][]bool) ( [][]bool, int ) {
	baseX, baseY, size := int(packet[1]), int(packet[2]), int(packet[3])

	for i := 0; i < size; i++ {
		mask[baseX + i][baseY] = false
	}

	return mask, size
}