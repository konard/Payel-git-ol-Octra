import { useState, useEffect } from 'react';
import { ReactFlowProvider } from '@xyflow/react';
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
import LandingPage from './components/LandingPage';

export default function App() {
  const [isExpanded, setIsExpanded] = useState(false);
  const [showAuthModal, setShowAuthModal] = useState(false);
  const [showSubscriptionModal, setShowSubscriptionModal] = useState(false);
  const { language, changeLanguage } = useI18n();
  const { isAuthenticated, hasSubscription, checkAuth, setSubscription } = useAuthStore();
  const { isDark, toggleTheme } = useThemeStore();

  // Определяем, показывать лендинг или приложение
  const isLandingPage = window.location.pathname === '/' && !isAuthenticated;

  // Если мы на главной и не авторизован - показываем лендинг
  if (isLandingPage) {
    return <LandingPage />;
  }

  // Иначе показываем приложение
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

  useEffect(() => {
    checkAuth();
  }, []);

  useEffect(() => {
    if (isAuthenticated) {
      connect();
    }
  }, [isAuthenticated]);

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

    useTaskStore.getState().resetTask();

    const tokenKey = data.provider === 'openrouter' ? 'openrouter'
      : data.provider === 'gemini' ? 'gemini'
      : data.provider === 'openai' ? 'openai'
      : data.provider === 'zai' ? 'zai'
      : 'claude';

    // Get current workflow from canvas (user-created nodes and edges)
    const { nodes, edges } = useTaskStore.getState();
    const userNodes = nodes.filter(node =>
      !node.id.startsWith('boss-') &&
      !node.id.startsWith('manager-') &&
      !node.id.startsWith('worker-') &&
      node.id !== 'zip-archive'
    );
    const userEdges = edges.filter(edge =>
      userNodes.some(node => node.id === edge.source || node.id === edge.target)
    );

    const workflow = userNodes.length > 0 ? {
      useAiPlanning: false, // Use custom workflow
      managers: userNodes.map(node => ({
        role: node.role || 'Manager',
        description: node.role || 'Custom manager',
        priority: 1,
        workers: [] // Workers will be created by AI based on manager role
      })),
      architecture: `Custom workflow with ${userNodes.length} managers`,
      techStack: []
    } : useTaskStore.getState().getWorkflow();

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
      ...(workflow && { workflow }),
    });

    useTaskStore.getState().setStartTime(Date.now());
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

  return (
    <div className="h-screen flex flex-col bg-[var(--background)] text-[var(--text)] overflow-hidden">
      <TopBar
        isAuthenticated={isAuthenticated}
        hasSubscription={hasSubscription}
        onShowAuth={handleShowAuth}
        onShowSubscription={() => setShowSubscriptionModal(true)}
      />

      <ReactFlowProvider>
        <Canvas />
      </ReactFlowProvider>

      <StatusBar />

      <ConsolePanel />

      <BottomInput
        onSubmit={handleCreateTask}
        onStop={handleStopTask}
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
