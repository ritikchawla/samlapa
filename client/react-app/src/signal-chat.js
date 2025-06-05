// src/signal-chat.js
import CryptoJS from "crypto-js";

// Shared secret for demonstration (in production, use per-user keys)
const SECRET = "demo-shared-secret";

export async function setupUser(username) {
  // No-op for simple encryption
  return { username };
}

export async function encryptMessage(sender, recipient, plaintext) {
  // Encrypt with AES using shared secret
  return CryptoJS.AES.encrypt(plaintext, SECRET).toString();
}

export async function decryptMessage(sender, ciphertext) {
  // Decrypt with AES using shared secret
  const bytes = CryptoJS.AES.decrypt(ciphertext, SECRET);
  return bytes.toString(CryptoJS.enc.Utf8);
}