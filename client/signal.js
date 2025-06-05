import * as libsignal from 'https://cdn.jsdelivr.net/npm/libsignal-protocol/dist/libsignal-protocol.js';

export async function generateAndUploadKeys(username) {
  const identityKeyPair = await libsignal.KeyHelper.generateIdentityKeyPair();
  const registrationId = await libsignal.KeyHelper.generateRegistrationId();
  const signedPreKey = await libsignal.KeyHelper.generateSignedPreKey(identityKeyPair, 1);
  const preKeys = [];
  for (let i = 1; i <= 5; i++) {
    preKeys.push(await libsignal.KeyHelper.generatePreKey(i));
  }

  // Convert keys to base64 for transport
  function keyToBase64(key) {
    return btoa(String.fromCharCode.apply(null, key.keyPair.pubKey || key.pubKey));
  }
  function signedKeyToBase64(key) {
    return btoa(String.fromCharCode.apply(null, key.keyPair.pubKey));
  }

  const payload = {
    identityKey: keyToBase64(identityKeyPair),
    signedPreKey: signedKeyToBase64(signedPreKey),
    preKeys: preKeys.map(keyToBase64)
  };

  await fetch(`/api/upload-keys?username=${encodeURIComponent(username)}`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload)
  });

  return { identityKeyPair, registrationId, signedPreKey, preKeys };
}

export async function fetchUserKeys(username) {
  const res = await fetch(`/api/get-keys?username=${encodeURIComponent(username)}`);
  if (!res.ok) throw new Error('Key fetch failed');
  return await res.json();
}