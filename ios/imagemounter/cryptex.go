package imagemounter

import (
	"bytes"
	"crypto/sha512"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path"
	"strconv"
	"strings"
	"syscall"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/golog"
	"github.com/danielpaulus/go-ios/ios/xpc"
	"github.com/google/uuid"
)

// cryptexServiceName is the RemoteXPC service of cryptexd. Recent devices install
// the developer disk image through it as a cryptex, instead of mounting it with
// the mobile_image_mounter.
const cryptexServiceName = "com.apple.security.cryptexd.remote"

// ddiCryptexIdentifier is the CFBundleIdentifier from the cryptex_info of the
// developer disk image, which cryptexd reports as remote-cryptex-identifier.
const ddiCryptexIdentifier = "com.apple.MobileAsset.DDI"

// The values below are what CoreDevice sends when it installs the developer
// disk image.
const (
	cryptexClientVersion    = uint64(3)
	cryptexImageTypeIndex   = int64(10)
	cryptexPersistence      = uint64(2)
	cryptexNoncePersistence = uint64(1)
)

// Manifest entries of the Cryptex1 build identity
const (
	cryptexDmgKey        = "Cryptex1,GenericDmg"
	cryptexTrustCacheKey = "Cryptex1,GenericTrustCache"
	cryptexVolumeKey     = "Cryptex1,GenericVolume"
	cryptexInfoPlistKey  = "Cryptex1,CryptexInfoPlist"
)

// CryptexDeveloperDiskImageMounter installs the developer disk image as a cryptex
// through cryptexd. The image is personalized with a Cryptex1 ticket from Apple's
// TSS server, which is bound to a nonce the device generates for the cryptex
// nonce domain.
type CryptexDeveloperDiskImageMounter struct {
	entry   ios.DeviceEntry
	connect func() (*xpc.Connection, error)
	sign    func(request map[string]interface{}) ([]byte, error)
}

// RemoteCryptex is a cryptex that was installed through cryptexd's remote service
type RemoteCryptex struct {
	Identifier string
	Version    string
}

// SupportsCryptexDDI reports whether the developer disk image can be installed
// through cryptexd on this device. That needs a tunnel to the device, as the
// service is only reachable through RemoteXPC.
func SupportsCryptexDDI(entry ios.DeviceEntry) bool {
	if !entry.SupportsRsd() {
		return false
	}
	_, err := ios.RsdPortForService(entry.Rsd, cryptexServiceName)
	return err == nil
}

// NewCryptexDeveloperDiskImageMounter creates a CryptexDeveloperDiskImageMounter for the device entry
func NewCryptexDeveloperDiskImageMounter(entry ios.DeviceEntry) (*CryptexDeveloperDiskImageMounter, error) {
	if !SupportsCryptexDDI(entry) {
		return nil, fmt.Errorf("NewCryptexDeveloperDiskImageMounter: %s is not available, make sure a tunnel is running", cryptexServiceName)
	}
	tss := newTssClient()
	return &CryptexDeveloperDiskImageMounter{
		entry: entry,
		connect: func() (*xpc.Connection, error) {
			return ios.ConnectToXpcServiceTunnelIface(entry, cryptexServiceName)
		},
		sign: func(request map[string]interface{}) ([]byte, error) {
			return tss.requestTicket(request, "Cryptex1,Ticket")
		},
	}, nil
}

// Close does nothing, every operation uses connections of its own
func (c *CryptexDeveloperDiskImageMounter) Close() error {
	return nil
}

// ListImages returns the identifiers of the installed developer disk image cryptexes
func (c *CryptexDeveloperDiskImageMounter) ListImages() ([][]byte, error) {
	installed, err := c.ListCryptexes()
	if err != nil {
		return nil, err
	}
	var images [][]byte
	for _, cryptex := range installed {
		if cryptex.Identifier == ddiCryptexIdentifier {
			images = append(images, []byte(cryptex.Identifier))
		}
	}
	return images, nil
}

// ListCryptexes returns all cryptexes that were installed through the remote service
func (c *CryptexDeveloperDiskImageMounter) ListCryptexes() ([]RemoteCryptex, error) {
	conn, err := c.connect()
	if err != nil {
		return nil, fmt.Errorf("ListCryptexes: %w", err)
	}
	defer conn.Close()
	return listCryptexes(conn)
}

