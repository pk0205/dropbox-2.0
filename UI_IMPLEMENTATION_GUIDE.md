# UI Implementation Guide

## Smart Upload/Download Decision Logic

Your UI should automatically choose the best method based on file size and capabilities.

---

## 📥 Download Decision Logic

### When to Use Each Method:

| File Size    | Method                     | Endpoint                                 | Why                                |
| ------------ | -------------------------- | ---------------------------------------- | ---------------------------------- |
| < 10MB       | Basic Download             | `GET /api/files/download/:fileName`      | Simple, fast, no special handling  |
| 10MB - 100MB | Stream Download            | `GET /api/files/stream-download/:fileId` | Memory efficient, supports resume  |
| > 100MB      | Stream Download + Progress | `GET /api/files/stream-download/:fileId` | Large files need progress tracking |

### Implementation:

```javascript
// Smart download function
async function smartDownload(file) {
  const { id, fileName, fileSize } = file;

  // Decision based on file size
  if (fileSize < 10 * 1024 * 1024) {
    // Small file: Basic download
    return basicDownload(fileName);
  } else if (fileSize < 100 * 1024 * 1024) {
    // Medium file: Stream download
    return streamDownload(id, fileName);
  } else {
    // Large file: Stream with progress
    return streamDownloadWithProgress(id, fileName, fileSize);
  }
}

// Basic download (< 10MB)
function basicDownload(fileName) {
  // Browser handles everything
  window.open(`/api/files/download/${fileName}`, "_blank");
}

// Stream download (10-100MB)
function streamDownload(fileId, fileName) {
  // Use stream endpoint
  window.open(`/api/files/stream-download/${fileId}`, "_blank");
}

// Stream with progress (> 100MB)
async function streamDownloadWithProgress(fileId, fileName, totalSize) {
  const response = await fetch(`/api/files/stream-download/${fileId}`, {
    credentials: "include",
  });

  if (!response.ok) throw new Error("Download failed");

  // Get reader for progress tracking
  const reader = response.body.getReader();
  const chunks = [];
  let downloaded = 0;

  while (true) {
    const { done, value } = await reader.read();
    if (done) break;

    chunks.push(value);
    downloaded += value.length;

    // Update progress
    const progress = (downloaded / totalSize) * 100;
    updateProgressBar(progress);
  }

  // Combine chunks and save
  const blob = new Blob(chunks);
  const url = URL.createObjectURL(blob);
  const a = document.createElement("a");
  a.href = url;
  a.download = fileName;
  a.click();
  URL.revokeObjectURL(url);
}
```

---

## 📤 Upload Decision Logic

### When to Use Each Method:

| Scenario           | Method          | Endpoint                          | Why                      |
| ------------------ | --------------- | --------------------------------- | ------------------------ |
| Single file < 10MB | Basic Upload    | `POST /api/files/upload`          | Simple, fast             |
| Multiple files     | Parallel Upload | `POST /api/files/parallel-upload` | Upload all at once       |
| Single file > 10MB | Chunked Upload  | `POST /api/files/chunk-upload/*`  | Reliable, resumable      |
| Slow connection    | Chunked Upload  | `POST /api/files/chunk-upload/*`  | Better for poor networks |

### Implementation:

