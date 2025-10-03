# Dropbox 2.0 API Documentation

## Overview

High-performance file storage system with advanced features like chunked uploads, parallel processing, deduplication, and resumable downloads.

## Performance Features

### 1. **Chunked Upload (For Large Files)**

- Files are split into 5MB chunks
- Each chunk uploaded independently
- Resume capability if connection drops
- Perfect for files > 10MB

**Flow:**

```
1. Initialize upload session → Get uploadId
2. Upload chunks in parallel → Track progress
3. Complete upload → Combine chunks and save
```

### 2. **Parallel Upload (Multiple Files)**

- Upload up to 10 files simultaneously
- Worker pool pattern for efficient resource usage
- Automatic deduplication

### 3. **Stream Download**

- Memory-efficient streaming
- Support for HTTP Range requests (resumable downloads)
- No need to load entire file in memory

### 4. **Deduplication**

- SHA-256 checksums for file identification
- Same file stored only once physically
- Multiple references in database

### 5. **Folder Structure**

- Hierarchical organization with parent-child relationships
- Efficient querying with database indexes

---

## API Endpoints

### Authentication

All authenticated endpoints use **HTTP-only cookies** for security. No Bearer tokens required!

#### Signup

```http
POST /api/user/signup
Content-Type: application/json

{
  "firstName": "John",
  "lastName": "Doe",
  "username": "johndoe",
  "email": "john@example.com",
  "password": "securepassword"
}
```

**Response:** JWT token automatically set in `AuthToken` cookie

#### Login

```http
POST /api/user/login
Content-Type: application/json

{
  "emailOrUsername": "johndoe",
  "password": "securepassword"
}
```

**Response:** JWT token automatically set in `AuthToken` cookie. The cookie is:

- HTTP-only (cannot be accessed by JavaScript)
- Secure (only sent over HTTPS in production)
- SameSite: Lax
- Expires in 30 days

---

### File Management (All require authentication via cookies)

#### 1. List Files

```http
GET /api/files?parentId={folderId}
Cookie: AuthToken=<your-token>
```

**Response:**

```json
[
  {
    "id": "file-uuid",
    "fileName": "document.pdf",
    "originalName": "My Document.pdf",
    "fileSize": 1024000,
    "isFolder": false,
    "createdAt": "2025-10-03T10:00:00Z",
    "updatedAt": "2025-10-03T10:00:00Z"
  }
]
```

#### 2. Basic Upload (Small Files < 10MB)

```http
POST /api/files/upload
Cookie: AuthToken=<your-token>
Content-Type: multipart/form-data

file: <binary data>
```

**Note:** The cookie is sent automatically by the browser.

#### 3. Parallel Upload (Multiple Files)

```http
POST /api/files/parallel-upload
Cookie: AuthToken=<your-token>
Content-Type: multipart/form-data

files: <file1>
files: <file2>
files: <file3>
```

**Response:**

```json
{
  "message": "Upload completed",
  "results": [
    {
      "fileId": "uuid-1",
      "fileName": "file1.pdf"
    },
    {
      "fileId": "uuid-2",
      "fileName": "file2.jpg"
    }
  ]
}
```

---

### Chunked Upload (Large Files)

#### Step 1: Initialize Upload

```http
POST /api/files/chunk-upload/init
Cookie: AuthToken=<your-token>
Content-Type: application/json

{
  "fileName": "large-video.mp4",
  "totalSize": 104857600,
  "totalChunks": 20,
  "parentId": "folder-uuid" // optional
}
```

**Response:**

```json
{
  "uploadId": "upload-session-uuid",
  "chunkSize": 5242880,
  "totalChunks": 20,
  "expiresAt": "2025-10-04T10:00:00Z"
}
```

#### Step 2: Upload Chunks (Parallel)

```http
POST /api/files/chunk-upload/{uploadId}
Cookie: AuthToken=<your-token>
Content-Type: multipart/form-data

chunkNumber: 0
chunk: <binary data>
```

Upload each chunk from 0 to totalChunks-1. You can upload multiple chunks in parallel!

#### Step 3: Complete Upload

```http
POST /api/files/chunk-upload/{uploadId}/complete
Cookie: AuthToken=<your-token>
```

