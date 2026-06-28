export type EnvironmentRecord = {
  id: string;
  name: string;
  endpoint: string;
  cliState: string;
  runtime: string;
  active: boolean;
  updatedAt: string;
};

type StorageAdapter = Pick<Storage, 'getItem' | 'setItem' | 'removeItem'>;

export const ENVIRONMENTS_STORAGE_KEY = 'octra_environments';

function stringValue(value: unknown, fallback: string): string {
  return typeof value === 'string' && value.trim() ? value.trim() : fallback;
}

function boolValue(value: unknown, status: unknown): boolean {
  if (typeof value === 'boolean') return value;
  if (typeof status === 'string') return status.toLowerCase() !== 'paused';
  return true;
}

export function normalizeEnvironment(value: unknown, index = 0): EnvironmentRecord | null {
  if (!value || typeof value !== 'object') return null;

  const record = value as Record<string, unknown>;
  const active = boolValue(record.active, record.status);
  const fallbackName = `Environment ${index + 1}`;

  return {
    id: stringValue(record.id ?? record.environment_id, `environment-${index + 1}`),
    name: stringValue(record.name ?? record.title, fallbackName),
    endpoint: stringValue(record.endpoint, '/api/chat'),
    cliState: stringValue(record.cliState ?? record.cli_state, active ? 'ready' : 'paused'),
    runtime: stringValue(record.runtime, 'nix'),
    active,
    updatedAt: stringValue(record.updatedAt ?? record.updated_at, new Date(0).toISOString()),
  };
}

export function readEnvironments(storage?: StorageAdapter | null): EnvironmentRecord[] {
  if (!storage) return [];

  const raw = storage.getItem(ENVIRONMENTS_STORAGE_KEY);
  if (!raw) return [];

  try {
    const parsed = JSON.parse(raw);
    if (!Array.isArray(parsed)) return [];
    return parsed
      .map((item, index) => normalizeEnvironment(item, index))
      .filter((item): item is EnvironmentRecord => item !== null);
  } catch {
    return [];
  }
}

export function writeEnvironments(storage: StorageAdapter | undefined | null, environments: EnvironmentRecord[]) {
  if (!storage) return;
  storage.setItem(ENVIRONMENTS_STORAGE_KEY, JSON.stringify(environments));
}

export function listActiveEnvironments(environments: EnvironmentRecord[]): EnvironmentRecord[] {
  return environments.filter((environment) => environment.active);
}

export function pauseEnvironment(environments: EnvironmentRecord[], id: string): EnvironmentRecord[] {
  const updatedAt = new Date().toISOString();
  return environments.map((environment) => (
    environment.id === id
      ? { ...environment, active: false, cliState: 'paused', updatedAt }
      : environment
  ));
}

export function startEnvironment(environments: EnvironmentRecord[], id: string): EnvironmentRecord[] {
  const updatedAt = new Date().toISOString();
  return environments.map((environment) => (
    environment.id === id
      ? {
          ...environment,
          active: true,
          cliState: environment.cliState === 'paused' ? 'ready' : environment.cliState,
          updatedAt,
        }
      : environment
  ));
}

export function deleteEnvironment(environments: EnvironmentRecord[], id: string): EnvironmentRecord[] {
  return environments.filter((environment) => environment.id !== id);
}
