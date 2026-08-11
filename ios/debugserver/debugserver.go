package debugserver

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path"
	"strings"
	"text/template"
	"time"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/golog"
	"github.com/danielpaulus/go-ios/ios/installationproxy"
	"howett.net/plist"
)

const logModule = "go-ios/debugserver"

const (
	serviceName    = "com.apple.debugserver"
	sslServiceName = "com.apple.debugserver.DVTSecureSocketProxy"
)

// ref: https://github.com/steeve/itool/blob/master/debugserver/debugserver.go#L14
type DebugClient struct {
	c         *ios.LockDownConnection
	gdbServer *GDBServer
}

func (c *DebugClient) Recv() (string, error) {
	return c.gdbServer.Recv()
}

func (c *DebugClient) Send(req string) error {
	return c.gdbServer.Send(req)
}

func (c *DebugClient) Request(req string) (string, error) {
	return c.gdbServer.Request(req)
}

func (c *DebugClient) Close() {
	c.c.Close()
}

func (c *DebugClient) Conn() net.Conn {
	return c.c.Conn()
}

// lldbScriptConfig holds everything needed to render the lldb command script
// and the python helper it imports. Either the launch fields or pid are set,
// never both.
type lldbScriptConfig struct {
	// launch mode
	appPath     string // local .app bundle, becomes the lldb target
	container   string // path of the installed app on the device
	stopAtEntry bool
	// attach mode
	pid int // when > 0, attach to this process instead of launching

	port int // local proxy port lldb connects to
}

// renderLLDBScripts renders the lldb command script and the python helper it
// imports. It is a pure function so the generated scripts can be unit tested.
func renderLLDBScripts(cfg lldbScriptConfig) (lldbScript string, pyScript string, err error) {
	var optionStopAtEntry string
	if cfg.stopAtEntry {
		optionStopAtEntry = STOP_AT_ENTRY
	}

	pyt, err := template.New("py").Parse(PY_FMT)
	if err != nil {
		return "", "", err
	}
	var py strings.Builder
	err = pyt.Execute(&py, struct {
		StopAtEntry string
	}{
		StopAtEntry: optionStopAtEntry,
	})
	if err != nil {
		return "", "", err
	}

	st, err := template.New("script").Parse(LLDB_FMT)
	if err != nil {
		return "", "", err
	}
	var script strings.Builder
	err = st.Execute(&script, struct {
		AppPath   string
		Container string
		Port      int
		Pid       int
		PyName    string
		PyPath    string
	}{
		AppPath:   cfg.appPath,
		Container: cfg.container,
		Port:      cfg.port,
		Pid:       cfg.pid,
		PyName:    strings.TrimSuffix(path.Base(PY_PATH), path.Ext(PY_PATH)),
		PyPath:    PY_PATH,
	})
	if err != nil {
		return "", "", err
	}
	return script.String(), py.String(), nil
}

