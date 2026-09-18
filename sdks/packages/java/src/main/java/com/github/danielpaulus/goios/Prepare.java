package com.github.danielpaulus.goios;

import com.github.danielpaulus.goios.generated.model.PrepareSkipOptions;
import com.github.danielpaulus.goios.generated.model.SupervisionCert;

/** Host-scoped device-preparation helpers ({@code /prepare/*}). */
public final class Prepare {

    private final RawHttp http;

    Prepare(RawHttp http) {
        this.http = http;
    }

    /**
     * Create a self-signed supervision certificate + key
     * ({@code POST /prepare/create-cert}).
     */
    public SupervisionCert createCert() {
        return http.postJson("/prepare/create-cert", null, null, SupervisionCert.class);
    }

    /**
     * List the setup panes that {@code prepare} can skip
     * ({@code GET /prepare/skip-options}).
     */
    public PrepareSkipOptions skipOptions() {
        return http.getJson("/prepare/skip-options", null, PrepareSkipOptions.class);
    }
}
