'use client';

import { Check, Loader2, X } from 'lucide-react';
import { useEffect, useRef, useState } from 'react';

export type SetupItem = {
  id: string;
  label: string;
  status: 'pending' | 'progress' | 'complete' | 'failed';
};

export type SetupGroup = {
  title: string;
  items: SetupItem[];
};

type SetupProgressModalProps = {
  open: boolean;
  groups: SetupGroup[];
  onClose: () => void;
};

export function SetupProgressModal({ open, groups, onClose }: SetupProgressModalProps) {
  const [mounted, setMounted] = useState(false);
  const [visible, setVisible] = useState(false);
  const [done, setDone] = useState(false);
  const timerRef = useRef<ReturnType<typeof setTimeout>>();

  useEffect(() => {
    if (open) {
      setMounted(true);
      requestAnimationFrame(() => {
        requestAnimationFrame(() => {
          setVisible(true);
        });
      });
    } else {
      setVisible(false);
      const t = setTimeout(() => {
        setMounted(false);
        setDone(false);
      }, 350);
      timerRef.current = t;
    }
    return () => {
      if (timerRef.current) clearTimeout(timerRef.current);
    };
  }, [open]);

  const allTerminal = groups.length > 0 && groups.every(g => g.items.every(i => i.status === 'complete' || i.status === 'failed'));
  const anyFailed = groups.some(g => g.items.some(i => i.status === 'failed'));

  useEffect(() => {
    if (!open || !mounted) return;
    if (allTerminal) {
      setDone(true);
      const t = setTimeout(() => {
        onClose();
      }, 1000);
      timerRef.current = t;
      return () => clearTimeout(t);
    } else {
      setDone(false);
    }
  }, [groups, open, mounted, onClose, allTerminal]);

  if (!mounted) return null;

  const pendingCount = groups.reduce((s, g) => s + g.items.filter(i => i.status !== 'complete' && i.status !== 'failed').length, 0);

  return (
    <div className={`setup-progress-overlay${visible ? ' visible' : ''}`}>
      <div className={`setup-progress-panel${visible ? ' visible' : ''}`}>
        {done ? (
          <div className="setup-progress-done">
            <div className={`setup-progress-icon-wrap ${anyFailed ? 'failed' : ''}`}>
              {anyFailed ? <X size={28} /> : <Check size={28} />}
            </div>
            <span className={`setup-progress-done-label ${anyFailed ? 'failed' : ''}`}>
              {anyFailed ? 'Failed' : 'Complete'}
            </span>
          </div>
        ) : (
          <>
            <div className="setup-progress-header">
              <Loader2 size={16} className="setup-progress-spinner" />
              <span>
                {pendingCount > 0
                  ? `Setting up ${pendingCount} dependenc${pendingCount > 1 ? 'ies' : 'y'}...`
                  : 'Processing...'}
              </span>
            </div>
            <div className="setup-progress-groups">
              {groups.map((group) => (
                <div key={group.title} className="setup-progress-group">
                  <div className="setup-progress-group-title">{group.title}</div>
                  {group.items.map((item) => (
                    <div key={item.id} className={`setup-progress-item status-${item.status}`}>
                      <span className="setup-progress-icon">
                        {item.status === 'complete' ? (
                          <Check size={14} />
                        ) : item.status === 'failed' ? (
                          <X size={14} />
                        ) : item.status === 'progress' ? (
                          <span className="setup-progress-dot active" />
                        ) : (
                          <span className="setup-progress-dot" />
                        )}
                      </span>
                      <span className="setup-progress-label">{item.label}</span>
                      {item.status === 'progress' && (
                        <span className="setup-progress-bar">
                          <span className="setup-progress-bar-fill" />
                        </span>
                      )}
                    </div>
                  ))}
                </div>
              ))}
            </div>
          </>
        )}
      </div>
    </div>
  );
}
