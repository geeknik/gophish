# Security Features

Gophish includes several security features designed to protect sensitive data and improve Operational Security (OPSEC) during engagements. This document outlines these features and how to configure them.

## Encryption at Rest

Gophish supports AES-256-GCM encryption for sensitive fields in the database. When configured, the following data is encrypted before being stored:

*   SMTP Passwords
*   IMAP Passwords
*   DKIM Private Keys

### Configuration

To enable encryption, add the `encryption_key` field to your `config.json`. The key must be a 32-byte value, which can be provided as a hex-encoded string or base64-encoded string.

You can generate a secure 32-byte hex key using OpenSSL:

```bash
openssl rand -hex 32
```

Add the generated key to your `config.json`:

```json
{
    "encryption_key": "YOUR_GENERATED_KEY_HERE",
    ...
}
```

### How It Works

*   **Encryption**: Data is encrypted using AES-256-GCM.
*   **Storage**: Encrypted fields in the database are prefixed with `ENC:` to distinguish them from plaintext data.
*   **Backward Compatibility**: The system checks for the `ENC:` prefix. If the prefix is missing, the data is treated as plaintext. This allows existing unencrypted data to continue working alongside encrypted data if a key is added later.

## Session Security

Gophish implements several measures to secure user sessions and the web interface:

*   **Session Invalidation**: Logging out correctly invalidates the session cookie by setting its `MaxAge` to -1, ensuring the cookie cannot be reused.
*   **API Key Protection**: The user's API key is not exposed in the DOM or HTML source of the web interface, mitigating the risk of key theft via XSS.
*   **Authentication Fallback**: API endpoints used by the web UI (e.g., in JavaScript) support session-based authentication as a fallback, removing the need to embed API keys in client-side code.

## Operational Security (OPSEC)

To reduce the likelihood of detection by defensive systems, Gophish provides configurable options to mask its identity.

### Server Fingerprinting

You can customize the HTTP headers and cookies to look like standard web server traffic instead of a phishing framework.

*   **`server_name`**: Sets the `Server` HTTP header.
    *   Default: `Apache/2.4.41 (Ubuntu)`
*   **`session_cookie_name`**: Sets the name of the session cookie.
    *   Default: `PHPSESSID`

### Webhook Signatures

Webhooks sent by Gophish use the generic header `X-Webhook-Signature` instead of `X-Gophish-Signature` to avoid simple header-based detection rules.

### TLS Certificates

When Gophish generates self-signed certificates (e.g., for the administration interface), the Organization (O) field is set to a generic "Web Server" rather than "Gophish".

### Email Indicators

The default subject line for configuration test emails is "Configuration Test Email". It does not contain the word "Gophish".

### Example Configuration

Here is an example `config.json` snippet demonstrating the security and OPSEC configuration options:

```json
{
	"admin_server": {
		"listen_url": "127.0.0.1:3333",
		"use_tls": true,
		"cert_path": "gophish_admin.crt",
		"key_path": "gophish_admin.key"
	},
	"phish_server": {
		"listen_url": "0.0.0.0:80",
		"use_tls": false,
		"cert_path": "example.crt",
		"key_path": "example.key"
	},
	"db_name": "sqlite3",
	"db_path": "gophish.db",
	"migrations_prefix": "db/db_",
	"contact_address": "",
	"logging": {
		"filename": "",
		"level": ""
	},
	"encryption_key": "d4e5f6...32_byte_hex_key...",
	"server_name": "nginx/1.18.0",
	"session_cookie_name": "JSESSIONID"
}
```
