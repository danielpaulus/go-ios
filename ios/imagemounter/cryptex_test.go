package imagemounter

import (
	"bytes"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"os"
	"path"
	"sync"
	"syscall"
	"testing"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/http"
	"github.com/danielpaulus/go-ios/ios/xpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/http2"
	"howett.net/plist"
)

func mustHex(s string) []byte {
	b, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return b
}

// values captured from an iPhone while Xcode installed the developer disk image
var (
	capturedNonce = mustHex("0000e565e4633875d66ef7080de57f575dee872fae090e0ba65b602fed0d0155be9fc4c619c2c4f7de129495114bb312ee7f000030000000")
	capturedChip  = map[string]interface{}{
		"img4_chip_bord": uint64(4), "img4_chip_cepo": uint64(1), "img4_chip_chip": uint64(33088),
		"img4_chip_clas": uint64(242), "img4_chip_cpro": uint64(1), "img4_chip_csec": uint64(1),
		"img4_chip_ecid": uint64(7388792203870236), "img4_chip_epro": uint64(1), "img4_chip_esdm": uint64(0),
		"img4_chip_esec": uint64(1), "img4_chip_euou": uint64(0), "img4_chip_fchp": uint64(65296),
		"img4_chip_fpgt": uint64(0), "img4_chip_iuou": uint64(0), "img4_chip_omit": uint64(0),
		"img4_chip_rsch": uint64(0), "img4_chip_sdom": uint64(1), "img4_chip_styp": uint64(255),
		"img4_chip_type": uint64(3),
	}
)

func TestParseImg4Nonce(t *testing.T) {
	nonce, err := parseImg4Nonce(capturedNonce)
	require.NoError(t, err)
	// the Cryptex1,Nonce ('cnch') of the ticket Apple issued for it
	assert.Equal(t, "e565e4633875d66ef7080de57f575dee872fae090e0ba65b602fed0d0155be9fc4c619c2c4f7de129495114bb312ee7f", hex.EncodeToString(nonce))

	_, err = parseImg4Nonce(capturedNonce[:40])
	assert.Error(t, err)
	invalidLength := bytes.Clone(capturedNonce)
	invalidLength[52] = 49
	_, err = parseImg4Nonce(invalidLength)
	assert.Error(t, err)
}

func TestParseCryptexChip(t *testing.T) {
	chip, err := parseCryptexChip(capturedChip)
	require.NoError(t, err)
	assert.Equal(t, cryptexChip{chip: 0x8140, ecid: 7388792203870236, fchp: 0xff10, typ: 3, clas: 0xf2, production: true, secure: true}, chip)

	_, err = parseCryptexChip(map[string]interface{}{"img4_chip_chip": uint64(1)})
	assert.Error(t, err)
}

func TestCryptexUDID(t *testing.T) {
	// the 'UDID' property of the ticket Apple issued
	assert.Equal(t, "0000000000008140001a40113ea1801c", hex.EncodeToString(cryptexUDID(0x8140, 7388792203870236)))
}

func TestParseCryptexReply(t *testing.T) {
	argv, err := parseCryptexReply("install", map[string]interface{}{"argv": map[string]interface{}{"a": "b"}})
	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"a": "b"}, argv)

	_, err = parseCryptexReply("install", map[string]interface{}{"error": uint64(22)})
	assert.ErrorIs(t, err, syscall.EINVAL)

	_, err = parseCryptexReply("install", map[string]interface{}{"cferr": map[string]interface{}{"code": int64(11)}})
	assert.Error(t, err)
}

type testImage struct {
	dir   string
	files map[string][]byte
}

// writeTestImage creates a developer disk image 'Restore' directory with a
// personalized DMG identity and a Cryptex1 identity like Xcode's DDI.
func writeTestImage(t *testing.T) testImage {
	return writeTestImageWithVersion(t, "27.1.9269.0")
}

