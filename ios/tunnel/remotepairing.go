package tunnel

import (
	"context"
	"crypto/sha512"
	"crypto/tls"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/danielpaulus/go-ios/ios"
	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/hkdf"
)

type RemoteTunnelService struct {
	pairRecords  PairRecordManager
	connSequence *ConnSequence
	cipher       *cipherStream
}

func (r *RemoteTunnelService) readEncrypted() (map[string]interface{}, error) {
	m, err := receiveRemotePair(r.connSequence.Conn)
	if err != nil {
		return nil, fmt.Errorf("readEncrypted: failed to read message: %w", err)
	}
	return getChildMap(m, "message")
}

func (r *RemoteTunnelService) writeEncrypted(data map[string]interface{}) error {
	return sendRemotePair(data, r.connSequence)
}

func (r *RemoteTunnelService) readResponse() (map[string]interface{}, error) {
	return receiveResponse(r.connSequence.Conn)
}

func (r *RemoteTunnelService) getCipher() *cipherStream {
	return r.cipher
}

func (r *RemoteTunnelService) writeEvent(event eventCodec) error {
	return writeEvent(event, r.connSequence)
}

func (r *RemoteTunnelService) readEvent(event eventCodec) error {
	return readEvent(event, r.connSequence)
}

func (r *RemoteTunnelService) setupCiphers(sessionKey []byte) error {
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

	r.cipher = newCipherStream(r, client, server)

	return nil
}

func (t *RemoteTunnelService) writeRequest(req map[string]interface{}) error {
	return writeRequest(req, t.connSequence)
}

func (r *RemoteTunnelService) getPairRecords() PairRecordManager {
	return r.pairRecords
}

type ConnSequence struct {
	Conn     io.ReadWriteCloser
	Sequence uint64
	mux      sync.Mutex
}

func (c *ConnSequence) GetAndIncrement() uint64 {
	c.mux.Lock()
	defer c.mux.Unlock()
	c.Sequence++
	return c.Sequence
}

var remotePairingMagic = []byte("RPPairing")

func Verify(conn *ConnSequence) error {
	data := map[string]interface{}{
		"handshake": map[string]interface{}{
			"_0": map[string]interface{}{
				"hostOptions": map[string]interface{}{
					"attemptPairVerify": true,
				},
				"wireProtocolVersion": int64(19),
			},
		},
	}
	err := writeRequest(data, conn)
	if err != nil {
		return fmt.Errorf("could not send remote pairing handshake: %w", err)
	}
	//data, err = receiveRemotePair(conn.Conn)
	//data, err = extractResponse(data)
	data, _ = receiveResponse(conn.Conn)
	print(data)
	return nil
}

func receiveResponse(conn io.Reader) (map[string]interface{}, error) {
	data, err := receiveRemotePair(conn)
	if err != nil {
		return nil, fmt.Errorf("could not receive remote pairing response: %w", err)
	}
	return extractResponse(data)
}

func extractResponse(data map[string]interface{}) (map[string]interface{}, error) {
	message, err := getChildMap(data, "message")
	if err != nil {
		return nil, fmt.Errorf("could not extract message: %w", err)
	}
	message, err = getChildMap(message, "plain")
	if err != nil {
		return nil, fmt.Errorf("could not extract message: %w", err)
	}
	message, err = getChildMap(message, "_0")
	if err != nil {
		return nil, fmt.Errorf("could not extract message: %w", err)
	}

	return message, nil
}

func writeEvent(e eventCodec, conn *ConnSequence) error {
	encoded := map[string]interface{}{
		"plain": map[string]interface{}{
			"_0": map[string]interface{}{
				"event": map[string]interface{}{
					"_0": e.Encode(),
				},
			},
		},
	}
	return sendRemotePair(encoded, conn)
}

func readEvent(e eventCodec, conn *ConnSequence) error {
	m, err := receiveRemotePair(conn.Conn)
	if err != nil {
		return fmt.Errorf("readEvent: failed to read message: %w", err)
	}
	event, err := getChildMap(m, "message", "plain", "_0", "event", "_0")
	if err != nil {
		return fmt.Errorf("readEvent: failed to get event payload: %w", err)
	}
	return e.Decode(event)
}

func writeRequest(req map[string]interface{}, conn *ConnSequence) error {
	err := sendRemotePair(map[string]interface{}{
		"plain": map[string]interface{}{
			"_0": map[string]interface{}{
				"request": map[string]interface{}{
					"_0": req,
				},
			},
		},
	}, conn)
	if err != nil {
		return fmt.Errorf("writeRequest: failed to write message: %w", err)
	}
	return nil
}

