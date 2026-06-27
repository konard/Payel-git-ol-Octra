'use client';

import { useState, useEffect, useCallback } from 'react';
import { Copy, Eye, EyeOff, Check, Plus, Trash2, KeyRound } from 'lucide-react';
import { ApiTabs } from '../ApiTabs';
import { listAPIKeys, createAPIKey, deleteAPIKey } from '../../../server/keys';
import { CreateKeyModal } from '../../../components/CreateKeyModal';

type APIKeyItem = {
  id: string;
  name: string;
  key?: string;
  expires_at: string | null;
  created_at: string;
};

export default function ApiKeysPage() {
  const [keys, setKeys] = useState<APIKeyItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [showCreate, setShowCreate] = useState(false);
  const [createError, setCreateError] = useState('');
  const [showKeyMap, setShowKeyMap] = useState<Record<string, boolean>>({});
  const [copiedId, setCopiedId] = useState('');

  const loadKeys = useCallback(() => {
    setLoading(true);
    listAPIKeys().then(async (res) => {
      if (res.ok) setKeys(await res.json());
    }).catch(() => {}).finally(() => setLoading(false));
  }, []);

  useEffect(() => { loadKeys(); }, [loadKeys]);

  async function copyKey(key: string, id: string) {
    try {
      await navigator.clipboard.writeText(key);
      setCopiedId(id);
      setTimeout(() => setCopiedId(''), 2000);
    } catch {}
  }

  async function handleCreate(name: string, expiresAt: string | null) {
    setCreateError('');
    const res = await createAPIKey(name, expiresAt);
    if (!res.ok) {
      const text = await res.text();
      setCreateError(text || 'Failed to create key');
      return;
    }
    const newKey = await res.json();
    setKeys((prev) => [newKey, ...prev]);
    setShowCreate(false);
  }

  async function handleDelete(id: string) {
    if (!confirm('Delete this API key? Any services using it will stop working.')) return;
    await deleteAPIKey(id);
    setKeys((prev) => prev.filter((k) => k.id !== id));
  }

  function formatExpiry(expiresAt: string | null): string {
    if (!expiresAt) return 'Never';
    const d = new Date(expiresAt);
    const now = new Date();
    if (d < now) return 'Expired';
    return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
  }

  function maskKey(key: string): string {
    if (key.length <= 10) return key;
    return key.slice(0, 8) + '••••••••' + key.slice(-4);
  }

  return (
    <div style={{ padding: 14 }}>
      <ApiTabs />
      <div className="dashboard-grid dashboard-grid-single">
        <article className="large-panel" aria-label="API keys" style={{ padding: 24, minHeight: 0 }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 20 }}>
            <h2 style={{ margin: 0, fontSize: '1rem', display: 'flex', alignItems: 'center', gap: 8 }}>
              <KeyRound size={18} />
              API keys
            </h2>
            <button className="small-command" onClick={() => { setCreateError(''); setShowCreate(true); }} style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
              <Plus size={14} />
              Create key
            </button>
          </div>

          {loading ? (
            <p style={{ color: 'var(--muted)' }}>Loading keys…</p>
          ) : keys.length === 0 ? (
            <div style={{ padding: '32px 16px', textAlign: 'center', border: '1px dashed var(--line)', borderRadius: 8 }}>
              <KeyRound size={28} style={{ color: 'var(--quiet)', marginBottom: 8 }} />
              <p style={{ margin: 0, color: 'var(--muted)', fontSize: '0.9rem' }}>No API keys yet.</p>
              <p style={{ margin: '4px 0 0', color: 'var(--quiet)', fontSize: '0.82rem' }}>Create one to use with your CLI or MCP client.</p>
            </div>
          ) : (
            <div style={{ display: 'grid', gap: 10 }}>
              {keys.map((item) => (
                <div key={item.id} className="api-key-row">
                  <div className="api-key-row-info">
                    <strong>{item.name}</strong>
                    <span style={{ color: 'var(--quiet)', fontSize: '0.78rem' }}>Expires: {formatExpiry(item.expires_at)}</span>
                  </div>
                  <div className="api-key-row-value">
                    {item.key && (
                      <>
                        <code>{showKeyMap[item.id] ? item.key : maskKey(item.key)}</code>
                        <button className="icon-button" onClick={() => setShowKeyMap((prev) => ({ ...prev, [item.id]: !prev[item.id] }))} aria-label={showKeyMap[item.id] ? 'Hide' : 'Show'}>
                          {showKeyMap[item.id] ? <EyeOff size={15} /> : <Eye size={15} />}
                        </button>
                        <button className="icon-button" onClick={() => copyKey(item.key!, item.id)} aria-label="Copy">
                          {copiedId === item.id ? <Check size={15} style={{ color: 'var(--metric-success)' }} /> : <Copy size={15} />}
                        </button>
                      </>
                    )}
                  </div>
                  <button className="icon-button icon-button-danger" onClick={() => handleDelete(item.id)} aria-label="Delete" style={{ color: 'var(--muted)', flex: '0 0 auto' }}>
                    <Trash2 size={15} />
                  </button>
                </div>
              ))}
            </div>
          )}

          <p style={{ margin: '20px 0 0', fontSize: '0.82rem', color: 'var(--muted)', lineHeight: 1.55 }}>
            Use these keys to authenticate API requests from your CLI or MCP client.
            Treat them like passwords — do not share them publicly.
            Deleting a key will invalidate it immediately.
          </p>
        </article>
      </div>

      {showCreate && (
        <CreateKeyModal
          onClose={() => setShowCreate(false)}
          onCreate={handleCreate}
          error={createError}
        />
      )}
    </div>
  );
}
