package com.github.danielpaulus.goios;

import com.fasterxml.jackson.core.type.TypeReference;
import com.github.danielpaulus.goios.generated.model.GenericResponse;
import com.github.danielpaulus.goios.generated.model.Job;
import com.github.danielpaulus.goios.generated.model.RunTestRequest;
import com.github.danielpaulus.goios.stream.EventDecoder;
import com.github.danielpaulus.goios.stream.SseReader;

import java.util.List;
import java.util.Map;

/** Asynchronous server-side jobs (test runs, WDA, port forwards) for a device. */
public final class Jobs {

    private final Device d;

    Jobs(Device d) {
        this.d = d;
    }

    /** Start an XCUITest run job ({@code POST /jobs/runtest}). */
    public Job runTest(RunTestRequest config) {
        return d.http().postJson(d.devicePath("/jobs/runtest"), null, config, Job.class);
    }

    /** Start a WebDriverAgent run job ({@code POST /jobs/runwda}). */
    public Job runWda(RunTestRequest config) {
        return d.http().postJson(d.devicePath("/jobs/runwda"), null, config, Job.class);
    }

    /** Start a TCP port-forward job ({@code POST /jobs/forward}). */
    public Job forward(int hostPort, int targetPort) {
        Map<String, Object> body = new java.util.LinkedHashMap<>();
        body.put("hostPort", hostPort);
        body.put("targetPort", targetPort);
        return d.http().postJson(d.devicePath("/jobs/forward"), null, body, Job.class);
    }

    /** List active jobs ({@code GET /jobs}). */
    public List<Job> list() {
        List<Job> jobs = d.http().getJson(d.devicePath("/jobs"), null, new TypeReference<List<Job>>() { });
        return jobs == null ? List.of() : jobs;
    }

    /** Get one job's status ({@code GET /jobs/{id}}). */
    public Job get(String jobId) {
        return d.http().getJson(d.devicePath("/jobs/" + Device.seg(jobId)), null, Job.class);
    }

    /** Stop/delete a job ({@code DELETE /jobs/{id}}). */
    public GenericResponse delete(String jobId) {
        return d.http().deleteJson(d.devicePath("/jobs/" + Device.seg(jobId)), null, GenericResponse.class);
    }

    /** Stream a job's log lines ({@code GET /jobs/{id}/logs}) as typed events. */
    public SseReader logs(String jobId) {
        return logs(jobId, false);
    }

    public SseReader logs(String jobId, boolean includeHeartbeats) {
        return d.http().sseStream(d.devicePath("/jobs/" + Device.seg(jobId) + "/logs"),
                null, EventDecoder.JOB_LOGS, includeHeartbeats);
    }
}
