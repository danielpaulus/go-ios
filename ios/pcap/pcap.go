package pcap

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/lunixbochs/struc"
	log "github.com/sirupsen/logrus"
	"howett.net/plist"
)

var (
	// IOSPacketHeader default is -1
	Pid              = int32(-2)
	ProcName         string
	PacketHeaderSize = uint32(95)
)

// IOSPacketHeader :)
// ref: https://github.com/gofmt/iOSSniffer/blob/master/pkg/sniffer/sniffer.go#L44
type IOSPacketHeader struct {
	HdrSize        uint32 `struc:"uint32,big"`
	Version        uint8  `struc:"uint8,big"`
	PacketSize     uint32 `struc:"uint32,big"`
	Type           uint8  `struc:"uint8,big"`
	Unit           uint16 `struc:"uint16,big"`
	IO             uint8  `struc:"uint8,big"`
	ProtocolFamily uint32 `struc:"uint32,big"`
	FramePreLength uint32 `struc:"uint32,big"`
	FramePstLength uint32 `struc:"uint32,big"`
	IFName         string `struc:"[16]byte"`
	Pid            int32  `struc:"int32,little"`
	ProcName       string `struc:"[17]byte"`
	Unknown        uint32 `struc:"uint32,little"`
	Pid2           int32  `struc:"int32,little"`
	ProcName2      string `struc:"[17]byte"`
	TsSec          int    `struc:"int32,big"` /* timestamp seconds */
	TsUsec         int    `struc:"int32,big"` /* timestamp microseconds */
}

func (iph *IOSPacketHeader) ToString() string {
	trim := func(src string) string {
		return strings.ReplaceAll(src, "\x00", "")
	}
	iph.IFName = trim(iph.IFName)
	iph.ProcName = trim(iph.ProcName)
	iph.ProcName2 = trim(iph.ProcName2)
	return fmt.Sprintf("%v", *iph)
}

func Start(device ios.DeviceEntry) error {
	intf, err := ios.ConnectToService(device, "com.apple.pcapd")
	if err != nil {
		return err
	}
	plistCodec := ios.NewPlistCodec()
	fname := fmt.Sprintf("dump-%d.pcap", time.Now().Unix())
	if Pid > 0 {
		fname = fmt.Sprintf("dump-%d-%d.pcap", Pid, time.Now().Unix())
	} else if ProcName != "" {
		fname = fmt.Sprintf("dump-%s-%d.pcap", ProcName, time.Now().Unix())
	}
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
		iph, packet, err := getPacket(decodedBytes)
		if err != nil {
			return err
		}
		if len(packet) > 0 {
			err = writePacket(f, iph, packet)
			if err != nil {
				return err
			}
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
	f, err := os.OpenFile(name, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0o755)
	if err != nil {
		return nil, err
	}
	// Write `pcap_hdr_s` with little endin to file.
	f.Write([]byte{
		0xd4, 0xc3, 0xb2, 0xa1, 0x02, 0x00, 0x04, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0xff, 0xff, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00,
	})
	return f, nil
}

func writePacket(f *os.File, iph IOSPacketHeader, packet []byte) error {
	phs := &PcaprecHdrS{
		iph.TsSec,
		iph.TsUsec,
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

func getPacket(buf []byte) (iph IOSPacketHeader, packet []byte, err error) {
	iph = IOSPacketHeader{}
	preader := bytes.NewReader(buf)
	struc.Unpack(preader, &iph)

	// support ios 15 beta4
	if iph.HdrSize > PacketHeaderSize {
		buf := make([]byte, iph.HdrSize-PacketHeaderSize)
		_, err = io.ReadFull(preader, buf)
		if err != nil {
			return iph, []byte{}, err
		}
	}

	// Only return specific packet
	if Pid > 0 {
		if iph.Pid != Pid && iph.Pid2 != Pid {
			return iph, []byte{}, nil
		}
	}

	if ProcName != "" {
		if !strings.HasPrefix(iph.ProcName, ProcName) && !strings.HasPrefix(iph.ProcName2, ProcName) {
			return iph, []byte{}, nil
		}
	}

	// log.Info("IOSPacketHeader: ", iph.ToString())
	packet, err = io.ReadAll(preader)
	if err != nil {
		return iph, packet, err
	}
	if iph.FramePreLength == 0 {
		ext := []byte{0xbe, 0xfe, 0xbe, 0xfe, 0xbe, 0xfe, 0xbe, 0xfe, 0xbe, 0xfe, 0xbe, 0xfe, 0x08, 0x00}
		return iph, append(ext, packet...), nil
	}
	return iph, packet, nil
}
