import { useEffect, useState } from 'react';
import { Wifi, WifiOff, Clock, Coins } from 'lucide-react';
import { useTaskStore } from '../../stores/taskStore';
import { useSettingsStore } from '../../stores/settingsStore';
import { t } from '../../hooks/useI18n';

export function StatusBar() {
  const isConnected = useTaskStore((state) => state.isConnected);
  const tokensUsed = useTaskStore((state) => state.tokensUsed);
  const status = useTaskStore((state) => state.status);
  const startTime = useTaskStore((state) => state.startTime);
  const hideServerStatus = useSettingsStore((state) => state.hideServerStatus);
  const [elapsed, setElapsed] = useState('0m 0s');

  useEffect(() => {
    if (!startTime) {
      setElapsed('0m 0s');
      return;
    }

    const interval = setInterval(() => {
      const diff = Math.floor((Date.now() - startTime) / 1000);
      const minutes = Math.floor(diff / 60);
      const seconds = diff % 60;
      setElapsed(`${minutes}m ${seconds}s`);
    }, 1000);

    return () => clearInterval(interval);
  }, [startTime]);

  if (hideServerStatus) return null;

  const statusLabels: Record<string, string> = {
    idle: t('statusbar.waiting'),
    creating: t('statusbar.creating'),
    planning: t('statusbar.planning'),
    executing: t('statusbar.executing'),
    done: t('statusbar.done'),
    error: t('statusbar.error'),
  };

  return (
    <footer className="bg-[var(--surface)] border-t border-[var(--border)] px-4 py-2 flex items-center justify-between text-xs text-[var(--text-muted)]">
      <div className="flex items-center gap-4">
        {/* Connection status */}
        <div className="flex items-center gap-1.5">
          {isConnected ? (
            <>
              <Wifi size={14} className="text-green-500" />
              <span className="text-green-500">{t('statusbar.connected')}</span>
            </>
          ) : (
            <>
              <WifiOff size={14} className="text-gray-500" />
              <span>{t('statusbar.disconnected')}</span>
            </>
          )}
        </div>

        {/* Task status */}
        <div className="flex items-center gap-1.5">
          <span>{t('statusbar.status')}:</span>
          <span className="text-[var(--text)] font-medium">{statusLabels[status] || status}</span>
        </div>
      </div>

      <div className="flex items-center gap-4">
        {/* Tokens used */}
        {tokensUsed > 0 && (
          <div className="flex items-center gap-1.5">
            <Coins size={14} />
            <span>{t('statusbar.tokens')}: {tokensUsed.toLocaleString()}</span>
          </div>
        )}

        {/* Elapsed time */}
        {startTime && (
          <div className="flex items-center gap-1.5">
            <Clock size={14} />
            <span>{t('statusbar.time')}: {elapsed}</span>
          </div>
        )}
      </div>
    </footer>
  );
}
