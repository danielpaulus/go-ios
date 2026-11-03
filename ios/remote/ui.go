package remote

// indexHTML is the entire, self-contained web UI. No external/CDN assets: the
// HTML, CSS and JS are inlined so `ios remote` serves it standalone.
//
// The screen is rendered two ways, chosen automatically:
//   - Primary ("h264/canvas"): the DeviceKit runner's hardware H.264 (GET
//     /video.h264, an Annex-B elementary stream) is fetched as a stream, parsed
//     into NAL units, GROUPED INTO ACCESS UNITS (all NALs for one picture), and
//     decoded with the WebCodecs VideoDecoder. Decode is decoupled from paint: a
//     requestAnimationFrame loop draws only the freshest decoded VideoFrame to
//     the <canvas> (older undrawn frames are dropped), which is what keeps the
//     mirror smooth and low-latency.
//   - Fallback ("mjpeg/img"): if WebCodecs is unavailable or the decoder errors,
//     an <img src="/screen"> shows the runner's MJPEG instead. The switch is
//     automatic.
//
// A diagnostic HUD (toggle with the 'd' key) overlays the current mode, the
// measured rendered fps, decoder state + decodeQueueSize, dropped-frame count,
// and the last decoder error — so a maintainer can tell at a glance whether the
// H.264 path is live or it fell back, and where any stall is.
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
    overflow: hidden; padding: 8px; position: relative; }
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
  /* Diagnostic HUD: fixed top-left overlay, toggled with 'd'. */
  #hud { position: absolute; top: 14px; left: 14px; z-index: 10;
    background: rgba(0,0,0,.72); color: #d7ffd7; border: 1px solid #2f5f2f;
    border-radius: 8px; padding: 8px 10px; font: 11px/1.5 ui-monospace, Menlo, Consolas, monospace;
    white-space: pre; pointer-events: none; min-width: 210px; }
  #hud .err { color: #ff9a9a; }
  #hud .hint { color: #7a7a7a; }
  #hud.hidden { display: none; }
