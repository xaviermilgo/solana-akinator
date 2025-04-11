import React from 'react';

function GameControls({
                          twitterHandle,
                          setTwitterHandle,
                          onStart,
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
                    {isConnected ? '● Connected to the Mystical Realm' : '○ Establishing Connection...'}
                </span>
            </div>

            <div className="control-panel">
                <button
                    className="start-button"
                    onClick={onStart}
                    disabled={!isConnected || isProcessing}
                >
                    {isProcessing ? 'Divination in Progress...' : 'Begin Mystical Divination'}
                </button>

                <div className="twitter-input">
                    <label htmlFor="twitter-handle">@</label>
                    <input
                        id="twitter-handle"
                        type="text"
                        value={twitterHandle}
                        onChange={handleTwitterChange}
                        placeholder="Enter your Twitter handle"
                        disabled={!isConnected || isProcessing}
                    />
                    <button
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