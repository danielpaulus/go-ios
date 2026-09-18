package com.github.danielpaulus.goios;

import com.fasterxml.jackson.core.type.TypeReference;
import com.github.danielpaulus.goios.generated.model.AssistiveTouchState;
import com.github.danielpaulus.goios.generated.model.BatteryInfo;
import com.github.danielpaulus.goios.generated.model.BatteryRegistry;
import com.github.danielpaulus.goios.generated.model.CrashListing;
import com.github.danielpaulus.goios.generated.model.DevModeState;
import com.github.danielpaulus.goios.generated.model.DeviceDate;
import com.github.danielpaulus.goios.generated.model.DeviceName;
import com.github.danielpaulus.goios.generated.model.DiskSpaceInfo;
import com.github.danielpaulus.goios.generated.model.FileListing;
import com.github.danielpaulus.goios.generated.model.FilePushResult;
import com.github.danielpaulus.goios.generated.model.GenericResponse;
import com.github.danielpaulus.goios.generated.model.LanguageConfiguration;
import com.github.danielpaulus.goios.generated.model.MemLimitResult;
import com.github.danielpaulus.goios.generated.model.MountedImages;
import com.github.danielpaulus.goios.generated.model.NetworkInfo;
import com.github.danielpaulus.goios.generated.model.PasteboardContent;
import com.github.danielpaulus.goios.generated.model.PrepareResult;
import com.github.danielpaulus.goios.generated.model.StatusOk;
import com.github.danielpaulus.goios.generated.model.TimeFormatState;
import com.github.danielpaulus.goios.generated.model.UnlockToken;
import com.github.danielpaulus.goios.generated.model.VoiceOverState;
import com.github.danielpaulus.goios.generated.model.ZoomTouchState;
import com.github.danielpaulus.goios.stream.BinaryStream;
import com.github.danielpaulus.goios.stream.EventDecoder;
import com.github.danielpaulus.goios.stream.SseReader;

import java.net.URLEncoder;
import java.nio.charset.StandardCharsets;
import java.util.List;
import java.util.Map;

/**
 * All operations scoped to a single device udid. Grouped sub-facades
 * ({@link #apps()}, {@link #files()}, {@link #ui()}, …) mirror the CLI/REST
 * grouping; flat device-level operations (battery, reboot, screenshot, streams)
 * live directly on this type.
 */
public final class Device {

    private final RawHttp http;
    private final String udid;

    // Grouped sub-facades.
    private final Apps apps;
    private final Wda wda;
    private final Files files;
    private final Crashes crashes;
    private final Jobs jobs;
    private final Settings settings;
    private final Media media;
    private final Mdm mdm;
    private final Fsync fsync;
    private final WebInspector webinspector;
    private final Ui ui;

    Device(RawHttp http, String udid) {
        this.http = http;
        this.udid = udid;
        this.apps = new Apps(this);
        this.wda = new Wda(this);
        this.files = new Files(this);
        this.crashes = new Crashes(this);
        this.jobs = new Jobs(this);
        this.settings = new Settings(this);
        this.media = new Media(this);
        this.mdm = new Mdm(this);
        this.fsync = new Fsync(this);
        this.webinspector = new WebInspector(this);
        this.ui = new Ui(this);
    }

    /** The udid this handle is scoped to. */
    public String udid() {
        return udid;
    }

    // -- URL helpers -------------------------------------------------------

    static String seg(String s) {
        return URLEncoder.encode(s, StandardCharsets.UTF_8);
    }

    private String path(String suffix) {
        return "/device/" + seg(udid) + suffix;
    }

    private static String bool(boolean b) {
        return b ? "true" : "false";
    }

    // -- device info -------------------------------------------------------

    /** Get device info (lockdown + instruments values) ({@code GET /info}). */
    public Object info() {
        return http.getJson(path("/info"), null, Object.class);
    }

    /** Get the device name ({@code GET /devicename}). */
    public DeviceName deviceName() {
        return http.getJson(path("/devicename"), null, DeviceName.class);
    }

    /** Get the device date/time ({@code GET /date}). */
    public DeviceDate date() {
        return http.getJson(path("/date"), null, DeviceDate.class);
    }

    /** Get battery info ({@code GET /battery}). */
    public BatteryInfo battery() {
        return http.getJson(path("/battery"), null, BatteryInfo.class);
    }

