package debuggertools

import (
	"fmt"
	"math"
	"strings"

	"github.com/danielpaulus/go-ios/ios/debugserver"
	log "github.com/sirupsen/logrus"
)

// ARM64 register numbers (from debugserver's qRegisterInfo numbering)
const (
	regX0 = 0  // first argument / return value
	regLR = 30 // link register (return address)
	regPC = 32 // program counter

	// NEON/FP registers v0-v7 for float/double arguments.
	// GPR set has 63 registers (x0-x28, fp, lr, sp, pc, cpsr, w0-w28),
	// then VFP set starts at index 63 with v0-v31.
	regV0 = 63
)

// dlopen flags
const (
	rtldNow = 2 // RTLD_NOW
)

// RTLD_DEFAULT = ((void*)-2) on Apple platforms.
const rtldDefault = ^uint64(0) - 1

const (
	arm64BrkInstruction = "000020d4" // brk #0 (little-endian)
	arm64BrkKind        = 4          // breakpoint size for Z0 packet
	dataPageSize        = 0x10000    // rw memory for strings/data
	codePageSize        = 0x100      // rwx memory for brk trap
	memReadChunkSize    = 0x4000     // max bytes per GDB 'm' request
)

// objCRuntime provides a high-level interface for calling ObjC methods
// in a remote process via GDB RSP.
type objCRuntime struct {
	mem      *gdbMem
	dlsym    uint64 // address of dlsym()
	dlopen   uint64 // address of dlopen()
	msgSend  uint64 // address of objc_msgSend()
	getClass uint64 // address of objc_getClass()
	selReg   uint64 // address of sel_registerName()
	selCache map[string]uint64
	clsCache map[string]uint64
}

// newObjCRuntime bootstraps an ObjC runtime interface by:
//  1. Saving register state for later restoration
//  2. Allocating memory for a brk trap and data
//  3. Finding dlsym via Mach-O export trie
//  4. Using dlsym to resolve dlopen, objc_msgSend, etc.
func newObjCRuntime(gdb *debugserver.GDBServer) (*objCRuntime, error) {
	mem, err := newGDBMem(gdb)
	if err != nil {
		return nil, err
	}

	dlsymAddrs, err := resolveSymbols(gdb, []symbolQuery{
		{"libdyld.dylib", "dlsym"},
	})
	if err != nil {
		mem.cleanup()
		return nil, fmt.Errorf("resolve dlsym: %w", err)
	}
	dlsymAddr := dlsymAddrs[0]

	symbolNames := []string{"dlopen", "objc_msgSend", "objc_getClass", "sel_registerName"}
	resolved := make([]uint64, len(symbolNames))
	for i, name := range symbolNames {
		nameAddr, _ := mem.writeCString(name)
		addr, err := mem.call(dlsymAddr, rtldDefault, nameAddr)
		if err != nil || addr == 0 {
			mem.cleanup()
			return nil, fmt.Errorf("dlsym(%s) failed", name)
		}
		resolved[i] = addr
		log.WithField("symbol", name).WithField("addr", fmt.Sprintf("0x%x", addr)).Debug("Resolved via dlsym")
	}

	return &objCRuntime{
		mem:      mem,
		dlsym:    dlsymAddr,
		dlopen:   resolved[0],
		msgSend:  resolved[1],
		getClass: resolved[2],
		selReg:   resolved[3],
		selCache: make(map[string]uint64),
		clsCache: make(map[string]uint64),
	}, nil
}

func (rt *objCRuntime) cleanup() {
	rt.mem.cleanup()
}

// CString writes a null-terminated C string to remote memory and returns its address.
func (rt *objCRuntime) CString(s string) uint64 {
	addr, err := rt.mem.writeCString(s)
	if err != nil {
		log.WithError(err).Fatal("write C string failed")
	}
	return addr
}

// Dlopen loads a dynamic library in the remote process.
func (rt *objCRuntime) Dlopen(path string) uint64 {
	pathAddr, _ := rt.mem.writeCString(path)
	handle, _ := rt.mem.call(rt.dlopen, pathAddr, rtldNow)
	short := path[strings.LastIndex(path, "/")+1:]
	if handle != 0 {
		log.WithField("lib", short).Debug("dlopen OK")
	}
	return handle
}

// Dlsym resolves a symbol by name via dlsym(RTLD_DEFAULT, name).
func (rt *objCRuntime) Dlsym(name string) (uint64, error) {
	nameAddr, _ := rt.mem.writeCString(name)
	addr, err := rt.mem.call(rt.dlsym, rtldDefault, nameAddr)
	if err != nil || addr == 0 {
		return 0, fmt.Errorf("dlsym(%s) failed", name)
	}
	return addr, nil
}

// CallFunc calls a C function with integer args (x0-x7) and double-precision float args (v0-v7).
func (rt *objCRuntime) CallFunc(funcAddr uint64, intArgs []uint64, floatArgs []float64) (uint64, error) {
	return rt.mem.callWithFloats(funcAddr, intArgs, floatArgs)
}

