import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { authAPI, type User } from "../lib/api";

// Query keys
export const authKeys = {
  me: ["auth", "me"] as const,
};

// Hook to get current user
export function useMe() {
  return useQuery({
    queryKey: authKeys.me,
    queryFn: authAPI.me,
    retry: false,
    staleTime: 5 * 60 * 1000, // 5 minutes
  });
}

// Hook for login mutation
export function useLogin() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: authAPI.login,
    onSuccess: (user: User) => {
      // Update the me query cache
      queryClient.setQueryData(authKeys.me, user);
    },
  });
}

// Hook for signup mutation
export function useSignup() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: authAPI.signup,
    onSuccess: (user: User) => {
      // Update the me query cache
      queryClient.setQueryData(authKeys.me, user);
    },
  });
}

// Hook for logout mutation
export function useLogout() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: authAPI.logout,
    onSuccess: () => {
      // Clear the me query cache
      queryClient.setQueryData(authKeys.me, null);
      queryClient.invalidateQueries({ queryKey: authKeys.me });
    },
  });
}
