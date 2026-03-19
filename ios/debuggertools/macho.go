package debuggertools

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/danielpaulus/go-ios/ios/debugserver"
)

// Mach-O constants (from <mach-o/loader.h>)
const (
	machOMagic64       = 0xFEEDFACF // MH_MAGIC_64 — 64-bit Mach-O
	machOHeaderSize64  = 32         // sizeof(mach_header_64)
	lcSegment64        = 0x19       // LC_SEGMENT_64
	lcSegment64MinSize = 72         // minimum size of segment_command_64
	lcDyldInfoOnly     = 0x80000022 // LC_DYLD_INFO_ONLY
	lcDyldExportsTrie  = 0x80000033 // LC_DYLD_EXPORTS_TRIE
	maxLoadCmdsSize    = 0x10000    // safety limit for reading load commands
)

// segment_command_64 field offsets (relative to command start)
const (
	segNameOff    = 8  // char segname[16]
	segVMAddrOff  = 24 // uint64 vmaddr
	segFileOffOff = 40 // uint64 fileoff
)

// dyld_all_image_infos field offsets
const (
	imageInfoCountOff = 4 // uint32 infoArrayCount
	imageInfoArrayOff = 8 // uint64 infoArray pointer
)

// dyld_image_info struct size and field offsets
const (
	imageInfoSize    = 24 // sizeof(dyld_image_info)
	imageLoadAddrOff = 0  // uint64 imageLoadAddress
	imagePathPtrOff  = 8  // uint64 imageFilePath pointer
)

// findExportInMachO reads a Mach-O header from process memory at baseAddr
// and resolves an exported symbol by walking the dyld export trie.
//
// It parses the load commands to find:
//   - __TEXT segment vmaddr (for computing ASLR slide)
//   - __LINKEDIT segment vmaddr + fileoff (for locating the export trie in memory)
//   - LC_DYLD_EXPORTS_TRIE or LC_DYLD_INFO_ONLY (for export trie file offset + size)
//
// Returns the symbol's absolute address, or 0 if not found.
func findExportInMachO(gdb *debugserver.GDBServer, baseAddr uint64, symbolName string) uint64 {
	// Read mach_header_64
	resp, err := gdb.Request(fmt.Sprintf("m%x,%x", baseAddr, machOHeaderSize64))
	if err != nil || resp == "" || strings.HasPrefix(resp, "E") {
		return 0
	}
	header, _ := hex.DecodeString(resp)
	if len(header) < machOHeaderSize64 || binary.LittleEndian.Uint32(header[0:4]) != machOMagic64 {
		return 0
	}

	ncmds := binary.LittleEndian.Uint32(header[16:20])
	sizeOfCmds := binary.LittleEndian.Uint32(header[20:24])
	if sizeOfCmds > maxLoadCmdsSize {
		sizeOfCmds = maxLoadCmdsSize
	}

	// Read all load commands
	resp, _ = gdb.Request(fmt.Sprintf("m%x,%x", baseAddr+machOHeaderSize64, sizeOfCmds))
	cmds, _ := hex.DecodeString(resp)

	var textVMAddr, linkeditVMAddr, linkeditFileOff uint64
	var exportFileOff, exportSize uint32

	// Walk load commands looking for segments and export trie info
	off := uint32(0)
	for i := uint32(0); i < ncmds && off+8 <= uint32(len(cmds)); i++ {
		cmd := binary.LittleEndian.Uint32(cmds[off : off+4])
		cmdSize := binary.LittleEndian.Uint32(cmds[off+4 : off+8])
		if cmdSize < 8 || off+cmdSize > uint32(len(cmds)) {
			break
		}

		switch cmd {
		case lcSegment64:
			if off+lcSegment64MinSize <= uint32(len(cmds)) {
				segName := cString(cmds[off+segNameOff : off+segNameOff+16])
				vmaddr := binary.LittleEndian.Uint64(cmds[off+segVMAddrOff : off+segVMAddrOff+8])
				fileoff := binary.LittleEndian.Uint64(cmds[off+segFileOffOff : off+segFileOffOff+8])
				switch segName {
				case "__TEXT":
					textVMAddr = vmaddr
				case "__LINKEDIT":
					linkeditVMAddr = vmaddr
					linkeditFileOff = fileoff
				}
			}

		case lcDyldInfoOnly:
			// export_off at offset 40, export_size at offset 44
			if off+48 <= uint32(len(cmds)) {
				exportFileOff = binary.LittleEndian.Uint32(cmds[off+40 : off+44])
				exportSize = binary.LittleEndian.Uint32(cmds[off+44 : off+48])
			}

		case lcDyldExportsTrie:
			// dataoff at offset 8, datasize at offset 12
			if off+16 <= uint32(len(cmds)) {
				exportFileOff = binary.LittleEndian.Uint32(cmds[off+8 : off+12])
				exportSize = binary.LittleEndian.Uint32(cmds[off+12 : off+16])
			}
		}

		off += cmdSize
	}

	if exportSize == 0 || linkeditVMAddr == 0 || textVMAddr == 0 {
		return 0
	}

	// Convert file offset → memory address using ASLR slide.
	// slide = actual load address - compiled __TEXT vmaddr.
	// Export trie lives within __LINKEDIT: addr = slide + linkedit_vmaddr + (export_fileoff - linkedit_fileoff)
	slide := baseAddr - textVMAddr
	trieAddr := slide + linkeditVMAddr + (uint64(exportFileOff) - linkeditFileOff)

	// Read and walk the export trie
	resp, _ = gdb.Request(fmt.Sprintf("m%x,%x", trieAddr, exportSize))
	trie, _ := hex.DecodeString(resp)
	if len(trie) == 0 {
		return 0
	}

	// Exported C symbols have a leading underscore
	return walkExportTrie(trie, "_"+symbolName, baseAddr)
}

