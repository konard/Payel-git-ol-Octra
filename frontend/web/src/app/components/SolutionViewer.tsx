import { useEffect, useMemo, useRef, useState } from 'react';
import type { CSSProperties } from 'react';
import Editor, { type OnMount } from '@monaco-editor/react';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import {
  CheckCircle2,
  ChevronDown,
  ChevronRight,
  CircleDotDashed,
  Code2,
  Eye,
  FileCode2,
  FileText,
  Files,
  FileBox,
  Folder,
  FolderOpen,
  PanelLeftClose,
  PanelLeftOpen,
  Presentation,
  X,
} from 'lucide-react';
import { useTaskStore, type CodeFile } from '../../stores/taskStore';
import { useThemeStore } from '../../stores/themeStore';
import { isMarkdownPath, isBinaryPath, binaryFileLabel } from '../../lib/markdown';
import '../../styles/markdown.css';

interface TreeNode {
  name: string;
  path: string;
  type: 'folder' | 'file';
  children: TreeNode[];
  file?: CodeFile;
}

const CODE_VIEW_THEME = {
  '--code-bg': 'var(--background)',
  '--code-surface': 'var(--surface)',
  '--code-border': 'var(--border)',
  '--code-text': 'var(--text)',
  '--code-text-muted': 'var(--text-muted)',
  '--code-accent': 'var(--accent)',
  '--code-line-active': 'color-mix(in srgb, var(--accent) 10%, transparent)',
  '--code-tree-active': 'color-mix(in srgb, var(--accent) 16%, transparent)',
  '--code-tree-hover': 'color-mix(in srgb, var(--text) 8%, transparent)',
} as CSSProperties & Record<string, string>;

function buildFileTree(files: CodeFile[]): TreeNode[] {
  const root: TreeNode[] = [];
  const folders = new Map<string, TreeNode>();

  files.forEach((file) => {
    const parts = file.path.split('/').filter(Boolean);
    let current = root;
    let currentPath = '';

    parts.forEach((part, index) => {
      currentPath = currentPath ? `${currentPath}/${part}` : part;
      const isFile = index === parts.length - 1;

      if (isFile) {
        current.push({
          name: part,
          path: file.path,
          type: 'file',
          children: [],
          file,
        });
        return;
      }

      let folder = folders.get(currentPath);
      if (!folder) {
        folder = {
          name: part,
          path: currentPath,
          type: 'folder',
          children: [],
        };
        folders.set(currentPath, folder);
        current.push(folder);
      }
      current = folder.children;
    });
  });

  const sortNodes = (nodes: TreeNode[]) => {
    nodes.sort((a, b) => {
      if (a.type !== b.type) return a.type === 'folder' ? -1 : 1;
      return a.name.localeCompare(b.name);
    });
    nodes.forEach((node) => sortNodes(node.children));
  };

  sortNodes(root);
  return root;
}

function statusLabel(status: CodeFile['status']) {
  return status === 'streaming' ? 'Generating' : 'Ready';
}

function CodeStatus({ status }: { status: CodeFile['status'] }) {
  if (status === 'streaming') {
    return <CircleDotDashed size={14} className="animate-spin text-[var(--code-accent)]" />;
  }
  return <CheckCircle2 size={14} className="text-[var(--success)]" />;
}

function TreeRows({
  nodes,
  depth,
  activePath,
  expandedFolders,
  onToggleFolder,
  onOpenFile,
}: {
  nodes: TreeNode[];
  depth: number;
  activePath: string | null;
  expandedFolders: Set<string>;
  onToggleFolder: (path: string) => void;
  onOpenFile: (path: string) => void;
}) {
  return (
    <>
      {nodes.map((node) => {
        const isExpanded = expandedFolders.has(node.path);
        const paddingLeft = 10 + depth * 16;

        if (node.type === 'folder') {
          return (
            <div key={node.path}>
              <button
                type="button"
                onClick={() => onToggleFolder(node.path)}
                className="flex h-8 w-full items-center gap-1.5 rounded-md pr-2 text-left text-sm text-[var(--code-text-muted)] transition-colors hover:bg-[var(--code-tree-hover)] hover:text-[var(--code-text)]"
                style={{ paddingLeft }}
              >
                {isExpanded ? <ChevronDown size={14} /> : <ChevronRight size={14} />}
                {isExpanded ? <FolderOpen size={15} /> : <Folder size={15} />}
                <span className="truncate">{node.name}</span>
              </button>
              {isExpanded && (
                <TreeRows
                  nodes={node.children}
                  depth={depth + 1}
                  activePath={activePath}
                  expandedFolders={expandedFolders}
                  onToggleFolder={onToggleFolder}
                  onOpenFile={onOpenFile}
                />
              )}
            </div>
          );
        }

        return (
          <button
            key={node.path}
            type="button"
            onClick={() => onOpenFile(node.path)}
            className={`flex h-8 w-full items-center gap-2 rounded-md pr-2 text-left text-sm transition-colors ${
              activePath === node.path
                ? 'bg-[var(--code-tree-active)] text-[var(--code-text)]'
                : 'text-[var(--code-text-muted)] hover:bg-[var(--code-tree-hover)] hover:text-[var(--code-text)]'
            }`}
            style={{ paddingLeft }}
          >
            {node.path.toLowerCase().endsWith('.pptx') ? (
              <Presentation size={15} />
            ) : isBinaryPath(node.path) ? (
              <FileBox size={15} />
            ) : isMarkdownPath(node.path) ? (
              <FileText size={15} />
            ) : (
              <FileCode2 size={15} />
            )}
            <span className="min-w-0 flex-1 truncate">{node.name}</span>
            {node.file && <CodeStatus status={node.file.status} />}
          </button>
        );
      })}
    </>
  );
}

