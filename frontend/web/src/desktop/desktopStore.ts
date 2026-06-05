/**
 * Desktop-only UI state: the currently open project, its file tree, the path of
 * the file last opened from the explorer, and whether the file explorer dock is
 * shown. Lives in its own store so the web app is untouched when not running in
 * Electron — none of this state is ever populated in a plain browser.
 *
 * Opening a file no longer pops a separate viewer window. Per the owner's
 * feedback on issue #50 ("Удали боковое окно просмотра файлов и открывай в
 * solution files"), an opened project file is pushed into the web app's task
 * store so it renders inside the existing "Solution files" panel, reusing its
 * Monaco editor, Markdown preview, binary download and tab UI.
 */

import { create } from 'zustand';
import {
  getBridge,
  isDesktopApp,
  type DesktopProject,
  type FileTreeNode,
  type ReadFileResult,
} from './bridge';
import { useTaskStore } from '../stores/taskStore';

interface DesktopState {
  project: DesktopProject | null;
  tree: FileTreeNode | null;
  recent: DesktopProject[];
  // Path of the file most recently opened from the explorer. Drives the active
  // row highlight; the file's contents live in the task store's codeFiles.
  openFilePath: string | null;
  // Bumped on every open so the Solution files panel can focus the freshly
  // opened file even when the same path is re-opened.
  openNonce: number;
  explorerOpen: boolean;
  loadingTree: boolean;
  loadingFile: boolean;
  error: string | null;

  refreshRecent: () => Promise<void>;
  openFolder: () => Promise<void>;
  openProjectPath: (path: string) => Promise<void>;
  reloadTree: () => Promise<void>;
  openPath: (relativePath: string) => Promise<void>;
  toggleExplorer: () => void;
  setExplorerOpen: (open: boolean) => void;
}

// Map a file extension to a Monaco language id. Mirrors the languages the
// SolutionViewer relies on; anything unknown falls back to plain text.
const LANGUAGE_BY_EXT: Record<string, string> = {
  ts: 'typescript', tsx: 'typescript', js: 'javascript', jsx: 'javascript',
  mjs: 'javascript', cjs: 'javascript', json: 'json', md: 'markdown',
  markdown: 'markdown', css: 'css', scss: 'scss', less: 'less', html: 'html',
  htm: 'html', xml: 'xml', svg: 'xml', yml: 'yaml', yaml: 'yaml', py: 'python',
  rb: 'ruby', go: 'go', rs: 'rust', java: 'java', kt: 'kotlin', c: 'c',
  h: 'c', cpp: 'cpp', cc: 'cpp', hpp: 'cpp', cs: 'csharp', php: 'php',
  sh: 'shell', bash: 'shell', zsh: 'shell', sql: 'sql', toml: 'ini',
  ini: 'ini', dockerfile: 'dockerfile', vue: 'html', svelte: 'html',
};

function languageForPath(filePath: string): string {
  const base = filePath.split('/').pop() || filePath;
  if (base.toLowerCase() === 'dockerfile') return 'dockerfile';
  const ext = base.includes('.') ? base.split('.').pop()!.toLowerCase() : '';
  return LANGUAGE_BY_EXT[ext] || 'plaintext';
}

// Map a file read from disk into the task store's CodeFile shape so it shows up
// in the "Solution files" panel. Binary files keep their base64 content and
// encoding so the panel can offer a download; text files get an inline note when
// they were truncated for size.
function toCodeFile(result: ReadFileResult) {
  return {
    path: result.path,
    name: result.path.split('/').pop() || result.path,
    language: result.binary ? 'plaintext' : languageForPath(result.path),
    encoding: result.encoding,
    content:
      result.binary || !result.truncated
        ? result.content
        : result.content + '\n\n… [file truncated]',
    status: 'ready' as const,
  };
}

export const useDesktopStore = create<DesktopState>((set, get) => ({
  project: null,
  tree: null,
  recent: [],
  openFilePath: null,
  openNonce: 0,
  explorerOpen: true,
  loadingTree: false,
  loadingFile: false,
  error: null,

  refreshRecent: async () => {
    const bridge = getBridge();
    if (!bridge) return;
    try {
      const recent = await bridge.projects.listRecent();
      set({ recent });
    } catch (err) {
      set({ error: String(err) });
    }
  },

  openFolder: async () => {
    const bridge = getBridge();
    if (!bridge) return;
    try {
      const result = await bridge.projects.open();
      if (result) {
        set({ project: result.project, recent: result.recent, openFilePath: null });
        await get().reloadTree();
      }
    } catch (err) {
      set({ error: String(err) });
    }
  },

  openProjectPath: async (path: string) => {
    const bridge = getBridge();
    if (!bridge) return;
    try {
      const result = await bridge.projects.openPath(path);
      set({ project: result.project, recent: result.recent, openFilePath: null });
      await get().reloadTree();
    } catch (err) {
      set({ error: String(err) });
    }
  },

  reloadTree: async () => {
    const bridge = getBridge();
    const project = get().project;
    if (!bridge || !project) return;
    set({ loadingTree: true, error: null });
    try {
      const tree = await bridge.fs.readTree(project.path);
      set({ tree, loadingTree: false });
    } catch (err) {
      set({ error: String(err), loadingTree: false });
    }
  },

  openPath: async (relativePath: string) => {
    const bridge = getBridge();
    const project = get().project;
    if (!bridge || !project) return;
    set({ loadingFile: true, error: null });
    try {
      const result = await bridge.fs.readFile(project.path, relativePath);
      // Surface the file in the web app's "Solution files" panel instead of a
      // separate desktop viewer window (issue #50 owner feedback).
      useTaskStore.getState().upsertCodeFiles([toCodeFile(result)]);
      set((s) => ({ openFilePath: result.path, openNonce: s.openNonce + 1, loadingFile: false }));
    } catch (err) {
      set({ error: String(err), loadingFile: false });
    }
  },

  toggleExplorer: () => set((s) => ({ explorerOpen: !s.explorerOpen })),
  setExplorerOpen: (open: boolean) => set({ explorerOpen: open }),
}));

// Hydrate the recent-projects list once at startup when in the desktop app.
if (isDesktopApp()) {
  void useDesktopStore.getState().refreshRecent();
}
