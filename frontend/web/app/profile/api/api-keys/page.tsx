'use client';

import { useState, useEffect } from 'react';
import { Copy, Eye, EyeOff, Check, RefreshCw } from 'lucide-react';
import { ApiTabs } from '../ApiTabs';
import { fetchMe } from '../../../server/user';

function readToken(): string | null {
  return window.localStorage.getItem('octra_access_token') ?? window.localStorage.getItem('access_token');
}

export default function ApiKeysPage() {
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

  return (
    <div style={{ padding: 14 }}>
      <ApiTabs />
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
    </div>
  );
}
