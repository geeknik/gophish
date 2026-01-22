# DKIM Configuration

DomainKeys Identified Mail (DKIM) is an email authentication method that helps prevent spammers and other malicious parties from impersonating a legitimate domain. It provides a cryptographic proof that an email was sent by a specific domain and that the message body has not been altered in transit.

Enabling DKIM in Gophish can significantly improve the deliverability of your phishing simulation emails, ensuring they land in the inbox rather than the spam folder.

## Prerequisites

Before configuring DKIM in Gophish, you need:
1. Access to your domain's DNS records.
2. `openssl` installed on your machine to generate keys.

## 1. Generate DKIM Key Pair

You need a private key for signing emails and a public key for your DNS records.

Generate a 2048-bit RSA private key:

```bash
openssl genrsa -out dkim_private.pem 2048
```

Extract the public key from the private key:

```bash
openssl rsa -in dkim_private.pem -pubout -out dkim_public.pem
```

## 2. Configure DNS

You need to add a TXT record to your domain's DNS.

* **Selector**: Choose a selector name (e.g., `gophish1`). This allows you to have multiple DKIM keys for the same domain.
* **Host/Name**: `[selector]._domainkey` (e.g., `gophish1._domainkey`).
* **Value**: `v=DKIM1; k=rsa; p=[content of dkim_public.pem]`

**Steps:**
1. Open `dkim_public.pem`.
2. Copy the base64 string between `-----BEGIN PUBLIC KEY-----` and `-----END PUBLIC KEY-----`. Remove all newlines so it is a single line string.
3. Create the TXT record.

**Example DNS Record:**

| Type | Host | Value |
|------|------|-------|
| TXT | `gophish1._domainkey` | `v=DKIM1; k=rsa; p=MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA...` |

## 3. Configure Gophish

Currently, DKIM configuration is supported via the Gophish API.

**API Endpoint:** `POST /api/smtp/` (to create) or `PUT /api/smtp/{id}` (to update)

**JSON Payload:**

Include the following fields in your SMTP profile configuration:

```json
{
  "name": "Corporate SMTP",
  "host": "smtp.example.com:587",
  "username": "user@example.com",
  "password": "password",
  "from_address": "security@example.com",
  "ignore_cert_errors": false,
  "dkim_enabled": true,
  "dkim_domain": "example.com",
  "dkim_selector": "gophish1",
  "dkim_private_key": "-----BEGIN RSA PRIVATE KEY-----\nMIIEowIBAAKCAQEA...\n...-----END RSA PRIVATE KEY-----"
}
```

> **Note:** The `dkim_private_key` must be the full content of your `dkim_private.pem` file, including the header and footer lines. Gophish will encrypt this key at rest if an encryption key is configured.

## 4. Verify Configuration

1. **Check DNS**: Use `dig` to verify your DNS record is propagated.
   ```bash
   dig TXT gophish1._domainkey.example.com
   ```
2. **Send Test Email**: Send a test email from Gophish to an external address (like Gmail).
3. **Inspect Headers**: View the "Original" message or headers in the receiving inbox. Look for `Authentication-Results` passing DKIM and a `DKIM-Signature` header.