func listCryptexes(conn *xpc.Connection) ([]RemoteCryptex, error) {
	argv, err := cryptexCall(conn, 1, "copy-installed", map[string]interface{}{})
	if err != nil {
		return nil, fmt.Errorf("listCryptexes: %w", err)
	}
	entries, _ := argv["remote-cryptex-array"].([]interface{})
	cryptexes := make([]RemoteCryptex, 0, len(entries))
	for _, e := range entries {
		if d, ok := e.(map[string]interface{}); ok {
			cryptexes = append(cryptexes, parseRemoteCryptex(d))
		}
	}
	return cryptexes, nil
}

func parseRemoteCryptex(d map[string]interface{}) RemoteCryptex {
	identifier, _ := d["remote-cryptex-identifier"].(string)
	version, _ := d["remote-cryptex-version"].(string)
	return RemoteCryptex{Identifier: identifier, Version: version}
}

// UnmountImage uninstalls the developer disk image cryptex
func (c *CryptexDeveloperDiskImageMounter) UnmountImage() error {
	conn, err := c.connect()
	if err != nil {
		return fmt.Errorf("UnmountImage: %w", err)
	}
	defer conn.Close()

	installed, err := listCryptexes(conn)
	if err != nil {
		return fmt.Errorf("UnmountImage: %w", err)
	}
	id := uint64(1)
	found := false
	for _, cryptex := range installed {
		if cryptex.Identifier != ddiCryptexIdentifier {
			continue
		}
		found = true
		_, err := cryptexCall(conn, id, "uninstall", map[string]interface{}{
			"remote-cryptex-identifier": cryptex.Identifier,
			"remote-cryptex-version":    cryptex.Version,
		})
		if err != nil {
			return fmt.Errorf("UnmountImage: failed to uninstall %s %s: %w", cryptex.Identifier, cryptex.Version, err)
		}
		//id += 2
		golog.Info("uninstalled developer disk image cryptex", "module", logModule, "udid", c.entry.Properties.SerialNumber, "identifier", cryptex.Identifier, "version", cryptex.Version)
	}
	if !found {
		return fmt.Errorf("UnmountImage: no developer disk image cryptex is installed")
	}
	return nil
}

// MountImage installs the developer disk image at imagePath as a cryptex.
// imagePath needs to point to the 'Restore' directory of the developer disk image,
// and its BuildManifest needs a Cryptex1 build identity.
func (c *CryptexDeveloperDiskImageMounter) MountImage(imagePath string) error {
	udid := c.entry.Properties.SerialNumber
	manifest, err := loadBuildManifest(path.Join(imagePath, "BuildManifest.plist"))
	if err != nil {
		return fmt.Errorf("MountImage: failed to load build manifest: %w", err)
	}

	conn, err := c.connect()
	if err != nil {
		return fmt.Errorf("MountImage: %w", err)
	}
	chip, err := readCryptexPersonalizationIdentifiers(conn)
	if err != nil {
		conn.Close()
		return fmt.Errorf("MountImage: %w", err)
	}
	identity, err := manifest.findCryptexIdentity(chip)
	if err != nil {
		conn.Close()
		return fmt.Errorf("MountImage: %w", err)
	}
	nonceDomain, _ := toUint64(identity.Cryptex1NonceDomain)
	nonce, err := getCryptexNonce(conn, nonceDomain)
	conn.Close()
	if err != nil {
		return fmt.Errorf("MountImage: %w", err)
	}

	digests, err := identity.cryptexDigests(imagePath)
	if err != nil {
		return fmt.Errorf("MountImage: %w", err)
	}
	im4m, err := c.sign(cryptexTSSRequest(identity, chip, nonce, digests))
	if err != nil {
		return fmt.Errorf("MountImage: failed to get Cryptex1 ticket from Apple: %w", err)
	}

	// CoreDevice installs the cryptex on a connection of its own
	conn, err = c.connect()
	if err != nil {
		return fmt.Errorf("MountImage: %w", err)
	}
	defer conn.Close()
	installed, err := installCryptex(conn, imagePath, identity, im4m)
	if err != nil {
		return fmt.Errorf("MountImage: %w", err)
	}
	golog.Info("installed developer disk image cryptex", "module", logModule, "udid", udid, "imagePath", imagePath, "identifier", installed.Identifier, "version", installed.Version)
	return nil
}

// hasCryptexIdentity reports whether the developer disk image at imagePath can be installed as a cryptex
func hasCryptexIdentity(imagePath string) bool {
	manifest, err := loadBuildManifest(path.Join(imagePath, "BuildManifest.plist"))
	if err != nil {
		return false
	}
	for _, identity := range manifest.BuildIdentities {
		if identity.Cryptex1ChipID != nil {
			return true
		}
	}
	return false
}