func writeTestImageWithVersion(t *testing.T, version string) testImage {
	dir := t.TempDir()
	require.NoError(t, os.Mkdir(path.Join(dir, "Firmware"), 0o755))
	info, err := plist.Marshal(map[string]interface{}{
		"CFBundleIdentifier": ddiCryptexIdentifier,
		"CFBundleVersion":    version,
	}, plist.XMLFormat)
	require.NoError(t, err)
	files := map[string][]byte{
		cryptexDmgKey:        bytes.Repeat([]byte("dmg-"), 20000),
		cryptexVolumeKey:     []byte("root hash"),
		cryptexInfoPlistKey:  info,
		cryptexTrustCacheKey: bytes.Repeat([]byte("tc"), 3000),
	}
	paths := map[string]string{
		cryptexDmgKey:        "c.dmg",
		cryptexVolumeKey:     "Firmware/c.dmg.root_hash",
		cryptexInfoPlistKey:  "Firmware/c.dmg.cryptex_info",
		cryptexTrustCacheKey: "Firmware/c.dmg.trustcache",
	}
	entries := map[string]interface{}{}
	for key, content := range files {
		require.NoError(t, os.WriteFile(path.Join(dir, paths[key]), content, 0o644))
		digest := sha512.Sum384(content)
		entries[key] = map[string]interface{}{
			"Digest": digest[:],
			"Info": map[string]interface{}{
				"Path":        paths[key],
				"Personalize": key != cryptexDmgKey,
			},
		}
	}
	manifest := map[string]interface{}{
		"ProductBuildVersion": "27A9269",
		"BuildIdentities": []interface{}{
			map[string]interface{}{
				"ApBoardID": "0x08", "ApChipID": "0x8140",
				"Manifest": map[string]interface{}{"PersonalizedDMG": map[string]interface{}{"Digest": []byte{1}, "Info": map[string]interface{}{"Path": "p.dmg"}}},
			},
			map[string]interface{}{
				"Cryptex1,ChipID":                  "0xFF10",
				"Cryptex1,NonceDomain":             uint64(4),
				"Cryptex1,PreauthorizationVersion": "39.999.999.0.0,0",
				"Cryptex1,ProductClass":            "0xF2",
				"Cryptex1,SubType":                 uint64(2),
				"Cryptex1,Type":                    uint64(3),
				"Cryptex1,UseProductClass":         true,
				"Cryptex1,Version":                 "39.999.999.0.0,0",
				"Manifest":                         entries,
			},
		},
	}
	b, err := plist.Marshal(manifest, plist.XMLFormat)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(path.Join(dir, "BuildManifest.plist"), b, 0o644))
	return testImage{dir: dir, files: files}
}

func TestCryptexImageVersion(t *testing.T) {
	img := writeTestImageWithVersion(t, "27.1.9266.5")
	image, err := CryptexImageVersion(img.dir)
	require.NoError(t, err)
	assert.Equal(t, RemoteCryptex{Identifier: ddiCryptexIdentifier, Version: "27.1.9266.5"}, image)

	_, err = CryptexImageVersion(t.TempDir())
	assert.Error(t, err)
}

func TestCryptexTSSRequest(t *testing.T) {
	img := writeTestImage(t)
	assert.True(t, hasCryptexIdentity(img.dir))

	manifest, err := loadBuildManifest(path.Join(img.dir, "BuildManifest.plist"))
	require.NoError(t, err)
	chip, err := parseCryptexChip(capturedChip)
	require.NoError(t, err)
	identity, err := manifest.findCryptexIdentity(chip)
	require.NoError(t, err)

	_, err = manifest.findCryptexIdentity(cryptexChip{fchp: 0xff10, typ: 3, clas: 0xf0})
	assert.Error(t, err, "product class must match")

	digests, err := identity.cryptexDigests(img.dir)
	require.NoError(t, err)
	assert.Len(t, digests, 3, "the dmg itself isn't personalized")

	nonce, err := parseImg4Nonce(capturedNonce)
	require.NoError(t, err)
	req := cryptexTSSRequest(identity, chip, nonce, digests)

	// This combination was verified against gs.apple.com: it yields a ticket
	// with the same properties as the one Xcode got.
	for key, want := range map[string]interface{}{
		"@Cryptex1,Ticket":                 true,
		"ApSecurityMode":                   true,
		"Cryptex1,UDID":                    mustHex("0000000000008140001a40113ea1801c"),
		"Cryptex1,ChipID":                  uint64(0xff10),
		"Cryptex1,Type":                    uint64(3),
		"Cryptex1,SubType":                 uint64(2),
		"Cryptex1,ProductClass":            uint64(0xf2),
		"Cryptex1,ProductionMode":          true,
		"Cryptex1,UseProductClass":         true,
		"Cryptex1,NonceDomain":             uint64(4),
		"Cryptex1,Nonce":                   nonce,
		"Cryptex1,Version":                 "39.999.999.0.0,0",
		"Cryptex1,PreauthorizationVersion": "39.999.999.0.0,0",
		"Cryptex1,UniqueTagList":           []byte{},
	} {
		assert.Equal(t, want, req[key], key)
	}
	// Apple TSS refuses Cryptex1 requests with these
	for _, key := range []string{"@ApImg4Ticket", "ApECID", "ApChipID", "ApBoardID", "ApSecurityDomain", "ApNonce", "SepNonce", "UID_MODE", cryptexDmgKey} {
		assert.NotContains(t, req, key)
	}
	for _, key := range []string{cryptexTrustCacheKey, cryptexVolumeKey, cryptexInfoPlistKey} {
		digest := sha512.Sum384(img.files[key])
		assert.Equal(t, map[string]interface{}{"Digest": digest[:]}, req[key], key)
	}
}

