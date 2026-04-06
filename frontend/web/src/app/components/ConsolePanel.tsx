import { useState, useEffect, useRef } from 'react';
import { useTaskStore } from '../../stores/taskStore';
import { ChevronDown, ChevronUp, Terminal, XCircle, CheckCircle, AlertTriangle, Info } from 'lucide-react';
import { t } from '../../hooks/useI18n';

export function ConsolePanel() {
  const logs = useTaskStore((state) => state.logs);
  const [isCollapsed, setIsCollapsed] = useState(true);
  const logsEndRef = useRef<HTMLDivElement>(null);
  const panelRef = useRef<HTMLDivElement>(null);

  // Auto-scroll to bottom when new logs arrive
  useEffect(() => {
    if (!isCollapsed && logsEndRef.current) {
      logsEndRef.current.scrollIntoView({ behavior: 'smooth' });
    }
  }, [logs, isCollapsed]);

  // Get last log message for preview when collapsed
  const lastLog = logs.length > 0 ? logs[logs.length - 1] : null;

  const getLogIcon = (type: string) => {
    switch (type) {
      case 'error':
        return <XCircle className="w-3.5 h-3.5 text-red-500 flex-shrink-0" />;
      case 'success':
        return <CheckCircle className="w-3.5 h-3.5 text-green-500 flex-shrink-0" />;
      case 'warning':
        return <AlertTriangle className="w-3.5 h-3.5 text-yellow-500 flex-shrink-0" />;
      default:
        return <Info className="w-3.5 h-3.5 text-blue-500 flex-shrink-0" />;
    }
  };

  const getLogColor = (type: string) => {
    switch (type) {
      case 'error':
        return 'text-red-500';
      case 'success':
        return 'text-green-500';
      case 'warning':
        return 'text-yellow-500';
      default:
        return 'text-[var(--text)]';
    }
  };

  const formatTime = (timestamp: Date | number) => {
    const date = timestamp instanceof Date ? timestamp : new Date(timestamp);
    return date.toLocaleTimeString('ru-RU', {
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
    });
  };

  return (
    <div ref={panelRef} className="border-t border-[var(--border)] bg-[var(--surface)] backdrop-blur-sm">
      {/* Header — always visible */}
      <button
        onClick={() => setIsCollapsed(!isCollapsed)}
        className="w-full flex items-center justify-between px-4 py-2 hover:bg-[var(--text-muted)]/10 transition-colors"
      >
        <div className="flex items-center gap-2">
          <Terminal className="w-4 h-4 text-[var(--text-muted)]" />
          <span className="text-sm font-medium text-[var(--text)]">
            {t('console.title')}
            <span className="ml-2 text-xs text-[var(--text-muted)]">({logs.length})</span>
          </span>
          {!isCollapsed || (
            lastLog && (
              <span className="text-xs text-[var(--text-muted)] truncate max-w-xs ml-2">
                {lastLog.message}
              </span>
            )
          )}
        </div>
        <div className="flex items-center gap-2">
          {lastLog && isCollapsed && (
            <span className={`text-xs ${getLogColor(lastLog.type)}`}>
              {getLogIcon(lastLog.type)}
            </span>
          )}
          {isCollapsed ? (
            <ChevronDown className="w-4 h-4 text-[var(--text-muted)]" />
          ) : (
            <ChevronUp className="w-4 h-4 text-[var(--text-muted)]" />
          )}
        </div>
      </button>

      {/* Logs — visible when expanded */}
      {!isCollapsed && (
        <div className="max-h-48 overflow-y-auto px-4 pb-3 font-mono text-xs space-y-0.5">
          {logs.length === 0 ? (
            <div className="text-[var(--text-muted)] py-4 text-center">
              {t('console.empty')}
            </div>
          ) : (
            logs.map((log) => (
              <div key={log.id} className="flex items-start gap-2 py-0.5">
                <span className="text-[var(--text-muted)] flex-shrink-0">
                  {formatTime(log.timestamp)}
                </span>
                {getLogIcon(log.type)}
                <span className={getLogColor(log.type)}>{log.message}</span>
              </div>
            ))
          )}
          <div ref={logsEndRef} />
        </div>
      )}
    </div>
  );
}
