package hid

// HID keyboard usage codes, page 0x07, for the bitmap of BuildKeyboardReport.
// Any usage below 240 works as a raw number; these are the common ones.
const (
	KeyA uint8 = 0x04
	KeyB uint8 = 0x05
	KeyC uint8 = 0x06
	KeyD uint8 = 0x07
	KeyE uint8 = 0x08
	KeyF uint8 = 0x09
	KeyG uint8 = 0x0A
	KeyH uint8 = 0x0B
	KeyI uint8 = 0x0C
	KeyJ uint8 = 0x0D
	KeyK uint8 = 0x0E
	KeyL uint8 = 0x0F
	KeyM uint8 = 0x10
	KeyN uint8 = 0x11
	KeyO uint8 = 0x12
	KeyP uint8 = 0x13
	KeyQ uint8 = 0x14
	KeyR uint8 = 0x15
	KeyS uint8 = 0x16
	KeyT uint8 = 0x17
	KeyU uint8 = 0x18
	KeyV uint8 = 0x19
	KeyW uint8 = 0x1A
	KeyX uint8 = 0x1B
	KeyY uint8 = 0x1C
	KeyZ uint8 = 0x1D

	Key1 uint8 = 0x1E
	Key2 uint8 = 0x1F
	Key3 uint8 = 0x20
	Key4 uint8 = 0x21
	Key5 uint8 = 0x22
	Key6 uint8 = 0x23
	Key7 uint8 = 0x24
	Key8 uint8 = 0x25
	Key9 uint8 = 0x26
	Key0 uint8 = 0x27

	KeyEnter     uint8 = 0x28
	KeyEsc       uint8 = 0x29
	KeyBackspace uint8 = 0x2A
	KeyTab       uint8 = 0x2B
	KeySpace     uint8 = 0x2C

	KeyMinus      uint8 = 0x2D
	KeyEqual      uint8 = 0x2E
	KeyLBracket   uint8 = 0x2F
	KeyRBracket   uint8 = 0x30
	KeyBackslash  uint8 = 0x31
	KeySemicolon  uint8 = 0x33
	KeyApostrophe uint8 = 0x34
	KeyGrave      uint8 = 0x35
	KeyComma      uint8 = 0x36
	KeyDot        uint8 = 0x37
	KeySlash      uint8 = 0x38
	KeyCapsLock   uint8 = 0x39

	KeyF1  uint8 = 0x3A
	KeyF2  uint8 = 0x3B
	KeyF3  uint8 = 0x3C
	KeyF4  uint8 = 0x3D
	KeyF5  uint8 = 0x3E
	KeyF6  uint8 = 0x3F
	KeyF7  uint8 = 0x40
	KeyF8  uint8 = 0x41
	KeyF9  uint8 = 0x42
	KeyF10 uint8 = 0x43
	KeyF11 uint8 = 0x44
	KeyF12 uint8 = 0x45

	KeyRight uint8 = 0x4F
	KeyLeft  uint8 = 0x50
	KeyDown  uint8 = 0x51
	KeyUp    uint8 = 0x52

	KeyLeftCtrl   uint8 = 0xE0
	KeyLeftShift  uint8 = 0xE1
	KeyLeftAlt    uint8 = 0xE2
	KeyLeftGUI    uint8 = 0xE3
	KeyRightCtrl  uint8 = 0xE4
	KeyRightShift uint8 = 0xE5
	KeyRightAlt   uint8 = 0xE6
	KeyRightGUI   uint8 = 0xE7
)

// Key is the usage code a character maps to, together with whether Shift must be
// held to produce it on a US layout.
type Key struct {
	Usage uint8
	Shift bool
}

// asciiToHID maps printable ASCII to its US-layout usage code. Built once at
// init rather than written out, so the letter and digit runs cannot drift.
var asciiToHID = func() map[rune]Key {
	table := make(map[rune]Key, 128)

	for i := rune(0); i < 26; i++ {
		usage := KeyA + uint8(i)
		table['a'+i] = Key{Usage: usage}
		table['A'+i] = Key{Usage: usage, Shift: true}
	}
	// Usages run 1..9 then 0, which is why '0' is handled after the loop.
	for i := rune(0); i < 9; i++ {
		table['1'+i] = Key{Usage: Key1 + uint8(i)}
	}
	table['0'] = Key{Usage: Key0}

	// Shifted digits.
	for ch, usage := range map[rune]uint8{
		'!': Key1, '@': Key2, '#': Key3, '$': Key4, '%': Key5,
		'^': Key6, '&': Key7, '*': Key8, '(': Key9, ')': Key0,
	} {
		table[ch] = Key{Usage: usage, Shift: true}
	}

	for ch, key := range map[rune]Key{
		' ':  {Usage: KeySpace},
		'\t': {Usage: KeyTab},
		'\n': {Usage: KeyEnter},
		'-':  {Usage: KeyMinus},
		'_':  {Usage: KeyMinus, Shift: true},
		'=':  {Usage: KeyEqual},
		'+':  {Usage: KeyEqual, Shift: true},
		'[':  {Usage: KeyLBracket},
		'{':  {Usage: KeyLBracket, Shift: true},
		']':  {Usage: KeyRBracket},
		'}':  {Usage: KeyRBracket, Shift: true},
		'\\': {Usage: KeyBackslash},
		'|':  {Usage: KeyBackslash, Shift: true},
		';':  {Usage: KeySemicolon},
		':':  {Usage: KeySemicolon, Shift: true},
		'\'': {Usage: KeyApostrophe},
		'"':  {Usage: KeyApostrophe, Shift: true},
		'`':  {Usage: KeyGrave},
		'~':  {Usage: KeyGrave, Shift: true},
		',':  {Usage: KeyComma},
		'<':  {Usage: KeyComma, Shift: true},
		'.':  {Usage: KeyDot},
		'>':  {Usage: KeyDot, Shift: true},
		'/':  {Usage: KeySlash},
		'?':  {Usage: KeySlash, Shift: true},
	} {
		table[ch] = key
	}

	return table
}()

// KeyForRune returns the usage code and shift state that produce ch on a US
// keyboard layout. The second result is false for characters with no mapping.
func KeyForRune(ch rune) (Key, bool) {
	key, ok := asciiToHID[ch]
	return key, ok
}