export function SolutionViewer() {
  const codeFiles = useTaskStore((state) => state.codeFiles);
  const latestCodeFilePath = useTaskStore((state) => state.latestCodeFilePath);
  const updateCodeFileContent = useTaskStore((state) => state.updateCodeFileContent);
  const isDark = useThemeStore((state) => state.isDark);
  const [openFilePaths, setOpenFilePaths] = useState<string[]>([]);
  const [activePath, setActivePath] = useState<string | null>(null);
  const [isExplorerOpen, setIsExplorerOpen] = useState(() => typeof window === 'undefined' || window.innerWidth >= 720);
  const [expandedFolders, setExpandedFolders] = useState<Set<string>>(new Set());
  const [displayContent, setDisplayContent] = useState('');
  const [cursor, setCursor] = useState({ line: 1, column: 1 });
  const [showSource, setShowSource] = useState(false);
  const animationRef = useRef<ReturnType<typeof setInterval> | null>(null);

  const filesByPath = useMemo(() => new Map(codeFiles.map((file) => [file.path, file])), [codeFiles]);
  const activeFile = activePath ? filesByPath.get(activePath) ?? null : null;
  const activeIsMarkdown = activeFile ? isMarkdownPath(activeFile.path) : false;
  const activeIsBinary = activeFile ? isBinaryPath(activeFile.path) : false;
  // Markdown documents default to the rendered preview; users can flip to source.
  const renderAsPreview = activeIsMarkdown && !showSource;
  const openFiles = openFilePaths
    .map((path) => filesByPath.get(path))
    .filter((file): file is CodeFile => Boolean(file));
  const tree = useMemo(() => buildFileTree(codeFiles), [codeFiles]);

  useEffect(() => {
    if (codeFiles.length === 0) {
      setActivePath(null);
      setOpenFilePaths([]);
      setExpandedFolders(new Set());
      return;
    }

    const preferredPath = latestCodeFilePath && filesByPath.has(latestCodeFilePath)
      ? latestCodeFilePath
      : activePath && filesByPath.has(activePath)
        ? activePath
        : codeFiles[0].path;

    setActivePath(preferredPath);
    setOpenFilePaths((prev) => {
      const existing = prev.filter((path) => filesByPath.has(path));
      return existing.includes(preferredPath) ? existing : [...existing, preferredPath];
    });
  }, [activePath, codeFiles, filesByPath, latestCodeFilePath]);

  useEffect(() => {
    const folders = new Set<string>();
    codeFiles.forEach((file) => {
      const parts = file.path.split('/').filter(Boolean);
      let path = '';
      parts.slice(0, -1).forEach((part) => {
        path = path ? `${path}/${part}` : part;
        folders.add(path);
      });
    });
    setExpandedFolders((prev) => new Set([...folders, ...prev]));
  }, [codeFiles]);

  useEffect(() => {
    if (animationRef.current) {
      clearInterval(animationRef.current);
      animationRef.current = null;
    }

    const target = activeFile?.content ?? '';
    if (!activeFile) {
      setDisplayContent('');
      return;
    }

    if (activeFile.status !== 'streaming') {
      setDisplayContent(target);
      return;
    }

    setDisplayContent((prev) => (target.startsWith(prev) ? prev : ''));
    animationRef.current = window.setInterval(() => {
      setDisplayContent((prev) => {
        if (prev.length >= target.length) {
          if (animationRef.current) {
            clearInterval(animationRef.current);
            animationRef.current = null;
          }
          return target;
        }

        const remaining = target.length - prev.length;
        const step = Math.max(12, Math.ceil(target.length / 90));
        return target.slice(0, prev.length + Math.min(step, remaining));
      });
    }, 18);

    return () => {
      if (animationRef.current) {
        clearInterval(animationRef.current);
        animationRef.current = null;
      }
    };
  }, [activeFile?.content, activeFile?.path, activeFile?.status]);

  // Reset to the preferred view (preview for Markdown) whenever the file changes.
  useEffect(() => {
    setShowSource(false);
  }, [activePath]);

  const toggleFolder = (path: string) => {
    setExpandedFolders((prev) => {
      const next = new Set(prev);
      if (next.has(path)) {
        next.delete(path);
      } else {
        next.add(path);
      }
      return next;
    });
  };

  const openFile = (path: string) => {
    setActivePath(path);
    setOpenFilePaths((prev) => (prev.includes(path) ? prev : [...prev, path]));
  };

  const closeFile = (path: string) => {
    setOpenFilePaths((prev) => {
      const next = prev.filter((openPath) => openPath !== path);
      if (activePath === path) {
        setActivePath(next.at(-1) ?? codeFiles.find((file) => file.path !== path)?.path ?? null);
      }
      return next;
    });
  };

  const handleMount: OnMount = (editor) => {
    editor.onDidChangeCursorPosition((event) => {
      setCursor({ line: event.position.lineNumber, column: event.position.column });
    });
  };

  return (
    <div className="flex h-full min-h-0 flex-col bg-[var(--code-bg)] text-[var(--code-text)]" style={CODE_VIEW_THEME}>
      <div className="flex h-10 shrink-0 items-center justify-between border-b border-[var(--code-border)] bg-[var(--code-surface)] px-3">
        <div className="flex min-w-0 items-center gap-2">
          <button
            type="button"
            onClick={() => setIsExplorerOpen((value) => !value)}
            className="flex h-7 w-7 items-center justify-center rounded-md text-[var(--code-text-muted)] transition-colors hover:bg-[var(--code-tree-hover)] hover:text-[var(--code-text)]"
            title={isExplorerOpen ? 'Hide explorer' : 'Show explorer'}
          >
            {isExplorerOpen ? <PanelLeftClose size={16} /> : <PanelLeftOpen size={16} />}
          </button>
          <Files size={16} className="text-[var(--code-text-muted)]" />
          <span className="truncate text-sm font-medium">Solution files</span>
        </div>
        <div className="text-xs text-[var(--code-text-muted)]">
          {codeFiles.length} {codeFiles.length === 1 ? 'file' : 'files'}
        </div>
      </div>

      <div className="flex min-h-0 flex-1">
        {isExplorerOpen && (
          <aside className="hidden w-[280px] shrink-0 flex-col border-r border-[var(--code-border)] bg-[var(--code-surface)] md:flex">
            <div className="border-b border-[var(--code-border)] px-3 py-2 text-xs font-semibold uppercase tracking-wide text-[var(--code-text-muted)]">
              Explorer
            </div>
            <div className="min-h-0 flex-1 overflow-auto p-2">
              {tree.length > 0 ? (
                <TreeRows
                  nodes={tree}
                  depth={0}
                  activePath={activePath}
                  expandedFolders={expandedFolders}
                  onToggleFolder={toggleFolder}
                  onOpenFile={openFile}
                />
              ) : (
                <div className="flex h-full flex-col items-center justify-center px-5 text-center text-sm text-[var(--code-text-muted)]">
                  <FileCode2 size={36} className="mb-3 opacity-70" />
                  <span>No generated files yet.</span>
                </div>
              )}
            </div>
          </aside>
        )}

        <main className="flex min-w-0 flex-1 flex-col">
          <div className="flex h-10 shrink-0 overflow-x-auto border-b border-[var(--code-border)] bg-[var(--code-surface)]">
            {openFiles.length > 0 ? (
              openFiles.map((file) => (
                <div
                  key={file.path}
                  className={`group flex h-10 max-w-[240px] shrink-0 items-center border-r border-[var(--code-border)] text-sm transition-colors ${
                    activePath === file.path
                      ? 'bg-[var(--code-bg)] text-[var(--code-text)]'
                      : 'text-[var(--code-text-muted)] hover:bg-[var(--code-tree-hover)] hover:text-[var(--code-text)]'
                  }`}
                >
                  <button
                    type="button"
                    onClick={() => setActivePath(file.path)}
                    className="flex min-w-0 flex-1 items-center gap-2 px-3"
                  >
                    {file.path.toLowerCase().endsWith('.pptx') ? (
                      <Presentation size={15} />
                    ) : isBinaryPath(file.path) ? (
                      <FileBox size={15} />
                    ) : isMarkdownPath(file.path) ? (
                      <FileText size={15} />
                    ) : (
                      <FileCode2 size={15} />
                    )}
                    <span className="truncate">{file.name}</span>
                    <CodeStatus status={file.status} />
                  </button>
                  <button
                    type="button"
                    onClick={() => closeFile(file.path)}
                    className="mr-2 rounded p-0.5 opacity-0 transition-opacity hover:bg-[var(--code-tree-hover)] group-hover:opacity-100"
                    title="Close file"
                  >
                    <X size={13} />
                  </button>
                </div>
              ))
            ) : (
              <div className="flex items-center px-3 text-sm text-[var(--code-text-muted)]">No generated files</div>
            )}
          </div>

          {activeFile ? (
            <>
              <div className="flex h-9 shrink-0 items-center justify-between gap-3 border-b border-[var(--code-border)] px-3 text-xs text-[var(--code-text-muted)]">
                <div className="min-w-0 truncate">{activeFile.path}</div>
                <div className="flex shrink-0 items-center gap-2">
                  {activeIsMarkdown && (
                    <button
                      type="button"
                      onClick={() => setShowSource((value) => !value)}
                      className="flex items-center gap-1 rounded-md border border-[var(--code-border)] px-2 py-0.5 text-[var(--code-text-muted)] transition-colors hover:bg-[var(--code-tree-hover)] hover:text-[var(--code-text)]"
                      title={showSource ? 'Show rendered preview' : 'Show Markdown source'}
                    >
                      {showSource ? <Eye size={13} /> : <Code2 size={13} />}
                      <span>{showSource ? 'Preview' : 'Source'}</span>
                    </button>
                  )}
                  <CodeStatus status={activeFile.status} />
                  <span>{statusLabel(activeFile.status)}</span>
                </div>
              </div>
              <div className="min-h-0 flex-1">
                {activeIsBinary ? (
                  <div className="flex h-full flex-col items-center justify-center gap-3 px-6 text-center text-[var(--code-text-muted)]">
                    {activeFile.path.toLowerCase().endsWith('.pptx') ? (
                      <Presentation size={48} className="opacity-80 text-[var(--code-accent)]" />
                    ) : (
                      <FileBox size={48} className="opacity-80" />
                    )}
                    <div className="text-sm font-medium text-[var(--code-text)]">
                      {binaryFileLabel(activeFile.path)}
                    </div>
                    <p className="max-w-sm text-xs leading-5">
                      This file was generated on the server and is stored in the
                      project repository. Binary documents can't be previewed in the
                      browser — open the downloaded project to view or edit it.
                    </p>
                  </div>
                ) : renderAsPreview ? (
                  <div className="markdown-preview h-full overflow-auto px-6 py-5">
                    <ReactMarkdown remarkPlugins={[remarkGfm]}>{displayContent}</ReactMarkdown>
                  </div>
                ) : (
                  <Editor
                    path={activeFile.path}
                    value={displayContent}
                    language={activeFile.language}
                    theme={isDark ? 'vs-dark' : 'light'}
                    onMount={handleMount}
                    onChange={(value) => {
                      if (activeFile.status !== 'streaming' && value !== undefined && value !== activeFile.content) {
                        updateCodeFileContent(activeFile.path, value);
                      }
                    }}
                    options={{
                      automaticLayout: true,
                      fontFamily: 'JetBrains Mono, ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace',
                      fontSize: 13,
                      lineHeight: 22,
                      minimap: { enabled: true },
                      padding: { top: 12, bottom: 12 },
                      readOnly: activeFile.status === 'streaming',
                      scrollBeyondLastLine: false,
                      smoothScrolling: true,
                      wordWrap: activeIsMarkdown ? 'on' : 'off',
                    }}
                  />
                )}
              </div>
              <div className="flex h-7 shrink-0 items-center justify-between border-t border-[var(--code-border)] bg-[var(--code-surface)] px-3 text-xs text-[var(--code-text-muted)]">
                <span className="truncate">{activeFile.workerRole || activeFile.managerRole || 'Worker output'}</span>
                <span>
                  {activeIsBinary
                    ? binaryFileLabel(activeFile.path)
                    : renderAsPreview
                      ? 'Markdown preview'
                      : `Ln ${cursor.line}, Col ${cursor.column}`}
                </span>
              </div>
            </>
          ) : (
            <div className="flex min-h-0 flex-1 flex-col items-center justify-center px-6 text-center text-[var(--code-text-muted)]">
              <FileCode2 size={46} className="mb-4 opacity-70" />
              <div className="max-w-sm text-sm leading-6">
                No generated files yet.
              </div>
            </div>
          )}
        </main>
      </div>
    </div>
  );
}
