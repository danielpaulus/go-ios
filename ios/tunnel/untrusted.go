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

	"io"

	"github.com/danielpaulus/go-ios/ios/opack"
	"github.com/danielpaulus/go-ios/ios/xpc"

	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/ed25519"
	"golang.org/x/crypto/hkdf"
)

// untrustedTunnelServiceName is the service name that is described in the Remote Service Discovery of the
// ethernet interface of the device (not the tunnel interface)
const untrustedTunnelServiceName = "com.apple.internal.dt.coredevice.untrusted.tunnelservice"

func newTunnelServiceWithXpc(xpcConn *xpc.Connection, c io.Closer, pairRecords PairRecordManager) *tunnelService {
	return &tunnelService{
		xpcConn:        xpcConn,
		c:              c,
		controlChannel: newControlChannelReadWriter(xpcConn),
		pairRecords:    pairRecords,
	}
}

// need to find a nicer name, my interface to decouple the implementation from the ncm xpc tunnel and the remote tunnel
type pairingService interface {
	getPairRecords() PairRecordManager
	setupCiphers(sharedSecret []byte) error
	writeEvent(event eventCodec) error
	readEvent(event eventCodec) error
	getCipher() *cipherStream
	writeRequest(req map[string]interface{}) error
	readResponse() (map[string]interface{}, error)
	writeEncrypted(msg map[string]interface{}) error
	readEncrypted() (map[string]interface{}, error)
}

type tunnelService struct {
	xpcConn *xpc.Connection
	c       io.Closer

	controlChannel *controlChannelReadWriter
	cipher         *cipherStream

	pairRecords PairRecordManager
}

func (t *tunnelService) Close() error {
	return t.c.Close()
}

func (t *tunnelService) readEncrypted() (map[string]interface{}, error) {
	return t.controlChannel.read()
}

func (t *tunnelService) writeEncrypted(msg map[string]interface{}) error {
	return t.controlChannel.write(msg)
}

func (t *tunnelService) getCipher() *cipherStream {
	return t.cipher
}

func (t *tunnelService) writeEvent(event eventCodec) error {
	return t.controlChannel.writeEvent(event)
}
func (t *tunnelService) readEvent(event eventCodec) error {
	return t.controlChannel.readEvent(event)
}

func (t *tunnelService) writeRequest(req map[string]interface{}) error {
	return t.controlChannel.writeRequest(req)
}

func (t *tunnelService) readResponse() (map[string]interface{}, error) {
	return t.controlChannel.read()
}

func (t *tunnelService) getPairRecords() PairRecordManager {
	return t.pairRecords
}

// ManualPair triggers a device pairing that requires the user to press the 'Trust' button on the device that appears
// when this operation is triggered
// If there is already an active pairing with the credentials stored in PairRecordManager this call does not trigger
// anything on the device and returns with an error
func ManualPair(t pairingService) error {
	err := t.writeRequest(map[string]interface{}{
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
		return fmt.Errorf("ManualPair: failed to send 'attemptPairVerify' request: %w", err)
	}
	// ignore the response for now
	r, err := t.readResponse()
	if err != nil {
		return fmt.Errorf("ManualPair: failed to read 'attemptPairVerify' response: %w", err)
	}
	print(r)

	err = verifyPair(t)
	if err == nil {
		return nil
	}
	log.WithError(err).Info("pair verify failed")

	err = setupManualPairing(t)
	if err != nil {
		return fmt.Errorf("ManualPair: failed to initiate manual pairing: %w", err)
	}

	sessionKey, err := setupSessionKey(t)
	if err != nil {
		return fmt.Errorf("ManualPair: failed to setup SRP session key: %w", err)
	}

	err = exchangeDeviceInfo(t, sessionKey)
	if err != nil {
		return fmt.Errorf("ManualPair: failed to exchange device info: %w", err)
	}

	err = t.setupCiphers(sessionKey)
	if err != nil {
		return fmt.Errorf("ManualPair: failed to setup session ciphers: %w", err)
	}

	_, err = createUnlockKey(t)
	if err != nil {
		return fmt.Errorf("ManualPair: failed to create unlock key: %w", err)
	}

	return nil
}

func createTunnelListener(t pairingService, protocol string) (tunnelListener, error) {
	log.Info("create tunnel listener")
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)

	if err != nil {
		return tunnelListener{}, err
	}
	der, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return tunnelListener{}, err
	}

	err = t.getCipher().write(map[string]interface{}{
		"request": map[string]interface{}{
			"_0": map[string]interface{}{
				"createListener": map[string]interface{}{
					"key":                   der,
					"transportProtocolType": protocol,
				},
			},
		},
	})
	if err != nil {
		return tunnelListener{}, err
	}

	var listenerRes map[string]interface{}
	err = t.getCipher().read(&listenerRes)
	if err != nil {
		return tunnelListener{}, err
	}

	createListener, err := getChildMap(listenerRes, "response", "_1", "createListener")
	if err != nil {
		return tunnelListener{}, err
	}
	port := createListener["port"].(float64)
	devPublicKey := createListener["devicePublicKey"].(string)
	devPK, err := base64.StdEncoding.DecodeString(devPublicKey)
	if err != nil {
		return tunnelListener{}, err
	}
	publicKey, err := x509.ParsePKIXPublicKey(devPK)
	if err != nil {
		return tunnelListener{}, err
	}
	return tunnelListener{
		PrivateKey:      privateKey,
		DevicePublicKey: publicKey,
		TunnelPort:      uint64(port),
	}, nil
}

