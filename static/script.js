document.body.style.margin = '0';
document.body.style.overflow = 'hidden';

var canvas = document.getElementById('canvas');
var ctx = canvas.getContext("2d", {"alpha": false});

function runCommand(c) {
  var fn = c.shift();
  window[fn].apply(undefined, c);
}

function measureText(text) {
  var w = ctx.measureText(text).width;
  syncSocket.send(JSON.stringify({
    "type": "TextWidth",
    "width": w,
  }));
}

function transform(a, b, c, d, e, f) { ctx.transform(a, b, c, d, e, f); }
function translate(x, y) { ctx.translate(x, y); }
function fillText(text, x, y) { ctx.fillText(text, x, y); }
function fillRect(x, y, w, h) { ctx.fillRect(x, y, w, h); }
function rect(x, y, w, h) { ctx.rect(x, y, w, h); }
function arc(x, y, r, alpha, beta, anticlockwise) { ctx.arc(x, y, r, alpha, beta, anticlockwise); }
function ellipse(x, y, rx, ry, rotation, alpha, beta, anticlockwise) { ctx.ellipse(x, y, rx, ry, rotation, alpha, beta, anticlockwise); }
function moveTo(x, y) { ctx.moveTo(x, y); }
function lineTo(x, y) { ctx.lineTo(x, y); }
function setLineDash(val) { ctx.setLineDash(val); }
function rotate(val) { ctx.rotate(val); }
function scale(val) { ctx.scale(val, val); }

function fillStyle(val) { ctx.fillStyle = val; }
function textAlign(val) { ctx.textAlign = val; }
function textBaseline(val) { ctx.textBaseline = val; }
function lineWidth(val) { ctx.lineWidth = val; }
function strokeStyle(val) { ctx.strokeStyle = val; }
function font(val) { ctx.font = val; }

function save() { ctx.save(); }
function restore() { ctx.restore(); }
function beginPath() { ctx.beginPath(); }
function closePath() { ctx.closePath(); }
function fill() { ctx.fill(); }
function stroke() { ctx.stroke(); }
function clip() { ctx.clip(); }

var eventSocket = undefined;
var syncSocket = undefined;

var lastData;
var lastMsg;
function SyncMessage(e) {
  var msg = JSON.parse(e.data);
  lastData = e.data;
  lastMsg = msg;
  if (typeof msg[0] === 'string') {
    runCommand(msg);
  } else {
    msg.forEach(runCommand);
    syncSocket.send(JSON.stringify({
      "type": "RenderDone",
      "time": performance.now()
    }));
    RequestRendering();
  }
};

function RequestRendering() {
  requestAnimationFrame(function() {
    if (eventSocket.readyState != 1) return;
    eventSocket.send(JSON.stringify({
      "type": "RenderReady",
      "time": performance.now()
    }));
  });
}

function Reconnect() {
  canvas.width = innerWidth;
  canvas.height = innerHeight;
  ctx.font = '20px Iosevka';
  ctx.fillStyle = "#ddd";
  ctx.fillRect(0,0,canvas.width, canvas.height);
  ctx.fillStyle = "#000";
  ctx.textAlign = "center";
  ctx.fillText("Stopped", canvas.width/2, canvas.height/2);
  setTimeout(Connect, 1000);
};

function Connect() {
  if (eventSocket) eventSocket.close();
  eventSocket = new WebSocket("ws://localhost:8000/events");
  eventSocket.onopen = EventOpen;
  eventSocket.onerror = Reconnect;
  if (syncSocket) syncSocket.close();
  syncSocket = new WebSocket("ws://localhost:8000/sync");
  syncSocket.onmessage = SyncMessage;
};

function WindowResize(e) {
  eventSocket.send(JSON.stringify({
    "type": "Size",
    "width": innerWidth,
    "height": innerHeight
  }));
  canvas.width = innerWidth;
  canvas.height = innerHeight;
  ctx.font = '20px Iosevka';
};

function CopyProperties(type, properties) {
  return function(e) {
	  var o = { "type": type };
	  for (var key in properties) {
	    o[key] = e[properties[key]];
	  }
	  eventSocket.send(JSON.stringify(o));
	  e.preventDefault();
	  return false;
  };
}

function Touch(type) {
  var properties = {"x": "clientX", "y": "clientY", "id": "identifier"};
  return function(e) {
	  var o = { "type": type, "changed": [] };
    for (var i = 0; i < e.changedTouches.length; ++i) {
      var t = {};
	    for (var key in properties) {
	      t[key] = e.changedTouches[i][properties[key]];
	    }
      o.changed.push(t);
    }
	  eventSocket.send(JSON.stringify(o));
	  e.preventDefault();
	  return false;
  };
}

function Ignore(e) {
  e.preventDefault();
  return false;
}

var binds = {
  "mousedown":   CopyProperties("MouseDown", {"x": "clientX", "y": "clientY", "button": "button"}),
  "mousemove":   CopyProperties("MouseMove", {"x": "clientX", "y": "clientY"}),
  "mouseup":     CopyProperties("MouseUp", {"x": "clientX", "y": "clientY", "button": "button"}),
  "wheel":       CopyProperties("Wheel", {"x": "deltaX",  "y": "deltaY"}),
  "keydown":     CopyProperties("KeyDown", {"code": "code", "key": "key"}),
  "keyup":       CopyProperties("KeyUp", {"code": "code", "key": "key"}),
  "touchstart":  Touch("TouchStart"),
  "touchend":    Touch("TouchEnd"),
  "touchmove":   Touch("TouchMove"),
  "contextmenu": Ignore,
  "resize":      WindowResize,
};

function EventOpen(e) {
  eventSocket.onerror = undefined;
  for (var eventtype in binds) {  
    window.addEventListener(eventtype, binds[eventtype], false);
  }
  WindowResize();
  RequestRendering();
  eventSocket.onclose = EventClose;
};

function EventClose() {
  for (var eventtype in binds) {  
    window.removeEventListener(eventtype, binds[eventtype], false);
  }
  Reconnect();
};

Connect();