// ClassCall calls a class method: [ClassName selector:args...]
func (rt *objCRuntime) ClassCall(className string, selector string, args ...uint64) (uint64, error) {
	cls, err := rt.class(className)
	if err != nil {
		return 0, err
	}
	return rt.Call(cls, selector, args...)
}

// Call calls an instance method: [receiver selector:args...]
func (rt *objCRuntime) Call(receiver uint64, selector string, args ...uint64) (uint64, error) {
	if receiver == 0 {
		return 0, fmt.Errorf("nil receiver for [? %s]", selector)
	}
	sel, err := rt.sel(selector)
	if err != nil {
		return 0, err
	}
	callArgs := make([]uint64, 0, 2+len(args))
	callArgs = append(callArgs, receiver, sel)
	callArgs = append(callArgs, args...)
	result, err := rt.mem.call(rt.msgSend, callArgs...)
	if err != nil {
		return 0, fmt.Errorf("objc_msgSend(%s): %w", selector, err)
	}
	if result == 0 {
		return 0, fmt.Errorf("[0x%x %s] returned nil", receiver, selector)
	}
	return result, nil
}

func (rt *objCRuntime) class(name string) (uint64, error) {
	if cached, ok := rt.clsCache[name]; ok {
		return cached, nil
	}
	nameAddr, _ := rt.mem.writeCString(name)
	cls, err := rt.mem.call(rt.getClass, nameAddr)
	if err != nil || cls == 0 {
		return 0, fmt.Errorf("objc_getClass(%q) failed", name)
	}
	rt.clsCache[name] = cls
	return cls, nil
}

func (rt *objCRuntime) sel(name string) (uint64, error) {
	if cached, ok := rt.selCache[name]; ok {
		return cached, nil
	}
	nameAddr, _ := rt.mem.writeCString(name)
	sel, err := rt.mem.call(rt.selReg, nameAddr)
	if err != nil || sel == 0 {
		return 0, fmt.Errorf("sel_registerName(%q) failed", name)
	}
	rt.selCache[name] = sel
	return sel, nil
}

// gdbMem manages memory allocation and function calls in a remote process via GDB RSP.
type gdbMem struct {
	gdb      *debugserver.GDBServer
	threadID string
	codeAddr uint64 // rwx page with brk #0 trap
	dataAddr uint64 // rw page for strings/data
	dataOff  uint64 // write cursor in data page
	saveID   string // saved register state ID
}

func newGDBMem(gdb *debugserver.GDBServer) (*gdbMem, error) {
	resp, _ := gdb.Request("qC")
	threadID := strings.TrimPrefix(resp, "QC")
	if threadID == "" {
		return nil, fmt.Errorf("no current thread")
	}

	m := &gdbMem{gdb: gdb, threadID: threadID}

	m.saveID, _ = gdb.Request(fmt.Sprintf("QSaveRegisterState;thread:%s;", threadID))

	resp, _ = gdb.Request(fmt.Sprintf("_M%x,rw", dataPageSize))
	fmt.Sscanf(resp, "%x", &m.dataAddr)

	resp, _ = gdb.Request(fmt.Sprintf("_M%x,rwx", codePageSize))
	fmt.Sscanf(resp, "%x", &m.codeAddr)

	if m.dataAddr == 0 || m.codeAddr == 0 {
		return nil, fmt.Errorf("memory allocation failed (data=0x%x code=0x%x)", m.dataAddr, m.codeAddr)
	}

	gdb.Request(fmt.Sprintf("M%x,%d:%s", m.codeAddr, arm64BrkKind, arm64BrkInstruction))
	gdb.Request(fmt.Sprintf("Z0,%x,%d", m.codeAddr, arm64BrkKind))

	return m, nil
}

// cleanup restores registers, removes breakpoint, and frees allocated memory.
// Must be called BEFORE detaching from the process.
func (m *gdbMem) cleanup() {
	m.gdb.Request(fmt.Sprintf("QRestoreRegisterState:%s;thread:%s;", m.saveID, m.threadID))
	m.gdb.Request(fmt.Sprintf("z0,%x,%d", m.codeAddr, arm64BrkKind))
	m.gdb.Request(fmt.Sprintf("_m%x", m.codeAddr))
	m.gdb.Request(fmt.Sprintf("_m%x", m.dataAddr))
}

func (m *gdbMem) writeCString(s string) (uint64, error) {
	return m.writeData(append([]byte(s), 0))
}

func (m *gdbMem) writeData(data []byte) (uint64, error) {
	addr := m.dataAddr + m.dataOff
	r, err := m.gdb.Request(fmt.Sprintf("M%x,%x:%s", addr, len(data), hexEncode(data)))
	if err != nil {
		return 0, err
	}
	if r != "OK" {
		return 0, fmt.Errorf("write: %s", r)
	}
	m.dataOff = (m.dataOff + uint64(len(data)) + 7) &^ 7
	return addr, nil
}

