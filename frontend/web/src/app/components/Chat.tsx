import { useState, useRef, useEffect } from 'react';
import { Send } from 'lucide-react';
import crewaiMascot from '../../images/crewai-mascot.png';
import { ChatProgressBar } from './ChatProgressBar';

interface ChatMessage {
  id: string;
  text: string;
  sender: 'boss' | 'user';
  timestamp: Date;
  read?: boolean;
}

interface ChatProps {
  messages: ChatMessage[];
  onSendMessage: (message: string) => void;
  onMarkAsRead: (messageId: string) => void;
}

export function Chat({ messages, onSendMessage, onMarkAsRead }: ChatProps) {
  const messagesEndRef = useRef<HTMLDivElement>(null);

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  useEffect(() => {
    // Mark messages as read when they come into view
    messages.forEach(message => {
      if (message.sender === 'boss' && !message.read) {
        onMarkAsRead(message.id);
      }
    });
  }, [messages, onMarkAsRead]);

  return (
    <div className="flex flex-col h-full bg-[var(--background)]">
      {/* Progress Bar */}
      <div className="p-4 pb-2">
        <ChatProgressBar />
      </div>

      {/* Messages area */}
      <div className="flex-1 overflow-y-auto px-4 space-y-4">
        {messages.length === 0 ? (
          <div className="flex items-center justify-center h-full">
            <div className="text-center">
              <img
                src={crewaiMascot}
                alt="CrewAI Mascot"
                className="w-28 h-28 rounded-lg object-contain mx-auto mb-6"
              />
              <h1 className="text-2xl font-bold text-[var(--text)] mb-2">
                Добро пожаловать в CrewAI!
              </h1>
              <p className="text-[var(--text-muted)]">
                Начните общение с вашим ИИ-ассистентом
              </p>
            </div>
          </div>
        ) : (
          messages.map((message) => (
            <div
              key={message.id}
              className={`flex ${message.sender === 'user' ? 'justify-end' : 'justify-start'}`}
            >
              <div
                className={`max-w-[70%] rounded-lg px-4 py-2 ${
                  message.sender === 'user'
                    ? 'bg-[var(--accent)] text-white'
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