func TestCryptexDigestsDetectModifiedFiles(t *testing.T) {
	img := writeTestImage(t)
	require.NoError(t, os.WriteFile(path.Join(img.dir, "Firmware/c.dmg.trustcache"), []byte("modified"), 0o644))
	manifest, err := loadBuildManifest(path.Join(img.dir, "BuildManifest.plist"))
	require.NoError(t, err)
	identity, err := manifest.findCryptexIdentity(cryptexChip{fchp: 0xff10, typ: 3, clas: 0xf2})
	require.NoError(t, err)
	_, err = identity.cryptexDigests(img.dir)
	assert.Error(t, err)
}

// fakeCryptexd emulates cryptexd's remote service on the HTTP/2 + XPC level
type fakeCryptexd struct {
	t *testing.T
	// window is the flow-control window the fake announces and keeps topping
	// up, small enough that uploads have to wait for WINDOW_UPDATE frames.
	window uint32

	mu        sync.Mutex
	requests  []map[string]interface{}
	uploads   map[string][]byte
	installed []RemoteCryptex
}

func (f *fakeCryptexd) connect() (*xpc.Connection, error) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}
	go func() {
		defer l.Close()
		c, err := l.Accept()
		if err != nil {
			return
		}
		defer c.Close()
		if err := f.serve(c); err != nil && err != io.EOF {
			f.t.Errorf("fake cryptexd: %v", err)
		}
	}()
	c, err := net.Dial("tcp", l.Addr().String())
	if err != nil {
		return nil, err
	}
	h, err := http.NewHttpConnection(c)
	if err != nil {
		return nil, err
	}
	return ios.CreateXpcConnection(h)
}

type fakeStream struct {
	buf        []byte
	transferId uint64
	opened     bool
	ended      bool
}

