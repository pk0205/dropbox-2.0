import { useState } from "react";
import { useNavigate } from "react-router";
import { PublicRoute } from "~/components/PublicRoute";
import { Button } from "~/components/ui/button";
import { Input } from "~/components/ui/input";
import { Label } from "~/components/ui/label";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "~/components/ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "~/components/ui/tabs";
import { useAuth } from "~/contexts/AuthContext";
import { AlertCircle, Cloud, Loader2 } from "lucide-react";
import type { Route } from "./+types/auth";

export function meta({}: Route.MetaArgs) {
  return [
    { title: "Login - Dropbox 2.0" },
    { name: "description", content: "Login or sign up to Dropbox 2.0" },
  ];
}

function AuthContent() {
  const navigate = useNavigate();
  const { login, signup } = useAuth();
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState("");

  // Login form state
  const [loginData, setLoginData] = useState({
    emailOrUsername: "",
    password: "",
  });

  // Signup form state
  const [signupData, setSignupData] = useState({
    firstName: "",
    lastName: "",
    username: "",
    email: "",
    password: "",
    confirmPassword: "",
  });

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    setIsLoading(true);

    try {
      await login(loginData.emailOrUsername, loginData.password);
      navigate("/dashboard");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Login failed");
    } finally {
      setIsLoading(false);
    }
  };

  const handleSignup = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");

    // Validation
    if (signupData.password !== signupData.confirmPassword) {
      setError("Passwords do not match");
      return;
    }

    if (signupData.password.length < 6) {
      setError("Password must be at least 6 characters");
      return;
    }

    setIsLoading(true);

    try {
      await signup(
        signupData.firstName,
        signupData.lastName,
        signupData.username,
        signupData.email,
        signupData.password
      );
      navigate("/dashboard");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Signup failed");
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-blue-50 via-white to-purple-50 p-4">
      <div className="w-full max-w-6xl grid md:grid-cols-2 gap-8 items-center">
        {/* Left side - Branding */}
        <div className="hidden md:flex flex-col justify-center space-y-6">
          <div className="flex items-center space-x-3">
            <div className="w-12 h-12 bg-blue-500 rounded-lg flex items-center justify-center">
              <Cloud className="w-8 h-8 text-white" />
            </div>
            <h1 className="text-4xl font-bold bg-gradient-to-r from-blue-600 to-purple-600 bg-clip-text text-transparent">
              Dropbox 2.0
            </h1>
          </div>

          <div className="space-y-4">
            <h2 className="text-3xl font-bold text-gray-800">
              Your files, anywhere you go
            </h2>
            <p className="text-lg text-gray-600">
              Store, share, and collaborate on files from any device. Experience
              fast, secure, and reliable cloud storage.
            </p>
          </div>

          <div className="grid grid-cols-2 gap-4 pt-4">
            <div className="space-y-2">
              <div className="w-10 h-10 bg-blue-100 rounded-lg flex items-center justify-center">
                <span className="text-2xl">🚀</span>
              </div>
              <h3 className="font-semibold text-gray-800">Lightning Fast</h3>
              <p className="text-sm text-gray-600">
                Optimized uploads with chunking
              </p>
            </div>
            <div className="space-y-2">
              <div className="w-10 h-10 bg-purple-100 rounded-lg flex items-center justify-center">
                <span className="text-2xl">🔒</span>
              </div>
              <h3 className="font-semibold text-gray-800">Secure</h3>
              <p className="text-sm text-gray-600">
                End-to-end encrypted storage
              </p>
            </div>
            <div className="space-y-2">
              <div className="w-10 h-10 bg-green-100 rounded-lg flex items-center justify-center">
                <span className="text-2xl">🔗</span>
              </div>
              <h3 className="font-semibold text-gray-800">Easy Sharing</h3>
              <p className="text-sm text-gray-600">Share files with anyone</p>
            </div>
            <div className="space-y-2">
              <div className="w-10 h-10 bg-orange-100 rounded-lg flex items-center justify-center">
                <span className="text-2xl">💾</span>
              </div>
              <h3 className="font-semibold text-gray-800">Deduplication</h3>
              <p className="text-sm text-gray-600">Save space automatically</p>
            </div>
          </div>
        </div>

        {/* Right side - Auth forms */}
        <Card className="w-full shadow-xl">
          <CardHeader className="space-y-1">
            <div className="md:hidden flex items-center justify-center mb-4">
              <div className="w-12 h-12 bg-blue-500 rounded-lg flex items-center justify-center">
                <Cloud className="w-8 h-8 text-white" />
              </div>
            </div>
            <CardTitle className="text-2xl text-center">
              Welcome to Dropbox 2.0
            </CardTitle>
            <CardDescription className="text-center">
              Login or create an account to get started
            </CardDescription>
          </CardHeader>
          <CardContent>
            <Tabs defaultValue="login" className="w-full">
              <TabsList className="grid w-full grid-cols-2 mb-6">
                <TabsTrigger value="login">Login</TabsTrigger>
                <TabsTrigger value="signup">Sign Up</TabsTrigger>
              </TabsList>

              {/* Error Message */}
              {error && (
                <div className="mb-4 p-3 rounded-lg bg-red-50 border border-red-200 flex items-start space-x-2">
                  <AlertCircle className="w-5 h-5 text-red-500 flex-shrink-0 mt-0.5" />
                  <p className="text-sm text-red-700">{error}</p>
                </div>
              )}

              {/* Login Tab */}
              <TabsContent value="login">
                <form onSubmit={handleLogin} className="space-y-4">
                  <div className="space-y-2">
                    <Label htmlFor="login-email">Email or Username</Label>
                    <Input
                      id="login-email"
                      type="text"
                      placeholder="Enter your email or username"
                      value={loginData.emailOrUsername}
                      onChange={(e) =>
                        setLoginData({
                          ...loginData,
                          emailOrUsername: e.target.value,
                        })
                      }
                      required
                      disabled={isLoading}
                    />
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="login-password">Password</Label>
                    <Input
                      id="login-password"
                      type="password"
                      placeholder="Enter your password"
                      value={loginData.password}
                      onChange={(e) =>
                        setLoginData({ ...loginData, password: e.target.value })
                      }
                      required
                      disabled={isLoading}
                    />
                  </div>
                  <Button type="submit" className="w-full" disabled={isLoading}>
                    {isLoading ? (
                      <>
                        <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                        Logging in...
                      </>
                    ) : (
                      "Login"
                    )}
                  </Button>
                </form>
              </TabsContent>

              {/* Signup Tab */}
              <TabsContent value="signup">
                <form onSubmit={handleSignup} className="space-y-4">
                  <div className="grid grid-cols-2 gap-4">
                    <div className="space-y-2">
                      <Label htmlFor="signup-firstname">First Name</Label>
                      <Input
                        id="signup-firstname"
                        type="text"
                        placeholder="John"
                        value={signupData.firstName}
                        onChange={(e) =>
                          setSignupData({
                            ...signupData,
                            firstName: e.target.value,
                          })
                        }
                        required
                        disabled={isLoading}
                      />
                    </div>
                    <div className="space-y-2">
                      <Label htmlFor="signup-lastname">Last Name</Label>
                      <Input
                        id="signup-lastname"
                        type="text"
                        placeholder="Doe"
                        value={signupData.lastName}
                        onChange={(e) =>
                          setSignupData({
                            ...signupData,
                            lastName: e.target.value,
                          })
                        }
                        required
                        disabled={isLoading}
                      />
                    </div>
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="signup-username">Username</Label>
                    <Input
                      id="signup-username"
                      type="text"
                      placeholder="johndoe"
                      value={signupData.username}
                      onChange={(e) =>
                        setSignupData({
                          ...signupData,
                          username: e.target.value,
                        })
                      }
                      required
                      disabled={isLoading}
                    />
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="signup-email">Email</Label>
                    <Input
                      id="signup-email"
                      type="email"
                      placeholder="john@example.com"
                      value={signupData.email}
                      onChange={(e) =>
                        setSignupData({ ...signupData, email: e.target.value })
                      }
                      required
                      disabled={isLoading}
                    />
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="signup-password">Password</Label>
                    <Input
                      id="signup-password"
                      type="password"
                      placeholder="At least 6 characters"
                      value={signupData.password}
                      onChange={(e) =>
                        setSignupData({
                          ...signupData,
                          password: e.target.value,
                        })
                      }
                      required
                      disabled={isLoading}
                    />
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="signup-confirm">Confirm Password</Label>
                    <Input
                      id="signup-confirm"
                      type="password"
                      placeholder="Confirm your password"
                      value={signupData.confirmPassword}
                      onChange={(e) =>
                        setSignupData({
                          ...signupData,
                          confirmPassword: e.target.value,
                        })
                      }
                      required
                      disabled={isLoading}
                    />
                  </div>
                  <Button type="submit" className="w-full" disabled={isLoading}>
                    {isLoading ? (
                      <>
                        <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                        Creating account...
                      </>
                    ) : (
                      "Create Account"
                    )}
                  </Button>
                </form>
              </TabsContent>
            </Tabs>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}

export default function Auth() {
  return (
    <PublicRoute>
      <AuthContent />
    </PublicRoute>
  );
}