func (t *tunnelService) setupCiphers(sessionKey []byte) error {
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

	t.cipher = newCipherStream(t, client, server)

	return nil
}

func setupManualPairing(t pairingService) error {
	buf := newTlvBuffer()
	buf.writeByte(typeMethod, 0x00)
	buf.writeByte(typeState, pairStateStartRequest)

	event := pairingData{
		data:            buf.bytes(),
		kind:            "setupManualPairing",
		startNewSession: true,
	}

	err := t.writeEvent(&event)
	if err != nil {
		return err
	}
	r, err := t.readResponse()
	if err != nil {
		return err
	}
	print(r)
	return err
}

func readDeviceKey(t pairingService) (publicKey []byte, salt []byte, err error) {
	var pairingData pairingData
	err = t.readEvent(&pairingData)
	if err != nil {
		return
	}
	publicKey, err = tlvReader(pairingData.data).readCoalesced(typePublicKey)
	if err != nil {
		return
	}
	salt, err = tlvReader(pairingData.data).readCoalesced(typeSalt)
	if err != nil {
		return
	}
	return
}

func createUnlockKey(t pairingService) ([]byte, error) {
	err := t.getCipher().write(map[string]interface{}{
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
	err = t.getCipher().read(&res)
	if err != nil {
		return nil, err
	}

	return nil, err
}

func verifyPair(t pairingService) error {
	key, _ := ecdh.X25519().GenerateKey(rand.Reader)
	tlv := newTlvBuffer()
	tlv.writeByte(typeState, pairStateStartRequest)
	tlv.writeData(typePublicKey, key.PublicKey().Bytes())

	event := pairingData{
		data:            tlv.bytes(),
		kind:            "verifyManualPairing",
		startNewSession: true,
	}

	selfId := t.getPairRecords().selfId

	err := t.writeEvent(&event)
	if err != nil {
		return err
	}
	var devP pairingData
	err = t.readEvent(&devP)
	if err != nil {
		return err
	}

	devicePublicKeyBytes, err := tlvReader(devP.data).readCoalesced(typePublicKey)
	if err != nil {
		return err
	}

	if devicePublicKeyBytes == nil {
		_ = t.writeEvent(&pairVerifyFailed{})
		return fmt.Errorf("verifyPair: did not get public key from device. Can not verify pairing")
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
	// Write on bytes.Buffer never returns an error
	_, _ = signBuf.Write(key.PublicKey().Bytes())
	_, _ = signBuf.Write([]byte(selfId.Identifier))
	_, _ = signBuf.Write(devicePublicKeyBytes)

	signature := ed25519.Sign(selfId.privateKey(), signBuf.Bytes())

	cTlv := newTlvBuffer()
	cTlv.writeData(typeSignature, signature)
	cTlv.writeData(typeIdentifier, []byte(selfId.Identifier))

	nonce := make([]byte, 12)
	copy(nonce[4:], "PV-Msg03")
	encrypted := ci.Seal(nil, nonce, cTlv.bytes(), nil)

	tlvOut := newTlvBuffer()
	tlvOut.writeByte(typeState, pairStateVerifyRequest)
	tlvOut.writeData(typeEncryptedData, encrypted)

	pd := pairingData{
		data:            tlvOut.bytes(),
		kind:            "verifyManualPairing",
		startNewSession: false,
	}

	err = t.writeEvent(&pd)
	if err != nil {
		return err
	}

	var responseEvent pairingData
	err = t.readEvent(&responseEvent)
	if err != nil {
		return err
	}

	errRes, err := tlvReader(responseEvent.data).readCoalesced(typeError)
	if err != nil {
		return err
	}
	if len(errRes) > 0 {
		log.Debug("send pair verify failed event")
		err := t.writeEvent(pairVerifyFailed{})
		if err != nil {
			return err
		}
		return tlvError(errRes[0])
	}

	err = t.setupCiphers(sharedSecret)
	if err != nil {
		return err
	}

	return nil
}

type tunnelListener struct {
	PrivateKey      *rsa.PrivateKey
	DevicePublicKey interface{}
	TunnelPort      uint64
}

type tunnelParameters struct {
	ServerAddress    string
	ServerRSDPort    uint64
	ClientParameters struct {
		Address string
		Netmask string
		Mtu     uint64
	}
}

func setupSessionKey(t pairingService) ([]byte, error) {
	devicePublicKey, deviceSalt, err := readDeviceKey(t)
	if err != nil {
		return nil, fmt.Errorf("setupSessionKey: failed to read device public key and salt value: %w", err)
	}

	srp, err := newSrpInfo(deviceSalt, devicePublicKey)
	if err != nil {
		return nil, fmt.Errorf("setupSessionKey: failed to setup SRP: %w", err)
	}

	proofTlv := newTlvBuffer()
	proofTlv.writeByte(typeState, pairStateVerifyRequest)
	proofTlv.writeData(typePublicKey, srp.ClientPublic)
	proofTlv.writeData(typeProof, srp.ClientProof)

	err = t.writeEvent(&pairingData{
		data: proofTlv.bytes(),
		kind: "setupManualPairing",
	})
	if err != nil {
		return nil, fmt.Errorf("setupSessionKey: failed to send SRP proof: %w", err)
	}

	var proofPairingData pairingData
	err = t.readEvent(&proofPairingData)
	if err != nil {
		return nil, fmt.Errorf("setupSessionKey: failed to read device SRP proof: %w", err)
	}

	serverProof, err := tlvReader(proofPairingData.data).readCoalesced(typeProof)
	if err != nil {
		return nil, fmt.Errorf("setupSessionKey: failed to parse device proof: %w", err)
	}
	verified := srp.verifyServerProof(serverProof)
	if !verified {
		return nil, fmt.Errorf("setupSessionKey: could not verify server proof")
	}
	return srp.SessionKey, nil
}

func exchangeDeviceInfo(t pairingService, sessionKey []byte) error {
	hkdfPairSetup := hkdf.New(sha512.New, sessionKey, []byte("Pair-Setup-Controller-Sign-Salt"), []byte("Pair-Setup-Controller-Sign-Info"))
	buf := bytes.NewBuffer(nil)
	// Write on bytes.Buffer never returns an error
	_, _ = io.CopyN(buf, hkdfPairSetup, 32)
	_, _ = buf.WriteString(t.getPairRecords().selfId.Identifier)
	_, _ = buf.Write(t.getPairRecords().selfId.publicKey())

	signature := ed25519.Sign(t.getPairRecords().selfId.privateKey(), buf.Bytes())

	// this represents the device info of this host that is stored on the device on a successful pairing.
	// The only relevant field is 'accountID' as it's used earlier in the pairing process already.
	// Everything else can be random data and is not needed later in any communication.
	deviceInfo, err := opack.Encode(map[string]interface{}{
		"accountID":                   t.getPairRecords().selfId.Identifier,
		"altIRK":                      []byte{0x5e, 0xca, 0x81, 0x91, 0x92, 0x02, 0x82, 0x00, 0x11, 0x22, 0x33, 0x44, 0xbb, 0xf2, 0x4a, 0xc8},
		"btAddr":                      "FF:DD:99:66:BB:AA",
		"mac":                         []byte{0xff, 0x44, 0x88, 0x66, 0x33, 0x99},
		"model":                       "go-ios",
		"name":                        "host-name",
		"remotepairing_serial_number": "remote-serial",
	})

	deviceInfoTlv := newTlvBuffer()
	deviceInfoTlv.writeData(typeSignature, signature)
	deviceInfoTlv.writeData(typePublicKey, t.getPairRecords().selfId.publicKey())
	deviceInfoTlv.writeData(typeIdentifier, []byte(t.getPairRecords().selfId.Identifier))
	deviceInfoTlv.writeData(typeInfo, deviceInfo)

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
	x := cipher.Seal(nil, nonce, deviceInfoTlv.bytes(), nil)

	encryptedTlv := newTlvBuffer()
	encryptedTlv.writeByte(typeState, 0x05)
	encryptedTlv.writeData(typeEncryptedData, x)

	err = t.writeEvent(&pairingData{
		data:        encryptedTlv.bytes(),
		kind:        "setupManualPairing",
		sendingHost: "SL-1876",
	})
	if err != nil {
		return err
	}

	var encRes pairingData
	err = t.readEvent(&encRes)
	if err != nil {
		return err
	}

	encrData, err := tlvReader(encRes.data).readCoalesced(typeEncryptedData)
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
