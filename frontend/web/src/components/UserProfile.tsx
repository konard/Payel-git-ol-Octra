import { useState } from 'react';
import { User, Mail, Crown, Calendar, Edit2, X, Check } from 'lucide-react';
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

  const isSubscriptionActive = hasSubscription && subscriptionEnd && subscriptionEnd > Date.now() / 1000;

  // Generate avatar from username/email
  const getInitials = () => {
    const name = user?.username || user?.email || 'U';
    const parts = name.split('@')[0].split(/[.\s_]/);
    if (parts.length >= 2) {
      return (parts[0][0] + parts[1][0]).toUpperCase();
    }
    return name.substring(0, 2).toUpperCase();
  };

  const getAvatarColor = () => {
    const name = user?.username || user?.email || 'user';
    let hash = 0;
    for (let i = 0; i < name.length; i++) {
      hash = name.charCodeAt(i) + ((hash << 5) - hash);
    }
    const hue = hash % 360;
    return `hsl(${hue}, 60%, 50%)`;
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50" onClick={onClose}>
      <div
        className="bg-[var(--surface)] border border-[var(--border)] rounded-xl shadow-2xl w-[500px] max-w-[95vw] overflow-hidden"
        onClick={(e) => e.stopPropagation()}
      >
        {/* Header */}
        <div className="relative bg-gradient-to-r from-indigo-500 to-purple-600 px-6 py-8">
          <button
            onClick={onClose}
            className="absolute top-4 right-4 p-1.5 hover:bg-white/20 rounded-md transition-colors text-white"
          >
            <X size={18} />
          </button>
          
          {/* Avatar */}
          <div className="flex items-center gap-4">
            <div
              className="w-20 h-20 rounded-full flex items-center justify-center text-white text-2xl font-bold shadow-lg"
              style={{ backgroundColor: getAvatarColor() }}
            >
              {getInitials()}
            </div>
            <div className="flex-1">
              <div className="flex items-center gap-2">
                {isEditing ? (
                  <div className="flex items-center gap-2 flex-1">
                    <input
                      type="text"
                      value={editUsername}
                      onChange={(e) => setEditUsername(e.target.value)}
                      className="flex-1 px-2 py-1 bg-white/20 border border-white/30 rounded-md text-white placeholder:text-white/60 text-sm focus:outline-none focus:border-white"
                      placeholder="Username"
                      autoFocus
                    />
                    <button
                      onClick={handleSaveUsername}
                      className="p-1 hover:bg-white/20 rounded transition-colors text-white"
                    >
                      <Check size={16} />
                    </button>
                    <button
                      onClick={() => {
                        setIsEditing(false);
                        setEditUsername(user?.username || '');
                      }}
                      className="p-1 hover:bg-white/20 rounded transition-colors text-white"
                    >
                      <X size={16} />
                    </button>
                  </div>
                ) : (
                  <>
                    <h2 className="text-xl font-bold text-white">{user?.username || 'User'}</h2>
                    <button
                      onClick={() => setIsEditing(true)}
                      className="p-1 hover:bg-white/20 rounded transition-colors text-white"
                    >
                      <Edit2 size={14} />
                    </button>
                  </>
                )}
              </div>
              <p className="text-sm text-white/80 mt-1 flex items-center gap-1.5">
                <Mail size={12} />
                {user?.email || 'No email'}
              </p>
            </div>
          </div>
        </div>

        {/* Content */}
        <div className="p-6 space-y-4">
          {/* Subscription Status */}
          <div className={`p-4 rounded-lg border ${
            isSubscriptionActive 
              ? 'bg-gradient-to-r from-emerald-500/10 to-teal-500/10 border-emerald-500/30' 
              : 'bg-[var(--background)] border-[var(--border)]'
          }`}>
            <div className="flex items-center gap-2 mb-2">
              <Crown 
                size={18} 
                className={isSubscriptionActive ? 'text-emerald-500' : 'text-gray-400'}
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
