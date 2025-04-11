// src/components/GameControls.jsx
import React from 'react';

function GameControls({
                          twitterHandle,
                          setTwitterHandle,
                          onSubmit,
                          isConnected,
                          isProcessing = false
                      }) {
    const handleTwitterChange = (e) => {
        // Remove @ if user typed it
        const handle = e.target.value.replace('@', '');
        setTwitterHandle(handle);
    };

    return (
        <div className="game-controls">
            <div className="connection-status">
                <span className={isConnected ? 'connected' : 'disconnected'}>
                    {isConnected ? '● Connected to the Mystical Realm' : '○ Connecting...'}
                </span>
            </div>

            <div className="control-panel">
                <div className="jinn-description">
                    <p>Enter your Twitter handle below and the Crypto Jinn will divine your wallet address using mystical blockchain powers!</p>
                </div>

                <div className="twitter-input-container">
                    <div className="twitter-input">
                        <label htmlFor="twitter-handle">@</label>
                        <input
                            id="twitter-handle"
                            type="text"
                            value={twitterHandle}
                            onChange={handleTwitterChange}
                            placeholder="Enter Twitter handle"
                            disabled={!isConnected || isProcessing}
                        />
                    </div>
                    <button
                        className="submit-button"
                        onClick={onSubmit}
                        disabled={!isConnected || !twitterHandle || isProcessing}
                    >
                        {isProcessing ? 'Divining...' : 'Consult the Jinn'}
                    </button>
                </div>
            </div>
        </div>
    );
}

export default GameControls;