# File Sharing Guide

## Overview

Share files and folders with anyone using secure, tokenized links. Features include:

- 🔗 **Shareable Links** - Generate unique URLs for files/folders
- 🔒 **Password Protection** - Optional password for secure sharing
- ⏰ **Expiration** - Set time limits on share links
- 📁 **Folder Sharing** - Share entire folders with contents
- 🔄 **Link Management** - Update, delete, and list all your shares

## API Endpoints

### 1. Create Share Link (Authenticated)

Generate a shareable link for a file or folder.

```http
POST /api/shares
Cookie: AuthToken=<your-token>
Content-Type: application/json

{
  "fileId": "file-uuid",
  "expiresIn": 24,           // Optional: hours until expiration
  "password": "secret123"    // Optional: password protection
}
```

**Response:**

```json
{
  "message": "Share link created successfully",
  "shareId": "share-uuid",
  "shareUrl": "http://localhost:3000/share/a1b2c3d4...",
  "token": "a1b2c3d4e5f6...",
  "fileName": "document.pdf",
  "isFolder": false,
  "expiresAt": "2025-10-04T10:00:00Z",
  "passwordProtected": true
}
```

### 2. Access Shared File (Public)

Download a file via share link. **No authentication required!**

```http
GET /share/{token}?password=secret123

# Password is optional, only needed if share is password-protected
```

**Response:** File download or folder contents (JSON)

### 3. Get Share Info (Public)

Get information about a share without downloading.

```http
GET /api/share/{token}/info
```

**Response:**

```json
{
  "fileName": "document.pdf",
  "fileSize": 1024000,
  "isFolder": false,
  "expiresAt": "2025-10-04T10:00:00Z",
  "passwordProtected": true,
  "createdAt": "2025-10-03T10:00:00Z"
}
```

### 4. List Your Shares (Authenticated)

Get all share links you've created.

```http
GET /api/shares
Cookie: AuthToken=<your-token>
```

**Response:**

```json
[
  {
    "id": "share-uuid",
    "token": "a1b2c3d4...",
    "fileId": "file-uuid",
    "fileName": "document.pdf",
    "isFolder": false,
    "expiresAt": "2025-10-04T10:00:00Z",
    "passwordProtected": true,
    "createdAt": "2025-10-03T10:00:00Z",
    "shareUrl": "http://localhost:3000/share/a1b2c3d4..."
  }
]
```

### 5. Update Share Link (Authenticated)

Extend expiration or change password.

```http
PUT /api/shares/{shareId}
Cookie: AuthToken=<your-token>
Content-Type: application/json

{
  "expiresIn": 48,           // Optional: extend for 48 more hours
  "password": "newsecret"    // Optional: change password (empty string to remove)
}
```

### 6. Delete Share Link (Authenticated)

Revoke a share link.

```http
DELETE /api/shares/{shareId}
Cookie: AuthToken=<your-token>
```

---

## Client Examples

### JavaScript: Create Share Link

```javascript
async function createShareLink(fileId, options = {}) {
  const response = await fetch("/api/shares", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    credentials: "include",
    body: JSON.stringify({
      fileId: fileId,
      expiresIn: options.expiresIn || null, // Hours
      password: options.password || null,
    }),
  });

  const data = await response.json();
  console.log("Share URL:", data.shareUrl);
  return data;
}

// Usage
const share = await createShareLink("file-uuid-123", {
  expiresIn: 24,
  password: "secret123",
});

// Copy to clipboard
navigator.clipboard.writeText(share.shareUrl);
```

### JavaScript: Access Shared File

```javascript
// Download file directly (browser handles authentication)
function downloadSharedFile(token, password = null) {
  const url = password
    ? `/share/${token}?password=${encodeURIComponent(password)}`
    : `/share/${token}`;

  window.open(url, "_blank");
}

// Get info first, then download
async function getShareInfo(token) {
  const response = await fetch(`/api/share/${token}/info`);
  const info = await response.json();

  if (info.passwordProtected) {
    const password = prompt("Enter password:");
    downloadSharedFile(token, password);
  } else {
    downloadSharedFile(token);
  }
}
```

