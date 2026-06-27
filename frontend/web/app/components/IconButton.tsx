'use client';

import { type ReactNode } from 'react';

interface IconButtonProps {
  children: ReactNode;
  onClick: (e: React.MouseEvent<HTMLButtonElement>) => void;
  'aria-label': string;
  variant?: 'default' | 'danger' | 'warning' | 'success';
  style?: React.CSSProperties;
}

export function IconButton({ children, onClick, 'aria-label': ariaLabel, variant = 'default', style }: IconButtonProps) {
  const cls = variant === 'danger' ? ' icon-button-danger' : variant === 'warning' ? ' icon-button-warning' : variant === 'success' ? ' icon-button-success' : '';
  return (
    <button
      className={`icon-button${cls}`}
      onClick={onClick}
      aria-label={ariaLabel}
      style={{ color: 'var(--muted)', flex: '0 0 auto', ...style }}
    >
      {children}
    </button>
  );
}
