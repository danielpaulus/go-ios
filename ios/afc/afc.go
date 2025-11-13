package afc

const (
	magic      uint64 = 0x4141504c36414643
	headerSize uint64 = 40
)

type opcode uint64

const (
	status                opcode = 0x00000001
	readDir               opcode = 0x00000003
	removePath            opcode = 0x00000008
	makeDir               opcode = 0x00000009
	fileInfo              opcode = 0x0000000A
	deviceInfo            opcode = 0x0000000B
	fileOpen              opcode = 0x0000000D
	fileClose             opcode = 0x00000014
	fileWrite             opcode = 0x00000010
	fileOpenResult        opcode = 0x0000000E
	fileRead              opcode = 0x0000000F
	removePathAndContents opcode = 0x00000022
)
