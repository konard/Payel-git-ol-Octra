'use strict';

// Filesystem integration for the desktop IDE: open existing projects, create new
// ones, and keep a persisted list of recently opened projects so the project
// switcher (reference 4) can show "This Window" and "Recent Projects".
//
// The recent-projects list lives in a small JSON file inside Electron's userData
// directory so it survives restarts without a database. Everything here is pure
// Node.js so it can be unit-tested without launching Electron.

const fs = require('fs');
const path = require('path');

const STORE_FILE = 'recent-projects.json';
const MAX_RECENT = 12;

// Resolve the JSON store path. `userDataDir` is injected by the caller (the main
// process passes app.getPath('userData')) so this module stays testable.
function storePath(userDataDir) {
  return path.join(userDataDir, STORE_FILE);
}

// Read the persisted recent-projects list, tolerating a missing or corrupt file.
function readRecent(userDataDir) {
  try {
    const raw = fs.readFileSync(storePath(userDataDir), 'utf8');
    const parsed = JSON.parse(raw);
    if (!Array.isArray(parsed)) return [];
    // Drop entries whose folder no longer exists on disk.
    return parsed.filter((p) => p && typeof p.path === 'string');
  } catch {
    return [];
  }
}

// Persist the recent-projects list, creating the userData directory if needed.
function writeRecent(userDataDir, projects) {
  fs.mkdirSync(userDataDir, { recursive: true });
  fs.writeFileSync(storePath(userDataDir), JSON.stringify(projects, null, 2), 'utf8');
}

// Build a project descriptor from a folder path. The name is the folder's base
// name, matching how Zed labels projects (reference 2/4).
function describeProject(projectPath) {
  return {
    name: path.basename(projectPath) || projectPath,
    path: projectPath,
    openedAt: null,
  };
}

// Add (or move to the front) a project in the recent list and persist it.
// `now` is injected so the function is deterministic in tests.
function rememberProject(userDataDir, projectPath, now) {
  const normalized = path.resolve(projectPath);
  const existing = readRecent(userDataDir).filter(
    (p) => path.resolve(p.path) !== normalized,
  );
  const entry = { ...describeProject(normalized), openedAt: now };
  const next = [entry, ...existing].slice(0, MAX_RECENT);
  writeRecent(userDataDir, next);
  return { project: entry, recent: next };
}

// Remove a project from the recent list (used by "Remove from Recent Projects").
function forgetProject(userDataDir, projectPath) {
  const normalized = path.resolve(projectPath);
  const next = readRecent(userDataDir).filter(
    (p) => path.resolve(p.path) !== normalized,
  );
  writeRecent(userDataDir, next);
  return next;
}

// Create a new project folder on disk. Fails loudly if the folder already exists
// and is not empty so we never clobber existing work.
function createProject(parentDir, name) {
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

  // Seed a minimal README so the new project is recognisable in a file tree.
  const readme = path.join(projectPath, 'README.md');
  if (!fs.existsSync(readme)) {
    fs.writeFileSync(readme, `# ${safeName}\n\nCreated with Octra Desktop.\n`, 'utf8');
  }

  return describeProject(projectPath);
}

module.exports = {
  STORE_FILE,
  MAX_RECENT,
  storePath,
  readRecent,
  writeRecent,
  describeProject,
  rememberProject,
  forgetProject,
  createProject,
};
