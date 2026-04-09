import { useState, useEffect } from 'react';
import { TopBar } from './components/TopBar';
import { StatusBar } from './components/StatusBar';
import { ConsolePanel } from './components/ConsolePanel';
import { Canvas } from './components/Canvas';
import { BottomInput } from './components/BottomInput';
import type { TaskData } from './components/BottomInput';
import { useWebSocket } from '../hooks/useWebSocket';
import { useTaskStore } from '../stores/taskStore';
import { useI18n } from '../hooks/useI18n';
import { AuthModal } from '../components/AuthModal';
import { SubscriptionModal } from '../components/SubscriptionModal';
import { useAuthStore } from '../stores/authStore';
import { useThemeStore } from '../stores/themeStore';

export default function App() {
  const [isExpanded, setIsExpanded] = useState(false);
  const [showAuthModal, setShowAuthModal] = useState(false);
  const [showSubscriptionModal, setShowSubscriptionModal] = useState(false);
  const { language, changeLanguage } = useI18n();
  const { isAuthenticated, hasSubscription, checkAuth, setSubscription } = useAuthStore();
  const { isDark, toggleTheme } = useThemeStore();

  const wsUrl = import.meta.env.VITE_WS_URL || `ws://${window.location.host}/api/task/create`;
  const { connect, send } = useWebSocket(wsUrl);

  const status = useTaskStore((state) => state.status);
  const isSubmitting = status === 'creating' || status === 'planning' || status === 'executing';

  useEffect(() => {
    if (isDark) {
      document.documentElement.classList.add('dark');
    } else {
      document.documentElement.classList.remove('dark');
    }
  }, [isDark]);

  // Проверяем авторизацию при загрузке
  useEffect(() => {
    checkAuth();
  }, []);

  // Подключаемся к WebSocket только если авторизованы
  useEffect(() => {
    if (isAuthenticated) {
      connect();
    }
  }, [isAuthenticated]);

  const handleAuthSuccess = () => {
    setShowAuthModal(false);
    // После авторизации показываем подписку
    setShowSubscriptionModal(true);
  };

  const handleShowAuth = () => {
    setShowAuthModal(true);
  };

  const handleCreateTask = (data: TaskData) => {
    // Проверяем подписку перед отправкой задачи
    if (!hasSubscription) {
      setShowSubscriptionModal(true);
      return;
    }

    // Reset previous task state (clear canvas)
    useTaskStore.getState().resetTask();

    const tokenKey = data.provider === 'openrouter' ? 'openrouter'
      : data.provider === 'gemini' ? 'gemini'
      : data.provider === 'openai' ? 'openai'
      : 'claude';

    send({
      username: 'user',
      title: data.title,
      description: data.description,
      tokens: {
        [tokenKey]: data.apiKey,
      },
      meta: {
        provider: data.provider,
        model: data.model,
      },
    });

    useTaskStore.getState().setStartTime(Date.now());
  };

  const toggleExpand = () => {
    setIsExpanded(!isExpanded);
  };

  return (
    <div className="h-screen flex flex-col bg-[var(--background)] text-[var(--text)] overflow-hidden">
      <TopBar
        isAuthenticated={isAuthenticated}
        hasSubscription={hasSubscription}
        onShowAuth={handleShowAuth}
        onShowSubscription={() => setShowSubscriptionModal(true)}
      />

      <Canvas />

      <StatusBar />

      <ConsolePanel />

      <BottomInput
        onSubmit={handleCreateTask}
        isSubmitting={isSubmitting}
        isExpanded={isExpanded}
        onToggleExpand={toggleExpand}
      />

      {showAuthModal && !isAuthenticated && (
        <AuthModal
          onClose={() => setShowAuthModal(false)}
          onAuthSuccess={handleAuthSuccess}
        />
      )}

      {showSubscriptionModal && isAuthenticated && (
        <SubscriptionModal
          onClose={() => setShowSubscriptionModal(false)}
          onSubscribe={() => {
            setSubscription(true);
            setShowSubscriptionModal(false);
          }}
        />
      )}
    </div>
  );
}