```javascript
// Smart upload function
async function smartUpload(files, onProgress) {
  // Single file
  if (files.length === 1) {
    const file = files[0];
    const sizeMB = file.size / (1024 * 1024);

    if (sizeMB < 10) {
      // Small single file: Basic upload
      return basicUpload(file, onProgress);
    } else {
      // Large single file: Chunked upload
      return chunkedUpload(file, onProgress);
    }
  }

  // Multiple files
  else {
    const totalSize = files.reduce((sum, f) => sum + f.size, 0);
    const avgSize = totalSize / files.length;

    if (avgSize < 10 * 1024 * 1024) {
      // Small files: Parallel upload
      return parallelUpload(files, onProgress);
    } else {
      // Mix of large files: Upload sequentially with chunking
      return sequentialChunkedUpload(files, onProgress);
    }
  }
}

// Basic upload (< 10MB single file)
async function basicUpload(file, onProgress) {
  const formData = new FormData();
  formData.append("file", file);

  const xhr = new XMLHttpRequest();

  // Progress tracking
  xhr.upload.addEventListener("progress", (e) => {
    if (e.lengthComputable) {
      const progress = (e.loaded / e.total) * 100;
      onProgress(progress);
    }
  });

  return new Promise((resolve, reject) => {
    xhr.onload = () => {
      if (xhr.status === 201) {
        resolve(JSON.parse(xhr.responseText));
      } else {
        reject(new Error("Upload failed"));
      }
    };

    xhr.onerror = () => reject(new Error("Network error"));

    xhr.open("POST", "/api/files/upload");
    xhr.withCredentials = true; // Send cookies
    xhr.send(formData);
  });
}

// Parallel upload (multiple small files)
async function parallelUpload(files, onProgress) {
  const formData = new FormData();
  files.forEach((file) => {
    formData.append("files", file);
  });

  const xhr = new XMLHttpRequest();

  xhr.upload.addEventListener("progress", (e) => {
    if (e.lengthComputable) {
      const progress = (e.loaded / e.total) * 100;
      onProgress(progress);
    }
  });

  return new Promise((resolve, reject) => {
    xhr.onload = () => {
      if (xhr.status === 200) {
        resolve(JSON.parse(xhr.responseText));
      } else {
        reject(new Error("Upload failed"));
      }
    };

    xhr.onerror = () => reject(new Error("Network error"));

    xhr.open("POST", "/api/files/parallel-upload");
    xhr.withCredentials = true;
    xhr.send(formData);
  });
}

// Chunked upload (large files)
async function chunkedUpload(file, onProgress) {
  const chunkSize = 5 * 1024 * 1024; // 5MB
  const totalChunks = Math.ceil(file.size / chunkSize);

  // Step 1: Initialize
  const initRes = await fetch("/api/files/chunk-upload/init", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    credentials: "include",
    body: JSON.stringify({
      fileName: file.name,
      totalSize: file.size,
      totalChunks: totalChunks,
    }),
  });

  const { uploadId } = await initRes.json();

  // Step 2: Upload chunks with concurrency control
  const concurrency = 3; // 3 parallel chunks
  let uploadedChunks = 0;

  for (let i = 0; i < totalChunks; i += concurrency) {
    const promises = [];

    for (let j = 0; j < concurrency && i + j < totalChunks; j++) {
      const chunkNum = i + j;
      promises.push(uploadChunk(file, uploadId, chunkNum, chunkSize));
    }

    await Promise.all(promises);
    uploadedChunks += promises.length;

    // Update progress
    const progress = (uploadedChunks / totalChunks) * 100;
    onProgress(progress);
  }

  // Step 3: Complete
  const completeRes = await fetch(
    `/api/files/chunk-upload/${uploadId}/complete`,
    {
      method: "POST",
      credentials: "include",
    }
  );

  return completeRes.json();
}

async function uploadChunk(file, uploadId, chunkNum, chunkSize) {
  const start = chunkNum * chunkSize;
  const end = Math.min(start + chunkSize, file.size);
  const chunk = file.slice(start, end);

  const formData = new FormData();
  formData.append("chunkNumber", chunkNum);
  formData.append("chunk", chunk);

  const response = await fetch(`/api/files/chunk-upload/${uploadId}`, {
    method: "POST",
    credentials: "include",
    body: formData,
  });

  if (!response.ok) {
    throw new Error(`Chunk ${chunkNum} upload failed`);
  }
}
```

---

## 🎯 Complete React Component Example

