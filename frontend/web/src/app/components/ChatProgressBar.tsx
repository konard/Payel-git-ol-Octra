import { useState } from 'react';

interface ChatProgressBarProps {
  progress?: number;
}

export function ChatProgressBar({ progress = 45 }: ChatProgressBarProps) {
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
      <style>{`
        @keyframes bubble-flow {
          0% {
            transform: translateX(0);
            opacity: 0;
          }
          20% {
            opacity: 0.9;
          }
          80% {
            opacity: 0.9;
          }
          100% {
            transform: translateX(320px);
            opacity: 0;
          }
        }

        @keyframes tentacle-wave {
          0%, 100% {
            d: path('M 0,0 Q 2,3 4,6 T 8,12 T 12,18');
          }
          50% {
            d: path('M 0,0 Q -2,3 -4,6 T 0,12 T 4,18');
          }
        }

        @keyframes tentacle-wave {
          0%, 100% {
            transform: scaleY(1) translateX(0);
          }
          50% {
            transform: scaleY(1.2) translateX(2px);
          }
        }

        @keyframes tentacle-sway {
          0%, 100% {
            transform: translateX(0) rotate(0deg);
          }
          25% {
            transform: translateX(-2px) rotate(-3deg);
          }
          75% {
            transform: translateX(2px) rotate(3deg);
          }
        }

        .bubble {
          animation: bubble-flow var(--duration) ease-in-out infinite;
          animation-delay: var(--delay);
        }

        .tentacle {
          animation: tentacle-sway 2.5s ease-in-out infinite;
          animation-delay: var(--delay);
        }

        .tentacle-wave {
          animation: tentacle-wave 1.5s ease-in-out infinite;
          animation-delay: var(--delay);
        }
      `}</style>

      <div className="mb-2">
        <div className="flex justify-between items-center mb-1">
          <span className="text-sm text-[var(--text-muted)]">Прогресс выполнения</span>
          <span className="text-sm text-[var(--accent)]">{progress}%</span>
        </div>

        {/* Progress Bar Container */}
        <div className="relative w-full h-6 bg-[var(--surface)] border border-[var(--border)] rounded-lg overflow-hidden">
          {/* Progress Fill */}
          <div
            className="absolute top-0 left-0 h-full bg-[var(--accent)] rounded-lg transition-all duration-300 ease-out"
            style={{ width: `${progress}%` }}
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