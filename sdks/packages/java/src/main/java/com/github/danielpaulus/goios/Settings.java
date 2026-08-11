package com.github.danielpaulus.goios;

import com.github.danielpaulus.goios.generated.model.AssistiveTouchState;
import com.github.danielpaulus.goios.generated.model.GenericResponse;
import com.github.danielpaulus.goios.generated.model.TimeFormatState;
import com.github.danielpaulus.goios.generated.model.WifiRequest;

import java.util.Map;

/** Device-settings operations (AssistiveTouch, 24h clock, Wi-Fi) for a device. */
public final class Settings {

    private final Device d;

    Settings(Device d) {
        this.d = d;
    }

    /** Get AssistiveTouch state ({@code GET /assistivetouch}). */
    public AssistiveTouchState assistiveTouch() {
        return d.http().getJson(d.devicePath("/assistivetouch"), null, AssistiveTouchState.class);
    }

    /** Enable/disable AssistiveTouch ({@code PUT /assistivetouch}). */
    public AssistiveTouchState setAssistiveTouch(boolean enabled) {
        return d.http().putJson(d.devicePath("/assistivetouch"), null,
                Map.of("enabled", enabled), AssistiveTouchState.class);
    }

    /** Get the 24-hour clock setting ({@code GET /timeformat}). */
    public TimeFormatState timeFormat() {
        return d.http().getJson(d.devicePath("/timeformat"), null, TimeFormatState.class);
    }

    /** Set the 24-hour clock setting ({@code PUT /timeformat}). */
    public TimeFormatState setTimeFormat(boolean uses24Hour) {
        return d.http().putJson(d.devicePath("/timeformat"), null,
                Map.of("uses24Hour", uses24Hour), TimeFormatState.class);
    }

    /** Configure a Wi-Fi network ({@code PUT /wifi}). */
    public GenericResponse setWifi(String ssid, String password, String encType) {
        WifiRequest req = new WifiRequest().ssid(ssid);
        if (password != null) {
            req.password(password);
        }
        if (encType != null) {
            req.encType(encType);
        }
        return d.http().putJson(d.devicePath("/wifi"), null, req, GenericResponse.class);
    }

    /** Forget a Wi-Fi network ({@code DELETE /wifi}). */
    public GenericResponse removeWifi(String ssid) {
        Map<String, String> q = RawHttp.query();
        q.put("ssid", ssid);
        return d.http().deleteJson(d.devicePath("/wifi"), q, GenericResponse.class);
    }
}
