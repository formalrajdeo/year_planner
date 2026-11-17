export interface PublicUser {
  id: string;
  email: string;
  displayName: string;
  createdAt: string;
  updatedAt: string;
}

export interface AuthSession {
  user: PublicUser;
}

export interface PasswordRecoveryResponse {
  message: string;
}

interface ApiError {
  error?: string;
}

export class ApiRequestError extends Error {
  constructor(
    message: string,
    readonly status: number,
  ) {
    super(message);
    this.name = "ApiRequestError";
  }
}

const apiBaseUrl = import.meta.env.VITE_API_BASE_URL ?? "/api/v1";

async function request<T>(
  path: string,
  options: RequestInit = {},
): Promise<T> {
  const headers = new Headers(options.headers);
  if (options.body && !headers.has("Content-Type")) {
    headers.set("Content-Type", "application/json");
  }

  const response = await fetch(`${apiBaseUrl}${path}`, {
    ...options,
    credentials: "include",
    headers,
  });

  if (!response.ok) {
    const rawBody = await response.text();
    let message = `Request failed (${response.status})`;
    if (rawBody) {
      try {
        const body = JSON.parse(rawBody) as ApiError;
        if (body.error) message = body.error;
      } catch {
        message = `Request failed (${response.status})`;
      }
    }
    throw new ApiRequestError(message, response.status);
  }

  return (await response.json()) as T;
}

function csrfHeaders(): HeadersInit {
  const csrf = document.cookie
    .split("; ")
    .find((cookie) => cookie.startsWith("yearplanner_csrf="))
    ?.slice("yearplanner_csrf=".length);

  return csrf ? { "X-CSRF-Token": decodeURIComponent(csrf) } : {};
}

export interface AuthCredentials {
  email: string;
  password: string;
  displayName?: string;
}

export function register(credentials: AuthCredentials): Promise<AuthSession> {
  return request<AuthSession>("/auth/register", {
    method: "POST",
    headers: csrfHeaders(),
    body: JSON.stringify(credentials),
  });
}

export function login(credentials: AuthCredentials): Promise<AuthSession> {
  return request<AuthSession>("/auth/login", {
    method: "POST",
    headers: csrfHeaders(),
    body: JSON.stringify({
      email: credentials.email,
      password: credentials.password,
    }),
  });
}

export function requestPasswordReset(email: string): Promise<PasswordRecoveryResponse> {
  return request<PasswordRecoveryResponse>("/auth/password/forgot", {
    method: "POST",
    headers: csrfHeaders(),
    body: JSON.stringify({ email }),
  });
}

export function resetPassword(
  token: string,
  password: string,
): Promise<PasswordRecoveryResponse> {
  return request<PasswordRecoveryResponse>("/auth/password/reset", {
    method: "POST",
    headers: csrfHeaders(),
    body: JSON.stringify({ token, password }),
  });
}

export function getCurrentUser(): Promise<PublicUser> {
  return request<PublicUser>("/auth/me");
}

export function logout(): Promise<{ message: string }> {
  return request<{ message: string }>("/auth/logout", {
    method: "POST",
    headers: csrfHeaders(),
  });
}
