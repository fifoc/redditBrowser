package main

func writeSetBg(bg uint32) []byte {
	r := byte(bg >> 16)
	g := byte((bg >> 8) & 0xFF)
	b := byte(bg & 0xFF)

	opcode := []byte{0x01, r, g, b}

	return opcode
}

func writeSetFg(bg uint32) []byte {
	r := byte(bg >> 16)
	g := byte((bg >> 8) & 0xFF)
	b := byte(bg & 0xFF)

	opcode := []byte{0x02, r, g, b}

	return opcode
}