'use client';

import { useState } from 'react';
import { Lock, LockOpen, Plus, X } from 'lucide-react';

type CreateEnvironmentModalProps = {
  onClose: () => void;
  onCreate: (name: string, visibility: 'private' | 'public') => Promise<void>;
  error: string;
};

export function CreateEnvironmentModal({ onClose, onCreate, error }: CreateEnvironmentModalProps) {
  const [name, setName] = useState('');
  const [visibility, setVisibility] = useState<'private' | 'public'>('private');
  const [submitting, setSubmitting] = useState(false);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!name.trim() || submitting) return;
    setSubmitting(true);
    await onCreate(name.trim(), visibility);
    setSubmitting(false);
  }

  return (
    <div className="overlay" onClick={onClose} role="presentation">
      <div className="create-env-modal" onClick={(e) => e.stopPropagation()} role="dialog" aria-label="Create environment">
        <div className="create-env-header">
          <h2>New environment</h2>
          <button className="icon-button" onClick={onClose} aria-label="Close">
            <X size={15} />
          </button>
        </div>
        <form onSubmit={handleSubmit}>
          <label className="create-env-field">
            <span>Name</span>
            <input
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="my-environment"
              autoFocus
              required
            />
          </label>
          <label className="create-env-field">
            <span>Visibility</span>
            <div className="create-env-toggle">
              <button
                type="button"
                className={visibility === 'private' ? 'active' : ''}
                onClick={() => setVisibility('private')}
              >
                <Lock size={15} />
                Private
              </button>
              <button
                type="button"
                className={visibility === 'public' ? 'active' : ''}
                onClick={() => setVisibility('public')}
              >
                <LockOpen size={15} />
                Public
              </button>
            </div>
          </label>
          {error && <p className="create-env-error">{error}</p>}
          <button className="primary-button" type="submit" disabled={!name.trim() || submitting}>
            <Plus size={15} />
            Create
          </button>
        </form>
      </div>
    </div>
  );
}