</style>
</head>
<body>
<div id="app">
  <div id="stage">
    <canvas id="canvas" width="390" height="844"></canvas>
    <img id="fallback" src="" alt="device screen" draggable="false">
    <div id="hud" class="hidden"></div>
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
  var hud = document.getElementById('hud');
  var DRAG_THRESHOLD = 0.02; // fraction of the screen; below this = tap
  var runnerNotReady = false;

  function setStatus(s) { status.textContent = s; }

  // --- diagnostics ---------------------------------------------------------
  // Shared HUD state, updated by the decode/paint paths and rendered on a timer.
  var diag = {
    mode: 'starting',        // 'h264/canvas' | 'mjpeg/img fallback'
    codec: '',
    decoderState: '-',       // VideoDecoder.state
    queue: 0,                // decoder.decodeQueueSize
    dropped: 0,              // frames decoded but never painted (superseded)
    auDropped: 0,            // access units dropped for backpressure
    lastError: '',
    painted: 0               // rAF paints since last fps sample
  };
  var renderedFps = 0;
  var hudVisible = false;

  function renderHUD() {
    if (!hudVisible) return;
    var lines = [
      'MODE     ' + diag.mode + (diag.codec ? ' (' + diag.codec + ')' : ''),
      'RENDER   ' + renderedFps.toFixed(1) + ' fps',
      'DECODER  ' + diag.decoderState + '  queue=' + diag.queue,
      'DROPPED  ' + diag.dropped + ' late, ' + diag.auDropped + ' AU (backpressure)'
    ];
    hud.textContent = lines.join('\n');
    if (diag.lastError) {
      var e = document.createElement('span');
      e.className = 'err';
      e.textContent = '\nERROR    ' + diag.lastError;
      hud.appendChild(e);
    }
    var hint = document.createElement('span');
    hint.className = 'hint';
    hint.textContent = "\n('d' hides)";
    hud.appendChild(hint);
  }
  // fps = paints per second, sampled once a second.
  setInterval(function () {
    renderedFps = diag.painted;
    diag.painted = 0;
    renderHUD();
  }, 1000);
  document.addEventListener('keydown', function (e) {
    if (e.key === 'd' || e.key === 'D') {
      hudVisible = !hudVisible;
      hud.classList.toggle('hidden', !hudVisible);
      renderHUD();
    }
  });

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
    diag.mode = 'mjpeg/img fallback';
    diag.codec = '';
    diag.lastError = reason;
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
    // is there a start code at p? returns its length (3 or 4) or 0.
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
    diag.mode = 'h264/canvas';
    var ctx = canvas.getContext('2d');
    var decoder = null;
    var configured = false;
    var sps = null, pps = null;
    var codec = '';
    var sawKey = false;
    var reconnectTimer = null;

    // 1-slot holder: only the freshest decoded frame is kept for the next paint;
    // any older undrawn frame is closed (dropped). This decouples decode from
    // paint and is the core smoothness fix.
    var latestFrame = null;
    var rafPending = false;

    function paintLoop() {
      rafPending = false;
      if (latestFrame) {
        var frame = latestFrame;
        latestFrame = null;
        if (canvas.width !== frame.displayWidth || canvas.height !== frame.displayHeight) {
          canvas.width = frame.displayWidth;
          canvas.height = frame.displayHeight;
        }
        try { ctx.drawImage(frame, 0, 0); } catch (e) { /* frame may be closed on teardown */ }
        frame.close();
        diag.painted++;
      }
      // Keep the loop alive so a newly-arrived frame is drawn next vsync.
      schedulePaint();
    }
    function schedulePaint() {
      if (rafPending) return;
      rafPending = true;
      requestAnimationFrame(paintLoop);
    }

    function ensureDecoder() {
      if (configured || !sps || !pps) return;
      codec = codecFromSPS(sps);
      try {
        decoder = new VideoDecoder({
          output: function (frame) {
            // Never paint here: stash the freshest frame and let rAF draw it.
            if (latestFrame) { latestFrame.close(); diag.dropped++; }
            latestFrame = frame;
            schedulePaint();
          },
          error: function (e) {
            diag.lastError = e.message;
            useFallback('decoder error: ' + e.message);
          }
        });
        // annexb: SPS/PPS travel in-band, so no description is needed.
        decoder.configure({ codec: codec, optimizeForLatency: true });
        configured = true;
        diag.codec = codec;
        if (!runnerNotReady) setStatus('screen: H.264 (' + codec + ')');
      } catch (e) {
        diag.lastError = e.message;
        useFallback('configure failed: ' + e.message);
      }
    }

    // Wrap one or more NALs (with 4-byte start codes) into an Annex-B buffer.
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

    // Access-unit assembler: NALs stream in; a new AU starts at each VCL NAL
    // (type 1 or 5). Leading non-VCL NALs (SPS/PPS/SEI/AUD) accumulate and are
    // emitted together with the following VCL slice as ONE EncodedVideoChunk.
    var pendingAU = [];       // NALs accumulated for the current (not-yet-VCL-terminated) AU
    var pendingHasKey = false;
    var ts = 0;

    function isVCL(type) { return type === 1 || type === 5; }

    function emitAU(nals, isKey) {
      if (!configured || !decoder) return;
      // The first chunk fed MUST be a keyframe; drop deltas until we've seen one.
      if (!sawKey && !isKey) return;
      if (isKey) sawKey = true;

      // Backpressure: if the decoder is falling behind, drop delta AUs until the
      // next keyframe rather than queueing (low latency > completeness for a live
      // mirror).
      diag.queue = decoder.decodeQueueSize;
      if (!isKey && diag.queue > 2) { diag.auDropped++; return; }

      // A keyframe AU must carry SPS+PPS in-band (annexb). Non-key AUs go as-is.
      var data;
      if (isKey) {
        var withParamSets = [sps, pps];
        for (var m = 0; m < nals.length; m++) withParamSets.push(nals[m]);
        data = wrapAnnexB(withParamSets);
      } else {
        data = wrapAnnexB(nals);
      }
      try {
        decoder.decode(new EncodedVideoChunk({
          type: isKey ? 'key' : 'delta',
          timestamp: ts,
          data: data
        }));
        ts += 16666; // ~60fps nominal; timestamps only need to be monotonic
        diag.queue = decoder.decodeQueueSize;
      } catch (e) {
        diag.lastError = e.message;
        useFallback('decode failed: ' + e.message);
      }
    }

    function flushPendingAU() {
      if (pendingAU.length === 0) return;
      emitAU(pendingAU, pendingHasKey);
      pendingAU = [];
      pendingHasKey = false;
    }

    function feed(nals) {
      for (var k = 0; k < nals.length; k++) {
        var nal = nals[k];
        if (nal.length === 0) continue;
        var type = nal[0] & 0x1f;
        // Cache parameter sets for the decoder config; they are (re)inserted into
        // key AUs, so they are not accumulated into pendingAU here.
        if (type === 7) {
          sps = nal; ensureDecoder();
          continue;
        }
        if (type === 8) {
          pps = nal; ensureDecoder();
          continue;
        }
        if (isVCL(type)) {
          // The VCL slice completes the current access unit. Any non-VCL NALs
          // accumulated (SEI/AUD) belong with it.
          pendingAU.push(nal);
          if (type === 5) pendingHasKey = true;
          flushPendingAU();
        } else {
          // Leading non-VCL NAL (SEI type 6, AUD type 9, …): hold for the AU.
          pendingAU.push(nal);
        }
      }
    }

    function scheduleReconnect() {
      if (usingFallback) return;
      if (reconnectTimer) return;
      // A dropped stream means the runner restarted; the next AU will be a fresh
      // keyframe, so require a keyframe again before feeding deltas.
      sawKey = false;
      pendingAU = [];
      pendingHasKey = false;
      reconnectTimer = setTimeout(function () { reconnectTimer = null; connect(); }, 800);
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
            var merged = new Uint8Array(pending.length + r.value.length);
            merged.set(pending, 0);
            merged.set(r.value, pending.length);
            var split = splitNALs(merged);
            if (split.nals.length) feed(split.nals);
            // keep the leftover as its own buffer so it doesn't retain merged.
            pending = merged.subarray(split.consumed).slice();
            pump();
          }).catch(function () { scheduleReconnect(); });
        }
        pump();
      }).catch(function () { scheduleReconnect(); });
    }

    // Keep the decoder-state line in the HUD fresh even when idle.
    setInterval(function () {
      if (decoder) { diag.decoderState = decoder.state; diag.queue = decoder.decodeQueueSize; }
    }, 250);

    schedulePaint();
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
