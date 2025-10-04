import { useState, useEffect } from "react";
import { useNavigate, useParams } from "react-router";
import { Button } from "~/components/ui/button";
import { ProtectedRoute } from "~/components/ProtectedRoute";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "~/components/ui/card";
import {
  Cloud,
  Upload,
  Download,
  Trash2,
  Share2,
  Folder,
  File,
  X,
  LogOut,
  User,
  FolderPlus,
  Search,
  Grid3x3,
  List,
  FileText,
  Image as ImageIcon,
  Video,
  Music,
  Archive,
  Home,
  ChevronRight,
  Loader2,
} from "lucide-react";
import { useAuth } from "~/contexts/AuthContext";
import {
  useFiles,
  useUploadFile,
  useDeleteFile,
  useCreateFolder,
} from "~/hooks/useFiles";
import { fileAPI, type FileItem as APIFileItem } from "~/lib/api";
export function meta() {
  return [
    { title: "Dashboard - Dropbox 2.0" },
    { name: "description", content: "Manage your files and folders" },
  ];
}

// Adapter type for display (maps API types to UI types)
interface FileItem {
  id: string;
  name: string;
  type: "file" | "folder";
  size?: number;
  mimeType?: string;
  createdAt: string;
  isShared?: boolean;
  parentId?: string;
}

// Convert API file to display file
function mapAPIFileToDisplayFile(apiFile: APIFileItem): FileItem {
  return {
    id: apiFile.id,
    name: apiFile.originalName,
    type: apiFile.isFolder ? "folder" : "file",
    size: apiFile.fileSize,
    mimeType: apiFile.mimeType,
    createdAt: new Date(apiFile.createdAt).toISOString().split("T")[0],
    isShared: apiFile.isShared,
    parentId: apiFile.parentId ?? undefined,
  };
}

