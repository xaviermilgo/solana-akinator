import React, { useState, useEffect, useRef } from 'react';
import Jinn from './Jinn';
import GameControls from './GameControls';
import { connectWebSocket } from '../services/websocket';
import backgroundImage from "../assets/background.png";

function Game() {
    const [jinnState, setJinnState] = useState('idle');
    const [message, setMessage] = useState('');
    const [twitterHandle, setTwitterHandle] = useState('');
    const [socket, setSocket] = useState(null);
    const [isConnected, setIsConnected] = useState(false);
    const [isProcessing, setIsProcessing] = useState(false);
    const orbsRef = useRef(null);

    // Create mystical orbs dynamically
    useEffect(() => {
        if (orbsRef.current) {
            // Create dynamic orbs
            const orb1 = document.createElement('div');
            orb1.className = 'orb1';
            const orb2 = document.createElement('div');
            orb2.className = 'orb2';

            orbsRef.current.appendChild(orb1);
            orbsRef.current.appendChild(orb2);

            // Cleanup
            return () => {
                if (orbsRef.current) {
                    orbsRef.current.removeChild(orb1);
                    orbsRef.current.removeChild(orb2);
                }
            };
        }
    }, []);

    // Add special effects when Jinn changes state
    useEffect(() => {
        const addTemporaryEffect = () => {
            if (!orbsRef.current) return;

            // Create a burst effect
            const burstEffect = document.createElement('div');
            burstEffect.className = `state-transition-effect ${jinnState}-transition`;
            orbsRef.current.appendChild(burstEffect);

            // Remove after animation completes
            setTimeout(() => {
                if (orbsRef.current && orbsRef.current.contains(burstEffect)) {
                    orbsRef.current.removeChild(burstEffect);
                }
            }, 1000);
        };

        if (jinnState !== 'idle') {
            addTemporaryEffect();
        }
    }, [jinnState]);

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

    return (
        <div className="game-container">
            <div className="game-background">
                <img src={backgroundImage} alt="Mystical gateway" />
            </div>

            <div className="game-content">
                {/* Add mystical orbs with ref */}
                <div className="mystical-orbs" ref={orbsRef}></div>

                <Jinn state={jinnState} />

                <div className="game-message">
                    {message}
                </div>

                <GameControls
                    twitterHandle={twitterHandle}
                    setTwitterHandle={setTwitterHandle}
                    onSubmit={handleSubmitTwitter}
                    isConnected={isConnected}
                    isProcessing={isProcessing}
                />
            </div>
        </div>
    );
}

export default Game;