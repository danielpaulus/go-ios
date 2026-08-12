package remote

// indexHTML is the entire, self-contained web UI. No external/CDN assets: the
// HTML, CSS and JS are inlined so `ios remote` serves it standalone. The screen
// is an <img src="/screen"> MJPEG stream; clicks/drags are translated to 0..1
// fractions and POSTed to the input endpoints.
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
  #screen { max-width: 100%; max-height: 100%; touch-action: none;
    border-radius: 14px; box-shadow: 0 0 0 1px #333, 0 8px 30px rgba(0,0,0,.6);
    cursor: crosshair; user-select: none; -webkit-user-drag: none; }
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
    <img id="screen" src="/screen" alt="device screen" draggable="false">
  </div>
  <div id="bar">
    <button data-btn="home">Home</button>
    <button data-btn="lock">Lock</button>
    <button data-btn="volumeup">Vol +</button>
    <button data-btn="volumedown">Vol -</button>
    <span class="sep"></span>
    <input id="text" type="text" placeholder="type text, Enter to send">
    <button id="send">Send</button>
    <span id="status">tap the screen</span>
  </div>
</div>
<script>
(function () {
  var img = document.getElementById('screen');
  var status = document.getElementById('status');
  var textbox = document.getElementById('text');
  var DRAG_THRESHOLD = 0.02; // fraction of the screen; below this = tap

  function setStatus(s) { status.textContent = s; }

  // Translate a pointer event to a 0..1 fraction of the *rendered* image,
  // accounting for letterboxing (object-fit style max-width/height).
  function fractionFromEvent(e) {
    var r = img.getBoundingClientRect();
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

  var down = null;
  img.addEventListener('pointerdown', function (e) {
    e.preventDefault();
    down = fractionFromEvent(e);
  });
  img.addEventListener('pointerup', function (e) {
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
  img.addEventListener('pointercancel', function () { down = null; });

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

  // If the MJPEG stream drops, nudge it back to life.
  img.addEventListener('error', function () {
    setStatus('screen stream error, retrying…');
    setTimeout(function () { img.src = '/screen?t=' + Date.now(); }, 1000);
  });
})();
</script>
</body>
</html>
`