// walkExportTrie walks a dyld export trie to find a symbol.
//
// The trie is a compact prefix tree where each node has:
//   - Terminal info size (ULEB128) — if >0, this node is an exported symbol
//   - Terminal info (flags + address offset, both ULEB128)
//   - Child count (1 byte)
//   - For each child: edge label (null-terminated string) + child node offset (ULEB128)
//
// Returns baseAddr + symbol offset, or 0 if not found.
func walkExportTrie(trie []byte, symbol string, baseAddr uint64) uint64 {
	target := []byte(symbol)
	nodeOff := 0

	for depth := 0; depth < len(target); {
		if nodeOff >= len(trie) {
			return 0
		}

		// Skip terminal info at current node
		termSize, n := readULEB128(trie[nodeOff:])
		nodeOff += n + int(termSize)
		if nodeOff >= len(trie) {
			return 0
		}

		// Read child count and search for matching edge
		childCount := int(trie[nodeOff])
		nodeOff++

		matched := false
		for i := 0; i < childCount; i++ {
			if nodeOff >= len(trie) {
				return 0
			}
			// Read edge label (null-terminated)
			labelStart := nodeOff
			for nodeOff < len(trie) && trie[nodeOff] != 0 {
				nodeOff++
			}
			if nodeOff >= len(trie) {
				return 0
			}
			label := trie[labelStart:nodeOff]
			nodeOff++ // skip null terminator

			// Read child node offset
			childOff, n := readULEB128(trie[nodeOff:])
			nodeOff += n

			// Check if this edge matches the remaining symbol prefix
			remaining := target[depth:]
			if len(label) <= len(remaining) && string(label) == string(remaining[:len(label)]) {
				depth += len(label)
				nodeOff = int(childOff)
				matched = true
				break
			}
		}
		if !matched {
			return 0
		}
	}

	// At the matching node — read terminal info
	if nodeOff >= len(trie) {
		return 0
	}
	termSize, n := readULEB128(trie[nodeOff:])
	nodeOff += n
	if termSize == 0 {
		return 0 // node exists but is not a terminal (no export here)
	}

	// Terminal info: flags (ULEB128) + address offset (ULEB128)
	_, n = readULEB128(trie[nodeOff:]) // flags (unused)
	nodeOff += n
	addrOff, _ := readULEB128(trie[nodeOff:])

	return baseAddr + addrOff
}