**Response:**

```json
{
  "message": "File uploaded successfully",
  "fileId": "file-uuid",
  "fileName": "large-video.mp4",
  "fileSize": 104857600,
  "checksum": "sha256-hash"
}
```

---

### Download Files

#### Basic Download

```http
GET /api/files/download/{fileName}
Cookie: AuthToken=<your-token>
```

#### Stream Download (Recommended for Large Files)

```http
GET /api/files/stream-download/{fileId}
Cookie: AuthToken=<your-token>
```

**Resumable Download (Range Request):**

```http
GET /api/files/stream-download/{fileId}
Cookie: AuthToken=<your-token>
Range: bytes=0-1048575
```

The server responds with status `206 Partial Content` and sends the requested byte range.

---

### Folder Management

#### Create Folder

```http
POST /api/folders
Cookie: AuthToken=<your-token>
Content-Type: application/json

{
  "folderName": "My Documents",
  "parentId": "parent-folder-uuid" // optional, null for root
}
```

#### List Folder Contents

```http
GET /api/files?parentId={folderId}
Cookie: AuthToken=<your-token>
```

---

### Delete File/Folder

```http
DELETE /api/files/{fileId}
Cookie: AuthToken=<your-token>
```

**Note:** Due to deduplication, physical file is only deleted if no other references exist.

---

## File Sharing

### Create Share Link

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

### Access Shared File (Public)

**No authentication required!** Anyone with the link can access.

```http
GET /share/{token}?password=secret123
```

**Response:** File download or folder contents

### Get Share Info

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

### List Your Shares

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

### Update Share Link

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

### Delete Share Link

Revoke a share link.

```http
DELETE /api/shares/{shareId}
Cookie: AuthToken=<your-token>
```

**For detailed sharing examples, see [SHARING_GUIDE.md](./SHARING_GUIDE.md)**

---

## Client-Side Implementation Examples

### JavaScript: Chunked Upload

```javascript
async function uploadLargeFile(file) {
  const chunkSize = 5 * 1024 * 1024; // 5MB
  const totalChunks = Math.ceil(file.size / chunkSize);

  // Step 1: Initialize
  const initRes = await fetch("/api/files/chunk-upload/init", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    credentials: "include", // Important: sends cookies
    body: JSON.stringify({
      fileName: file.name,
      totalSize: file.size,
      totalChunks: totalChunks,
    }),
  });

  const { uploadId } = await initRes.json();

  // Step 2: Upload chunks in parallel (max 3 concurrent)
  const uploadChunk = async (chunkNum) => {
    const start = chunkNum * chunkSize;
    const end = Math.min(start + chunkSize, file.size);
    const chunk = file.slice(start, end);

    const formData = new FormData();
    formData.append("chunkNumber", chunkNum);
    formData.append("chunk", chunk);

    await fetch(`/api/files/chunk-upload/${uploadId}`, {
      method: "POST",
      credentials: "include", // Important: sends cookies
      body: formData,
    });
  };

  // Upload with concurrency control
  const concurrency = 3;
  for (let i = 0; i < totalChunks; i += concurrency) {
    const promises = [];
    for (let j = 0; j < concurrency && i + j < totalChunks; j++) {
      promises.push(uploadChunk(i + j));
    }
    await Promise.all(promises);
    console.log(`Progress: ${Math.round((i / totalChunks) * 100)}%`);
  }

  // Step 3: Complete
  const completeRes = await fetch(
    `/api/files/chunk-upload/${uploadId}/complete`,
    {
      method: "POST",
      credentials: "include", // Important: sends cookies
    }
  );

  const result = await completeRes.json();
  console.log("Upload complete!", result);
}
```

**Note:** Always include `credentials: "include"` in fetch requests to send cookies!

### JavaScript: Resumable Download

