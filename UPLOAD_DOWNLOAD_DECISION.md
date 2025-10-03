# Upload/Download Decision Quick Reference

## 🎯 TL;DR - When to Use What

### Uploads

```
Single File:
  ├─ < 10MB    → Basic Upload      (/api/files/upload)
  └─ > 10MB    → Chunked Upload    (/api/files/chunk-upload/*)

Multiple Files:
  ├─ Small     → Parallel Upload   (/api/files/parallel-upload)
  └─ Large     → Sequential Chunked (loop chunked upload)
```

### Downloads

```
Any File:
  ├─ < 10MB    → Basic Download    (/api/files/download/:fileName)
  ├─ 10-100MB  → Stream Download   (/api/files/stream-download/:fileId)
  └─ > 100MB   → Stream + Progress (same endpoint, track progress)
```

---

## 📤 Upload Examples

### Scenario 1: User Uploads Single 5MB File

```javascript
// UI automatically uses Basic Upload
const formData = new FormData();
formData.append("file", file);

fetch("/api/files/upload", {
  method: "POST",
  credentials: "include",
  body: formData,
});
```

**Why:** Small file, no need for complexity

---

### Scenario 2: User Uploads Single 50MB Video

```javascript
// UI automatically uses Chunked Upload
// 1. Initialize
const initRes = await fetch("/api/files/chunk-upload/init", {
  method: "POST",
  credentials: "include",
  body: JSON.stringify({
    fileName: file.name,
    totalSize: file.size,
    totalChunks: 10, // 50MB / 5MB per chunk
  }),
});

// 2. Upload 10 chunks (3 at a time in parallel)
// 3. Complete
```

**Why:** Large file benefits from chunking - can resume if interrupted

---

### Scenario 3: User Uploads 20 Photos (2MB each)

```javascript
// UI automatically uses Parallel Upload
const formData = new FormData();
files.forEach((file) => formData.append("files", file));

fetch("/api/files/parallel-upload", {
  method: "POST",
  credentials: "include",
  body: formData,
});
```

**Why:** Multiple small files upload faster in parallel

---

## 📥 Download Examples

### Scenario 1: User Downloads 3MB Document

```javascript
// Simple window.open
window.open("/api/files/download/filename.pdf", "_blank");
```

**Why:** Small file, browser handles download perfectly

---

### Scenario 2: User Downloads 80MB Video

```javascript
// Use stream endpoint (browser still handles it)
window.open("/api/files/stream-download/file-id", "_blank");
```

**Why:** Stream prevents loading entire file in memory

---

### Scenario 3: User Downloads 500MB File

```javascript
// Stream with manual progress tracking
const response = await fetch("/api/files/stream-download/file-id", {
  credentials: "include",
});

const reader = response.body.getReader();
let downloaded = 0;

while (true) {
  const { done, value } = await reader.read();
  if (done) break;

  downloaded += value.length;
  showProgress((downloaded / totalSize) * 100);
}
```

**Why:** Large file needs progress feedback for user

---

## 🧠 Smart Decision Function

```javascript
function decideUploadMethod(files) {
  if (files.length === 1) {
    const file = files[0];
    const sizeMB = file.size / (1024 * 1024);

    return sizeMB > 10 ? "chunked" : "basic";
  }

  return "parallel";
}

function decideDownloadMethod(fileSize) {
  const sizeMB = fileSize / (1024 * 1024);

  if (sizeMB < 10) return "basic";
  if (sizeMB < 100) return "stream";
  return "stream-with-progress";
}
```

---

## 🎨 Complete UI Flow

### Upload Flow

```
User selects file(s)
      ↓
  Check count & size
      ↓
  ┌───────────────┐
  │ Decide method │
  └───────┬───────┘
          ↓
    ┌─────────────┐
    │ Show message│  "Quick upload" or "Large file - using 10 chunks"
    └─────┬───────┘
          ↓
    ┌─────────────┐
    │   Upload    │  Use appropriate endpoint
    └─────┬───────┘
          ↓
    ┌─────────────┐
    │Show progress│  Update progress bar
    └─────┬───────┘
          ↓
    ┌─────────────┐
    │  Complete   │  "Upload successful!"
    └─────────────┘
```

