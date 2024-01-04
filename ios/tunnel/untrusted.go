package tunnel

import (
	"bytes"
	"crypto/ecdh"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha512"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"github.com/danielpaulus/go-ios/ios/opack"
	"github.com/danielpaulus/go-ios/ios/xpc"
	//"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/ed25519"
	"golang.org/x/crypto/hkdf"
	"io"
)

const UntrustedTunnelServiceName = "com.apple.internal.dt.coredevice.untrusted.tunnelservice"

func NewTunnelServiceWithXpc(xpcConn *xpc.Connection, c io.Closer, pairRecords PairRecordManager) (*TunnelService, error) {
	return &TunnelService{
		xpcConn:        xpcConn,
		c:              c,
		controlChannel: newControlChannelReadWriter(xpcConn),
		pairRecords:    pairRecords,
	}, nil
}

type TunnelService struct {
	xpcConn *xpc.Connection
	c       io.Closer

	controlChannel *controlChannelReadWriter
	cipher         *cipherStream

	pairRecords PairRecordManager
}

func (t *TunnelService) Close() error {
	return t.c.Close()
}

func (t *TunnelService) Pair() error {
	err := t.controlChannel.writeRequest(map[string]interface{}{
		"handshake": map[string]interface{}{
			"_0": map[string]interface{}{
				"hostOptions": map[string]interface{}{
					"attemptPairVerify": true,
				},
				"wireProtocolVersion": int64(19),
			},
		},
	})

	if err != nil {
		return err
	}
	// ignore the response for now
	_, err = t.controlChannel.read()
	if err != nil {
		return err
	}

	err = t.verifyPair()
	if err == nil {
		return nil
	}
	log.WithError(err).Info("pair verify failed")

	err = t.setupManualPairing()
	if err != nil {
		return err
	}

	sessionKey, err := t.setupSessionKey()
	if err != nil {
		return err
	}

	err = t.exchangeDeviceInfo(sessionKey)
	if err != nil {
		return err
	}

	err = t.setupCiphers(sessionKey)
	if err != nil {
		return err
	}

	_, err = t.createUnlockKey()
	if err != nil {
		return err
	}

	return nil
}

func (t *TunnelService) CreateTunnelListener() (TunnelListener, error) {
	log.Info("create tunnel listener")
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)

	if err != nil {
		return TunnelListener{}, err
	}
	der, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return TunnelListener{}, err
	}

	err = t.cipher.write(map[string]interface{}{
		"request": map[string]interface{}{
			"_0": map[string]interface{}{
				"createListener": map[string]interface{}{
					"key":                   der,
					"transportProtocolType": "quic",
				},
			},
		},
	})
	if err != nil {
		return TunnelListener{}, err
	}

	var listenerRes map[string]interface{}
	err = t.cipher.read(&listenerRes)
	if err != nil {
		return TunnelListener{}, err
	}

	createListener, err := getChildMap(listenerRes, "response", "_1", "createListener")
	if err != nil {
		return TunnelListener{}, err
	}
	port := createListener["port"].(float64)
	devPublicKey := createListener["devicePublicKey"].(string)
	devPK, err := base64.StdEncoding.DecodeString(devPublicKey)
	if err != nil {
		return TunnelListener{}, err
	}
	publicKey, err := x509.ParsePKIXPublicKey(devPK)
	if err != nil {
		return TunnelListener{}, err
	}
	return TunnelListener{
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
	server, err := chacha20poly1305.New(serverKey)
	if err != nil {
		return err
	}
	client, err := chacha20poly1305.New(clientKey)
	if err != nil {
		return err
	}

	t.cipher = newCipherStream(t.controlChannel, client, server)

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

	err := t.controlChannel.writeEvent(&event)
	if err != nil {
		return err
	}
	_, err = t.controlChannel.read()
	if err != nil {
		return err
	}
	return err
}

func (t *TunnelService) readDeviceKey() (publicKey []byte, salt []byte, err error) {
	var pairingData pairingData
	err = t.controlChannel.readEvent(&pairingData)
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
	err := t.cipher.write(map[string]interface{}{
		"request": map[string]interface{}{
			"_0": map[string]interface{}{
				"createRemoteUnlockKey": map[string]interface{}{},
			},
		},
	})
	if err != nil {
		return nil, err
	}

	var res map[string]interface{}
	err = t.cipher.read(&res)
	if err != nil {
		return nil, err
	}

	return nil, err
}

