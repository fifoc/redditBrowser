package main

func tryVerticalLine(x int, y int, data [][]*FifSegment, mask [][]bool) (int, []byte){ // Strategy 2: Vertical line
	baseline := data[x][y]
	covered := 1
	cY := y + 1

	opCode := []byte{0x13, byte(x), byte(y), 0x1, data[x][y].ToByte()}

	for {
		if cY < len(data[0]) {
			if data[x][cY].bg == baseline.bg && data[x][cY].fg == baseline.fg {
				if mask[x][cY] == true {
					covered++
					opCode[3] = byte(covered)
					opCode = append(opCode, data[x][cY].ToByte())
					cY++
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

func updateMaskVerticalLine(packet []byte, mask [][]bool)  ( [][]bool, int ) {
	baseX, baseY, size := int(packet[1]), int(packet[2]), int(packet[3])

	for i := 0; i < size; i++ {
		mask[baseX][baseY + i] = false
	}

	return mask, size
}