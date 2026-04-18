import { useState } from 'react';

export default function App() {
  const [progress, setProgress] = useState(45);

  const increment = () => {
    setProgress(prev => Math.min(prev + 10, 100));
  };

  const decrement = () => {
    setProgress(prev => Math.max(prev - 10, 0));
  };

  const bubbles = Array.from({ length: 8 }, (_, i) => ({
    id: i,
    left: `${10 + i * 12}%`,
    delay: `${i * 0.7}s`,
    duration: `${3 + (i % 3)}s`,
  }));

  const tentacles = Array.from({ length: 5 }, (_, i) => ({
    id: i,
    left: `${15 + i * 18}%`,
    delay: `${i * 0.4}s`,
  }));

  return (
    <div className="size-full flex items-center justify-center bg-[#1e1e1e]">
      <style>{`
        @keyframes bubble-rise {
          0% {
            transform: translateY(40px) translateX(0);
            opacity: 0;
          }
          10% {
            opacity: 0.6;
          }
          90% {
            opacity: 0.6;
          }
          100% {
            transform: translateY(-10px) translateX(10px);
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

        @keyframes tentacle-sway {
          0%, 100% {
            transform: translateX(0) rotate(0deg);
          }
          25% {
            transform: translateX(-3px) rotate(-5deg);
          }
          75% {
            transform: translateX(3px) rotate(5deg);
          }
        }

        .bubble {
          animation: bubble-rise var(--duration) ease-in-out infinite;
          animation-delay: var(--delay);
        }

        .tentacle {
          animation: tentacle-sway 3s ease-in-out infinite;
          animation-delay: var(--delay);
        }
      `}</style>

      <div className="w-full max-w-2xl px-8">
        <div className="mb-8">
          <div className="flex justify-between items-center mb-3">
            <h2 className="text-xl text-gray-300">Прогресс</h2>
            <span className="text-2xl text-orange-500">{progress}%</span>
          </div>

          {/* Progress Bar Container */}
          <div className="relative w-full h-10 bg-[#2d2d2d] rounded-sm overflow-hidden">
            {/* Progress Fill */}
            <div
              className="absolute top-0 left-0 h-full bg-orange-500 rounded-sm transition-all duration-300 ease-out overflow-hidden"
              style={{ width: `${progress}%` }}
            >
              {/* Tentacles */}
              {tentacles.map((t) => (
                <div
                  key={t.id}
                  className="absolute tentacle"
                  style={{
                    left: t.left,
                    top: '-2px',
                    '--delay': t.delay,
                  } as React.CSSProperties}
                >
                  <svg width="16" height="24" viewBox="0 0 16 24">
                    <path
                      d="M 8,0 Q 10,4 9,8 T 7,16 T 8,24"
                      fill="none"
                      stroke="rgba(255,255,255,0.3)"
                      strokeWidth="2"
                      strokeLinecap="round"
                    >
                      <animate
                        attributeName="d"
                        values="M 8,0 Q 10,4 9,8 T 7,16 T 8,24;
                                M 8,0 Q 6,4 7,8 T 9,16 T 8,24;
                                M 8,0 Q 10,4 9,8 T 7,16 T 8,24"
                        dur="2s"
                        repeatCount="indefinite"
                      />
                    </path>
                  </svg>
                </div>
              ))}

              {/* Bubbles */}
              {bubbles.map((b) => (
                <div
                  key={b.id}
                  className="absolute bubble rounded-full bg-white/20"
                  style={{
                    left: b.left,
                    bottom: '0',
                    width: '6px',
                    height: '6px',
                    '--delay': b.delay,
                    '--duration': b.duration,
                  } as React.CSSProperties}
                />
              ))}
            </div>
          </div>
        </div>

        {/* Controls */}
        <div className="flex gap-4 justify-center">
          <button
            onClick={decrement}
            className="px-6 py-2 bg-[#2d2d2d] hover:bg-[#3d3d3d] text-gray-300 rounded transition-colors duration-150"
          >
            - 10%
          </button>
          <button
            onClick={increment}
            className="px-6 py-2 bg-orange-500 hover:bg-orange-600 text-white rounded transition-colors duration-150"
          >
            + 10%
          </button>
        </div>
      </div>
    </div>
  );
}