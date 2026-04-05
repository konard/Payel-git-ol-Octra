import { useState, useEffect } from 'react';
import { TopBar } from './components/TopBar';
import { StatusBar } from './components/StatusBar';
import { Canvas } from './components/Canvas';
import { BottomInput } from './components/BottomInput';
import type { TaskData } from './components/BottomInput';
import { useWebSocket } from '../hooks/useWebSocket';
import { useTaskStore } from '../stores/taskStore';

export default function App() {
  const [isDark, setIsDark] = useState(false);
  const [isExpanded, setIsExpanded] = useState(false);
  
  const { connect, send } = useWebSocket('ws://localhost:3111/task/create');
  
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
    connect();
  }, []);

  const handleCreateTask = (data: TaskData) => {
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

  const toggleTheme = () => {
    setIsDark(!isDark);
  };

  const toggleExpand = () => {
    setIsExpanded(!isExpanded);
  };

  return (
    <div className="h-screen flex flex-col bg-[var(--background)] text-[var(--text)] overflow-hidden">
      <TopBar
        onToggleTheme={toggleTheme}
        isDark={isDark}
      />
      
      <Canvas />
      
      <StatusBar />
      
      <BottomInput
        onSubmit={handleCreateTask}
        isSubmitting={isSubmitting}
        isExpanded={isExpanded}
        onToggleExpand={toggleExpand}
      />
    </div>
  );
}
