$(document).ready(function() {
    console.log("Ready!");
});

function beginGame(p2Nick) {
    $('#p2-nick').text(p2Nick);
    $('#play-container').slideToggle(300, function() {
        $('#game-container').fadeIn();
    });
}

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
        var msg     = JSON.parse(event.data);
        var action  = msg['Action'];
        var payload = msg['Payload'];
        if (action == "found_game") {
            beginGame(payload['match']);
        }
    };
});
