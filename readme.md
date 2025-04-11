# Wallet Guesser

A magical game inspired by Akinator that guesses a user's wallet address based on their Twitter handle. This project features a React frontend and a Golang backend that communicate via WebSockets.

## Project Structure

The project is organized using clean architecture principles with clear separation of concerns:

- `cmd/` - Entry points for the application
   - `server/` - The main server application
   - `updateavoidlist/` - Command to update the avoid list
- `frontend/` - React application
- `internal/` - Backend application code with clear domain boundaries
   - `api/` - API endpoints and handlers
   - `avoidlist/` - Services for managing the avoid list
   - `blockchain/` - Blockchain client and utilities
   - `config/` - Configuration management
   - `domain/` - Domain models and interfaces
   - `game/` - Game logic
   - `twitter/` - Twitter client and utilities

## Features

- WebSocket connection between frontend and backend
- Animated character with different states controlled by the backend
- Twitter handle input and processing
- Avoid list for filtering spammy addresses
- Caching of results for better performance

## Setup

### Prerequisites

- Node.js (v16+)
- Go (v1.21+)
- NPM or Yarn

### Backend Setup

1. Copy the example environment file and configure it:
   ```
   cp .env.example .env
   ```
   Then edit the `.env` file to add your API keys.

2. Install dependencies:
   ```
   go mod download
   ```

3. Update the avoid list:
   ```
   go run cmd/updateavoidlist/main.go
   ```

4. Run the server:
   ```
   go run cmd/server/main.go
   ```

The server will start on port 8080 by default. You can configure the port using the `.env` file.

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
- `DEBUG` - Enable debug logging (default: false)
- `APIFY_TOKEN` - Apify API token for Twitter data
- `DUNE_API_KEY` - Dune Analytics API key for avoid list
- `SOLANA_RPC_ENDPOINT` - Solana RPC endpoint (default: https://api.mainnet-beta.solana.com)
- `AVOID_LIST_PATH` - Path to the avoid list file (default: data/avoidlist.json)

### Frontend
- `REACT_APP_WS_URL` - WebSocket server URL (default: ws://localhost:8080/ws)

## Avoid List

The avoid list helps filter out spammy addresses:
- Wallets that hold more than 500 tokens
- Tokens that have more than 100,000 holders

To update the avoid list:

```
go run cmd/updateavoidlist/main.go
```

This command fetches the latest data from Dune Analytics and updates the local avoid list file.

## API Documentation

### WebSocket API

The frontend and backend communicate via WebSocket messages with the following format:

```json
{
  "type": "MESSAGE_TYPE",
  "payload": {}
}
```

Message types:
- `START_GAME` - Initialize a new game
- `USER_INPUT` - Send user input (Twitter handle)
- `JINN_STATE` - Update the Jinn character's state
- `PROGRESS_UPDATE` - Send progress updates
- `WALLET_RESULT` - Send the wallet guess result

## Adding New Features

The project is designed with clean architecture principles, making it easy to add new features:

1. Define new domain models and interfaces in `internal/domain/`
2. Implement the interfaces in the appropriate packages
3. Wire everything together in `cmd/server/main.go`

## License

This project is licensed under the MIT License - see the LICENSE file for details.