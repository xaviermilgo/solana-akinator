const WS_URL = process.env.REACT_APP_WS_URL || 'ws://localhost:8080/ws';

export function connectWebSocket({ onMessage, onConnect, onDisconnect }) {
    let socket = null;
    let reconnectTimer = null;
    let reconnectAttempts = 0;
    const maxReconnectAttempts = 5;
    let messageQueue = [];

    // Create WebSocket connection
    function connect() {
        // Clear any existing reconnect timer
        if (reconnectTimer) {
            clearTimeout(reconnectTimer);
            reconnectTimer = null;
        }

        // Create a new WebSocket connection
        const ws = new WebSocket(WS_URL);

        // Connection opened
        ws.addEventListener('open', () => {
            console.log('WebSocket connection established');
            reconnectAttempts = 0; // Reset reconnect attempts on successful connection

            // Process any queued messages
            if (messageQueue.length > 0) {
                console.log(`Processing ${messageQueue.length} queued messages`);
                messageQueue.forEach(msg => {
                    if (ws.readyState === WebSocket.OPEN) {
                        ws.send(msg);
                    }
                });
                messageQueue = []; // Clear the queue
            }

            if (onConnect) onConnect();
        });

        // Listen for messages
        ws.addEventListener('message', (event) => {
            try {
                const data = JSON.parse(event.data);
                if (onMessage) onMessage(data);
            } catch (error) {
                console.error('Error parsing WebSocket message:', error);
            }
        });

        // Handle errors
        ws.addEventListener('error', (error) => {
            console.error('WebSocket error:', error);
        });

        // Connection closed
        ws.addEventListener('close', (event) => {
            console.log('WebSocket connection closed:', event.code, event.reason);
            if (onDisconnect) onDisconnect();

            // Attempt to reconnect with exponential backoff
            if (reconnectAttempts < maxReconnectAttempts) {
                const delay = Math.min(1000 * Math.pow(2, reconnectAttempts), 10000);
                console.log(`Attempting to reconnect in ${delay}ms (attempt ${reconnectAttempts + 1}/${maxReconnectAttempts})...`);

                reconnectTimer = setTimeout(() => {
                    reconnectAttempts++;
                    console.log(`Reconnecting... Attempt ${reconnectAttempts}/${maxReconnectAttempts}`);
                    connect();
                }, delay);
            } else {
                console.log('Maximum reconnect attempts reached. Please refresh the page to try again.');
            }
        });

        // Add custom send method that queues messages when connection isn't open
        ws.safeSend = function(data) {
            if (this.readyState === WebSocket.OPEN) {
                this.send(data);
            } else {
                console.log('Connection not open, queueing message');
                messageQueue.push(data);
            }
        };

        socket = ws;
        return ws;
    }

    // Disconnect WebSocket
    function disconnect() {
        if (socket && socket.readyState === WebSocket.OPEN) {
            socket.close();
        }

        if (reconnectTimer) {
            clearTimeout(reconnectTimer);
            reconnectTimer = null;
        }

        // Clear message queue
        messageQueue = [];
    }

    // Start the connection
    const initialSocket = connect();

    return {
        socket: initialSocket,
        disconnect,
    };
}