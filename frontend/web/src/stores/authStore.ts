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
import { clearUserScopedData } from './storageScope';

// 14-day trial period
const TRIAL_DAYS = 14;

function isUserInTrial(createdAt?: string): boolean {
  if (!createdAt) return false;
  const createdDate = new Date(createdAt);
  const now = new Date();
  const trialEnd = new Date(createdDate.getTime() + TRIAL_DAYS * 24 * 60 * 60 * 1000);
  return now < trialEnd;
}

const ACCESS_TOKEN_KEY = 'access_token';
const REFRESH_TOKEN_KEY = 'refresh_token';

// These helpers run at module-load time, so they must be safe in non-browser
// (SSR / test) environments where localStorage and document are undefined.
const hasLocalStorage = typeof localStorage !== 'undefined';
const hasDocument = typeof document !== 'undefined';

// Получаем токены из localStorage при инициализации
function getStoredAccessToken(): string | null {
  return hasLocalStorage ? localStorage.getItem(ACCESS_TOKEN_KEY) : null;
}

function getStoredRefreshToken(): string | null {
  return hasLocalStorage ? localStorage.getItem(REFRESH_TOKEN_KEY) : null;
}

function setAccessTokenCookie(token: string): void {
  if (!hasDocument) return;
  document.cookie = `${ACCESS_TOKEN_KEY}=${token}; path=/; SameSite=Lax`;
}

function removeAccessTokenCookie(): void {
  if (!hasDocument) return;
  document.cookie = `${ACCESS_TOKEN_KEY}=; path=/; SameSite=Lax; Max-Age=0`;
}

function storeTokens(accessToken: string, refreshToken: string): void {
  if (hasLocalStorage) {
    localStorage.setItem(ACCESS_TOKEN_KEY, accessToken);
    localStorage.setItem(REFRESH_TOKEN_KEY, refreshToken);
  }
  setAccessTokenCookie(accessToken);
}

function removeTokens(): void {
  if (hasLocalStorage) {
    localStorage.removeItem(ACCESS_TOKEN_KEY);
    localStorage.removeItem(REFRESH_TOKEN_KEY);
  }
  removeAccessTokenCookie();
}

interface AuthState {
  user: User | null;
  accessToken: string | null;
  refreshToken: string | null;
  isAuthenticated: boolean;
  hasSubscription: boolean;
  subscriptionEnd: number | null;
  isInTrial: boolean;
  isLoading: boolean;
  error: string | null;

  // Actions
  register: (username: string, email: string, password: string) => Promise<void>;
  login: (email: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
  checkAuth: () => Promise<void>;
  setSubscription: (has: boolean) => void;
  clearError: () => void;
}

const initialToken = getStoredAccessToken();
if (initialToken) {
  setAccessTokenCookie(initialToken);
}

export const useAuthStore = create<AuthState>((set, get) => ({
  user: null,
  accessToken: initialToken,
  refreshToken: getStoredRefreshToken(),
  isAuthenticated: !!initialToken,
  hasSubscription: false,
  subscriptionEnd: null,
  isInTrial: false,
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
        hasSubscription: false,
        isInTrial: true, // New users start in trial
        isLoading: false,
        error: null,
      });
      // Check subscription status (new users don't have one)
      try {
        const userResponse = await getCurrentUser(response.data.access_token);
        set({
          hasSubscription: userResponse.data.has_subscription || false,
          subscriptionEnd: userResponse.data.subscription_end || null,
          isInTrial: isUserInTrial(userResponse.data.created_at),
          user: { ...response.data.user, subscription_end: userResponse.data.subscription_end },
        });
      } catch {
        // ignore
      }
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
        hasSubscription: false,
        isInTrial: false,
        isLoading: false,
        error: null,
      });
      // Immediately check subscription status from server
      try {
        const userResponse = await getCurrentUser(response.data.access_token);
        set({
          hasSubscription: userResponse.data.has_subscription || false,
          subscriptionEnd: userResponse.data.subscription_end || null,
          isInTrial: isUserInTrial(userResponse.data.created_at),
          isAuthenticated: true,
          error: null,
        });
      } catch {
        // ignore - will check on next page load
      }
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
      clearUserScopedData();
      removeTokens();
      set({
        user: null,
        accessToken: null,
        refreshToken: null,
        isAuthenticated: false,
        hasSubscription: false,
        subscriptionEnd: null,
        error: null,
      });
    }
  },

  setSubscription: (has: boolean) => {
    set({ hasSubscription: has });
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
          const userResponse = await getCurrentUser(response.data.access_token);
          set({
            user: {
              id: userResponse.data.user_id,
              username: userResponse.data.username,
              email: userResponse.data.email,
              subscription_end: userResponse.data.subscription_end,
            },
            hasSubscription: userResponse.data.has_subscription || false,
            subscriptionEnd: userResponse.data.subscription_end || null,
            isAuthenticated: true,
            error: null,
          });
          return;
        } catch (error) {
          removeTokens();
          set({
            user: null,
            accessToken: null,
            refreshToken: null,
            isAuthenticated: false,
            hasSubscription: false,
            subscriptionEnd: null,
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
          id: response.data.user_id,
          username: response.data.username,
          email: response.data.email,
          subscription_end: response.data.subscription_end,
          created_at: (response.data as any).created_at,
        },
        hasSubscription: response.data.has_subscription || false,
        subscriptionEnd: response.data.subscription_end || null,
        isInTrial: isUserInTrial(response.data.created_at),
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
          const userResponse = await getCurrentUser(refreshResponse.data.access_token);
          set({
            user: {
              id: userResponse.data.user_id,
              username: userResponse.data.username,
              email: userResponse.data.email,
              subscription_end: userResponse.data.subscription_end,
            },
            hasSubscription: userResponse.data.has_subscription || false,
            subscriptionEnd: userResponse.data.subscription_end || null,
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
            hasSubscription: false,
            subscriptionEnd: null,
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
