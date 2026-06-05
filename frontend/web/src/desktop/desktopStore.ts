/**
 * Desktop-only UI state: the currently open project, its file tree, the file
 * being viewed, and whether the file explorer dock is shown. Lives in its own
 * store so the web app is untouched when not running in Electron — none of this
 * state is ever populated in a plain browser.
 */

import { create } from 'zustand';
import {
  getBridge,
  isDesktopApp,
  type DesktopProject,
  type FileTreeNode,
  type ReadFileResult,
} from './bridge';

export interface OpenFile {
  path: string;
  name: string;
  content: string;
  language: string;
  binary: boolean;
  truncated: boolean;
}

interface DesktopState {
  project: DesktopProject | null;
  tree: FileTreeNode | null;
  recent: DesktopProject[];
  openFile: OpenFile | null;
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

function toOpenFile(result: ReadFileResult): OpenFile {
  return {
    path: result.path,
    name: result.path.split('/').pop() || result.path,
    content: result.binary
      ? `[Binary file — ${result.size} bytes, not shown]`
      : result.content + (result.truncated ? '\n\n… [file truncated]' : ''),
    language: result.binary ? 'plaintext' : languageForPath(result.path),
    binary: result.binary,
    truncated: result.truncated,
  };
}

export const useDesktopStore = create<DesktopState>((set, get) => ({
  project: null,
  tree: null,
  recent: [],
  openFile: null,
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
        set({ project: result.project, recent: result.recent, openFile: null });
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
      set({ project: result.project, recent: result.recent, openFile: null });
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
      set({ openFile: toOpenFile(result), loadingFile: false });
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
