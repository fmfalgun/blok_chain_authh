const fs = require('fs');
const crypto = require('crypto');

// Generate a new key pair in PKCS1 format (different from default PKCS8)
const { publicKey, privateKey } = crypto.generateKeyPairSync('rsa', {
  modulusLength: 2048,
  publicKeyEncoding: {
    type: 'spki',
    format: 'pem'
  },
  privateKeyEncoding: {
    type: 'pkcs1', // Use PKCS1 instead of PKCS8
    format: 'pem'
  }
});

// Save the keys to files
fs.writeFileSync('test-private.pem', privateKey);
fs.writeFileSync('test-public.pem', publicKey);

console.log('Test keys generated:');
console.log('Private key (PKCS1):', privateKey.substring(0, 100) + '...');
console.log('Public key:', publicKey.substring(0, 100) + '...');

// Test signing
const testData = Buffer.from('test data');
const hash = crypto.createHash('sha256').update(testData).digest();

// Attempt to sign with the private key
const signature = crypto.sign('sha256', hash, {
  key: privateKey,
  padding: crypto.constants.RSA_PKCS1_PADDING
});

console.log('Signature created successfully, length:', signature.length);

// Verify the signature
const verified = crypto.verify(
  'sha256',
  hash,
  {
    key: publicKey,
    padding: crypto.constants.RSA_PKCS1_PADDING
  },
  signature
);

console.log('Signature verification result:', verified);
