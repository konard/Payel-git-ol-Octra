/**
 * Typed access to the Electron preload bridge (window.octra).
 *
 * The desktop app (frontend/desktop) runs THIS web app as its renderer and
 * injects `window.octra` via a contextBridge. In a normal browser that object is
 * absent, so every desktop feature is gated behind `isDesktopApp()` and the web
 * build keeps behaving exactly as before. This module is the single, typed place
 * the React code touches the bridge.
 */

export interface FileTreeNode {
  name: string;
  path: string;
  type: 'file' | 'directory';
  children?: FileTreeNode[];
  size?: number;
}

export interface ReadFileResult {
  path: string;
  content: string;
  encoding: 'utf8' | 'base64';
  size: number;
  truncated: boolean;
  binary: boolean;
}

export interface DesktopProject {
  name: string;
  path: string;
  openedAt: number | null;
}

export interface RememberResult {
  project: DesktopProject;
  recent: DesktopProject[];
}

export interface DesktopConfig {
  apiBaseUrl: string;
  wsUrl: string;
  platform: string;
}

export interface OctraBridge {
  isElectron: boolean;
  platform: string;
  window: {
    minimize: () => Promise<void>;
    toggleMaximize: () => Promise<boolean>;
    close: () => Promise<void>;
    isMaximized: () => Promise<boolean>;
    onMaximizeChange: (cb: (maximized: boolean) => void) => () => void;
  };
  projects: {
    listRecent: () => Promise<DesktopProject[]>;
    open: () => Promise<RememberResult | null>;
    openPath: (path: string) => Promise<RememberResult>;
    create: () => Promise<RememberResult | null>;
    forget: (path: string) => Promise<DesktopProject[]>;
  };
  fs: {
    readTree: (root: string) => Promise<FileTreeNode>;
    readFile: (root: string, rel: string) => Promise<ReadFileResult>;
  };
  shell: {
    openExternal: (url: string) => Promise<void>;
  };
  getConfig: () => Promise<DesktopConfig>;
  onMenuCommand: (cb: (command: string, payload?: unknown) => void) => () => void;
}

declare global {
  interface Window {
    octra?: OctraBridge;
  }
}

// True only when running inside the Electron desktop shell.
export function isDesktopApp(): boolean {
  return typeof window !== 'undefined' && Boolean(window.octra?.isElectron);
}

// The bridge, or null in a plain browser. Callers should guard with isDesktopApp.
export function getBridge(): OctraBridge | null {
  return isDesktopApp() ? (window.octra as OctraBridge) : null;
}
