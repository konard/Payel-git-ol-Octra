import { useState, useEffect, useRef } from 'react';
import { ReactFlowProvider } from '@xyflow/react';
import { TopBar } from './components/TopBar';
import { StatusBar } from './components/StatusBar';
import { ConsolePanel } from './components/ConsolePanel';
import { Canvas } from './components/Canvas';
import { Chat, type ChatMessage } from './components/Chat';
import { SolutionViewer } from './components/SolutionViewer';
import { BottomInput } from './components/BottomInput';
import { ChatInput } from './components/ChatInput';
import type { TaskData } from './components/BottomInput';
import { useWebSocket } from '../hooks/useWebSocket';
import { useTaskStore } from '../stores/taskStore';
import { useI18n } from '../hooks/useI18n';
import { AuthModal } from '../components/AuthModal';
import { SubscriptionModal } from '../components/SubscriptionModal';
import { addMessage, getChat, updateChatWorkflow, createChat } from '../services/chatHistoryService';
import { PaymentSuccess } from '../components/PaymentSuccess';
import { Sidebar } from '../components/Sidebar';
import { useAuthStore } from '../stores/authStore';
import { useThemeStore } from '../stores/themeStore';
import { useCustomProvidersStore } from '../stores/customProvidersStore';
import { useSettingsStore } from '../stores/settingsStore';
import LandingPage from './components/LandingPage';
import { buildTaskProviderAuth, buildWorkflowConfigFromGraph } from './taskPayload';

const SHOW_STATUS_BAR = false;

// Serialize the current canvas graph into the compact JSON shape stored per chat.
function serializeCurrentWorkflow(): string {
  const { nodes, edges } = useTaskStore.getState();
  return JSON.stringify({
    nodes: nodes.map((n) => ({
      id: n.id,
      type: n.type,
      role: n.role,
      status: n.status,
      position: n.position,
      filesCount: n.filesCount,
      workerCount: n.workerCount,
      techStack: n.techStack,
      scale: n.scale,
      repoUrl: n.repoUrl,
    })),
    edges: edges.map((e) => ({ from: e.from, to: e.to })),
  });
}

// Persist the current canvas graph to the given chat so it can be restored later.
async function persistWorkflow(chatId: string | null) {
  if (!chatId) return;
  try {
    await updateChatWorkflow(chatId, serializeCurrentWorkflow());
  } catch (error) {
    console.error('Failed to persist chat workflow:', error);
  }
}

// Parse a stored chat workflow string into store-ready nodes and edges.
function parseChatWorkflow(raw?: string): { nodes: any[]; edges: any[] } {
  if (!raw) return { nodes: [], edges: [] };
  try {
    const data = JSON.parse(raw);
    const nodes = (data.nodes || []).map((node: any) => ({
      id: node.id,
      type: node.type,
      role: node.role,
      status: node.status || 'done',
      position: node.position,
      filesCount: node.filesCount,
      workerCount: node.workerCount,
      techStack: node.techStack,
      scale: node.scale,
      repoUrl: node.repoUrl,
    }));
    const edges = (data.edges || []).map((edge: any) => ({ from: edge.from, to: edge.to }));
    return { nodes, edges };
  } catch (error) {
    console.error('Failed to parse chat workflow:', error);
    return { nodes: [], edges: [] };
  }
}

