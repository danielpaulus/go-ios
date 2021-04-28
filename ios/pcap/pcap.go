package pcap

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/lunixbochs/struc"
	log "github.com/sirupsen/logrus"
	"howett.net/plist"
)

// PcapPlistHeader :)
// ref: https://github.com/iOSForensics/pymobiledevice/blob/master/pymobiledevice/pcapd.py
type PcapPlistHeader struct {
	HdrSize        int `struc:"uint32,big"`
	Xxx            int `struc:"uint8,big"`
	PacketSize     int `struc:"uint32,big"`
	Flag1          int `struc:"uint32,big"`
	Flag2          int `struc:"uint32,big"`
	OffsetToIPData int `struc:"uint32,big"`
	Zero           int `struc:"uint32,big"`
}

func Start(device ios.DeviceEntry) error {
	intf, err := ios.ConnectToService(device, "com.apple.pcapd")
	if err != nil {
		return err
	}
	plistCodec := ios.NewPlistCodec()
	fname := fmt.Sprintf("dump-%d.pcap",time.Now().Unix())
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
		int(now.UTC().Unix()),
		int(now.UTC().UnixNano() * 1000),
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

	interfacetype := make([]byte, pph.HdrSize-25)
	cnt, err := preader.Read(interfacetype)
	if err != nil {
		return []byte{}, err
	}
	if cnt != pph.HdrSize-25 {
		return []byte{}, errors.New("invalid HdrSize")
	}
	log.Debug(interfacetype)
	packet, err := ioutil.ReadAll(preader)
	if err != nil {
		return packet, err
	}
	if pph.OffsetToIPData == 0 {
		ext := []byte{0xbe, 0xfe, 0xbe, 0xfe, 0xbe, 0xfe, 0xbe, 0xfe, 0xbe, 0xfe, 0xbe, 0xfe, 0x08, 0x00}
		return append(ext, packet...), nil
	}
	return packet, nil
}
