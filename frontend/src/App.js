import React from 'react';
import Game from './components/Game';

function App() {
    return (
        <div className="app">
            <header className="app-header">
                <h1>✨ The Magical Wallet Guesser ✨</h1>
            </header>
            <main>
                <Game />
            </main>
            <footer className="app-footer">
                <p>&copy; 2025 Magical Wallet Guesser | Powered by the Crypto Jinn</p>
            </footer>
        </div>
    );
}

export default App;