func sendRemotePair(data map[string]interface{}, conn *ConnSequence) error {
	d := map[string]interface{}{
		"message":        data,
		"originatedBy":   "host",
		"sequenceNumber": conn.GetAndIncrement(),
	}
	b, err := json.Marshal(d)
	print(string(b) + "\n")
	if err != nil {
		return fmt.Errorf("could not marshal remote pairing data: %w", err)
	}

	_, err = conn.Conn.Write(remotePairingMagic)
	l := make([]byte, 2)
	binary.BigEndian.PutUint16(l, uint16(len(b)))
	_, err2 := conn.Conn.Write(l)
	_, err3 := conn.Conn.Write(b)
	return errors.Join(err, err2, err3)
}

func receiveRemotePair(conn io.Reader) (map[string]interface{}, error) {
	magic := make([]byte, 9)
	_, err := conn.Read(magic)
	if err != nil {
		return nil, fmt.Errorf("could not read remote pairing magic: %w", err)
	}
	if string(magic) != string(remotePairingMagic) {
		return nil, fmt.Errorf("invalid remote pairing magic: %s", magic)
	}

	var length uint16
	err = binary.Read(conn, binary.BigEndian, &length)
	if err != nil {
		return nil, fmt.Errorf("could not read remote pairing length: %w", err)
	}

	data := make([]byte, length)
	_, err = conn.Read(data)
	if err != nil {
		return nil, fmt.Errorf("could not read remote pairing data: %w", err)
	}

	var decoded map[string]interface{}
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal remote pairing data: %w", err)
	}

	return decoded, nil
}

func ConnectRemotePairingTunnel(ctx context.Context, device ios.DeviceEntry, p PairRecordManager) (Tunnel, error) {
	const amod = "_apple-mobdev2._tcp"
	const r = "_remotepairing._tcp"
	const rmp = "_remotepairing-manual-pairing._tcp.local."
	var c = func(a string, port int) (ios.RsdService, error) {
		return NewTCP(a, port, p)
	}
	ios.NewTCP = c
	_, err := ios.FindDevicesForService(context.Background(), r)
	if err != nil {
		return Tunnel{}, err
	}
	time.Sleep(15 * time.Second)

	if err != nil {
		return Tunnel{}, fmt.Errorf("ManualPairAndConnectToTunnel: failed to create tunnel listener: %w", err)
	}

	return Tunnel{}, nil
}

func NewTCP(addr1 string, port1 int, p PairRecordManager) (ios.RsdService, error) {
	addr, err := net.ResolveTCPAddr("tcp6", fmt.Sprintf("[%s]:%d", addr1, port1))
	if err != nil {
		return ios.RsdService{}, fmt.Errorf("ConnectToHttp2WithAddr: failed to resolve address: %w", err)
	}

	/*
	   ctx = SSLPSKContext(ssl.PROTOCOL_TLSv1_2)
	           ctx.psk = self.encryption_key
	           ctx.set_ciphers('PSK')
	*/
	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		return ios.RsdService{}, fmt.Errorf("ConnectToHttp2WithAddr: failed to dial: %w", err)
	}

	err = conn.SetKeepAlive(true)
	if err != nil {
		return ios.RsdService{}, fmt.Errorf("ConnectToHttp2WithAddr: failed to set keepalive: %w", err)
	}
	err = conn.SetKeepAlivePeriod(1 * time.Second)
	if err != nil {
		return ios.RsdService{}, fmt.Errorf("ConnectToHttp2WithAddr: failed to set keepalive period: %w", err)
	}
	c := ConnSequence{
		Conn: conn,
		mux:  sync.Mutex{},
	}
	//Verify(&c)
	remoteTunnelService := RemoteTunnelService{
		pairRecords:  p,
		connSequence: &c,
	}
	err = ManualPair(&remoteTunnelService)
	if err != nil {
		slog.Error("failed to verify pair", "err", err)
		return ios.RsdService{}, fmt.Errorf("ConnectToHttp2WithAddr: failed to verify pair: %w", err)
	}
	tunnelInfo, err := createTunnelListener(&remoteTunnelService, "tcp")
	conf, err := createTlsConfig(tunnelInfo)
	if err != nil {
		return ios.RsdService{}, err
	}
	addr1 = strings.Replace(addr1, "%en12", "%en0", -1)
	cs, err := net.Dial("tcp", fmt.Sprintf("[%s]:%d", addr1, tunnelInfo.TunnelPort))
	tlsconn, err := tls.Dial("tcp", fmt.Sprintf("[%s]:%d", addr1, tunnelInfo.TunnelPort), conf)
	tun, err := connectToTunnelLockdown(context.Background(), ios.DeviceEntry{Properties: ios.DeviceProperties{SerialNumber: "d"}}, tlsconn)

	slog.Info("sdf", tun, cs)
	return ios.RsdService{}, nil
}
