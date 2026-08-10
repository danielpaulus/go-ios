package api

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/forward"
	"github.com/danielpaulus/go-ios/ios/testmanagerd"
	"github.com/gin-gonic/gin"
)

const (
	// defaults for the runwda convenience endpoint
	defaultWDABundleID     = "com.deviceboxhq.goios.WebDriverAgentRunner.xctrunner"
	defaultWDAXctestConfig = "WebDriverAgentRunner.xctest"
)

// registerJobRoutes registers the async-job endpoints for long-running device
// operations (test runs, port forwards). Each job runs in the background with an
// isolated, streamable log; clients poll status or stream logs and stop when done.
// Routes live under /device/:udid.
func registerJobRoutes(device *gin.RouterGroup) {
	device.POST("/jobs/runtest", StartRunTest)
	device.POST("/jobs/runwda", StartRunWda)
	device.POST("/jobs/forward", StartForward)
	device.GET("/jobs", ListJobs)
	device.GET("/jobs/:id", GetJob)
	device.GET("/jobs/:id/logs", streamingMiddleWare, StreamJobLogs)
	device.DELETE("/jobs/:id", StopJob)
}

type runTestRequest struct {
	BundleId           string         `json:"bundleId"`
	TestRunnerBundleId string         `json:"testRunnerBundleId"`
	XctestConfig       string         `json:"xctestConfig"`
	Env                map[string]any `json:"env"`
	Args               []string       `json:"args"`
	TestsToRun         []string       `json:"testsToRun"`
	TestsToSkip        []string       `json:"testsToSkip"`
	XcTest             bool           `json:"xctest"`
}

// startTestJob runs a testmanagerd test in the background, routing its output to
// the job's isolated log sink, and returns the created job.
func startTestJob(device ios.DeviceEntry, kind string, cfg testmanagerd.TestConfig) *Job {
	ctx, cancel := context.WithCancel(context.Background())
	j := jobs.create(kind, device.Properties.SerialNumber, func() error { cancel(); return nil })
	cfg.Device = device
	cfg.Listener = testmanagerd.NewTestListener(j.log, j.log, os.TempDir())
	go func() {
		suites, err := testmanagerd.RunTestWithConfig(ctx, cfg)
		j.finish(suites, err)
	}()
	return j
}

// StartRunTest starts an XCUITest/unit-test run (CLI: ios runtest).
// @Summary Start a test run (async job)
// @Accept json
// @Produce json
// @Param udid path string true "Device UDID"
// @Param body body runTestRequest true "test configuration"
// @Success 202 {object} jobView
// @Failure 400 {object} map[string]string
// @Router /device/{udid}/jobs/runtest [post]
func StartRunTest(c *gin.Context) {
	device := c.MustGet(IOS_KEY).(ios.DeviceEntry)
	var req runTestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, err)
		return
	}
	if req.TestRunnerBundleId == "" {
		req.TestRunnerBundleId = req.BundleId
	}
	if req.TestRunnerBundleId == "" {
		RespondError(c, http.StatusBadRequest, errMissingBundleID)
		return
	}
	j := startTestJob(device, "runtest", testmanagerd.TestConfig{
		BundleId:           req.BundleId,
		TestRunnerBundleId: req.TestRunnerBundleId,
		XctestConfigName:   req.XctestConfig,
		Env:                req.Env,
		Args:               req.Args,
		TestsToRun:         req.TestsToRun,
		TestsToSkip:        req.TestsToSkip,
		XcTest:             req.XcTest,
	})
	c.JSON(http.StatusAccepted, j.view())
}

// StartRunWda starts the WebDriverAgent runner (CLI: ios runwda). Body fields are
// optional and default to the standard WDA bundle id and xctest config.
// @Summary Start the WebDriverAgent runner (async job)
// @Accept json
// @Produce json
// @Param udid path string true "Device UDID"
// @Param body body runTestRequest false "optional overrides"
// @Success 202 {object} jobView
// @Router /device/{udid}/jobs/runwda [post]
func StartRunWda(c *gin.Context) {
	device := c.MustGet(IOS_KEY).(ios.DeviceEntry)
	var req runTestRequest
	_ = c.ShouldBindJSON(&req)
	if req.BundleId == "" {
		req.BundleId = defaultWDABundleID
	}
	if req.TestRunnerBundleId == "" {
		req.TestRunnerBundleId = req.BundleId
	}
	if req.XctestConfig == "" {
		req.XctestConfig = defaultWDAXctestConfig
	}
	j := startTestJob(device, "runwda", testmanagerd.TestConfig{
		BundleId:           req.BundleId,
		TestRunnerBundleId: req.TestRunnerBundleId,
		XctestConfigName:   req.XctestConfig,
		Env:                req.Env,
		Args:               req.Args,
	})
	c.JSON(http.StatusAccepted, j.view())
}

