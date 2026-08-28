package com.github.danielpaulus.goios;

import com.github.danielpaulus.goios.generated.model.GenericResponse;
import com.github.danielpaulus.goios.generated.model.PasteboardContent;

import java.nio.charset.StandardCharsets;

/** Wallpaper / icon-layout / pasteboard operations for a single device. */
public final class Media {

    private final Device d;

    Media(Device d) {
        this.d = d;
    }

    /** Get the current wallpaper PNG bytes ({@code GET /wallpaper}). */
    public byte[] wallpaper() {
        return d.http().getBytes(d.devicePath("/wallpaper"), null);
    }

    /** Set the wallpaper (supervised) ({@code PUT /wallpaper}), multipart. */
    public GenericResponse setWallpaper(byte[] image, byte[] p12, String password, String screen) {
        return d.http().multipart("PUT", d.devicePath("/wallpaper"), null, RawHttp.parts(
                RawHttp.Part.file("image", "wallpaper.png", image),
                RawHttp.Part.file("p12", "identity.p12", p12),
                password == null ? null : RawHttp.Part.field("password", password),
                screen == null ? null : RawHttp.Part.field("screen", screen)),
                GenericResponse.class);
    }

    /** Get the SpringBoard icon layout ({@code GET /icon-layout}). */
    public Object iconLayout() {
        return d.http().getJson(d.devicePath("/icon-layout"), null, Object.class);
    }

    /** Set the SpringBoard icon layout ({@code PUT /icon-layout}). */
    public GenericResponse setIconLayout(Object layout) {
        return d.http().putJson(d.devicePath("/icon-layout"), null, layout, GenericResponse.class);
    }

    /** Read the device pasteboard ({@code GET /pasteboard}). */
    public PasteboardContent pasteboard() {
        return d.http().getJson(d.devicePath("/pasteboard"), null, PasteboardContent.class);
    }

    /** Write text to the device pasteboard ({@code PUT /pasteboard}), {@code text/plain}. */
    public GenericResponse setPasteboard(String text) {
        return d.http().requestJson("PUT", d.devicePath("/pasteboard"), null,
                text.getBytes(StandardCharsets.UTF_8), "text/plain", GenericResponse.class);
    }
}
