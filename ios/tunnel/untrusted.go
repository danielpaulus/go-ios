package tunnel

import (
	"bytes"
	"context"
	"crypto/cipher"
	"crypto/ecdh"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha512"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"github.com/danielpaulus/go-ios/ios/opack"
	"github.com/danielpaulus/go-ios/ios/xpc"
	"github.com/google/uuid"
	"github.com/quic-go/quic-go"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/ed25519"
	"golang.org/x/crypto/hkdf"
	"io"
	"math/big"
	"net"
	"sync/atomic"
	"time"
)

const UntrustedTunnelServiceName = "com.apple.internal.dt.coredevice.untrusted.tunnelservice"

func NewTunnelServiceWithXpc(xpcConn *xpc.Connection, c io.Closer) (*TunnelService, error) {
	key, err := ecdh.P256().GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}
	sequence := atomic.Uint64{}
	sequence.Store(1)
	return &TunnelService{xpcConn: xpcConn, c: c, key: key, messageReadWriter: newControlChannelCodec()}, nil
}

type TunnelService struct {
	xpcConn *xpc.Connection
	c       io.Closer
	key     *ecdh.PrivateKey

	clientEncryption  cipher.AEAD
	serverEncryption  cipher.AEAD
	cs                *cipherStream
	messageReadWriter *controlChannelCodec
}

func (receiver *TunnelService) Close() error {
	return receiver.c.Close()
}

func (receiver *TunnelService) Pair() error {
	err := receiver.xpcConn.Send(EncodeRequest(receiver.messageReadWriter, map[string]interface{}{
		"handshake": map[string]interface{}{
			"_0": map[string]interface{}{
				"hostOptions": map[string]interface{}{
					"attemptPairVerify": false,
				},
				"wireProtocolVersion": int64(19),
			},
		},
	}))

	if err != nil {
		return err
	}
	m, err := receiver.xpcConn.ReceiveOnClientServerStream()
	if err != nil {
		return err
	}
	m, err = receiver.messageReadWriter.Decode(m)

	err = receiver.setupManualPairing()
	if err != nil {
		return err
	}

	devPublicKey, devSaltKey, err := receiver.readDeviceKey()
	if err != nil {
		return err
	}

	srp, err := NewSrpInfo(devSaltKey, devPublicKey)
	if err != nil {
		return err
	}

	proofTlv := NewTlvBuffer()
	proofTlv.WriteByte(TypeState, PairStateVerifyRequest)
	proofTlv.WriteData(TypePublicKey, srp.ClientPublic)
	proofTlv.WriteData(TypeProof, srp.ClientProof)

	err = receiver.xpcConn.Send(EncodeEvent(receiver.messageReadWriter, &pairingData{
		data: proofTlv.Bytes(),
		kind: "setupManualPairing",
	}))
	if err != nil {
		return err
	}

	m, err = receiver.xpcConn.ReceiveOnClientServerStream()
	if err != nil {
		return err
	}

	var proofPairingData pairingData
	DecodeEvent(receiver.messageReadWriter, m, &proofPairingData)

	serverProof, err := TlvReader(proofPairingData.data).ReadCoalesced(TypeProof)
	if err != nil {
		return err
	}
	verified := srp.VerifyServerProof(serverProof)
	if !verified {
		return fmt.Errorf("could not verify server proof")
	}

	identifier := uuid.New()
	public, private, err := ed25519.GenerateKey(rand.Reader)
	hkdfPairSetup := hkdf.New(sha512.New, srp.SessionKey, []byte("Pair-Setup-Controller-Sign-Salt"), []byte("Pair-Setup-Controller-Sign-Info"))
	buf := bytes.NewBuffer(nil)
	io.CopyN(buf, hkdfPairSetup, 32)
	buf.WriteString(identifier.String())
	buf.Write(public)

	if err != nil {
		return err
	}
	signature := ed25519.Sign(private, buf.Bytes())

	deviceInfo, err := opack.Encode(map[string]interface{}{
		"accountID":                   identifier.String(),
		"altIRK":                      []byte{0x5e, 0xca, 0x81, 0x91, 0x92, 0x02, 0x82, 0x00, 0x11, 0x22, 0x33, 0x44, 0xbb, 0xf2, 0x4a, 0xc8},
		"btAddr":                      "FF:DD:99:66:BB:AA",
		"mac":                         []byte{0xff, 0x44, 0x88, 0x66, 0x33, 0x99},
		"model":                       "MacBookPro18,3",
		"name":                        "host-name",
		"remotepairing_serial_number": "YY9944YY99",
	})

	deviceInfoTlv := NewTlvBuffer()
	deviceInfoTlv.WriteData(TypeSignature, signature)
	deviceInfoTlv.WriteData(TypePublicKey, public)
	deviceInfoTlv.WriteData(TypeIdentifier, []byte(identifier.String()))
	deviceInfoTlv.WriteData(TypeInfo, deviceInfo)

	sessionKeyBuf := bytes.NewBuffer(nil)
	_, err = io.CopyN(sessionKeyBuf, hkdf.New(sha512.New, srp.SessionKey, []byte("Pair-Setup-Encrypt-Salt"), []byte("Pair-Setup-Encrypt-Info")), 32)
	if err != nil {
		return err
	}
	setupKey := sessionKeyBuf.Bytes()

	cipher, err := chacha20poly1305.New(setupKey)
	if err != nil {
		return err
	}

	//deviceInfoLen := len(deviceInfoTlv.Bytes())
	nonce := make([]byte, cipher.NonceSize())
	for x, y := range "PS-Msg05" {
		nonce[4+x] = byte(y)
	}
	x := cipher.Seal(nil, nonce, deviceInfoTlv.Bytes(), nil)

	encryptedTlv := NewTlvBuffer()
	encryptedTlv.WriteByte(TypeState, 0x05)
	encryptedTlv.WriteData(TypeEncryptedData, x)

	err = receiver.xpcConn.Send(EncodeEvent(receiver.messageReadWriter, &pairingData{
		data:        encryptedTlv.Bytes(),
		kind:        "setupManualPairing",
		sendingHost: "SL-1876",
	}))
	if err != nil {
		return err
	}

	m, err = receiver.xpcConn.ReceiveOnClientServerStream()
	log.WithField("data", m).Info("response")
	if err != nil {
		return err
	}

	var encRes pairingData
	err = DecodeEvent(receiver.messageReadWriter, m, &encRes)
	if err != nil {
		return err
	}

	encrData, err := TlvReader(encRes.data).ReadCoalesced(TypeEncryptedData)
	if err != nil {
		return err
	}
	copy(nonce[4:], "PS-Msg06")
	decrypted, err := cipher.Open(nil, nonce, encrData, nil)

	log.WithField("decrypted", hex.EncodeToString(decrypted)).Info("response")
	log.Printf("%s", decrypted)

	err = receiver.setupCiphers(srp.SessionKey)
	if err != nil {
		return err
	}

	receiver.cs = &cipherStream{}

	_, err = receiver.createUnlockKey()

	return err
}

