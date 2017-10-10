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
  socket.send(JSON.stringify({
    "type": "TextWidth",
    "width": w,
  }));
}

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

var socket = undefined;
var binds = [
  {"html": "onmousedown", "mvm": "MouseDown", "x": "clientX", "y": "clientY", "button": "button"},
  {"html": "onmousemove", "mvm": "MouseMove", "x": "clientX", "y": "clientY"},
  {"html": "onmouseup",   "mvm": "MouseUp",   "x": "clientX", "y": "clientY", "button": "button"},
  {"html": "onwheel",     "mvm": "Wheel",     "x": "deltaX",  "y": "deltaY"},
  {"html": "onkeydown",   "mvm": "KeyDown",   "code": "code", "key": "key"},
  {"html": "onkeyup",     "mvm": "KeyUp",     "code": "code", "key": "key"},
  {"html": "oncontextmenu"},
];

function SocketMessage(e) {
  var msg = JSON.parse(e.data);
  if (typeof msg[0] === 'string') {
    runCommand(msg);
  } else {
    msg.forEach(runCommand);
    socket.send(JSON.stringify({
      "type": "RenderDone",
      "time": performance.now()
    }));
    requestAnimationFrame(function() {
      if (socket.readyState != 1) return;
      socket.send(JSON.stringify({
	"type": "RenderReady",
	"time": performance.now()
      }));
    });
  }
};

function Reconnect() {
  ctx.clearRect(0,0,canvas.width, canvas.height);
  ctx.save();
  ctx.textAlign = "center";
  ctx.translate(canvas.width/2, canvas.height/2);
  ctx.fillText("Stopped", 0, 0);
  ctx.restore();
  setTimeout(Connect, 1000);
};

function Connect() {
  socket = new WebSocket("ws://localhost:8000/events");
  socket.onmessage = SocketMessage;
  socket.onopen = SocketOpen;
  socket.onerror = Reconnect;
};

function SocketClose() {
  window.onresize = undefined;
  binds.forEach(function(bind) { window[bind.html] = undefined; });
  Reconnect();
};

function WindowResize(e) {
  socket.send(JSON.stringify({
    "type": "Size",
    "width": innerWidth,
    "height": innerHeight
  }));
  canvas.width = innerWidth;
  canvas.height = innerHeight;
  ctx.font = '20px Iosevka';
  draw();
};

function Bind(bind) {
  window[bind.html] = function(e) {
    if (typeof bind.mvm !== 'undefined') {
      var o = { "type": bind.mvm };
      for (var key in bind) {
	if (key == "html" || key == "mvm") continue;
	o[key] = e[bind[key]];
      }
      socket.send(JSON.stringify(o));
    }
    e.preventDefault();
    return true;
  }
};

function SocketOpen(e) {
  socket.onerror = undefined;
  window.onresize = WindowResize;
  window.onresize();
  binds.forEach(Bind);
  socket.onclose = SocketClose;
};

Connect();
