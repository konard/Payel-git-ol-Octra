/**
 * Auth Store
 * Zustand store для управления состоянием авторизации
 */

import { create } from 'zustand';
import {
  registerUser,
  loginUser,
  logoutUser,
  getCurrentUser,
  refreshToken as refreshAccessToken,
  type User,
} from '../services/authService';

const ACCESS_TOKEN_KEY = 'access_token';
const REFRESH_TOKEN_KEY = 'refresh_token';

// Получаем токены из localStorage при инициализации
function getStoredAccessToken(): string | null {
  return localStorage.getItem(ACCESS_TOKEN_KEY);
}

function getStoredRefreshToken(): string | null {
  return localStorage.getItem(REFRESH_TOKEN_KEY);
}

function storeTokens(accessToken: string, refreshToken: string): void {
  localStorage.setItem(ACCESS_TOKEN_KEY, accessToken);
  localStorage.setItem(REFRESH_TOKEN_KEY, refreshToken);
}

function storeAccessToken(token: string): void {
  localStorage.setItem(ACCESS_TOKEN_KEY, token);
}

function removeTokens(): void {
  localStorage.removeItem(ACCESS_TOKEN_KEY);
  localStorage.removeItem(REFRESH_TOKEN_KEY);
}

interface AuthState {
  user: User | null;
  accessToken: string | null;
  refreshToken: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  error: string | null;

  // Actions
  register: (username: string, email: string, password: string) => Promise<void>;
  login: (email: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
  checkAuth: () => Promise<void>;
  clearError: () => void;
}

export const useAuthStore = create<AuthState>((set, get) => ({
  user: null,
  accessToken: getStoredAccessToken(),
  refreshToken: getStoredRefreshToken(),
  isAuthenticated: !!getStoredAccessToken(),
  isLoading: false,
  error: null,

  register: async (username: string, email: string, password: string) => {
    set({ isLoading: true, error: null });
    try {
      const response = await registerUser({ username, email, password });
      storeTokens(response.data.access_token, response.data.refresh_token);
      set({
        user: response.data.user,
        accessToken: response.data.access_token,
        refreshToken: response.data.refresh_token,
        isAuthenticated: true,
        isLoading: false,
        error: null,
      });
    } catch (error) {
      set({
        isLoading: false,
        error: error instanceof Error ? error.message : 'Registration failed',
      });
      throw error;
    }
  },

  login: async (email: string, password: string) => {
    set({ isLoading: true, error: null });
    try {
      const response = await loginUser({ email, password });
      storeTokens(response.data.access_token, response.data.refresh_token);
      set({
        user: response.data.user,
        accessToken: response.data.access_token,
        refreshToken: response.data.refresh_token,
        isAuthenticated: true,
        isLoading: false,
        error: null,
      });
    } catch (error) {
      set({
        isLoading: false,
        error: error instanceof Error ? error.message : 'Login failed',
      });
      throw error;
    }
  },

  logout: async () => {
    try {
      await logoutUser();
    } catch (error) {
      console.error('Logout error:', error);
    } finally {
      removeTokens();
      set({
        user: null,
        accessToken: null,
        refreshToken: null,
        isAuthenticated: false,
        error: null,
      });
    }
  },

  checkAuth: async () => {
    const { accessToken, refreshToken } = get();
    
    // Если нет access токена, пробуем обновить через refresh токен
    if (!accessToken) {
      if (refreshToken) {
        try {
          const response = await refreshAccessToken(refreshToken);
          storeTokens(response.data.access_token, response.data.refresh_token);
          set({
            accessToken: response.data.access_token,
            refreshToken: response.data.refresh_token,
            isAuthenticated: true,
            error: null,
          });
          // Теперь проверяем пользователя
          const userResponse = await getCurrentUser(response.data.access_token);
          set({
            user: {
              id: userResponse.data.user_id.toString(),
              username: userResponse.data.username,
              email: userResponse.data.email,
            },
            isAuthenticated: true,
            error: null,
          });
          return;
        } catch (error) {
          // Refresh токен истек или невалиден
          removeTokens();
          set({
            user: null,
            accessToken: null,
            refreshToken: null,
            isAuthenticated: false,
            error: null,
          });
          return;
        }
      } else {
        set({ isAuthenticated: false, user: null });
        return;
      }
    }

    try {
      const response = await getCurrentUser(accessToken);
      set({
        user: {
          id: response.data.user_id.toString(),
          username: response.data.username,
          email: response.data.email,
        },
        isAuthenticated: true,
        error: null,
      });
    } catch (error) {
      // Если токен истек, пробуем обновить
      if (refreshToken) {
        try {
          const refreshResponse = await refreshAccessToken(refreshToken);
          storeTokens(refreshResponse.data.access_token, refreshResponse.data.refresh_token);
          set({
            accessToken: refreshResponse.data.access_token,
            refreshToken: refreshResponse.data.refresh_token,
          });
          // Повторно проверяем пользователя
          const userResponse = await getCurrentUser(refreshResponse.data.access_token);
          set({
            user: {
              id: userResponse.data.user_id.toString(),
              username: userResponse.data.username,
              email: userResponse.data.email,
            },
            isAuthenticated: true,
            error: null,
          });
        } catch (refreshError) {
          removeTokens();
          set({
            user: null,
            accessToken: null,
            refreshToken: null,
            isAuthenticated: false,
            error: null,
          });
        }
      } else {
        removeTokens();
        set({
          user: null,
          accessToken: null,
          refreshToken: null,
          isAuthenticated: false,
          error: null,
        });
      }
    }
  },

  clearError: () => {
    set({ error: null });
  },
}));
