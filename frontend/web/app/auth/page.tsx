'use client';

import { FormEvent, useState } from 'react';
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

const providers = [
  {
    name: 'Google',
    href: '/auth/google',
    icon: Chrome,
  },
  {
    name: 'GitHub',
    href: '/auth/github',
    icon: Github,
  },
  {
    name: 'Lefine',
    href: '/auth/lefine',
    icon: Sparkles,
  },
];

export default function AuthPage() {
  const [loginError, setLoginError] = useState('');
  const [registerError, setRegisterError] = useState('');

  async function handleLogin(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setLoginError('');
    const form = new FormData(e.currentTarget);
    try {
      const res = await login(form.get('email') as string, form.get('password') as string);
      if (!res.ok) {
        const text = await res.text();
        setLoginError(text || 'Sign in failed');
        return;
      }
      window.location.href = '/app';
    } catch {
      setLoginError('Network error');
    }
  }

  async function handleRegister(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setRegisterError('');
    const form = new FormData(e.currentTarget);
    try {
      const res = await register(form.get('username') as string, form.get('email') as string);
      if (!res.ok) {
        const text = await res.text();
        setRegisterError(text || 'Registration failed');
        return;
      }
      window.location.href = '/app';
    } catch {
      setRegisterError('Network error');
    }
  }
  return (
    <main className="auth-shell">
      <header className="top-nav auth-nav">
        <a className="brand-link" href="/" aria-label="Octra home">
          <img src="/assets/octra-node-logo.svg" alt="" className="brand-mark" />
          <span>Octra</span>
        </a>
        <div className="nav-actions">
          <a className="text-button" href="/dashboard">
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
