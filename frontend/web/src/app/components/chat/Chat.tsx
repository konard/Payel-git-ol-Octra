import { useEffect, useRef, useState } from 'react';
import { CheckCircle2, ChevronDown, ChevronRight, Circle, CircleDotDashed, Download, GitBranch, Globe, Loader2, UserRound } from 'lucide-react';
import octraMascot from '../../../images/octra-mascot.png';
import { useTaskStore } from '../../../stores/taskStore';

export interface ChatMessage {
  id: string;
  text: string;
  sender: 'boss' | 'user';
  timestamp: Date;
  read?: boolean;
  isClarification?: boolean;
  progress?: number;
  showProgress?: boolean;
}

interface ChatProps {
  messages: ChatMessage[];
  onMarkAsRead: (messageId: string) => void;
}

function formatTime(date: Date) {
  return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
}

export function Chat({ messages, onMarkAsRead }: ChatProps) {
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const nodes = useTaskStore((state) => state.nodes);
  const zipUrl = useTaskStore((state) => state.zipUrl);
  const status = useTaskStore((state) => state.status);
  const searchSteps = useTaskStore((state) => state.searchSteps);
  const searchPhase = useTaskStore((state) => state.searchPhase);
  const searchStepsCount = useTaskStore((state) => state.searchStepsCount);
  const managers = nodes.filter((node) => node.type === 'manager');
  const workers = nodes.filter((node) => node.type === 'worker');
  const universalNodes = nodes.filter((node) => node.type === 'universal');
  const [searchCollapsed, setSearchCollapsed] = useState(false);
  const isSearching = searchPhase === 'searching';
  const completedSteps = searchStepsCount || searchSteps.length;

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages, nodes, searchSteps]);

  useEffect(() => {
    const timer = setTimeout(() => {
      messages.forEach((message) => {
        if (message.sender === 'boss' && !message.read) {
          onMarkAsRead(message.id);
        }
      });
    }, 120);
    return () => clearTimeout(timer);
  }, [messages, onMarkAsRead]);

  return (
    <div className="flex h-full min-h-0 flex-col bg-[var(--background)]">
      <div className="flex min-h-0 flex-1 flex-col overflow-y-auto px-4 py-5">
        <div className="mx-auto flex w-full max-w-4xl flex-1 flex-col gap-5">
          {messages.length === 0 && nodes.length === 0 && searchSteps.length === 0 ? (
            <div className="m-auto text-center">
              <img
                src={octraMascot}
                alt="Octra Mascot"
                className="mx-auto mb-4 h-28 w-28 object-contain drop-shadow"
              />
              <p className="text-sm text-[var(--text-muted)]">Ready.</p>
            </div>
          ) : (
            messages.map((message) => (
              <div key={message.id} className={`flex ${message.sender === 'user' ? 'justify-end' : 'justify-start'}`}>
                <div className={`flex max-w-[82%] gap-3 ${message.sender === 'user' ? 'flex-row-reverse' : ''}`}>
                  <div className={`flex h-9 w-9 shrink-0 items-center justify-center rounded-lg border border-[var(--border)] ${
                    message.sender === 'user' ? 'bg-[var(--accent)] text-white' : 'bg-[var(--surface)] text-[var(--text)]'
                  }`}>
                    {message.sender === 'user' ? (
                      <UserRound size={18} />
                    ) : (
                      <img src={octraMascot} alt="Octra Mascot" className="h-7 w-7 object-contain" />
                    )}
                  </div>

                  <div className={`min-w-0 rounded-lg border px-4 py-3 ${
                    message.sender === 'user'
                      ? 'border-transparent bg-[var(--accent)] text-white'
                      : message.isClarification
                        ? 'border-[var(--warning)]/40 bg-[var(--warning)]/10 text-[var(--text)]'
                        : 'border-[var(--border)] bg-[var(--surface)] text-[var(--text)]'
                  }`}>
                    <div className="whitespace-pre-wrap text-sm leading-6">{message.text}</div>

                    {message.showProgress && (
                      <div className="mt-3">
                        <div className="mb-1 flex items-center justify-between text-xs text-[var(--text-muted)]">
                          <span>{status === 'done' ? 'Workflow complete' : 'Workflow running'}</span>
                          <span>{message.progress ?? 0}%</span>
                        </div>
                        <div className="h-2 overflow-hidden rounded-full bg-[var(--background)]">
                          <div
                            className="h-full rounded-full bg-[var(--accent)] transition-all duration-300"
                            style={{ width: `${Math.max(0, Math.min(100, message.progress ?? 0))}%` }}
                          />
                        </div>
                      </div>
                    )}

                    {message.progress === 100 && zipUrl && (
                      <a
                        href={zipUrl}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="mt-3 inline-flex items-center gap-2 text-sm font-medium text-[var(--accent)] hover:underline"
                      >
                        <Download size={15} />
                        Open result
                      </a>
                    )}

                    <div className={`mt-2 text-xs ${
                      message.sender === 'user' ? 'text-white/70' : 'text-[var(--text-muted)]'
                    }`}>
                      {formatTime(message.timestamp)}
                    </div>
                  </div>
                </div>
              </div>
            ))
          )}

          {searchSteps.length > 0 && (
            <div className="ml-12 max-w-2xl">
              <button
                type="button"
                onClick={() => setSearchCollapsed((v) => !v)}
                className="flex w-full items-center gap-2 rounded-lg border border-[var(--border)] bg-[var(--surface)] px-3 py-2 text-sm font-medium text-[var(--text)] transition-colors hover:bg-[var(--background)]"
              >
                <Globe size={16} className="shrink-0 text-[var(--accent)]" />
                {isSearching ? (
                  <>
                    <span>Searching the web</span>
                    <Loader2 size={14} className="animate-spin text-[var(--accent)]" />
                  </>
                ) : (
                  <span>{`Completed ${completedSteps} step${completedSteps === 1 ? '' : 's'}`}</span>
                )}
                {searchCollapsed ? (
                  <ChevronRight size={16} className="ml-auto text-[var(--text-muted)]" />
                ) : (
                  <ChevronDown size={16} className="ml-auto text-[var(--text-muted)]" />
                )}
              </button>
              {!searchCollapsed && (
                <div className="mt-2 space-y-2 border-l border-[var(--border)] pl-5">
                  {searchSteps.map((step) => (
                    <div key={step.id} className="flex items-start gap-2 text-sm text-[var(--text-muted)]">
                      {isSearching ? (
                        <Loader2 size={14} className="mt-0.5 shrink-0 animate-spin text-[var(--accent)]" />
                      ) : (
                        <CheckCircle2 size={14} className="mt-0.5 shrink-0 text-[var(--success)]" />
                      )}
                      <span className="flex-1">
                        {step.text}
                        {step.resultCount && step.resultCount > 0 && (
                          <span className="ml-2 rounded bg-[var(--accent)]/10 px-1.5 py-0.5 text-xs text-[var(--accent)]">
                            {step.resultCount} sources
                          </span>
                        )}
                      </span>
                      {step.provider && (
                        <span className="shrink-0 text-xs text-[var(--text-muted)]">
                          {step.provider}{step.model ? ` · ${step.model}` : ''}
                        </span>
                      )}
                    </div>
                  ))}
                </div>
              )}
            </div>
          )}

          {(managers.length > 0 || workers.length > 0 || universalNodes.length > 0) && (
            <div className="ml-12 max-w-2xl border-l border-[var(--border)] pl-5">
              <div className="mb-3 flex items-center gap-2 text-sm font-semibold text-[var(--text)]">
                <GitBranch size={16} />
                Workflow activity
              </div>
              <div className="space-y-3">
                {universalNodes.map((node) => (
                  <div key={node.id} className="flex items-center gap-3 text-sm">
                    {node.status === 'done' ? (
                      <CheckCircle2 size={16} className="text-[var(--success)]" />
                    ) : node.status === 'working' || node.status === 'reviewing' ? (
                      <CircleDotDashed size={16} className="animate-spin text-[var(--accent)]" />
                    ) : (
                      <Circle size={16} className="text-[var(--text-muted)]" />
                    )}
                    <span className="text-[var(--text)]">{node.role}</span>
                    <span className="text-xs text-[var(--text-muted)]">
                      {node.filesCount ? `${node.filesCount} files` : node.status}
                    </span>
                  </div>
                ))}
                {managers.map((manager) => (
                  <div key={manager.id} className="flex items-center gap-3 text-sm">
                    {manager.status === 'done' ? (
                      <CheckCircle2 size={16} className="text-[var(--success)]" />
                    ) : manager.status === 'working' || manager.status === 'reviewing' ? (
                      <CircleDotDashed size={16} className="animate-spin text-[var(--accent)]" />
                    ) : (
                      <Circle size={16} className="text-[var(--text-muted)]" />
                    )}
                    <span className="text-[var(--text)]">{manager.role}</span>
                    <span className="text-xs text-[var(--text-muted)]">{manager.status}</span>
                  </div>
                ))}
                {workers.map((worker) => (
                  <div key={worker.id} className="ml-6 flex items-center gap-3 text-sm">
                    {worker.status === 'done' ? (
                      <CheckCircle2 size={16} className="text-[var(--success)]" />
                    ) : worker.status === 'working' ? (
                      <CircleDotDashed size={16} className="animate-spin text-[var(--accent)]" />
                    ) : (
                      <Circle size={16} className="text-[var(--text-muted)]" />
                    )}
                    <span className="text-[var(--text)]">{worker.role}</span>
                    <span className="text-xs text-[var(--text-muted)]">
                      {worker.filesCount ? `${worker.filesCount} files` : worker.status}
                    </span>
                  </div>
                ))}
              </div>
            </div>
          )}
          <div ref={messagesEndRef} />
        </div>
      </div>
    </div>
  );
}
