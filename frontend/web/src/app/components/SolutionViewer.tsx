import { useEffect, useMemo, useRef, useState } from 'react';
import type { CSSProperties } from 'react';
import Editor, { type OnMount } from '@monaco-editor/react';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import {
  CheckCircle2,
  ChevronDown,
  ChevronLeft,
  ChevronRight,
  CircleDotDashed,
  Code2,
  Download,
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
import {
  isMarkdownPath,
  isBinaryPath,
  binaryFileLabel,
  isDocumentSolution,
  paginateMarkdown,
} from '../../lib/markdown';
import { findPresentationPreviewFile, parsePresentationMarkdown } from '../../lib/presentation';
import { choosePreferredCodeFilePath } from '../../lib/solutionFiles';
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
  '--code-tree-guide': 'color-mix(in srgb, var(--text) 20%, transparent)',
} as CSSProperties & Record<string, string>;

// Width in pixels of one nesting level in the file tree. Each level draws a
// vertical guide line so it is obvious which files sit inside a folder.
const TREE_INDENT_WIDTH = 16;

// TreeIndent renders one vertical guide line per nesting level, making the
// folder hierarchy visually unambiguous (issue #31: nesting looked flat with no
// spacing, so it was unclear whether a file was inside a folder).
function TreeIndent({ depth }: { depth: number }) {
  if (depth <= 0) {
    return null;
  }
  return (
    <span className="flex flex-shrink-0 self-stretch" aria-hidden="true">
      {Array.from({ length: depth }).map((_, level) => (
        <span
          key={level}
          className="border-l border-[var(--code-tree-guide)]"
          style={{ width: TREE_INDENT_WIDTH }}
        />
      ))}
    </span>
  );
}

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

function binaryMimeType(path: string): string {
  const ext = path.toLowerCase().split('.').at(-1);
  switch (ext) {
    case 'pptx':
      return 'application/vnd.openxmlformats-officedocument.presentationml.presentation';
    case 'docx':
      return 'application/vnd.openxmlformats-officedocument.wordprocessingml.document';
    case 'xlsx':
      return 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet';
    case 'pdf':
      return 'application/pdf';
    case 'png':
      return 'image/png';
    case 'jpg':
    case 'jpeg':
      return 'image/jpeg';
    case 'gif':
      return 'image/gif';
    case 'zip':
      return 'application/zip';
    default:
      return 'application/octet-stream';
  }
}

function createBinaryDownloadUrl(file: CodeFile): string | null {
  if (file.encoding !== 'base64' || !file.content || typeof window === 'undefined' || typeof window.atob !== 'function') {
    return null;
  }

  try {
    const binary = window.atob(file.content);
    const bytes = new Uint8Array(binary.length);
    for (let i = 0; i < binary.length; i += 1) {
      bytes[i] = binary.charCodeAt(i);
    }
    return URL.createObjectURL(new Blob([bytes], { type: binaryMimeType(file.path) }));
  } catch {
    return null;
  }
}

function BinaryPlaceholder({ file }: { file: CodeFile }) {
  const path = file.path;
  const isPptx = path.toLowerCase().endsWith('.pptx');
  const downloadUrl = useMemo(() => createBinaryDownloadUrl(file), [file.content, file.encoding, file.path]);
  const downloadName = file.name || path.split('/').filter(Boolean).at(-1) || 'download';

  useEffect(() => {
    return () => {
      if (downloadUrl) {
        URL.revokeObjectURL(downloadUrl);
      }
    };
  }, [downloadUrl]);

  return (
    <div className="flex h-full flex-col items-center justify-center gap-3 px-6 text-center text-[var(--code-text-muted)]">
      {isPptx ? (
        <Presentation size={48} className="opacity-80 text-[var(--code-accent)]" />
      ) : (
        <FileBox size={48} className="opacity-80" />
      )}
      <div className="text-sm font-medium text-[var(--code-text)]">{binaryFileLabel(path)}</div>
      <p className="max-w-sm text-xs leading-5">
        This binary document can't be previewed in the browser.
        Download it to view or edit it locally.
      </p>
      {downloadUrl && (
        <a
          href={downloadUrl}
          download={downloadName}
          className="mt-1 inline-flex h-9 items-center gap-2 rounded-md border border-[var(--code-border)] px-3 text-sm font-medium text-[var(--code-text)] transition-colors hover:bg-[var(--code-tree-hover)]"
        >
          <Download size={15} />
          Download
        </a>
      )}
    </div>
  );
}

