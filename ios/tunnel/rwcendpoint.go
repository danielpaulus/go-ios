package tunnel

// I took this code from https://github.com/xjasonlyu/tun2socks which is an amazing project
// but GPL licensed.
// Technically this would require me to have go-ios be GPL licensed too I think.
// I think though the code is pretty simple and not very original to what tun2socks is about.
// So I hope it's okay to use this code without changing my license and that I can get away with
// attribution. If not let me know, then I can uhm.. rewrite it.
import (
	"context"
	"errors"
	"io"
	"sync"

	"gvisor.dev/gvisor/pkg/buffer"
	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/header"
	"gvisor.dev/gvisor/pkg/tcpip/link/channel"

	"gvisor.dev/gvisor/pkg/tcpip/stack"
)

const (
	// Queue length for outbound packet, arriving for read. Overflow
	// causes packet drops.
	defaultOutQueueLen = 1 << 10
)

// Endpoint implements the interface of stack.LinkEndpoint from io.ReadWriter.
type Endpoint struct {
	*channel.Endpoint

	// rw is the io.ReadWriter for reading and writing packets.
	rw io.ReadWriter

	// mtu (maximum transmission unit) is the maximum size of a packet.
	mtu uint32

	// offset can be useful when perform TUN device I/O with TUN_PI enabled.
	offset int

	// once is used to perform the init action once when attaching.
	once sync.Once

	// wg keeps track of running goroutines.
	wg sync.WaitGroup
}

// RWCEndpointNew creates a new io.ReadWriter based endpoint. It is used to
// connect the virtual TUN device with the lockdown connection to the device.
// The lockdown connection expects IP packets as it is a proxy to some virtual
// network interface on the device. So in essence, the virtual gVisor TUN device will
// receive normal tcp ip connections and then encode them into IP packets sent to
// lockdown using a RWCEndpoint.
func RWCEndpointNew(rw io.ReadWriter, mtu uint32, offset int) (*Endpoint, error) {
	if mtu == 0 {
		return nil, errors.New("MTU size is zero")
	}

	if rw == nil {
		return nil, errors.New("RW interface is nil")
	}

	if offset < 0 {
		return nil, errors.New("offset must be non-negative")
	}

	return &Endpoint{
		Endpoint: channel.New(defaultOutQueueLen, mtu, ""),
		rw:       rw,
		mtu:      mtu,
		offset:   offset,
	}, nil
}

// Attach launches the goroutine that reads packets from io.Reader and
// dispatches them via the provided dispatcher.
func (e *Endpoint) Attach(dispatcher stack.NetworkDispatcher) {
	e.Endpoint.Attach(dispatcher)
	e.once.Do(func() {
		ctx, cancel := context.WithCancel(context.Background())
		e.wg.Add(2)
		go func() {
			e.outboundLoop(ctx)
			e.wg.Done()
		}()
		go func() {
			e.dispatchLoop(cancel)
			e.wg.Done()
		}()
	})
}

func (e *Endpoint) Wait() {
	e.wg.Wait()
}

// dispatchLoop dispatches packets to upper layer.
func (e *Endpoint) dispatchLoop(cancel context.CancelFunc) {
	// Call cancel() to ensure (*Endpoint).outboundLoop(context.Context) exits
	// gracefully after (*Endpoint).dispatchLoop(context.CancelFunc) returns.
	defer cancel()

	offset, mtu := e.offset, int(e.mtu)

	for {
		data := make([]byte, offset+mtu)

		n, err := e.rw.Read(data)
		if err != nil {
			break
		}

		if n == 0 || n > mtu {
			continue
		}

		if !e.IsAttached() {
			continue /* unattached, drop packet */
		}

		pkt := stack.NewPacketBuffer(stack.PacketBufferOptions{
			Payload: buffer.MakeWithData(data[offset : offset+n]),
		})

		switch header.IPVersion(data[offset:]) {
		case header.IPv4Version:
			e.InjectInbound(header.IPv4ProtocolNumber, pkt)
		case header.IPv6Version:
			e.InjectInbound(header.IPv6ProtocolNumber, pkt)
		}
		pkt.DecRef()
	}
}

// outboundLoop reads outbound packets from channel, and then it calls
// writePacket to send those packets back to lower layer.
func (e *Endpoint) outboundLoop(ctx context.Context) {
	for {
		pkt := e.ReadContext(ctx)
		if pkt == nil {
			break
		}
		e.writePacket(pkt)
	}
}

// writePacket writes outbound packets to the io.Writer.
func (e *Endpoint) writePacket(pkt *stack.PacketBuffer) tcpip.Error {
	defer pkt.DecRef()

	buf := pkt.ToBuffer()
	defer buf.Release()
	if e.offset != 0 {
		v := buffer.NewViewWithData(make([]byte, e.offset))
		_ = buf.Prepend(v)
	}

	if _, err := e.rw.Write(buf.Flatten()); err != nil {
		return &tcpip.ErrInvalidEndpointState{}
	}
	return nil
}