### JavaScript: List and Manage Shares

```javascript
async function listMyShares() {
  const response = await fetch("/api/shares", {
    credentials: "include",
  });

  const shares = await response.json();
  return shares;
}

async function deleteShare(shareId) {
  await fetch(`/api/shares/${shareId}`, {
    method: "DELETE",
    credentials: "include",
  });
}

async function extendShare(shareId, hours) {
  await fetch(`/api/shares/${shareId}`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    credentials: "include",
    body: JSON.stringify({ expiresIn: hours }),
  });
}
```

### curl: Create and Use Share

```bash
# 1. Login
curl -X POST http://localhost:3000/api/user/login \
  -H "Content-Type: application/json" \
  -c cookies.txt \
  -d '{"emailOrUsername":"user","password":"pass"}'

# 2. Create share link
curl -X POST http://localhost:3000/api/shares \
  -b cookies.txt \
  -H "Content-Type: application/json" \
  -d '{
    "fileId": "file-uuid",
    "expiresIn": 24,
    "password": "secret123"
  }'

# Response includes shareUrl and token

# 3. Access shared file (no cookies needed!)
curl "http://localhost:3000/share/TOKEN?password=secret123" \
  -o downloaded-file.pdf

# 4. Get share info
curl "http://localhost:3000/api/share/TOKEN/info"

# 5. List your shares
curl http://localhost:3000/api/shares -b cookies.txt

# 6. Delete share
curl -X DELETE http://localhost:3000/api/shares/SHARE_ID -b cookies.txt
```

### Python: Share Management

```python
import requests

session = requests.Session()

# Login
session.post('http://localhost:3000/api/user/login', json={
    'emailOrUsername': 'user',
    'password': 'password'
})

# Create share
response = session.post('http://localhost:3000/api/shares', json={
    'fileId': 'file-uuid',
    'expiresIn': 24,
    'password': 'secret123'
})
share = response.json()
print(f"Share URL: {share['shareUrl']}")

# List shares
shares = session.get('http://localhost:3000/api/shares').json()
for s in shares:
    print(f"{s['fileName']}: {s['shareUrl']}")

# Download shared file (no authentication!)
file_response = requests.get(f"http://localhost:3000/share/{share['token']}",
                              params={'password': 'secret123'})
with open('downloaded.pdf', 'wb') as f:
    f.write(file_response.content)
```

---

## Use Cases

### 1. Temporary File Sharing

Share a file for 24 hours:

```javascript
const share = await createShareLink(fileId, { expiresIn: 24 });
// Send share.shareUrl to recipient via email/chat
```

### 2. Password-Protected Documents

Share sensitive documents with password:

```javascript
const share = await createShareLink(fileId, {
  password: "confidential2024",
  expiresIn: 48, // 2 days
});
```

### 3. Public File Distribution

Share without password or expiration:

```javascript
const share = await createShareLink(fileId);
// Permanent link (until manually revoked)
```

### 4. Folder Sharing

Share entire folder with contents:

```javascript
// First, create folder and upload files
const folderId = await createFolder("Project Files");

// Share the folder
const share = await createShareLink(folderId);

// Recipients get folder contents as JSON
// Can then download individual files
```

### 5. Link Rotation

Periodically update share links for security:

```javascript
// Delete old share
await deleteShare(oldShareId);

// Create new share
const newShare = await createShareLink(fileId, { expiresIn: 24 });
```

---

## Security Features

### 1. Tokenization

- 256-bit random tokens (64 hex characters)
- Cryptographically secure random generation
- Impossible to guess or brute force

### 2. Password Protection

- Passwords hashed with bcrypt
- Never stored or transmitted in plain text
- Optional per-share basis

### 3. Expiration

- Automatic expiration after set duration
- Server-side validation on every access
- Expired links return 410 Gone

### 4. Access Control

- Only file owners can create shares
- Only owners can delete/update shares
- File permissions inherited from owner

### 5. Audit Trail

- Created timestamp for every share
- Can track who created which shares
- Easy to revoke access by deleting share

---

## Best Practices

