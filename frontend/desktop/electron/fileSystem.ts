'use strict';

// Reading project files from disk — the feature the previous desktop shell was
// missing entirely ("Чтение файлов нету … я открыл не пустой проект но ничего не
// отобразилось", issue #50.4). The renderer asks the main process to walk a
// project folder into a tree and to read individual files; both are implemented
// here as pure Node so they can be unit-tested without launching Electron.

import * as fs from 'fs';
import * as path from 'path';

export interface FileTreeNode {
  name: string;
  path: string; // POSIX-style path relative to the project root
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

// Folders and files that never carry useful project content and that would make
// the tree huge / slow if walked. Mirrors a typical editor's default excludes.
const IGNORED = new Set([
  'node_modules',
  '.git',
  '.hg',
  '.svn',
  '.cache',
  '.next',
  '.nuxt',
  '.turbo',
  'dist',
  'build',
  'out',
  'coverage',
  '.idea',
  '.vscode',
  '__pycache__',
  '.venv',
  'venv',
  '.DS_Store',
  'dist-electron',
]);

// Safety limits so a pathological project can never hang or OOM the renderer.
const MAX_DEPTH = 12;
const MAX_ENTRIES_PER_DIR = 2000;
const MAX_FILE_BYTES = 2 * 1024 * 1024; // 2 MB — files above this are truncated.

// Common binary extensions. We still let the caller open them (e.g. to show an
// image), but text rendering of arbitrary binaries is avoided.
const BINARY_EXTENSIONS = new Set([
  'png', 'jpg', 'jpeg', 'gif', 'webp', 'bmp', 'ico', 'icns', 'svgz',
  'pdf', 'zip', 'gz', 'tar', 'tgz', 'rar', '7z',
  'mp3', 'mp4', 'mov', 'avi', 'webm', 'wav', 'ogg', 'flac',
  'woff', 'woff2', 'ttf', 'otf', 'eot',
  'exe', 'dll', 'so', 'dylib', 'bin', 'class', 'o', 'a',
  'doc', 'docx', 'xls', 'xlsx', 'ppt', 'pptx',
]);

function extensionOf(filePath: string): string {
  const base = path.basename(filePath);
  const dot = base.lastIndexOf('.');
  return dot > 0 ? base.slice(dot + 1).toLowerCase() : '';
}

export function isBinaryPath(filePath: string): boolean {
  return BINARY_EXTENSIONS.has(extensionOf(filePath));
}

// Resolve a (possibly relative) child path against the project root, refusing to
// escape it. Returns the absolute path, or null if the resolved path would step
// outside the root (defence against path traversal from the renderer).
export function resolveWithinRoot(root: string, relativePath: string): string | null {
  const absRoot = path.resolve(root);
  const candidate = path.resolve(absRoot, relativePath);
  const rel = path.relative(absRoot, candidate);
  if (rel === '') return absRoot;
  if (rel.startsWith('..') || path.isAbsolute(rel)) return null;
  return candidate;
}

function toPosix(p: string): string {
  return p.split(path.sep).join('/');
}

// Recursively read a directory into a sorted tree (directories first, then files,
// each alphabetically). Hidden ignore-listed entries are skipped.
function readDirTree(absRoot: string, absDir: string, depth: number): FileTreeNode[] {
  if (depth > MAX_DEPTH) return [];

  let entries: fs.Dirent[];
  try {
    entries = fs.readdirSync(absDir, { withFileTypes: true });
  } catch {
    return [];
  }

  const nodes: FileTreeNode[] = [];
  for (const entry of entries.slice(0, MAX_ENTRIES_PER_DIR)) {
    if (IGNORED.has(entry.name)) continue;
    const absChild = path.join(absDir, entry.name);
    const relChild = toPosix(path.relative(absRoot, absChild));

    if (entry.isDirectory()) {
      nodes.push({
        name: entry.name,
        path: relChild,
        type: 'directory',
        children: readDirTree(absRoot, absChild, depth + 1),
      });
    } else if (entry.isFile()) {
      let size = 0;
      try {
        size = fs.statSync(absChild).size;
      } catch {
        // ignore — keep the node with an unknown size
      }
      nodes.push({ name: entry.name, path: relChild, type: 'file', size });
    }
    // Symlinks and special files are intentionally skipped.
  }

  nodes.sort((a, b) => {
    if (a.type !== b.type) return a.type === 'directory' ? -1 : 1;
    return a.name.localeCompare(b.name);
  });
  return nodes;
}

// Build the file tree for a project root. Throws if the root is not a directory.
export function readProjectTree(root: string): FileTreeNode {
  const absRoot = path.resolve(root);
  const stat = fs.statSync(absRoot); // throws if missing
  if (!stat.isDirectory()) {
    throw new Error(`Not a directory: ${absRoot}`);
  }
  return {
    name: path.basename(absRoot) || absRoot,
    path: '',
    type: 'directory',
    children: readDirTree(absRoot, absRoot, 0),
  };
}

// Read a single file inside the project root. Large files are truncated; binary
// files are returned base64-encoded so the renderer can still preview them.
export function readProjectFile(root: string, relativePath: string): ReadFileResult {
  const abs = resolveWithinRoot(root, relativePath);
  if (!abs) {
    throw new Error(`Path escapes project root: ${relativePath}`);
  }

  const stat = fs.statSync(abs);
  if (!stat.isFile()) {
    throw new Error(`Not a file: ${relativePath}`);
  }

  const binary = isBinaryPath(abs);
  const truncated = stat.size > MAX_FILE_BYTES;
  const bytesToRead = truncated ? MAX_FILE_BYTES : stat.size;

  const fd = fs.openSync(abs, 'r');
  try {
    const buffer = Buffer.alloc(bytesToRead);
    fs.readSync(fd, buffer, 0, bytesToRead, 0);
    return {
      path: toPosix(relativePath),
      content: binary ? buffer.toString('base64') : buffer.toString('utf8'),
      encoding: binary ? 'base64' : 'utf8',
      size: stat.size,
      truncated,
      binary,
    };
  } finally {
    fs.closeSync(fd);
  }
}
