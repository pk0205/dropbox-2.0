# TanStack Query Migration ✅ COMPLETE

All API calls in the frontend now use TanStack Query for better performance, caching, and state management.

## ✅ What Was Done

### 1. Created API Service Layer (`app/lib/api.ts`)

- Centralized all API endpoints
- Type-safe API calls with TypeScript interfaces
- Consistent error handling
- Automatic credentials inclusion for authentication

### 2. Created Custom Hooks

**Auth Hooks (`app/hooks/useAuth.ts`)**

- `useMe()` - Query hook to fetch current user (auto-runs on mount)
- `useLogin()` - Mutation hook for login
- `useSignup()` - Mutation hook for signup
- `useLogout()` - Mutation hook for logout

**File Hooks (`app/hooks/useFiles.ts`)**

- `useFiles(parentId?)` - Query hook to fetch files/folders
- `useUploadFile()` - Mutation hook for uploading files
- `useDeleteFile()` - Mutation hook for deleting files
- `useCreateFolder()` - Mutation hook for creating folders

### 3. Updated AuthContext (`app/contexts/AuthContext.tsx`)

- Now uses TanStack Query hooks internally
- Maintains same API for components (no breaking changes)
- Automatic cache management
- Performance optimized with `useMemo`

### 4. Updated Dashboard (`app/routes/dashboard.tsx`)

- Replaced all mock data with real API calls
- Integrated file upload, download, delete, and folder creation
- Added loading and error states
- Connected to backend file management endpoints

### 5. Configured QueryClient (`app/root.tsx`)

- Set default stale time (1 minute)
- Disabled refetch on window focus
- Limited retries to 1

### 6. Fixed Backend Middleware (`middleware/requireAuth.go`)

- Updated to set both `userName` and `userID` in context
- Added database query to fetch user ID from username

## 🎯 Benefits

### Automatic Features You Get For Free:

- ✅ **Automatic caching** - Reduces unnecessary API calls
- ✅ **Loading states** - Built-in `isLoading` states
- ✅ **Error handling** - Automatic error capture
- ✅ **Request deduplication** - Multiple components requesting same data = 1 API call
- ✅ **Background refetching** - Keeps data fresh automatically
- ✅ **Optimistic updates** - Can update UI before API responds
- ✅ **Query invalidation** - Automatic refresh when data changes

## 📖 How to Use

### Authentication (Already Implemented)

The `useAuth` hook works exactly the same as before:

```tsx
import { useAuth } from "~/contexts/AuthContext";

function MyComponent() {
  const { user, isAuthenticated, isLoading, login, signup, logout } = useAuth();

  // Use as normal - now powered by TanStack Query!
}
```

### Adding New API Calls

When you need to add file operations or other API calls:

#### Step 1: Add API function to `app/lib/api.ts`

```typescript
export const fileAPI = {
  list: (folderId?: string) =>
    fetchAPI<FileItem[]>(`/api/files?folderId=${folderId || ""}`),

  upload: (file: File, folderId?: string) => {
    const formData = new FormData();
    formData.append("file", file);
    if (folderId) formData.append("folderId", folderId);

    return fetchAPI<FileItem>("/api/files/upload", {
      method: "POST",
      body: formData,
    });
  },

  delete: (fileId: string) =>
    fetchAPI<void>(`/api/files/${fileId}`, {
      method: "DELETE",
    }),
};
```

#### Step 2: Create custom hook in `app/hooks/`

```typescript
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { fileAPI } from "../lib/api";

export const fileKeys = {
  all: ["files"] as const,
  list: (folderId?: string) => [...fileKeys.all, "list", folderId] as const,
};

export function useFiles(folderId?: string) {
  return useQuery({
    queryKey: fileKeys.list(folderId),
    queryFn: () => fileAPI.list(folderId),
  });
}

export function useUploadFile() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: fileAPI.upload,
    onSuccess: () => {
      // Refresh the file list after upload
      queryClient.invalidateQueries({ queryKey: fileKeys.all });
    },
  });
}

export function useDeleteFile() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: fileAPI.delete,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: fileKeys.all });
    },
  });
}
```

#### Step 3: Use in your component

```tsx
import { useFiles, useUploadFile, useDeleteFile } from "~/hooks/useFiles";

function Dashboard() {
  const { data: files, isLoading, error } = useFiles();
  const uploadMutation = useUploadFile();
  const deleteMutation = useDeleteFile();

  const handleUpload = async (file: File) => {
    try {
      await uploadMutation.mutateAsync(file);
      alert("Upload successful!");
    } catch (error) {
      alert("Upload failed: " + error.message);
    }
  };

  const handleDelete = async (fileId: string) => {
    try {
      await deleteMutation.mutateAsync(fileId);
      alert("Deleted successfully!");
    } catch (error) {
      alert("Delete failed: " + error.message);
    }
  };

  if (isLoading) return <div>Loading...</div>;
  if (error) return <div>Error: {error.message}</div>;

  return (
    <div>
      {files?.map((file) => (
        <div key={file.id}>
          {file.name}
          <button onClick={() => handleDelete(file.id)}>Delete</button>
        </div>
      ))}
    </div>
  );
}
```

## 📚 Resources

- [TanStack Query Docs](https://tanstack.com/query/latest)
- [Query Keys Guide](https://tanstack.com/query/latest/docs/react/guides/query-keys)
- [Mutations Guide](https://tanstack.com/query/latest/docs/react/guides/mutations)

## 💡 Tips

1. **Use Query Keys Wisely** - Organize them hierarchically for easy invalidation
2. **Leverage Optimistic Updates** - Update UI immediately for better UX
3. **Use `mutateAsync` vs `mutate`** - `mutateAsync` returns a promise, better for async/await
4. **Invalidate Queries** - After mutations, invalidate related queries to refresh data
5. **Check `useFiles.example.ts`** - See complete examples for file operations

## 🎉 All Features Implemented

✅ Authentication (login, signup, logout, me)
✅ File listing with folder support
✅ File upload
✅ File download
✅ File deletion
✅ Folder creation
✅ Loading and error states
✅ Automatic cache management

## 🔌 Backend Integration

The frontend is now fully connected to these backend endpoints:

- `GET /api/me` - Get current user
- `GET /api/files?parentId=` - List files/folders
- `POST /api/files/upload` - Upload file
- `GET /api/files/stream-download/:fileId` - Download file
- `DELETE /api/files/:fileId` - Delete file
- `POST /api/folders` - Create folder
