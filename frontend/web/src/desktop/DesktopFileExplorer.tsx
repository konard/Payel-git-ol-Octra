import { useMemo, useState } from 'react';
import {
  ChevronRight,
  ChevronDown,
  Folder,
  FolderOpen,
  RefreshCw,
  FolderPlus,
} from 'lucide-react';
import { isDesktopApp, type FileTreeNode } from './bridge';
import { useDesktopStore } from './desktopStore';
import { FileTypeIcon } from '../lib/FileTypeIcon';

/**
 * Filesystem-backed file explorer for the desktop app — the feature the old
 * shell was missing ("Чтение файлов нету … ничего не отобразилось", issue #50.4).
 * Reads the open project's tree through the Electron bridge and lets the user
 * open any file, which is shown in the web app's "Solution files" panel
 * (SolutionViewer). Renders nothing in a normal browser.
 */
export function DesktopFileExplorer() {
  if (!isDesktopApp()) return null;
  return <ExplorerInner />;
}

function ExplorerInner() {
  const project = useDesktopStore((s) => s.project);
  const tree = useDesktopStore((s) => s.tree);
  const loadingTree = useDesktopStore((s) => s.loadingTree);
  const error = useDesktopStore((s) => s.error);
  const openFolder = useDesktopStore((s) => s.openFolder);
  const reloadTree = useDesktopStore((s) => s.reloadTree);

  return (
    <div className="flex h-full min-h-0 flex-col bg-[var(--surface)] text-[var(--text)]">
      <div className="flex h-9 shrink-0 items-center justify-between border-b border-[var(--border)] px-3">
        <span className="truncate text-xs font-semibold uppercase tracking-wide text-[var(--text-muted)]">
          {project ? project.name : 'Explorer'}
        </span>
        <div className="flex items-center gap-0.5">
          <button
            type="button"
            onClick={() => void reloadTree()}
            title="Refresh"
            disabled={!project}
            className="flex h-6 w-6 items-center justify-center rounded text-[var(--text-muted)] transition-colors hover:bg-[var(--surface-sunken)] hover:text-[var(--text)] disabled:opacity-40"
          >
            <RefreshCw size={13} className={loadingTree ? 'animate-spin' : ''} />
          </button>
          <button
            type="button"
            onClick={() => void openFolder()}
            title="Open Folder"
            className="flex h-6 w-6 items-center justify-center rounded text-[var(--text-muted)] transition-colors hover:bg-[var(--surface-sunken)] hover:text-[var(--text)]"
          >
            <FolderPlus size={14} />
          </button>
        </div>
      </div>

      <div className="min-h-0 flex-1 overflow-auto py-1">
        {!project ? (
          <div className="flex h-full flex-col items-center justify-center gap-3 px-4 text-center">
            <Folder size={28} className="text-[var(--text-muted)]" />
            <p className="text-sm text-[var(--text-muted)]">No project open</p>
            <button
              type="button"
              onClick={() => void openFolder()}
              className="rounded-md bg-[var(--accent)] px-3 py-1.5 text-sm font-medium text-white transition-opacity hover:opacity-90"
            >
              Open Folder
            </button>
          </div>
        ) : error ? (
          <div className="px-3 py-2 text-sm text-red-400">{error}</div>
        ) : tree && tree.children ? (
          <TreeNodes nodes={tree.children} depth={0} />
        ) : (
          <div className="px-3 py-2 text-sm text-[var(--text-muted)]">Loading…</div>
        )}
      </div>
    </div>
  );
}

function TreeNodes({ nodes, depth }: { nodes: FileTreeNode[]; depth: number }) {
  return (
    <>
      {nodes.map((node) =>
        node.type === 'directory' ? (
          <DirectoryRow key={node.path} node={node} depth={depth} />
        ) : (
          <FileRow key={node.path} node={node} depth={depth} />
        ),
      )}
    </>
  );
}

function DirectoryRow({ node, depth }: { node: FileTreeNode; depth: number }) {
  // Top-level folders start expanded so the project is visible at a glance.
  const [open, setOpen] = useState(depth === 0);
  const children = node.children || [];
  return (
    <div>
      <button
        type="button"
        onClick={() => setOpen((v) => !v)}
        className="flex w-full items-center gap-1 py-1 pr-2 text-left text-sm transition-colors hover:bg-[var(--surface-sunken)]"
        style={{ paddingLeft: 8 + depth * 14 }}
      >
        {open ? (
          <ChevronDown size={14} className="shrink-0 text-[var(--text-muted)]" />
        ) : (
          <ChevronRight size={14} className="shrink-0 text-[var(--text-muted)]" />
        )}
        {open ? (
          <FolderOpen size={14} className="shrink-0 text-[var(--accent)]" />
        ) : (
          <Folder size={14} className="shrink-0 text-[var(--accent)]" />
        )}
        <span className="truncate">{node.name}</span>
      </button>
      {open && children.length > 0 && <TreeNodes nodes={children} depth={depth + 1} />}
    </div>
  );
}

function FileRow({ node, depth }: { node: FileTreeNode; depth: number }) {
  const openPath = useDesktopStore((s) => s.openPath);
  const activePath = useDesktopStore((s) => s.openFilePath);
  const isActive = activePath === node.path;
  return (
    <button
      type="button"
      onClick={() => void openPath(node.path)}
      title={node.path}
      className={`flex w-full items-center gap-1.5 py-1 pr-2 text-left text-sm transition-colors hover:bg-[var(--surface-sunken)] ${
        isActive ? 'bg-[color-mix(in_srgb,var(--accent)_16%,transparent)] text-[var(--text)]' : ''
      }`}
      style={{ paddingLeft: 8 + depth * 14 + 14 }}
    >
      <FileTypeIcon filePath={node.path} size={13} />
      <span className="truncate">{node.name}</span>
    </button>
  );
}

// True when the explorer should be shown in the workspace (desktop app only).
export const showDesktopExplorer = isDesktopApp;
