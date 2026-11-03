package api

import (
	"net/http"
	"time"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/tunnel"
	"github.com/gin-gonic/gin"
)

// tunnelRefreshTimeout mirrors the CLI's refresh wait.
const tunnelRefreshTimeout = 30 * time.Second

// registerTunnelRoutes registers tunnel-agent endpoints (CLI: ios tunnel ...).
// These are NOT device-scoped — they query the running tunnel agent by udid
// string over its info API (host/port from ios.HttpApiHost/HttpApiPort) — so they
// live at the /api/v1 level rather than under /device/:udid.
//
// `ios tunnel start` is intentionally not exposed: it starts a long-running
// privileged daemon (sudo / CAP_NET_ADMIN / admin shell), which is a host
// process-lifecycle concern, not a REST call.
func registerTunnelRoutes(router *gin.RouterGroup) {
	router.GET("/tunnels", ListTunnels)
	router.DELETE("/tunnels/:udid", StopTunnel)
	router.POST("/tunnels/:udid/refresh", RefreshTunnel)
	router.POST("/tunnel-agent/shutdown", ShutdownTunnelAgent)
}

// ListTunnels lists running tunnels (CLI: ios tunnel ls).
// @Summary List running tunnels
// @Produce json
// @Success 200 {array} tunnel.Tunnel
// @Router /tunnels [get]
func ListTunnels(c *gin.Context) {
	tunnels, err := tunnel.ListRunningTunnels(ios.HttpApiHost(), ios.HttpApiPort())
	if err != nil {
		RespondError(c, http.StatusBadGateway, err)
		return
	}
	c.JSON(http.StatusOK, tunnels)
}

// StopTunnel stops the tunnel for a device (CLI: ios tunnel stop --udid).
// @Summary Stop a device tunnel
// @Param udid path string true "Device UDID"
// @Success 200 {object} map[string]string
// @Router /tunnels/{udid} [delete]
func StopTunnel(c *gin.Context) {
	udid := c.Param("udid")
	if err := tunnel.StopTunnelForDevice(udid, ios.HttpApiHost(), ios.HttpApiPort()); err != nil {
		RespondError(c, http.StatusBadGateway, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"udid": udid, "status": "stopped"})
}

// RefreshTunnel restarts the tunnel for a device and waits for it (CLI: ios tunnel refresh).
// @Summary Refresh a device tunnel
// @Produce json
// @Param udid path string true "Device UDID"
// @Success 200 {object} tunnel.Tunnel
// @Router /tunnels/{udid}/refresh [post]
func RefreshTunnel(c *gin.Context) {
	udid := c.Param("udid")
	tun, err := tunnel.RefreshTunnelForDevice(udid, ios.HttpApiHost(), ios.HttpApiPort(), tunnelRefreshTimeout)
	if err != nil {
		RespondError(c, http.StatusBadGateway, err)
		return
	}
	c.JSON(http.StatusOK, tun)
}

// ShutdownTunnelAgent stops the tunnel agent (CLI: ios tunnel stopagent).
// @Summary Shut down the tunnel agent
// @Success 200 {object} map[string]string
// @Router /tunnel-agent/shutdown [post]
func ShutdownTunnelAgent(c *gin.Context) {
	if err := tunnel.CloseAgent(); err != nil {
		RespondError(c, http.StatusBadGateway, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "agent shutdown requested"})
}
