import { useState } from "react";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { AuthPage, type AuthMode } from "./features/auth/AuthPage";
import { PlannerHome } from "./features/year-board/PlannerHome";
import {
  ApiRequestError,
  getCurrentUser,
  login,
  logout,
  register,
  requestPasswordReset,
  resetPassword,
  type AuthCredentials,
} from "./lib/api";

const initialPasswordResetToken = readPasswordResetToken();

export default function App() {
  const queryClient = useQueryClient();
  const [signedOut, setSignedOut] = useState(false);
  const [authError, setAuthError] = useState("");
  const [logoutError, setLogoutError] = useState("");
  const [isSigningOut, setIsSigningOut] = useState(false);
  const [passwordResetToken, setPasswordResetToken] = useState(initialPasswordResetToken);

  const profileQuery = useQuery({
    queryKey: ["auth", "me"],
    queryFn: getCurrentUser,
    retry: false,
  });

  async function handleAuth(mode: AuthMode, credentials: AuthCredentials) {
    setAuthError("");
    try {
      const nextSession =
        mode === "register"
          ? await register(credentials)
          : await login(credentials);
      setSignedOut(false);
      queryClient.setQueryData(["auth", "me"], nextSession.user);
    } catch (error) {
      setAuthError(errorMessage(error));
    }
  }

  async function handleLogout() {
    setLogoutError("");
    setIsSigningOut(true);
    try {
      await logout();
      setSignedOut(true);
      await queryClient.removeQueries({ queryKey: ["auth", "me"] });
    } catch (error) {
      if (isUnauthorized(error)) {
        setSignedOut(true);
        await queryClient.removeQueries({ queryKey: ["auth", "me"] });
      } else {
        setLogoutError(errorMessage(error));
      }
    } finally {
      setIsSigningOut(false);
    }
  }

  if (!passwordResetToken && !signedOut && profileQuery.isLoading) {
    return <div className="app-loading">Opening your planning space…</div>;
  }

  if (!passwordResetToken && !signedOut && profileQuery.data && !isUnauthorized(profileQuery.error)) {
    return (
      <PlannerHome
        user={profileQuery.data}
        profileLoading={profileQuery.isFetching}
        profileError={profileQuery.isError && !isUnauthorized(profileQuery.error)}
        logoutError={logoutError}
        isSigningOut={isSigningOut}
        onLogout={handleLogout}
      />
    );
  }

  const formError = authError || (
    profileQuery.error && !isUnauthorized(profileQuery.error)
      ? errorMessage(profileQuery.error)
      : ""
  );
  return (
    <AuthPage
      onSubmit={handleAuth}
      onRequestPasswordReset={requestPasswordReset}
      onResetPassword={async (token, password) => {
        await resetPassword(token, password);
        window.history.replaceState({}, "", window.location.pathname + window.location.search);
        setPasswordResetToken(null);
      }}
      passwordResetToken={passwordResetToken}
      error={formError}
    />
  );
}

function readPasswordResetToken() {
  const fragment = window.location.hash;
  if (!fragment.startsWith("#/reset-password?")) {
    return null;
  }
  return new URLSearchParams(fragment.slice(fragment.indexOf("?") + 1)).get("token");
}

function isUnauthorized(error: unknown) {
  return error instanceof ApiRequestError && error.status === 401;
}

function errorMessage(error: unknown) {
  return error instanceof Error
    ? error.message
    : "Something went wrong. Please try again.";
}
