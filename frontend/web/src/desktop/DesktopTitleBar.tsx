import { useEffect, useRef, useState } from 'react';
import { Minus, Square, Copy, X, PanelLeft, FolderOpen, ChevronDown, Folder } from 'lucide-react';
import { getBridge, isDesktopApp } from './bridge';
import { useDesktopStore } from './desktopStore';

/**
 * Frameless-window title bar for the Electron desktop app.
 *
 * The previous desktop shell had an ugly, barely-functional window strip
 * (issue #50.5). This replaces it with a clean Octra-styled bar that is
 * draggable, exposes a working File menu and project switcher, and renders
 * crisp minimize / maximize / close controls. It renders nothing in a normal
 * browser, so the web build is unaffected.
 */
export function DesktopTitleBar() {
  if (!isDesktopApp()) return null;
  return <TitleBarInner />;
}

// The whole bar is a drag region except interactive controls, which opt out via
// the `no-drag` class (see the WebkitAppRegion styles below).
const DRAG_STYLE = { WebkitAppRegion: 'drag' } as React.CSSProperties;
const NO_DRAG_STYLE = { WebkitAppRegion: 'no-drag' } as React.CSSProperties;

function TitleBarInner() {
  const bridge = getBridge()!;
  const isMac = bridge.platform === 'darwin';
  const project = useDesktopStore((s) => s.project);
  const recent = useDesktopStore((s) => s.recent);
  const openFolder = useDesktopStore((s) => s.openFolder);
  const openProjectPath = useDesktopStore((s) => s.openProjectPath);
  const toggleExplorer = useDesktopStore((s) => s.toggleExplorer);
  const refreshRecent = useDesktopStore((s) => s.refreshRecent);

  const [maximized, setMaximized] = useState(false);
  const [menuOpen, setMenuOpen] = useState(false);
  const menuRef = useRef<HTMLDivElement>(null);

  // Keep the maximize/restore icon in sync with the real window state.
  useEffect(() => {
    let mounted = true;
    void bridge.window.isMaximized().then((m) => mounted && setMaximized(m));
    const off = bridge.window.onMaximizeChange((m) => mounted && setMaximized(m));
    return () => {
      mounted = false;
      off();
    };
  }, [bridge]);

  // Route native-menu / accelerator commands to the same actions as the bar.
  useEffect(() => {
    const off = bridge.onMenuCommand((command) => {
      switch (command) {
        case 'project.open':
        case 'project.new':
          void openFolder();
          break;
        case 'project.recent':
          void refreshRecent().then(() => setMenuOpen(true));
          break;
        case 'view.toggleExplorer':
          toggleExplorer();
          break;
        case 'view.reload':
          window.location.reload();
          break;
        case 'window.minimize':
          void bridge.window.minimize();
          break;
        case 'window.toggleMaximize':
          void bridge.window.toggleMaximize();
          break;
        case 'window.close':
        case 'app.quit':
          void bridge.window.close();
          break;
        case 'help.docs':
        case 'help.report':
          void bridge.shell.openExternal('https://github.com/Payel-git-ol/Octra');
          break;
        default:
          break;
      }
    });
    return off;
  }, [bridge, openFolder, refreshRecent, toggleExplorer]);

  // Close the dropdown on any outside click.
  useEffect(() => {
    if (!menuOpen) return;
    const handler = (e: MouseEvent) => {
      if (menuRef.current && !menuRef.current.contains(e.target as Node)) setMenuOpen(false);
    };
    document.addEventListener('mousedown', handler);
    return () => document.removeEventListener('mousedown', handler);
  }, [menuOpen]);

  return (
    <div
      className="flex h-9 shrink-0 items-center gap-1 border-b border-[var(--border)] bg-[var(--surface)] pr-2 text-[var(--text)] select-none"
      style={DRAG_STYLE}
    >
      {/* macOS keeps its OS traffic lights at top-left, so reserve space there. */}
      {isMac && <div className="w-[72px] shrink-0" />}

      {/* Toggle the file explorer dock. */}
      <button
        type="button"
        onClick={toggleExplorer}
        title="Toggle Explorer"
        className="ml-1 flex h-7 w-7 items-center justify-center rounded-md text-[var(--text-muted)] transition-colors hover:bg-[var(--surface-sunken)] hover:text-[var(--text)]"
        style={NO_DRAG_STYLE}
      >
        <PanelLeft size={16} />
      </button>

      {/* File menu + project switcher. */}
      <div className="relative" ref={menuRef} style={NO_DRAG_STYLE}>
        <button
          type="button"
          onClick={() => {
            void refreshRecent();
            setMenuOpen((v) => !v);
          }}
          className="flex h-7 items-center gap-1.5 rounded-md px-2 text-sm font-medium transition-colors hover:bg-[var(--surface-sunken)]"
        >
          <Folder size={14} className="text-[var(--accent)]" />
          <span className="max-w-[220px] truncate">{project ? project.name : 'No Project'}</span>
          <ChevronDown size={13} className="text-[var(--text-muted)]" />
        </button>

        {menuOpen && (
          <div className="absolute left-0 top-9 z-50 w-72 overflow-hidden rounded-lg border border-[var(--border)] bg-[var(--surface)] py-1 shadow-xl">
            <button
              type="button"
              onClick={() => {
                setMenuOpen(false);
                void openFolder();
              }}
              className="flex w-full items-center gap-2 px-3 py-2 text-left text-sm transition-colors hover:bg-[var(--surface-sunken)]"
            >
              <FolderOpen size={15} className="text-[var(--text-muted)]" />
              Open Folder…
            </button>
            <div className="my-1 h-px bg-[var(--border)]" />
            <div className="px-3 pb-1 pt-1 text-xs font-medium uppercase tracking-wide text-[var(--text-muted)]">
              Recent
            </div>
            {recent.length === 0 ? (
              <div className="px-3 py-2 text-sm text-[var(--text-muted)]">No recent projects</div>
            ) : (
              recent.map((p) => (
                <button
                  key={p.path}
                  type="button"
                  title={p.path}
                  onClick={() => {
                    setMenuOpen(false);
                    void openProjectPath(p.path);
                  }}
                  className="flex w-full flex-col items-start px-3 py-1.5 text-left transition-colors hover:bg-[var(--surface-sunken)]"
                >
                  <span className="text-sm">{p.name}</span>
                  <span className="max-w-full truncate text-xs text-[var(--text-muted)]">{p.path}</span>
                </button>
              ))
            )}
          </div>
        )}
      </div>

      {/* Centred app title; also the main draggable area. */}
      <div className="flex flex-1 items-center justify-center text-xs font-medium text-[var(--text-muted)]">
        Octra
      </div>

      {/* Windows / Linux window controls. macOS uses its native traffic lights. */}
      {!isMac && (
        <div className="flex items-center" style={NO_DRAG_STYLE}>
          <button
            type="button"
            onClick={() => void bridge.window.minimize()}
            title="Minimize"
            className="flex h-7 w-9 items-center justify-center rounded-md text-[var(--text-muted)] transition-colors hover:bg-[var(--surface-sunken)] hover:text-[var(--text)]"
          >
            <Minus size={15} />
          </button>
          <button
            type="button"
            onClick={() => void bridge.window.toggleMaximize()}
            title={maximized ? 'Restore' : 'Maximize'}
            className="flex h-7 w-9 items-center justify-center rounded-md text-[var(--text-muted)] transition-colors hover:bg-[var(--surface-sunken)] hover:text-[var(--text)]"
          >
            {maximized ? <Copy size={13} /> : <Square size={12} />}
          </button>
          <button
            type="button"
            onClick={() => void bridge.window.close()}
            title="Close"
            className="flex h-7 w-9 items-center justify-center rounded-md text-[var(--text-muted)] transition-colors hover:bg-[#e81123] hover:text-white"
          >
            <X size={16} />
          </button>
        </div>
      )}
    </div>
  );
}
