package com.github.danielpaulus.goios;

import com.fasterxml.jackson.databind.JsonNode;

import java.io.IOException;
import java.io.UncheckedIOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.function.Function;

/**
 * Locates the local go-ios REST daemon so an {@link IosClient} can be built with
 * no explicit {@code baseUrl}.
 *
 * <p>The daemon writes a discovery file at {@code <home>/rest-api.json} after it
 * binds (see the discovery contract), where {@code <home>} is {@code GO_IOS_HOME}
 * if set and non-empty, else {@code ~/.go-ios}. The file's authoritative
 * {@code baseUrl} field (scheme + host + port) is what this reads.
 *
 * <p>Resolution order used by the {@link IosClient.Builder}: an explicit
 * {@code .baseUrl(...)} wins; then the {@code GO_IOS_BASE_URL} env var; then this
 * discovery file; otherwise a clear error.
 */
public final class Discovery {

    /** Name of the discovery file the daemon writes inside the go-ios home dir. */
    public static final String DISCOVERY_FILE = "rest-api.json";

    /** Env var overriding the go-ios home directory. */
    static final String HOME_ENV = "GO_IOS_HOME";

    /** Env var overriding the base URL (takes precedence over the discovery file). */
    static final String BASE_URL_ENV = "GO_IOS_BASE_URL";

    private final Function<String, String> env;
    private final Function<String, String> props;

    private Discovery(Function<String, String> env, Function<String, String> props) {
        this.env = env;
        this.props = props;
    }

    /** Discovery backed by the real process environment and system properties. */
    static Discovery system() {
        return new Discovery(System::getenv, System::getProperty);
    }

    /** Testable variant with injected environment / system-property lookups. */
    static Discovery of(Function<String, String> env, Function<String, String> props) {
        return new Discovery(env, props);
    }

    /**
     * The go-ios home directory: {@code GO_IOS_HOME} if set and non-empty, else
     * {@code <user.home>/.go-ios}.
     */
    Path home() {
        String h = env.apply(HOME_ENV);
        if (h != null && !h.isBlank()) {
            return Path.of(h);
        }
        String userHome = props.apply("user.home");
        return Path.of(userHome == null ? "" : userHome, ".go-ios");
    }

    /** Absolute path of the discovery file within {@link #home()}. */
    Path discoveryFile() {
        return home().resolve(DISCOVERY_FILE);
    }

    /**
     * Resolve the daemon base URL, honoring the {@code GO_IOS_BASE_URL} env var
     * first, then the on-disk discovery file.
     *
     * @throws IosDiscoveryException if neither is available.
     */
    String resolveBaseUrl() {
        String envUrl = env.apply(BASE_URL_ENV);
        if (envUrl != null && !envUrl.isBlank()) {
            return envUrl;
        }
        return readDiscoveryFile();
    }

    /**
     * Read {@code <home>/rest-api.json} and return its {@code baseUrl}.
     *
     * @throws IosDiscoveryException if the file is missing, unreadable, malformed,
     *                               or lacks a usable {@code baseUrl}.
     */
    String readDiscoveryFile() {
        Path file = discoveryFile();
        if (!Files.isRegularFile(file)) {
            throw notFound(file, null);
        }
        byte[] raw;
        try {
            raw = Files.readAllBytes(file);
        } catch (IOException e) {
            throw notFound(file, e);
        }
        JsonNode root;
        try {
            root = RawHttp.MAPPER.readTree(raw);
        } catch (IOException e) {
            throw notFound(file, e);
        }
        JsonNode baseUrl = root == null ? null : root.get("baseUrl");
        if (baseUrl == null || !baseUrl.isTextual() || baseUrl.asText().isBlank()) {
            throw notFound(file, null);
        }
        return baseUrl.asText();
    }

    private static IosDiscoveryException notFound(Path file, Throwable cause) {
        String msg = "no local go-ios REST daemon found at " + file
                + "; start the go-ios REST API or set baseUrl";
        return cause == null ? new IosDiscoveryException(msg)
                : new IosDiscoveryException(msg, cause);
    }
}
