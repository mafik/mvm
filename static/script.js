var elems = [];

document.body.style.margin = '0';
document.body.style.overflow = 'hidden';

var canvas = document.getElementById('canvas');
var ctx = canvas.getContext("2d");

function drawWidget(w) {
  ctx.save();
  ctx.translate(w.Pos.X, w.Pos.Y);
  ctx.scale(w.Scale, w.Scale);
  window[w.Type](w.Value);
  ctx.restore();
}

function draw() {
  ctx.clearRect(0, 0, canvas.width, canvas.height);
  ctx.fillStyle = '#eee';
  ctx.fillRect(0, 0, canvas.width, canvas.height);
  ctx.fillStyle = '#000';
  elems.forEach(drawWidget);
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

var margin = 10;

function button(value) {
  if (value.Rect) {
    rect(value.Rect);
    ctx.clip();
  }
  if (value.Text) {
    if (value.Rect) {
      if (value.Text.Align != "center") {
	ctx.translate(-value.Rect.X/2 + margin, 0);
      }
      ctx.translate(0, margin + 20 - value.Rect.Y/2);
    }
    text(value.Text);
  }
}

function text(value) {
  ctx.textAlign = value.Align || 'left';
  ctx.fillStyle = value.Color || '#000';
  var cars = value.Text.split("\n");
  for (var i = 0; i < cars.length; ++i) {
    ctx.fillText(cars[i], 0, 20*i);
  }
  if (value.Caret) {
    var endX = ctx.measureText(cars[cars.length - 1]).width;
    var endY = 20 * (cars.length - 1);
    ctx.fillRect(endX, endY + 5, 2, - 25);
  }
}

function rect(value) {
  ctx.fillStyle = value.Color || '#fff';
  ctx.beginPath();
  ctx.rect(-value.X/2, -value.Y/2, value.X, value.Y);
  ctx.closePath();
  ctx.fill();
}

function arrow(value) {
  var A = 13;
  ctx.beginPath();
  ctx.moveTo(0, 0);
  ctx.arc(0, 0, A, Math.PI*5/6, Math.PI*7/6);
  ctx.closePath();
  ctx.fillStyle = '#000';
  ctx.fill();
}


function hourglass(value) {
  var LW = 1.5; // line width
  var W = 8; // width
  var H = 10; // top height
  var H2 = H - 2; // angled height
  var h = 1.5; // gap height
  var F = 4 + Math.cos((+new Date) / 2000); // fill
  ctx.translate(-W-LW/2, -H-LW/2);
  ctx.beginPath();
  ctx.moveTo(-W, -H);
  ctx.lineTo(W, -H);
  ctx.lineTo(W, -H2);
  ctx.lineTo(2, -h);
  ctx.lineTo(2, h);
  ctx.lineTo(W, H2);
  ctx.lineTo(W, H);
  ctx.lineTo(-W, H);
  ctx.lineTo(-W, H2);
  ctx.lineTo(-2, h);
  ctx.lineTo(-2, -h);
  ctx.lineTo(-W, -H2);
  ctx.closePath();
  ctx.lineWidth = 1.5;
  //ctx.lineJoin = 'round';
  ctx.strokeStyle = value.Color;
  ctx.stroke();

  ctx.translate(0, -LW/2 * Math.sqrt(2) -h);
  ctx.beginPath();
  ctx.moveTo(0, 0);
  ctx.lineTo(-F, -F);
  ctx.lineTo(F, -F);
  ctx.closePath();
  ctx.fillStyle = value.Color;
  ctx.fill();
}

function line(value) {
  ctx.lineWidth = value.Width || 2;
  if (value.Dash) {
    ctx.setLineDash(value.Dash);
  }

  ctx.rotate(Math.atan2(value.Y, value.X));
  var length = Math.hypot(value.Y, value.X);
  var end = length;
  var start = 0;
  if (value.Start) {
    start += 2;
  }
  if (value.End) {
    end -= 2;
  }

  ctx.beginPath();
  ctx.moveTo(start, 0);
  ctx.lineTo(end, 0);
  ctx.stroke();

  if (value.Middle) {
    ctx.save();
    ctx.translate(length / 2, 0);
    if (value.X <= 0) {
      ctx.rotate(Math.PI);
    }
    ctx.translate(0,-1);
    drawWidget(value.Middle);
    ctx.restore();
  }

  if (value.Start) {
    ctx.rotate(Math.PI);
    drawWidget(value.Start);
    ctx.rotate(Math.PI);
  }

  if (value.End) {
    ctx.translate(Math.hypot(value.X, value.Y), 0);
    drawWidget(value.End);
  }
}

function circle(value) {
  ctx.beginPath();
  ctx.arc(0, 0, value.R, 0, 2*Math.PI, false);
  ctx.closePath();
  ctx.fillStyle = value.Color || "#f00";
  ctx.fill();
}

var socket = new WebSocket("ws://localhost:8000/events");
var binds = [
  {"html": "onmousedown", "mvm": "MouseDown", "x": "clientX", "y": "clientY", "button": "button"},
  {"html": "onmousemove", "mvm": "MouseMove", "x": "clientX", "y": "clientY"},
  {"html": "onmouseup",   "mvm": "MouseUp",   "x": "clientX", "y": "clientY", "button": "button"},
  {"html": "onwheel",     "mvm": "Wheel",     "x": "deltaX",  "y": "deltaY"},
  {"html": "onkeydown",   "mvm": "KeyDown",   "code": "code", "key": "key"},
  {"html": "onkeyup",     "mvm": "KeyUp",     "code": "code", "key": "key"},
  {"html": "oncontextmenu", "mvm": "ContextMenu"},
];

function SocketMessage(e) {
  elems = JSON.parse(e.data);
  draw();
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
    var o = { "type": bind.mvm };
    for (var key in bind) {
      if (key == "html" || key == "mvm") continue;
      o[key] = e[bind[key]];
    }
    socket.send(JSON.stringify(o));
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
