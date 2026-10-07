import "@typespec/http";
import "@typespec/openapi";
import "@typespec/sse";

import "./models.tsp";
import "./streaming.tsp";

using Http;
using OpenAPI;
using SSE;

namespace GoIos.Devices;

// ===========================================================================
// Async jobs subsystem (CLI: ios runtest/runwda/forward). Long-running device
// operations run in the background with an isolated, streamable log; clients
// poll status or stream logs and stop when done. All under /device/{udid}.
// ===========================================================================

/**
 * Start an XCUITest/unit-test run as an async job (CLI: `ios runtest`).
 * Returns `202` with the created job.
 */
@route("/jobs/runtest")
@post
@summary("Start test run (job)")
op startRunTest(
  ...UdidPath,
  @body body: RunTestRequest,
): {
  @statusCode statusCode: 202;
  @body job: Job;
} | BadRequestError | DeviceErrors;

/**
 * Start the WebDriverAgent runner as an async job (CLI: `ios runwda`). Body
 * fields are optional and default to the standard WDA bundle id and config.
 */
@route("/jobs/runwda")
@post
@summary("Start WDA runner (job)")
op startRunWda(
  ...UdidPath,
  @body body?: RunTestRequest,
): {
  @statusCode statusCode: 202;
  @body job: Job;
} | DeviceErrors;

/**
 * Start a TCP port forward host→device as an async job (CLI: `ios forward`).
 */
@route("/jobs/forward")
@post
@summary("Start port forward (job)")
op startForward(
  ...UdidPath,
  @body body: ForwardRequest,
): {
  @statusCode statusCode: 202;
  @body job: Job;
} | BadRequestError | DeviceErrors;

/**
 * List jobs for a device.
 */
@route("/jobs")
@get
@summary("List jobs")
op listJobs(...UdidPath): {
  @body jobs: Job[];
} | DeviceErrors;

/**
 * Get a job's status. Returns `404` for an unknown job on this device.
 */
@route("/jobs/{id}")
@get
@summary("Get job")
op getJob(
  ...UdidPath,

  /** The job id. */
  @path id: string,
): Job | NotFoundError | DeviceErrors;

/**
 * Stream a job's log output as Server-Sent Events: the buffered history first,
 * then live lines until the job ends or the client disconnects.
 */
@route("/jobs/{id}/logs")
@get
@summary("Stream job logs (SSE)")
@extension(
  "x-sse-events",
  #{
    schema: "JobLogEvents",
    events: #{ log: "JobLogLine", heartbeat: "Heartbeat" },
  }
)
op streamJobLogs(
  ...UdidPath,

  /** The job id. */
  @path id: string,
): SSEStream<JobLogEvents> | NotFoundError | DeviceErrors;

/**
 * Stop a running job, or purge an already-terminal one from the registry
 * (CLI: Ctrl-C on the equivalent command).
 */
@route("/jobs/{id}")
@delete
@summary("Stop or delete job")
op stopJob(
  ...UdidPath,

  /** The job id. */
  @path id: string,
): GenericResponse | NotFoundError | DeviceErrors;
