var socket

window.addEventListener('load', function () {
    var term = new Terminal({
        fontSize: 24,
    });
    term.open(document.getElementById('terminal'));

    var socketUrl = 'ws://localhost:8080/init';
    
    if (socket) {
        socket.close();
    }
    socket = new WebSocket(socketUrl);

    socket.onopen = function (e) {
        console.log(`Connection established to ${socketUrl}`);
    };

    socket.onmessage = function (event) {
        term.write(event.data);
    };

    socket.onclose = function (event) {
        console.log(`Connection to ${socketUrl} closed`);
    };

    socket.onerror = function (error) {
        console.error(`Websocket error: ${error.message}`);
    };
});
