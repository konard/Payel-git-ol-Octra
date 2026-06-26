import { useState, useEffect } from 'react';
import { Mail, Crown, Calendar, Edit2, Check, X, Copy } from 'lucide-react';
import { useAuthStore } from '../../stores/authStore';
import { t } from '../../hooks/useI18n';
import { createAvatarDataUrl } from '../../utils/avatar';

interface UserProfileProps {
  onClose: () => void;
}

export function UserProfile({ onClose }: UserProfileProps) {
  const { user, hasSubscription, subscriptionEnd, logout } = useAuthStore();
  const [isEditing, setIsEditing] = useState(false);
  const [editUsername, setEditUsername] = useState(user?.username || '');
  const [copied, setCopied] = useState(false);

  const handleSaveUsername = async () => {
    // TODO: Call backend API to update username
    // await updateUsername(editUsername);
    setIsEditing(false);
  };

  const formatDate = (timestamp?: number) => {
    if (!timestamp) return null;
    return new Date(timestamp).toLocaleDateString('ru-RU', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
    });
  };

  const isSubscriptionActive = hasSubscription;

  const [avatarUrl, setAvatarUrl] = useState<string>('');

  useEffect(() => {
    // Bump the cache key version whenever the generator changes so users get
    // the new isometric sculpture instead of a stale flat identicon.
    const seed = user?.id || user?.email || 'default';
    const key = `avatar-iso-v1-${seed}`;
    const cached = localStorage.getItem(key);
    if (cached) {
      setAvatarUrl(cached);
      return;
    }

    const dataUrl = createAvatarDataUrl(seed);
    if (dataUrl) {
      localStorage.setItem(key, dataUrl);
      setAvatarUrl(dataUrl);
    }
  }, [user?.id, user?.email]);

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50" onClick={onClose}>
      <div
        className="bg-[var(--surface)] border border-[var(--border)] rounded-xl shadow-2xl w-[500px] max-w-[95vw] overflow-hidden"
        onClick={(e) => e.stopPropagation()}
      >
        {/* Header */}
        <div className="relative bg-[var(--surface)] border-b border-[var(--border)] px-6 py-8">
          <button
            onClick={onClose}
            className="absolute top-4 right-4 p-1.5 hover:bg-[var(--background)] rounded-md transition-colors text-[var(--text)]"
          >
            <X size={18} />
          </button>

          {/* Avatar */}
          <div className="flex items-center gap-4">
            <img
              src={avatarUrl}
              alt="Avatar"
              className="w-20 h-20 rounded-full shadow-lg border-2 border-[var(--border)]"
            />
            <div className="flex-1">
              <div className="flex items-center gap-2">
                {isEditing ? (
                  <div className="flex items-center gap-2 flex-1">
                    <input
                      type="text"
                      value={editUsername}
                      onChange={(e) => setEditUsername(e.target.value)}
                      className="flex-1 px-2 py-1 bg-[var(--background)] border border-[var(--border)] rounded-md text-[var(--text)] placeholder:text-[var(--text-muted)] text-sm focus:outline-none focus:border-[var(--accent)]"
                      placeholder="Username"
                      autoFocus
                    />
                    <button
                      onClick={handleSaveUsername}
                      className="p-1 hover:bg-[var(--background)] rounded transition-colors text-[var(--text)]"
                    >
                      <Check size={16} />
                    </button>
                    <button
                      onClick={() => {
                        setIsEditing(false);
                        setEditUsername(user?.username || '');
                      }}
                      className="p-1 hover:bg-[var(--background)] rounded transition-colors text-[var(--text)]"
                    >
                      <X size={16} />
                    </button>
                  </div>
                ) : (
                  <>
                    <h2 className="text-xl font-bold text-[var(--text)]">{user?.username || 'User'}</h2>
                    <button
                      onClick={() => setIsEditing(true)}
                      className="p-1 hover:bg-[var(--background)] rounded transition-colors text-[var(--text)]"
                    >
                      <Edit2 size={14} />
                    </button>
                  </>
                )}
              </div>
              <p className="text-sm text-[var(--text-muted)] mt-1 flex items-center gap-1.5">
                <Mail size={12} />
                {user?.email || 'No email'}
              </p>
            </div>
          </div>
        </div>

        {/* Content */}
        <div className="p-6 space-y-4">
          {/* Subscription Status */}
          <div class={`p-4 rounded-lg border ${
            isSubscriptionActive
              ? 'bg-gradient-to-r from-orange-500/10 to-orange-600/10 border-orange-500/30'
              : 'bg-[var(--background)] border-[var(--border)]'
          }`}>
            <div className="flex items-center gap-2 mb-2">
              <Crown 
                size={18} 
                className={isSubscriptionActive ? 'text-orange-500' : 'text-gray-400'}
              />
              <span className="font-semibold text-[var(--text)]">
                {isSubscriptionActive ? 'Pro Plan' : 'Free Plan'}
              </span>
            </div>
            {isSubscriptionActive && subscriptionEnd && (
              <p className="text-sm text-[var(--text-muted)] flex items-center gap-1.5">
                <Calendar size={12} />
                {t('profile.subscriptionEnd')}: {formatDate(subscriptionEnd)}
              </p>
            )}
            {!isSubscriptionActive && (
              <p className="text-sm text-[var(--text-muted)]">
                {t('profile.noSubscription')}
              </p>
            )}
          </div>

          {/* User Info */}
          <div className="space-y-3">
            <div className="flex items-center justify-between py-2 border-b border-[var(--border)]">
              <span className="text-sm text-[var(--text-muted)]">{t('profile.userId')}</span>
              <div className="flex items-center gap-1.5">
                <span className="text-sm text-[var(--text)] font-mono">
                  {user?.id ? `${user.id.substring(0, 8)}...` : 'N/A'}
                </span>
                {user?.id && (
                  <button
                    onClick={() => {
                      navigator.clipboard.writeText(user.id);
                      setCopied(true);
                      setTimeout(() => setCopied(false), 1500);
                    }}
                    className={`p-1 rounded transition-all ${copied ? 'text-[var(--accent)]' : 'text-[var(--text-muted)] hover:text-[var(--text)] hover:bg-[var(--background)]'}`}
                    title={copied ? 'Copied!' : 'Copy ID'}
                  >
                    {copied ? <Check size={14} /> : <Copy size={14} />}
                  </button>
                )}
              </div>
            </div>
            <div className="flex items-center justify-between py-2">
              <span className="text-sm text-[var(--text-muted)]">{t('profile.memberSince')}</span>
              <span className="text-sm text-[var(--text)]">
                {user?.created_at ? formatDate(new Date(user.created_at).getTime()) : 'N/A'}
              </span>
            </div>
          </div>

          {/* Actions */}
          <div className="flex gap-2 pt-2">
            <button
              onClick={() => {
                logout();
                onClose();
              }}
              className="flex-1 px-4 py-2.5 bg-red-500 hover:bg-red-600 text-white rounded-lg text-sm font-medium transition-colors"
            >
              {t('profile.logout')}
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
