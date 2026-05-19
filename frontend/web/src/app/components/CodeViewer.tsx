import { useMemo, useRef, useState, useEffect } from 'react';
import type { CSSProperties, MouseEvent } from 'react';
import {
  ArrowLeft,
  ArrowRight,
  Bug,
  ChevronDown,
  ChevronRight,
  Circle,
  File,
  FileCode2,
  Files,
  Folder,
  FolderOpen,
  GitBranch,
  Search,
  X,
} from 'lucide-react';
import hljs from '../../lib/hljs';

function escapeHtml(str: string): string {
  return str
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#039;');
}

interface EditorFile {
  id: string;
  name: string;
  path: string;
  language: string;
  content: string;
}

const LINE_HEIGHT = 24;
const EDITOR_PADDING_TOP = 16;
const EDITOR_GUTTER_WIDTH = 48;
const EDITOR_GUTTER_PADDING_X = 8;
const EDITOR_CONTENT_PADDING_X = 12;

const CODE_VIEW_THEME = {
  '--code-bg': 'var(--background)',
  '--code-surface': 'var(--surface)',
  '--code-surface-soft': 'var(--surface)',
  '--code-surface-muted': 'var(--surface)',
  '--code-border': 'var(--border)',
  '--code-border-soft': 'var(--border)',
  '--code-text': 'var(--text)',
  '--code-text-muted': 'var(--text-muted)',
  '--code-text-faint': 'color-mix(in srgb, var(--text-muted) 70%, transparent)',
  '--code-accent': 'var(--accent)',
  '--code-accent-strong': 'var(--accent)',
  '--code-selection': 'color-mix(in srgb, var(--accent) 35%, transparent)',
  '--code-line-active': 'color-mix(in srgb, var(--accent) 8%, transparent)',
  '--code-tree-active': 'color-mix(in srgb, var(--accent) 15%, transparent)',
  '--code-tree-hover': 'color-mix(in srgb, var(--text) 8%, transparent)',
} as CSSProperties & Record<string, string>;

const INITIAL_FILES: EditorFile[] = [
  {
    id: 'main-go',
    name: 'main.go',
    path: 'go-server/src/main.go',
    language: 'Go',
    content: `package main

import (
	"fmt"
	"net/http"
	"encoding/json"
	"context"
)

type Config struct {
	Port string \`json:"port"\`
	Host string \`json:"host"\`
}

func main() {
	cfg := Config{
		Port: "8080",
		Host: "localhost",
	}
	
	ctx := context.Background()
	_ = ctx

	server := &http.Server{
		Addr: ":" + cfg.Port,
	}
	
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := map[string]string{
			"message": "Welcome to Octra",
			"status":  "running",
		}
		_ = json.NewEncoder(w).Encode(resp)
		_ = r
	})
	
	fmt.Printf("Server starting on http://%s:%s\\n", cfg.Host, cfg.Port)
	fmt.Println("Press Ctrl+C to stop")
	
	if err := server.ListenAndServe(); err != nil {
		fmt.Println("Server error:", err)
	}
}`,
  },
  {
    id: 'config-go',
    name: 'config.go',
    path: 'go-server/src/config.go',
    language: 'Go',
    content: `package main

type ServerConfig struct {
	Host string
	Port string
}

func DefaultServerConfig() ServerConfig {
	return ServerConfig{
		Host: "localhost",
		Port: "8080",
	}
}
`,
  },
  {
    id: 'handler-go',
    name: 'handler.go',
    path: 'go-server/src/handler.go',
    language: 'Go',
    content: `package main

import "net/http"

func RootHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}
`,
  },
  {
    id: 'utils-go',
    name: 'utils.go',
    path: 'go-server/src/utils.go',
    language: 'Go',
    content: `package main

func Must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}
`,
  },
];

function getCursorInfo(text: string, position: number) {
  const before = text.slice(0, position);
  const parts = before.split('\n');
  const line = parts.length;
  const col = (parts.at(-1)?.length ?? 0) + 1;

  return { line, col };
}

