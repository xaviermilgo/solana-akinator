import React from 'react';

function WalletResults({ result, onBack }) {
    if (!result || !result.addresses || result.addresses.length === 0) {
        return (
            <div className="wallet-results empty-results">
                <div className="result-message">
                    No wallet addresses found. The blockchain spirits remain silent.
                </div>
                <button className="back-button" onClick={onBack}>Try Another Handle</button>
            </div>
        );
    }

    // Function to truncate wallet addresses for display
    const truncateAddress = (address) => {
        if (address.length <= 16) return address;
        return `${address.substring(0, 8)}...${address.substring(address.length - 8)}`;
    };

    // Function to copy address to clipboard
    const copyToClipboard = (text) => {
        navigator.clipboard.writeText(text).then(
            () => {
                // Show a temporary "Copied!" message
                const copyButton = document.getElementById(`copy-${text.substring(0, 8)}`);
                if (copyButton) {
                    const originalText = copyButton.textContent;
                    copyButton.textContent = "Copied!";
                    copyButton.classList.add("copied");

                    setTimeout(() => {
                        copyButton.textContent = originalText;
                        copyButton.classList.remove("copied");
                    }, 2000);
                }
            },
            (err) => {
                console.error('Could not copy text: ', err);
            }
        );
    };

    // Calculate confidence class
    const getConfidenceClass = (confidence) => {
        if (confidence >= 70) return "high-confidence";
        if (confidence >= 40) return "medium-confidence";
        return "low-confidence";
    };

    return (
        <div className="wallet-results">
            <div className="results-header">
                <h3>Divination Results for @{result.twitterHandle}</h3>
                <div className={`confidence-meter ${getConfidenceClass(result.confidence)}`}>
                    <span className="confidence-label">Mystical Confidence:</span>
                    <div className="confidence-bar">
                        <div
                            className="confidence-fill"
                            style={{ width: `${result.confidence}%` }}
                        ></div>
                    </div>
                    <span className="confidence-value">{result.confidence}%</span>
                </div>
            </div>

            <div className="wallet-list">
                <h4>Potential Wallet Addresses:</h4>
                {result.addresses.map((address, index) => (
                    <div key={index} className="wallet-item">
                        <div className="wallet-address">
                            <span className="address-text">{truncateAddress(address)}</span>
                            <button
                                id={`copy-${address.substring(0, 8)}`}
                                className="copy-button"
                                onClick={() => copyToClipboard(address)}
                            >
                                Copy
                            </button>
                        </div>
                        {result.sources && result.sources[index] && (
                            <div className="wallet-source">{result.sources[index]}</div>
                        )}
                    </div>
                ))}
            </div>

            <div className="actions">
                <button className="back-button" onClick={onBack}>Try Another Handle</button>
            </div>
        </div>
    );
}

export default WalletResults;