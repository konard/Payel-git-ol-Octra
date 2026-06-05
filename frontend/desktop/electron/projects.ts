'use strict';

// Filesystem integration for the desktop IDE: open existing projects, create new
// ones, and keep a persisted list of recently opened projects so the welcome
// screen and the project switcher can show "Recent Projects".
//
// The recent-projects list lives in a small JSON file inside Electron's userData
// directory so it survives restarts without a database. Everything here is pure
// Node.js so it can be unit-tested without launching Electron.

import * as fs from 'fs';
import * as path from 'path';

export interface Project {
  name: string;
  path: string;
  openedAt: number | null;
}

export interface RememberResult {
  project: Project;
  recent: Project[];
}

export const STORE_FILE = 'recent-projects.json';
export const MAX_RECENT = 12;

// Resolve the JSON store path. `userDataDir` is injected by the caller (the main
// process passes app.getPath('userData')) so this module stays testable.
function storePath(userDataDir: string): string {
  return path.join(userDataDir, STORE_FILE);
}

// Read the persisted recent-projects list, tolerating a missing or corrupt file.
export function readRecent(userDataDir: string): Project[] {
  try {
    const raw = fs.readFileSync(storePath(userDataDir), 'utf8');
    const parsed = JSON.parse(raw);
    if (!Array.isArray(parsed)) return [];
    // Drop entries whose folder no longer exists on disk so the list never shows
    // projects that have since been deleted or moved.
    return parsed
      .filter((p): p is Project => Boolean(p) && typeof p.path === 'string')
      .filter((p) => {
        try {
          return fs.statSync(p.path).isDirectory();
        } catch {
          return false;
        }
      });
  } catch {
    return [];
  }
}

// Persist the recent-projects list, creating the userData directory if needed.
export function writeRecent(userDataDir: string, projects: Project[]): void {
  fs.mkdirSync(userDataDir, { recursive: true });
  fs.writeFileSync(storePath(userDataDir), JSON.stringify(projects, null, 2), 'utf8');
}

// Build a project descriptor from a folder path. The name is the folder's base
// name, matching how editors label projects in their title bar.
export function describeProject(projectPath: string): Project {
  return {
    name: path.basename(projectPath) || projectPath,
    path: projectPath,
    openedAt: null,
  };
}

// Add (or move to the front) a project in the recent list and persist it.
// `now` is injected so the function is deterministic in tests.
export function rememberProject(
  userDataDir: string,
  projectPath: string,
  now: number,
): RememberResult {
  const normalized = path.resolve(projectPath);
  const existing = readRecent(userDataDir).filter(
    (p) => path.resolve(p.path) !== normalized,
  );
  const entry: Project = { ...describeProject(normalized), openedAt: now };
  const next = [entry, ...existing].slice(0, MAX_RECENT);
  writeRecent(userDataDir, next);
  return { project: entry, recent: next };
}

// Remove a project from the recent list (used by "Remove from Recent Projects").
export function forgetProject(userDataDir: string, projectPath: string): Project[] {
  const normalized = path.resolve(projectPath);
  const next = readRecent(userDataDir).filter(
    (p) => path.resolve(p.path) !== normalized,
  );
  writeRecent(userDataDir, next);
  return next;
}

// Create a new project folder on disk. Fails loudly if the folder already exists
// and is not empty so we never clobber existing work.
export function createProject(parentDir: string, name: string): Project {
  const safeName = String(name || '').trim();
  if (!safeName) throw new Error('Project name is required');
  if (/[\\/]/.test(safeName)) throw new Error('Project name cannot contain path separators');

  const projectPath = path.join(parentDir, safeName);
  if (fs.existsSync(projectPath)) {
    const entries = fs.readdirSync(projectPath);
    if (entries.length > 0) {
      throw new Error(`Folder "${safeName}" already exists and is not empty`);
    }
  } else {
    fs.mkdirSync(projectPath, { recursive: true });
  }

  // Seed a minimal README so the new project is recognisable in the file tree.
  const readme = path.join(projectPath, 'README.md');
  if (!fs.existsSync(readme)) {
    fs.writeFileSync(readme, `# ${safeName}\n\nCreated with Octra Desktop.\n`, 'utf8');
  }

  return describeProject(projectPath);
}
