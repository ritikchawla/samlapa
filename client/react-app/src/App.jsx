import React, { useState, useEffect, useRef } from 'react';
import './App.css';

// TODO: Move these modules to src/ and import as local modules
import { generateAndUploadKeys, fetchUserKeys } from './signal.js';
import { encryptMessage, decryptMessage, setupUser } from './signal-chat.js';

function App() {
  const [username, setUsername] = useState('');
  const [connected, setConnected] = useState(false);
  const [ws, setWs] = useState(null);
  const [messages, setMessages] = useState([]);
  const [typingUsers, setTypingUsers] = useState({});
  const [input, setInput] = useState('');
  const inputRef = useRef();

  useEffect(() => {
    if (!connected) return;
    const socket = new WebSocket(`ws://${window.location.hostname}:8080/ws`);
    setWs(socket);

    socket.onopen = () => console.log('Connected');
    socket.onmessage = async (evt) => {
      const env = JSON.parse(evt.data);
      if (env.type === 'typing') {
        setTypingUsers(t => ({ ...t, [env.sender]: true }));
        setTimeout(() => {
          setTypingUsers(t => {
            const copy = { ...t };
            delete copy[env.sender];
            return copy;
          });
        }, 1000);
      } else if (env.type === 'message') {
        // Decrypt message
        let text = env.content;
        try {
          text = await decryptMessage(env.sender, env.content);
        } catch (e) {
          text = '[decryption failed]';
        }
        setMessages(msgs => [...msgs, { sender: env.sender, text }]);
      }
    };
    socket.onclose = () => setConnected(false);

    return () => socket.close();
  }, [connected]);

  const handleConnect = async () => {
    await setupUser(username);
    setConnected(true);
  };

  const handleSend = async (e) => {
    e.preventDefault();
    if (!input.trim() || !ws) return;
    // Encrypt message
    let encrypted = input;
    try {
      encrypted = await encryptMessage(username, 'all', input);
    } catch (e) {
      encrypted = '[encryption failed]';
    }
    ws.send(JSON.stringify({ type: 'message', sender: username, content: encrypted }));
    setInput('');
  };

  const handleTyping = () => {
    if (ws) {
      ws.send(JSON.stringify({ type: 'typing', sender: username }));
    }
  };

  return (
    <div id="chat">
      {!connected ? (
        <form onSubmit={e => { e.preventDefault(); handleConnect(); }}>
          <input
            placeholder="Enter username"
            value={username}
            onChange={e => setUsername(e.target.value)}
            required
          />
          <button type="submit">Connect</button>
        </form>
      ) : (
        <>
          <div style={{ maxHeight: 400, overflowY: 'auto', border: '1px solid #eee', marginBottom: 8 }}>
            {messages.map((msg, i) => (
              <div key={i} className="message">
                <b>{msg.sender}:</b> {msg.text}
              </div>
            ))}
            {Object.keys(typingUsers).map(user => (
              <div key={user} className="typing">{user} is typing...</div>
            ))}
          </div>
          <form onSubmit={handleSend}>
            <input
              ref={inputRef}
              value={input}
              onChange={e => setInput(e.target.value)}
              onKeyDown={handleTyping}
              placeholder="Type a message"
              autoFocus
              required
            />
            <button type="submit">Send</button>
          </form>
        </>
      )}
    </div>
  );
}

export default App;
