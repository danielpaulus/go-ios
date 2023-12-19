package tunnel

import (
	"crypto/sha512"
	"fmt"
	"github.com/tadglines/go-pkgs/crypto/srp"
)

type SrpInfo struct {
	ClientPublic []byte
	ClientProof  []byte
	Salt         []byte
	SessionKey   []byte
	c            *srp.ClientSession
}

func NewSrpInfo(salt, publicKey []byte) (SrpInfo, error) {
	s, err := srp.NewSRP("rfc5054.3072", sha512.New, func(salt, password []byte) []byte {
		h1 := sha512.New()
		h2 := sha512.New()
		h2.Write([]byte(fmt.Sprintf("%s:%s", "Pair-Setup", string(password))))
		h1.Write(salt)
		h1.Write(h2.Sum(nil))
		return h1.Sum(nil)
	})
	if err != nil {
		return SrpInfo{}, err
	}
	c := s.NewClientSession([]byte("Pair-Setup"), []byte("000000"))
	if err != nil {
		return SrpInfo{}, err
	}
	key, err := c.ComputeKey(salt, publicKey)
	if err != nil {
		return SrpInfo{}, err
	}
	a := c.ComputeAuthenticator()
	return SrpInfo{
		ClientPublic: c.GetA(),
		ClientProof:  a,
		Salt:         salt,
		SessionKey:   key,
		c:            c,
	}, nil
}

func (s SrpInfo) VerifyServerProof(p []byte) bool {
	return s.c.VerifyServerAuthenticator(p)
}
