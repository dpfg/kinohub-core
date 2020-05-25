(function () {
  var landing = document.getElementById("landing");
  var playerEl = document.getElementById("player");

  var showPlayer = function () {
    playerEl.style.display = "block";

    var player = videojs("player");
    player.ready(function () {
      player.play();
      player.requestFullscreen();
    });
  };

  landing.onclick = function () {
    showPlayer();
  };

  var player = videojs("player");
  player.on("fullscreenchange", function () {
    if (player.isFullscreen()) {
      landing.style.display = "none";
    } else {
      landing.style.display = "block";
      playerEl.style.display = "none";
    }
  });

  var ws = new WebSocket("ws://localhost:8090/ui/pws/?pid=LG-TV");
  ws.onmessage = function (event) {
    var msg = JSON.parse(event.data);
    switch (msg["type_id"]) {
      case "play":
        player.play();
        showPlayer();
        break;
      case "pause":
        player.pause();
        break;
      case "set-source":
        player.src([{ src: msg["data"]["url"] }]);
        player.play();
        break;
      default:
        console.log("Unknown message");
    }
  };
})();