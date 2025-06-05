import * as libsignal from 'https://cdn.jsdelivr.net/npm/libsignal-protocol/dist/libsignal-protocol.js';
import { generateAndUploadKeys, fetchUserKeys } from './signal.js';

const store = new libsignal.SignalProtocolStore();

export async function setupUser(username) {
  // Generate and upload keys if not already done
  const keys = await generateAndUploadKeys(username);
  await store.put('identityKey', keys.identityKeyPair);
  await store.put('registrationId', keys.registrationId);
  await store.put('signedPreKey', keys.signedPreKey);
  for (let i = 0; i < keys.preKeys.length; i++) {
    await store.put('preKey' + keys.preKeys[i].keyId, keys.preKeys[i]);
  }
  return keys;
}

export async function encryptMessage(sender, recipient, plaintext) {
  // Fetch recipient keys
  const remoteKeys = await fetchUserKeys(recipient);
  // Convert base64 to Uint8Array
  function b64ToArr(b64) {
    return new Uint8Array(atob(b64).split('').map(c => c.charCodeAt(0)));
  }
  const address = new libsignal.SignalProtocolAddress(recipient, 1);
  const sessionBuilder = new libsignal.SessionBuilder(store, address);
  await sessionBuilder.processPreKey({
    registrationId: 1,
    identityKey: b64ToArr(remoteKeys.identityKey),
    signedPreKey: {
      keyId: 1,
      publicKey: b64ToArr(remoteKeys.signedPreKey),
      signature: new Uint8Array(64)
    },
    preKey: {
      keyId: 1,
      publicKey: b64ToArr(remoteKeys.preKeys[0])
    }
  });
  const sessionCipher = new libsignal.SessionCipher(store, address);
  const ciphertext = await sessionCipher.encrypt(plaintext);
  return btoa(String.fromCharCode.apply(null, ciphertext.body));
}

export async function decryptMessage(sender, ciphertextB64) {
  const address = new libsignal.SignalProtocolAddress(sender, 1);
  const sessionCipher = new libsignal.SessionCipher(store, address);
  const ciphertext = {
    type: 3,
    body: new Uint8Array(atob(ciphertextB64).split('').map(c => c.charCodeAt(0)))
  };
  return await sessionCipher.decryptPreKeyWhisperMessage(ciphertext.body, 'binary');
}