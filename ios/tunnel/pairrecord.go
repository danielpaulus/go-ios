package tunnel

import (
	"crypto/rand"
	"errors"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/crypto/ed25519"
	"howett.net/plist"
)

type selfIdentity struct {
	Identifier string `plist:"identifier"`
	Irk        []byte `plist:"irk"`
	PrivateKey []byte `plist:"privateKey"`
	PublicKey  []byte `plist:"publicKey"`
}

func (s selfIdentity) publicKey() ed25519.PublicKey {
	return s.PublicKey
}

func (s selfIdentity) privateKey() ed25519.PrivateKey {
	return ed25519.NewKeyFromSeed(s.PrivateKey)
}

type device struct {
	Identifier string `plist:"identifier"`
	Info       []byte `plist:"info"`
	Irk        []byte `plist:"irk"`
	Model      string `plist:"model"`
	Name       string `plist:"name"`
	PublicKey  []byte `plist:"publicKey"`
}

// PairRecordManager implements the same logic as macOS related to remote pair records. Those pair records are used
// whenever a tunnel gets created.
type PairRecordManager struct {
	selfId        selfIdentity
	peersLocation string
}

// NewPairRecordManager creates a PairRecordManager that reads/stores the pair records information at the given path
// To use the same pair records as macOS does, this path should be /var/db/lockdown/RemotePairing/user_501
// (user_501 is the default for the root user)
func NewPairRecordManager(p string) (PairRecordManager, error) {
	selfIdPath := path.Join(p, "selfIdentity.plist")
	selfId, err := getOrCreateSelfIdentity(selfIdPath)
	if err != nil {
		return PairRecordManager{}, fmt.Errorf("NewPairRecordManager: failed to get self identity: %w", err)
	}
	return PairRecordManager{
		selfId:        selfId,
		peersLocation: path.Join(p, "peers"),
	}, nil
}

// StoreDeviceInfo stores the provided Device info as a plist encoded file in the `peers/` directory
func (p PairRecordManager) StoreDeviceInfo(d device) error {
	devicePath := path.Join(p.peersLocation, fmt.Sprintf("%s.plist", d.Identifier))
	f, err := os.OpenFile(devicePath, os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("StoreDeviceInfo: could open file for writing: %w", err)
	}
	defer f.Close()

	enc := plist.NewEncoderForFormat(f, plist.BinaryFormat)
	err = enc.Encode(d)
	if err != nil {
		return fmt.Errorf("StoreDeviceInfo: could not encode device info: %w", err)
	}
	return nil
}

func readSelfIdentity(p string) (selfIdentity, error) {
	content, err := os.ReadFile(p)
	if err != nil {
		return selfIdentity{}, fmt.Errorf("readSelfIdentity: could not read file: %w", err)
	}
	var s selfIdentity
	_, err = plist.Unmarshal(content, &s)
	if err != nil {
		return selfIdentity{}, fmt.Errorf("readSelfIdentity: could not parse plist content: %w", err)
	}

	return s, nil
}

func getOrCreateSelfIdentity(p string) (selfIdentity, error) {
	info, err := os.Stat(p)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return createSelfIdentity(p)
		} else {
			return selfIdentity{}, fmt.Errorf("getOrCreateSelfIdentity: failed to get file info: %w", err)
		}
	}
	if info.IsDir() {
		return selfIdentity{}, fmt.Errorf("getOrCreateSelfIdentity: '%s' is a directory", p)
	}
	return readSelfIdentity(p)
}

func createSelfIdentity(p string) (selfIdentity, error) {
	irk := make([]byte, 16)
	_, _ = rand.Read(irk)

	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return selfIdentity{}, fmt.Errorf("createSelfIdentity: failed to create key pair: %w", err)
	}

	si := selfIdentity{
		Identifier: strings.ToUpper(uuid.New().String()),
		Irk:        irk,
		PrivateKey: priv.Seed(),
		PublicKey:  pub,
	}

	f, err := os.OpenFile(p, os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return selfIdentity{}, fmt.Errorf("createSelfIdentity: failed to open file for writing: %w", err)
	}
	defer f.Close()

	enc := plist.NewEncoderForFormat(f, plist.BinaryFormat)
	err = enc.Encode(si)
	if err != nil {
		return selfIdentity{}, fmt.Errorf("createSelfIdentity: failed to encode self identity as plist: %w", err)
	}

	return si, nil
}
