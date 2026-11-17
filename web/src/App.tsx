import { useEffect, useState } from "react";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { AuthPage, type AuthMode } from "./features/auth/AuthPage";
import { PlannerHome } from "./features/year-board/PlannerHome";
import {
  ApiRequestError,
  getCurrentUser,
  login,
  logout,
  register,
  type AuthCredentials,
  type AuthSession,
} from "./lib/api";

export default function App() {
  const queryClient = useQueryClient();
  const [session, setSession] = useState<AuthSession | null>(null);
  const [authError, setAuthError] = useState("");
  const [logoutError, setLogoutError] = useState("");
  const [isSigningOut, setIsSigningOut] = useState(false);

  const profileQuery = useQuery({
    queryKey: ["auth", "me", session?.token],
    queryFn: () => getCurrentUser(session!.token),
    enabled: Boolean(session?.token),
    retry: false,
  });
  const sessionExpired = isUnauthorized(profileQuery.error);

  useEffect(() => {
    if (sessionExpired) {
      void queryClient.removeQueries({
        queryKey: ["auth", "me", session?.token],
      });
    }
  }, [queryClient, session?.token, sessionExpired]);

  async function handleAuth(mode: AuthMode, credentials: AuthCredentials) {
    setAuthError("");
    try {
      const nextSession =
        mode === "register"
          ? await register(credentials)
          : await login(credentials);
      setSession(nextSession);
      await queryClient.invalidateQueries({ queryKey: ["auth", "me"] });
    } catch (error) {
      setAuthError(errorMessage(error));
    }
  }

  async function handleLogout() {
    if (!session) return;
    setLogoutError("");
    setIsSigningOut(true);
    try {
      await logout(session.token);
      setSession(null);
      await queryClient.removeQueries({ queryKey: ["auth", "me"] });
    } catch (error) {
      if (isUnauthorized(error)) {
        setSession(null);
        await queryClient.removeQueries({ queryKey: ["auth", "me"] });
      } else {
        setLogoutError(errorMessage(error));
      }
    } finally {
      setIsSigningOut(false);
    }
  }

  if (!session || sessionExpired) {
    return <AuthPage onSubmit={handleAuth} error={authError} />;
  }

  return (
    <PlannerHome
      user={profileQuery.data ?? session.user}
      profileLoading={profileQuery.isLoading}
      profileError={profileQuery.isError && !isUnauthorized(profileQuery.error)}
      logoutError={logoutError}
      isSigningOut={isSigningOut}
      onLogout={handleLogout}
    />
  );
}

function isUnauthorized(error: unknown) {
  return error instanceof ApiRequestError && error.status === 401;
}

function errorMessage(error: unknown) {
  return error instanceof Error
    ? error.message
    : "Something went wrong. Please try again.";
}
