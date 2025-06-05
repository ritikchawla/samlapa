# Samlapa: Real-time Encrypted Chat

## Features
- Real-time chat with WebSocket (Golang backend)
- Encrypted messaging using AES (crypto-js) in React frontend
- User registration, login, presence, and user listing (REST API)
- Redis Pub/Sub for scalable presence and message delivery
- Modern React frontend (Vite)

## Project Structure
```
.
├── client/
│   ├── index.html
│   ├── signal.js
│   ├── signal-chat.js
│   └── react-app/
│       ├── src/
│       │   ├── App.jsx
│       │   ├── signal.js
│       │   └── signal-chat.js
│       └── ...
└── server/
    ├── main.go
    ├── go.mod
    ├── go.sum
    └── api_test.go
```

## Running the App

1. **Start Redis**  
   On macOS:  
   ```
   brew install redis
   brew services start redis
   ```

2. **Start the Go backend**  
   ```
   cd server
   go run main.go
   ```

3. **Start the React frontend**
   ```
   cd client/react-app
   npm install
   npm run dev
   ```
   Visit [http://localhost:5173](http://localhost:5173) in your browser.

   > The backend and frontend must both be running for full functionality.

## API Endpoints

### REST
- `POST /api/register` — `{ "username": "...", "password": "..." }`
- `POST /api/login` — `{ "username": "...", "password": "..." }`
- `GET /api/users`
- `POST /api/presence?username=...&online=true|false`
- `POST /api/upload-keys?username=...` — Signal protocol public keys
- `GET /api/get-keys?username=...` — Fetch Signal keys

### WebSocket
- `/ws` — Real-time chat (encrypted payloads)


