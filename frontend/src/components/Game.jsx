import React, { useState, useEffect, useRef } from 'react';
import { Copy, ExternalLink, RefreshCw } from 'lucide-react';
import { connectWebSocket } from '../services/websocket';

const ImprovedGame = () => {
    // State management
    const [jinnState, setJinnState] = useState('idle');
    const [message, setMessage] = useState('I am the Crypto Jinn! I can divine your wallet address from your Twitter handle!');
    const [twitterHandle, setTwitterHandle] = useState('');
    const [isConnected, setIsConnected] = useState(false);
    const [isProcessing, setIsProcessing] = useState(false);
    const [progressLogs, setProgressLogs] = useState([]);
    const [walletResults, setWalletResults] = useState(null);
    const [allWallets, setAllWallets] = useState([]);
    const [selectedWallets, setSelectedWallets] = useState({});
    const [socket, setSocket] = useState(null);
    const logsEndRef = useRef(null);

    // Connect to WebSocket when component mounts
    useEffect(() => {
        const { socket: newSocket, disconnect } = connectWebSocket({
            onMessage: handleWebSocketMessage,
            onConnect: () => setIsConnected(true),
            onDisconnect: () => setIsConnected(false),
        });

        setSocket(newSocket);

        // Clean up WebSocket connection when component unmounts
        return () => disconnect();
    }, []);

    // Auto-scroll logs
    useEffect(() => {
        if (logsEndRef.current) {
            logsEndRef.current.scrollIntoView({ behavior: 'smooth' });
        }
    }, [progressLogs]);

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
        }
    }, [jinnState]);

    // WebSocket message handler
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
            case 'PROGRESS_UPDATE':
                // Add new progress log
                setProgressLogs(logs => [...logs, data.payload.message]);
                break;
            case 'WALLET_RESULT':
                // Set wallet result
                setWalletResults(data.payload);

                // Create a comprehensive list of all wallets with at least one match
                if (data.payload && data.payload.addresses) {
                    // Extract score from sources (assuming format "Match score: XX/100. ...")
                    const extractScore = (source) => {
                        const match = source.match(/Match score: (\d+)\/100/);
                        return match ? parseInt(match[1]) : 0;
                    };

                    const wallets = data.payload.addresses.map((address, index) => {
                        const source = data.payload.sources[index] || '';
                        const score = extractScore(source);

                        // Extract sources from the string (assuming format "... Matched tokens from: @source1, @source2")
                        const sourcesMatch = source.match(/Matched tokens from: (.*)/);
                        const sourcesList = sourcesMatch ?
                            sourcesMatch[1].split(', ').map(s => s.trim()) :
                            [];

                        return {
                            address,
                            score,
                            sources: sourcesList,
                            selected: index === 0 // Select the first wallet by default
                        };
                    });

                    setAllWallets(wallets);

                    // Initialize selected wallets (select the first one by default)
                    const selected = {};
                    if (wallets.length > 0) {
                        selected[wallets[0].address] = true;
                    }
                    setSelectedWallets(selected);
                }
                break;
        }
    };

    // Handle Twitter handle submission
    const handleSubmit = (e) => {
        e.preventDefault();
        if (!twitterHandle || !isConnected || isProcessing) return;

        // Reset state for new search
        setWalletResults(null);
        setAllWallets([]);
        setSelectedWallets({});
        setProgressLogs([]);

        // Show thinking animation immediately for better UX
        setJinnState('thinking');
        setIsProcessing(true);

        // Send request to backend
        socket.send(JSON.stringify({
            type: 'USER_INPUT',
            payload: {
                twitter: twitterHandle,
            },
        }));
    };

    // Copy address to clipboard
    const copyToClipboard = (address) => {
        navigator.clipboard.writeText(address);
    };

    // Toggle wallet selection
    const toggleWalletSelection = (address) => {
        setSelectedWallets(prev => ({
            ...prev,
            [address]: !prev[address]
        }));
    };

    // Reset the game
    const resetGame = () => {
        setWalletResults(null);
        setAllWallets([]);
        setSelectedWallets({});
        setProgressLogs([]);
        setJinnState('idle');
        setIsProcessing(false);
        setTwitterHandle('');
    };

    // Truncate wallet address for display
    const truncateAddress = (address) => {
        if (address.length <= 16) return address;
        return `${address.substring(0, 8)}...${address.substring(address.length - 8)}`;
    };

    // Get confidence class based on score
    const getConfidenceClass = (score) => {
        if (score >= 70) return "text-emerald-400";
        if (score >= 40) return "text-amber-400";
        return "text-rose-400";
    };

    return (
        <div className="flex flex-col h-screen bg-slate-900 text-slate-100 overflow-auto">
            {/* Header with connection status */}
            <header className="bg-slate-800 border-b border-slate-700 py-3 px-4">
                <div className="flex items-center justify-between max-w-7xl mx-auto">
                    <h1 className="text-xl font-bold bg-gradient-to-r from-blue-400 to-purple-500 bg-clip-text text-transparent">
                        ✨ The Magical Wallet Guesser ✨
                    </h1>
                    <div className="flex items-center">
            <span className={`flex items-center text-sm ${isConnected ? 'text-emerald-400' : 'text-rose-400'}`}>
              <span className="inline-block w-2 h-2 rounded-full mr-2 bg-current"></span>
                {isConnected ? 'Connected to the Mystical Realm' : 'Connecting...'}
            </span>
                    </div>
                </div>
            </header>

            {/* Main content area */}
            <main className="flex-1 flex flex-col md:flex-row max-w-7xl mx-auto w-full p-4 gap-4">
                {/* Left column - Jinn and Status */}
                <div className="w-full md:w-1/3 flex flex-col">
                    {/* Jinn and Message */}
                    <div className="bg-slate-800 rounded-lg border border-slate-700 p-4 mb-4 flex flex-col items-center">
                        <div className={`w-32 h-32 flex items-center justify-center mt-2 mb-4 jinn jinn-${jinnState}`}>
                            {/* Jinn image would be here - using a placeholder */}
                            <div className="w-24 h-24 rounded-full bg-gradient-to-br from-blue-500 to-purple-600 flex items-center justify-center text-2xl">
                                🧞
                            </div>
                        </div>
                        <div className="text-center text-sm text-slate-200 mb-2">{message}</div>
                    </div>

                    {/* Logs Panel */}
                    {progressLogs.length > 0 && (
                        <div className="bg-slate-800 rounded-lg border border-slate-700 p-4 flex-1 mb-4 overflow-hidden">
                            <h2 className="text-blue-400 text-sm font-medium mb-2">The Jinn's Mystical Process:</h2>
                            <div className="h-56 overflow-y-auto text-xs pr-2 space-y-1">
                                {progressLogs.map((log, index) => (
                                    <div key={index} className="py-1 border-b text-left border-slate-700 text-slate-300">
                                        <span className="text-emerald-400 mr-1">✧</span> {log}
                                    </div>
                                ))}
                                <div ref={logsEndRef} />
                            </div>
                        </div>
                    )}

                    {/* Input Form */}
                    {!walletResults && (
                        <div className="bg-slate-800 rounded-lg border border-slate-700 p-4">
                            <form onSubmit={handleSubmit}>
                                <div className="mb-2 text-sm text-slate-300 text-center">
                                    Enter your Twitter handle below and the Crypto Jinn will divine your wallet address
                                </div>
                                <div className="flex">
                  <span className="inline-flex items-center px-3 rounded-l-md border border-r-0 border-slate-600 bg-slate-700 text-slate-300">
                    @
                  </span>
                                    <input
                                        type="text"
                                        value={twitterHandle}
                                        onChange={(e) => setTwitterHandle(e.target.value.replace('@', ''))}
                                        className="flex-1 min-w-0 block w-full px-3 py-2 rounded-none bg-slate-700 border border-slate-600 text-slate-200 focus:outline-none focus:ring-1 focus:ring-blue-500 focus:border-blue-500"
                                        placeholder="Twitter handle"
                                        disabled={!isConnected || isProcessing}
                                    />
                                    <button
                                        type="submit"
                                        className="inline-flex items-center px-4 py-2 border border-l-0 border-transparent rounded-r-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50 disabled:cursor-not-allowed"
                                        disabled={!isConnected || !twitterHandle || isProcessing}
                                    >
                                        {isProcessing ? (
                                            <div className="flex items-center">
                                                <RefreshCw className="animate-spin -ml-1 mr-2 h-4 w-4" />
                                                <span>Divining...</span>
                                            </div>
                                        ) : 'Consult the Jinn'}
                                    </button>
                                </div>
                            </form>
                        </div>
                    )}
                </div>

                {/* Right column - Results */}
                {walletResults && (
                    <div className="w-full md:w-2/3 bg-slate-800 rounded-lg border border-slate-700 p-4">
                        <div className="flex justify-between items-center mb-4">
                            <h2 className="text-lg font-medium text-blue-400">
                                Divination Results for @{walletResults.twitterHandle}
                            </h2>
                            <button
                                onClick={resetGame}
                                className="px-3 py-1 text-xs bg-slate-700 hover:bg-slate-600 rounded text-slate-200"
                            >
                                Try Another Handle
                            </button>
                        </div>

                        {/* Confidence Meter */}
                        <div className="mb-6">
                            <div className="flex items-center justify-between mb-1">
                                <span className="text-sm text-slate-300">Mystical Confidence:</span>
                                <span className={`text-sm font-medium ${getConfidenceClass(walletResults.confidence)}`}>
                  {walletResults.confidence}%
                </span>
                            </div>
                            <div className="w-full bg-slate-700 rounded-full h-2">
                                <div
                                    className={`h-2 rounded-full ${
                                        walletResults.confidence >= 70 ? 'bg-gradient-to-r from-emerald-500 to-blue-500' :
                                            walletResults.confidence >= 40 ? 'bg-gradient-to-r from-amber-500 to-amber-400' :
                                                'bg-gradient-to-r from-rose-500 to-rose-400'
                                    }`}
                                    style={{ width: `${walletResults.confidence}%` }}
                                ></div>
                            </div>
                        </div>

                        {/* Wallet Results Table */}
                        <div className="mb-4">
                            <h3 className="text-md font-medium text-slate-300 mb-2">All Potential Wallet Addresses:</h3>
                            <div className="overflow-x-auto">
                                <table className="min-w-full divide-y divide-slate-700">
                                    <thead className="bg-slate-700">
                                    <tr>
                                        <th scope="col" className="px-3 py-2 text-left text-xs font-medium text-slate-300 uppercase tracking-wider w-10">
                                            Select
                                        </th>
                                        <th scope="col" className="px-3 py-2 text-left text-xs font-medium text-slate-300 uppercase tracking-wider">
                                            Wallet Address
                                        </th>
                                        <th scope="col" className="px-3 py-2 text-left text-xs font-medium text-slate-300 uppercase tracking-wider w-20">
                                            Score
                                        </th>
                                        <th scope="col" className="px-3 py-2 text-left text-xs font-medium text-slate-300 uppercase tracking-wider">
                                            Sources
                                        </th>
                                        <th scope="col" className="px-3 py-2 text-left text-xs font-medium text-slate-300 uppercase tracking-wider w-20">
                                            Actions
                                        </th>
                                    </tr>
                                    </thead>
                                    <tbody className="bg-slate-800 divide-y divide-slate-700">
                                    {allWallets.map((wallet, idx) => (
                                        <tr key={wallet.address} className={idx % 2 === 0 ? 'bg-slate-800' : 'bg-slate-750'}>
                                            <td className="px-3 py-2 whitespace-nowrap text-sm">
                                                <input
                                                    type="checkbox"
                                                    checked={!!selectedWallets[wallet.address]}
                                                    onChange={() => toggleWalletSelection(wallet.address)}
                                                    className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-slate-600 rounded"
                                                />
                                            </td>
                                            <td className="px-3 py-2 whitespace-nowrap font-mono text-xs text-slate-300">
                                                {wallet.address}
                                            </td>
                                            <td className="px-3 py-2 whitespace-nowrap text-sm">
                          <span className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium ${getConfidenceClass(wallet.score)}`}>
                            {wallet.score}/100
                          </span>
                                            </td>
                                            <td className="px-3 py-2 whitespace-nowrap text-sm text-slate-400">
                                                {wallet.sources.join(', ')}
                                            </td>
                                            <td className="px-3 py-2 whitespace-nowrap text-sm text-slate-300">
                                                <button
                                                    onClick={() => copyToClipboard(wallet.address)}
                                                    className="inline-flex items-center p-1 mr-2 border border-transparent rounded text-slate-300 hover:bg-slate-700"
                                                    title="Copy address"
                                                >
                                                    <Copy size={16} />
                                                </button>
                                                <a
                                                    href={`https://solscan.io/account/${wallet.address}`}
                                                    target="_blank"
                                                    rel="noopener noreferrer"
                                                    className="inline-flex items-center p-1 border border-transparent rounded text-slate-300 hover:bg-slate-700"
                                                    title="View on Solscan"
                                                >
                                                    <ExternalLink size={16} />
                                                </a>
                                            </td>
                                        </tr>
                                    ))}
                                    </tbody>
                                </table>
                            </div>
                        </div>

                        {/* Selected Wallets Summary */}
                        <div className="mt-6 bg-slate-700 rounded-md p-3">
                            <h3 className="text-sm font-medium text-slate-200 mb-2">Selected Wallets:</h3>
                            <div className="space-y-2">
                                {Object.entries(selectedWallets).filter(([_, isSelected]) => isSelected).map(([address]) => {
                                    const wallet = allWallets.find(w => w.address === address);
                                    return wallet ? (
                                        <div key={address} className="flex items-center justify-between bg-slate-600 px-3 py-2 rounded">
                                            <div className="font-mono text-xs text-slate-200">{truncateAddress(address)}</div>
                                            <div className="flex items-center">
                        <span className={`text-xs ${getConfidenceClass(wallet.score)} mr-2`}>
                          {wallet.score}/100
                        </span>
                                                <button
                                                    onClick={() => copyToClipboard(address)}
                                                    className="inline-flex items-center p-1 border border-transparent rounded text-slate-300 hover:bg-slate-700"
                                                >
                                                    <Copy size={14} />
                                                </button>
                                            </div>
                                        </div>
                                    ) : null;
                                })}
                            </div>
                        </div>
                    </div>
                )}
            </main>

            {/* Footer */}
            <footer className="bg-slate-800 border-t border-slate-700 py-2 px-4 text-sm text-slate-400 text-center">
                &copy; 2025 Magical Wallet Guesser | Made with ❤️ by <a href="https://twitter.com/xaviermilgo" target="_blank" rel="noopener noreferrer" className="text-blue-400 hover:text-blue-300">@xaviermilgo</a>
            </footer>
        </div>
    );
};

export default ImprovedGame;