function DashboardContent() {
  const navigate = useNavigate();
  const params = useParams();
  const { user, logout } = useAuth();

  const [showUploadModal, setShowUploadModal] = useState(false);
  const [isDragging, setIsDragging] = useState(false);
  const [viewMode, setViewMode] = useState<"grid" | "list">("grid");
  const [searchQuery, setSearchQuery] = useState("");
  const [selectedFiles, setSelectedFiles] = useState<string[]>([]);
  const [currentFolderId, setCurrentFolderId] = useState<string | undefined>(
    undefined
  );

  // Update currentFolderId when URL changes
  useEffect(() => {
    const folderId = params.folderId || undefined;
    setCurrentFolderId(folderId);
  }, [params.folderId]);

  // Fetch files using TanStack Query
  const { data: apiFiles, isLoading, error } = useFiles(currentFolderId);
  const uploadMutation = useUploadFile();
  const deleteMutation = useDeleteFile();
  const createFolderMutation = useCreateFolder();

  // Convert API files to display files
  const files: FileItem[] = apiFiles
    ? apiFiles.map(mapAPIFileToDisplayFile)
    : [];

  // Format file size
  const formatBytes = (bytes: number) => {
    if (bytes === 0) return "0 Bytes";
    const k = 1024;
    const sizes = ["Bytes", "KB", "MB", "GB"];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return Math.round((bytes / Math.pow(k, i)) * 100) / 100 + " " + sizes[i];
  };

  // Get file icon based on mime type
  const getFileIcon = (mimeType?: string) => {
    if (!mimeType) return <File className="w-6 h-6" />;

    if (mimeType.startsWith("image/"))
      return <ImageIcon className="w-6 h-6 text-blue-500" />;
    if (mimeType.startsWith("video/"))
      return <Video className="w-6 h-6 text-purple-500" />;
    if (mimeType.startsWith("audio/"))
      return <Music className="w-6 h-6 text-pink-500" />;
    if (mimeType.includes("pdf"))
      return <FileText className="w-6 h-6 text-red-500" />;
    if (mimeType.includes("zip") || mimeType.includes("compressed"))
      return <Archive className="w-6 h-6 text-orange-500" />;

    return <File className="w-6 h-6 text-gray-500" />;
  };

  // Handle file upload (real)
  const handleFileUpload = async (uploadedFiles: FileList | null) => {
    if (!uploadedFiles) return;

    try {
      // Upload each file
      for (const file of Array.from(uploadedFiles)) {
        await uploadMutation.mutateAsync({
          file,
          parentId: currentFolderId,
        });
      }
      setShowUploadModal(false);
    } catch (error) {
      console.error("Upload failed:", error);
      alert(
        `Upload failed: ${error instanceof Error ? error.message : "Unknown error"}`
      );
    }
  };

  // Handle drag and drop
  const handleDragOver = (e: React.DragEvent) => {
    e.preventDefault();
    setIsDragging(true);
  };

  const handleDragLeave = () => {
    setIsDragging(false);
  };

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault();
    setIsDragging(false);
    handleFileUpload(e.dataTransfer.files);
  };

  // Handle delete (real)
  const handleDelete = async (id: string) => {
    if (confirm("Are you sure you want to delete this item?")) {
      try {
        await deleteMutation.mutateAsync(id);
        setSelectedFiles(selectedFiles.filter((sid) => sid !== id));
      } catch (error) {
        console.error("Delete failed:", error);
        alert(
          `Delete failed: ${error instanceof Error ? error.message : "Unknown error"}`
        );
      }
    }
  };

  // Handle share (mock)
  const handleShare = (id: string) => {
    alert(`Share link created for: ${files.find((f) => f.id === id)?.name}`);
  };

  // Handle download (real)
  const handleDownload = (id: string) => {
    const file = files.find((f) => f.id === id);
    if (!file) return;

    // Open download URL in new tab
    const downloadUrl = fileAPI.download(id);
    window.open(downloadUrl, "_blank");
  };

  // Create folder (real)
  const handleCreateFolder = async () => {
    const folderName = prompt("Enter folder name:");
    if (folderName) {
      try {
        await createFolderMutation.mutateAsync({
          folderName,
          parentId: currentFolderId,
        });
      } catch (error) {
        console.error("Create folder failed:", error);
        alert(
          `Create folder failed: ${error instanceof Error ? error.message : "Unknown error"}`
        );
      }
    }
  };

  // Handle folder navigation
  const handleFolderClick = (folderId: string) => {
    navigate(`/dashboard/folder/${folderId}`);
    setSearchQuery(""); // Clear search when navigating
  };

  const handleBackToRoot = () => {
    navigate("/dashboard");
  };

  // Get current folder details
  const currentFolder = currentFolderId
    ? files.find((f) => f.id === currentFolderId && f.type === "folder")
    : null;

  // Build breadcrumb trail
  const getBreadcrumbs = () => {
    const breadcrumbs: Array<{ id: string | null; name: string }> = [
      { id: null, name: "Home" },
    ];
    if (currentFolder) {
      breadcrumbs.push({ id: currentFolder.id, name: currentFolder.name });
    }
    return breadcrumbs;
  };

  // Filter files based on search (folder filtering is handled by API)
  const filteredFiles = files.filter((file) => {
    // Filter by search query
    const matchesSearch = file.name
      .toLowerCase()
      .includes(searchQuery.toLowerCase());

    return matchesSearch;
  });

  // Calculate storage usage
  const totalStorage = files.reduce((acc, file) => acc + (file.size || 0), 0);
  const storageLimit = 1024 * 1024 * 1024; // 1GB mock limit
  const storagePercentage = (totalStorage / storageLimit) * 100;

  // Show loading state
  if (isLoading) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-center">
          <Loader2 className="w-12 h-12 animate-spin mx-auto text-blue-500 mb-4" />
          <p className="text-gray-600">Loading files...</p>
        </div>
      </div>
    );
  }

  // Show error state
  if (error) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-center">
          <p className="text-red-600 mb-4">
            Error loading files: {error.message}
          </p>
          <Button onClick={() => window.location.reload()}>Try Again</Button>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <header className="bg-white border-b sticky top-0 z-40">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center h-16">
            <div className="flex items-center space-x-3">
              <div className="w-10 h-10 bg-gradient-to-br from-blue-500 to-purple-600 rounded-lg flex items-center justify-center">
                <Cloud className="w-6 h-6 text-white" />
              </div>
              <span className="text-xl font-bold bg-gradient-to-r from-blue-600 to-purple-600 bg-clip-text text-transparent">
                Dropbox 2.0
              </span>
            </div>

            <div className="flex items-center space-x-4">
              <Button variant="ghost" size="sm">
                <User className="w-4 h-4 mr-2" />
                {user?.firstName || "Profile"}
              </Button>
              <Button
                variant="ghost"
                size="sm"
                onClick={async () => {
                  await logout();
                  navigate("/");
                }}
              >
                <LogOut className="w-4 h-4 mr-2" />
                Logout
              </Button>
            </div>
          </div>
        </div>
      </header>

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Storage Stats */}
        <div className="grid md:grid-cols-3 gap-6 mb-8">
          <Card>
            <CardHeader className="pb-3">
              <CardDescription>Total Files</CardDescription>
              <CardTitle className="text-3xl">
                {files.filter((f) => f.type === "file").length}
              </CardTitle>
            </CardHeader>
          </Card>

          <Card>
            <CardHeader className="pb-3">
              <CardDescription>Total Folders</CardDescription>
              <CardTitle className="text-3xl">
                {files.filter((f) => f.type === "folder").length}
              </CardTitle>
            </CardHeader>
          </Card>

          <Card>
            <CardHeader className="pb-3">
              <CardDescription>Storage Used</CardDescription>
              <CardTitle className="text-3xl">
                {formatBytes(totalStorage)}
              </CardTitle>
              <div className="mt-2">
                <div className="w-full bg-gray-200 rounded-full h-2">
                  <div
                    className="bg-gradient-to-r from-blue-500 to-purple-600 h-2 rounded-full"
                    style={{ width: `${Math.min(storagePercentage, 100)}%` }}
                  />
                </div>
                <p className="text-xs text-gray-500 mt-1">
                  {storagePercentage.toFixed(1)}% of 1GB used
                </p>
              </div>
            </CardHeader>
          </Card>
        </div>

        {/* Actions Bar */}
        <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4 mb-6">
          <div className="flex flex-wrap items-center gap-2">
            <Button
              onClick={() => setShowUploadModal(true)}
              className="bg-gradient-to-r from-blue-500 to-purple-600"
            >
              <Upload className="w-4 h-4 mr-2" />
              Upload Files
            </Button>
            <Button variant="outline" onClick={handleCreateFolder}>
              <FolderPlus className="w-4 h-4 mr-2" />
              New Folder
            </Button>
          </div>

          <div className="flex items-center gap-2 w-full sm:w-auto">
            <div className="relative flex-1 sm:flex-initial">
              <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 w-4 h-4 text-gray-400" />
              <input
                type="text"
                placeholder="Search files..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="pl-10 pr-4 py-2 border rounded-lg w-full sm:w-64 focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>
            <div className="flex border rounded-lg">
              <button
                onClick={() => setViewMode("grid")}
                className={`p-2 cursor-pointer hover:bg-gray-50 transition-colors ${
                  viewMode === "grid" ? "bg-gray-100" : ""
                }`}
              >
                <Grid3x3 className="w-4 h-4" />
              </button>
              <button
                onClick={() => setViewMode("list")}
                className={`p-2 cursor-pointer hover:bg-gray-50 transition-colors ${
                  viewMode === "list" ? "bg-gray-100" : ""
                }`}
              >
                <List className="w-4 h-4" />
              </button>
            </div>
          </div>
        </div>

        {/* Breadcrumbs */}
        <div className="flex items-center space-x-2 mb-4 text-sm">
          {getBreadcrumbs().map((crumb, index) => (
            <div key={crumb.id || "root"} className="flex items-center">
              {index > 0 && (
                <ChevronRight className="w-4 h-4 text-gray-400 mx-1" />
              )}
              <button
                onClick={() => {
                  if (crumb.id === null) {
                    handleBackToRoot();
                  } else {
                    handleFolderClick(crumb.id);
                  }
                }}
                className={`flex items-center space-x-1 hover:text-blue-600 transition-colors cursor-pointer ${
                  index === getBreadcrumbs().length - 1
                    ? "text-blue-600 font-medium"
                    : "text-gray-600"
                }`}
              >
                {crumb.id === null && <Home className="w-4 h-4" />}
                <span>{crumb.name}</span>
              </button>
            </div>
          ))}
        </div>

        {/* Files Grid/List */}
        {filteredFiles.length === 0 && (
          <Card className="p-12">
            <div className="text-center space-y-4">
              <div className="w-16 h-16 bg-gray-100 rounded-full flex items-center justify-center mx-auto">
                <Cloud className="w-8 h-8 text-gray-400" />
              </div>
              <div>
                <h3 className="text-lg font-semibold text-gray-900">
                  No files found
                </h3>
                <p className="text-gray-500">
                  {searchQuery
                    ? "Try a different search term"
                    : "Upload your first file to get started"}
                </p>
              </div>
              <Button
                onClick={() => setShowUploadModal(true)}
                className="bg-gradient-to-r from-blue-500 to-purple-600"
              >
                <Upload className="w-4 h-4 mr-2" />
                Upload Files
              </Button>
            </div>
          </Card>
        )}

        {filteredFiles.length > 0 && viewMode === "grid" && (
          <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 gap-4">
            {filteredFiles.map((file) => (
              <Card
                key={file.id}
                className="group hover:shadow-lg transition-all duration-200 hover:-translate-y-1"
              >
                <CardContent className="p-4">
                  <div className="flex flex-col items-center text-center space-y-3">
                    {/* Icon */}
                    <div
                      onClick={() =>
                        file.type === "folder" && handleFolderClick(file.id)
                      }
                      onKeyDown={(e) => {
                        if (
                          file.type === "folder" &&
                          (e.key === "Enter" || e.key === " ")
                        ) {
                          e.preventDefault();
                          handleFolderClick(file.id);
                        }
                      }}
                      role={file.type === "folder" ? "button" : undefined}
                      tabIndex={file.type === "folder" ? 0 : undefined}
                      aria-label={
                        file.type === "folder"
                          ? `Open ${file.name} folder`
                          : undefined
                      }
                      className={`w-16 h-16 rounded-lg flex items-center justify-center ${
                        file.type === "folder"
                          ? "bg-blue-100 cursor-pointer hover:bg-blue-200 transition-colors"
                          : "bg-gray-100"
                      }`}
                    >
                      {file.type === "folder" ? (
                        <Folder className="w-8 h-8 text-blue-500" />
                      ) : (
                        getFileIcon(file.mimeType)
                      )}
                    </div>

                    {/* Name */}
                    <div className="w-full">
                      <p className="font-medium text-sm truncate">
                        {file.name}
                      </p>
                      {file.size && (
                        <p className="text-xs text-gray-500">
                          {formatBytes(file.size)}
                        </p>
                      )}
                      {file.isShared && (
                        <span className="inline-flex items-center text-xs text-blue-600 mt-1">
                          <Share2 className="w-3 h-3 mr-1" />
                          Shared
                        </span>
                      )}
                    </div>

                    {/* Actions */}
                    <div className="flex gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                      {file.type === "file" && (
                        <Button
                          size="sm"
                          variant="ghost"
                          onClick={() => handleDownload(file.id)}
                        >
                          <Download className="w-3 h-3" />
                        </Button>
                      )}
                      <Button
                        size="sm"
                        variant="ghost"
                        onClick={() => handleShare(file.id)}
                      >
                        <Share2 className="w-3 h-3" />
                      </Button>
                      <Button
                        size="sm"
                        variant="ghost"
                        onClick={() => handleDelete(file.id)}
                      >
                        <Trash2 className="w-3 h-3 text-red-500" />
                      </Button>
                    </div>
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>
        )}

        {filteredFiles.length > 0 && viewMode === "list" && (
          <Card>
            <div className="divide-y">
              {filteredFiles.map((file) => (
                <div
                  key={file.id}
                  className="flex items-center justify-between p-4 hover:bg-gray-50 transition-colors"
                >
                  <div className="flex items-center space-x-3 flex-1 min-w-0">
                    <div
                      onClick={() =>
                        file.type === "folder" && handleFolderClick(file.id)
                      }
                      onKeyDown={(e) => {
                        if (
                          file.type === "folder" &&
                          (e.key === "Enter" || e.key === " ")
                        ) {
                          e.preventDefault();
                          handleFolderClick(file.id);
                        }
                      }}
                      role={file.type === "folder" ? "button" : undefined}
                      tabIndex={file.type === "folder" ? 0 : undefined}
                      aria-label={
                        file.type === "folder"
                          ? `Open ${file.name} folder`
                          : undefined
                      }
                      className={`w-10 h-10 rounded-lg flex items-center justify-center flex-shrink-0 ${
                        file.type === "folder"
                          ? "bg-blue-100 cursor-pointer hover:bg-blue-200 transition-colors"
                          : "bg-gray-100"
                      }`}
                    >
                      {file.type === "folder" ? (
                        <Folder className="w-5 h-5 text-blue-500" />
                      ) : (
                        getFileIcon(file.mimeType)
                      )}
                    </div>
                    <div className="flex-1 min-w-0">
                      <p className="font-medium truncate">{file.name}</p>
                      <div className="flex items-center gap-3 text-sm text-gray-500">
                        {file.size && <span>{formatBytes(file.size)}</span>}
                        <span>{file.createdAt}</span>
                        {file.isShared && (
                          <span className="inline-flex items-center text-blue-600">
                            <Share2 className="w-3 h-3 mr-1" />
                            Shared
                          </span>
                        )}
                      </div>
                    </div>
                  </div>
                  <div className="flex items-center gap-1">
                    {file.type === "file" && (
                      <Button
                        size="sm"
                        variant="ghost"
                        onClick={() => handleDownload(file.id)}
                      >
                        <Download className="w-4 h-4" />
                      </Button>
                    )}
                    <Button
                      size="sm"
                      variant="ghost"
                      onClick={() => handleShare(file.id)}
                    >
                      <Share2 className="w-4 h-4" />
                    </Button>
                    <Button
                      size="sm"
                      variant="ghost"
                      onClick={() => handleDelete(file.id)}
                    >
                      <Trash2 className="w-4 h-4 text-red-500" />
                    </Button>
                  </div>
                </div>
              ))}
            </div>
          </Card>
        )}
      </div>

      {/* Upload Modal */}
      {showUploadModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
          <Card className="w-full max-w-lg">
            <CardHeader>
              <div className="flex justify-between items-center">
                <CardTitle>Upload Files</CardTitle>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => setShowUploadModal(false)}
                >
                  <X className="w-4 h-4" />
                </Button>
              </div>
              <CardDescription>
                Drag and drop files here or click to browse
              </CardDescription>
            </CardHeader>
            <CardContent>
              <section
                onDragOver={handleDragOver}
                onDragLeave={handleDragLeave}
                onDrop={handleDrop}
                aria-label="File upload drop zone"
                className={`border-2 border-dashed rounded-lg p-12 text-center transition-colors ${
                  isDragging
                    ? "border-blue-500 bg-blue-50"
                    : "border-gray-300 hover:border-gray-400"
                }`}
              >
                <input
                  type="file"
                  multiple
                  onChange={(e) => handleFileUpload(e.target.files)}
                  className="hidden"
                  id="file-upload"
                  aria-label="File upload input"
                />
                <label
                  htmlFor="file-upload"
                  className="cursor-pointer"
                  aria-label="Choose files to upload"
                >
                  <div className="space-y-4">
                    <div className="w-16 h-16 bg-blue-100 rounded-full flex items-center justify-center mx-auto">
                      <Upload className="w-8 h-8 text-blue-500" />
                    </div>
                    <div>
                      <p className="text-lg font-medium text-gray-900">
                        {isDragging ? "Drop files here" : "Choose files"}
                      </p>
                      <p className="text-sm text-gray-500 mt-1">
                        or drag and drop
                      </p>
                    </div>
                    <span className="inline-flex items-center justify-center gap-2 whitespace-nowrap rounded-md text-sm font-medium transition-all cursor-pointer px-8 py-3 bg-gradient-to-r from-blue-500 to-purple-600 text-white hover:opacity-90">
                      Browse Files
                    </span>
                  </div>
                </label>
              </section>
              <p className="text-xs text-gray-500 mt-4 text-center">
                Max file size: 100MB • Supports all file types
              </p>
            </CardContent>
          </Card>
        </div>
      )}
    </div>
  );
}

export default function Dashboard() {
  return (
    <ProtectedRoute>
      <DashboardContent />
    </ProtectedRoute>
  );
}
