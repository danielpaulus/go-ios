package com.github.danielpaulus.goios;

/**
 * Thrown when an {@link IosClient} is built without an explicit {@code baseUrl}
 * and no local go-ios REST daemon can be discovered (no {@code GO_IOS_BASE_URL}
 * env var and no readable {@code <home>/rest-api.json} discovery file).
 */
public final class IosDiscoveryException extends RuntimeException {

    IosDiscoveryException(String message) {
        super(message);
    }

    IosDiscoveryException(String message, Throwable cause) {
        super(message, cause);
    }
}