```jsx
import { useState } from "react";

function FileUploadComponent() {
  const [files, setFiles] = useState([]);
  const [progress, setProgress] = useState(0);
  const [uploading, setUploading] = useState(false);
  const [method, setMethod] = useState("");

  const handleFileSelect = (e) => {
    setFiles(Array.from(e.target.files));
  };

  const handleUpload = async () => {
    if (files.length === 0) return;

    setUploading(true);
    setProgress(0);

    try {
      // Determine upload method
      const uploadMethod = determineUploadMethod(files);
      setMethod(uploadMethod);

      // Use smart upload
      await smartUpload(files, (p) => setProgress(p));

      alert("Upload complete!");
      setFiles([]);
      setProgress(0);
    } catch (error) {
      alert("Upload failed: " + error.message);
    } finally {
      setUploading(false);
    }
  };

  const determineUploadMethod = (files) => {
    if (files.length === 1) {
      const sizeMB = files[0].size / (1024 * 1024);
      return sizeMB < 10 ? "Basic Upload" : "Chunked Upload";
    }
    return "Parallel Upload";
  };

  return (
    <div className="upload-container">
      <input
        type="file"
        multiple
        onChange={handleFileSelect}
        disabled={uploading}
      />

      {files.length > 0 && (
        <div className="file-list">
          <h3>Selected Files:</h3>
          {files.map((file, i) => (
            <div key={i}>
              {file.name} - {formatBytes(file.size)}
            </div>
          ))}

          {method && (
            <p>
              <strong>Method:</strong> {method}
            </p>
          )}
        </div>
      )}

      <button onClick={handleUpload} disabled={uploading || files.length === 0}>
        {uploading ? "Uploading..." : "Upload Files"}
      </button>

      {uploading && (
        <div className="progress-bar">
          <div className="progress-fill" style={{ width: `${progress}%` }}>
            {Math.round(progress)}%
          </div>
        </div>
      )}
    </div>
  );
}

function formatBytes(bytes) {
  if (bytes === 0) return "0 Bytes";
  const k = 1024;
  const sizes = ["Bytes", "KB", "MB", "GB"];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return Math.round((bytes / Math.pow(k, i)) * 100) / 100 + " " + sizes[i];
}
```

---

## 📊 Decision Tree Visualization

```
Upload Decision:
┌─────────────────┐
│  Files Selected │
└────────┬────────┘
         │
    ┌────▼────┐
    │ Count?  │
    └────┬────┘
         │
    ┌────▼────────────────┐
    │  Single File        │  Multiple Files
    │                     │
    ▼                     ▼
┌────────┐          ┌─────────┐
│ Size?  │          │ Avg Size│
└───┬────┘          └────┬────┘
    │                    │
┌───▼───┐           ┌────▼────┐
│< 10MB │ > 10MB    │ < 10MB  │ > 10MB
│       │           │         │
▼       ▼           ▼         ▼
Basic   Chunked     Parallel  Sequential
Upload  Upload      Upload    Chunked


Download Decision:
┌──────────────┐
│  File Size   │
└──────┬───────┘
       │
  ┌────▼──────┬──────────┐
  │           │          │
  ▼           ▼          ▼
< 10MB     10-100MB    > 100MB
  │           │          │
  ▼           ▼          ▼
Basic      Stream     Stream +
Download   Download   Progress
```

---

## 🎨 User Experience Recommendations

### 1. **Show Upload Method to User**

```javascript
function getUploadMethodMessage(files) {
  if (files.length === 1) {
    const sizeMB = files[0].size / (1024 * 1024);
    if (sizeMB < 10) {
      return "Quick upload";
    } else {
      return `Large file - using optimized upload (${Math.ceil(
        sizeMB / 5
      )} chunks)`;
    }
  }
  return `Uploading ${files.length} files in parallel`;
}
```

### 2. **Estimate Upload Time**

```javascript
function estimateUploadTime(fileSize, connectionSpeed = 1) {
  // connectionSpeed in MB/s (default: 1 MB/s)
  const seconds = fileSize / (1024 * 1024 * connectionSpeed);

  if (seconds < 60) {
    return `~${Math.ceil(seconds)} seconds`;
  } else {
    return `~${Math.ceil(seconds / 60)} minutes`;
  }
}
```

### 3. **Show Chunk Progress for Large Files**

```jsx
function ChunkProgress({ currentChunk, totalChunks }) {
  return (
    <div className="chunk-progress">
      <p>
        Uploading chunk {currentChunk} of {totalChunks}
      </p>
      <div className="chunk-bar">
        {Array.from({ length: totalChunks }).map((_, i) => (
          <div
            key={i}
            className={`chunk ${i < currentChunk ? "complete" : "pending"}`}
          />
        ))}
      </div>
    </div>
  );
}
```

