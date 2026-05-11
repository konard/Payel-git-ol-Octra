import { useState, useRef, useEffect } from 'react';
import { Send } from 'lucide-react';
import crewaiMascot from '../../images/crewai-mascot.png';
import { useTaskStore } from '../../stores/taskStore';

interface ChatMessage {
  id: string;
  text: string;
  sender: 'boss' | 'user';
  timestamp: Date;
  read?: boolean;
  isClarification?: boolean;
  progress?: number; // Progress percentage for progress messages
  showProgress?: boolean; // Whether to show progress bar for this message
}

interface ChatProps {
  messages: ChatMessage[];
  onSendMessage: (message: string) => void;
  onMarkAsRead: (messageId: string) => void;
}

export function Chat({ messages, onSendMessage, onMarkAsRead }: ChatProps) {
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const nodes = useTaskStore((state) => state.nodes);
  const zipUrl = useTaskStore((state) => state.zipUrl);

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
            <div key={message.id} className="space-y-2">
              {/* User message */}
              {message.sender === 'user' && (
                <div className="flex justify-end">
                  <div className="bg-[var(--accent)] text-white rounded-lg px-4 py-2 max-w-[70%]">
                    <div className="text-sm">{message.text}</div>
                    <div className="text-xs mt-1 text-white/70">
                      {message.timestamp.toLocaleTimeString([], {hour: '2-digit', minute: '2-digit'})}
                    </div>
                  </div>
                </div>
              )}

              {/* Boss message with progress */}
              {message.sender === 'boss' && (
                <div className="flex justify-start space-y-2">
                  <div className="space-y-3 max-w-[70%]">
                    {/* CrewAI mascot and progress bar */}
                    <div className="flex items-center gap-4">
                      <img
                        src={crewaiMascot}
                        alt="CrewAI Mascot"
                        className="w-12 h-12 rounded-lg object-contain"
                      />
                      <div className="flex-1">
                        <div className="bg-[var(--surface)] border border-[var(--border)] rounded-lg p-2">
                          <div
                            className="w-full bg-[var(--accent)] rounded-lg h-2 transition-all duration-300"
                            style={{ width: `${message.progress || 0}%` }}
                          ></div>
                        </div>
                      </div>
                    </div>

                    {/* Team status */}
                    {(managers.length > 0 || workers.length > 0) && (
                      <div className="space-y-1">
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
                    )}

                    {/* Boss message text */}
                    {message.text && (
                      <div className={`rounded-lg px-4 py-2 ${
                        message.isClarification
                          ? 'bg-orange-100 dark:bg-orange-900 border border-orange-300 dark:border-orange-700 text-orange-900 dark:text-orange-100'
                          : 'bg-[var(--surface)] text-[var(--text)] border border-[var(--border)]'
                      }`}>
                        <div className="text-sm">{message.text}</div>
                        {message.progress === 100 && zipUrl && (
                          <div className="mt-2 pt-2 border-t border-[var(--border)]">
                            <a
                              href={zipUrl}
                              target="_blank"
                              rel="noopener noreferrer"
                              className="flex items-center gap-2 text-sm text-[var(--accent)] hover:underline"
                            >
                              <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                                <path d="M15 22v-4a4.8 4.8 0 0 0-1-3.5c3 0 6-2 6-5.5.08-1.25-.27-2.48-1-3.5.28-1.15.28-2.35 0-3.5 0 0-1 0-3 1.5-2.64-.5-5.36.5-8 0C6 2 5 2 5 2c-.3 1.15-.3 2.35 0 3.5A5.403 5.403 0 0 0 4 9c0 3.5 3 5.5 6 5.5-.39.49-.68 1.05-.85 1.65-.17.6-.22 1.23-.15 1.85v4"/>
                                <path d="M9 18c-4.51 2-5-2-7-2"/>
                              </svg>
                              {zipUrl}
                            </a>
                          </div>
                        )}
                        <div className="text-xs mt-1 text-[var(--text-muted)]">
                          {message.timestamp.toLocaleTimeString([], {hour: '2-digit', minute: '2-digit'})}
                        </div>
                      </div>
                    )}
                  </div>
                </div>
              )}
            </div>
          ))
        )}
        <div ref={messagesEndRef} />
      </div>
    </div>
  );
}