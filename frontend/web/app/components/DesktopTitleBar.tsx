'use client';

import { Maximize2, Minimize2, Square, X } from 'lucide-react';
import { useEffect, useState } from 'react';

type OctraDesktopBridge = {
  isElectron: true;
  window: {
    minimize: () => Promise<void>;
    toggleMaximize: () => Promise<boolean>;
    close: () => Promise<void>;
    isMaximized: () => Promise<boolean>;
    onMaximizeChange: (callback: (maximized: boolean) => void) => () => void;
  };
};

declare global {
  interface Window {
    octra?: OctraDesktopBridge;
  }
}

export function DesktopTitleBar() {
  const [isDesktop, setIsDesktop] = useState(false);
  const [isMaximized, setIsMaximized] = useState(false);

  useEffect(() => {
    const bridge = window.octra;
    if (!bridge?.isElectron) return;

    setIsDesktop(true);
    bridge.window.isMaximized().then(setIsMaximized).catch(() => setIsMaximized(false));
    return bridge.window.onMaximizeChange(setIsMaximized);
  }, []);

  if (!isDesktop || !window.octra) return null;

  return (
    <div className="desktop-titlebar">
      <div className="desktop-titlebar-title">
        <img src="/assets/octra-node-logo.svg" alt="" />
        <span>Octra</span>
      </div>
      <div className="desktop-window-controls" aria-label="Window controls">
        <button type="button" onClick={() => window.octra?.window.minimize()} aria-label="Minimize window">
          <Minimize2 size={14} />
        </button>
        <button type="button" onClick={() => window.octra?.window.toggleMaximize()} aria-label="Toggle maximize window">
          {isMaximized ? <Square size={13} /> : <Maximize2 size={14} />}
        </button>
        <button className="close-control" type="button" onClick={() => window.octra?.window.close()} aria-label="Close window">
          <X size={15} />
        </button>
      </div>
    </div>
  );
}
