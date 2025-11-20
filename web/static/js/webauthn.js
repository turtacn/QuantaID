/**
 * WebAuthn (FIDO2) client-side logic.
 * This script handles the browser-side part of the WebAuthn ceremony.
 */

// --- Helper Functions for ArrayBuffer/Base64URL conversion ---

// Decodes a Base64URL string into an ArrayBuffer.
function bufferDecode(value) {
    const s = atob(value.replace(/_/g, '/').replace(/-/g, '+'));
    const a = new Uint8Array(s.length);
    for (let i = 0; i < s.length; i++) {
        a[i] = s.charCodeAt(i);
    }
    return a;
}

// Encodes an ArrayBuffer into a Base64URL string.
function bufferEncode(value) {
    return btoa(String.fromCharCode.apply(null, new Uint8Array(value)))
        .replace(/\+/g, '-')
        .replace(/\//g, '_')
        .replace(/=/g, '');
}


/**
 * Converts the credential object received from the browser's WebAuthn API
 * into a JSON-friendly format that can be sent to the server.
 * @param {PublicKeyCredential} cred - The credential object.
 * @returns {object} A JSON-serializable representation of the credential.
 */
function publicKeyCredentialToJSON(cred) {
    if (cred instanceof Array) {
        return cred.map(publicKeyCredentialToJSON);
    }

    if (cred instanceof ArrayBuffer) {
        return bufferEncode(cred);
    }

    if (cred instanceof Object) {
        const obj = {};
        for (const key in cred) {
            obj[key] = publicKeyCredentialToJSON(cred[key]);
        }
        return obj;
    }

    return cred;
}


/**
 * Initiates the registration process for a new WebAuthn credential (e.g., a security key).
 * @param {string} username - The username for which the credential is to be registered.
 */
async function registerWebAuthn(username) {
    try {
        // 1. Get options from the server
        const resp = await fetch('/api/mfa/webauthn/register/begin', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ username: username }), // Pass username if required by your endpoint
        });

        if (!resp.ok) {
            const err = await resp.json();
            throw new Error(err.message || 'Failed to get registration options');
        }

        const options = await resp.json();

        // 2. Decode options from Base64URL to ArrayBuffer for the browser API
        options.publicKey.challenge = bufferDecode(options.publicKey.challenge);
        options.publicKey.user.id = bufferDecode(options.publicKey.user.id);
        if (options.publicKey.excludeCredentials) {
            options.publicKey.excludeCredentials.forEach(cred => {
                cred.id = bufferDecode(cred.id);
            });
        }

        // 3. Prompt the browser/user to create a new credential
        const credential = await navigator.credentials.create({
            publicKey: options.publicKey
        });

        // 4. Encode the new credential and send it to the server to be stored
        const credentialJSON = publicKeyCredentialToJSON(credential);

        const finishResp = await fetch('/api/mfa/webauthn/register/finish', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(credentialJSON),
        });

        if (!finishResp.ok) {
            const err = await finishResp.json();
            throw new Error(err.message || 'Failed to finish registration');
        }

        alert("Security Key registered successfully!");
        window.location.reload(); // Reload to show the new credential

    } catch (err) {
        console.error("WebAuthn registration failed:", err);
        alert(`Error: ${err.message}`);
    }
}


/**
 * Initiates the authentication process using an existing WebAuthn credential.
 * @param {string} username - The username attempting to log in.
 */
async function loginWithWebAuthn(username) {
    try {
        // 1. Get options from the server
        const resp = await fetch('/api/mfa/webauthn/login/begin', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ username: username }),
        });

        if (!resp.ok) {
            const err = await resp.json();
            throw new Error(err.message || 'Failed to get login options');
        }

        const options = await resp.json();

        // 2. Decode options for the browser API
        options.publicKey.challenge = bufferDecode(options.publicKey.challenge);
        if (options.publicKey.allowCredentials) {
            options.publicKey.allowCredentials.forEach(cred => {
                cred.id = bufferDecode(cred.id);
            });
        }

        // 3. Prompt the user to use their security key
        const assertion = await navigator.credentials.get({
            publicKey: options.publicKey
        });

        // 4. Encode the result and send it to the server for verification
        const assertionJSON = publicKeyCredentialToJSON(assertion);

        const finishResp = await fetch('/api/mfa/webauthn/login/finish', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(assertionJSON),
        });

        if (finishResp.ok) {
            alert("Login successful!");
            // Redirect the user to their dashboard or the next step in the flow
            window.location.href = "/dashboard";
        } else {
            const err = await finishResp.json();
            throw new Error(err.message || 'WebAuthn authentication failed');
        }

    } catch (err) {
        console.error("WebAuthn login failed:", err);
        alert(`Error: ${err.message}`);
    }
}