func (f *fakeCryptexd) serve(c net.Conn) error {
	preface := make([]byte, len(http2.ClientPreface))
	if _, err := io.ReadFull(c, preface); err != nil {
		return err
	}
	fr := http2.NewFramer(c, c)
	if err := fr.WriteSettings(http2.Setting{ID: http2.SettingInitialWindowSize, Val: f.window}); err != nil {
		return err
	}

	streams := map[uint32]*fakeStream{}
	transfers := map[uint64]uint32{} // file transfer id -> stream
	var install map[string]interface{}
	var installId uint64

	reply := func(stream uint32, msg xpc.Message) error {
		buf := bytes.NewBuffer(nil)
		if err := xpc.EncodeMessage(buf, msg); err != nil {
			return err
		}
		return fr.WriteData(stream, false, buf.Bytes())
	}

	for {
		frame, err := fr.ReadFrame()
		if err != nil {
			return err
		}
		switch frame := frame.(type) {
		case *http2.SettingsFrame:
			if !frame.IsAck() {
				if err := fr.WriteSettingsAck(); err != nil {
					return err
				}
			}
		case *http2.DataFrame:
			s, ok := streams[frame.StreamID]
			if !ok {
				s = &fakeStream{}
				streams[frame.StreamID] = s
			}
			s.buf = append(s.buf, frame.Data()...)
			s.ended = frame.StreamEnded()
			if n := uint32(len(frame.Data())); n > 0 {
				if err := fr.WriteWindowUpdate(0, n); err != nil {
					return err
				}
				if err := fr.WriteWindowUpdate(frame.StreamID, n); err != nil {
					return err
				}
			}
			if frame.StreamID > 3 {
				if !s.opened && len(s.buf) >= 24 {
					msg, err := xpc.DecodeMessage(bytes.NewReader(s.buf[:24]))
					if err != nil {
						return err
					}
					assert.Equal(f.t, xpc.AlwaysSetFlag|xpc.FileOpenFlag, msg.Flags)
					s.opened, s.transferId, s.buf = true, msg.Id, s.buf[24:]
					transfers[msg.Id] = frame.StreamID
				}
				break
			}
			// XPC messages on the request (1) and reply (3) streams
			for len(s.buf) > 0 {
				r := bytes.NewReader(s.buf)
				msg, err := xpc.DecodeMessage(r)
				if err != nil {
					break // wait for the rest of the message
				}
				s.buf = s.buf[len(s.buf)-r.Len():]
				if msg.Flags&xpc.HeartbeatRequestFlag == 0 {
					// connection handshake, answered in kind
					if err := reply(frame.StreamID, xpc.Message{Flags: msg.Flags, Body: msg.Body}); err != nil {
						return err
					}
					continue
				}
				f.mu.Lock()
				f.requests = append(f.requests, msg.Body)
				f.mu.Unlock()
				argv, _ := msg.Body["argv"].(map[string]interface{})
				var resp map[string]interface{}
				switch msg.Body["routine"] {
				case "read-personalization-id":
					resp = map[string]interface{}{"argv": capturedChip}
				case "get-nonce":
					resp = map[string]interface{}{"argv": map[string]interface{}{"nonce": capturedNonce}}
				case "copy-installed":
					var list []interface{}
					f.mu.Lock()
					for _, c := range f.installed {
						list = append(list, map[string]interface{}{"remote-cryptex-identifier": c.Identifier, "remote-cryptex-version": c.Version})
					}
					f.mu.Unlock()
					resp = map[string]interface{}{"argv": map[string]interface{}{"remote-cryptex-array": list}}
				case "uninstall":
					f.mu.Lock()
					f.installed = nil
					f.mu.Unlock()
					resp = map[string]interface{}{"argv": map[string]interface{}{}}
				case "install":
					// accept the file transfers, the reply follows once all of them arrived
					install, installId = argv, msg.Id
					for _, key := range []string{"image", "trustcache", "im4m", "info", "volumehash"} {
						ft, ok := argv[key].(xpc.FileTransfer)
						if !ok {
							return fmt.Errorf("install: %s is no file transfer", key)
						}
						stream, ok := transfers[ft.MsgId]
						if !ok {
							return fmt.Errorf("install: no stream was opened for %s", key)
						}
						if err := reply(stream, xpc.Message{Flags: xpc.AlwaysSetFlag | xpc.FileOpenReplyFlag, Id: ft.MsgId}); err != nil {
							return err
						}
					}
					continue
				default:
					resp = map[string]interface{}{"error": uint64(syscall.ENOTSUP)}
				}
				if err := reply(3, xpc.Message{Flags: xpc.AlwaysSetFlag | xpc.DataFlag | xpc.HeartbeatReplyFlag, Body: resp, Id: msg.Id}); err != nil {
					return err
				}
			}
		}

		if install != nil && f.uploadsComplete(install, streams, transfers) {
			f.mu.Lock()
			f.uploads = map[string][]byte{}
			for _, key := range []string{"image", "trustcache", "im4m", "info", "volumehash"} {
				ft := install[key].(xpc.FileTransfer)
				data := streams[transfers[ft.MsgId]].buf
				assert.Equal(f.t, ft.TransferSize, uint64(len(data)), key)
				f.uploads[key] = data
			}
			// like cryptexd, report the identifier and version of the cryptex_info
			var info cryptexInfo
			_, err := plist.Unmarshal(f.uploads["info"], &info)
			assert.NoError(f.t, err)
			f.installed = append(f.installed, RemoteCryptex{Identifier: info.CFBundleIdentifier, Version: info.CFBundleVersion})
			f.mu.Unlock()
			resp := map[string]interface{}{"argv": map[string]interface{}{"remote-cryptex": map[string]interface{}{
				"remote-cryptex-identifier": info.CFBundleIdentifier,
				"remote-cryptex-version":    info.CFBundleVersion,
			}}}
			if err := reply(3, xpc.Message{Flags: xpc.AlwaysSetFlag | xpc.DataFlag | xpc.HeartbeatReplyFlag, Body: resp, Id: installId}); err != nil {
				return err
			}
			install = nil
		}
	}
}

