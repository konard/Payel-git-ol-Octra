'use client';

import { Gift } from 'lucide-react';

const CONFETTI_COLORS = [
  '#4ade80', '#22c55e', '#16a34a', '#15803d',
  '#facc15', '#fb923c', '#c084fc', '#67e8f9',
];

function createConfetti() {
  const container = document.querySelector('.confetti-wrap');
  if (!container) return;

  for (let i = 0; i < 120; i++) {
    const el = document.createElement('div');
    el.className = 'welcome-confetti-piece';
    const color = CONFETTI_COLORS[Math.floor(Math.random() * CONFETTI_COLORS.length)];
    const left = Math.random() * 100;
    const delay = Math.random() * 2;
    const duration = 1.5 + Math.random() * 2.5;
    const size = 4 + Math.random() * 6;
    const rotation = Math.random() * 360;
    el.style.cssText = `
      left: ${left}%;
      width: ${size}px;
      height: ${size * (0.4 + Math.random() * 0.6)}px;
      background: ${color};
      animation-delay: ${delay}s;
      animation-duration: ${duration}s;
      --rotation: ${rotation}deg;
      border-radius: ${Math.random() > 0.5 ? '50%' : '2px'};
    `;
    container.appendChild(el);
  }
}

interface Props {
  username: string;
  onClose: () => void;
}

export function WelcomeModal({ username, onClose }: Props) {
  function handleCelebrate() {
    const btn = document.querySelector('.welcome-button') as HTMLButtonElement | null;
    if (btn?.disabled) return;
    if (btn) btn.disabled = true;

    createConfetti();

    setTimeout(() => {
      onClose();
    }, 3000);
  }

  return (
    <div className="welcome-overlay">
      <div className="confetti-wrap" />
      <div className="welcome-dialog">
        <div className="welcome-icon-ring">
          <Gift size={32} />
        </div>
        <h1 className="welcome-title">Welcome, {username}!</h1>
        <p className="welcome-subtitle">Your account is ready.</p>
        <div className="welcome-bonus">
          <span>We&rsquo;ve added <strong>100 free credits</strong> to your account as a welcome gift.</span>
        </div>
        <button
          className="primary-button welcome-button"
          onClick={handleCelebrate}
        >
          <span>Start building</span>
        </button>
      </div>
    </div>
  );
}