### 4. **Allow Resume on Failed Uploads**

```javascript
async function resumableChunkedUpload(file, onProgress) {
  const uploadId = localStorage.getItem(`upload_${file.name}`);

  if (uploadId) {
    // Resume existing upload
    const resumePoint = await getUploadProgress(uploadId);
    return continueChunkedUpload(file, uploadId, resumePoint, onProgress);
  } else {
    // Start new upload
    return chunkedUpload(file, onProgress);
  }
}
```

---

## 🚀 Performance Optimization Tips

### 1. **Detect Connection Speed**

```javascript
async function detectConnectionSpeed() {
  const startTime = Date.now();

  // Download a small test file
  const response = await fetch("/api/test-speed");
  await response.blob();

  const endTime = Date.now();
  const duration = (endTime - startTime) / 1000; // seconds
  const fileSize = 1; // MB

  return fileSize / duration; // MB/s
}

// Adjust chunk size based on speed
function getOptimalChunkSize(connectionSpeed) {
  if (connectionSpeed > 10) return 20 * 1024 * 1024; // 20MB
  if (connectionSpeed > 5) return 10 * 1024 * 1024; // 10MB
  return 5 * 1024 * 1024; // 5MB
}
```

### 2. **Retry Failed Chunks**

```javascript
async function uploadChunkWithRetry(
  file,
  uploadId,
  chunkNum,
  chunkSize,
  maxRetries = 3
) {
  for (let attempt = 0; attempt < maxRetries; attempt++) {
    try {
      await uploadChunk(file, uploadId, chunkNum, chunkSize);
      return; // Success
    } catch (error) {
      if (attempt === maxRetries - 1) throw error;

      // Wait before retry (exponential backoff)
      await new Promise((resolve) =>
        setTimeout(resolve, Math.pow(2, attempt) * 1000)
      );
    }
  }
}
```

### 3. **Cancel Uploads**

```javascript
class UploadManager {
  constructor() {
    this.abortController = new AbortController();
  }

  async upload(file, onProgress) {
    try {
      await smartUpload(file, onProgress, {
        signal: this.abortController.signal,
      });
    } catch (error) {
      if (error.name === "AbortError") {
        console.log("Upload cancelled");
      } else {
        throw error;
      }
    }
  }

  cancel() {
    this.abortController.abort();
  }
}
```

---

## 📱 Mobile Considerations

### Adjust for Mobile Networks:

```javascript
function isMobile() {
  return /Android|webOS|iPhone|iPad|iPod/i.test(navigator.userAgent);
}

function getUploadStrategy() {
  if (isMobile()) {
    // More aggressive chunking on mobile
    return {
      chunkSize: 2 * 1024 * 1024, // 2MB chunks
      concurrency: 2, // Fewer parallel requests
      threshold: 5 * 1024 * 1024, // Start chunking at 5MB
    };
  }

  return {
    chunkSize: 5 * 1024 * 1024,
    concurrency: 3,
    threshold: 10 * 1024 * 1024,
  };
}
```

---

## 🎯 Summary

### Upload Strategy:

- **< 10MB single file** → Basic Upload (fast, simple)
- **> 10MB single file** → Chunked Upload (reliable, resumable)
- **Multiple small files** → Parallel Upload (efficient)
- **Multiple large files** → Sequential Chunked (controlled)

### Download Strategy:

- **< 10MB** → Basic Download (browser handles it)
- **10-100MB** → Stream Download (memory efficient)
- **> 100MB** → Stream + Progress (user feedback)

### Key Principles:

1. ✅ **Automatic** - User doesn't choose, UI decides
2. ✅ **Progressive** - Show progress for long operations
3. ✅ **Resilient** - Retry failed chunks, allow resume
4. ✅ **Responsive** - Adapt to connection speed
5. ✅ **Transparent** - Show what method is being used

Your UI should **"just work"** for users while using the optimal method behind the scenes! 🎉
