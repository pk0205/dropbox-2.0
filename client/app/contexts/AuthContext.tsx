import React, { createContext, useContext, useMemo } from "react";
import { useMe, useLogin, useSignup, useLogout } from "../hooks/useAuth";

interface User {
  id: string;
  username: string;
  email: string;
  firstName: string;
  lastName: string;
}

interface AuthContextType {
  user: User | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  login: (emailOrUsername: string, password: string) => Promise<void>;
  signup: (
    firstName: string,
    lastName: string,
    username: string,
    email: string,
    password: string
  ) => Promise<void>;
  logout: () => Promise<void>;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function AuthProvider({
  children,
}: {
  readonly children: React.ReactNode;
}) {
  // Use TanStack Query hooks
  const { data: user, isLoading } = useMe();
  const loginMutation = useLogin();
  const signupMutation = useSignup();
  const logoutMutation = useLogout();

  const login = async (emailOrUsername: string, password: string) => {
    await loginMutation.mutateAsync({ emailOrUsername, password });
  };

  const signup = async (
    firstName: string,
    lastName: string,
    username: string,
    email: string,
    password: string
  ) => {
    await signupMutation.mutateAsync({
      firstName,
      lastName,
      username,
      email,
      password,
    });
  };

  const logout = async () => {
    await logoutMutation.mutateAsync();
  };

  const value = useMemo(
    () => ({
      user: user ?? null,
      isAuthenticated: !!user,
      isLoading,
      login,
      signup,
      logout,
    }),
    [user, isLoading, login, signup, logout]
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error("useAuth must be used within an AuthProvider");
  }
  return context;
}
