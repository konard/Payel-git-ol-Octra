import Editor from '@monaco-editor/react';
import { X } from 'lucide-react';
import { isDesktopApp } from './bridge';
import { useDesktopStore } from './desktopStore';
import { useThemeStore } from '../stores/themeStore';

/**
 * Read-only viewer for a file opened from the DesktopFileExplorer. Rendered as a
 * layer over the workspace so files get the full editor width. Uses the same
 * Monaco editor the web SolutionViewer uses, so opened project files render with
 * proper syntax highlighting (issue #50.4). Renders nothing in a browser or when
 * no file is open.
 */
export function DesktopFileViewer() {
  if (!isDesktopApp()) return null;
  return <ViewerInner />;
}

function ViewerInner() {
  const openFile = useDesktopStore((s) => s.openFile);
  const loadingFile = useDesktopStore((s) => s.loadingFile);
  const isDark = useThemeStore((s) => s.isDark);

  if (!openFile) return null;

  return (
    <div className="absolute inset-2 z-20 flex flex-col overflow-hidden rounded-xl border border-[var(--border)] bg-[var(--surface)] shadow-2xl">
      <div className="flex h-9 shrink-0 items-center justify-between border-b border-[var(--border)] bg-[var(--surface)] px-3">
        <span className="truncate text-sm text-[var(--text)]" title={openFile.path}>
          {openFile.path}
          {openFile.truncated && (
            <span className="ml-2 text-xs text-[var(--text-muted)]">(truncated)</span>
          )}
        </span>
        <button
          type="button"
          onClick={() => useDesktopStore.setState({ openFile: null })}
          title="Close"
          className="flex h-6 w-6 items-center justify-center rounded text-[var(--text-muted)] transition-colors hover:bg-[var(--surface-sunken)] hover:text-[var(--text)]"
        >
          <X size={15} />
        </button>
      </div>
      <div className="min-h-0 flex-1">
        {loadingFile ? (
          <div className="flex h-full items-center justify-center text-sm text-[var(--text-muted)]">
            Loading…
          </div>
        ) : (
          <Editor
            path={openFile.path}
            value={openFile.content}
            language={openFile.language}
            theme={isDark ? 'vs-dark' : 'light'}
            options={{
              readOnly: true,
              automaticLayout: true,
              fontFamily: 'JetBrains Mono, ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace',
              fontSize: 13,
              lineHeight: 22,
              minimap: { enabled: true },
              padding: { top: 12, bottom: 12 },
              scrollBeyondLastLine: false,
              smoothScrolling: true,
            }}
          />
        )}
      </div>
    </div>
  );
}
