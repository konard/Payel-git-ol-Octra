import { useState, useRef, useEffect } from 'react';
import { Send } from 'lucide-react';
import crewaiMascot from '../../images/crewai-mascot.png';
import { ChatProgressBar } from './ChatProgressBar';
import { useTaskStore } from '../../stores/taskStore';

interface ChatMessage {
  id: string;
  text: string;
  sender: 'boss' | 'user';
  timestamp: Date;
  read?: boolean;
  isClarification?: boolean;
}

interface ChatProps {
  messages: ChatMessage[];
  onSendMessage: (message: string) => void;
  onMarkAsRead: (messageId: string) => void;
}

export function Chat({ messages, onSendMessage, onMarkAsRead }: ChatProps) {
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const nodes = useTaskStore((state) => state.nodes);

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  useEffect(() => {
    // Mark messages as read when they come into view
    const timer = setTimeout(() => {
      messages.forEach(message => {
        if (message.sender === 'boss' && !message.read) {
          onMarkAsRead(message.id);
        }
      });
    }, 100); // Small delay to ensure DOM is updated
    return () => clearTimeout(timer);
  }, [messages, onMarkAsRead]);

  // Get managers and workers
  const managers = nodes.filter(node => node.type === 'manager');
  const workers = nodes.filter(node => node.type === 'worker');

  return (
    <div className="flex flex-col h-full bg-[var(--background)]">
      {/* Progress Bar */}
      <div className="p-4 pb-2">
        <ChatProgressBar />
      </div>

      {/* User task area */}
      <div className="px-4 pb-4">
        <div className="bg-[var(--surface)] rounded-lg p-4 border border-[var(--border)]">
          <div className="text-sm text-[var(--text-muted)] mb-2">user</div>
          <div className="text-[var(--text)]">
            {messages.find(m => m.sender === 'user')?.text || 'Создание новой задачи...'}
          </div>
        </div>
      </div>

      {/* Mascot and progress */}
      <div className="px-4 pb-4">
        <div className="flex items-center gap-4">
          <img
            src={crewaiMascot}
            alt="CrewAI Mascot"
            className="w-16 h-16 rounded-lg object-contain"
          />
          <div className="flex-1">
            <div className="bg-[var(--surface)] border border-[var(--border)] rounded-lg p-2">
              <div className="w-full bg-[var(--accent)] rounded-lg h-2" style={{ width: '50%' }}></div>
            </div>
          </div>
        </div>
      </div>

      {/* Team status */}
      <div className="px-4 pb-4 flex-1 overflow-y-auto">
        <div className="space-y-2">
          {managers.map((manager) => (
            <div key={manager.id} className="flex items-center gap-2">
              <div className={`w-2 h-2 rounded-full ${
                manager.status === 'done' ? 'bg-green-500' :
                manager.status === 'working' ? 'bg-blue-500' :
                'bg-gray-400'
              }`}></div>
              <span className="text-sm text-[var(--text)]">{manager.role}</span>
            </div>
          ))}
          {workers.map((worker) => (
            <div key={worker.id} className="flex items-center gap-2 ml-4">
              <div className={`w-2 h-2 rounded-full ${
                worker.status === 'done' ? 'bg-green-500' :
                worker.status === 'working' ? 'bg-blue-500' :
                'bg-gray-400'
              }`}></div>
              <span className="text-sm text-[var(--text)]">{worker.role}</span>
            </div>
          ))}
        </div>
      </div>

      {/* Messages area */}
      <div className="flex-1 overflow-y-auto px-4 space-y-4">
        {messages.length > 0 && (
          messages.map((message) => (
            <div
              key={message.id}
              className={`flex ${message.sender === 'user' ? 'justify-end' : 'justify-start'}`}
            >
              <div
                className={`max-w-[70%] rounded-lg px-4 py-2 ${
                  message.sender === 'user'
                    ? 'bg-[var(--accent)] text-white'
                    : message.isClarification
                    ? 'bg-orange-100 dark:bg-orange-900 border border-orange-300 dark:border-orange-700 text-orange-900 dark:text-orange-100'
                    : 'bg-[var(--surface)] text-[var(--text)] border border-[var(--border)]'
                }`}
              >
                <div className="text-sm">{message.text}</div>
                <div className={`text-xs mt-1 ${
                  message.sender === 'user' ? 'text-white/70' : 'text-[var(--text-muted)]'
                }`}>
                  {message.timestamp.toLocaleTimeString([], {hour: '2-digit', minute: '2-digit'})}
                </div>
              </div>
            </div>
          ))
        )}
        <div ref={messagesEndRef} />
      </div>
    </div>
  );
}