// cryptexChip holds the img4 chip properties reported by read-personalization-id
type cryptexChip struct {
	chip, ecid         uint64
	fchp, typ, clas    uint64
	production, secure bool
}

func readCryptexPersonalizationIdentifiers(conn *xpc.Connection) (cryptexChip, error) {
	argv, err := cryptexCall(conn, 1, "read-personalization-id", map[string]interface{}{
		"client-version": cryptexClientVersion,
	})
	if err != nil {
		return cryptexChip{}, fmt.Errorf("readCryptexPersonalizationIdentifiers: %w", err)
	}
	return parseCryptexChip(argv)
}

func parseCryptexChip(argv map[string]interface{}) (cryptexChip, error) {
	get := func(name string) (uint64, error) {
		v, ok := argv["img4_chip_"+name].(uint64)
		if !ok {
			return 0, fmt.Errorf("parseCryptexChip: missing img4_chip_%s in %+v", name, argv)
		}
		return v, nil
	}
	var c cryptexChip
	var err error
	for _, f := range []struct {
		name string
		dst  *uint64
	}{{"chip", &c.chip}, {"ecid", &c.ecid}, {"fchp", &c.fchp}, {"type", &c.typ}, {"clas", &c.clas}} {
		if *f.dst, err = get(f.name); err != nil {
			return cryptexChip{}, err
		}
	}
	cpro, err := get("cpro")
	if err != nil {
		return cryptexChip{}, err
	}
	csec, err := get("csec")
	if err != nil {
		return cryptexChip{}, err
	}
	c.production = cpro != 0
	c.secure = csec != 0
	return c, nil
}

func getCryptexNonce(conn *xpc.Connection, nonceDomain uint64) ([]byte, error) {
	argv, err := cryptexCall(conn, 3, "get-nonce", map[string]interface{}{
		"nonce-domain-handle": nonceDomain,
	})
	if err != nil {
		return nil, fmt.Errorf("getCryptexNonce: %w", err)
	}
	raw, ok := argv["nonce"].([]byte)
	if !ok {
		return nil, fmt.Errorf("getCryptexNonce: no nonce in response %+v", argv)
	}
	return parseImg4Nonce(raw)
}

// parseImg4Nonce extracts the nonce from a serialized img4_nonce_t:
// uint16 version, 48 bytes of nonce, padding, uint32 length of the nonce.
func parseImg4Nonce(b []byte) ([]byte, error) {
	const maxNonceLength = 48
	if len(b) != 56 {
		return nil, fmt.Errorf("parseImg4Nonce: expected 56 bytes but got %d", len(b))
	}
	l := binary.LittleEndian.Uint32(b[52:])
	if l == 0 || l > maxNonceLength {
		return nil, fmt.Errorf("parseImg4Nonce: invalid nonce length %d", l)
	}
	return b[2 : 2+l], nil
}

// findCryptexIdentity returns the Cryptex1 build identity that matches the chip
func (m buildManifest) findCryptexIdentity(chip cryptexChip) (buildIdentity, error) {
	for _, i := range m.BuildIdentities {
		chipID, ok := toUint64(i.Cryptex1ChipID)
		if !ok || chipID != chip.fchp {
			continue
		}
		if t, ok := toUint64(i.Cryptex1Type); !ok || t != chip.typ {
			continue
		}
		if i.Cryptex1UseProductClass {
			if c, ok := toUint64(i.Cryptex1ProductClass); !ok || c != chip.clas {
				continue
			}
		}
		return i, nil
	}
	return buildIdentity{}, fmt.Errorf("findCryptexIdentity: no Cryptex1 identity for chip 0x%x, type 0x%x and product class 0x%x. Is this a recent developer disk image?", chip.fchp, chip.typ, chip.clas)
}

// cryptexDigests computes the SHA-384 digests of all manifest entries that get
// personalized, keyed by their manifest names. The device checks the files it
// receives against these digests, so they are computed from the files and
// compared with the build manifest.
func (b buildIdentity) cryptexDigests(imagePath string) (map[string][]byte, error) {
	digests := map[string][]byte{}
	for key, entry := range b.Manifest {
		if !strings.HasPrefix(key, "Cryptex1,") || !entry.Info.Personalize {
			continue
		}
		f, err := os.Open(path.Join(imagePath, entry.Info.Path))
		if err != nil {
			return nil, fmt.Errorf("cryptexDigests: %w", err)
		}
		h := sha512.New384()
		_, err = io.Copy(h, f)
		f.Close()
		if err != nil {
			return nil, fmt.Errorf("cryptexDigests: failed to hash %s: %w", entry.Info.Path, err)
		}
		digest := h.Sum(nil)
		if len(entry.Digest) != 0 && !bytes.Equal(entry.Digest, digest) {
			return nil, fmt.Errorf("cryptexDigests: digest of %s does not match the build manifest", entry.Info.Path)
		}
		digests[key] = digest
	}
	if len(digests) == 0 {
		return nil, fmt.Errorf("cryptexDigests: the Cryptex1 identity has no entries to personalize")
	}
	return digests, nil
}

