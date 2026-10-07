package com.github.danielpaulus.goios.stream;

import java.io.IOException;
import java.io.InputStream;

/**
 * A raw byte stream over an {@code x-stream: binary} endpoint (UI video stream,
 * MJPEG screenshot stream, live pcap capture).
 *
 * <p>Unlike {@link SseReader} these endpoints emit an opaque byte stream (MJPEG
 * multipart, an H.264 elementary stream, or a libpcap capture), not typed SSE
 * frames. {@code BinaryStream} is a plain {@link InputStream} the caller reads
 * and consumes directly; closing it releases (and cancels) the underlying HTTP
 * response so a long-lived capture can be stopped at any time.
 */
public final class BinaryStream extends InputStream {

    private final InputStream body;
    private final Runnable onClose;
    private final String contentType;
    private boolean closed;

    public BinaryStream(InputStream body, String contentType, Runnable onClose) {
        this.body = body;
        this.contentType = contentType;
        this.onClose = onClose;
    }

    /** The response {@code Content-Type} (e.g. {@code multipart/x-mixed-replace}, {@code application/vnd.tcpdump.pcap}). */
    public String contentType() {
        return contentType;
    }

    @Override
    public int read() throws IOException {
        return body.read();
    }

    @Override
    public int read(byte[] b, int off, int len) throws IOException {
        return body.read(b, off, len);
    }

    @Override
    public int available() throws IOException {
        return body.available();
    }

    @Override
    public void close() throws IOException {
        if (closed) {
            return;
        }
        closed = true;
        try {
            body.close();
        } finally {
            if (onClose != null) {
                onClose.run();
            }
        }
    }
}