function getLanguage(lang: string): string {
  const langMap: Record<string, string> = {
    Go: 'go',
    JavaScript: 'javascript',
    TypeScript: 'typescript',
    Python: 'python',
    Java: 'java',
    C: 'c',
    Cpp: 'cpp',
    CSS: 'css',
    JSON: 'json',
    Markdown: 'markdown',
    Bash: 'bash',
    SQL: 'sql',
    HTML: 'html',
    YAML: 'yaml',
  };
  return langMap[lang] || 'plaintext';
}

export function CodeViewer() {
  const [files, setFiles] = useState<EditorFile[]>(INITIAL_FILES);
  const [openFileIds, setOpenFileIds] = useState<string[]>([INITIAL_FILES[0].id]);
  const [activeFileId, setActiveFileId] = useState(INITIAL_FILES[0].id);
  const [isExplorerOpen, setIsExplorerOpen] = useState(() => typeof window === 'undefined' || window.innerWidth >= 720);
  const [explorerWidth, setExplorerWidth] = useState(320);
  const [cursor, setCursor] = useState({ line: 1, col: 1 });
  const [editorScroll, setEditorScroll] = useState({ top: 0, left: 0 });
  const [expandedFolders, setExpandedFolders] = useState<Set<string>>(new Set(['go-server', 'src', 'pkg']));
  const textAreaRef = useRef<HTMLTextAreaElement>(null);
  const explorerRef = useRef<HTMLDivElement>(null);
  const isResizing = useRef(false);

  const openFiles = openFileIds
    .map((id) => files.find((file) => file.id === id))
    .filter((file): file is EditorFile => Boolean(file));
  const activeFile = openFiles.find((file) => file.id === activeFileId) ?? null;

  useEffect(() => {
    const handleMouseMove = (e: globalThis.MouseEvent) => {
      if (!isResizing.current || !explorerRef.current) return;
      const newWidth = e.clientX - explorerRef.current.getBoundingClientRect().left;
      setExplorerWidth(Math.max(220, Math.min(520, newWidth)));
    };

    const handleMouseUp = () => {
      isResizing.current = false;
      document.body.style.cursor = '';
      document.body.style.userSelect = '';
    };

    document.addEventListener('mousemove', handleMouseMove);
    document.addEventListener('mouseup', handleMouseUp);

    return () => {
      document.removeEventListener('mousemove', handleMouseMove);
      document.removeEventListener('mouseup', handleMouseUp);
    };
  }, []);

  const lineCount = useMemo(() => {
    if (!activeFile) return 1;
    return Math.max(activeFile.content.split('\n').length, 1);
  }, [activeFile]);

  const lines = useMemo(() => Array.from({ length: lineCount }, (_, i) => i + 1), [lineCount]);

  const highlightedCode = useMemo(() => {
    if (!activeFile?.content) return '';
    const lang = getLanguage(activeFile.language);

    try {
      if (lang === 'plaintext' || !hljs.getLanguage(lang)) {
        return escapeHtml(activeFile.content);
      }

      return hljs.highlight(activeFile.content, { language: lang }).value;
    } catch {
      return escapeHtml(activeFile.content);
    }
  }, [activeFile?.content, activeFile?.language]);

  const pathParts = useMemo(() => activeFile?.path.split('/') ?? [], [activeFile?.path]);

  const activeLineTop = EDITOR_PADDING_TOP + (cursor.line - 1) * LINE_HEIGHT - editorScroll.top;

  const startResize = () => {
    isResizing.current = true;
    document.body.style.cursor = 'col-resize';
    document.body.style.userSelect = 'none';
  };

  const toggleFolder = (folder: string) => {
    setExpandedFolders((prev) => {
      const next = new Set(prev);
      if (next.has(folder)) {
        next.delete(folder);
      } else {
        next.add(folder);
      }
      return next;
    });
  };

  const closeFile = (fileId: string, e: MouseEvent<HTMLButtonElement>) => {
    e.stopPropagation();

    const nextOpenFileIds = openFileIds.filter((id) => id !== fileId);
    setOpenFileIds(nextOpenFileIds);

    if (activeFileId === fileId) {
      setActiveFileId(nextOpenFileIds.at(-1) ?? '');
    }
  };

  const openFile = (fileId: string) => {
    setOpenFileIds((prev) => (prev.includes(fileId) ? prev : [...prev, fileId]));
    setActiveFileId(fileId);
    setCursor({ line: 1, col: 1 });
    setEditorScroll({ top: 0, left: 0 });
  };

  const updateActiveFile = (nextContent: string, cursorPos: number) => {
    setFiles((prev) =>
      prev.map((file) =>
        file.id === activeFileId
          ? {
              ...file,
              content: nextContent,
            }
          : file,
      ),
    );
    setCursor(getCursorInfo(nextContent, cursorPos));
  };

  return (
    <div
      className="w-full h-full flex overflow-hidden bg-[var(--code-bg)] text-[var(--code-text)]"
      style={CODE_VIEW_THEME}
    >
      <div className="w-[50px] bg-[var(--code-surface-muted)] border-r border-[var(--code-border)] flex flex-col items-center py-2 shrink-0">
        <button
          className={`relative w-9 h-9 flex items-center justify-center rounded-md transition-colors ${
            isExplorerOpen
              ? 'bg-[var(--code-tree-active)] text-[var(--code-text)]'
              : 'text-[var(--code-text-muted)] hover:bg-[var(--code-tree-hover)] hover:text-[var(--code-text)]'
          }`}
          type="button"
          aria-label="Explorer"
          title="Explorer"
          onClick={() => setIsExplorerOpen((prev) => !prev)}
        >
          {isExplorerOpen && <span className="absolute left-0 top-2 bottom-2 w-0.5 rounded-full bg-[var(--code-accent)]" />}
          <Files size={20} strokeWidth={1.8} />
        </button>

        <button
          className="mt-2 w-9 h-9 flex items-center justify-center rounded-md text-[var(--code-text-muted)] transition-colors hover:bg-[var(--code-tree-hover)] hover:text-[var(--code-text)]"
          type="button"
          aria-label="Search"
          title="Search"
        >
          <Search size={20} strokeWidth={1.8} />
        </button>

        <button
          className="w-9 h-9 flex items-center justify-center rounded-md text-[var(--code-text-muted)] transition-colors hover:bg-[var(--code-tree-hover)] hover:text-[var(--code-text)]"
          type="button"
          aria-label="Source Control"
          title="Source Control"
        >
          <GitBranch size={20} strokeWidth={1.8} />
        </button>

        <button
          className="w-9 h-9 flex items-center justify-center rounded-md text-[var(--code-text-muted)] transition-colors hover:bg-[var(--code-tree-hover)] hover:text-[var(--code-text)]"
          type="button"
          aria-label="Run and Debug"
          title="Run and Debug"
        >
          <Bug size={20} strokeWidth={1.8} />
        </button>
      </div>

      {isExplorerOpen && (
        <aside
          ref={explorerRef}
          style={{ width: `min(${explorerWidth}px, calc(100vw - 180px))` }}
          className="relative bg-[var(--code-surface-soft)] border-r border-[var(--code-border)] flex flex-col shrink-0"
        >
          <div
            className="absolute right-0 top-0 bottom-0 w-1 cursor-col-resize hover:bg-[var(--code-accent)]/40 transition-colors z-10"
            onMouseDown={startResize}
          />

          <div className="h-11 px-4 border-b border-[var(--code-border-soft)] flex items-center justify-between">
            <span className="text-xs font-medium text-[var(--code-text-muted)] uppercase">Project</span>
            <span className="text-[11px] text-[var(--code-text-faint)]">go-server</span>
          </div>

          <div className="flex-1 overflow-auto px-2 py-3 text-sm">
            <button
              type="button"
              onClick={() => toggleFolder('go-server')}
              className="w-full h-7 px-2 flex items-center gap-2 rounded text-[var(--code-text)]/90 transition-colors hover:bg-[var(--code-tree-hover)]"
            >
              {expandedFolders.has('go-server') ? <ChevronDown size={15} /> : <ChevronRight size={15} />}
              {expandedFolders.has('go-server') ? <FolderOpen size={16} /> : <Folder size={16} />}
              <span className="truncate">go-server</span>
            </button>

            {expandedFolders.has('go-server') && (
              <div className="ml-4 border-l border-[var(--code-border-soft)] pl-2">
                <button
                  type="button"
                  onClick={() => toggleFolder('src')}
                  className="w-full h-7 px-2 flex items-center gap-2 rounded text-[var(--code-text)]/90 transition-colors hover:bg-[var(--code-tree-hover)]"
                >
                  {expandedFolders.has('src') ? <ChevronDown size={15} /> : <ChevronRight size={15} />}
                  {expandedFolders.has('src') ? <FolderOpen size={16} /> : <Folder size={16} />}
                  <span className="truncate">src</span>
                </button>

                {expandedFolders.has('src') && (
                  <div className="ml-4 border-l border-[var(--code-border-soft)] pl-2">
                    {files.map((file) => {
                      const isActive = file.id === activeFileId;
                      const isOpen = openFileIds.includes(file.id);

                      return (
                        <button
                          key={file.id}
                          type="button"
                          onClick={() => openFile(file.id)}
                          className={`w-full h-7 pl-2 pr-2 flex items-center gap-2 rounded transition-colors ${
                            isActive
                              ? 'bg-[var(--code-tree-active)] text-[var(--code-text)]'
                              : 'text-[var(--code-text-muted)] hover:bg-[var(--code-tree-hover)] hover:text-[var(--code-text)]'
                          }`}
                        >
                          <FileCode2 size={15} className="text-[var(--code-accent)]" />
                          <span className="truncate">{file.name}</span>
                          {isOpen && <Circle size={6} fill="currentColor" className="ml-auto text-[var(--code-text-faint)]" />}
                        </button>
                      );
                    })}
                  </div>
                )}

                <button
                  type="button"
                  onClick={() => toggleFolder('pkg')}
                  className="mt-1 w-full h-7 px-2 flex items-center gap-2 rounded text-[var(--code-text)]/80 transition-colors hover:bg-[var(--code-tree-hover)]"
                >
                  {expandedFolders.has('pkg') ? <ChevronDown size={15} /> : <ChevronRight size={15} />}
                  {expandedFolders.has('pkg') ? <FolderOpen size={16} /> : <Folder size={16} />}
                  <span className="truncate">pkg</span>
                </button>

                {expandedFolders.has('pkg') && (
                  <div className="ml-4 border-l border-[var(--code-border-soft)] pl-2">
                    <div className="h-7 pl-2 pr-2 flex items-center gap-2 text-[var(--code-text-muted)]">
                      <File size={15} className="text-[var(--code-text-faint)]" />
                      <span className="truncate">go.mod</span>
                    </div>
                  </div>
                )}
              </div>
            )}
          </div>
        </aside>
      )}

      <main className="flex-1 min-w-0 flex flex-col bg-[var(--code-bg)]">
        <div className="h-11 bg-[var(--code-surface-soft)] border-b border-[var(--code-border)] flex items-end overflow-hidden">
          <div className="h-full px-2 border-r border-[var(--code-border)] flex items-center gap-1 shrink-0">
            <button
              type="button"
              className="w-8 h-8 rounded-md flex items-center justify-center text-[var(--code-text-muted)] transition-colors hover:bg-[var(--code-tree-hover)] hover:text-[var(--code-text)]"
              aria-label="Back"
              title="Back"
            >
              <ArrowLeft size={17} />
            </button>
            <button
              type="button"
              className="w-8 h-8 rounded-md flex items-center justify-center text-[var(--code-text-faint)] transition-colors hover:bg-[var(--code-tree-hover)] hover:text-[var(--code-text-muted)]"
              aria-label="Forward"
              title="Forward"
            >
              <ArrowRight size={17} />
            </button>
          </div>

          <div className="flex-1 min-w-0 flex overflow-x-auto">
            {openFiles.length > 0 ? (
              openFiles.map((file) => {
                const isActive = file.id === activeFileId;

                return (
                  <div
                    key={file.id}
                    className={`h-11 min-w-[148px] max-w-[240px] px-3 border-r border-[var(--code-border)] flex items-center gap-2 text-sm transition-colors ${
                      isActive
                        ? 'bg-[var(--code-bg)] text-[var(--code-text)]'
                        : 'bg-[var(--code-surface-soft)] text-[var(--code-text-muted)] hover:bg-[var(--code-tree-hover)] hover:text-[var(--code-text)]'
                    }`}
                  >
                    <button
                      type="button"
                      onClick={() => {
                        setActiveFileId(file.id);
                        setCursor({ line: 1, col: 1 });
                        setEditorScroll({ top: 0, left: 0 });
                      }}
                      className="min-w-0 flex-1 h-full flex items-center gap-2 text-left"
                    >
                      <FileCode2 size={15} className={isActive ? 'text-[var(--code-accent-strong)]' : 'text-[var(--code-text-faint)]'} />
                      <span className="truncate">{file.name}</span>
                    </button>
                    <button
                      type="button"
                      onClick={(e) => closeFile(file.id, e)}
                      className="ml-auto w-5 h-5 rounded flex items-center justify-center text-[var(--code-text-faint)] transition-colors hover:bg-[var(--code-tree-active)] hover:text-[var(--code-text)]"
                      aria-label={`Close ${file.name}`}
                      title={`Close ${file.name}`}
                    >
                      <X size={13} />
                    </button>
                  </div>
                );
              })
            ) : (
              <div className="h-11 px-4 flex items-center text-sm text-[var(--code-text-faint)]">No file open</div>
            )}
          </div>
        </div>

        {activeFile && (
          <div className="h-10 px-5 text-sm border-b border-[var(--code-border-soft)] bg-[var(--code-bg)] flex items-center gap-2 overflow-hidden">
            {pathParts.map((part, index) => {
              const isLast = index === pathParts.length - 1;
              return (
                <span key={`${part}-${index}`} className="flex items-center gap-2 min-w-0">
                  <span className={isLast ? 'text-[var(--code-text)] truncate' : 'text-[var(--code-text-muted)] truncate'}>
                    {part}
                  </span>
                  {!isLast && <ChevronRight size={14} className="text-[var(--code-text-faint)] shrink-0" />}
                </span>
              );
            })}
          </div>
        )}

        {activeFile ? (
          <div className="flex-1 min-h-0 flex overflow-hidden bg-[var(--code-bg)] font-mono text-[13px]">
            <div
              className="shrink-0 border-r border-[var(--code-border-soft)] bg-[var(--code-bg)] text-right text-[var(--code-text-faint)] overflow-hidden select-none"
              style={{ width: EDITOR_GUTTER_WIDTH }}
            >
              <div
                style={{
                  paddingTop: EDITOR_PADDING_TOP,
                  paddingLeft: EDITOR_GUTTER_PADDING_X,
                  paddingRight: EDITOR_GUTTER_PADDING_X,
                  transform: `translateY(${-editorScroll.top}px)`,
                  lineHeight: `${LINE_HEIGHT}px`,
                }}
              >
                {lines.map((line) => (
                  <div
                    key={line}
                    className={`h-6 text-xs ${line === cursor.line ? 'text-[var(--code-accent-strong)]' : ''}`}
                  >
                    {line}
                  </div>
                ))}
              </div>
            </div>

            <div className="flex-1 min-w-0 relative overflow-hidden">
              {activeLineTop > -LINE_HEIGHT && (
                <div
                  className="absolute left-0 right-0 h-6 bg-[var(--code-line-active)] pointer-events-none"
                  style={{ top: activeLineTop }}
                />
              )}

              <pre
                className="absolute inset-0 z-10 p-0 m-0 overflow-hidden whitespace-pre pointer-events-none code-editor"
                aria-hidden="true"
              >
                <code
                  className="block"
                  style={{
                    paddingTop: EDITOR_PADDING_TOP,
                    paddingLeft: EDITOR_CONTENT_PADDING_X,
                    paddingRight: EDITOR_CONTENT_PADDING_X,
                    tabSize: 2,
                    lineHeight: `${LINE_HEIGHT}px`,
                    transform: `translate(${-editorScroll.left}px, ${-editorScroll.top}px)`,
                    fontFamily:
                      'JetBrains Mono, SFMono-Regular, Menlo, Monaco, Consolas, Liberation Mono, monospace',
                  }}
                  dangerouslySetInnerHTML={{ __html: highlightedCode + '\n' }}
                />
              </pre>

              <textarea
                key={activeFile.id}
                ref={textAreaRef}
                value={activeFile.content}
                onChange={(e) => updateActiveFile(e.target.value, e.target.selectionStart)}
                onSelect={(e) => {
                  const target = e.currentTarget;
                  setCursor(getCursorInfo(target.value, target.selectionStart));
                }}
                onClick={(e) => {
                  const target = e.currentTarget;
                  setCursor(getCursorInfo(target.value, target.selectionStart));
                }}
                onKeyUp={(e) => {
                  const target = e.currentTarget;
                  setCursor(getCursorInfo(target.value, target.selectionStart));
                }}
                onScroll={(e) => {
                  const target = e.currentTarget;
                  setEditorScroll({ top: target.scrollTop, left: target.scrollLeft });
                }}
                spellCheck={false}
                className="absolute inset-0 z-20 w-full h-full resize-none border-0 outline-none bg-transparent text-transparent caret-[var(--code-text)] selection:bg-[var(--code-selection)] selection:text-transparent overflow-auto whitespace-pre"
                style={{
                  paddingTop: EDITOR_PADDING_TOP,
                  paddingLeft: EDITOR_CONTENT_PADDING_X,
                  paddingRight: EDITOR_CONTENT_PADDING_X,
                  tabSize: 2,
                  lineHeight: `${LINE_HEIGHT}px`,
                  fontFamily:
                    'JetBrains Mono, SFMono-Regular, Menlo, Monaco, Consolas, Liberation Mono, monospace',
                }}
              />
            </div>
          </div>
        ) : (
          <div className="flex-1 min-h-0 bg-[var(--code-bg)] flex items-center justify-center text-[var(--code-text-muted)]">
            <div className="flex flex-col items-center gap-3">
              <FileCode2 size={42} strokeWidth={1.4} className="text-[var(--code-text-faint)]" />
              <p className="text-sm">Open a file from the project tree</p>
            </div>
          </div>
        )}

        <div className="h-7 bg-[var(--code-surface-muted)] border-t border-[var(--code-border)] text-[12px] text-[var(--code-text-muted)] px-3 flex items-center justify-between">
          <div className="min-w-0 flex items-center gap-3">
            <span className="flex items-center gap-1.5 text-[var(--code-text)]">
              <GitBranch size={14} />
              main
            </span>
            {activeFile && (
              <span className="hidden sm:inline truncate text-[var(--code-text-faint)]">{activeFile.path}</span>
            )}
          </div>

          {activeFile && (
            <div className="flex items-center gap-3 shrink-0">
              <span>Ln {cursor.line}, Col {cursor.col}</span>
              <span>Spaces: 2</span>
              <span>UTF-8</span>
              <span>{activeFile.language}</span>
            </div>
          )}
        </div>
      </main>
    </div>
  );
}

export default CodeViewer;
