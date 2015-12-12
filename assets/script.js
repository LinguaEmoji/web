$(document).ready(function() {
    console.log("Ready!");
});

$('#play-button').click(function() {
    var nick = $('#nick-input').val().trim();
    if (nick == "") {
        alert("Nickname cannot be blank!");
        return;
    }
    $('#play-button').addClass('disabled');
    $('#play-container .input').addClass('disabled');

    var ws = new WebSocket("ws://localhost:3000/websocket");
    ws.onerror = function() {
        alert("Something bad happened!");
    };
    ws.onopen = function() {
        console.log("Connected to websocket");

        console.log("Sending nick");
        ws.send(nick);

        $('#load-status h2').text("Joined game queue");
        $('#load-status').slideToggle();

    };
    ws.onmessage = function(event) {
        console.log("Got message", event.data);
    };
});
