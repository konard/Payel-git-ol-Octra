import fs from 'node:fs';
import path from 'node:path';

const root = process.cwd();
const helperPath = path.join(root, 'app/lib/environments.ts');

function assert(condition, message) {
  if (!condition) {
    throw new Error(message);
  }
}

class MemoryStorage {
  #items = new Map();

  getItem(key) {
    return this.#items.has(key) ? this.#items.get(key) : null;
  }

  setItem(key, value) {
    this.#items.set(key, String(value));
  }

  removeItem(key) {
    this.#items.delete(key);
  }
}

async function importHelper() {
  assert(fs.existsSync(helperPath), 'Missing app/lib/environments.ts helper');
  const ts = await import('typescript');
  const source = fs.readFileSync(helperPath, 'utf8');
  const compiled = ts.default.transpileModule(source, {
    compilerOptions: {
      module: ts.default.ModuleKind.ES2022,
      target: ts.default.ScriptTarget.ES2022,
      isolatedModules: true,
    },
    fileName: helperPath,
  }).outputText;
  const url = `data:text/javascript;base64,${Buffer.from(compiled).toString('base64')}`;
  return import(url);
}

const {
  ENVIRONMENTS_STORAGE_KEY,
  deleteEnvironment,
  listActiveEnvironments,
  pauseEnvironment,
  readEnvironments,
  startEnvironment,
  writeEnvironments,
} = await importHelper();

const storage = new MemoryStorage();
writeEnvironments(storage, [
  {
    id: 'env-active',
    name: 'Active test',
    endpoint: '/api/chat',
    cliState: 'ready',
    runtime: 'nix',
    active: true,
    updatedAt: '2026-06-27T20:00:00.000Z',
  },
  {
    id: 'env-paused',
    name: 'Paused test',
    endpoint: '/api/chat',
    cliState: 'paused',
    runtime: 'nix',
    active: false,
    updatedAt: '2026-06-27T20:01:00.000Z',
  },
]);

assert(storage.getItem(ENVIRONMENTS_STORAGE_KEY), 'Environment state must be persisted to storage');
assert(readEnvironments(storage).length === 2, 'Dashboard view must retain active and paused environments');
assert(listActiveEnvironments(readEnvironments(storage)).map((env) => env.id).join(',') === 'env-active', '/app must only list active environments');

writeEnvironments(storage, pauseEnvironment(readEnvironments(storage), 'env-active'));
const afterReload = readEnvironments(storage);
assert(afterReload.find((env) => env.id === 'env-active')?.active === false, 'Paused state must survive a reload');
assert(listActiveEnvironments(afterReload).length === 0, 'Paused environments must not reappear in /app after refresh');

writeEnvironments(storage, startEnvironment(afterReload, 'env-active'));
assert(listActiveEnvironments(readEnvironments(storage)).map((env) => env.id).join(',') === 'env-active', 'Starting from dashboard must restore the environment to /app');

storage.setItem(ENVIRONMENTS_STORAGE_KEY, JSON.stringify([{ id: 'legacy', name: 'Legacy env' }]));
assert(readEnvironments(storage)[0].active === true, 'Legacy environment records without active must default to active');

writeEnvironments(storage, deleteEnvironment(readEnvironments(storage), 'legacy'));
assert(readEnvironments(storage).length === 0, 'Deleting an environment must remove it from dashboard storage');

console.log('Environment active/paused persistence rules are correct.');
