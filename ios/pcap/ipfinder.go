package pcap

import (
	"fmt"
	"io"
	"time"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	log "github.com/sirupsen/logrus"
)

type NetworkInfo struct {
	Mac  string
	IPv4 string
	IPv6 string
}

type PacketInfo struct {
	Mac  string
	IPv4 string
	IPv6 string
}

func (n PacketInfo) complete() bool {
	return (n.IPv6 != "" && n.Mac != "") || (n.IPv4 != "" && n.Mac != "")
}

// FindIPByMac reads pcap packets until one is found that matches the given MAC
// and contains an IP address. This won't work if the iOS device "automatic Wifi address" privacy
// feature is enabled. The MAC needs to be static.
func FindIPByMac(device ios.DeviceEntry, capture_timeout time.Duration) (NetworkInfo, error) {
	mac, err := ios.GetWifiMac(device)
	if err != nil {
		return NetworkInfo{}, fmt.Errorf("FindIPMyMac: unable to get WiFi MAC Address: %w", err)
	}

	log.Infof("FindIPByMac: connected to pcapd. timeout %v", capture_timeout)

	pcapService, err := ios.ConnectToService(device, "com.apple.pcapd")
	if err != nil {
		return NetworkInfo{}, fmt.Errorf("FindIPByMac: failed connecting to com.apple.pcapd with err: %w", err)
	}

	endSignal := time.After(capture_timeout)

L:
	for {
		select {
		case <-endSignal:
			break L
		default:
			packet, err := readPacket(pcapService.Reader())
			if err != nil {
				return NetworkInfo{}, fmt.Errorf("FindIPByMac: error reading pcap packet: %w", err)
			}

			if packet.Mac == mac {
				return NetworkInfo{Mac: packet.Mac, IPv4: packet.IPv4, IPv6: packet.IPv6}, nil
			}
		}
	}

	return NetworkInfo{}, fmt.Errorf("failed to get any IP matching the MAC: %s", mac)
}

// FindIPByLazy reads pcap packets for a specified duration, whereafter it will
// try to find the IP address that occurs the most times, and assume that is the IP of the device.
// This is of course based on, that the device contacts mulitple IPs, and that there is some traffic.
// If the device only contains a single IP, then it would be 50/50 which IP will be returned.
// This is best effort! It's important to generate some traffic, when this function runs to get better results.
func FindIPByLazy(device ios.DeviceEntry, capture_duration time.Duration) (NetworkInfo, error) {
	pcapService, err := ios.ConnectToService(device, "com.apple.pcapd")
	if err != nil {
		return NetworkInfo{}, fmt.Errorf("FindIPByLazy: failed connecting to com.apple.pcapd with err: %w", err)
	}

	log.Infof("FindIPByLazy: connected to pcapd. waiting %v", capture_duration)

	ipv6Hits := make(map[string]int)
	ipv4Hits := make(map[string]int)
	endSignal := time.After(capture_duration)

L:
	for {
		select {
		case <-endSignal:
			break L
		default:
			packet, err := readPacket(pcapService.Reader())
			if err != nil {
				return NetworkInfo{}, fmt.Errorf("FindIPByLazy: error reading pcap packet: %w", err)
			}
			if packet.IPv4 != "" {
				ipv4Hits[packet.IPv4] += 1
			}
			if packet.IPv6 != "" {
				ipv6Hits[packet.IPv6] += 1
			}

		}
	}

	highestIPv4Hits, highestIPv4Addr := 0, ""
	for ipv4, hits := range ipv4Hits {
		if hits > highestIPv4Hits {
			highestIPv4Hits = hits
			highestIPv4Addr = ipv4
		}
	}

	highestIPv6Hits, highestIPv6Addr := 0, ""
	for ipv6, hits := range ipv6Hits {
		if hits > highestIPv6Hits {
			highestIPv6Hits = hits
			highestIPv6Addr = ipv6
		}
	}

	return NetworkInfo{IPv4: highestIPv4Addr, IPv6: highestIPv6Addr}, nil
}

var plistCodec = ios.NewPlistCodec()

func readPacket(r io.Reader) (PacketInfo, error) {
	b, err := plistCodec.Decode(r)
	if err != nil {
		return PacketInfo{}, fmt.Errorf("readPacket: failed decoding plistCodec err: %w", err)
	}
	decodedBytes, err := fromBytes(b)
	if err != nil {
		return PacketInfo{}, fmt.Errorf("readPacket: failed decoding fromBytes err: %w", err)
	}
	_, packet, err := getPacket(decodedBytes)
	if err != nil {
		return PacketInfo{}, fmt.Errorf("readPacket: failed getPacket err: %w", err)
	}
	if len(packet) > 0 {
		pInfo := parsePacket(packet)
		if pInfo.complete() {
			return pInfo, nil
		}
	}
	return PacketInfo{}, nil
}

func parsePacket(p []byte) PacketInfo {
	packet := gopacket.NewPacket(p, layers.LayerTypeEthernet, gopacket.Default)
	res := PacketInfo{}

	// Get the TCP layer from this packet
	if tcpLayer := packet.Layer(layers.LayerTypeEthernet); tcpLayer != nil {
		tcp, _ := tcpLayer.(*layers.Ethernet)
		res.Mac = tcp.SrcMAC.String()

		if ipv4Layer := packet.Layer(layers.LayerTypeIPv4); ipv4Layer != nil {
			ipv4, ok := ipv4Layer.(*layers.IPv4)
			if ok {
				res.IPv4 = ipv4.SrcIP.String()
				log.Debugf("ip4 found:%s", res.IPv4)
			}
		}
		if ipv6Layer := packet.Layer(layers.LayerTypeIPv6); ipv6Layer != nil {
			ipv6, ok := ipv6Layer.(*layers.IPv6)
			if ok {
				res.IPv6 = ipv6.SrcIP.String()
				log.Debugf("ip6 found:%s", res.IPv6)
			}
		}
	}

	return res
}