func (t *TunnelService) verifyPair() error {
	key, _ := ecdh.X25519().GenerateKey(rand.Reader)
	tlv := NewTlvBuffer()
	tlv.WriteByte(TypeState, PairStateStartRequest)
	tlv.WriteData(TypePublicKey, key.PublicKey().Bytes())

	event := pairingData{
		data:            tlv.Bytes(),
		kind:            "verifyManualPairing",
		startNewSession: true,
	}

	selfId := t.pairRecords.SelfId

	err := t.controlChannel.writeEvent(&event)

	var devP pairingData
	err = t.controlChannel.readEvent(&devP)
	if err != nil {
		return err
	}

	devicePublicKeyBytes, err := TlvReader(devP.data).ReadCoalesced(TypePublicKey)
	if err != nil {
		return err
	}

	devicePublicKey, err := ecdh.X25519().NewPublicKey(devicePublicKeyBytes)
	if err != nil {
		return err
	}

	sharedSecret, err := key.ECDH(devicePublicKey)
	if err != nil {
		return err
	}

	derived := make([]byte, 32)
	_, err = hkdf.New(sha512.New, sharedSecret, []byte("Pair-Verify-Encrypt-Salt"), []byte("Pair-Verify-Encrypt-Info")).Read(derived)
	if err != nil {
		return err
	}

	ci, err := chacha20poly1305.New(derived)
	if err != nil {
		return err
	}

	signBuf := bytes.NewBuffer(nil)
	signBuf.Write(key.PublicKey().Bytes())
	signBuf.Write([]byte(selfId.Identifier))
	signBuf.Write(devicePublicKeyBytes)

	signature := ed25519.Sign(selfId.PrivateKey, signBuf.Bytes())

	cTlv := NewTlvBuffer()
	cTlv.WriteData(TypeSignature, signature)
	cTlv.WriteData(TypeIdentifier, []byte(selfId.Identifier))

	nonce := make([]byte, 12)
	copy(nonce[4:], "PV-Msg03")
	encrypted := ci.Seal(nil, nonce, cTlv.Bytes(), nil)

	tlvOut := NewTlvBuffer()
	tlvOut.WriteByte(TypeState, PairStateVerifyRequest)
	tlvOut.WriteData(TypeEncryptedData, encrypted)

	pd := pairingData{
		data:            tlvOut.Bytes(),
		kind:            "verifyManualPairing",
		startNewSession: false,
	}

	err = t.controlChannel.writeEvent(&pd)
	if err != nil {
		return err
	}

	var responseEvent pairingData
	err = t.controlChannel.readEvent(&responseEvent)
	if err != nil {
		return err
	}

	errRes, err := TlvReader(responseEvent.data).ReadCoalesced(TypeError)
	if err != nil {
		return err
	}
	if len(errRes) > 0 {
		log.Debug("send pair verify failed event")
		err := t.controlChannel.writeEvent(pairVerifyFailed{})
		if err != nil {
			return err
		}
		return Error(errRes[0])
	}

	err = t.setupCiphers(sharedSecret)
	if err != nil {
		return err
	}

	return nil
}

type TunnelListener struct {
	PrivateKey      *rsa.PrivateKey
	DevicePublicKey interface{}
	TunnelPort      uint64
}

type TunnelInfo struct {
	ServerAddress    string
	ServerRSDPort    uint64
	ClientParameters struct {
		Address string
		Netmask string
		Mtu     uint64
	}
}

