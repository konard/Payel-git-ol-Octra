import { useEffect, useState } from 'react';
import { useTaskStore } from '../../stores/taskStore';

interface ChatProgressBarProps {
  progress?: number;
}

export function ChatProgressBar({ progress }: ChatProgressBarProps) {
  const taskStatus = useTaskStore((state) => state.status);
  const taskId = useTaskStore((state) => state.taskId);
  const [currentProgress, setCurrentProgress] = useState(progress || 0);

  // Get progress from task store nodes
  const nodes = useTaskStore((state) => state.nodes);
  const activeProgress = nodes.length > 0
    ? nodes.reduce((acc, node) => acc + (node.progress || 0), 0) / nodes.length
    : 0;

  useEffect(() => {
    if (nodes.length > 0) {
      setCurrentProgress(Math.round(activeProgress));
    } else if (progress !== undefined) {
      setCurrentProgress(progress);
    }
  }, [activeProgress, nodes.length, progress]);

  // Show progress bar if there's an active task OR if there are chat messages with progress
  const isTaskActive = taskId && ['creating', 'planning', 'executing'].includes(taskStatus);
  const hasProgress = currentProgress > 0;

  if (!isTaskActive && !hasProgress) {
    return null;
  }
  const bubbles = Array.from({ length: 15 }, (_, i) => ({
    id: i,
    delay: `${i * 0.6}s`,
    duration: `${3 + (i % 4)}s`,
    top: `${10 + (i % 5) * 16}%`, // Разные вертикальные позиции
    size: `${2 + (i % 2)}px`, // Разные размеры
  }));

  const tentacles = Array.from({ length: 5 }, (_, i) => ({
    id: i,
    left: `${15 + i * 18}%`,
    delay: `${i * 0.4}s`,
  }));

  return (
    <div className="w-full mb-4">
      <div className="mb-2">
        <div className="flex justify-between items-center mb-1">
          <span className="text-sm text-[var(--text-muted)]">
            {isTaskActive ? 'Прогресс выполнения' : 'Прогресс чата'}
          </span>
          <span className="text-sm text-[var(--accent)]">{currentProgress}%</span>
        </div>

        {/* Progress Bar Container */}
        <div className="relative w-full h-6 bg-[var(--surface)] border border-[var(--border)] rounded-lg overflow-hidden">
          {/* Progress Fill */}
          <div
            className="absolute top-0 left-0 h-full bg-[var(--accent)] rounded-lg transition-all duration-300 ease-out"
            style={{ width: `${currentProgress}%` }}
          >
            {/* Tentacles */}
            {tentacles.map((t) => (
              <div
                key={t.id}
                className="absolute tentacle"
                style={{
                  left: t.left,
                  top: '-1px',
                  '--delay': t.delay,
                } as React.CSSProperties}
              >
                <svg width="12" height="18" viewBox="0 0 12 18" className="tentacle-wave">
                  <path
                    d="M 6,0 Q 7,3 6.5,6 T 5,12 T 6,18"
                    fill="none"
                    stroke="rgba(255,255,255,0.5)"
                    strokeWidth="1.5"
                    strokeLinecap="round"
                  />
                </svg>
              </div>
            ))}
          </div>

          {/* Bubbles - теперь на уровне контейнера, двигаются через всю ширину */}
          {bubbles.map((b) => (
            <div
              key={b.id}
              className="absolute bubble rounded-full bg-white/60"
                style={{
                  left: '0px', // Начинают от левого края контейнера
                  top: b.top,
                  width: b.size,
                  height: b.size,
                  '--delay': b.delay,
                  '--duration': b.duration,
                } as React.CSSProperties}
            />
          ))}
        </div>
      </div>
    </div>
  );
}