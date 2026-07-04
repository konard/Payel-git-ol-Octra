'use client';

import { FormEvent, useEffect, useState } from 'react';
import {
  ArrowRight,
  Chrome,
  Github,
  KeyRound,
  LockKeyhole,
  Mail,
  ShieldCheck,
  Sparkles,
  UserPlus,
} from 'lucide-react';
import { login, register } from '../server/auth';
import { ROUTES } from '../config/routes';
import { ASSETS } from '../config/images';

const providers = [
  {
    name: 'Google',
    href: ROUTES.LOGIN_GOOGLE,
    icon: Chrome,
  },
  {
    name: 'GitHub',
    href: ROUTES.LOGIN_GITHUB,
    icon: Github,
  },
  {
    name: 'Lefine',
    href: ROUTES.LOGIN_LEFINE,
    icon: Sparkles,
  },
];

export default function AuthPage() {
  const [loginError, setLoginError] = useState('');
  const [registerError, setRegisterError] = useState('');

  useEffect(() => {
    const hasToken = ['octra_access_token', 'access_token'].some((key) => window.localStorage.getItem(key));
    if (hasToken) {
      window.location.href = ROUTES.WORKSPACE;
    }
  }, []);

  async function handleLogin(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setLoginError('');
    const form = new FormData(e.currentTarget);
    try {
      const res = await login(form.get('email') as string, form.get('password') as string);
      if (!res.ok) {
        const text = await res.text();
        let msg = text || 'Sign in failed';
        try { const j = JSON.parse(text); msg = j.error || msg; } catch {}
        const friendly: Record<string, string> = {
          'invalid email or password': 'Wrong email or password. Check your credentials or create a new account.',
          'email already registered': 'This email is already registered. Try signing in instead.',
          'username already taken': 'This username is taken. Try a different one.',
          'username, email and password are required': 'Fill in all fields — username, email, and password.',
          'invalid json body': 'Something went wrong. Please try again.',
          'failed to create account': 'Could not create account. Try again later.',
        };
        setLoginError(friendly[msg] || msg);
        return;
      }
      const body = await res.json();
      const accessToken = body?.data?.access_token;
      const refreshToken = body?.data?.refresh_token;
      const uname = body?.data?.user?.username;
      const userId = body?.data?.user?.id ?? body?.data?.user?.user_id;
      if (accessToken) {
        window.localStorage.setItem('octra_access_token', accessToken);
        window.localStorage.setItem('access_token', accessToken);
        if (refreshToken) {
          window.localStorage.setItem('octra_refresh_token', refreshToken);
          window.localStorage.setItem('refresh_token', refreshToken);
        }
      }
      if (uname) {
        window.localStorage.setItem('octra_username', uname);
      }
      if (userId) {
        window.localStorage.setItem('octra_user_id', userId);
      }
      window.location.href = ROUTES.WORKSPACE;
    } catch {
      setLoginError('Network error');
    }
  }

  async function handleRegister(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setRegisterError('');
    const form = new FormData(e.currentTarget);
    try {
      const res = await register(form.get('username') as string, form.get('email') as string, form.get('password') as string);
      if (!res.ok) {
        const text = await res.text();
        let msg = text || 'Registration failed';
        try { const j = JSON.parse(text); msg = j.error || msg; } catch {}
        const friendly: Record<string, string> = {
          'invalid email or password': 'Wrong email or password. Check your credentials.',
          'email already registered': 'This email is already registered. Try signing in instead.',
          'username already taken': 'This username is taken. Try a different one.',
          'username, email and password are required': 'Fill in all fields — username, email, and password.',
          'invalid json body': 'Something went wrong. Please try again.',
          'failed to create account': 'Could not create account. Try again later.',
        };
        setRegisterError(friendly[msg] || msg);
        return;
      }
      const body = await res.json();
      const apiKey = body?.api_key;
      const userId = body?.user_id;
      const uname = form.get('username') as string;
      if (apiKey) {
        window.localStorage.setItem('octra_access_token', apiKey);
        window.localStorage.setItem('access_token', apiKey);
        window.localStorage.setItem('octra_username', uname);
        if (userId) window.localStorage.setItem('octra_user_id', userId);
        window.localStorage.setItem('octra_show_welcome', '1');
      }
      window.location.href = ROUTES.WORKSPACE;
    } catch {
      setRegisterError('Network error');
    }
  }

  return (
    <main className="auth-shell">
      <header className="top-nav auth-nav">
        <a className="brand-link" href={ROUTES.HOME} aria-label="Octra home">
          <img src={ASSETS.LOGO} alt="" className="brand-mark" />
          <span>Octra</span>
        </a>
        <div className="nav-actions">
          <a className="text-button" href={ROUTES.DASHBOARD}>
            <ArrowRight size={17} />
            <span>Dashboard</span>
          </a>
        </div>
      </header>

      <section className="auth-grid" aria-labelledby="auth-title">
        <div className="auth-visual">
          <p className="eyebrow">Access layer</p>
          <h1 id="auth-title">Sign in or create account</h1>
          <p>
            Octra keeps identity, provider access, and delivery limits visible before work enters the pipeline.
          </p>
          <div className="auth-status-board" aria-label="Access status">
            <div className="status-line">
              <ShieldCheck size={18} />
              <span>OAuth providers</span>
              <strong>3 ready</strong>
            </div>
            <div className="status-line">
              <LockKeyhole size={18} />
              <span>Session policy</span>
              <strong>Strict</strong>
            </div>
            <div className="status-line">
              <KeyRound size={18} />
              <span>Token scope</span>
              <strong>Workspace</strong>
            </div>
          </div>
        </div>

        <div className="auth-panel" aria-label="Authentication options">
          <div className="provider-stack">
            {providers.map((provider) => (
              <a className="provider-button" href={provider.href} key={provider.name}>
                <provider.icon size={19} />
                <span>Continue with {provider.name}</span>
                <ArrowRight size={16} />
              </a>
            ))}
          </div>

          <div className="auth-divider">
            <span>or use email</span>
          </div>

          <div className="auth-forms">
            <form className="auth-form" onSubmit={handleLogin}>
              <div className="form-heading">
                <KeyRound size={18} />
                <h2>Sign in</h2>
              </div>
              {loginError && <p className="form-error">{loginError}</p>}
              <label>
                <span>Email</span>
                <span className="input-shell">
                  <Mail size={16} />
                  <input type="email" name="email" autoComplete="email" placeholder="you@company.com" />
                </span>
              </label>
              <label>
                <span>Password</span>
                <span className="input-shell">
                  <LockKeyhole size={16} />
                  <input type="password" name="password" autoComplete="current-password" placeholder="Password" />
                </span>
              </label>
              <button className="primary-button full-button" type="submit">
                <KeyRound size={18} />
                <span>Sign in</span>
              </button>
            </form>

            <form className="auth-form muted-form" onSubmit={handleRegister}>
              <div className="form-heading">
                <UserPlus size={18} />
                <h2>Create account</h2>
              </div>
              {registerError && <p className="form-error">{registerError}</p>}
              <label>
                <span>Username</span>
                <span className="input-shell">
                  <UserPlus size={16} />
                  <input type="text" name="username" autoComplete="username" placeholder="octra-user" />
                </span>
              </label>
              <label>
                <span>Email</span>
                <span className="input-shell">
                  <Mail size={16} />
                  <input type="email" name="email" autoComplete="email" placeholder="you@company.com" />
                </span>
              </label>
              <label>
                <span>Password</span>
                <span className="input-shell">
                  <LockKeyhole size={16} />
                  <input type="password" name="password" autoComplete="new-password" placeholder="Password" />
                </span>
              </label>
              <button className="secondary-button full-button" type="submit">
                <UserPlus size={18} />
                <span>Create account</span>
              </button>
            </form>
          </div>
        </div>
      </section>
    </main>
  );
}