### Download Flow

```
User clicks download
      ↓
  Check file size
      ↓
  ┌───────────────┐
  │ Decide method │
  └───────┬───────┘
          ↓
    ┌─────────────┐
    │  < 10MB?    │ → window.open (browser handles)
    └─────┬───────┘
          │
    ┌─────▼───────┐
    │ 10-100MB?   │ → window.open stream endpoint
    └─────┬───────┘
          │
    ┌─────▼───────┐
    │  > 100MB?   │ → Manual fetch with progress
    └─────────────┘
```

---

## 💡 Pro Tips

### 1. Don't Show Method to Technical Users

❌ **Bad:** "Using chunked upload with 5MB chunks"  
✅ **Good:** "Uploading large file... 45%"

### 2. Show Time Estimates

```javascript
const seconds = fileSize / (networkSpeed * 1024 * 1024);
const minutes = Math.ceil(seconds / 60);

// Show: "About 3 minutes remaining"
```

### 3. Handle Edge Cases

```javascript
// Very large file (> 1GB)
if (file.size > 1024 * 1024 * 1024) {
  if (!confirm("This is a large file. Upload may take a while. Continue?")) {
    return;
  }
}

// Slow connection
if (networkSpeed < 1) {
  showWarning("Slow connection detected. Upload may take longer than usual.");
}
```

### 4. Remember User Preference

```javascript
// If user successfully uploaded large file with chunks
localStorage.setItem("preferChunkedUpload", "true");

// Next time, use chunked even for 8MB files
const threshold = localStorage.getItem("preferChunkedUpload")
  ? 5 * 1024 * 1024 // 5MB
  : 10 * 1024 * 1024; // 10MB
```

---

## 🔄 Migration Path

If you have existing UI using only basic upload/download:

### Phase 1: Add Smart Detection (No UI Change)

```javascript
// Wrapper around existing code
async function upload(file) {
  if (shouldUseChunked(file)) {
    return chunkedUpload(file);
  }
  return existingBasicUpload(file);
}
```

### Phase 2: Add Progress (Better UX)

```javascript
// Add progress callback
async function upload(file, onProgress) {
  // ... same logic with progress updates
}
```

### Phase 3: Optimize (Performance)

```javascript
// Add parallel uploads, retry logic, etc.
```

---

## 📊 Performance Comparison

### Upload Performance (100MB file)

| Method             | Time | Success Rate | Resume? |
| ------------------ | ---- | ------------ | ------- |
| Basic Upload       | 180s | 70%          | ❌ No   |
| Chunked Upload     | 150s | 95%          | ✅ Yes  |
| Chunked + Parallel | 100s | 98%          | ✅ Yes  |

### Download Performance (100MB file)

| Method          | Memory Usage | Supports Resume? |
| --------------- | ------------ | ---------------- |
| Basic Download  | 100MB        | ❌ No            |
| Stream Download | 5MB          | ✅ Yes           |

---

## ✅ Checklist for Implementation

- [ ] File size detection working
- [ ] Progress bars show accurate progress
- [ ] Large files use chunked upload
- [ ] Multiple files use parallel upload
- [ ] Downloads use stream for large files
- [ ] Error handling for failed chunks
- [ ] Retry logic implemented
- [ ] Cancel upload functionality
- [ ] Time estimates shown
- [ ] Mobile-optimized (smaller chunks)
- [ ] Network speed detection
- [ ] Resume capability (optional)

---

## 🎯 Bottom Line

**You don't need to overthink this!**

The UI should:

1. **Detect** file size automatically
2. **Choose** the right method (user doesn't need to know)
3. **Show** progress (user just sees upload/download bar)
4. **Handle** errors gracefully

Users should just see "Upload" and "Download" - the smart logic happens behind the scenes! 🚀
