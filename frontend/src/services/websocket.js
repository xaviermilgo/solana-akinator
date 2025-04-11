// src/services/websocket.js

const WS_URL = process.env.REACT_APP_WS_URL || 'ws://localhost:8080/ws';

export function connectWebSocket({ onMessage, onConnect, onDisconnect }) {
    let socket = null;
    let reconnectTimer = null;

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

            // Attempt to reconnect after a delay
            reconnectTimer = setTimeout(() => {
                console.log('Attempting to reconnect...');
                connect();
            }, 3000);
        });

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
    }

    // Start the connection
    const initialSocket = connect();

    return {
        socket: initialSocket,
        disconnect,
    };
}