function PresentationDeckPreview({ file, previewFile }: { file: CodeFile; previewFile: CodeFile }) {
  const deck = useMemo(() => parsePresentationMarkdown(previewFile.content), [previewFile.content]);
  const slides = deck.slides.length > 0
    ? deck.slides
    : [{ title: deck.title || file.name || 'Presentation', bullets: [], visual: '', sources: [], notes: '' }];
  const [slideIndex, setSlideIndex] = useState(0);
  const currentSlide = slides[Math.min(slideIndex, Math.max(0, slides.length - 1))];
  const downloadUrl = useMemo(() => createBinaryDownloadUrl(file), [file.content, file.encoding, file.path]);
  const downloadName = file.name || file.path.split('/').filter(Boolean).at(-1) || 'presentation.pptx';

  useEffect(() => {
    setSlideIndex(0);
  }, [file.path, previewFile.path]);

  useEffect(() => {
    if (slideIndex >= slides.length) {
      setSlideIndex(Math.max(0, slides.length - 1));
    }
  }, [slideIndex, slides.length]);

  useEffect(() => {
    return () => {
      if (downloadUrl) {
        URL.revokeObjectURL(downloadUrl);
      }
    };
  }, [downloadUrl]);

  return (
    <div className="flex h-full min-h-0 flex-col bg-[var(--code-bg)] text-[var(--code-text)]">
      <div className="flex shrink-0 flex-wrap items-center justify-between gap-3 border-b border-[var(--code-border)] bg-[var(--code-surface)] px-4 py-3">
        <div className="flex min-w-0 items-center gap-3">
          <Presentation size={24} className="shrink-0 text-[var(--code-accent)]" />
          <div className="min-w-0">
            <div className="truncate text-sm font-semibold">{deck.title || binaryFileLabel(file.path)}</div>
            <div className="truncate text-xs text-[var(--code-text-muted)]">{file.name}</div>
          </div>
        </div>
        {downloadUrl && (
          <a
            href={downloadUrl}
            download={downloadName}
            className="inline-flex h-9 shrink-0 items-center gap-2 rounded-md border border-[var(--code-border)] px-3 text-sm font-medium text-[var(--code-text)] transition-colors hover:bg-[var(--code-tree-hover)]"
          >
            <Download size={15} />
            Download PPTX
          </a>
        )}
      </div>

      <div className="min-h-0 flex-1 overflow-auto px-4 py-5 sm:px-6">
        <div className="mx-auto flex w-full max-w-5xl flex-col gap-4">
          <section className="aspect-video overflow-auto rounded-md border border-[var(--code-border)] bg-[var(--code-surface)] p-6 shadow-sm sm:p-8">
            <div className="flex min-h-full flex-col">
              <div className="flex items-center justify-between gap-3 text-xs font-medium uppercase text-[var(--code-accent)]">
                <span>{slideIndex + 1} / {slides.length}</span>
                <span className="truncate text-[var(--code-text-muted)]">{previewFile.name}</span>
              </div>

              <div className="mt-8 max-w-3xl">
                <h1 className="text-2xl font-semibold leading-tight text-[var(--code-text)] sm:text-3xl">
                  {currentSlide.title || deck.title || 'Presentation'}
                </h1>
                {currentSlide.bullets.length > 0 && (
                  <ul className="mt-6 space-y-3 text-base leading-7 text-[var(--code-text)] sm:text-lg">
                    {currentSlide.bullets.map((bullet, index) => (
                      <li key={`${bullet}-${index}`} className="flex gap-3">
                        <span className="mt-3 h-1.5 w-1.5 shrink-0 rounded-full bg-[var(--code-accent)]" />
                        <span>{bullet}</span>
                      </li>
                    ))}
                  </ul>
                )}
              </div>

              {(currentSlide.visual || currentSlide.sources.length > 0 || currentSlide.notes) && (
                <div className="mt-auto grid gap-3 pt-8 text-sm leading-6 text-[var(--code-text-muted)] sm:grid-cols-2">
                  {currentSlide.visual && (
                    <div className="rounded-md border border-[var(--code-border)] px-3 py-2">
                      <div className="text-xs font-semibold uppercase text-[var(--code-text)]">Visual</div>
                      <div>{currentSlide.visual}</div>
                    </div>
                  )}
                  {currentSlide.sources.length > 0 && (
                    <div className="rounded-md border border-[var(--code-border)] px-3 py-2">
                      <div className="text-xs font-semibold uppercase text-[var(--code-text)]">Source</div>
                      <div>{currentSlide.sources.join(' | ')}</div>
                    </div>
                  )}
                  {currentSlide.notes && (
                    <div className="rounded-md border border-[var(--code-border)] px-3 py-2 sm:col-span-2">
                      <div className="text-xs font-semibold uppercase text-[var(--code-text)]">Notes</div>
                      <div>{currentSlide.notes}</div>
                    </div>
                  )}
                </div>
              )}
            </div>
          </section>
        </div>
      </div>

      {slides.length > 1 && (
        <div className="flex h-14 shrink-0 items-center justify-center border-t border-[var(--code-border)] bg-[var(--code-surface)]">
          <div className="flex items-center gap-2 rounded-full border border-[var(--code-border)] bg-[var(--code-bg)] px-2 py-1 shadow-sm">
            <button
              type="button"
              onClick={() => setSlideIndex((value) => Math.max(0, value - 1))}
              disabled={slideIndex <= 0}
              aria-label="Previous slide"
              className="flex h-8 w-8 items-center justify-center rounded-full text-[var(--code-text-muted)] transition-colors hover:bg-[var(--code-tree-hover)] hover:text-[var(--code-text)] disabled:cursor-not-allowed disabled:opacity-40 disabled:hover:bg-transparent"
            >
              <ChevronLeft size={18} />
            </button>
            <span className="min-w-[74px] text-center text-sm font-medium tabular-nums text-[var(--code-text)]">
              {slideIndex + 1} / {slides.length}
            </span>
            <button
              type="button"
              onClick={() => setSlideIndex((value) => Math.min(slides.length - 1, value + 1))}
              disabled={slideIndex >= slides.length - 1}
              aria-label="Next slide"
              className="flex h-8 w-8 items-center justify-center rounded-full text-[var(--code-text-muted)] transition-colors hover:bg-[var(--code-tree-hover)] hover:text-[var(--code-text)] disabled:cursor-not-allowed disabled:opacity-40 disabled:hover:bg-transparent"
            >
              <ChevronRight size={18} />
            </button>
          </div>
        </div>
      )}
    </div>
  );
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

        if (node.type === 'folder') {
          return (
            <div key={node.path}>
              <button
                type="button"
                onClick={() => onToggleFolder(node.path)}
                className="flex h-8 w-full items-center gap-1.5 rounded-md pl-2 pr-2 text-left text-sm text-[var(--code-text-muted)] transition-colors hover:bg-[var(--code-tree-hover)] hover:text-[var(--code-text)]"
              >
                <TreeIndent depth={depth} />
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
            className={`flex h-8 w-full items-center gap-2 rounded-md pl-2 pr-2 text-left text-sm transition-colors ${
              activePath === node.path
                ? 'bg-[var(--code-tree-active)] text-[var(--code-text)]'
                : 'text-[var(--code-text-muted)] hover:bg-[var(--code-tree-hover)] hover:text-[var(--code-text)]'
            }`}
          >
            <TreeIndent depth={depth} />
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
  const displayedContentByPathRef = useRef<Record<string, string>>({});

  const [pageIndex, setPageIndex] = useState(0);
  const readerRef = useRef<HTMLDivElement | null>(null);

  const filesByPath = useMemo(() => new Map(codeFiles.map((file) => [file.path, file])), [codeFiles]);
  const activeFile = activePath ? filesByPath.get(activePath) ?? null : null;
  const activeIsMarkdown = activeFile ? isMarkdownPath(activeFile.path) : false;
  const activeIsBinary = activeFile ? isBinaryPath(activeFile.path) : false;
  const presentationPreviewFile = useMemo(
    () => activeFile && activeFile.path.toLowerCase().endsWith('.pptx')
      ? findPresentationPreviewFile(codeFiles, activeFile.path)
      : null,
    [activeFile, codeFiles],
  );
  // A document solution (research report, generated document, presentation) gets
  // a clean reader — no explorer or tabs — so a regular user just reads the text.
  const documentMode = useMemo(() => isDocumentSolution(codeFiles.map((file) => file.path)), [codeFiles]);
  // Markdown documents default to the rendered preview; users can flip to source.
  const renderAsPreview = activeIsMarkdown && !showSource;
  const openFiles = openFilePaths
    .map((path) => filesByPath.get(path))
    .filter((file): file is CodeFile => Boolean(file));
  const tree = useMemo(() => buildFileTree(codeFiles), [codeFiles]);

  // In reader mode large Markdown documents are split into pages so the user can
  // flip through them with the bottom page switcher.
  const pages = useMemo(
    () => (documentMode && activeIsMarkdown ? paginateMarkdown(displayContent) : [displayContent]),
    [documentMode, activeIsMarkdown, displayContent],
  );
  const currentPage = Math.min(pageIndex, Math.max(0, pages.length - 1));

  useEffect(() => {
    if (codeFiles.length === 0) {
      displayedContentByPathRef.current = {};
      setActivePath(null);
      setOpenFilePaths([]);
      setExpandedFolders(new Set());
      return;
    }

    const preferredPath = choosePreferredCodeFilePath(
      codeFiles.map((file) => file.path),
      activePath,
      latestCodeFilePath,
    ) ?? codeFiles[0].path;

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

    const path = activeFile.path;
    if (activeFile.status !== 'streaming') {
      displayedContentByPathRef.current[path] = target;
      setDisplayContent(target);
      return;
    }

    const cached = displayedContentByPathRef.current[path] ?? '';
    setDisplayContent((prev) => {
      const candidate = target.startsWith(prev) ? prev : cached;
      const next = target.startsWith(candidate) ? candidate : '';
      displayedContentByPathRef.current[path] = next;
      return next;
    });
    animationRef.current = window.setInterval(() => {
      setDisplayContent((prev) => {
        const cachedValue = displayedContentByPathRef.current[path] ?? '';
        const current = target.startsWith(cachedValue) && cachedValue.length > prev.length ? cachedValue : prev;
        if (current.length >= target.length) {
          if (animationRef.current) {
            clearInterval(animationRef.current);
            animationRef.current = null;
          }
          displayedContentByPathRef.current[path] = target;
          return target;
        }

        const remaining = target.length - current.length;
        const step = Math.max(12, Math.ceil(target.length / 90));
        const next = target.slice(0, current.length + Math.min(step, remaining));
        displayedContentByPathRef.current[path] = next;
        return next;
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
    setPageIndex(0);
  }, [activePath]);

  // Scroll the reader back to the top whenever the visible page changes.
  useEffect(() => {
    readerRef.current?.scrollTo({ top: 0 });
  }, [currentPage, activePath]);

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

  // Reader mode: a single, beautifully formatted document with no file explorer
  // or tabs. Multiple documents get a minimal centered switcher; large documents
  // get a bottom page switcher.
  if (documentMode) {
    return (
      <div
        className="flex h-full min-h-0 flex-col bg-[var(--code-bg)] text-[var(--code-text)]"
        style={CODE_VIEW_THEME}
      >
        {codeFiles.length > 1 && (
          <div className="flex shrink-0 flex-wrap items-center justify-center gap-2 border-b border-[var(--code-border)] bg-[var(--code-surface)] px-4 py-2">
            {codeFiles.map((file) => (
              <button
                key={file.path}
                type="button"
                onClick={() => setActivePath(file.path)}
                className={`flex items-center gap-1.5 rounded-full border px-3 py-1 text-xs transition-colors ${
                  activePath === file.path
                    ? 'border-transparent bg-[var(--code-tree-active)] text-[var(--code-text)]'
                    : 'border-[var(--code-border)] text-[var(--code-text-muted)] hover:bg-[var(--code-tree-hover)] hover:text-[var(--code-text)]'
                }`}
              >
                {file.path.toLowerCase().endsWith('.pptx') ? (
                  <Presentation size={13} />
                ) : isBinaryPath(file.path) ? (
                  <FileBox size={13} />
                ) : (
                  <FileText size={13} />
                )}
                <span className="max-w-[160px] truncate">{file.name}</span>
              </button>
            ))}
          </div>
        )}

        <div ref={readerRef} className="min-h-0 flex-1 overflow-auto">
          {activeFile ? (
            activeIsBinary ? (
              presentationPreviewFile ? (
                <PresentationDeckPreview file={activeFile} previewFile={presentationPreviewFile} />
              ) : (
                <BinaryPlaceholder file={activeFile} />
              )
            ) : (
              <article className="markdown-preview mx-auto w-full max-w-[820px] px-6 py-10 sm:px-10">
                <ReactMarkdown remarkPlugins={[remarkGfm]}>{pages[currentPage] ?? ''}</ReactMarkdown>
              </article>
            )
          ) : (
            <div className="flex h-full flex-col items-center justify-center px-6 text-center text-[var(--code-text-muted)]">
              <FileText size={46} className="mb-4 opacity-70" />
              <div className="max-w-sm text-sm leading-6">No document yet.</div>
            </div>
          )}
        </div>

        {activeFile && !activeIsBinary && pages.length > 1 && (
          <div className="flex h-14 shrink-0 items-center justify-center border-t border-[var(--code-border)] bg-[var(--code-surface)]">
            <div className="flex items-center gap-2 rounded-full border border-[var(--code-border)] bg-[var(--code-bg)] px-2 py-1 shadow-sm">
              <button
                type="button"
                onClick={() => setPageIndex((value) => Math.max(0, value - 1))}
                disabled={currentPage <= 0}
                aria-label="Previous page"
                className="flex h-8 w-8 items-center justify-center rounded-full text-[var(--code-text-muted)] transition-colors hover:bg-[var(--code-tree-hover)] hover:text-[var(--code-text)] disabled:cursor-not-allowed disabled:opacity-40 disabled:hover:bg-transparent"
              >
                <ChevronLeft size={18} />
              </button>
              <span className="min-w-[64px] text-center text-sm font-medium tabular-nums text-[var(--code-text)]">
                {currentPage + 1} / {pages.length}
              </span>
              <button
                type="button"
                onClick={() => setPageIndex((value) => Math.min(pages.length - 1, value + 1))}
                disabled={currentPage >= pages.length - 1}
                aria-label="Next page"
                className="flex h-8 w-8 items-center justify-center rounded-full text-[var(--code-text-muted)] transition-colors hover:bg-[var(--code-tree-hover)] hover:text-[var(--code-text)] disabled:cursor-not-allowed disabled:opacity-40 disabled:hover:bg-transparent"
              >
                <ChevronRight size={18} />
              </button>
            </div>
          </div>
        )}
      </div>
    );
  }

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
                  presentationPreviewFile ? (
                    <PresentationDeckPreview file={activeFile} previewFile={presentationPreviewFile} />
                  ) : (
                    <BinaryPlaceholder file={activeFile} />
                  )
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
