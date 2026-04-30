import { useMemo, useState, useRef, useEffect } from 'react';
import hljs from '../../lib/hljs';

interface EditorFile {
  id: string;
  name: string;
  path: string;
  language: string;
  content: string;
}

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
			"message": "Welcome to CrewAI",
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

export function CodeViewer() {
  const [files, setFiles] = useState<EditorFile[]>(INITIAL_FILES);
  const [activeFileId, setActiveFileId] = useState(INITIAL_FILES[0].id);
  const [isExplorerOpen, setIsExplorerOpen] = useState(true);
  const [cursor, setCursor] = useState({ line: 1, col: 1 });
  const [expandedFolders, setExpandedFolders] = useState<Set<string>>(new Set(['src']));
  const textAreaRef = useRef<HTMLTextAreaElement>(null);
  const lineNumbersRef = useRef<HTMLDivElement>(null);

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

  const activeFile = files.find((file) => file.id === activeFileId) ?? files[0];

  const lineCount = useMemo(() => {
    if (!activeFile) return 1;
    return Math.max(activeFile.content.split('\n').length, 1);
  }, [activeFile]);

  const lines = useMemo(() => Array.from({ length: lineCount }, (_, i) => i + 1), [lineCount]);

  const getLanguage = (lang: string): string => {
    const langMap: Record<string, string> = {
      'Go': 'go',
      'JavaScript': 'javascript',
      'TypeScript': 'typescript',
      'Python': 'python',
      'Java': 'java',
      'C': 'c',
      'Cpp': 'cpp',
      'CSS': 'css',
      'JSON': 'json',
      'Markdown': 'markdown',
      'Bash': 'bash',
      'SQL': 'sql',
      'HTML': 'html',
      'YAML': 'yaml',
    };
    return langMap[lang] || 'plaintext';
  };

  const highlightedCode = useMemo(() => {
    if (!activeFile?.content) return '';
    const lang = getLanguage(activeFile.language);
    try {
      if (lang === 'plaintext' || !hljs.getLanguage(lang)) {
        return activeFile.content;
      }
      return hljs.highlight(activeFile.content, { language: lang }).value;
    } catch {
      return activeFile.content;
    }
  }, [activeFile?.content, activeFile?.language]);

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

  if (!activeFile) {
    return <div className="w-full h-full" />;
  }

  return (
    <div className="w-full h-full flex bg-[var(--background)] text-[var(--text)] overflow-hidden">
      <div className="w-12 bg-[var(--surface)] border-r border-[var(--border)] flex flex-col items-center py-1 shrink-0">
        <button
          className="w-12 h-12 flex items-center justify-center text-[var(--text)] relative"
          type="button"
          aria-label="Explorer"
          onClick={() => setIsExplorerOpen((prev) => !prev)}
        >
          <span className="absolute left-0 top-0 bottom-0 w-0.5 bg-[var(--accent)]" />
          <svg width="24" height="24" viewBox="0 0 24 24" fill="currentColor">
            <path d="M17.5 0h-9L7 1.5V6H2.5L1 7.5v15.07L2.5 24h12.07L16 22.57V18h4.7l1.3-1.43V4.5L17.5 0zm0 2.12l2.38 2.38H17.5V2.12zm-3 20.38h-12v-15H7v9.07L8.5 18h6v4.5zm6-6h-12v-15H16V6h4.5v10.5z" />
          </svg>
        </button>

        <button className="w-12 h-12 flex items-center justify-center text-[var(--text-muted)] hover:text-[var(--text)]" type="button" aria-label="Search">
          <svg width="24" height="24" viewBox="0 0 24 24" fill="currentColor">
            <path d="M15.25 0a8.25 8.25 0 0 0-6.18 13.72L1 22.88l1.12 1.12 8.05-9.12A8.251 8.251 0 1 0 15.25.01V0zm0 15a6.75 6.75 0 1 1 0-13.5 6.75 6.75 0 0 1 0 13.5z" />
          </svg>
        </button>

        <button className="w-12 h-12 flex items-center justify-center text-[var(--text-muted)] hover:text-[var(--text)]" type="button" aria-label="Source Control">
          <svg width="24" height="24" viewBox="0 0 24 24" fill="currentColor">
            <path d="M21.007 8.222A3.738 3.738 0 0 0 17.5 5.5a3.73 3.73 0 0 0-3.007 1.528 3.738 3.738 0 0 0-3.007-1.528 3.73 3.73 0 0 0-3.007 1.528A3.738 3.738 0 0 0 5.5 5.5a3.75 3.75 0 0 0 0 7.5 3.73 3.73 0 0 0 3.007-1.528 3.738 3.738 0 0 0 3.007 1.528 3.73 3.73 0 0 0 3.007-1.528A3.738 3.738 0 0 0 17.5 13a3.75 3.75 0 0 0 3.507-4.778z" />
          </svg>
        </button>

        <button className="w-12 h-12 flex items-center justify-center text-[var(--text-muted)] hover:text-[var(--text)]" type="button" aria-label="Run and Debug">
          <svg width="24" height="24" viewBox="0 0 24 24" fill="currentColor">
            <path d="M23 5.5h-2V3h-2v2.5h-2.5v2H19v2.5h2V7.5h2V5.5zM6 13.5a2.25 2.25 0 0 1 2.25 2.25h-4.5A2.25 2.25 0 0 1 6 13.5zm3 6a3 3 0 0 1-6 0v-2.25h6v2.25zm9.5-6a2.25 2.25 0 0 1 2.25 2.25h-4.5a2.25 2.25 0 0 1 2.25-2.25zm3 6a3 3 0 0 1-6 0v-2.25h6v2.25z" />
          </svg>
        </button>
      </div>

      {isExplorerOpen && (
        <div className="w-[250px] bg-[var(--surface)] border-r border-[var(--border)] flex flex-col shrink-0">
          <div className="h-9 px-4 flex items-center text-xs uppercase tracking-wider text-[var(--text-muted)] border-b border-[var(--border)]">Explorer</div>

          <div className="p-2 text-sm overflow-auto">
            <button
              type="button"
              onClick={() => toggleFolder('src')}
              className="w-full text-left px-2 py-1 rounded text-[var(--text)]/90 hover:bg-[var(--background)]"
            >
              <span className="mr-1">{expandedFolders.has('src') ? '▼' : '▶'}</span>
              src
            </button>
            {expandedFolders.has('src') && (
              <div className="ml-3">
                {files.map((file) => {
                  const isActive = file.id === activeFileId;
                  return (
                    <button
                      key={file.id}
                      type="button"
                      onClick={() => setActiveFileId(file.id)}
                      className={`w-full text-left pl-6 pr-2 py-1 rounded transition-colors ${
                        isActive
                          ? 'bg-[var(--accent)]/20 text-[var(--text)]'
                          : 'text-[var(--text)]/85 hover:bg-[var(--background)]'
                      }`}
                    >
                      <span className="text-[#519aba] font-semibold mr-2">G</span>
                      {file.name}
                    </button>
                  );
                })}
              </div>
            )}
            <button
              type="button"
              onClick={() => toggleFolder('pkg')}
              className="w-full text-left px-2 py-1 rounded text-[var(--text)]/80 hover:bg-[var(--background)] mt-1"
            >
              <span className="mr-1">{expandedFolders.has('pkg') ? '▼' : '▶'}</span>
              pkg
            </button>
            {expandedFolders.has('pkg') && (
              <div className="ml-3">
                <div className="pl-6 pr-2 py-1 rounded text-[var(--text)]/80">go.mod</div>
              </div>
            )}
          </div>
        </div>
      )}

      <div className="flex-1 min-w-0 flex flex-col bg-[var(--background)]">
        <div className="h-9 bg-[var(--surface)] flex overflow-x-auto border-b border-[var(--border)]">
          {files.map((file) => {
            const isActive = file.id === activeFileId;
            return (
              <button
                key={file.id}
                type="button"
                onClick={() => setActiveFileId(file.id)}
                className={`min-w-[120px] max-w-[220px] h-full px-3 border-r border-[var(--border)] text-sm flex items-center justify-between gap-2 ${
                  isActive
                    ? 'bg-[var(--background)] text-[var(--text)] border-t-2 border-t-[var(--accent)]'
                    : 'bg-[var(--surface)] text-[var(--text-muted)] hover:text-[var(--text)] hover:bg-[var(--background)]'
                }`}
              >
                <span className="truncate">{file.name}</span>
                <span className="opacity-60">×</span>
              </button>
            );
          })}
        </div>

        <div className="h-6 px-4 text-xs text-[var(--text-muted)] border-b border-[var(--border)] bg-[var(--background)] flex items-center gap-2">
          <span className="hover:text-[var(--text)] cursor-pointer">go-server</span>
          <span>›</span>
          <span className="hover:text-[var(--text)] cursor-pointer">src</span>
          <span>›</span>
          <span className="text-[var(--text)]">{activeFile.name}</span>
        </div>

        <div className="flex-1 min-h-0 flex overflow-hidden font-mono text-[13px] leading-6 bg-[var(--background)]">
          <div ref={lineNumbersRef} className="w-12 shrink-0 border-r border-[var(--border)] bg-[var(--surface)] text-right pr-2 text-[var(--text-muted)] overflow-hidden select-none">
            {lines.map((line) => (
              <div key={line} className={`h-6 text-xs leading-6 ${line === cursor.line ? 'bg-[var(--accent)]/20 text-[var(--text)]' : ''}`}>{line}</div>
            ))}
          </div>

          <div className="flex-1 relative overflow-hidden">
            <pre
              className="absolute inset-0 p-0 px-4 py-0 m-0 overflow-auto whitespace-pre pointer-events-none code-editor"
              style={{ tabSize: 2 }}
              dangerouslySetInnerHTML={{ __html: highlightedCode + '\n' }}
            />
            <textarea
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
                const pre = target.previousElementSibling as HTMLElement;
                if (pre) {
                  pre.scrollTop = target.scrollTop;
                }
                if (lineNumbersRef.current) {
                  lineNumbersRef.current.scrollTop = target.scrollTop;
                }
              }}
              spellCheck={false}
              className="absolute inset-0 w-full h-full resize-none border-0 outline-none bg-transparent text-transparent caret-[var(--text)] p-0 px-4 py-0 whitespace-pre"
              style={{ tabSize: 2 }}
            />
          </div>
        </div>

        <div className="h-6 bg-[var(--accent)] text-white px-3 text-xs flex items-center justify-between">
          <div className="flex items-center gap-3">
            <span>main</span>
          </div>
          <div className="flex items-center gap-3">
            <span>Ln {cursor.line}, Col {cursor.col}</span>
            <span>Spaces: 2</span>
            <span>UTF-8</span>
            <span>{activeFile.language}</span>
          </div>
        </div>
      </div>
    </div>
  );
}

export default CodeViewer;