import { useState, useEffect, useRef } from 'react';
import { Plus, MessageSquare, MoreHorizontal, Edit2, Trash2, X } from 'lucide-react';
import { useAuthStore } from '../stores/authStore';
import { getChatHistory, createChat, deleteChat, type ChatHistoryItem } from '../services/chatHistoryService';
import { t } from '../hooks/useI18n';

interface SidebarProps {
  isOpen: boolean;
  onClose: () => void;
  onSelectChat: (chatId: string) => void;
  onNewChat: () => void;
}

export function Sidebar({ isOpen, onClose, onSelectChat, onNewChat }: SidebarProps) {
  const { user } = useAuthStore();
  const [chats, setChats] = useState<ChatHistoryItem[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [selectedChatId, setSelectedChatId] = useState<string | null>(null);
  const [openMenuId, setOpenMenuId] = useState<string | null>(null);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [editTitle, setEditTitle] = useState('');
  const menuRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (user?.id && isOpen) {
      loadChats();
    }
  }, [user?.id, isOpen]);

  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (menuRef.current && !menuRef.current.contains(event.target as Node)) {
        setOpenMenuId(null);
      }
    }
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  const loadChats = async () => {
    try {
      setIsLoading(true);
      const history = await getChatHistory(user!.id);
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
    try {
      const newChat = await createChat(user!.id, t('chatSidebar.newChat'));
      setChats(prev => [newChat, ...prev]);
      setSelectedChatId(newChat.id);
      onNewChat();
    } catch (error) {
      console.error('Failed to create chat:', error);
    }
  };

  const handleDeleteChat = async (chatId: string, e: React.MouseEvent) => {
    e.stopPropagation();
    try {
      await deleteChat(chatId);
      setChats(prev => prev.filter(c => c.id !== chatId));
      if (selectedChatId === chatId) {
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
      const { updateChatTitle } = await import('../services/chatHistoryService');
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
    <div className={`fixed inset-y-0 left-0 z-40 w-72 bg-[var(--surface)] border-r border-[var(--border)] flex flex-col transition-transform duration-200 ${isOpen ? 'translate-x-0' : '-translate-x-full'} ${isOpen ? '' : 'pointer-events-none'}`}>
      {/* Header */}
      <div className="flex items-center justify-between px-3 py-3 border-b border-[var(--border)]">
        <div className="flex items-center gap-2">
          <MessageSquare size={18} className="text-[var(--accent)]" />
          <span className="font-medium text-[var(--text)]">{t('chatSidebar.history')}</span>
        </div>
        <button
          onClick={onClose}
          className="p-1.5 rounded-lg hover:bg-[var(--background)] text-[var(--text-secondary)] transition-colors"
          title={t('common.close')}
        >
          <X size={18} />
        </button>
      </div>

      {/* New Chat Button */}
      <div className="px-3 py-2">
        <button
          onClick={handleNewChat}
          className="w-full flex items-center justify-center gap-2 px-3 py-2.5 bg-[var(--accent)] hover:bg-[var(--accent)]/90 text-white font-medium rounded-lg transition-colors text-sm"
        >
          <Plus size={16} />
          {t('chatSidebar.newChat')}
        </button>
      </div>

      {/* Chat List */}
      <div className="flex-1 overflow-y-auto px-2 py-2" ref={menuRef}>
        {isLoading ? (
          <div className="flex items-center justify-center py-8">
            <div className="animate-spin w-5 h-5 border-2 border-[var(--accent)] border-t-transparent rounded-full" />
          </div>
        ) : chats.length === 0 ? (
          <div className="text-center py-8 text-[var(--text-muted)] text-sm">
            {t('chatSidebar.noChats')}
          </div>
        ) : (
          Object.entries(groupedChats).map(([month, monthChats]) => (
            <div key={month} className="mb-4">
              <div className="px-2 py-1.5 text-xs font-medium text-[var(--text-muted)] uppercase tracking-wider">
                {month}
              </div>
              <div className="space-y-0.5">
                {monthChats.map(chat => (
                  <div
                    key={chat.id}
                    className={`relative px-2 py-2 rounded-lg transition-colors group ${
                      selectedChatId === chat.id
                        ? 'bg-[var(--accent)]/10'
                        : 'hover:bg-[var(--background)]'
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
                        className="w-full bg-[var(--background)] text-[var(--text)] text-sm px-2 py-1 rounded border border-[var(--accent)] outline-none"
                        autoFocus
                        onClick={(e) => e.stopPropagation()}
                      />
                    ) : (
                      <>
                        <div className="text-sm font-medium truncate pr-8 text-[var(--text)]">
                          {chat.title || t('chatSidebar.newChat')}
                        </div>
                        <div className="text-xs text-[var(--text-muted)] truncate mt-0.5">
                          {formatDate(chat.updated_at)}
                        </div>
                      </>
                    )}
                    
                    {/* Menu button */}
                    <div className="absolute right-2 top-1/2 -translate-y-1/2">
                      <button
                        onClick={(e) => {
                          e.stopPropagation();
                          setOpenMenuId(openMenuId === chat.id ? null : chat.id);
                        }}
                        className="p-1 rounded opacity-0 group-hover:opacity-100 hover:bg-[var(--background)] text-[var(--text-secondary)] transition-all"
                      >
                        <MoreHorizontal size={16} />
                      </button>
                      
                      {/* Dropdown menu */}
                      {openMenuId === chat.id && (
                        <div className="absolute right-0 top-full mt-1 w-32 bg-[var(--surface)] border border-[var(--border)] rounded-lg shadow-lg py-1 z-50">
                          <button
                            onClick={(e) => handleRename(chat.id, chat.title || t('chatSidebar.newChat'), e)}
                            className="w-full flex items-center gap-2 px-3 py-2 text-sm text-[var(--text)] hover:bg-[var(--background)]"
                          >
                            <Edit2 size={14} />
                            {t('chatSidebar.rename')}
                          </button>
                          <button
                            onClick={(e) => handleDeleteChat(chat.id, e)}
                            className="w-full flex items-center gap-2 px-3 py-2 text-sm text-red-500 hover:bg-red-500/10"
                          >
                            <Trash2 size={14} />
                            {t('chatSidebar.delete')}
                          </button>
                        </div>
                      )}
                    </div>
                    
                    {selectedChatId === chat.id && (
                      <div className="absolute left-0 top-1/2 -translate-y-1/2 w-0.5 h-4 bg-[var(--accent)] rounded-r" />
                    )}
                  </div>
                ))}
              </div>
            </div>
          ))
        )}
      </div>
    </div>
  );
}