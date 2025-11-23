window.addEventListener('load', function () {
    const TERMINAL_FONT_SIZE = 22;
    const TERMINAL_FONT_FAMILY = '"SF Mono", "Fira Code", "Consolas", monospace';
    const TERMINAL_SCROLLBACK = 10000;
    const WS_INITIAL_RETRY_DELAY_MS = 1000;
    const WS_MAX_RETRY_DELAY_MS = 30000;
    const WS_BACKOFF_MULTIPLIER = 2;

    var term = new Terminal({
        fontSize: TERMINAL_FONT_SIZE,
        fontFamily: TERMINAL_FONT_FAMILY,
        scrollback: TERMINAL_SCROLLBACK,
        theme: {
            background: 'transparent',
            foreground: '#00ff00',
            cursor: '#00ff00',
            cursorAccent: '#000',
            selectionBackground: 'rgba(0, 255, 0, 0.3)',
        }
    });
    term.open(document.getElementById('terminal'));

    var socket;
    var retryDelay = WS_INITIAL_RETRY_DELAY_MS;

    function initWebSocket() {
        var socketUrl = 'ws://' + window.location.host + '/ws';

        if (socket) {
            socket.close();
        }
        socket = new WebSocket(socketUrl);

        socket.onopen = function (e) {
            console.log(`Connection established to ${socketUrl}`);
            retryDelay = WS_INITIAL_RETRY_DELAY_MS;
        };

        socket.onmessage = function (event) {
            term.write(event.data);
        };

        socket.onclose = function (event) {
            var jitter = Math.random() * 0.3 * retryDelay;
            var delayWithJitter = retryDelay + jitter;

            console.log(`Connection to ${socketUrl} closed. Attempting to reconnect in ${Math.round(delayWithJitter / 1000)} seconds.`);
            setTimeout(initWebSocket, delayWithJitter);

            retryDelay = Math.min(retryDelay * WS_BACKOFF_MULTIPLIER, WS_MAX_RETRY_DELAY_MS);
        };

        socket.onerror = function (error) {
            console.error(`WebSocket error: ${error.message}`);
        };
    }

    initWebSocket();

    function setupHighlighting() {
        hljs.highlightAll();

        var pendingHighlight = false;

        const observer = new MutationObserver(function(mutations) {
            if (!pendingHighlight) {
                pendingHighlight = true;
                requestAnimationFrame(function() {
                    hljs.highlightAll();
                    pendingHighlight = false;
                });
            }
        });

        observer.observe(document.body, {
            childList: true,
            subtree: true,
        });
    }

    setupHighlighting();

    // Track slide type for conditional fade-in and only fade when transitioning between different slide types
    var previousSlideType = null;

    function getSlideType() {
        var slideContent = document.getElementById('slide-content');
        if (!slideContent) return null;
        if (slideContent.querySelector('.command-string')) return 'command';
        if (slideContent.querySelector('.text-string')) return 'text';
        return null;
    }

    document.body.addEventListener('htmx:beforeSwap', function(event) {
        if (event.detail.target && event.detail.target.id === 'command') {
            previousSlideType = getSlideType();
        }
    });

    // After swap, conditionally apply fade-in
    document.body.addEventListener('htmx:afterSwap', function(event) {
        if (event.detail.target && event.detail.target.id === 'command') {
            var newSlideType = getSlideType();
            var slideContent = document.getElementById('slide-content');

            if (slideContent && previousSlideType !== newSlideType) {
                slideContent.classList.add('fade-in');
            }

            previousSlideType = newSlideType;
        }
    });

    previousSlideType = getSlideType();
});
