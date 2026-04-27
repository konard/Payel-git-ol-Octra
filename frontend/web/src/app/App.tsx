import { useState, useEffect } from 'react';
import { ReactFlowProvider } from '@xyflow/react';
import { TopBar } from './components/TopBar';
import { StatusBar } from './components/StatusBar';
import { ConsolePanel } from './components/ConsolePanel';
import { Canvas } from './components/Canvas';
import { Chat } from './components/Chat';
import { BottomInput } from './components/BottomInput';
import { ChatInput } from './components/ChatInput';
import type { TaskData } from './components/BottomInput';
import { useWebSocket } from '../hooks/useWebSocket';
import { useTaskStore } from '../stores/taskStore';
import { useI18n } from '../hooks/useI18n';
import { AuthModal } from '../components/AuthModal';
import { SubscriptionModal } from '../components/SubscriptionModal';
import { getChat, updateChatWorkflow } from '../services/chatHistoryService';
import { PaymentSuccess } from '../components/PaymentSuccess';
import { Sidebar } from '../components/Sidebar';
import { useAuthStore } from '../stores/authStore';
import { useThemeStore } from '../stores/themeStore';
import { useCustomProvidersStore } from '../stores/customProvidersStore';
import LandingPage from './components/LandingPage';

const SHOW_STATUS_BAR = false;