    /** Get raw IOKit battery-registry values ({@code GET /battery/registry}). */
    public BatteryRegistry batteryRegistry() {
        return http.getJson(path("/battery/registry"), null, BatteryRegistry.class);
    }

    /** Get IORegistry diagnostics ({@code GET /diagnostics}). */
    public Object diagnostics() {
        return http.getJson(path("/diagnostics"), null, Object.class);
    }

    /** Get filesystem disk-space usage ({@code GET /diskspace}). */
    public DiskSpaceInfo diskSpace() {
        return http.getJson(path("/diskspace"), null, DiskSpaceInfo.class);
    }

    /** Get the device's network/IP info ({@code GET /ip}). */
    public NetworkInfo ip() {
        return http.getJson(path("/ip"), null, NetworkInfo.class);
    }

    /** List RemoteServiceDiscovery services exposed over the tunnel ({@code GET /rsd}). */
    public Object rsd() {
        return http.getJson(path("/rsd"), null, Object.class);
    }

    /** Query MobileGestalt values by key ({@code GET /mobilegestalt}). */
    public Object mobileGestalt(List<String> keys) {
        Map<String, String> q = RawHttp.query();
        if (keys != null && !keys.isEmpty()) {
            q.put("key", String.join(",", keys)); // spec: explode:false -> comma-joined
        }
        return http.getJson(path("/mobilegestalt"), q, Object.class);
    }

    /** List running processes ({@code GET /processes}). */
    public Object processes(Boolean apps) {
        Map<String, String> q = RawHttp.query();
        if (apps != null) {
            q.put("apps", bool(apps));
        }
        return http.getJson(path("/processes"), q, Object.class);
    }

    /** Read all lockdown values ({@code GET /lockdown}). */
    public Object lockdown() {
        return http.getJson(path("/lockdown"), null, Object.class);
    }

    /** Read lockdown values scoped to a domain ({@code GET /lockdown?domain=...}). */
    public Object lockdown(String domain) {
        Map<String, String> q = RawHttp.query();
        if (domain != null) {
            q.put("domain", domain);
        }
        return http.getJson(path("/lockdown"), q, Object.class);
    }

    // -- management --------------------------------------------------------

    /** Activate the device ({@code POST /activate}). */
    public GenericResponse activate() {
        return http.postJson(path("/activate"), null, null, GenericResponse.class);
    }

    /** Reboot the device ({@code POST /reboot}). */
    public GenericResponse reboot() {
        return http.postJson(path("/reboot"), null, null, GenericResponse.class);
    }

    /** Shut down the device ({@code POST /shutdown}). */
    public GenericResponse shutdown() {
        return http.postJson(path("/shutdown"), null, null, GenericResponse.class);
    }

    /** Erase the device ({@code POST /erase}); requires {@code confirm=true}. */
    public GenericResponse erase(boolean confirm) {
        Map<String, String> q = RawHttp.query();
        q.put("confirm", bool(confirm));
        return http.postJson(path("/erase"), q, null, GenericResponse.class);
    }

    /** Get developer-mode state ({@code GET /devmode}). */
    public DevModeState devMode() {
        return http.getJson(path("/devmode"), null, DevModeState.class);
    }

    /** Set developer mode ({@code POST /devmode}); {@code action} is {@code enable} or {@code reveal}. */
    public GenericResponse setDevMode(String action, Boolean enablePostRestart) {
        Map<String, Object> body = new java.util.LinkedHashMap<>();
        body.put("action", action);
        if (enablePostRestart != null) {
            body.put("enablePostRestart", enablePostRestart);
        }
        return http.postJson(path("/devmode"), null, body, GenericResponse.class);
    }

    /** Get language/locale configuration ({@code GET /lang}). */
    public LanguageConfiguration lang() {
        return http.getJson(path("/lang"), null, LanguageConfiguration.class);
    }

    /** Set language and/or locale ({@code PUT /lang}). */
    public LanguageConfiguration setLang(String language, String locale) {
        Map<String, Object> body = new java.util.LinkedHashMap<>();
        if (language != null) {
            body.put("language", language);
        }
        if (locale != null) {
            body.put("locale", locale);
        }
        return http.putJson(path("/lang"), null, body, LanguageConfiguration.class);
    }

    /** Waive the memory limit for a process ({@code POST /memlimitoff}). */
    public MemLimitResult memlimitoff(String process) {
        Map<String, Object> body = new java.util.LinkedHashMap<>();
        if (process != null) {
            body.put("process", process);
        }
        return http.postJson(path("/memlimitoff"), null, body, MemLimitResult.class);
    }

