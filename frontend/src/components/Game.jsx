// src/components/Game.jsx
import React, { useState, useEffect } from 'react';
import Jinn from './Jinn';
import GameControls from './GameControls';
import { connectWebSocket } from '../services/websocket';
import backgroundImage from "../assets/background.png"

function Game() {
    const [jinnState, setJinnState] = useState('idle');
    const [message, setMessage] = useState('');
    const [twitterHandle, setTwitterHandle] = useState('');
    const [socket, setSocket] = useState(null);
    const [isConnected, setIsConnected] = useState(false);

    useEffect(() => {
        // Connect to WebSocket when component mounts
        const { socket: newSocket, disconnect } = connectWebSocket({
            onMessage: handleWebSocketMessage,
            onConnect: () => setIsConnected(true),
            onDisconnect: () => setIsConnected(false),
        });

        setSocket(newSocket);

        // Clean up WebSocket connection when component unmounts
        return () => {
            disconnect();
        };
    }, []);

    const handleWebSocketMessage = (data) => {
        switch (data.type) {
            case 'JINN_STATE':
                setJinnState(data.payload.state);
                break;
            case 'GAME_STATE':
                // Update game state from backend
                setJinnState(data.payload.jinnState);
                break;
            default:
                console.log('Unknown message type:', data.type);
        }
    };

    const handleStartGame = () => {
        if (socket && isConnected) {
            socket.send(JSON.stringify({
                type: 'START_GAME',
                payload: {},
            }));
        }
    };

    const handleSubmitTwitter = () => {
        if (socket && isConnected && twitterHandle) {
            socket.send(JSON.stringify({
                type: 'USER_INPUT',
                payload: {
                    twitter: twitterHandle,
                },
            }));
        }
    };

    return (
        <div className="game-container">
            <div className="game-background">
                <img src={backgroundImage} alt="Mystical background" />
            </div>

            <div className="game-content">
                <Jinn state={jinnState} />

                <div className="game-message">
                    {message || "I can guess your wallet address from your Twitter handle!"}
                </div>

                <GameControls
                    twitterHandle={twitterHandle}
                    setTwitterHandle={setTwitterHandle}
                    onStart={handleStartGame}
                    onSubmit={handleSubmitTwitter}
                    isConnected={isConnected}
                />
            </div>
        </div>
    );
}

export default Game;