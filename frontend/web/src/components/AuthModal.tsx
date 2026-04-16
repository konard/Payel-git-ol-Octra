/**
 * AuthModal Component
 * Модальное окно авторизации/регистрации поверх основного экрана
 */

import { useState, useEffect } from 'react';
import { useAuthStore } from '../stores/authStore';
import { useI18n } from '../hooks/useI18n';

type AuthView = 'login' | 'register';

interface AuthModalProps {
  onClose: () => void;
  onAuthSuccess: () => void;
}

export function AuthModal({ onClose, onAuthSuccess }: AuthModalProps) {
  const { t } = useI18n();
  const [view, setView] = useState<AuthView>('login');
  
  // Login state
  const [loginEmail, setLoginEmail] = useState('');
  const [loginPassword, setLoginPassword] = useState('');
  
  // Register state
  const [regUsername, setRegUsername] = useState('');
  const [regEmail, setRegEmail] = useState('');
  const [regPassword, setRegPassword] = useState('');
  const [regConfirmPassword, setRegConfirmPassword] = useState('');
  
  const [formError, setFormError] = useState('');
  const { login, register, isLoading, error, clearError } = useAuthStore();

  useEffect(() => {
    clearError();
    setFormError('');
  }, []);

  useEffect(() => {
    if (error) {
      setFormError(error);
    }
  }, [error]);

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    setFormError('');

    if (!loginEmail.trim()) {
      setFormError(t('auth.emailRequired'));
      return;
    }

    if (!loginPassword) {
      setFormError(t('auth.passwordRequired'));
      return;
    }

    try {
      await login(loginEmail, loginPassword);
      onAuthSuccess();
    } catch (err) {
      console.error('Login error:', err);
    }
  };

  const handleRegister = async (e: React.FormEvent) => {
    e.preventDefault();
    setFormError('');

    if (!regUsername.trim()) {
      setFormError(t('auth.usernameRequired'));
      return;
    }

    if (!regEmail.trim()) {
      setFormError(t('auth.emailRequired'));
      return;
    }

    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    if (!emailRegex.test(regEmail)) {
      setFormError(t('auth.invalidEmail'));
      return;
    }

    if (!regPassword) {
      setFormError(t('auth.passwordRequired'));
      return;
    }

    if (regPassword.length < 6) {
      setFormError(t('auth.passwordMinLength'));
      return;
    }

    if (regPassword !== regConfirmPassword) {
      setFormError(t('auth.passwordsNotMatch'));
      return;
    }

    try {
      await register(regUsername, regEmail, regPassword);
      onAuthSuccess();
    } catch (err) {
      console.error('Registration error:', err);
    }
  };

  const switchToLogin = () => {
    setView('login');
    setFormError('');
    clearError();
  };

  const switchToRegister = () => {
    setView('register');
    setFormError('');
    clearError();
  };

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/50"
      onClick={onClose}
    >
      <div
        className="relative bg-[var(--surface)] border border-[var(--border)] rounded-xl shadow-2xl w-full max-w-[400px] mx-4 flex flex-col overflow-hidden"
        onClick={(e) => e.stopPropagation()}
      >
        {/* Header */}
        <div className="flex items-center justify-between px-5 py-4 border-b border-[var(--border)]">
          <h2 className="text-xl font-semibold text-[var(--text)]">
            {view === 'login' ? t('auth.loginTitle') : t('auth.registerTitle')}
          </h2>
          <button
            onClick={onClose}
            className="p-1 hover:bg-[var(--background)] rounded-md transition-colors"
            aria-label={t('auth.close')}
          >
            <svg width="14" height="14" viewBox="0 0 14 14" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round">
              <line x1="1" y1="1" x2="13" y2="13" />
              <line x1="13" y1="1" x2="1" y2="13" />
            </svg>
          </button>
        </div>

        {/* Content */}
        <div className="flex-1 overflow-y-auto px-6 py-6">
          {formError && (
            <div className="mb-4 p-3 bg-red-500/10 border border-red-500 text-red-500 rounded-md text-sm">
              {formError}
            </div>
          )}

          {view === 'login' ? (
            <form onSubmit={handleLogin} className="flex flex-col gap-4">
              <div className="flex flex-col gap-2">
                <label htmlFor="login-email" className="text-sm font-medium text-[var(--text)]">
                  Email
                </label>
                <input
                  id="login-email"
                  type="email"
                  value={loginEmail}
                  onChange={(e) => setLoginEmail(e.target.value)}
                  placeholder="your@email.com"
                  className="w-full px-3 py-2 bg-[var(--background)] border border-[var(--border)] rounded-md text-[var(--text)] text-sm placeholder:text-[var(--text-muted)] focus:outline-none focus:border-[var(--accent)] transition-colors"
                  disabled={isLoading}
                  required
                  autoComplete="email"
                />
              </div>

              <div className="flex flex-col gap-2">
                <label htmlFor="login-password" className="text-sm font-medium text-[var(--text)]">
                  {t('auth.password')}
                </label>
                <input
                  id="login-password"
                  type="password"
                  value={loginPassword}
                  onChange={(e) => setLoginPassword(e.target.value)}
                  placeholder={t('auth.enterPassword')}
                  className="w-full px-3 py-2 bg-[var(--background)] border border-[var(--border)] rounded-md text-[var(--text)] text-sm placeholder:text-[var(--text-muted)] focus:outline-none focus:border-[var(--accent)] transition-colors"
                  disabled={isLoading}
                  required
                  autoComplete="current-password"
                />
              </div>

              <button
                type="submit"
                className="w-full py-2 px-4 bg-[var(--accent)] hover:bg-[var(--accent)]/90 text-white font-medium rounded-md transition-colors text-sm disabled:opacity-60 disabled:cursor-not-allowed"
                disabled={isLoading}
              >
                {isLoading ? t('auth.loggingIn') : t('auth.login')}
              </button>
            </form>
          ) : (
            <form onSubmit={handleRegister} className="flex flex-col gap-4">
              <div className="flex flex-col gap-2">
                <label htmlFor="reg-username" className="text-sm font-medium text-[var(--text)]">
                  {t('auth.username')}
                </label>
                <input
                  id="reg-username"
                  type="text"
                  value={regUsername}
                  onChange={(e) => setRegUsername(e.target.value)}
                  placeholder={t('auth.yourName')}
                  className="w-full px-3 py-2 bg-[var(--background)] border border-[var(--border)] rounded-md text-[var(--text)] text-sm placeholder:text-[var(--text-muted)] focus:outline-none focus:border-[var(--accent)] transition-colors"
                  disabled={isLoading}
                  required
                  autoComplete="username"
                />
              </div>

              <div className="flex flex-col gap-2">
                <label htmlFor="reg-email" className="text-sm font-medium text-[var(--text)]">
                  Email
                </label>
                <input
                  id="reg-email"
                  type="email"
                  value={regEmail}
                  onChange={(e) => setRegEmail(e.target.value)}
                  placeholder="your@email.com"
                  className="w-full px-3 py-2 bg-[var(--background)] border border-[var(--border)] rounded-md text-[var(--text)] text-sm placeholder:text-[var(--text-muted)] focus:outline-none focus:border-[var(--accent)] transition-colors"
                  disabled={isLoading}
                  required
                  autoComplete="email"
                />
              </div>

              <div className="flex flex-col gap-2">
                <label htmlFor="reg-password" className="text-sm font-medium text-[var(--text)]">
                  {t('auth.password')}
                </label>
                <input
                  id="reg-password"
                  type="password"
                  value={regPassword}
                  onChange={(e) => setRegPassword(e.target.value)}
                  placeholder={t('auth.min6Chars')}
                  className="w-full px-3 py-2 bg-[var(--background)] border border-[var(--border)] rounded-md text-[var(--text)] text-sm placeholder:text-[var(--text-muted)] focus:outline-none focus:border-[var(--accent)] transition-colors"
                  disabled={isLoading}
                  required
                  autoComplete="new-password"
                />
              </div>

              <div className="flex flex-col gap-2">
                <label htmlFor="reg-confirm-password" className="text-sm font-medium text-[var(--text)]">
                  {t('auth.confirmPassword')}
                </label>
                <input
                  id="reg-confirm-password"
                  type="password"
                  value={regConfirmPassword}
                  onChange={(e) => setRegConfirmPassword(e.target.value)}
                  placeholder={t('auth.repeatPassword')}
                  className="w-full px-3 py-2 bg-[var(--background)] border border-[var(--border)] rounded-md text-[var(--text)] text-sm placeholder:text-[var(--text-muted)] focus:outline-none focus:border-[var(--accent)] transition-colors"
                  disabled={isLoading}
                  required
                  autoComplete="new-password"
                />
              </div>

              <button
                type="submit"
                className="w-full py-2 px-4 bg-[var(--accent)] hover:bg-[var(--accent)]/90 text-white font-medium rounded-md transition-colors text-sm disabled:opacity-60 disabled:cursor-not-allowed"
                disabled={isLoading}
              >
                {isLoading ? t('auth.registering') : t('auth.register')}
              </button>
            </form>
          )}
        </div>

        {/* Footer */}
        <div className="px-6 py-4 border-t border-[var(--border)] text-center">
          <p className="text-sm text-[var(--text-secondary)]">
            {view === 'login' ? (
              <>
                {t('auth.noAccount')}{' '}
                <button
                  onClick={switchToRegister}
                  className="text-[var(--accent)] hover:text-[var(--accent)]/90 font-medium transition-colors"
                  disabled={isLoading}
                >
                  {t('auth.register')}
                </button>
              </>
            ) : (
              <>
                {t('auth.hasAccount')}{' '}
                <button
                  onClick={switchToLogin}
                  className="text-[var(--accent)] hover:text-[var(--accent)]/90 font-medium transition-colors"
                  disabled={isLoading}
                >
                  {t('auth.login')}
                </button>
              </>
            )}
          </p>
        </div>
      </div>
    </div>
  );
}
