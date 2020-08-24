package testmanagerd

type XCTestManager_IDEInterface interface{}
type XCTestManager_DaemonConnectionInterface interface{}

type dtxproxy struct {
	ideInterface     XCTestManager_IDEInterface
	daemonConnection XCTestManager_DaemonConnectionInterface
}

func newDtxProxy(conn DtxConnection) dtxproxy {
	conn.requestChannelWithCodeAndIdentifier(1, "")
	return dtxproxy{}
}
