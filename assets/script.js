$('#emoji-input').ready(function() {
    $('#emoji-input').emojiPicker({
        button: false,
        height: 200
    });
});

function beginGame(p2Nick) {
    $('.p2-nick').text(p2Nick);
    $('#play-container').slideToggle(300, function() {
        $('#game-container').fadeIn();
    });
}

function playTurn(phrase) {
    $('#wait-container').hide();
    $('#answer-container').hide();

    $('#emoji-input').val("");

    $('#starting-phrase').text(phrase);
    $('#emojify-container').fadeIn(400, function() {
        $('#emoji-input').emojiPicker('toggle'); 
        var style = $('.emojiPicker')[0].style;
        style.position = "relative";
        style.top = "";
        style.left = "";
        style.zoom = "1.5";
        style.display = "block";
    });
}

function playTurnEnglish(clue) {
    $('#wait-container').hide();
    $('#answer-container').hide();
    $('#englishify-container').fadeIn();

    $('#starting-english').text(clue);
}

function waitTurn() {
    $('#emojify-container').hide();
    $('.emojiPicker').hide();
    $('#answer-container').hide();
    $('#wait-container').fadeIn();
}

function displayAnswer(answer, correct, clue, real) {
    $('#wait-container').hide();
    $('#englishify-container').hide();
    $('#emojify-container').hide();
    
    $('#ans-original-phrase').text(real);
    $('#ans-emoji-clue').text(clue);
    $('#ans-english').text(answer);
    if (correct) {
        $('#ans-result').text("Good job! 👌");
    } else {
        $('#ans-result').text("Not quite. 😕");
    }
    $('#answer-container').show();

}

var GLOBAL_WS;

$('#play-button').click(function() {
    var nick = $('#nick-input').val().trim();
    if (nick == "") {
        alert("Nickname cannot be blank!");
        return;
    }
    $('#play-button').addClass('disabled');
    $('#play-container .input').addClass('disabled');

    var ws = new WebSocket("ws://localhost:3000/websocket");
    GLOBAL_WS = ws;
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
        console.log(msg);
        var action  = msg['Action'];
        var payload = msg['Payload'];
        if (action == "found_game") {
            beginGame(payload['match']);
        } else if (action == "turn") {
            var turn = payload['turn'];
            if (turn == "your") {
                if (payload['state'] == null || payload['state'] == "give_clue") {
                    playTurn(payload['word']);
                } else if (payload['state'] == 'give_answer') {
                    playTurnEnglish(payload['clue']);
                }
            } else if (turn == "their") {
                waitTurn();
            }
        } else if (action == "answer") {
            var answer  = payload['answer'];
            var correct = payload['boolean'];
            var clue    = payload['clue'];
            var real    = payload['real_answer'];
            displayAnswer(answer, correct, clue, real);
        }
    };
});

$('#emoji-input-delete').click(function() {
    var elem = $('#emoji-input');
    elem.val(elem.val().slice(0, -2));
});

$('#emoji-submit').click(function() {
    var msg = {
        Action: 'submit_clue',
        Payload: {
            clue: $('#emoji-input').val()
        }
    }
    GLOBAL_WS.send(JSON.stringify(msg));
});

$('#english-submit').click(function() {
    var msg = {
        Action: 'submit_answer',
        Payload: {
            answer: $('#english-input').val()
        }
    };
    GLOBAL_WS.send(JSON.stringify(msg));
});
