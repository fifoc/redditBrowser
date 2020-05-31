package main

func tryVerticalLineAlt(x int, y int, data [][]*FifSegment, mask [][]bool) (int, []byte, []int){ // Strategy 1: Straight line
	baseline := data[x][y]
	covered := 1
	cY := y + 1

	opCode := []byte{0x13, byte(x), byte(y), 0x1, data[x][y].ToByte()}
	good := make([]int, 1)
	good[0] = y

	for {
		if cY < len(data[0]) {
			if mask[x][cY] == true {
				covered++
				opCode[3] = byte(covered)
				opCode = append(opCode, data[x][cY].ToByte())
				if data[x][cY].bg == baseline.bg && data[x][cY].fg == baseline.fg {
					good = append(good, cY)
				}
				cY++
			} else {
				break
			}
		} else {
			break
		}
	}

	maxLen := good[len(good) - 1] - y + 1
	opCode = opCode[:4 + maxLen]
	opCode[3] = byte(maxLen)

	return len(good), opCode, good
}

func updateMaskVerticalLineAlt(packet []byte, mask [][]bool, good []int) ( [][]bool, int ) {
	baseX,_, size := int(packet[1]), int(packet[2]), int(packet[3])

	for i := 0; i < len(good); i++ {
		mask[baseX][good[i]] = false
	}

	return mask, size
}