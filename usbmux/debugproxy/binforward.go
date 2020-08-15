package debugproxy

import "io"

type BinDumpCodec struct {
	received chan []byte
}

func NewBinDumpCodec(channel chan []byte) *BinDumpCodec {
	return &BinDumpCodec{channel}
}

func (b BinDumpCodec) Encode(msg interface{}) ([]byte, error) {
	return msg.([]byte), nil
}

func (b *BinDumpCodec) Decode(r io.Reader) error {
	buffer := make([]byte, 1024)
	n, err := r.Read(buffer)
	if err != nil {
		return err
	}
	b.received <- buffer[0:n]
	return nil
}
func readOnDeviceConnectionAndForwardToUnixDomainConnectionGeneric(p *ProxyConnection) {
	/*for {
		msg := <-p.deviceChannel

		if msg == nil {
			log.Info("device disconnected")
			p.connListeningOnUnixSocket.Close()
			return
		}

		//log.Info(hex.Dump(msg))
		p.connListeningOnUnixSocket.Send(msg)
	}*/
}

func readOnUnixDomainSocketAndForwardToDeviceGeneric(p *ProxyConnection) {
	/*
		for {
			msg := <-p.unixSocketChannel

			if msg == nil {
				log.Info("service on host disconnected")
				p.connectionToDevice.Close()
				return
			}
			//log.Info(hex.Dump(msg))
			p.connectionToDevice.Send(msg)

		}
	*/
}
