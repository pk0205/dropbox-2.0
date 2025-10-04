const API_URL = "http://localhost:4000";

export interface User {
  id: string;
  username: string;
  email: string;
  firstName: string;
  lastName: string;
}

interface LoginRequest {
  emailOrUsername: string;
  password: string;
}

interface SignupRequest {
  firstName: string;
  lastName: string;
  username: string;
  email: string;
  password: string;
}

// Helper function for API calls
async function fetchAPI<T>(
  endpoint: string,
  options?: RequestInit
): Promise<T> {
  const response = await fetch(`${API_URL}${endpoint}`, {
    credentials: "include",
    headers: {
      "Content-Type": "application/json",
      ...options?.headers,
    },
    ...options,
  });

  if (!response.ok) {
    const error = await response
      .json()
      .catch(() => ({ error: "Request failed" }));
    throw new Error(
      error.error || `HTTP ${response.status}: ${response.statusText}`
    );
  }

  return response.json();
}

export interface FileItem {
  id: string;
  userId: string;
  fileName: string;
  originalName: string;
  filePath: string;
  fileSize: number;
  mimeType?: string;
  checksum: string;
  parentId?: string | null;
  isFolder: boolean;
  isShared: boolean;
  createdAt: string;
  updatedAt: string;
}

interface UploadResponse {
  message: string;
  file?: string;
  fileId?: string;
  fileName?: string;
}

interface CreateFolderRequest {
  folderName: string;
  parentId?: string;
}

// Auth API
export const authAPI = {
  me: () => fetchAPI<User>("/api/me"),

  login: (data: LoginRequest) =>
    fetchAPI<User>("/api/user/login", {
      method: "POST",
      body: JSON.stringify(data),
    }),

  signup: (data: SignupRequest) =>
    fetchAPI<User>("/api/user/signup", {
      method: "POST",
      body: JSON.stringify(data),
    }),

  logout: () =>
    fetchAPI<{ message: string }>("/api/user/logout", {
      method: "POST",
    }),
};

// File API
export const fileAPI = {
  list: (parentId?: string) => {
    const url = parentId ? `/api/files?parentId=${parentId}` : "/api/files";
    return fetchAPI<FileItem[]>(url);
  },

  upload: async (file: File, parentId?: string) => {
    const formData = new FormData();
    formData.append("file", file);
    if (parentId) formData.append("parentId", parentId);

    const response = await fetch(`${API_URL}/api/files/upload`, {
      method: "POST",
      credentials: "include",
      body: formData, // Don't set Content-Type, browser will set it with boundary
    });

    if (!response.ok) {
      const error = await response
        .json()
        .catch(() => ({ error: "Upload failed" }));
      throw new Error(error.error || `HTTP ${response.status}`);
    }

    return response.json() as Promise<UploadResponse>;
  },

  delete: (fileId: string) =>
    fetchAPI<{ message: string }>(`/api/files/${fileId}`, {
      method: "DELETE",
    }),

  download: (fileId: string) => {
    // Return download URL
    return `${API_URL}/api/files/stream-download/${fileId}`;
  },

  createFolder: (data: CreateFolderRequest) =>
    fetchAPI<{ message: string; folderId: string }>("/api/folders", {
      method: "POST",
      body: JSON.stringify(data),
    }),
};
