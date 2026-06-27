'use client';

import { FormEvent, useState } from 'react';
import { X, KeyRound } from 'lucide-react';
import { Select } from './Select';

interface Props {
  onClose: () => void;
  onCreate: (name: string, expiresAt: string | null) => void;
  error: string;
}

const EXPIRY_OPTIONS = [
  { label: '1 day', value: '1d' },
  { label: '7 days', value: '7d' },
  { label: '1 month', value: '30d' },
  { label: '6 months', value: '180d' },
  { label: '1 year', value: '365d' },
  { label: 'Never expires', value: '' },
];

function computeExpiresAt(value: string): string | null {
  if (!value) return null;
  const match = value.match(/^(\d+)d$/);
  if (!match) return null;
  const days = parseInt(match[1], 10);
  const d = new Date();
  d.setDate(d.getDate() + days);
  return d.toISOString();
}

export function CreateKeyModal({ onClose, onCreate, error }: Props) {
  const [name, setName] = useState('');
  const [expiry, setExpiry] = useState('');

  function handleSubmit(e: FormEvent) {
    e.preventDefault();
    if (!name.trim()) return;
    onCreate(name.trim(), computeExpiresAt(expiry));
  }

  return (
    <div className="welcome-overlay" onClick={onClose}>
      <div className="welcome-dialog" style={{ maxWidth: 400, padding: '32px 28px' }} onClick={(e) => e.stopPropagation()}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', width: '100%' }}>
          <h2 style={{ margin: 0, fontSize: '1rem', display: 'flex', alignItems: 'center', gap: 8 }}>
            <KeyRound size={18} />
            Create API key
          </h2>
          <button className="icon-button" onClick={onClose} aria-label="Close" style={{ width: 32, height: 32 }}>
            <X size={18} />
          </button>
        </div>

        <form onSubmit={handleSubmit} style={{ width: '100%', display: 'grid', gap: 16 }}>
          <label>
            <span style={{ display: 'block', marginBottom: 6, fontSize: '0.84rem', color: 'var(--muted)' }}>Key name</span>
            <span className="input-shell" style={{ width: '100%' }}>
              <input
                type="text"
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="e.g. Production CLI"
                autoFocus
                style={{ width: '100%' }}
              />
            </span>
          </label>

          <label>
            <span style={{ display: 'block', marginBottom: 6, fontSize: '0.84rem', color: 'var(--muted)' }}>Expiration</span>
            <Select options={EXPIRY_OPTIONS} value={expiry} onChange={setExpiry} />
          </label>

          {error && <p style={{ margin: 0, color: 'var(--metric-danger)', fontSize: '0.84rem' }}>{error}</p>}

          <button className="primary-button full-button" type="submit" disabled={!name.trim()}>
            <KeyRound size={18} />
            <span>Create key</span>
          </button>
        </form>
      </div>
    </div>
  );
}
