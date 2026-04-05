import { useEffect, useState } from 'react';
import { Wifi, WifiOff, Clock, Coins } from 'lucide-react';
import { useTaskStore } from '../../stores/taskStore';

export function StatusBar() {
  const isConnected = useTaskStore((state) => state.isConnected);
  const tokensUsed = useTaskStore((state) => state.tokensUsed);
  const status = useTaskStore((state) => state.status);
  const startTime = useTaskStore((state) => state.startTime);
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

  const statusLabels: Record<string, string> = {
    idle: 'Ожидание',
    creating: 'Создание задачи',
    planning: 'Планирование',
    executing: 'Выполнение',
    done: 'Завершено',
    error: 'Ошибка',
  };

  return (
    <footer className="bg-[var(--surface)] border-t border-[var(--border)] px-4 py-2 flex items-center justify-between text-xs text-[var(--text-muted)]">
      <div className="flex items-center gap-4">
        {/* Connection status */}
        <div className="flex items-center gap-1.5">
          {isConnected ? (
            <>
              <Wifi size={14} className="text-green-500" />
              <span className="text-green-500">Подключено</span>
            </>
          ) : (
            <>
              <WifiOff size={14} className="text-gray-500" />
              <span>Отключено</span>
            </>
          )}
        </div>

        {/* Task status */}
        <div className="flex items-center gap-1.5">
          <span>Статус:</span>
          <span className="text-[var(--text)] font-medium">{statusLabels[status] || status}</span>
        </div>
      </div>

      <div className="flex items-center gap-4">
        {/* Tokens used */}
        {tokensUsed > 0 && (
          <div className="flex items-center gap-1.5">
            <Coins size={14} />
            <span>Токенов: {tokensUsed.toLocaleString()}</span>
          </div>
        )}

        {/* Elapsed time */}
        {startTime && (
          <div className="flex items-center gap-1.5">
            <Clock size={14} />
            <span>Время: {elapsed}</span>
          </div>
        )}
      </div>
    </footer>
  );
}