export default function App() {
  const [isExpanded, setIsExpanded] = useState(false);
  const [showAuthModal, setShowAuthModal] = useState(false);
  const [showSubscriptionModal, setShowSubscriptionModal] = useState(false);
  const [showSidebar, setShowSidebar] = useState(false);
  const [mode, setMode] = useState<'canvas' | 'chat' | 'solution'>('canvas');
  const [currentChatId, setCurrentChatId] = useState<string | null>(null);
  const currentChatIdRef = useRef<string | null>(null);

  // Chat state
  const [chatMessages, setChatMessages] = useState<ChatMessage[]>([]);
  const [hasUnreadMessages, setHasUnreadMessages] = useState(true);

  const { language, changeLanguage } = useI18n();
  const { isAuthenticated, hasSubscription, isInTrial, checkAuth, setSubscription } = useAuthStore();
  const { isDark, toggleTheme } = useThemeStore();
  const defaultToken = useSettingsStore((state) => state.defaultToken);
  const defaultProvider = useSettingsStore((state) => state.defaultProvider);
  const defaultModel = useSettingsStore((state) => state.defaultModel);

  useEffect(() => {
    currentChatIdRef.current = currentChatId;
  }, [currentChatId]);

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

  const persistChatMessage = (role: 'user' | 'boss', content: string) => {
    const chatId = currentChatIdRef.current;
    if (!chatId || !content.trim()) return;

    addMessage(chatId, role, content, {
      provider: defaultProvider,
      model: defaultModel,
    }).catch((error) => {
      console.error('Failed to persist chat message:', error);
    });
  };

  const ensureChatSession = async (title: string) => {
    if (currentChatIdRef.current) {
      return currentChatIdRef.current;
    }

    const authUser = useAuthStore.getState().user;
    if (!authUser?.id) {
      return null;
    }

    const newChat = await createChat(authUser.id, title);
    setCurrentChatId(newChat.id);
    currentChatIdRef.current = newChat.id;
    return newChat.id;
  };

  const buildChatTaskPayload = (message: string) => {
    const authUser = useAuthStore.getState().user;
    if (!authUser || !isAuthenticated || (!hasSubscription && !isInTrial)) {
      return null;
    }

    const customProviders = useCustomProvidersStore.getState().providers;
    const customProvider = customProviders.find(p => p.id === defaultProvider);
    const title = message.trim().slice(0, 60) || 'Chat workflow';

    const taskPayload: any = {
      username: authUser.username || 'user',
      user_id: authUser.id,
      title,
      description: message,
      meta: {
        model: defaultModel,
      },
    };

    const providerAuth = buildTaskProviderAuth({
      provider: defaultProvider,
      defaultToken,
      customProvider,
    });

    taskPayload.tokens = providerAuth.tokens;
    taskPayload.meta.provider = providerAuth.provider;

    const workflow = buildWorkflowConfigFromGraph(
      useTaskStore.getState().nodes,
      useTaskStore.getState().edges,
    );
    if (workflow) {
      taskPayload.workflow = workflow;
    }

    return taskPayload;
  };

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
      persistChatMessage('boss', message);
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
  const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
  const wsUrl = import.meta.env.VITE_WS_URL || `${wsProtocol}//${window.location.host}/api/task/create`;
  const { send, sendChat } = useWebSocket(wsUrl, handleIncomingChatMessage, updateBossProgress);

  const status = useTaskStore((state) => state.status);
  const isSubmitting = status === 'creating' || status === 'planning' || status === 'executing';

  // Persist the workflow to the chat whenever a task finishes.
  useEffect(() => {
    if (status === 'done' && currentChatId) {
      void persistWorkflow(currentChatId);
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
    // Handle OAuth callback tokens
    const urlParams = new URLSearchParams(window.location.search);
    const oauthToken = urlParams.get('token');
    const oauthRefreshToken = urlParams.get('refresh_token');

    if (oauthToken) {
      // Save the access token
      localStorage.setItem('access_token', oauthToken);
      if (oauthRefreshToken) {
        localStorage.setItem('refresh_token', oauthRefreshToken);
      }
      document.cookie = `access_token=${oauthToken}; path=/; SameSite=Lax`;
      
      // Clean the URL
      window.history.replaceState({}, document.title, window.location.pathname);
      
      // Update store with token
      useAuthStore.setState({
        accessToken: oauthToken,
        refreshToken: oauthRefreshToken || useAuthStore.getState().refreshToken,
        isAuthenticated: true,
      });

      // Immediately fetch user data
      const { checkAuth } = useAuthStore.getState();
      checkAuth().catch(() => {
        // If checkAuth fails, at least keep the token
        console.log('OAuth login successful, token saved');
      });
    } else {
      // Normal auth check
      checkAuth();
    }
  }, []);

  const handleAuthSuccess = () => {
    setShowAuthModal(false);
    setShowSubscriptionModal(true);
  };

  const handleShowAuth = () => {
    setShowAuthModal(true);
  };

  const handleCreateTask = async (data: TaskData) => {
    await checkAuth();
    const { isAuthenticated, hasSubscription, isInTrial } = useAuthStore.getState();

    if (!isAuthenticated) {
      setShowAuthModal(true);
      return;
    }

    // Allow usage during 14-day trial or with active subscription
    if (!hasSubscription && !isInTrial) {
      setShowSubscriptionModal(true);
      return;
    }

    const authUser = useAuthStore.getState().user;

    // Create a new chat session if none exists
    let chatId = currentChatId;
    if (!chatId && authUser?.id) {
      try {
        const newChat = await createChat(authUser.id, data.title);
        chatId = newChat.id;
        setCurrentChatId(chatId);
        currentChatIdRef.current = chatId;
        // Update sidebar to reflect the new chat
        // Note: Sidebar will reload chats automatically when opened
      } catch (error) {
        console.error('Failed to create chat session:', error);
        // Continue without chat session for now
      }
    }

    const store = useTaskStore.getState();
    const workflow = buildWorkflowConfigFromGraph(store.nodes, store.edges);

    // Reset only task execution state, keep user workflow
    store.resetTaskExecution();

    // Check if this is a custom provider
    const customProviders = useCustomProvidersStore.getState().providers;
    const customProvider = customProviders.find(p => p.id === data.provider);

    let taskPayload: any = {
      username: authUser?.username || 'user',
      user_id: authUser?.id || 'local',
      title: data.title,
      description: data.description,
      meta: {
        model: data.model,
      },
    };

    // Send user workflow in the backend's predefined manager/worker shape.
    if (workflow) {
      taskPayload.workflow = workflow;
    }

    const providerAuth = buildTaskProviderAuth({
      provider: data.provider,
      apiKey: data.apiKey,
      defaultToken,
      customProvider,
    });

    taskPayload.tokens = providerAuth.tokens;
    taskPayload.meta.provider = providerAuth.provider;

    console.log('Sending task payload:', JSON.stringify(taskPayload, null, 2));
    send(taskPayload);
    useTaskStore.getState().setStartTime(Date.now());
  };

  const handleSendChatMessage = async (message: string) => {
    await checkAuth();
    const { isAuthenticated } = useAuthStore.getState();
    if (!isAuthenticated) {
      setShowAuthModal(true);
      return;
    }

    try {
      await ensureChatSession(message.slice(0, 60) || 'New Chat');
    } catch (error) {
      console.error('Failed to create chat session:', error);
    }

    const newMessage = {
      id: Date.now().toString(),
      text: message,
      sender: 'user' as const,
      timestamp: new Date(),
      read: true,
    };
    setChatMessages(prev => [...prev, newMessage]);
    setMode('chat');
    persistChatMessage('user', message);

    const taskPayload = buildChatTaskPayload(message);
    const sent = await sendChat(message, 'user', taskPayload);

    if (!sent) {
      handleIncomingChatMessage('I cannot reach the boss service right now. Your message is saved in this chat.', 'boss');
    }
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

  const handleNewChat = async (chatId?: string) => {
    // Save the workflow of the chat we are leaving before clearing the canvas.
    const prevChatId = currentChatIdRef.current;
    if (prevChatId && prevChatId !== chatId) {
      await persistWorkflow(prevChatId);
    }

    setCurrentChatId(chatId ?? null);
    currentChatIdRef.current = chatId ?? null;
    setChatMessages([]);
    setMode('chat');
    // Start every chat from a clean canvas so nodes never leak across chats.
    useTaskStore.getState().setGraph([], []);
  };

  const handleSelectChat = async (chatId: string) => {
    // Save the workflow we are leaving so it is restored next time it is opened.
    const prevChatId = currentChatIdRef.current;
    if (prevChatId && prevChatId !== chatId) {
      await persistWorkflow(prevChatId);
    }

    setCurrentChatId(chatId);
    currentChatIdRef.current = chatId;
    setMode('chat');
    setShowSidebar(false);

    const store = useTaskStore.getState();
    // Clear the previous chat's graph immediately to avoid cross-chat contamination
    // while the selected chat loads asynchronously.
    store.setGraph([], []);

    try {
      const chat = await getChat(chatId);
      setChatMessages((chat.messages || []).map((message) => ({
        id: message.id,
        text: message.content,
        sender: message.role,
        timestamp: new Date(message.created_at || message.timestamp || Date.now()),
        read: true,
      })));

      // Restore this chat's saved workflow (a full graph swap, empty if none).
      const { nodes, edges } = parseChatWorkflow(chat.workflow);
      useTaskStore.getState().setGraph(nodes, edges);
    } catch (err) {
      console.error('Failed to load chat:', err);
    }
  };

  return (
    <div className="h-screen flex flex-col bg-[var(--background)] text-[var(--text)] overflow-hidden">
      <div
        className={`fixed inset-0 bg-black/60 z-30 transition-opacity duration-200 ${showSidebar ? 'opacity-100' : 'opacity-0 pointer-events-none'}`}
        onClick={() => setShowSidebar(false)}
      />
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
        ) : mode === 'solution' ? (
          <SolutionViewer />
        ) : (
          <Chat
            messages={chatMessages}
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
      ) : mode === 'solution' ? null : (
        <ChatInput onSendMessage={handleSendChatMessage} />
      )}

      {showAuthModal && !isAuthenticated && (
        <AuthModal
          onClose={() => setShowAuthModal(false)}
          onAuthSuccess={handleAuthSuccess}
        />
      )}

      {showSubscriptionModal && isAuthenticated && !hasSubscription && !isInTrial && (
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
