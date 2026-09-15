package hid

import (
	"encoding/binary"
	"time"
)

const (
	reportIDTouchscreen = 0x09
)

// The device rejects a touchscreen report of any other size.
const touchscreenReportLen = 58

// timestampBits is the width of the report timestamp field. Six bytes on the
// wire, so values are masked to 48 bits.
const timestampBits = 48

type touchState uint8

const (
	touchContact touchState = 0xC2
	touchRelease touchState = 0x02
)

// processStart anchors timestamp to a monotonic origin. time.Since reads Go's
// monotonic clock, so the sequence is unaffected by wall-clock adjustments.
var processStart = time.Now()

// timestamp returns the value to stamp a report with. Only monotonicity and the
// deltas between reports are read from it, not absolute time.
func timestamp() uint64 {
	return uint64(time.Since(processStart).Nanoseconds())
}

// putTimestamp writes the low 48 bits, because the slot is six bytes and a
// wider value would overwrite the reserved bytes that follow it.
func putTimestamp(report []byte, ts uint64) {
	var full [8]byte
	binary.LittleEndian.PutUint64(full[:], ts&(1<<timestampBits-1))
	copy(report, full[:6])
}

// buildTouchscreenReport carries a contact state and position.
// Layout: [0]=report ID, [1:3]=constants 0x01 0x05, [3]=state, [4:6]=X uint16 LE,
// [6:8]=Y uint16 LE, [8:40] reserved, [40:44]=constant 0x02 0x00 0x00 0x00,
// [44:50]=timestamp, [50:58] reserved.
//
// A contact at x=0x1234 y=0x5678 with timestamp 0xa1b2c3d4:
//
//	09 0105 c2 3412 7856 00..00 02000000 d4c3b2a10000 00..00
//	ID const st x    y    res    const    timestamp    res
func buildTouchscreenReport(state touchState, x, y uint16, ts uint64) []byte {
	report := make([]byte, touchscreenReportLen)
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
