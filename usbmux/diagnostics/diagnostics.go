package diagnostics

/*
import "github.com/danielpaulus/go-ios/usbmux"

const serviceName = "com.apple.mobile.diagnostics_relay"

type Connection struct {
	muxConn    *usbmux.MuxConnection
	plistCodec *usbmux.PlistCodec
}

func New(deviceID int, udid string, pairRecord usbmux.PairRecord) (*Connection, error) {
	startServiceResponse := usbmux.StartService(deviceID, udid, serviceName)
	var diagnosticsConn Connection
	diagnosticsConn.muxConn = usbmux.NewUsbMuxConnection()
	responseChannel := make(chan []byte)

	plistCodec := usbmux.NewPlistCodec(responseChannel)
	diagnosticsConn.plistCodec = plistCodec
	err := diagnosticsConn.muxConn.ConnectWithStartServiceResponse(deviceID, *startServiceResponse, plistCodec, pairRecord)
	if err != nil {
		return &Connection{}, err
	}

	return &diagnosticsConn, nil
}

func (diagnosticsConn *Connection) AllValues() allDiagnosticsResponse {
	allReq := diagnosticsRequest{"All"}
	diagnosticsConn.muxConn.Send(allReq)
	response := <-diagnosticsConn.plistCodec.ResponseChannel
	return diagnosticsfromBytes(response)
}

func (diagnosticsConn *Connection) Close() {
	closeReq := diagnosticsRequest{"Goodbye"}
	diagnosticsConn.muxConn.Send(closeReq)
	<-diagnosticsConn.plistCodec.ResponseChannel
	diagnosticsConn.muxConn.Close()

}
*/