type forwardRequest struct {
	HostPort   uint16 `json:"hostPort"`
	TargetPort uint16 `json:"targetPort"`
}

// StartForward starts a TCP port forward host->device (CLI: ios forward).
// @Summary Start a port forward (async job)
// @Accept json
// @Produce json
// @Param udid path string true "Device UDID"
// @Param body body forwardRequest true "hostPort/targetPort"
// @Success 202 {object} jobView
// @Failure 400 {object} map[string]string
// @Router /device/{udid}/jobs/forward [post]
func StartForward(c *gin.Context) {
	device := c.MustGet(IOS_KEY).(ios.DeviceEntry)
	var req forwardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, err)
		return
	}
	if req.HostPort == 0 || req.TargetPort == 0 {
		RespondError(c, http.StatusBadRequest, errMissingPorts)
		return
	}
	cl, err := forward.Forward(device, req.HostPort, req.TargetPort)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}
	j := jobs.create("forward", device.Properties.SerialNumber, func() error { return cl.Close() })
	fmt.Fprintf(j.log, "forwarding 127.0.0.1:%d -> device:%d\n", req.HostPort, req.TargetPort)
	c.JSON(http.StatusAccepted, j.view())
}

// ListJobs lists jobs for a device.
// @Summary List jobs
// @Produce json
// @Param udid path string true "Device UDID"
// @Success 200 {array} jobView
// @Router /device/{udid}/jobs [get]
func ListJobs(c *gin.Context) {
	c.JSON(http.StatusOK, jobs.listForUDID(c.Param("udid")))
}

// jobForRequest fetches a job and verifies it belongs to the path's device.
func jobForRequest(c *gin.Context) (*Job, bool) {
	j, ok := jobs.get(c.Param("id"))
	if !ok || j.UDID != c.Param("udid") {
		RespondError(c, http.StatusNotFound, errJobNotFound)
		return nil, false
	}
	return j, true
}

// GetJob returns a job's status.
// @Summary Get job status
// @Produce json
// @Param udid path string true "Device UDID"
// @Param id path string true "job id"
// @Success 200 {object} jobView
// @Failure 404 {object} map[string]string
// @Router /device/{udid}/jobs/{id} [get]
func GetJob(c *gin.Context) {
	j, ok := jobForRequest(c)
	if !ok {
		return
	}
	c.JSON(http.StatusOK, j.view())
}

// StreamJobLogs streams a job's isolated log output: the buffered history first,
// then live lines until the job ends or the client disconnects.
// @Summary Stream a job's logs
// @Produce text/plain
// @Param udid path string true "Device UDID"
// @Param id path string true "job id"
// @Success 200 {string} string
// @Router /device/{udid}/jobs/{id}/logs [get]
func StreamJobLogs(c *gin.Context) {
	j, ok := jobForRequest(c)
	if !ok {
		return
	}
	for _, line := range j.log.snapshot() {
		c.Writer.WriteString(line)
	}
	c.Writer.Flush()

	ch, unsubscribe := j.log.subscribe()
	defer unsubscribe()
	c.Stream(func(w io.Writer) bool {
		line, ok := <-ch
		if !ok {
			return false
		}
		w.Write([]byte(line))
		return true
	})
}

// StopJob stops a running job (CLI: Ctrl-C on the equivalent command).
// @Summary Stop a job
// @Produce json
// @Param udid path string true "Device UDID"
// @Param id path string true "job id"
// @Success 200 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /device/{udid}/jobs/{id} [delete]
func StopJob(c *gin.Context) {
	if _, ok := jobForRequest(c); !ok {
		return
	}
	if _, err := jobs.stop(c.Param("id")); err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "job stopped"})
}
