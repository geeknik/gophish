# OAuth 2.0 IMAP Configuration

## Overview

Gophish supports OAuth 2.0 for IMAP reply tracking, specifically designed for Microsoft 365 integration. This implementation uses the Azure AD client credentials flow, which is an app-only authentication method requiring no user interaction.

This configuration is required for organizations that have disabled basic authentication on Exchange Online, as Microsoft has deprecated legacy authentication methods.

## Azure AD Setup

To configure OAuth 2.0 for Gophish, you must first set up an application in Azure Active Directory (Entra ID). Follow these steps:

### 1. Register Application
1. Log in to the [Azure Portal](https://portal.azure.com).
2. Navigate to **App registrations**.
3. Click **New registration**.
4. Enter a name (e.g., "Gophish IMAP Monitor") and register the application.
5. Note the **Application (client) ID** and **Directory (tenant) ID** from the Overview page.

### 2. Create Client Secret
1. Select **Certificates & secrets** from the left menu.
2. Click **New client secret**.
3. Add a description and select an expiration period.
4. Click **Add**.
5. **Important:** Copy the **Value** of the client secret immediately. You will not be able to see it again.

### 3. Configure API Permissions
1. Select **API permissions** from the left menu.
2. Click **Add a permission**.
3. Select **Office 365 Exchange Online**.
4. Select **Application permissions**.
5. Check **IMAP.AccessAsApp**.
6. Click **Add permissions**.

### 4. Grant Admin Consent
1. On the **API permissions** page, click **Grant admin consent for [Your Organization]**.
2. Confirm the action to ensure the permissions are active.

### 5. Configure Service Principal Access
You must grant the application access to the specific mailbox Gophish will monitor. This is done via Exchange Online PowerShell.

```powershell
# Register the Service Principal for your application
New-ServicePrincipal -AppId <app-id> -ServiceId <object-id>

# Grant FullAccess permissions to the specific mailbox
Add-MailboxPermission -Identity "shared@example.com" -User <service-principal-id> -AccessRights FullAccess
```

*Note: `<app-id>` is your Application (client) ID. `<object-id>` is the Object ID of the Service Principal (found in Enterprise Applications, not App Registrations).*

## Gophish Configuration

Update your Gophish IMAP settings to include the OAuth configuration parameters. The configuration uses a JSON structure.

### OAuth Fields
Ensure the following fields are set in your configuration:

*   `oauth_enabled`: Set to `true` to enable OAuth.
*   `oauth_client_id`: The **Application (client) ID** from Azure.
*   `oauth_client_secret`: The **Client secret value** generated in step 2. (Note: This will be encrypted at rest by Gophish).
*   `oauth_tenant_id`: The **Directory (tenant) ID** from Azure.

### Standard Fields
You must still provide the standard IMAP connection details:
*   `host`: `outlook.office365.com:993`
*   `username`: The email address of the mailbox being monitored (e.g., `shared@example.com`).

**Example Configuration Snippet:**
```json
{
    "host": "outlook.office365.com:993",
    "username": "shared@example.com",
    "oauth_enabled": true,
    "oauth_client_id": "00000000-0000-0000-0000-000000000000",
    "oauth_client_secret": "your-client-secret-value",
    "oauth_tenant_id": "00000000-0000-0000-0000-000000000000"
}
```

## Troubleshooting

### Token Errors
*   **Symptom:** Authentication fails with token retrieval errors.
*   **Solution:** Verify that the Client Secret has not expired. If it has, generate a new one in the Azure Portal and update the Gophish configuration.

### Permission Denied
*   **Symptom:** Gophish connects but fails to access the mailbox (e.g., "AUTHENTICATE failed").
*   **Solution:**
    1. Ensure **Grant admin consent** was clicked in the API Permissions section.
    2. Verify the PowerShell commands were run successfully to grant the Service Principal access to the specific mailbox.

### Connection Failed
*   **Symptom:** Timeout or inability to connect to the host.
*   **Solution:**
    1. Verify the host is `outlook.office365.com:993`.
    2. Ensure IMAP is enabled for the target mailbox in the Microsoft 365 Admin Center (Users > Active users > Mail > Manage email apps).
