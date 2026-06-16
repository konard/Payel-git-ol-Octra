import { useState, useEffect, useRef } from 'react';
import { Plus, BookOpen, MoreHorizontal, Edit2, Trash2, X } from 'lucide-react';
import { useAuthStore } from '../../stores/authStore';
import { getChatHistory, createChat, deleteChat, type ChatHistoryItem } from '../../services/chatHistoryService';
import { t } from '../../hooks/useI18n';

interface SidebarProps {
  isOpen: boolean;
  onClose: () => void;
  onSelectChat: (chatId: string) => void;
  onNewChat: (chatId?: string) => void;
  // 'overlay' is the slide-in drawer used on narrow screens; 'dock' renders the
  // same session list inline as a persistent left pane in the desktop workspace.
  variant?: 'overlay' | 'dock';
}

export function Sidebar({ isOpen, onClose, onSelectChat, onNewChat, variant = 'overlay' }: SidebarProps) {
  const isDock = variant === 'dock';
  const { user, isAuthenticated, accessToken, refreshToken } = useAuthStore();
  const [chats, setChats] = useState<ChatHistoryItem[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [selectedChatId, setSelectedChatId] = useState<string | null>(null);
  const [openMenuId, setOpenMenuId] = useState<string | null>(null);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [editTitle, setEditTitle] = useState('');
  const menuRef = useRef<HTMLDivElement>(null);
  const hasAuthSession = isAuthenticated || Boolean(accessToken || refreshToken);
  const notLoggedInLabel = t('chatSidebar.notLoggedIn');
  const notLoggedInMessage = notLoggedInLabel === 'chatSidebar.notLoggedIn'
    ? 'Вы не вошли в аккаунт'
    : notLoggedInLabel;

  useEffect(() => {
    if (!isOpen && !isDock) return;

    if (!hasAuthSession) {
      setChats([]);
      setIsLoading(false);
      return;
    }

    if (user?.id) {
      loadChats(user.id);
      return;
    }

    setIsLoading(true);
  }, [hasAuthSession, user?.id, isOpen]);

  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (menuRef.current && !menuRef.current.contains(event.target as Node)) {
        setOpenMenuId(null);
      }
    }
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  const loadChats = async (userId: string) => {
    try {
      setIsLoading(true);
      const history = await getChatHistory(userId);
      setChats(history);
    } catch (error) {
      console.error('Failed to load chat history:', error);
    } finally {
      setIsLoading(false);
    }
  };

  const handleSelectChat = async (chatId: string) => {
    setSelectedChatId(chatId);
    onSelectChat(chatId);
    setOpenMenuId(null);
  };

  const handleNewChat = async () => {
    if (!user?.id) {
      setSelectedChatId(null);
      setOpenMenuId(null);
      onNewChat();
      return;
    }

    try {
      const newChat = await createChat(user.id, t('chatSidebar.newChat'));
      setChats(prev => [newChat, ...prev]);
      setSelectedChatId(newChat.id);
      onNewChat(newChat.id);
    } catch (error) {
      console.error('Failed to create chat:', error);
    }
  };

  // Ctrl+N (Cmd+N) starts a new chat from the docked history pane, matching the
  // shortcut advertised next to the header button.
  useEffect(() => {
    if (!isDock) return;
    const onKey = (e: KeyboardEvent) => {
      if ((e.ctrlKey || e.metaKey) && (e.key === 'n' || e.key === 'N')) {
        e.preventDefault();
        handleNewChat();
      }
    };
    window.addEventListener('keydown', onKey);
    return () => window.removeEventListener('keydown', onKey);
  });

  const handleDeleteChat = async (chatId: string, e: React.MouseEvent) => {
    e.stopPropagation();
    try {
      await deleteChat(chatId);
      setChats(prev => prev.filter(c => c.id !== chatId));
      if (selectedChatId === chatId) {
        setSelectedChatId(null);
        onNewChat();
      }
      setOpenMenuId(null);
    } catch (error) {
      console.error('Failed to delete chat:', error);
    }
  };

  const handleRename = (chatId: string, currentTitle: string, e: React.MouseEvent) => {
    e.stopPropagation();
    setEditingId(chatId);
    setEditTitle(currentTitle);
    setOpenMenuId(null);
  };

  const handleSaveRename = async (chatId: string) => {
    if (!editTitle.trim()) {
      setEditingId(null);
      return;
    }
    try {
      const { updateChatTitle } = await import('../../services/chatHistoryService');
      await updateChatTitle(chatId, editTitle.trim());
      setChats(prev => prev.map(c => c.id === chatId ? { ...c, title: editTitle.trim() } : c));
      setEditingId(null);
    } catch (error) {
      console.error('Failed to rename chat:', error);
      setEditingId(null);
    }
  };

  const groupChatsByMonth = (chats: ChatHistoryItem[]) => {
    const grouped: Record<string, ChatHistoryItem[]> = {};
    const now = new Date();
    
    chats.forEach(chat => {
      const chatDate = new Date(chat.created_at);
      const diffMonths = (now.getFullYear() - chatDate.getFullYear()) * 12 + now.getMonth() - chatDate.getMonth();
      
      let key: string;
      if (diffMonths === 0) key = t('chatSidebar.thisMonth');
      else if (diffMonths === 1) key = t('chatSidebar.lastMonth');
      else if (diffMonths < 12) key = chatDate.toLocaleString(t('language'), { month: 'long' });
      else key = chatDate.getFullYear().toString();
      
      if (!grouped[key]) grouped[key] = [];
      grouped[key].push(chat);
    });
    
    return grouped;
  };

  const groupedChats = groupChatsByMonth(chats);

  // A small status dot reflects how recently a session was touched: a live
  // accent dot for activity in the last day, a calmer muted dot otherwise.
  const isRecentlyActive = (date: Date) => {
    const diffHours = (new Date().getTime() - new Date(date).getTime()) / 3600000;
    return diffHours < 24;
  };

  const formatDate = (date: Date) => {
    const d = new Date(date);
    const now = new Date();
    const diffMs = now.getTime() - d.getTime();
    const diffMins = Math.floor(diffMs / 60000);
    const diffHours = Math.floor(diffMs / 3600000);
    const diffDays = Math.floor(diffMs / 86400000);
    
    if (diffMins < 1) return 'Just now';
    if (diffMins < 60) return `${diffMins}m ago`;
    if (diffHours < 24) return `${diffHours}h ago`;
    if (diffDays < 7) return `${diffDays}d ago`;
    return d.toLocaleDateString(t('language'), { month: 'short', day: 'numeric' });
  };

  return (
    <div
      className={
        isDock
          ? 'relative h-full w-full bg-[var(--surface-sunken)] flex flex-col'
          : `fixed inset-y-0 left-0 z-40 w-72 bg-[var(--surface)] border-r border-[var(--border)] flex flex-col transition-transform duration-200 ${isOpen ? 'translate-x-0' : '-translate-x-full'} ${isOpen ? '' : 'pointer-events-none'}`
      }
    >
      {/* Header */}
      <div className="flex items-center justify-between px-4 py-3.5 border-b border-[var(--border)]">
        <div className="flex items-center gap-3">
          <div className="w-9 h-9 rounded-xl bg-[var(--background)] border border-[var(--border)] flex items-center justify-center text-[var(--accent)]">
            <BookOpen size={18} />
          </div>
          <span className="font-semibold text-[15px] text-[var(--text)]">{t('chatSidebar.history')}</span>
        </div>
        {isDock ? (
          /* Compact "New" action with a Ctrl+N hint, matching the reference
             history header. The full-width button is reserved for the overlay
             drawer where vertical space is plentiful. */
          <button
            onClick={handleNewChat}
            className="flex items-center gap-1.5 rounded-lg border border-[var(--border)] bg-[var(--surface)] px-2.5 py-1.5 text-xs font-medium text-[var(--text)] transition-colors hover:border-[var(--accent)] hover:text-[var(--accent)]"
            title={`${t('chatSidebar.newChat')} (Ctrl+N)`}
            data-tour="new-chat"
          >
            <Plus size={14} />
            <span>{t('chatSidebar.newChat')}</span>
            <kbd className="ml-0.5 rounded border border-[var(--border)] bg-[var(--background)] px-1 py-px font-sans text-[10px] text-[var(--text-muted)]">
              Ctrl+N
            </kbd>
          </button>
        ) : (
          <button
            onClick={onClose}
            className="w-9 h-9 rounded-xl flex items-center justify-center text-[var(--text-secondary)] hover:bg-[var(--background)] hover:text-[var(--text)] transition-colors"
            title={t('common.close')}
          >
            <X size={18} />
          </button>
        )}
      </div>

      {/* Full-width New Chat button — overlay drawer only. */}
      {!isDock && (
        <div className="px-4 py-3.5">
          <button
            onClick={handleNewChat}
            className="w-full h-12 flex items-center justify-center gap-2.5 bg-[var(--accent)] hover:opacity-90 text-white font-semibold rounded-xl transition-all text-sm shadow-[0_10px_30px_rgba(255,132,0,0.18)] active:scale-[0.985]"
            data-tour="new-chat"
          >
            <Plus size={18} />
            {t('chatSidebar.newChat')}
          </button>
        </div>
      )}

      {/* Chat List */}
      <div className="flex-1 overflow-y-auto px-2.5 pb-4" ref={menuRef}>
        {!hasAuthSession ? (
          <div className="px-4 py-8 text-center text-sm text-[var(--text-muted)]">
            {notLoggedInMessage}
          </div>
        ) : isLoading ? (
          <div className="flex items-center justify-center py-8">
            <div className="animate-spin w-5 h-5 border-2 border-[var(--accent)] border-t-transparent rounded-full" />
          </div>
        ) : chats.length === 0 ? (
          <div className="text-center py-8 text-[var(--text-muted)] text-sm">
            {t('chatSidebar.noChats')}
          </div>
        ) : (
          Object.entries(groupedChats).map(([month, monthChats]) => (
            <div key={month} className="mb-3">
              <div className="px-2 py-1.5 text-xs font-medium text-[var(--text-muted)] uppercase tracking-wider">
                {month}
              </div>
              <div className="space-y-1">
                {monthChats.map(chat => (
                  <div
                    key={chat.id}
                    className={`relative px-3.5 py-3.5 rounded-2xl transition-all group cursor-pointer border border-transparent ${
                      selectedChatId === chat.id
                        ? 'bg-[var(--accent)]/8 border-[var(--accent)]/20'
                        : 'hover:bg-[var(--background)] hover:border-[var(--border)] hover:translate-x-0.5'
                    }`}
                    onClick={() => handleSelectChat(chat.id)}
                  >
                    {editingId === chat.id ? (
                      <input
                        type="text"
                        value={editTitle}
                        onChange={(e) => setEditTitle(e.target.value)}
                        onBlur={() => handleSaveRename(chat.id)}
                        onKeyDown={(e) => e.key === 'Enter' && handleSaveRename(chat.id)}
                        className="w-full bg-[var(--background)] text-[var(--text)] text-sm px-3 py-1.5 rounded-lg border border-[var(--accent)] outline-none"
                        autoFocus
                        onClick={(e) => e.stopPropagation()}
                      />
                    ) : (
                      <>
                        <div className="flex items-center gap-2 pr-8">
                          <span
                            className={`h-2 w-2 shrink-0 rounded-full ${
                              isRecentlyActive(chat.updated_at)
                                ? 'bg-[var(--accent)]'
                                : 'bg-[var(--text-muted)]/50'
                            }`}
                            aria-hidden="true"
                          />
                          <div className="text-[14px] font-semibold text-[var(--text)] truncate">
                            {chat.title || t('chatSidebar.newChat')}
                          </div>
                        </div>
                        <div className="text-xs text-[var(--text-muted)] mt-1 ml-4 truncate">
                          {formatDate(chat.updated_at)}
                        </div>
                      </>
                    )}
                    
                    {/* Menu button */}
                    <div className="absolute right-3 top-1/2 -translate-y-1/2">
                      <button
                        onClick={(e) => {
                          e.stopPropagation();
                          setOpenMenuId(openMenuId === chat.id ? null : chat.id);
                        }}
                        className="p-1.5 rounded-lg opacity-0 group-hover:opacity-100 hover:bg-[var(--background)] text-[var(--text-secondary)] transition-all"
                      >
                        <MoreHorizontal size={16} />
                      </button>
                      
                      {/* Dropdown menu */}
                      {openMenuId === chat.id && (
                        <div className="absolute right-0 top-full mt-1.5 w-36 bg-[var(--surface)] border border-[var(--border)] rounded-xl shadow-xl py-1 z-50">
                          <button
                            onClick={(e) => handleRename(chat.id, chat.title || t('chatSidebar.newChat'), e)}
                            className="w-full flex items-center gap-2 px-3.5 py-2 text-sm text-[var(--text)] hover:bg-[var(--background)]"
                          >
                            <Edit2 size={15} />
                            {t('chatSidebar.rename')}
                          </button>
                          <button
                            onClick={(e) => handleDeleteChat(chat.id, e)}
                            className="w-full flex items-center gap-2 px-3.5 py-2 text-sm text-red-500 hover:bg-red-500/10"
                          >
                            <Trash2 size={15} />
                            {t('chatSidebar.delete')}
                          </button>
                        </div>
                      )}
                    </div>
                    
                    {selectedChatId === chat.id && (
                      <div className="absolute left-0 top-1/2 -translate-y-1/2 w-1 h-5 bg-[var(--accent)] rounded-r-full" />
                    )}
                  </div>
                ))}
              </div>
            </div>
          ))
        )}
      </div>

      {/* Footer — a quiet session counter, dock pane only. */}
      {isDock && (
        <div className="border-t border-[var(--border)] px-4 py-2.5 text-xs text-[var(--text-muted)]">
          {hasAuthSession
            ? `${chats.length} ${chats.length === 1 ? 'session' : 'sessions'}`
            : notLoggedInMessage}
        </div>
      )}
    </div>
  );
}