func (t *TunnelService) setupSessionKey() ([]byte, error) {
	devicePublicKey, deviceSalt, err := t.readDeviceKey()
	if err != nil {
		return nil, err
	}

	srp, err := NewSrpInfo(deviceSalt, devicePublicKey)
	if err != nil {
		return nil, err
	}

	proofTlv := NewTlvBuffer()
	proofTlv.WriteByte(TypeState, PairStateVerifyRequest)
	proofTlv.WriteData(TypePublicKey, srp.ClientPublic)
	proofTlv.WriteData(TypeProof, srp.ClientProof)

	err = t.controlChannel.writeEvent(&pairingData{
		data: proofTlv.Bytes(),
		kind: "setupManualPairing",
	})
	if err != nil {
		return nil, err
	}

	var proofPairingData pairingData
	err = t.controlChannel.readEvent(&proofPairingData)
	if err != nil {
		return nil, err
	}

	serverProof, err := TlvReader(proofPairingData.data).ReadCoalesced(TypeProof)
	if err != nil {
		return nil, err
	}
	verified := srp.VerifyServerProof(serverProof)
	if !verified {
		return nil, fmt.Errorf("could not verify server proof")
	}
	return srp.SessionKey, nil
}

func (t *TunnelService) exchangeDeviceInfo(sessionKey []byte) error {
	hkdfPairSetup := hkdf.New(sha512.New, sessionKey, []byte("Pair-Setup-Controller-Sign-Salt"), []byte("Pair-Setup-Controller-Sign-Info"))
	buf := bytes.NewBuffer(nil)
	io.CopyN(buf, hkdfPairSetup, 32)
	buf.WriteString(t.pairRecords.SelfId.Identifier)
	buf.Write(t.pairRecords.SelfId.PublicKey)

	signature := ed25519.Sign(t.pairRecords.SelfId.PrivateKey, buf.Bytes())

	deviceInfo, err := opack.Encode(map[string]interface{}{
		"accountID":                   t.pairRecords.SelfId.Identifier,
		"altIRK":                      []byte{0x5e, 0xca, 0x81, 0x91, 0x92, 0x02, 0x82, 0x00, 0x11, 0x22, 0x33, 0x44, 0xbb, 0xf2, 0x4a, 0xc8},
		"btAddr":                      "FF:DD:99:66:BB:AA",
		"mac":                         []byte{0xff, 0x44, 0x88, 0x66, 0x33, 0x99},
		"model":                       "go-ios",
		"name":                        "host-name",
		"remotepairing_serial_number": "remote-serial",
	})

	deviceInfoTlv := NewTlvBuffer()
	deviceInfoTlv.WriteData(TypeSignature, signature)
	deviceInfoTlv.WriteData(TypePublicKey, t.pairRecords.SelfId.PublicKey)
	deviceInfoTlv.WriteData(TypeIdentifier, []byte(t.pairRecords.SelfId.Identifier))
	deviceInfoTlv.WriteData(TypeInfo, deviceInfo)

	sessionKeyBuf := bytes.NewBuffer(nil)
	_, err = io.CopyN(sessionKeyBuf, hkdf.New(sha512.New, sessionKey, []byte("Pair-Setup-Encrypt-Salt"), []byte("Pair-Setup-Encrypt-Info")), 32)
	if err != nil {
		return err
	}
	setupKey := sessionKeyBuf.Bytes()

	cipher, err := chacha20poly1305.New(setupKey)
	if err != nil {
		return err
	}

	nonce := make([]byte, cipher.NonceSize())
	copy(nonce[4:], "PS-Msg05")
	x := cipher.Seal(nil, nonce, deviceInfoTlv.Bytes(), nil)

	encryptedTlv := NewTlvBuffer()
	encryptedTlv.WriteByte(TypeState, 0x05)
	encryptedTlv.WriteData(TypeEncryptedData, x)

	err = t.controlChannel.writeEvent(&pairingData{
		data:        encryptedTlv.Bytes(),
		kind:        "setupManualPairing",
		sendingHost: "SL-1876",
	})
	if err != nil {
		return err
	}

	var encRes pairingData
	err = t.controlChannel.readEvent(&encRes)
	if err != nil {
		return err
	}

	encrData, err := TlvReader(encRes.data).ReadCoalesced(TypeEncryptedData)
	if err != nil {
		return err
	}
	copy(nonce[4:], "PS-Msg06")
	// the device info response from the device is not needed. we just make sure that there's no error decrypting it
	// TODO: decode the opack encoded data and persist it using the PairRecordManager.StoreDeviceInfo method
	_, err = cipher.Open(nil, nonce, encrData, nil)
	if err != nil {
		return err
	}
	return nil
}
