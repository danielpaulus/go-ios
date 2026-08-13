package remote

// indexHTML is the entire, self-contained web UI. No external/CDN assets: the
// HTML, CSS and JS are inlined so `ios remote` serves it standalone.
//
// The screen is rendered two ways, chosen automatically:
//   - Primary: the DeviceKit runner's hardware H.264 (GET /video.h264, an
//     Annex-B elementary stream) is fetched as a stream, parsed into NAL units,
//     and decoded with the WebCodecs VideoDecoder into a <canvas>. This is the
//     efficient path (a few KB/s, delta-compressed).
//   - Fallback: if WebCodecs is unavailable or the decoder errors, an
//     <img src="/screen"> shows the runner's MJPEG instead. The switch is
//     automatic.
//
// Clicks/drags are translated to 0..1 fractions against the visible screen
// element (canvas or img, whichever is active) and POSTed to the input
// endpoints — unchanged from before.
const indexHTML = `<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1, maximum-scale=1, user-scalable=no">
<title>ios remote</title>
<style>
  :root { color-scheme: dark; }
  * { box-sizing: border-box; }
  html, body { margin: 0; height: 100%; background: #111; color: #eee;
    font: 14px/1.4 -apple-system, system-ui, sans-serif; }
  #app { display: flex; flex-direction: column; height: 100%; }
  #stage { flex: 1; display: flex; align-items: center; justify-content: center;
    overflow: hidden; padding: 8px; }
  #canvas, #fallback { max-width: 100%; max-height: 100%; touch-action: none;
    border-radius: 14px; box-shadow: 0 0 0 1px #333, 0 8px 30px rgba(0,0,0,.6);
    cursor: crosshair; user-select: none; -webkit-user-drag: none; }
  #fallback { display: none; }
  #bar { display: flex; flex-wrap: wrap; gap: 6px; align-items: center;
    padding: 8px; background: #1a1a1a; border-top: 1px solid #2a2a2a; }
  button { background: #2a2a2a; color: #eee; border: 1px solid #3a3a3a;
    border-radius: 8px; padding: 8px 12px; font-size: 14px; cursor: pointer; }
  button:hover { background: #333; }
  button:active { background: #444; }
  input[type=text] { flex: 1; min-width: 120px; background: #222; color: #eee;
    border: 1px solid #3a3a3a; border-radius: 8px; padding: 8px 10px; font-size: 14px; }
  #status { margin-left: auto; font-size: 12px; color: #9a9a9a;
    max-width: 40%; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  .sep { width: 1px; height: 24px; background: #333; margin: 0 2px; }
</style>
</head>
<body>
<div id="app">
  <div id="stage">
    <canvas id="canvas" width="390" height="844"></canvas>
    <img id="fallback" src="" alt="device screen" draggable="false">
  </div>
  <div id="bar">
    <button data-btn="home">Home</button>
    <button data-btn="lock">Lock</button>
    <button data-btn="volumeup">Vol +</button>
    <button data-btn="volumedown">Vol -</button>
    <span class="sep"></span>
    <input id="text" type="text" placeholder="type text, Enter to send">
    <button id="send">Send</button>
    <span id="status">connecting…</span>
  </div>
</div>
<script>
(function () {
  var canvas = document.getElementById('canvas');
  var fallback = document.getElementById('fallback');
  var status = document.getElementById('status');
  var textbox = document.getElementById('text');
  var DRAG_THRESHOLD = 0.02; // fraction of the screen; below this = tap
  var runnerNotReady = false;

  function setStatus(s) { status.textContent = s; }

  // activeScreen is whichever element is currently visible; input fractions are
  // measured against its rendered rect so taps map identically for canvas/img.
  function activeScreen() {
    return canvas.style.display === 'none' ? fallback : canvas;
  }

  // Translate a pointer event to a 0..1 fraction of the *rendered* screen,
  // accounting for letterboxing (max-width/height keeps aspect ratio).
  function fractionFromEvent(e) {
    var r = activeScreen().getBoundingClientRect();
    var x = (e.clientX - r.left) / r.width;
    var y = (e.clientY - r.top) / r.height;
    return { x: Math.min(1, Math.max(0, x)), y: Math.min(1, Math.max(0, y)) };
  }

  function post(path, body) {
    return fetch(path, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body)
    }).then(function (res) {
      return res.text().then(function (t) {
        setStatus(res.ok ? (path + ' ok') : (path + ' ' + res.status + ': ' + t.slice(0, 80)));
        return res.ok;
      });
    }).catch(function (err) { setStatus(path + ' error: ' + err); });
  }

  // --- input wiring (identical to before; bound to both screen elements) ---
  var down = null;
  function bindInput(el) {
    el.addEventListener('pointerdown', function (e) {
      e.preventDefault();
      down = fractionFromEvent(e);
    });
    el.addEventListener('pointerup', function (e) {
      if (!down) return;
      e.preventDefault();
      var up = fractionFromEvent(e);
      var dx = up.x - down.x, dy = up.y - down.y;
      if (Math.abs(dx) < DRAG_THRESHOLD && Math.abs(dy) < DRAG_THRESHOLD) {
        post('/tap', { x: down.x, y: down.y });
      } else {
        post('/swipe', { fromX: down.x, fromY: down.y, toX: up.x, toY: up.y });
      }
      down = null;
    });
    el.addEventListener('pointercancel', function () { down = null; });
  }
  bindInput(canvas);
  bindInput(fallback);

  document.querySelectorAll('button[data-btn]').forEach(function (b) {
    b.addEventListener('click', function () { post('/button', { name: b.dataset.btn }); });
  });

  function send() {
    var t = textbox.value;
    if (!t) return;
    post('/type', { text: t }).then(function () { textbox.value = ''; });
  }
  document.getElementById('send').addEventListener('click', send);
  textbox.addEventListener('keydown', function (e) { if (e.key === 'Enter') send(); });

  // --- MJPEG fallback ---
  var usingFallback = false;
  function useFallback(reason) {
    if (usingFallback) return;
    usingFallback = true;
    canvas.style.display = 'none';
    fallback.style.display = 'block';
    setStatus('screen: MJPEG fallback (' + reason + ')');
    fallback.src = '/screen?t=' + Date.now();
    fallback.addEventListener('error', function () {
      setTimeout(function () { fallback.src = '/screen?t=' + Date.now(); }, 1000);
    });
  }

  // --- H.264 Annex-B → WebCodecs → canvas (primary) ---
  // Derive the avc1.PPCCLL codec string from the SPS: bytes after the NAL header
  // are profile_idc, constraint_set flags, level_idc.
  function codecFromSPS(sps) {
    // sps[0] is the NAL header; profile/constraint/level are the next 3 bytes.
    var profile = sps[1], constraint = sps[2], level = sps[3];
    function hex(b) { return ('0' + b.toString(16)).slice(-2); }
    return 'avc1.' + hex(profile) + hex(constraint) + hex(level);
  }

  // Split an Annex-B buffer into NAL payloads (without start codes). Returns the
  // NALs found and the number of bytes consumed (a trailing partial NAL is left
  // for the next chunk).
  function splitNALs(buf) {
    var nals = [];
    var i = 0, n = buf.length;
    // find first start code
    function startAt(p) {
      if (p + 3 < n && buf[p] === 0 && buf[p+1] === 0 && buf[p+2] === 0 && buf[p+3] === 1) return 4;
      if (p + 2 < n && buf[p] === 0 && buf[p+1] === 0 && buf[p+2] === 1) return 3;
      return 0;
    }
    // Position i at the first start code.
    while (i < n && startAt(i) === 0) i++;
    var consumed = i;
    while (i < n) {
      var sc = startAt(i);
      if (sc === 0) { i++; continue; }
      var payloadStart = i + sc;
      // find next start code
      var j = payloadStart;
      while (j < n && startAt(j) === 0) j++;
      if (j >= n) {
        // No next start code yet: this NAL may be incomplete, keep for next time.
        break;
      }
      nals.push(buf.subarray(payloadStart, j));
      consumed = j;
      i = j;
    }
    return { nals: nals, consumed: consumed };
  }

  function startH264() {
    if (!('VideoDecoder' in window)) {
      useFallback('WebCodecs unavailable');
      return;
    }
    var ctx = canvas.getContext('2d');
    var decoder = null;
    var configured = false;
    var sps = null, pps = null;
    var sawKey = false;
    var reconnectTimer = null;

    function scheduleReconnect() {
      if (usingFallback) return;
      if (reconnectTimer) return;
      reconnectTimer = setTimeout(function () { reconnectTimer = null; connect(); }, 800);
    }

    function ensureDecoder() {
      if (configured || !sps || !pps) return;
      var codec = codecFromSPS(sps);
      try {
        decoder = new VideoDecoder({
          output: function (frame) {
            if (canvas.width !== frame.displayWidth || canvas.height !== frame.displayHeight) {
              canvas.width = frame.displayWidth;
              canvas.height = frame.displayHeight;
            }
            ctx.drawImage(frame, 0, 0);
            frame.close();
          },
          error: function (e) { useFallback('decoder error: ' + e.message); }
        });
        decoder.configure({ codec: codec, optimizeForLatency: true });
        configured = true;
        if (!runnerNotReady) setStatus('screen: H.264 (' + codec + ')');
      } catch (e) {
        useFallback('configure failed: ' + e.message);
      }
    }

    // Wrap one or more NALs (with 4-byte start codes) into an EncodedVideoChunk.
    function wrapAnnexB(nals) {
      var len = 0, k;
      for (k = 0; k < nals.length; k++) len += 4 + nals[k].length;
      var out = new Uint8Array(len);
      var off = 0;
      for (k = 0; k < nals.length; k++) {
        out[off] = 0; out[off+1] = 0; out[off+2] = 0; out[off+3] = 1;
        out.set(nals[k], off + 4);
        off += 4 + nals[k].length;
      }
      return out;
    }

    var ts = 0;
    function feed(nals) {
      // Collect a set of NALs into one access unit and decode. SPS/PPS are cached
      // for the decoder config and prepended to the first IDR.
      var au = [];
      for (var k = 0; k < nals.length; k++) {
        var nal = nals[k];
        if (nal.length === 0) continue;
        var type = nal[0] & 0x1f;
        if (type === 7) { sps = nal; ensureDecoder(); continue; }
        if (type === 8) { pps = nal; ensureDecoder(); continue; }
        au.push(nal);
        var isKey = (type === 5);
        if (isKey) sawKey = true;
        if (!configured || !sawKey || !decoder) { au = []; continue; }
        var data = isKey ? wrapAnnexB([sps, pps, nal]) : wrapAnnexB([nal]);
        try {
          decoder.decode(new EncodedVideoChunk({
            type: isKey ? 'key' : 'delta',
            timestamp: ts,
            data: data
          }));
          ts += 33333; // ~30fps nominal; timestamps only need to be monotonic
        } catch (e) {
          useFallback('decode failed: ' + e.message);
          return;
        }
        au = [];
      }
    }

    function connect() {
      if (usingFallback) return;
      fetch('/video.h264', { cache: 'no-store' }).then(function (res) {
        if (!res.ok || !res.body) { scheduleReconnect(); return; }
        var reader = res.body.getReader();
        var pending = new Uint8Array(0);
        function pump() {
          reader.read().then(function (r) {
            if (r.done) { scheduleReconnect(); return; }
            // append to pending
            var merged = new Uint8Array(pending.length + r.value.length);
            merged.set(pending, 0);
            merged.set(r.value, pending.length);
            var split = splitNALs(merged);
            if (split.nals.length) feed(split.nals);
            pending = merged.subarray(split.consumed);
            // keep pending as its own buffer so it doesn't retain the big merged one
            pending = pending.slice();
            pump();
          }).catch(function () { scheduleReconnect(); });
        }
        pump();
      }).catch(function () { scheduleReconnect(); });
    }

    connect();
  }

  // Poll the input-runner lifecycle so the user sees when the runner (which
  // backs BOTH screen and input) is (re)starting. /status returns 503 until the
  // runner is ready; the screen proxy reconnects on its own in lockstep.
  function pollStatus() {
    fetch('/status').then(function (res) {
      return res.json().then(function (j) {
        runnerNotReady = j.runnerState && j.runnerState !== 'ready';
        if (runnerNotReady) setStatus('runner ' + j.runnerState + '… (reconnecting)');
      });
    }).catch(function () {}).then(function () {
      setTimeout(pollStatus, 2000);
    });
  }

  startH264();
  pollStatus();
})();
</script>
</body>
</html>
`
