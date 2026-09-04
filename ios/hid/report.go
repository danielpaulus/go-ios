package hid

import (
	"encoding/binary"
	"time"
)

const (
	reportIDKeyboard    = 0x01
	reportIDTouchscreen = 0x09
	reportIDDigitizer   = 0x13
)

// Report lengths. The device rejects reports of any other size for these surfaces.
const (
	DigitizerReportLen   = 19
	TouchscreenReportLen = 58
	KeyboardReportLen    = 39
)

// timestampBits is the width of the report timestamp field. Six bytes on the
// wire, so values are masked to 48 bits.
const timestampBits = 48

type TouchState uint8

const (
	TouchContact TouchState = 0xC2
	TouchRelease TouchState = 0x02
)

const keyboardBitmapLen = 30

const maxKeyboardUsage = keyboardBitmapLen * 8 // 240

// processStart anchors Timestamp to a monotonic origin. time.Since reads Go's
// monotonic clock, so the sequence is unaffected by wall-clock adjustments.
var processStart = time.Now()

// Timestamp returns the 48-bit value to stamp a report with. Only monotonicity
// and the deltas between reports are read from it, not absolute time.
func Timestamp() uint64 {
	return uint64(time.Since(processStart).Nanoseconds()) & (1<<timestampBits - 1)
}

func putTimestamp(report []byte, ts uint64) {
	var full [8]byte
	binary.LittleEndian.PutUint64(full[:], ts&(1<<timestampBits-1))
	copy(report, full[:6])
}

// BuildDigitizerReport places the pointer at (x, y), signed 32-bit.
// Layout: [0]=report ID, [1:5]=X int32 LE, [5:9]=Y int32 LE, [9:11] reserved,
// [11:17]=timestamp, [17:19] reserved.
//
// x=100 y=200 with timestamp 0xa1b2c3d4:
//
//	13 64000000 c8000000 0000 d4c3b2a10000 0000
//	ID x=100    y=200    res  timestamp    res
func BuildDigitizerReport(x, y int32, ts uint64) []byte {
	report := make([]byte, DigitizerReportLen)
	report[0] = reportIDDigitizer
	binary.LittleEndian.PutUint32(report[1:5], uint32(x))
	binary.LittleEndian.PutUint32(report[5:9], uint32(y))
	putTimestamp(report[11:17], ts)
	return report
}

// BuildTouchscreenReport carries a contact state and position.
// Layout: [0]=report ID, [1:3]=constants 0x01 0x05, [3]=state, [4:6]=X uint16 LE,
// [6:8]=Y uint16 LE, [8:40] reserved, [40:44]=constant 0x02 0x00 0x00 0x00,
// [44:50]=timestamp, [50:58] reserved.
//
// A contact at x=0x1234 y=0x5678 with timestamp 0xa1b2c3d4:
//
//	09 0105 c2 3412 7856 00..00 02000000 d4c3b2a10000 00..00
//	ID const st x    y    res    const    timestamp    res
func BuildTouchscreenReport(state TouchState, x, y uint16, ts uint64) []byte {
	report := make([]byte, TouchscreenReportLen)
	report[0] = reportIDTouchscreen
	report[1] = 0x01
	report[2] = 0x05
	report[3] = uint8(state)
	binary.LittleEndian.PutUint16(report[4:6], x)
	binary.LittleEndian.PutUint16(report[6:8], y)
	report[40] = 0x02
	putTimestamp(report[44:50], ts)
	return report
}

// BuildKeyboardReport takes the full set of usages held down, not a delta.
// Usages at or above 240 are ignored. Layout: [0]=report ID, [1:31]=240-bit usage
// bitmap, [31:37]=timestamp, [37:39] reserved.
//
// Usage 4, the letter a, with timestamp 0xa1b2c3d4. Bit 4 lands in byte 1 of the
// bitmap, which is why the second byte is 0x10:
//
//	01 10 00..00 d4c3b2a10000 0000
//	ID bitmap    timestamp    res
func BuildKeyboardReport(usages []uint8, ts uint64) []byte {
	report := make([]byte, KeyboardReportLen)
	report[0] = reportIDKeyboard
	for _, usage := range usages {
		if usage >= maxKeyboardUsage {
			continue
		}
		report[1+usage/8] |= 1 << (usage % 8)
	}
	putTimestamp(report[31:37], ts)
	return report
}
