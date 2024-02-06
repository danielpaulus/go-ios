package main

import (
	"context"
	"testing"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/tunnel"
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

	ts.On("StartTunnel", mock.Anything, d1, mock.Anything).Return(tunnel.Tunnel{
		Address: "addr1",
		RsdPort: 1,
	}, nil)
	ts.On("StartTunnel", mock.Anything, d2, mock.Anything).Return(tunnel.Tunnel{
		Address: "addr2",
		RsdPort: 2,
	}, nil)

	err := tm.UpdateTunnels()
	assert.NoError(t, err)

	tunnels, err := tm.ListTunnels()

	assert.Contains(t, tunnels, tunnelInfo{
		Address: "addr1",
		RsdPort: 1,
		Udid:    "serial1",
	})
	assert.Contains(t, tunnels, tunnelInfo{
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

	closer := new(mockCloser)
	closer.On("Close").Return(nil)

	dl.On("ListDevices").
		Return(ios.DeviceList{DeviceList: []ios.DeviceEntry{d1}}, nil).
		Once()
	ts.On("StartTunnel", mock.Anything, d1, mock.Anything).Return(tunnel.Tunnel{
		Address: "addr1",
		RsdPort: 1,
		Closer:  closer,
	}, nil)

	err := tm.UpdateTunnels()
	assert.NoError(t, err)

	tunnels, _ := tm.ListTunnels()
	assert.Len(t, tunnels, 1)

	dl.On("ListDevices").
		Return(ios.DeviceList{}, nil).
		Once()

	err = tm.UpdateTunnels()
	assert.NoError(t, err)
	tunnels, _ = tm.ListTunnels()
	assert.Len(t, tunnels, 0)
	closer.AssertCalled(t, "Close")
}

func TestBridgeIsOnlyStarteOnce(t *testing.T) {
	tm, ts, dl := setupTunnelManager()

	d1 := ios.DeviceEntry{
		Properties: ios.DeviceProperties{
			SerialNumber: "serial",
		},
	}

	closer := new(mockCloser)
	closer.On("Close").Return(nil)

	dl.On("ListDevices").
		Return(ios.DeviceList{DeviceList: []ios.DeviceEntry{d1}}, nil)
	ts.On("StartTunnel", mock.Anything, d1, mock.Anything).Return(tunnel.Tunnel{
		Address: "addr1",
		RsdPort: 1,
		Closer:  closer,
	}, nil)

	err := tm.UpdateTunnels()
	assert.NoError(t, err)
	err = tm.UpdateTunnels()
	assert.NoError(t, err)

	ts.AssertNumberOfCalls(t, "StartTunnel", 1)
}

func setupTunnelManager() (*tunnelManager, *tunnelStarterMock, *deviceListerMock) {
	ts := new(tunnelStarterMock)
	dl := new(deviceListerMock)

	return &tunnelManager{
		ts:      ts,
		dl:      dl,
		tunnels: map[string]tunnelInfo{},
	}, ts, dl
}

type tunnelStarterMock struct {
	mock.Mock
}

func (t *tunnelStarterMock) StartTunnel(ctx context.Context, device ios.DeviceEntry, p tunnel.PairRecordManager) (tunnel.Tunnel, error) {
	args := t.Mock.Called(ctx, device, p)
	return args.Get(0).(tunnel.Tunnel), args.Error(1)
}

type deviceListerMock struct {
	mock.Mock
}

func (d *deviceListerMock) ListDevices() (ios.DeviceList, error) {
	args := d.Called()
	return args.Get(0).(ios.DeviceList), args.Error(1)
}

type mockCloser struct {
	mock.Mock
}

func (m *mockCloser) Close() error {
	return m.Called().Error(0)
}