// readULEB128 decodes an unsigned LEB128 value.
// Returns the value and the number of bytes consumed.
func readULEB128(data []byte) (uint64, int) {
	var result uint64
	var shift uint
	for i, b := range data {
		result |= uint64(b&0x7f) << shift
		if b&0x80 == 0 {
			return result, i + 1
		}
		shift += 7
	}
	return result, len(data)
}

func cString(b []byte) string {
	if i := bytes.IndexByte(b, 0); i >= 0 {
		return string(b[:i])
	}
	return string(b)
}

// symbolQuery describes a symbol to look up in a specific library.
type symbolQuery struct {
	lib    string // substring to match in the library path
	symbol string // exported symbol name (without leading underscore)
}

// resolveSymbols finds function addresses by scanning dyld_all_image_infos
// and parsing Mach-O export tries from process memory.
//
// dyld_all_image_infos contains an array of dyld_image_info structs,
// each with (imageLoadAddress, imageFilePath, imageFileModDate).
// We scan in batches, reading library paths to find matching libraries,
// then parse their Mach-O headers to resolve the requested symbols.
func resolveSymbols(gdb *debugserver.GDBServer, queries []symbolQuery) ([]uint64, error) {
	// Get dyld_all_image_infos address from debugserver
	resp, _ := gdb.Request("qShlibInfoAddr")
	if resp == "" || strings.HasPrefix(resp, "E") {
		return nil, fmt.Errorf("qShlibInfoAddr: %s", resp)
	}
	var shlibAddr uint64
	fmt.Sscanf(resp, "%x", &shlibAddr)

	// Read version + count + infoArray pointer
	resp, _ = gdb.Request(fmt.Sprintf("m%x,%x", shlibAddr, imageInfoArrayOff+8))
	shlibData, _ := hex.DecodeString(resp)
	if len(shlibData) < imageInfoArrayOff+8 {
		return nil, fmt.Errorf("cannot read dyld_all_image_infos")
	}
	count := binary.LittleEndian.Uint32(shlibData[imageInfoCountOff : imageInfoCountOff+4])
	infoArray := binary.LittleEndian.Uint64(shlibData[imageInfoArrayOff : imageInfoArrayOff+8])

	results := make([]uint64, len(queries))
	found := 0

	const batchCount = 50 // number of image infos to read per GDB request

	for batch := uint32(0); batch < count && found < len(queries); batch += batchCount {
		batchSize := count - batch
		if batchSize > batchCount {
			batchSize = batchCount
		}

		// Read a batch of dyld_image_info structs
		resp, _ = gdb.Request(fmt.Sprintf("m%x,%x", infoArray+uint64(batch)*imageInfoSize, batchSize*imageInfoSize))
		imgData, _ := hex.DecodeString(resp)

		for i := uint32(0); i < batchSize && (i+1)*imageInfoSize <= uint32(len(imgData)); i++ {
			off := i * imageInfoSize
			loadAddr := binary.LittleEndian.Uint64(imgData[off+imageLoadAddrOff : off+imageLoadAddrOff+8])
			pathPtr := binary.LittleEndian.Uint64(imgData[off+imagePathPtrOff : off+imagePathPtrOff+8])

			// Read library path string (first 64 bytes is enough to identify the library)
			pathResp, _ := gdb.Request(fmt.Sprintf("m%x,40", pathPtr))
			if pathResp == "" || strings.HasPrefix(pathResp, "E") {
				continue
			}
			pathBytes, _ := hex.DecodeString(pathResp)
			pathStr := cString(pathBytes)

			// Check if this library matches any pending queries
			for qi, q := range queries {
				if results[qi] != 0 || !strings.Contains(pathStr, q.lib) {
					continue
				}
				if addr := findExportInMachO(gdb, loadAddr, q.symbol); addr != 0 {
					results[qi] = addr
					found++
				}
			}
			if found >= len(queries) {
				break
			}
		}
	}

	for i, q := range queries {
		if results[i] == 0 {
			return nil, fmt.Errorf("symbol %s not found in %s", q.symbol, q.lib)
		}
	}
	return results, nil
}