```javascript
async function downloadFileResumable(fileId, onProgress) {
  const chunkSize = 1024 * 1024; // 1MB chunks
  let downloaded = 0;
  const chunks = [];

  // Get file size first
  const headRes = await fetch(`/api/files/stream-download/${fileId}`, {
    method: "HEAD",
    credentials: "include", // Important: sends cookies
  });
  const totalSize = parseInt(headRes.headers.get("Content-Length"));

  while (downloaded < totalSize) {
    const end = Math.min(downloaded + chunkSize - 1, totalSize - 1);

    const res = await fetch(`/api/files/stream-download/${fileId}`, {
      credentials: "include", // Important: sends cookies
      headers: {
        Range: `bytes=${downloaded}-${end}`,
      },
    });

    const chunk = await res.arrayBuffer();
    chunks.push(chunk);
    downloaded += chunk.byteLength;

    onProgress(downloaded / totalSize);
  }

  // Combine chunks into Blob
  const blob = new Blob(chunks);
  return blob;
}
```

### Python: Parallel Upload Multiple Files

```python
import requests
import concurrent.futures

# Create a session to maintain cookies
session = requests.Session()

def login(username, password):
    """Login and get cookie"""
    response = session.post(
        'http://localhost:3000/api/user/login',
        json={'emailOrUsername': username, 'password': password}
    )
    return response.json()

def upload_file(file_path):
    """Upload file using session cookies"""
    with open(file_path, 'rb') as f:
        files = {'files': f}
        response = session.post(
            'http://localhost:3000/api/files/parallel-upload',
            files=files
        )
        return response.json()

# Login first to get cookie
login('your_username', 'your_password')

# Upload multiple files in parallel
files_to_upload = ['file1.pdf', 'file2.jpg', 'file3.docx']

with concurrent.futures.ThreadPoolExecutor(max_workers=3) as executor:
    futures = [executor.submit(upload_file, f) for f in files_to_upload]
    results = [f.result() for f in concurrent.futures.as_completed(futures)]

print(results)
```

---

## Performance Tuning

### Server-Side

1. **Adjust chunk size** based on network conditions:

   ```go
   const ChunkSize = 10 * 1024 * 1024 // 10MB for faster networks
   ```

2. **Increase worker pool** for more parallelism:

   ```go
   const MaxWorkers = 20
   ```

3. **Database connection pooling** - Use pgxpool for better performance:

   ```go
   pool, err := pgxpool.New(context.Background(), connString)
   ```

4. **File compression** - Add gzip compression for text files

### Client-Side

1. **Concurrent chunk uploads** - 3-5 parallel uploads optimal
2. **Network retry logic** - Retry failed chunks automatically
3. **Upload queue** - Queue files and upload sequentially to avoid overwhelming server

---

## Database Schema

### Files Table

- `id` - Unique file identifier
- `user_id` - Owner reference
- `file_name` - Stored filename
- `original_name` - User's original filename
- `file_path` - Physical storage path
- `file_size` - Size in bytes
- `checksum` - SHA-256 hash for deduplication
- `parent_id` - Parent folder (NULL for root)
- `is_folder` - Boolean flag
- `created_at`, `updated_at` - Timestamps

**Indexes:** user_id, parent_id, checksum

### Chunk Uploads Table

- Temporary storage for upload sessions
- Expires after 24 hours
- Tracks uploaded chunks as array

---

## Security Considerations

1. **Authentication** - JWT tokens with 30-day expiry
2. **File Access Control** - Users can only access their own files
3. **Path Traversal Protection** - UUIDs prevent directory traversal
4. **File Size Limits** - 100MB default body limit
5. **Upload Session Expiry** - 24-hour timeout for incomplete uploads

---

## Future Enhancements

1. **File Sharing** - Share links with expiration and passwords
2. **File Versioning** - Keep history of file changes
3. **Search** - Full-text search across file names and contents
4. **Compression** - Automatic compression for text files
5. **CDN Integration** - Serve static files from CDN
6. **Thumbnail Generation** - For images and videos
7. **Trash/Recycle Bin** - Soft delete with recovery period
8. **Real-time Sync** - WebSocket for live file updates

---

## Troubleshooting

### Upload Fails

- Check body size limit in Fiber config
- Verify storage directory permissions
- Check available disk space

### Slow Performance

- Enable database connection pooling
- Increase MaxWorkers for more parallelism
- Use chunked upload for large files
- Check network bandwidth

### Authentication Issues

- Verify JWT secret is set in .env
- Check token expiration
- Ensure middleware is applied to routes

---

## License

MIT
