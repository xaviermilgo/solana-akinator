// src/App.jsx
import React from 'react';
import Game from './components/Game';

function App() {
  return (
      <div className="app">
        <header className="app-header">
          <h1>Wallet Address Guesser</h1>
        </header>
        <main>
          <Game />
        </main>
        <footer className="app-footer">
          <p>&copy; 2025 Wallet Guesser</p>
        </footer>
      </div>
  );
}

export default App;