import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { fileAPI } from "../lib/api";

// Query keys
export const fileKeys = {
  all: ["files"] as const,
  lists: () => [...fileKeys.all, "list"] as const,
  list: (parentId?: string) =>
    [...fileKeys.lists(), parentId ?? "root"] as const,
};

// Fetch files list
export function useFiles(parentId?: string) {
  return useQuery({
    queryKey: fileKeys.list(parentId),
    queryFn: () => fileAPI.list(parentId),
    staleTime: 30 * 1000, // 30 seconds
  });
}

// Upload file mutation
export function useUploadFile() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ file, parentId }: { file: File; parentId?: string }) =>
      fileAPI.upload(file, parentId),
    onSuccess: (_, variables) => {
      // Invalidate and refetch files list
      queryClient.invalidateQueries({
        queryKey: fileKeys.list(variables.parentId),
      });
    },
  });
}

// Delete file mutation
export function useDeleteFile() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (fileId: string) => fileAPI.delete(fileId),
    onSuccess: () => {
      // Invalidate all file lists
      queryClient.invalidateQueries({ queryKey: fileKeys.lists() });
    },
  });
}

// Create folder mutation
export function useCreateFolder() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      folderName,
      parentId,
    }: {
      folderName: string;
      parentId?: string;
    }) => fileAPI.createFolder({ folderName, parentId }),
    onSuccess: (_, variables) => {
      // Invalidate the parent folder's file list
      queryClient.invalidateQueries({
        queryKey: fileKeys.list(variables.parentId),
      });
    },
  });
}
