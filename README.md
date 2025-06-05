# Samlapa: Real-time Encrypted Chat

## Features
- Real-time chat with WebSocket (Golang backend)
- End-to-end encryption using Signal protocol (browser JS)
- User registration, login, presence, and user listing (REST API)
- Redis Pub/Sub for scalable presence and message delivery

## Project Structure
```
.
├── client/
│   ├── index.html
│   ├── signal.js
│   └── signal-chat.js
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

3. **Open the frontend**  
   Visit [http://localhost:8080](http://localhost:8080) in your browser.

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


