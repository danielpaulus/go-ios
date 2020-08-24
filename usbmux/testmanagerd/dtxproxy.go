package testmanagerd

type XCTestManager_IDEInterface interface{}
type XCTestManager_DaemonConnectionInterface interface{}

type dtxproxy struct {
	ideInterface     XCTestManager_IDEInterface
	daemonConnection XCTestManager_DaemonConnectionInterface
}

func newDtxProxy(conn dtxConnection) dtxproxy {
	conn.requestChannelWithCodeAndIdentifier(1, "")
	return dtxproxy{}
}

type dtxConnection struct {
}

func (d dtxConnection) requestChannelWithCodeAndIdentifier(code int, identifier string) dtxChannel {
	return dtxChannel{}
}

type dtxChannel struct {
	channelCode int
	identifier  int
}