    // -- location ----------------------------------------------------------

    /** Set a simulated GPS location ({@code PUT /setlocation}). */
    public GenericResponse setLocation(double latitude, double longitude) {
        Map<String, String> q = RawHttp.query();
        q.put("latitude", Double.toString(latitude));
        q.put("longitude", Double.toString(longitude));
        return http.requestJson("PUT", path("/setlocation"), q, null, null, GenericResponse.class);
    }

    /** Replay a GPX track as simulated location ({@code PUT /setlocation/gpx}), multipart {@code gpx}. */
    public GenericResponse setLocationGpx(byte[] gpx) {
        return http.multipart("PUT", path("/setlocation/gpx"), null,
                RawHttp.parts(RawHttp.Part.file("gpx", "track.gpx", gpx)),
                GenericResponse.class);
    }

    /** Reset the simulated location ({@code POST /resetlocation}). */
    public GenericResponse resetLocation() {
        return http.postJson(path("/resetlocation"), null, null, GenericResponse.class);
    }

    /** Reset accessibility settings ({@code POST /resetaccessibility}). */
    public GenericResponse resetAccessibility() {
        return http.postJson(path("/resetaccessibility"), null, null, GenericResponse.class);
    }

    // -- accessibility -----------------------------------------------------

    /** Get a snapshot of the focused accessibility element ({@code GET /ax}). */
    public Object ax() {
        return http.getJson(path("/ax"), null, Object.class);
    }

    /** Run the accessibility audit against the focused app ({@code POST /ax/audit}). */
    public Object axAudit(Integer timeoutSeconds) {
        Map<String, String> q = RawHttp.query();
        if (timeoutSeconds != null) {
            q.put("timeout", Integer.toString(timeoutSeconds));
        }
        return http.postJson(path("/ax/audit"), q, null, Object.class);
    }

    /** Get VoiceOver state ({@code GET /voiceover}). */
    public VoiceOverState voiceOver() {
        return http.getJson(path("/voiceover"), null, VoiceOverState.class);
    }

    /** Enable/disable VoiceOver ({@code PUT /voiceover}). */
    public VoiceOverState setVoiceOver(boolean enabled) {
        return http.putJson(path("/voiceover"), null, Map.of("enabled", enabled), VoiceOverState.class);
    }

    /** Get Zoom (touch) state ({@code GET /zoom}). */
    public ZoomTouchState zoom() {
        return http.getJson(path("/zoom"), null, ZoomTouchState.class);
    }

    /** Enable/disable Zoom (touch) ({@code PUT /zoom}). */
    public ZoomTouchState setZoom(boolean enabled) {
        return http.putJson(path("/zoom"), null, Map.of("enabled", enabled), ZoomTouchState.class);
    }

    // -- screenshot / media flat ops ---------------------------------------

    /** Capture a PNG screenshot as raw bytes ({@code GET /screenshot}). */
    public byte[] screenshot() {
        return http.getBytes(path("/screenshot"), null);
    }

    // -- conditions / images / profiles ------------------------------------

    /** List available condition profile types ({@code GET /conditions}). */
    public Object conditions() {
        return http.getJson(path("/conditions"), null, Object.class);
    }

    /** Enable a device condition ({@code PUT /enable-condition}). */
    public GenericResponse enableCondition(String profileTypeId, String profileId) {
        Map<String, String> q = RawHttp.query();
        q.put("profileTypeID", profileTypeId);
        q.put("profileID", profileId);
        return http.requestJson("PUT", path("/enable-condition"), q, null, null, GenericResponse.class);
    }

    /** Disable the active device condition ({@code POST /disable-condition}). */
    public GenericResponse disableCondition() {
        return http.postJson(path("/disable-condition"), null, null, GenericResponse.class);
    }

    /** List available developer disk images on the server ({@code GET /image}). */
    public Object images() {
        return http.getJson(path("/image"), null, Object.class);
    }

    /** List mounted developer disk images ({@code GET /image/list}). */
    public MountedImages mountedImages() {
        return http.getJson(path("/image/list"), null, MountedImages.class);
    }

    /** Auto-resolve and mount the matching developer disk image ({@code PUT /image?auto=true}). */
    public GenericResponse mountImageAuto(String basedir) {
        Map<String, String> q = RawHttp.query();
        q.put("auto", "true");
        if (basedir != null) {
            q.put("basedir", basedir);
        }
        return http.requestJson("PUT", path("/image"), q, null, null, GenericResponse.class);
    }

