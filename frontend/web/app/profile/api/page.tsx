'use client';

import { useState, useEffect } from 'react';
import { Copy, Eye, EyeOff, Check, KeyRound, Globe, Settings, RefreshCw } from 'lucide-react';
import { fetchMe } from '../../server/user';

type UserData = { api_key: string };
type Tab = 'api-keys' | 'public-api' | 'settings';

function readToken(): string | null {
  return window.localStorage.getItem('octra_access_token') ?? window.localStorage.getItem('access_token');
}

export default function ApiPage() {
  const [tab, setTab] = useState<Tab>('api-keys');
  const [showKey, setShowKey] = useState(false);
  const [copied, setCopied] = useState(false);
  const [apiKey, setApiKey] = useState<string | null>(null);

  useEffect(() => {
    const token = readToken();
    if (!token) return;
    fetchMe(token).then(async (res) => {
      if (!res.ok) return;
      const body = await res.json();
      const user = body?.data ?? body;
      if (user?.api_key) setApiKey(user.api_key);
    }).catch(() => {});
  }, []);

  async function copyKey() {
    if (!apiKey) return;
    try {
      await navigator.clipboard.writeText(apiKey);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch {}
  }

  const tabs: { key: Tab; label: string; icon: typeof KeyRound }[] = [
    { key: 'api-keys', label: 'API keys', icon: KeyRound },
    { key: 'public-api', label: 'Public API', icon: Globe },
    { key: 'settings', label: 'Settings', icon: Settings },
  ];

  return (
    <div style={{ padding: 14 }}>
      <div className="dashboard-tabs" aria-label="API sections" style={{ marginBottom: 14 }}>
        {tabs.map((t) => {
          const Icon = t.icon;
          return (
            <button
              className={tab === t.key ? 'active' : ''}
              onClick={() => setTab(t.key)}
              key={t.key}
              style={{ cursor: 'pointer', background: 'none', border: 'none', font: 'inherit', display: 'inline-flex', alignItems: 'center', gap: 6 }}
            >
              <Icon size={15} />
              {t.label}
            </button>
          );
        })}
      </div>

      {tab === 'api-keys' && (
        <div className="dashboard-grid dashboard-grid-single">
          <article className="large-panel" aria-label="API keys" style={{ padding: 24, minHeight: 0 }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
              <h2 style={{ margin: 0, fontSize: '1rem' }}>Your API key</h2>
              <button className="small-command" style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
                <RefreshCw size={14} />
                Regenerate
              </button>
            </div>

            <div className="api-key-shell">
              <code className="api-key-value">
                {!apiKey ? 'Loading…' : showKey ? apiKey : apiKey.slice(0, 8) + '••••••••' + apiKey.slice(-4)}
              </code>
              <div className="api-key-actions">
                <button className="icon-button" onClick={() => setShowKey(!showKey)} aria-label={showKey ? 'Hide' : 'Show'}>
                  {showKey ? <EyeOff size={16} /> : <Eye size={16} />}
                </button>
                <button className="icon-button" onClick={copyKey} aria-label="Copy">
                  <Copy size={16} />
                </button>
              </div>
              {copied && <span className="copy-check"><Check size={16} /></span>}
            </div>

            <p style={{ margin: '16px 0 0', fontSize: '0.82rem', color: 'var(--muted)', lineHeight: 1.55 }}>
              Use this key to authenticate API requests from your CLI or MCP client.
              Treat it like a password — do not share it publicly.
              Regenerating the key will invalidate the previous one immediately.
            </p>
          </article>
        </div>
      )}

      {tab === 'public-api' && (
        <div className="dashboard-grid dashboard-grid-single">
          <article className="large-panel" aria-label="Public API" style={{ padding: 24, minHeight: 0 }}>
            <div style={{ display: 'flex', gap: 12, alignItems: 'center', marginBottom: 16 }}>
              <Globe size={20} />
              <h2 style={{ margin: 0, fontSize: '1rem' }}>Public API</h2>
            </div>
            <p style={{ color: 'var(--muted)', lineHeight: 1.6, margin: 0 }}>
              Public API endpoints and documentation will be available here.
            </p>
          </article>
        </div>
      )}

      {tab === 'settings' && (
        <div className="dashboard-grid dashboard-grid-single">
          <article className="large-panel" aria-label="API settings" style={{ padding: 24, minHeight: 0 }}>
            <div style={{ display: 'flex', gap: 12, alignItems: 'center', marginBottom: 16 }}>
              <Settings size={20} />
              <h2 style={{ margin: 0, fontSize: '1rem' }}>API settings</h2>
            </div>
            <p style={{ color: 'var(--muted)', lineHeight: 1.6, margin: 0 }}>
              Rate limits, IP whitelisting, and other API configuration options will be managed here.
            </p>
          </article>
        </div>
      )}
    </div>
  );
}
