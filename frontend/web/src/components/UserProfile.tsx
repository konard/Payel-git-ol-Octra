import { useState, useEffect } from 'react';
import { User, Mail, Crown, Calendar, Edit2, Check, X } from 'lucide-react';
import { useAuthStore } from '../stores/authStore';
import { t } from '../hooks/useI18n';

interface UserProfileProps {
  onClose: () => void;
}

export function UserProfile({ onClose }: UserProfileProps) {
  const { user, hasSubscription, subscriptionEnd, logout } = useAuthStore();
  const [isEditing, setIsEditing] = useState(false);
  const [editUsername, setEditUsername] = useState(user?.username || '');

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

  const isSubscriptionActive = true; // TODO: user?.subscription_end && user.subscription_end > Date.now() / 1000;

  const [avatarUrl, setAvatarUrl] = useState<string>('');

  useEffect(() => {
    const generateAvatar = async () => {
      const key = `avatar-canvas-${user?.id || 'default'}`;
      const cached = localStorage.getItem(key);
      if (cached) {
        setAvatarUrl(cached);
        return;
      }

      const hash = user?.id || user?.email || 'default';
      const fullHash = await sha256(hash);
      const matrix = generateMatrix(fullHash);
      const dataUrl = drawAvatar(matrix);
      localStorage.setItem(key, dataUrl);
      setAvatarUrl(dataUrl);
    };

    generateAvatar();
  }, [user?.id, user?.email]);

  const sha256 = async (message: string): Promise<string> => {
    const msgBuffer = new TextEncoder().encode(message);
    const hashBuffer = await crypto.subtle.digest('SHA-256', msgBuffer);
    const hashArray = Array.from(new Uint8Array(hashBuffer));
    return hashArray.map(b => ('00' + b.toString(16)).slice(-2)).join('');
  };

  const generateMatrix = (hash: string): number[][] => {
    const m = Array(5).fill(0).map(() => Array(5).fill(0));
    for (let i = 0; i < 5; i++) {
      for (let j = 0; j < 5; j++) {
        const n = parseInt(hash.substr(i * 5 + j, 1), 16);
        m[i][j] = n > 7 ? 0 : 1;
      }
    }
    // make symmetric
    for (let i = 0; i < 5; i++) {
      for (let j = Math.floor(5 / 2), k = 2; j < 5; j++, k += 2) {
        m[i][j] = m[i][j - k];
      }
    }
    return m;
  };

  const drawAvatar = (m: number[][]): string => {
    const canvas = document.createElement('canvas');
    canvas.width = 100;
    canvas.height = 100;
    const ctx = canvas.getContext('2d')!;
    ctx.fillStyle = '#F8F8F8';
    ctx.fillRect(0, 0, 100, 100);

    const r = Math.floor(Math.random() * 128 + 128);
    const g = Math.floor(Math.random() * 128 + 128);
    const b = Math.floor(Math.random() * 128 + 128);
    ctx.fillStyle = `rgba(${r}, ${g}, ${b}, 1)`;

    for (let i = 0; i < 5; i++) {
      for (let j = 0; j < 5; j++) {
        if (m[i][j] === 1) {
          ctx.fillRect(j * 20, i * 20, 20, 20);
        }
      }
    }
    return canvas.toDataURL();
  };

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
              <span className="text-sm text-[var(--text)] font-mono">
                {user?.id ? `${user.id.substring(0, 8)}...` : 'N/A'}
              </span>
            </div>
            <div className="flex items-center justify-between py-2">
              <span className="text-sm text-[var(--text-muted)]">{t('profile.memberSince')}</span>
              <span className="text-sm text-[var(--text)]">
                {user?.createdAt ? formatDate(new Date(user.createdAt).getTime() / 1000) : 'N/A'}
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