// cryptexTSSRequest builds the personalization request for a Cryptex1 ticket,
// mirroring what CryptexKitHost sends. Unlike an ApImg4Ticket request it carries
// no ApECID/ApChipID/ApBoardID (Apple TSS rejects those here), the device is
// identified by Cryptex1,UDID instead, and the entries only carry digests.
func cryptexTSSRequest(identity buildIdentity, chip cryptexChip, nonce []byte, digests map[string][]byte) map[string]interface{} {
	subType, _ := toUint64(identity.Cryptex1SubType)
	nonceDomain, _ := toUint64(identity.Cryptex1NonceDomain)
	req := map[string]interface{}{
		"@Cryptex1,Ticket":                 true,
		"@HostPlatformInfo":                "mac",
		"@VersionInfo":                     "libauthinstall-1104.0.9",
		"@UUID":                            strings.ToUpper(uuid.New().String()),
		"ApSecurityMode":                   chip.secure,
		"Cryptex1,UDID":                    cryptexUDID(chip.chip, chip.ecid),
		"Cryptex1,ChipID":                  chip.fchp,
		"Cryptex1,Type":                    chip.typ,
		"Cryptex1,ProductClass":            chip.clas,
		"Cryptex1,ProductionMode":          chip.production,
		"Cryptex1,UseProductClass":         identity.Cryptex1UseProductClass,
		"Cryptex1,SubType":                 subType,
		"Cryptex1,NonceDomain":             nonceDomain,
		"Cryptex1,Nonce":                   nonce,
		"Cryptex1,Version":                 identity.Cryptex1Version,
		"Cryptex1,PreauthorizationVersion": identity.Cryptex1PreauthorizationVersion,
		// Without it the ticket lacks the 'uniq' tag
		"Cryptex1,UniqueTagList": []byte{},
	}
	for key, digest := range digests {
		req[key] = map[string]interface{}{"Digest": digest}
	}
	return req
}

// cryptexUDID is 4 zero bytes followed by the big endian chip id and ECID
func cryptexUDID(chip uint64, ecid uint64) []byte {
	udid := make([]byte, 16)
	binary.BigEndian.PutUint32(udid[4:], uint32(chip))
	binary.BigEndian.PutUint64(udid[8:], ecid)
	return udid
}

type cryptexUpload struct {
	arg  string
	open func() (io.ReadCloser, uint64, error)
}

