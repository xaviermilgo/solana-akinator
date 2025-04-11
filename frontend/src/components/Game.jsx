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
    const [isProcessing, setIsProcessing] = useState(false);

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

    // Update message based on Jinn state
    useEffect(() => {
        switch(jinnState) {
            case 'idle':
                setMessage("I am the Crypto Jinn! I can divine your wallet address from your Twitter handle!");
                break;
            case 'thinking':
                setMessage("Hmm... I'm consulting the mystical blockchain ledgers...");
                setIsProcessing(true);
                break;
            case 'asking':
                setMessage("Tell me more... Is your Twitter activity primarily focused on crypto?");
                setIsProcessing(false);
                break;
            case 'confident':
                setMessage("Aha! I'm sensing a strong connection to the blockchain realm!");
                break;
            case 'correct':
                setMessage("Behold! I have divined your wallet correctly. The stars were aligned!");
                setIsProcessing(false);
                break;
            case 'wrong':
                setMessage("The crypto spirits have misled me. Perhaps you've obscured your digital footprints well!");
                setIsProcessing(false);
                break;
            case 'glitched':
                setMessage("The fabric between realms is disturbed! Something has interfered with my powers...");
                setIsProcessing(false);
                break;
            default:
                setMessage("I am the Crypto Jinn! I can divine your wallet address from your Twitter handle!");
        }
    }, [jinnState]);

    const handleWebSocketMessage = (data) => {
        switch (data.type) {
            case 'JINN_STATE':
                setJinnState(data.payload.state);
                // Custom message if provided
                if (data.payload.message) {
                    setMessage(data.payload.message);
                }
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
            // Reset any previous state
            setTwitterHandle('');
            // Set initial idle state to ensure animation plays
            setJinnState('idle');
            setMessage("I am the Crypto Jinn! Share your Twitter handle, and I shall reveal your crypto identity!");

            socket.send(JSON.stringify({
                type: 'START_GAME',
                payload: {},
            }));
        }
    };

    const handleSubmitTwitter = () => {
        if (socket && isConnected && twitterHandle) {
            // Show thinking animation immediately for better UX
            setJinnState('thinking');
            setMessage("Hmm... I'm consulting the mystical blockchain ledgers...");
            setIsProcessing(true);

            socket.send(JSON.stringify({
                type: 'USER_INPUT',
                payload: {
                    twitter: twitterHandle,
                },
            }));
        }
    };

    // Add mystical effects with CSS
    const backgroundStyle = {
        background: `radial-gradient(circle at center, rgba(91, 33, 182, 0.1) 0%, rgba(10, 7, 33, 0) 70%)`,
    };

    return (
        <div className="game-container">
            <div className="game-background">
                <img src={backgroundImage} alt="Mystical gateway" />
            </div>

            <div className="game-content" style={backgroundStyle}>
                {/* Add subtle particle effects or mystical elements */}
                <div className="mystical-orbs"></div>

                <Jinn state={jinnState} />

                <div className="game-message">
                    {message}
                </div>

                <GameControls
                    twitterHandle={twitterHandle}
                    setTwitterHandle={setTwitterHandle}
                    onStart={handleStartGame}
                    onSubmit={handleSubmitTwitter}
                    isConnected={isConnected}
                    isProcessing={isProcessing}
                />
            </div>
        </div>
    );
}

export default Game;