    /** Mount a developer disk image from raw bytes ({@code PUT /image}). */
    public GenericResponse mountImage(byte[] image) {
        return http.requestJson("PUT", path("/image"), null, image,
                "application/octet-stream", GenericResponse.class);
    }

    /** Unmount the developer disk image ({@code DELETE /image}). */
    public GenericResponse unmountImage() {
        return http.deleteJson(path("/image"), null, GenericResponse.class);
    }

    /** List installed configuration profiles ({@code GET /profiles}). */
    public Object profiles() {
        return http.getJson(path("/profiles"), null, Object.class);
    }

    /** Install a {@code .mobileconfig} profile ({@code POST /profiles}), multipart. */
    public GenericResponse addProfile(byte[] profile, byte[] p12, String password) {
        return http.multipart("POST", path("/profiles"), null, RawHttp.parts(
                RawHttp.Part.file("profile", "profile.mobileconfig", profile),
                p12 == null ? null : RawHttp.Part.file("p12", "identity.p12", p12),
                password == null ? null : RawHttp.Part.field("password", password)),
                GenericResponse.class);
    }

    /** Remove an installed profile by identifier ({@code DELETE /profiles/{name}}). */
    public GenericResponse removeProfile(String name) {
        return http.deleteJson(path("/profiles/" + seg(name)), null, GenericResponse.class);
    }

    // -- pairing / prepare / proxy -----------------------------------------

    /** Pair the device ({@code POST /pair}); pass {@code p12} for supervised pairing. */
    public GenericResponse pair(boolean supervised, byte[] p12, String supervisionPassword) {
        Map<String, String> q = RawHttp.query();
        q.put("supervised", bool(supervised));
        if (p12 == null) {
            return http.postJson(path("/pair"), q, null, GenericResponse.class);
        }
        return http.multipart("POST", path("/pair"), q, RawHttp.parts(
                RawHttp.Part.file("p12file", "identity.p12", p12),
                supervisionPassword == null ? null
                        : RawHttp.Part.field("supervisionPassword", supervisionPassword)),
                GenericResponse.class);
    }

    /**
     * Run the device preparation/provisioning flow ({@code POST /prepare}), multipart.
     * Pass {@code cert} to supervise; omit it to prepare unsupervised.
     */
    public PrepareResult prepare(byte[] cert, String p12password, List<String> skip,
                                 String orgname, String locale, String lang) {
        java.util.List<RawHttp.Part> parts = new java.util.ArrayList<>();
        if (cert != null) {
            parts.add(RawHttp.Part.file("cert", "supervision.p12", cert));
        }
        if (p12password != null) {
            parts.add(RawHttp.Part.field("p12password", p12password));
        }
        if (skip != null) {
            for (String s : skip) {
                parts.add(RawHttp.Part.field("skip", s));
            }
        }
        if (orgname != null) {
            parts.add(RawHttp.Part.field("orgname", orgname));
        }
        if (locale != null) {
            parts.add(RawHttp.Part.field("locale", locale));
        }
        if (lang != null) {
            parts.add(RawHttp.Part.field("lang", lang));
        }
        return http.multipart("POST", path("/prepare"), null, parts, PrepareResult.class);
    }

    /** Set a global HTTP proxy (supervised) ({@code PUT /httpproxy}), multipart. */
    public GenericResponse setHttpProxy(String host, String port, String user, String pass,
                                        byte[] p12, String p12password) {
        return http.multipart("PUT", path("/httpproxy"), null, RawHttp.parts(
                RawHttp.Part.field("host", host),
                RawHttp.Part.field("port", port),
                user == null ? null : RawHttp.Part.field("user", user),
                pass == null ? null : RawHttp.Part.field("pass", pass),
                p12 == null ? null : RawHttp.Part.file("p12", "identity.p12", p12),
                p12password == null ? null : RawHttp.Part.field("password", p12password)),
                GenericResponse.class);
    }

    /** Remove the global HTTP proxy ({@code DELETE /httpproxy}). */
    public GenericResponse removeHttpProxy() {
        return http.deleteJson(path("/httpproxy"), null, GenericResponse.class);
    }

    // -- mdm (also exposed via mdm() group) --------------------------------

