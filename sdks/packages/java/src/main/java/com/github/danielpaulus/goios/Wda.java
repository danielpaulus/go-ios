package com.github.danielpaulus.goios;

import com.github.danielpaulus.goios.generated.model.WdaConfig;
import com.github.danielpaulus.goios.generated.model.WdaSession;

/** WebDriverAgent (XCUITest) session operations for a single device. */
public final class Wda {

    private final Device d;

    Wda(Device d) {
        this.d = d;
    }

    /** Create a WDA session ({@code POST /wda/session}). */
    public WdaSession createSession(WdaConfig config) {
        return d.http().postJson(d.devicePath("/wda/session"), null, config, WdaSession.class);
    }

    /** Read a WDA session ({@code GET /wda/session/{id}}). */
    public WdaSession getSession(String sessionId) {
        return d.http().getJson(d.devicePath("/wda/session/" + Device.seg(sessionId)), null, WdaSession.class);
    }

    /** Delete a WDA session ({@code DELETE /wda/session/{id}}). */
    public WdaSession deleteSession(String sessionId) {
        return d.http().deleteJson(d.devicePath("/wda/session/" + Device.seg(sessionId)), null, WdaSession.class);
    }
}
