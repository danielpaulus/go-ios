package tunnel

//disabled for now
/*
import (
	"context"
	"sync/atomic"
	"testing"

	"github.com/Masterminds/semver"
	"github.com/danielpaulus/go-ios/ios"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSuccessStartForMultipleConnectedDevices(t *testing.T) {
	tm, ts, dl := setupTunnelManager()

	d1 := ios.DeviceEntry{
		Properties: ios.DeviceProperties{
			SerialNumber: "serial1",
		},
	}
	d2 := ios.DeviceEntry{
		Properties: ios.DeviceProperties{
			SerialNumber: "serial2",
		},
	}

	dl.On("ListDevices").Return(ios.DeviceList{DeviceList: []ios.DeviceEntry{d1, d2}}, nil)

	ts.On("StartTunnel", mock.Anything, d1, mock.Anything).Return(Tunnel{
		Address: "addr1",
		RsdPort: 1,
		Udid:    "serial1",
	}, nil)
	ts.On("StartTunnel", mock.Anything, d2, mock.Anything).Return(Tunnel{
		Address: "addr2",
		RsdPort: 2,
		Udid:    "serial2",
	}, nil)

	err := tm.UpdateTunnels(context.Background())
	assert.NoError(t, err)

	tunnels, err := tm.ListTunnels()

	assert.Contains(t, tunnels, Tunnel{
		Address: "addr1",
		RsdPort: 1,
		Udid:    "serial1",
	})
	assert.Contains(t, tunnels, Tunnel{
		Address: "addr2",
		RsdPort: 2,
		Udid:    "serial2",
	})
}

func TestCloseTunnelsOnDisconnect(t *testing.T) {
	tm, ts, dl := setupTunnelManager()

	d1 := ios.DeviceEntry{
		Properties: ios.DeviceProperties{
			SerialNumber: "serial",
		},
	}
	var closerCalls atomic.Uint64
	closer := func() error {
		closerCalls.Add(1)
		return nil
	}

	dl.On("ListDevices").
		Return(ios.DeviceList{DeviceList: []ios.DeviceEntry{d1}}, nil).
		Once()
	ts.On("StartTunnel", mock.Anything, d1, mock.Anything).Return(Tunnel{
		Address: "addr1",
		RsdPort: 1,
		closer:  closer,
	}, nil)

	err := tm.UpdateTunnels(context.Background())
	assert.NoError(t, err)

	tunnels, _ := tm.ListTunnels()
	assert.Len(t, tunnels, 1)

	dl.On("ListDevices").
		Return(ios.DeviceList{}, nil).
		Once()

	err = tm.UpdateTunnels(context.Background())
	assert.NoError(t, err)
	tunnels, _ = tm.ListTunnels()
	assert.Len(t, tunnels, 0)
	assert.GreaterOrEqual(t, closerCalls.Load(), uint64(1))
}

func TestBridgeIsOnlyStarteOnce(t *testing.T) {
	tm, ts, dl := setupTunnelManager()

	d1 := ios.DeviceEntry{
		Properties: ios.DeviceProperties{
			SerialNumber: "serial",
		},
	}

	closer := func() error { return nil }
	dl.On("ListDevices").
		Return(ios.DeviceList{DeviceList: []ios.DeviceEntry{d1}}, nil)
	ts.On("StartTunnel", mock.Anything, d1, mock.Anything).Return(Tunnel{
		Address: "addr1",
		RsdPort: 1,
		closer:  closer,
	}, nil)

	err := tm.UpdateTunnels(context.Background())
	assert.NoError(t, err)
	err = tm.UpdateTunnels(context.Background())
	assert.NoError(t, err)

	ts.AssertNumberOfCalls(t, "StartTunnel", 1)
}

func setupTunnelManager() (*TunnelManager, *tunnelStarterMock, *deviceListerMock) {
	ts := new(tunnelStarterMock)
	dl := new(deviceListerMock)

	return &TunnelManager{
		ts:      ts,
		dl:      dl,
		tunnels: map[string]Tunnel{},
	}, ts, dl
}

type tunnelStarterMock struct {
	mock.Mock
}

func (t *tunnelStarterMock) StartTunnel(ctx context.Context, device ios.DeviceEntry, p PairRecordManager, version *semver.Version) (Tunnel, error) {
	args := t.Mock.Called(ctx, device, p)
	return args.Get(0).(Tunnel), args.Error(1)
}

type deviceListerMock struct {
	mock.Mock
}

func (d *deviceListerMock) ListDevices() (ios.DeviceList, error) {
	args := d.Called()
	return args.Get(0).(ios.DeviceList), args.Error(1)
}
*/
