// src/components/GameControls.css.jsx
import React from 'react';

function GameControls({
                          twitterHandle,
                          setTwitterHandle,
                          onStart,
                          onSubmit,
                          isConnected
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
          {isConnected ? 'Connected' : 'Connecting...'}
        </span>
            </div>

            <div className="control-panel">
                <button
                    className="start-button"
                    onClick={onStart}
                    disabled={!isConnected}
                >
                    Start Game
                </button>

                <div className="twitter-input">
                    <label htmlFor="twitter-handle">@</label>
                    <input
                        id="twitter-handle"
                        type="text"
                        value={twitterHandle}
                        onChange={handleTwitterChange}
                        placeholder="Twitter handle"
                        disabled={!isConnected}
                    />
                    <button
                        onClick={onSubmit}
                        disabled={!isConnected || !twitterHandle}
                    >
                        Submit
                    </button>
                </div>
            </div>
        </div>
    );
}

export default GameControls;