func (t *TunnelService) CreateTunnelListener() (TunnelInfo, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return TunnelInfo{}, err
	}
	der := x509.MarshalPKCS1PublicKey(&privateKey.PublicKey)
	fmt.Printf("tunnel key: %s\n", hex.EncodeToString(der))
	if err != nil {
		return TunnelInfo{}, err
	}

	createListenerRequest, err := EncodeStreamEncrypted(t.messageReadWriter, t.clientEncryption, t.cs, map[string]interface{}{
		"request": map[string]interface{}{
			"_0": map[string]interface{}{
				"createListener": map[string]interface{}{
					"key":                   der,
					"transportProtocolType": "tcp",
				},
			},
		},
	})
	if err != nil {
		return TunnelInfo{}, err
	}
	err = t.xpcConn.Send(createListenerRequest)
	if err != nil {
		return TunnelInfo{}, err
	}

	m, err := t.xpcConn.ReceiveOnClientServerStream()
	if err != nil {
		return TunnelInfo{}, err
	}

	listenerRes, err := DecodeStreamEncrypted(t.messageReadWriter, t.serverEncryption, t.cs, m)
	//listenerRes := new(cipherMessage)
	//err = listenerRes.Decode(t.serverEncryption, t.cs, m)
	if err != nil {
		return TunnelInfo{}, err
	}
	log.Infof("Tunnel listener %v", listenerRes)

	createListener, err := getChildMap(listenerRes, "response", "_1", "createListener")
	if err != nil {
		return TunnelInfo{}, err
	}
	port := createListener["port"].(float64)
	devPublicKey := createListener["devicePublicKey"].(string)
	devPK, err := base64.StdEncoding.DecodeString(devPublicKey)
	if err != nil {
		return TunnelInfo{}, err
	}
	publicKey, err := x509.ParsePKIXPublicKey(devPK)
	if err != nil {
		return TunnelInfo{}, err
	}
	return TunnelInfo{
		PrivateKey:      privateKey,
		DevicePublicKey: publicKey,
		TunnelPort:      uint64(port),
	}, nil
}

