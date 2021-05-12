package pcap

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/lunixbochs/struc"
	log "github.com/sirupsen/logrus"
	"howett.net/plist"
)

// PcapPlistHeader :)
// ref: https://github.com/gofmt/iOSSniffer/blob/master/pkg/sniffer/sniffer.go#L44`~
type PcapPlistHeader struct {
	HdrSize        uint32  `struc:"uint32,big"`
	Version            uint8   `struc:"uint8,big"`
	PacketSize     uint32  `struc:"uint32,big"`
	Type           uint8   `struc:"uint8,big"`
	Unit           uint16  `struc:"uint16,big"`
	IO             uint8   `struc:"uint8,big"`
	ProtocolFamily uint32  `struc:"uint32,big"`
	FramePreLength uint32  `struc:"uint32,big"`
	FramePstLength uint32  `struc:"uint32,big"`
	IFName         string  `struc:"[16]byte"`
	Pid            int32  `struc:"int32,little"`
	ProcName       string  `struc:"[17]byte"`
	Unknown        uint32  `struc:"uint32,big"`
	Pid2           int32  `struc:"int32,little"`
	ProcName2      string  `struc:"[17]byte"`
	Unknown2       [8]byte `struc:"[8]byte"`
}

func (pph *PcapPlistHeader) ToString() string {
	trim := func(src string) string {
		return strings.ReplaceAll(src, "\x00", "")
	}
	pph.IFName = trim(pph.IFName)
	pph.ProcName = trim(pph.ProcName)
	pph.ProcName2 = trim(pph.ProcName2)
	return fmt.Sprintf("%v", *pph)
}

func Start(device ios.DeviceEntry) error {
	intf, err := ios.ConnectToService(device, "com.apple.pcapd")
	if err != nil {
		return err
	}
	plistCodec := ios.NewPlistCodec()
	fname := fmt.Sprintf("dump-%d.pcap", time.Now().Unix())
	f, err := createPcap(fname)
	if err != nil {
		return err
	}
	defer f.Close()
	log.Info("Create pcap file: ", fname)
	for {
		b, err := plistCodec.Decode(intf.Reader())
		if err != nil {
			return err
		}
		decodedBytes, err := fromBytes(b)
		if err != nil {
			return err
		}
		packet, err := getPacket(decodedBytes)
		if err != nil {
			return err
		}

		err = writePacket(f, packet)
		if err != nil {
			return err
		}
	}
}

func fromBytes(data []byte) ([]byte, error) {
	var result []byte
	_, err := plist.Unmarshal(data, &result)
	return result, err
}

// struct pcap_hdr_s {
//         guint32 magic_number;   /* magic number */
//         guint16 version_major;  /* major version number */
//         guint16 version_minor;  /* minor version number */
//         gint32  thiszone;       /* GMT to local correction */
//         guint32 sigfigs;        /* accuracy of timestamps */
//         guint32 snaplen;        /* max length of captured packets, in octets */
//         guint32 network;        /* data link type */
// } pcap_hdr_t;

// typedef struct pcaprec_hdr_s {
//         guint32 ts_sec;         /* timestamp seconds */
//         guint32 ts_usec;        /* timestamp microseconds */
//         guint32 incl_len;       /* number of octets of packet saved in file */
//         guint32 orig_len;       /* actual length of packet */
// } pcaprec_hdr_t;

// ref: https://www.wireshark.org/~martinm/mac_pcap_sample_code.c
type PcaprecHdrS struct {
	TsSec   int `struc:"uint32,little"` /* timestamp seconds */
	TsUsec  int `struc:"uint32,little"` /* timestamp microseconds */
	InclLen int `struc:"uint32,little"` /* number of octets of packet saved in file */
	OrigLen int `struc:"uint32,little"` /* actual length of packet */
}

func createPcap(name string) (*os.File, error) {
	f, err := os.OpenFile(name, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0755)
	if err != nil {
		return nil, err
	}
	// Write `pcap_hdr_s` with little endin to file.
	f.Write([]byte{
		0xd4, 0xc3, 0xb2, 0xa1, 0x02, 0x00, 0x04, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0xff, 0xff, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00})
	return f, nil
}

func writePacket(f *os.File, packet []byte) error {
	now := time.Now()
	phs := &PcaprecHdrS{
		int(now.Unix()),
		int(now.UnixNano() / 1e6),
		len(packet),
		len(packet),
	}
	var buf bytes.Buffer
	err := struc.Pack(&buf, phs)
	if err != nil {
		return err
	}
	f.Write(buf.Bytes())
	f.Write(packet)
	return nil
}

func getPacket(buf []byte) ([]byte, error) {
	pph := PcapPlistHeader{}
	preader := bytes.NewReader(buf)
	struc.Unpack(preader, &pph)
	log.Info("PcapPlistHeader: ", pph.ToString())
	packet, err := ioutil.ReadAll(preader)
	if err != nil {
		return packet, err
	}
	if pph.FramePreLength == 0 {
		ext := []byte{0xbe, 0xfe, 0xbe, 0xfe, 0xbe, 0xfe, 0xbe, 0xfe, 0xbe, 0xfe, 0xbe, 0xfe, 0x08, 0x00}
		return append(ext, packet...), nil
	}
	return packet, nil
}
