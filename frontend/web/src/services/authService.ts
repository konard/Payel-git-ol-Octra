/**
 * Auth API Service
 * Сервис для работы с API авторизации
 */

const AUTH_API_URL = import.meta.env.VITE_AUTH_URL || '/auth';
const API_GATEWAY_URL = import.meta.env.VITE_API_GATEWAY_URL || '/api';

export interface UserRegisterRequest {
  username: string;
  email: string;
  password: string;
}

export interface UserLoginRequest {
  email: string;
  password: string;
}

export interface RefreshTokenRequest {
  refresh_token: string;
}

export interface User {
  id: string;
  username: string;
  email: string;
  subscription_end?: number;
  created_at?: string;
}

export interface AuthResponse {
  status: string;
  data: {
    access_token: string;
    refresh_token: string;
    user: User;
  };
}

export interface UserResponse {
  status: string;
  data: {
    user_id: string;
    username: string;
    email: string;
    has_subscription?: boolean;
    subscription_end?: number;
  };
}

export interface ErrorResponse {
  status: string;
  error: string;
}

export interface RefreshResponse {
  status: string;
  data: {
    access_token: string;
    refresh_token: string;
  };
}

/**
 * Регистрация пользователя
 */
export async function registerUser(request: UserRegisterRequest): Promise<AuthResponse> {
  const response = await fetch(`${AUTH_API_URL}/register`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(request),
    credentials: 'include',
  });

  if (!response.ok) {
    const error: ErrorResponse = await response.json();
    throw new Error(error.error || 'Registration failed');
  }

  return response.json();
}

/**
 * Авторизация пользователя
 */
export async function loginUser(request: UserLoginRequest): Promise<AuthResponse> {
  const response = await fetch(`${AUTH_API_URL}/login`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(request),
    credentials: 'include',
  });

  if (!response.ok) {
    const error: ErrorResponse = await response.json();
    throw new Error(error.error || 'Login failed');
  }

  return response.json();
}

/**
 * Обновление токена
 */
export async function refreshToken(refreshToken: string): Promise<RefreshResponse> {
  const response = await fetch(`${AUTH_API_URL}/refresh`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ refresh_token: refreshToken }),
    credentials: 'include',
  });

  if (!response.ok) {
    const error: ErrorResponse = await response.json();
    throw new Error(error.error || 'Token refresh failed');
  }

  return response.json();
}

/**
 * Получение информации о текущем пользователе
 */
export async function getCurrentUser(accessToken: string): Promise<UserResponse> {
  const response = await fetch(`${AUTH_API_URL}/me`, {
    method: 'GET',
    headers: {
      'Authorization': `Bearer ${accessToken}`,
    },
    credentials: 'include',
  });

  if (!response.ok) {
    const error: ErrorResponse = await response.json();
    throw new Error(error.error || 'Failed to get user info');
  }

  return response.json();
}

/**
 * Выход из системы
 */
export async function logoutUser(): Promise<void> {
  await fetch(`${AUTH_API_URL}/logout`, {
    method: 'POST',
    credentials: 'include',
  });
}
