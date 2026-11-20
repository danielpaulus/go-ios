package com.github.danielpaulus.goios;

import com.fasterxml.jackson.core.type.TypeReference;
import com.github.danielpaulus.goios.generated.model.AgentShutdown;
import com.github.danielpaulus.goios.generated.model.Tunnel;
import com.github.danielpaulus.goios.generated.model.TunnelStopped;

import java.util.List;

/** userspace-tunnel (RemoteXPC) management (iOS 17+). */
public final class Tunnels {

    private final RawHttp http;

    Tunnels(RawHttp http) {
        this.http = http;
    }

    /** List active tunnels ({@code GET /tunnels}). */
    public List<Tunnel> list() {
        List<Tunnel> tunnels = http.getJson("/tunnels", null, new TypeReference<List<Tunnel>>() { });
        return tunnels == null ? List.of() : tunnels;
    }

    /** Refresh the tunnel for {@code udid} ({@code POST /tunnels/{udid}/refresh}). */
    public Tunnel refresh(String udid) {
        return http.postJson("/tunnels/" + Device.seg(udid) + "/refresh", null, null, Tunnel.class);
    }

    /** Stop the tunnel for {@code udid} ({@code DELETE /tunnels/{udid}}). */
    public TunnelStopped delete(String udid) {
        return http.deleteJson("/tunnels/" + Device.seg(udid), null, TunnelStopped.class);
    }

    /** Shut down the whole tunnel agent ({@code POST /tunnel-agent/shutdown}). */
    public AgentShutdown shutdownAgent() {
        return http.postJson("/tunnel-agent/shutdown", null, null, AgentShutdown.class);
    }
}
