(function () {
  var canvas = document.getElementById("firefly-canvas");
  if (!canvas) return;
  var ctx = canvas.getContext("2d");

  function resize() {
    canvas.width = window.innerWidth;
    canvas.height = window.innerHeight;
  }
  resize();
  window.addEventListener("resize", resize);

  var ZONE = { xMin: 0.44, xMax: 0.96, yMin: 0.02, yMax: 0.50 };

  var flies = [
    { x: 0.56, y: 0.17, pulseOffset: 0.0, speed: 0.00008, angle: 1.1, turnRate: 0.004 },
    { x: 0.73, y: 0.27, pulseOffset: 3.8, speed: 0.00006, angle: 2.7, turnRate: 0.003 },
    { x: 0.63, y: 0.11, pulseOffset: 7.1, speed: 0.00010, angle: 0.4, turnRate: 0.005 },
  ];

  var lastTime = null;

  function drawFly(fly, now) {
    var W = canvas.width, H = canvas.height;
    var px = fly.x * W, py = fly.y * H;

    var raw = Math.sin(now * 0.0007 + fly.pulseOffset);
    var pulse = Math.pow(Math.max(0, raw), 3);
    if (pulse < 0.01) return;

    var alpha = pulse * 0.85;
    var glowR = 2.5 + pulse * 7;

    var grad = ctx.createRadialGradient(px, py, 0, px, py, glowR * 3);
    grad.addColorStop(0,   "rgba(195, 235, 85, " + (alpha * 0.65) + ")");
    grad.addColorStop(0.4, "rgba(165, 215, 65, " + (alpha * 0.28) + ")");
    grad.addColorStop(1,   "rgba(140, 200, 50, 0)");

    ctx.beginPath();
    ctx.arc(px, py, glowR * 3, 0, Math.PI * 2);
    ctx.fillStyle = grad;
    ctx.fill();

    ctx.beginPath();
    ctx.arc(px, py, glowR * 0.35, 0, Math.PI * 2);
    ctx.fillStyle = "rgba(225, 252, 145, " + alpha + ")";
    ctx.fill();
  }

  function moveFly(fly, dt) {
    fly.angle += (Math.random() - 0.5) * fly.turnRate;

    var margin = 0.03;
    if (fly.x < ZONE.xMin + margin) fly.angle += 0.05;
    if (fly.x > ZONE.xMax - margin) fly.angle -= 0.05;
    if (fly.y < ZONE.yMin + margin) fly.angle += 0.05;
    if (fly.y > ZONE.yMax - margin) fly.angle -= 0.05;

    fly.x += Math.cos(fly.angle) * fly.speed * dt;
    fly.y += Math.sin(fly.angle) * fly.speed * dt;

    fly.x = Math.max(ZONE.xMin, Math.min(ZONE.xMax, fly.x));
    fly.y = Math.max(ZONE.yMin, Math.min(ZONE.yMax, fly.y));
  }

  function frame(now) {
    if (!lastTime) lastTime = now;
    var dt = Math.min(now - lastTime, 50);
    lastTime = now;

    ctx.clearRect(0, 0, canvas.width, canvas.height);
    flies.forEach(function (fly) { moveFly(fly, dt); drawFly(fly, now); });
    requestAnimationFrame(frame);
  }
  requestAnimationFrame(frame);
})();