// call invokes a function with integer arguments only.
func (m *gdbMem) call(funcAddr uint64, args ...uint64) (uint64, error) {
	return m.callWithFloats(funcAddr, args, nil)
}

// callWithFloats invokes a function with integer (x0-x7) and double (v0-v7) arguments.
func (m *gdbMem) callWithFloats(funcAddr uint64, intArgs []uint64, floatArgs []float64) (uint64, error) {
	for i, arg := range intArgs {
		if err := m.writeReg(regX0+i, arg); err != nil {
			return 0, err
		}
	}
	for i, farg := range floatArgs {
		if err := m.writeFloatReg(regV0+i, farg); err != nil {
			return 0, err
		}
	}
	m.writeReg(regLR, m.codeAddr)
	m.writeReg(regPC, funcAddr)

	resp, err := m.gdb.Request(fmt.Sprintf("vCont;c:%s", m.threadID))
	if err != nil {
		return 0, fmt.Errorf("vCont: %w", err)
	}
	if !strings.HasPrefix(resp, "T") {
		return 0, fmt.Errorf("unexpected stop reply: %s", truncate(resp, 60))
	}
	return m.readReg(regX0)
}

// writeFloatReg writes a float64 (double) to a 128-bit NEON register.
func (m *gdbMem) writeFloatReg(reg int, val float64) error {
	bits := math.Float64bits(val)
	b := make([]byte, 16)
	b[0] = byte(bits)
	b[1] = byte(bits >> 8)
	b[2] = byte(bits >> 16)
	b[3] = byte(bits >> 24)
	b[4] = byte(bits >> 32)
	b[5] = byte(bits >> 40)
	b[6] = byte(bits >> 48)
	b[7] = byte(bits >> 56)
	r, _ := m.gdb.Request(fmt.Sprintf("P%x=%s;thread:%s;", reg, hexEncode(b), m.threadID))
	if r != "OK" {
		return fmt.Errorf("write float reg v%d: %s", reg-regV0, r)
	}
	return nil
}

func (m *gdbMem) readMemory(addr, size uint64) ([]byte, error) {
	var result []byte
	for off := uint64(0); off < size; off += memReadChunkSize {
		n := size - off
		if n > memReadChunkSize {
			n = memReadChunkSize
		}
		resp, err := m.gdb.Request(fmt.Sprintf("m%x,%x", addr+off, n))
		if err != nil {
			return nil, err
		}
		if strings.HasPrefix(resp, "E") {
			return nil, fmt.Errorf("read 0x%x: %s", addr+off, resp)
		}
		b, _ := hexDecode(resp)
		result = append(result, b...)
	}
	return result, nil
}

func (m *gdbMem) readReg(reg int) (uint64, error) {
	r, err := m.gdb.Request(fmt.Sprintf("p%x;thread:%s;", reg, m.threadID))
	if err != nil {
		return 0, err
	}
	b, _ := hexDecode(r)
	if len(b) < 8 {
		return 0, fmt.Errorf("short register: %q", r)
	}
	return leUint64(b), nil
}

func (m *gdbMem) writeReg(reg int, val uint64) error {
	b := leBytes(val)
	r, _ := m.gdb.Request(fmt.Sprintf("P%x=%s;thread:%s;", reg, hexEncode(b), m.threadID))
	if r != "OK" {
		return fmt.Errorf("write reg %d: %s", reg, r)
	}
	return nil
}

// --- hex/endian helpers ---

func truncate(s string, n int) string {
	if len(s) > n {
		return s[:n] + "..."
	}
	return s
}

func hexEncode(b []byte) string {
	const hextable = "0123456789abcdef"
	dst := make([]byte, len(b)*2)
	for i, v := range b {
		dst[i*2] = hextable[v>>4]
		dst[i*2+1] = hextable[v&0x0f]
	}
	return string(dst)
}

func hexDecode(s string) ([]byte, error) {
	n := len(s) / 2
	b := make([]byte, n)
	for i := 0; i < n; i++ {
		b[i] = unhex(s[i*2])<<4 | unhex(s[i*2+1])
	}
	return b, nil
}

func unhex(c byte) byte {
	switch {
	case '0' <= c && c <= '9':
		return c - '0'
	case 'a' <= c && c <= 'f':
		return c - 'a' + 10
	case 'A' <= c && c <= 'F':
		return c - 'A' + 10
	}
	return 0
}

func leUint64(b []byte) uint64 {
	return uint64(b[0]) | uint64(b[1])<<8 | uint64(b[2])<<16 | uint64(b[3])<<24 |
		uint64(b[4])<<32 | uint64(b[5])<<40 | uint64(b[6])<<48 | uint64(b[7])<<56
}

func leBytes(v uint64) []byte {
	return []byte{byte(v), byte(v >> 8), byte(v >> 16), byte(v >> 24),
		byte(v >> 32), byte(v >> 40), byte(v >> 48), byte(v >> 56)}
}
