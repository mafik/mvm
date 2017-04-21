var elems = [];

document.body.style.margin = '0';
document.body.style.overflow = 'hidden';

var canvas = document.getElementById('canvas');
var ctx = canvas.getContext("2d");

function draw() {
    ctx.clearRect(0, 0, canvas.width, canvas.height);
    for (var i = 0; i < elems.length; ++i) {
        var elem = elems[i];
        ctx.save();
        ctx.translate(elem.x, elem.y);
        window[elem.type](elem.value);
        ctx.restore();
    }
    socket.send(JSON.stringify({
        "type": "RenderDone",
        "time": performance.now()
    }));
    requestAnimationFrame(function() {
        socket.send(JSON.stringify({
            "type": "RenderReady",
            "time": performance.now()
        }));
    });
}

function text(value) {
    ctx.fillText(value, 0, 0);
}

var socket = new WebSocket("ws://localhost:8000/events");
socket.onmessage = function(e) {
    elems = JSON.parse(e.data);
    draw();
};
socket.onopen = function(e) {
    window.onresize = function (e) {
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
    window.onresize();

    [
        {"html": "onmousedown", "mvm": "TouchStart"},
        {"html": "onmousemove", "mvm": "TouchMove"},
        {"html": "onmouseup", "mvm": "TouchEnd"}
    ].forEach(function(bind) {
        window[bind.html] = function(e) {
            socket.send(JSON.stringify({
                "type": bind.mvm,
                "id": 0,
                "x": e.clientX,
                "y": e.clientY
            }));
        }
    });
    socket.onclose = function() {
        console.log("reloading!");
        setTimeout("location.reload()", 1000);
    };
};