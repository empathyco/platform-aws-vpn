
function encodePEM (pemType, ab) {
  let b64Encoded = btoa(String.fromCharCode(...new Uint8Array(ab)))
  let pemEncoded = b64Encoded.match(/.{1,64}/g).join('\n') + '='.repeat(b64Encoded.length % 4)
  return `-----BEGIN ${pemType}-----\n${pemEncoded}\n-----END ${pemType}-----\n`
}

export async function newKeyPair () {
  let keyParams
  switch (process.env.VUE_APP_KEY_TYPE) {
    case 'RSA':
      keyParams = { name: 'RSA-PSS', modulusLength: 2048, publicExponent: new Uint8Array([1, 0, 1]), hash: 'SHA-256' }
      break
    case 'EC':
      keyParams = { name: 'ECDSA', namedCurve: 'P-256' }
      break
    default:
      throw Error('invalid Key Type: ' + process.env.VUE_APP_KEY_TYPE)
  }

  return window.crypto.subtle.generateKey(keyParams, true, ['sign', 'verify'])
}

export async function exportKeys (keyPair) {
  let publicKey = await window.crypto.subtle.exportKey('jwk', keyPair.publicKey)
  let privateKey = await window.crypto.subtle.exportKey('pkcs8', keyPair.privateKey)

  return { 'public': publicKey, 'private': encodePEM(publicKey.kty + ' PRIVATE KEY', privateKey) }
}
