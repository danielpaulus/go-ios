package tunnel

import (
	"crypto/rand"
	"fmt"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ed25519"
	"howett.net/plist"
	"io"
	"os"
	"path"
)

type PairRecord struct {
	Public   ed25519.PublicKey
	Private  ed25519.PrivateKey
	HostName string
}

type pairRecordData struct {
	Seed     []byte
	Hostname string
}

func NewPairRecord() (PairRecord, error) {
	hostname, err := os.Hostname()
	if err != nil {
		log.WithError(err).Warn("could not get hostname. generate a random one")
	}
	hostname = uuid.New().String() // FIXME: this should be the hostname, but pairing fails with it
	var pairRecord PairRecord
	priv, pub, err := ed25519.GenerateKey(rand.Reader)
	pairRecord.Public = priv
	pairRecord.Private = pub
	pairRecord.HostName = hostname
	if err != nil {
		return PairRecord{}, err
	}
	log.WithField("hostname", pairRecord.HostName).Info("created new pair record")
	return pairRecord, nil
}

func ParsePairRecord(r io.ReadSeeker) (PairRecord, error) {
	dec := plist.NewDecoder(r)
	var seed pairRecordData
	err := dec.Decode(&seed)
	if err != nil {
		return PairRecord{}, err
	}
	key := ed25519.NewKeyFromSeed(seed.Seed)
	return PairRecord{
		Public:   key.Public().(ed25519.PublicKey),
		Private:  key,
		HostName: seed.Hostname,
	}, err
}

func StorePairRecord(w io.Writer, p PairRecord) error {
	enc := plist.NewEncoderForFormat(w, plist.BinaryFormat)
	return enc.Encode(pairRecordData{
		Seed:     p.Private.Seed(),
		Hostname: p.HostName,
	})
}

type PairRecordStore struct {
	p string
}

func NewPairRecordStore(directory string) PairRecordStore {
	return PairRecordStore{p: directory}
}

func (p PairRecordStore) Load(udid string) (PairRecord, error) {
	f, err := os.Open(p.pairRecordPath(udid))
	if err != nil {
		return PairRecord{}, err
	}
	defer f.Close()

	return ParsePairRecord(f)
}

func (p PairRecordStore) Store(udid string, pr PairRecord) error {
	f, err := os.OpenFile(p.pairRecordPath(udid), os.O_CREATE|os.O_WRONLY, 0o666)
	if err != nil {
		return err
	}
	defer f.Close()
	return StorePairRecord(f, pr)
}

func (p PairRecordStore) pairRecordPath(udid string) string {
	return path.Join(p.p, fmt.Sprintf("%s.plist", udid))
}

func (p PairRecordStore) LoadOrCreate(udid string) (PairRecord, error) {
	pr, err := p.Load(udid)
	if err != nil {
		log.WithError(err).Info("could load pair record. creating new one")
		return NewPairRecord()
	}
	return pr, nil
}