// Write the script files to the tmp directory and start lldb
func startLLDB(cfg lldbScriptConfig) error {
	lldbScript, pyScript, err := renderLLDBScripts(cfg)
	if err != nil {
		return err
	}
	if err := os.WriteFile(PY_PATH, []byte(pyScript), 0o644); err != nil {
		return err
	}
	if err := os.WriteFile(SCRIPT_PATH, []byte(lldbScript), 0o644); err != nil {
		return err
	}
	cmd := exec.Command(LLDB_SHELL, "-s", SCRIPT_PATH)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func getBundleidFromApp(appPath string) (string, error) {
	plistPath := path.Join(appPath, "Info.plist")
	// check path
	if !fileExists(plistPath) {
		return "", errors.New("cannot find info.plist")
	}
	// read bundleId
	pcontent, err := os.ReadFile(plistPath)
	if err != nil {
		return "", err
	}
	pmap := make(map[string]interface{})
	_, err = plist.Unmarshal(pcontent, pmap)
	if err != nil {
		return "", err
	}

	bundleId, ok := pmap["CFBundleIdentifier"]
	if !ok || bundleId == nil {
		return "", errors.New("cannot find CFBundleIdentifier in Info.plist")
	}
	return bundleId.(string), nil
}

func connectToDevice(device ios.DeviceEntry) (ios.DeviceConnectionInterface, error) {
	info, err := ios.GetValuesPlist(device)
	if err != nil {
		return nil, err
	}
	version, ok := info["ProductVersion"]
	if !ok {
		golog.Error("cannot find version, default use ssl debug server", "module", logModule, "udid", device.Properties.SerialNumber)
		return ios.ConnectToService(device, sslServiceName)
	}
	if version.(string) > "14" {
		return ios.ConnectToService(device, sslServiceName)
	}
	intf, err := ios.ConnectToService(device, serviceName)
	if err != nil {
		return intf, err
	}
	return intf, err
}

// Start launches the app at appPath on the device and opens an interactive
// lldb session for it.
func Start(device ios.DeviceEntry, appPath string, stopAtEntry bool) error {
	bundleId, err := getBundleidFromApp(appPath)
	if err != nil {
		return err
	}
	conn, err := installationproxy.New(device)
	if err != nil {
		return err
	}
	appinfo, err := conn.BrowseUserApps()
	if err != nil {
		return err
	}
	var container string
	for _, ai := range appinfo {
		if ai.CFBundleIdentifier() == bundleId {
			container = ai.Path()
			break
		}
	}
	if container == "" {
		return errors.New("cannot find container of bundleid: " + bundleId)
	}

	return runSession(device, lldbScriptConfig{appPath: appPath, container: container, stopAtEntry: stopAtEntry})
}

// AttachByPid opens an interactive lldb session attached to the process with
// the given pid on the device.
func AttachByPid(device ios.DeviceEntry, pid int) error {
	if pid <= 0 {
		return fmt.Errorf("invalid pid %d, must be > 0", pid)
	}
	golog.Info("attaching lldb to process", "module", logModule, "udid", device.Properties.SerialNumber, "pid", pid)
	return runSession(device, lldbScriptConfig{pid: pid})
}

// runSession connects to the debugserver on the device, proxies it on a local
// port and runs lldb against that port until the debug session ends.
func runSession(device ios.DeviceEntry, cfg lldbScriptConfig) error {
	intf, err := connectToDevice(device)
	if err != nil {
		return err
	}
	// listen at random port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return fmt.Errorf("failed to listen on 127.0.0.1: %w", err)
	}
	defer listener.Close()
	port := listener.Addr().(*net.TCPAddr).Port
	cfg.port = port
	golog.Info("debug proxy listening", "module", logModule, "udid", device.Properties.SerialNumber, "port", port)

	// Run lldb in the background; its result ends the debug session. Start
	// returns that result (nil when the session ends cleanly), so the caller
	// decides what to do instead of this library exiting the process.
	lldbDone := make(chan error, 1)
	go func() {
		time.Sleep(time.Second)
		lldbDone <- startLLDB(cfg)
	}()

	// Proxy connections until Start returns and closes the listener.
	go func() {
		for {
			localConn, err := listener.Accept()
			if err != nil {
				if errors.Is(err, net.ErrClosed) {
					return // listener closed because the session ended
				}
				golog.Error("accept failed", "module", logModule, "udid", device.Properties.SerialNumber, "port", port, "error", err)
				continue
			}
			go func() {
				lc := ios.NewLockDownConnection(intf)
				cli := &DebugClient{
					c:         lc,
					gdbServer: NewGDBServer(lc.Conn()),
				}
				// start proxy
				go io.Copy(localConn, cli.Conn())
				io.Copy(cli.Conn(), localConn)
			}()
		}
	}()

	if err := <-lldbDone; err != nil {
		return fmt.Errorf("lldb failed: %w", err)
	}
	return nil
}