### For Users

1. **Use Expiration** - Always set expiration for temporary sharing

   ```javascript
   {
     expiresIn: 24;
   } // 24 hours
   ```

2. **Password Protect Sensitive Files** - Add passwords for confidential data

   ```javascript
   {
     password: "strong-password-here";
   }
   ```

3. **Revoke After Use** - Delete shares when no longer needed

   ```javascript
   await deleteShare(shareId);
   ```

4. **Check Before Sharing** - Verify what you're sharing
   ```javascript
   const info = await getShareInfo(token);
   console.log(info.fileName); // Confirm it's the right file
   ```

### For Developers

1. **Configure BASE_URL** - Set in environment variables

   ```bash
   BASE_URL=https://yourdomain.com
   ```

2. **Monitor Share Usage** - Track download counts (future enhancement)

3. **Clean Expired Shares** - Implement cron job to delete expired shares

   ```sql
   DELETE FROM share_links WHERE expires_at < NOW();
   ```

4. **Rate Limiting** - Add rate limits to prevent abuse
   ```go
   app.Use(limiter.New(limiter.Config{
     Max: 100,
     Duration: time.Minute,
   }))
   ```

---

## Database Schema

### share_links Table

```sql
CREATE TABLE share_links (
    id TEXT PRIMARY KEY,
    file_id TEXT NOT NULL REFERENCES files(id) ON DELETE CASCADE,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMP,
    password TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

**Indexes:**

- `token` (unique) - Fast lookup by share URL
- `file_id` - Check if file has shares
- `user_id` - List user's shares

---

## UI Component Example

### Share Button Component (React)

```jsx
import { useState } from "react";

function ShareButton({ fileId, fileName }) {
  const [shareUrl, setShareUrl] = useState(null);
  const [showDialog, setShowDialog] = useState(false);

  const handleShare = async () => {
    const response = await fetch("/api/shares", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      credentials: "include",
      body: JSON.stringify({
        fileId: fileId,
        expiresIn: 24,
      }),
    });

    const data = await response.json();
    setShareUrl(data.shareUrl);
    setShowDialog(true);
  };

  const copyToClipboard = () => {
    navigator.clipboard.writeText(shareUrl);
    alert("Link copied!");
  };

  return (
    <>
      <button onClick={handleShare}>Share {fileName}</button>

      {showDialog && (
        <div className="share-dialog">
          <h3>Share Link</h3>
          <input value={shareUrl} readOnly />
          <button onClick={copyToClipboard}>Copy Link</button>
          <p>Link expires in 24 hours</p>
        </div>
      )}
    </>
  );
}
```

---

## Troubleshooting

### Share Link Not Working

**Problem:** 404 Not Found when accessing share link

**Solutions:**

1. Check token is correct (64 hex characters)
2. Verify share hasn't been deleted
3. Check if share expired

```javascript
// Get info to debug
const info = await fetch(`/api/share/${token}/info`).then((r) => r.json());
console.log(info);
```

### Password Not Accepted

**Problem:** 401 Invalid password

**Solutions:**

1. Verify password is correct (case-sensitive)
2. Check if password was recently changed
3. URL encode password in query string

```javascript
const encodedPassword = encodeURIComponent(password);
const url = `/share/${token}?password=${encodedPassword}`;
```

### Can't Create Share

**Problem:** 404 File not found when creating share

**Solutions:**

1. Verify you own the file
2. Check fileId is correct
3. Ensure you're authenticated

---

## Future Enhancements

- [ ] Download counters
- [ ] Access logs
- [ ] Custom short URLs
- [ ] QR codes for share links
- [ ] Email sharing directly from API
- [ ] Bulk sharing (multiple files)
- [ ] Share templates
- [ ] Whitelist IPs
- [ ] Geographic restrictions

---

## Summary

File sharing is now easy and secure:

1. **Create** - `POST /api/shares` with fileId
2. **Share** - Send the shareUrl to anyone
3. **Access** - Recipients use `/share/{token}` (no auth!)
4. **Manage** - List, update, or delete shares anytime

**No authentication required for recipients!** 🎉
