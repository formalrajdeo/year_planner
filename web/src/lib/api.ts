export interface PublicUser {
  id: string;
  email: string;
  displayName: string;
  createdAt: string;
  updatedAt: string;
}

export interface AuthSession {
  token: string;
  user: PublicUser;
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

export interface AuthCredentials {
  email: string;
  password: string;
  displayName?: string;
}

export function register(credentials: AuthCredentials): Promise<AuthSession> {
  return request<AuthSession>("/auth/register", {
    method: "POST",
    body: JSON.stringify(credentials),
  });
}

export function login(credentials: AuthCredentials): Promise<AuthSession> {
  return request<AuthSession>("/auth/login", {
    method: "POST",
    body: JSON.stringify({
      email: credentials.email,
      password: credentials.password,
    }),
  });
}

export function getCurrentUser(token: string): Promise<PublicUser> {
  return request<PublicUser>("/auth/me", {
    headers: { Authorization: `Bearer ${token}` },
  });
}

export function logout(token: string): Promise<{ message: string }> {
  return request<{ message: string }>("/auth/logout", {
    method: "POST",
    headers: { Authorization: `Bearer ${token}` },
  });
}
