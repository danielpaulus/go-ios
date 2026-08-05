package debugserver

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func lastCommand(script string) string {
	lines := strings.Split(strings.TrimSpace(script), "\n")
	return lines[len(lines)-1]
}

func TestRenderLLDBScriptsLaunch(t *testing.T) {
	lldbScript, pyScript, err := renderLLDBScripts(lldbScriptConfig{
		appPath:   "/local/path/Wda.app",
		container: "/private/var/containers/Bundle/Application/GUID/Wda.app",
		port:      54321,
	})
	require.NoError(t, err)

	assert.Contains(t, lldbScript, "platform select remote-ios\n")
	assert.Contains(t, lldbScript, `target create "/local/path/Wda.app"`)
	assert.Contains(t, lldbScript, `script device_app="/private/var/containers/Bundle/Application/GUID/Wda.app"`)
	assert.Contains(t, lldbScript, `script connect_url="connect://127.0.0.1:54321"`)
	assert.Contains(t, lldbScript, `command script import "`+PY_PATH+`"`)
	assert.Contains(t, lldbScript, "\nconnect\n")
	assert.Equal(t, "run", lastCommand(lldbScript))
	assert.NotContains(t, lldbScript, "process attach")

	assert.Contains(t, pyScript, "def connect_command(")
	assert.Contains(t, pyScript, "def run_command(")
	// stop-at-entry is off by default
	assert.NotContains(t, pyScript, STOP_AT_ENTRY)
}

func TestRenderLLDBScriptsLaunchStopAtEntry(t *testing.T) {
	lldbScript, pyScript, err := renderLLDBScripts(lldbScriptConfig{
		appPath:     "/local/path/Wda.app",
		container:   "/private/var/containers/Bundle/Application/GUID/Wda.app",
		port:        54321,
		stopAtEntry: true,
	})
	require.NoError(t, err)

	assert.Equal(t, "run", lastCommand(lldbScript))
	assert.Contains(t, pyScript, STOP_AT_ENTRY)
}

func TestRenderLLDBScriptsAttach(t *testing.T) {
	lldbScript, pyScript, err := renderLLDBScripts(lldbScriptConfig{
		pid:  1337,
		port: 54321,
	})
	require.NoError(t, err)

	assert.Contains(t, lldbScript, "platform select remote-ios\n")
	assert.Contains(t, lldbScript, `script connect_url="connect://127.0.0.1:54321"`)
	assert.Contains(t, lldbScript, "\nconnect\n")
	assert.Equal(t, "process attach --pid 1337", lastCommand(lldbScript))
	// no local binary in attach mode: no target is created and the app is not launched
	assert.NotContains(t, lldbScript, "target create")
	assert.NotContains(t, lldbScript, "\nrun\n")

	// the python helper creates an empty target for attach mode
	assert.Contains(t, pyScript, "debugger.CreateTarget('')")
	assert.NotContains(t, pyScript, STOP_AT_ENTRY)
}
