# Wallet Guesser

A magical game inspired by Akinator that guesses a user's wallet address based on their Twitter handle. This project features a React frontend and a Golang backend that communicate via WebSockets.

## Project Structure

- `frontend/` - React application
- `.` (root) - Golang server application

## Features

- Websocket connection between frontend and backend
- Animated character with different states controlled by the backend
- Twitter handle input and processing

## Setup

### Prerequisites

- Node.js (v16+)
- Go (v1.21+)
- NPM or Yarn

### Backend Setup

1. Install dependencies:
   ```
   go mod download
   ```

2. Run the server:
   ```
   go run cmd/server/main.go
   ```

The server will start on port 8080 by default. You can configure the port using the `PORT` environment variable.

### Frontend Setup

1. Navigate to the frontend directory:
   ```
   cd frontend
   ```

2. Install dependencies:
   ```
   npm install
   ```
   or
   ```
   yarn
   ```

3. Start the development server:
   ```
   npm start
   ```
   or
   ```
   yarn start
   ```

The frontend development server will start on port 3000 by default.

## Environment Variables

### Backend
- `PORT` - Server port (default: 8080)
- `TWITTER_API_KEY` - Twitter API key
- `TWITTER_API_SECRET` - Twitter API secret

### Frontend
- `REACT_APP_WS_URL` - WebSocket server URL (default: ws://localhost:8080/ws)


# Project Structure

```
wallet-guesser/
├── frontend/                  # React frontend
│   ├── public/
│   │   └── assets/            # Asset files (you'll add these)
│   │       ├── background.png
│   │       └── jinn/
│   │           ├── state-idle.png
│   │           ├── state-thinking.png
│   │           ├── state-asking.png
│   │           ├── state-confident.png
│   │           ├── state-correct.png
│   │           ├── state-wrong.png
│   │           └── state-glitched.png
│   ├── src/
│   │   ├── components/        # React components
│   │   │   ├── Game.jsx       # Main game component
│   │   │   ├── Jinn.jsx       # Character component
│   │   │   └── GameControls.jsx # User interaction elements
│   │   ├── services/
│   │   │   └── websocket.js   # WebSocket connection management
│   │   ├── App.jsx           # Root component
│   │   ├── index.jsx         # Entry point
│   │   └── styles/           # CSS/styling files
│   ├── package.json
│   └── README.md
├── cmd/
│   └── server/           # Application entry point
│       └── main.go
├── internal/
│   ├── api/              # API handlers
│   │   └── websocket.go  # WebSocket connection handler
│   ├── config/           # Configuration
│   │   └── config.go
│   ├── game/             # Game logic
│   │   └── state.go      # Game state management
│   └── twitter/          # Twitter interaction
│       └── client.go     # Twitter API client
├── go.mod
├── go.sum
└── README.md                # Project overview
```