func (t *TunnelService) setupCiphers(sessionKey []byte) error {
	clientKey := make([]byte, 32)
	_, err := hkdf.New(sha512.New, sessionKey, nil, []byte("ClientEncrypt-main")).Read(clientKey)
	if err != nil {
		return err
	}
	serverKey := make([]byte, 32)
	_, err = hkdf.New(sha512.New, sessionKey, nil, []byte("ServerEncrypt-main")).Read(serverKey)
	if err != nil {
		return err
	}
	t.serverEncryption, err = chacha20poly1305.New(serverKey)
	if err != nil {
		return err
	}
	t.clientEncryption, err = chacha20poly1305.New(clientKey)
	if err != nil {
		return err
	}
	return nil
}

func (t *TunnelService) setupManualPairing() error {
	buf := NewTlvBuffer()
	buf.WriteByte(TypeMethod, 0x00)
	buf.WriteByte(TypeState, PairStateStartRequest)

	event := pairingData{
		data:            buf.Bytes(),
		kind:            "setupManualPairing",
		startNewSession: true,
	}

	err := t.xpcConn.Send(EncodeEvent(t.messageReadWriter, &event))
	if err != nil {
		return err
	}
	res, err := t.xpcConn.ReceiveOnClientServerStream()
	if err != nil {
		return err
	}
	_, err = t.messageReadWriter.Decode(res)
	return err
}

func (t *TunnelService) readDeviceKey() (publicKey []byte, salt []byte, err error) {
	m, err := t.xpcConn.ReceiveOnClientServerStream()
	if err != nil {
		return
	}
	var pairingData pairingData
	err = DecodeEvent(t.messageReadWriter, m, &pairingData)
	if err != nil {
		return
	}
	publicKey, err = TlvReader(pairingData.data).ReadCoalesced(TypePublicKey)
	if err != nil {
		return
	}
	salt, err = TlvReader(pairingData.data).ReadCoalesced(TypeSalt)
	if err != nil {
		return
	}
	return
}

func (t *TunnelService) createUnlockKey() ([]byte, error) {
	unlockReqMsg, err := EncodeStreamEncrypted(t.messageReadWriter, t.clientEncryption, t.cs, map[string]interface{}{
		"request": map[string]interface{}{
			"_0": map[string]interface{}{
				"createRemoteUnlockKey": map[string]interface{}{},
			},
		},
	})

	err = t.xpcConn.Send(unlockReqMsg)
	if err != nil {
		return nil, err
	}

	m, err := t.xpcConn.ReceiveOnClientServerStream()
	if err != nil {
		return nil, err
	}

	_, _ = DecodeStreamEncrypted(t.messageReadWriter, t.serverEncryption, t.cs, m)
	return nil, err
}

type TunnelInfo struct {
	PrivateKey      *rsa.PrivateKey
	DevicePublicKey interface{}
	TunnelPort      uint64
}

func ConnectToTunnel(info TunnelInfo, addr string) error {
	log.WithField("address", addr).WithField("port", info.TunnelPort).Info("connect to tunnel")
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(2019),
		Subject: pkix.Name{
			Organization: []string{"go-ios"},
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().AddDate(10, 0, 0),
	}

	cert, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &info.PrivateKey.PublicKey, info.PrivateKey)
	if err != nil {
		return err
	}
	privateKeyPem := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(info.PrivateKey),
		},
	)
	certPem := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert,
	})
	cert5, err := tls.X509KeyPair(certPem, privateKeyPem)

	conf := &tls.Config{
		InsecureSkipVerify: true,
		Certificates:       []tls.Certificate{cert5},
		ClientAuth:         tls.NoClientCert,
	}

	udpAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("[%s]:%d", addr, info.TunnelPort))
	if err != nil {
		return err
	}
	udpConn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		return err
	}

	//conn, err := quic.Dial(context.TODO(), udpConn, udpAddr, conf, nil)
	conn, err := quic.DialAddr(context.TODO(), fmt.Sprintf("[%s]:%d", addr, info.TunnelPort), conf, nil)
	if err != nil {
		return err
	}
	defer udpConn.Close()
	conn.Context()
	return nil
}