    /** MDM security info ({@code POST /mdm/security-info}), multipart. */
    public Object securityInfo(byte[] p12, String password) {
        return mdm.securityInfo(p12, password);
    }

    /** Fetch the escrow unlock token ({@code POST /mdm/fetch-unlock-token}), multipart. */
    public UnlockToken fetchUnlockToken(byte[] p12, String password) {
        return mdm.fetchUnlockToken(p12, password);
    }

    // -- SSE streams -------------------------------------------------------

    /** Stream syslog messages ({@code GET /syslog}). */
    public SseReader syslog() {
        return syslog(false);
    }

    public SseReader syslog(boolean includeHeartbeats) {
        return http.sseStream(path("/syslog"), null, EventDecoder.SYSLOG, includeHeartbeats);
    }

    /** Stream app-state notifications ({@code GET /notifications}). */
    public SseReader notifications() {
        return notifications(false);
    }

    public SseReader notifications(boolean includeHeartbeats) {
        return http.sseStream(path("/notifications"), null, EventDecoder.NOTIFICATIONS, includeHeartbeats);
    }

    /** Stream os_trace entries ({@code GET /ostrace}) with optional AND filters. */
    public SseReader ostrace(Integer pid, String level, String subsystem,
                             String match, String exclude, boolean includeHeartbeats) {
        Map<String, String> q = RawHttp.query();
        if (pid != null) {
            q.put("pid", Integer.toString(pid));
        }
        if (level != null) {
            q.put("level", level);
        }
        if (subsystem != null) {
            q.put("subsystem", subsystem);
        }
        if (match != null) {
            q.put("match", match);
        }
        if (exclude != null) {
            q.put("exclude", exclude);
        }
        return http.sseStream(path("/ostrace"), q, EventDecoder.OSTRACE, includeHeartbeats);
    }

    public SseReader ostrace() {
        return ostrace(null, null, null, null, null, false);
    }

    /** Stream device attach/detach/pair events ({@code GET /listen}). */
    public SseReader listen() {
        return listen(false);
    }

    public SseReader listen(boolean includeHeartbeats) {
        return http.sseStream(path("/listen"), null, EventDecoder.LISTEN, includeHeartbeats);
    }

    /** Stream CPU-usage samples ({@code GET /sysmontap}). */
    public SseReader sysmontap() {
        return sysmontap(false);
    }

    public SseReader sysmontap(boolean includeHeartbeats) {
        return http.sseStream(path("/sysmontap"), null, EventDecoder.SYSMONTAP, includeHeartbeats);
    }

    // -- binary streams ----------------------------------------------------

    /** Live MJPEG screenshot stream ({@code GET /screenshot/stream}); returns raw bytes. */
    public BinaryStream screenshotStream() {
        return screenshotStream(null);
    }

    public BinaryStream screenshotStream(Integer quality) {
        Map<String, String> q = RawHttp.query();
        if (quality != null) {
            q.put("quality", Integer.toString(quality));
        }
        return http.binaryStream(path("/screenshot/stream"), q);
    }

    /** Live packet capture as a libpcap byte stream ({@code GET /pcap}). */
    public BinaryStream pcap() {
        return pcap(null);
    }

    public BinaryStream pcap(Integer timeoutSeconds) {
        Map<String, String> q = RawHttp.query();
        if (timeoutSeconds != null) {
            q.put("timeout", Integer.toString(timeoutSeconds));
        }
        return http.binaryStream(path("/pcap"), q);
    }

    // -- group accessors ---------------------------------------------------

    public Apps apps() {
        return apps;
    }

    public Wda wda() {
        return wda;
    }

    public Files files() {
        return files;
    }

    public Crashes crashes() {
        return crashes;
    }

    public Jobs jobs() {
        return jobs;
    }

    public Settings settings() {
        return settings;
    }

    public Media media() {
        return media;
    }

    public Mdm mdm() {
        return mdm;
    }

    public Fsync fsync() {
        return fsync;
    }

    /** Read the device supervision/cloud configuration ({@code GET /cloudconfig}). */
    public Object cloudConfig() {
        return http.getJson(path("/cloudconfig"), null, Object.class);
    }

    public WebInspector webinspector() {
        return webinspector;
    }

    public Ui ui() {
        return ui;
    }

    // -- internal accessors for grouped sub-facades ------------------------

    RawHttp http() {
        return http;
    }

    String devicePath(String suffix) {
        return path(suffix);
    }
}
