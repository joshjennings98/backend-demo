window.addEventListener('load', function () {
    var term = new Terminal({
        fontSize: 24,
        fontFamily: "monospace",
    });
    term.open(document.getElementById('terminal'));

    var socket;

    function initWebSocket() {
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
            console.log(`Connection to ${socketUrl} closed. Attempting to reconnect in 1 second.`);
            setTimeout(initWebSocket, 1000);  // Reconnect after 1 second
        };

        socket.onerror = function (error) {
            console.error(`WebSocket error: ${error.message}`);
        };
    }

    initWebSocket(); 
});