export default function App() {
  const [isExpanded, setIsExpanded] = useState(false);
  const [showAuthModal, setShowAuthModal] = useState(false);
  const [showSubscriptionModal, setShowSubscriptionModal] = useState(false);
  const [showSidebar, setShowSidebar] = useState(false);
  const [mode, setMode] = useState<'canvas' | 'chat'>('canvas');
  const [currentChatId, setCurrentChatId] = useState<string | null>(null);

  // Chat state
  const [chatMessages, setChatMessages] = useState<Array<{
    id: string;
    text: string;
    sender: 'boss' | 'user';
    timestamp: Date;
    read?: boolean;
    isClarification?: boolean;
  }>>([]);
  const [hasUnreadMessages, setHasUnreadMessages] = useState(true);

  const { language, changeLanguage } = useI18n();
  const { isAuthenticated, hasSubscription, checkAuth, setSubscription } = useAuthStore();
  const { isDark, toggleTheme } = useThemeStore();

  // Определяем, показывать лендинг или приложение
  const isLandingPage = window.location.pathname === '/' && !isAuthenticated;
  const isPaymentSuccess = window.location.pathname === '/payment-success';

  // Если мы на главной и не авторизован - показываем лендинг
  if (isLandingPage) {
    return <LandingPage />;
  }

  // Если успешная оплата - показываем страницу успеха
  if (isPaymentSuccess) {
    return <PaymentSuccess />;
  }

  const handleIncomingChatMessage = (message: string, sender: 'boss' | 'user', isClarification = false, progress?: number, showProgress?: boolean) => {
    const newMessage = {
      id: Date.now().toString(),
      text: message,
      sender,
      timestamp: new Date(),
      read: sender === 'user', // User messages are automatically read
      isClarification,
      progress,
      showProgress,
    };
    setChatMessages(prev => [...prev, newMessage]);
    if (sender === 'boss') {
      setHasUnreadMessages(true);
    }
  };

  const updateBossProgress = (progress: number, message?: string) => {
    setChatMessages(prev => {
      const lastBossIndex = prev.findLastIndex(m => m.sender === 'boss');
      if (lastBossIndex >= 0) {
        const updatedMessages = [...prev];
        updatedMessages[lastBossIndex] = {
          ...updatedMessages[lastBossIndex],
          progress,
          text: message || updatedMessages[lastBossIndex].text,
          showProgress: true,
        };
        return updatedMessages;
      } else {
        // Create new boss message with progress
        const newMessage = {
          id: Date.now().toString(),
          text: message || 'Processing your request...',
          sender: 'boss' as const,
          timestamp: new Date(),
          read: false,
          progress,
          showProgress: true,
        };
        return [...prev, newMessage];
      }
    });
    setHasUnreadMessages(true);
  };

  // Иначе показываем приложение
  const wsUrl = import.meta.env.VITE_WS_URL || `ws://${window.location.host}/api/task/create`;
  const { send, sendChat } = useWebSocket(wsUrl, handleIncomingChatMessage, updateBossProgress);

  const status = useTaskStore((state) => state.status);
  const isSubmitting = status === 'creating' || status === 'planning' || status === 'executing';

  // Сохранять workflow в чат при завершении таски
  useEffect(() => {
    if (status === 'done' && currentChatId) {
      const nodes = useTaskStore.getState().nodes;
      const edges = useTaskStore.getState().edges;
      if (nodes.length > 0) {
        const workflowData = JSON.stringify({
          nodes: nodes.map(n => ({
            id: n.id,
            type: n.type,
            role: n.role,
            status: n.status,
            position: n.position,
            filesCount: n.filesCount,
            workerCount: n.workerCount,
          })),
          edges: edges.map(e => ({ from: e.from, to: e.to })),
        });
        updateChatWorkflow(currentChatId, workflowData).catch(console.error);
      }
    }
  }, [status, currentChatId]);

  useEffect(() => {
    if (isDark) {
      document.documentElement.classList.add('dark');
    } else {
      document.documentElement.classList.remove('dark');
    }
  }, [isDark]);

  useEffect(() => {
    checkAuth();
  }, []);

  const handleAuthSuccess = () => {
    setShowAuthModal(false);
    setShowSubscriptionModal(true);
  };

  const handleShowAuth = () => {
    setShowAuthModal(true);
  };

  const handleCreateTask = (data: TaskData) => {
    if (!hasSubscription) {
      setShowSubscriptionModal(true);
      return;
    }

    const store = useTaskStore.getState();
    const hasUserDefinedWorkflow = store.nodes.some(node =>
      !node.id.startsWith('boss-') &&
      !node.id.startsWith('manager-') &&
      !node.id.startsWith('worker-') &&
      node.id !== 'zip-archive'
    );
    const workflow = hasUserDefinedWorkflow ? store.getWorkflow() : null;

    store.resetTask();

    // Check if this is a custom provider
    const customProviders = useCustomProvidersStore.getState().providers;
    const customProvider = customProviders.find(p => p.id === data.provider);

    let taskPayload: any = {
      username: 'user',
      title: data.title,
      description: data.description,
      meta: {
        model: data.model,
      },
      ...(workflow && { workflow }),
    };

    if (customProvider) {
      // For custom providers, send base_url in tokens for the custom provider
      taskPayload.meta.provider = 'ollama';
      taskPayload.tokens = {
        ollama: data.apiKey,
        base_url: customProvider.base_url,
      };
    } else {
      // For standard providers, use existing logic
      const tokenKey = data.provider === 'openrouter' ? 'openrouter'
        : data.provider === 'gemini' ? 'gemini'
        : data.provider === 'openai' ? 'openai'
        : data.provider === 'zai' ? 'zai'
        : 'claude';

      taskPayload.tokens = {
        [tokenKey]: data.apiKey,
      };
      taskPayload.meta.provider = data.provider;
    }

    console.log('Sending task payload:', JSON.stringify(taskPayload, null, 2));
    send(taskPayload);
    useTaskStore.getState().setStartTime(Date.now());
  };

  const handleSendChatMessage = (message: string) => {
    const newMessage = {
      id: Date.now().toString(),
      text: message,
      sender: 'user' as const,
      timestamp: new Date(),
    };
    setChatMessages(prev => [...prev, newMessage]);

    // Send chat message via WebSocket
    sendChat(message, 'user');
  };

  const handleMarkChatMessageAsRead = (messageId: string) => {
    setChatMessages(prev =>
      prev.map(msg =>
        msg.id === messageId ? { ...msg, read: true } : msg
      )
    );
    // Check if there are any unread boss messages left
    setHasUnreadMessages(chatMessages.some(msg => msg.sender === 'boss' && !msg.read && msg.id !== messageId));
  };

  const handleStopTask = async () => {
    const taskId = useTaskStore.getState().taskId;
    if (!taskId) return;

    try {
      // Отправляем HTTP запрос для остановки задачи
      const response = await fetch(`/api/task/${taskId}/stop`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
      });

      if (response.ok) {
        useTaskStore.getState().setTaskStatus('cancelled');
        useTaskStore.getState().addLog({
          message: 'Task cancelled by user',
          type: 'warning',
        });
      } else {
        throw new Error('Failed to stop task');
      }
    } catch (error) {
      console.error('Error stopping task:', error);
      useTaskStore.getState().addLog({
        message: 'Failed to cancel task',
        type: 'error',
      });
    }
  };



  const toggleExpand = () => {
    setIsExpanded(!isExpanded);
  };

  const handleNewChat = () => {
    setCurrentChatId(null);
    setChatMessages([]);
    setMode('chat');
  };

  const handleSelectChat = async (chatId: string) => {
    setCurrentChatId(chatId);
    try {
      const chat = await getChat(chatId);
      if (chat.workflow) {
        const workflowData = JSON.parse(chat.workflow);
        if (workflowData.nodes?.length || workflowData.edges?.length) {
          useTaskStore.getState().resetTask();
          workflowData.nodes?.forEach((node: any) => {
            useTaskStore.getState().addNode({
              id: node.id,
              type: node.type,
              role: node.role,
              status: node.status || 'done',
              position: node.position,
            });
          });
          workflowData.edges?.forEach((edge: any) => {
            useTaskStore.getState().addEdge(edge);
          });
        }
      }
    } catch (err) {
      console.error('Failed to load chat workflow:', err);
    }
  };

  return (
    <div className="h-screen flex flex-col bg-[var(--background)] text-[var(--text)] overflow-hidden">
      <Sidebar
        isOpen={showSidebar}
        onClose={() => setShowSidebar(false)}
        onSelectChat={handleSelectChat}
        onNewChat={handleNewChat}
      />
      <TopBar
        isAuthenticated={isAuthenticated}
        hasSubscription={hasSubscription}
        onShowAuth={handleShowAuth}
        onShowSubscription={() => setShowSubscriptionModal(true)}
        mode={mode}
        onModeChange={setMode}
        hasUnreadMessages={hasUnreadMessages}
        onToggleSidebar={() => setShowSidebar(true)}
      />

      <ReactFlowProvider>
        {mode === 'canvas' ? (
          <Canvas mode={mode} onModeChange={setMode} hasUnreadMessages={hasUnreadMessages} />
        ) : (
          <Chat
            messages={chatMessages}
            onSendMessage={handleSendChatMessage}
            onMarkAsRead={handleMarkChatMessageAsRead}
          />
        )}
      </ReactFlowProvider>

      {SHOW_STATUS_BAR && <StatusBar />}

      <ConsolePanel />

      {mode === 'canvas' ? (
        <BottomInput
          onSubmit={handleCreateTask}
          onStop={handleStopTask}
          isSubmitting={isSubmitting}
          isExpanded={isExpanded}
          onToggleExpand={toggleExpand}
        />
      ) : (
        <ChatInput onSendMessage={handleSendChatMessage} />
      )}

      {showAuthModal && !isAuthenticated && (
        <AuthModal
          onClose={() => setShowAuthModal(false)}
          onAuthSuccess={handleAuthSuccess}
        />
      )}

      {showSubscriptionModal && isAuthenticated && !hasSubscription && false && (
        <SubscriptionModal
          onClose={() => setShowSubscriptionModal(false)}
          onSubscribe={async () => {
            setSubscription(true);
            setShowSubscriptionModal(false);
            // Refresh auth status from server to get updated subscription info
            await checkAuth();
          }}
        />
      )}
    </div>
  );
}
