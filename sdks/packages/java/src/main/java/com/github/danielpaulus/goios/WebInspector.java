package com.github.danielpaulus.goios;

import com.github.danielpaulus.goios.generated.model.WebInspectorEvalResult;
import com.github.danielpaulus.goios.generated.model.WebInspectorLaunchResult;

import java.util.LinkedHashMap;
import java.util.Map;

/** Safari Web Inspector operations for a single device ({@code /webinspector/*}). */
public final class WebInspector {

    private final Device d;

    WebInspector(Device d) {
        this.d = d;
    }

    /** List inspectable pages ({@code GET /webinspector/pages}). */
    public Object pages() {
        return d.http().getJson(d.devicePath("/webinspector/pages"), null, Object.class);
    }

    /** Open a URL in a new inspectable page ({@code POST /webinspector/launch}). */
    public WebInspectorLaunchResult launch(String url, String bundleId) {
        Map<String, Object> body = new LinkedHashMap<>();
        if (url != null) {
            body.put("url", url);
        }
        if (bundleId != null) {
            body.put("bundleId", bundleId);
        }
        return d.http().postJson(d.devicePath("/webinspector/launch"), null, body, WebInspectorLaunchResult.class);
    }

    /** Evaluate JavaScript in an inspectable page ({@code POST /webinspector/eval}). */
    public WebInspectorEvalResult eval(String script, String page, String bundleId) {
        Map<String, Object> body = new LinkedHashMap<>();
        body.put("script", script);
        if (page != null) {
            body.put("page", page);
        }
        if (bundleId != null) {
            body.put("bundleId", bundleId);
        }
        return d.http().postJson(d.devicePath("/webinspector/eval"), null, body, WebInspectorEvalResult.class);
    }
}