func installCryptex(conn *xpc.Connection, imagePath string, identity buildIdentity, im4m []byte) (RemoteCryptex, error) {
	fromManifest := func(key string) func() (io.ReadCloser, uint64, error) {
		return func() (io.ReadCloser, uint64, error) {
			entry, ok := identity.Manifest[key]
			if !ok {
				return nil, 0, fmt.Errorf("the Cryptex1 identity has no %s entry", key)
			}
			p := path.Join(imagePath, entry.Info.Path)
			size, err := getFileSize(p)
			if err != nil {
				return nil, 0, err
			}
			f, err := os.Open(p)
			return f, size, err
		}
	}
	uploads := []cryptexUpload{
		{"image", fromManifest(cryptexDmgKey)},
		{"volumehash", fromManifest(cryptexVolumeKey)},
		{"info", fromManifest(cryptexInfoPlistKey)},
		{"trustcache", fromManifest(cryptexTrustCacheKey)},
		{"im4m", func() (io.ReadCloser, uint64, error) {
			return io.NopCloser(bytes.NewReader(im4m)), uint64(len(im4m)), nil
		}},
	}

	subType, _ := toUint64(identity.Cryptex1SubType)
	nonceDomain, _ := toUint64(identity.Cryptex1NonceDomain)
	argv := map[string]interface{}{
		"auth":              uint64(0),
		"client-version":    cryptexClientVersion,
		"image-type-index":  cryptexImageTypeIndex,
		"persistence":       cryptexPersistence,
		"nonce-persistence": cryptexNoncePersistence,
		"cryptex1-properties": map[string]interface{}{
			"Cryptex1,NonceDomain":     nonceDomain,
			"Cryptex1,PreauthVersion":  identity.Cryptex1PreauthorizationVersion,
			"Cryptex1,SubType":         subType,
			"Cryptex1,UseProductClass": identity.Cryptex1UseProductClass,
			"Cryptex1,Version":         identity.Cryptex1Version,
			"MountedCryptex":           false,
		},
	}

	// Each file goes on a stream of its own, announced with the message id that
	// the FileTransfer object in the request references. The install request
	// itself uses id 1.
	sources := make([]io.ReadCloser, len(uploads))
	defer func() {
		for _, s := range sources {
			if s != nil {
				s.Close()
			}
		}
	}()
	streams := make([]*xpc.FileTransferStream, len(uploads))
	for i, u := range uploads {
		src, size, err := u.open()
		if err != nil {
			return RemoteCryptex{}, fmt.Errorf("installCryptex: %s: %w", u.arg, err)
		}
		sources[i] = src
		id := uint64(2 + i)
		streams[i], err = conn.OpenFileTransfer(id)
		if err != nil {
			return RemoteCryptex{}, fmt.Errorf("installCryptex: %s: %w", u.arg, err)
		}
		argv[u.arg] = xpc.FileTransfer{MsgId: id, TransferSize: size}
	}

	err := conn.Send(map[string]interface{}{"routine": "install", "argv": argv}, xpc.HeartbeatRequestFlag)
	if err != nil {
		return RemoteCryptex{}, fmt.Errorf("installCryptex: failed to send install request: %w", err)
	}
	for i, u := range uploads {
		if err := streams[i].WaitAccepted(); err != nil {
			return RemoteCryptex{}, fmt.Errorf("installCryptex: device did not accept %s: %w", u.arg, err)
		}
		if _, err := io.Copy(streams[i], sources[i]); err != nil {
			return RemoteCryptex{}, fmt.Errorf("installCryptex: failed to send %s: %w", u.arg, err)
		}
		if err := streams[i].Close(); err != nil {
			return RemoteCryptex{}, fmt.Errorf("installCryptex: failed to finish %s: %w", u.arg, err)
		}
	}

	reply, err := conn.ReceiveOnServerClientStream()
	if err != nil {
		return RemoteCryptex{}, fmt.Errorf("installCryptex: failed to read install reply: %w", err)
	}
	result, err := parseCryptexReply("install", reply)
	if err != nil {
		return RemoteCryptex{}, fmt.Errorf("installCryptex: %w", err)
	}
	installed, _ := result["remote-cryptex"].(map[string]interface{})
	return parseRemoteCryptex(installed), nil
}

// cryptexCall invokes a cryptexd routine and returns the argv of its reply
func cryptexCall(conn *xpc.Connection, id uint64, routine string, argv map[string]interface{}) (map[string]interface{}, error) {
	err := conn.Send(map[string]interface{}{"routine": routine, "argv": argv}, xpc.HeartbeatRequestFlag)
	if err != nil {
		return nil, fmt.Errorf("cryptexCall: failed to send '%s': %w", routine, err)
	}
	reply, err := conn.ReceiveOnServerClientStream()
	if err != nil {
		return nil, fmt.Errorf("cryptexCall: failed to read reply for '%s': %w", routine, err)
	}
	return parseCryptexReply(routine, reply)
}

// parseCryptexReply checks the reply of a routine for errors. cryptexd reports
// them either as errno in 'error' or as serialized CFError in 'cferr'.
func parseCryptexReply(routine string, reply map[string]interface{}) (map[string]interface{}, error) {
	if cferr, ok := reply["cferr"]; ok {
		return nil, fmt.Errorf("'%s' failed: %v", routine, cferr)
	}
	if errno, ok := reply["error"].(uint64); ok && errno != 0 {
		return nil, fmt.Errorf("'%s' failed: %w", routine, syscall.Errno(errno))
	}
	argv, _ := reply["argv"].(map[string]interface{})
	if argv == nil {
		argv = map[string]interface{}{}
	}
	return argv, nil
}

// toUint64 converts build manifest numbers, which are either integers or hex strings
func toUint64(v interface{}) (uint64, bool) {
	switch n := v.(type) {
	case uint64:
		return n, true
	case int64:
		return uint64(n), n >= 0
	case string:
		i, err := strconv.ParseUint(n, 0, 64)
		return i, err == nil
	}
	return 0, false
}