func (f *fakeCryptexd) uploadsComplete(install map[string]interface{}, streams map[uint32]*fakeStream, transfers map[uint64]uint32) bool {
	for _, key := range []string{"image", "trustcache", "im4m", "info", "volumehash"} {
		ft := install[key].(xpc.FileTransfer)
		if s := streams[transfers[ft.MsgId]]; s == nil || !s.ended {
			return false
		}
	}
	return true
}

func TestCryptexMountListUnmount(t *testing.T) {
	img := writeTestImage(t)
	fake := &fakeCryptexd{t: t, window: 1000}
	im4m := bytes.Repeat([]byte("ticket"), 500)
	var signRequest map[string]interface{}
	mounter := &CryptexDeveloperDiskImageMounter{
		connect: fake.connect,
		sign: func(request map[string]interface{}) ([]byte, error) {
			signRequest = request
			return im4m, nil
		},
	}

	images, err := mounter.ListImages()
	require.NoError(t, err)
	assert.Empty(t, images)
	mounted, err := mounter.IsImageMounted(img.dir)
	require.NoError(t, err)
	assert.False(t, mounted)

	require.NoError(t, mounter.MountImage(img.dir))

	nonce, _ := parseImg4Nonce(capturedNonce)
	assert.Equal(t, nonce, signRequest["Cryptex1,Nonce"])
	assert.Equal(t, img.files[cryptexDmgKey], fake.uploads["image"])
	assert.Equal(t, img.files[cryptexTrustCacheKey], fake.uploads["trustcache"])
	assert.Equal(t, img.files[cryptexVolumeKey], fake.uploads["volumehash"])
	assert.Equal(t, img.files[cryptexInfoPlistKey], fake.uploads["info"])
	assert.Equal(t, im4m, fake.uploads["im4m"])

	fake.mu.Lock()
	var routines []interface{}
	for _, r := range fake.requests {
		routines = append(routines, r["routine"])
	}
	getNonce := fake.requests[3]
	installReq := fake.requests[4]
	fake.mu.Unlock()
	assert.Equal(t, []interface{}{"copy-installed", "copy-installed", "read-personalization-id", "get-nonce", "install"}, routines)
	assert.Equal(t, map[string]interface{}{"nonce-domain-handle": uint64(4)}, getNonce["argv"])
	argv := installReq["argv"].(map[string]interface{})
	assert.Equal(t, int64(10), argv["image-type-index"])
	assert.Equal(t, uint64(2), argv["persistence"])
	assert.Equal(t, uint64(1), argv["nonce-persistence"])
	assert.Equal(t, map[string]interface{}{
		"Cryptex1,NonceDomain":     uint64(4),
		"Cryptex1,PreauthVersion":  "39.999.999.0.0,0",
		"Cryptex1,SubType":         uint64(2),
		"Cryptex1,UseProductClass": true,
		"Cryptex1,Version":         "39.999.999.0.0,0",
		"MountedCryptex":           false,
	}, argv["cryptex1-properties"])

	mounted, err = mounter.IsImageMounted(img.dir)
	require.NoError(t, err)
	assert.True(t, mounted)
	mounted, err = mounter.IsImageMounted(writeTestImageWithVersion(t, "27.1.9270.0").dir)
	require.NoError(t, err)
	assert.False(t, mounted, "a different version of the image is not mounted")

	images, err = mounter.ListImages()
	require.NoError(t, err)
	assert.Equal(t, [][]byte{[]byte(ddiCryptexIdentifier)}, images)

	require.NoError(t, mounter.UnmountImage())
	images, err = mounter.ListImages()
	require.NoError(t, err)
	assert.Empty(t, images)
	assert.Error(t, mounter.UnmountImage(), "nothing left to uninstall")
}
