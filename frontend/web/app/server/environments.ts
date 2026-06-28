import { API } from '../config/routes';

function authHeaders(): HeadersInit {
  const token = window.localStorage.getItem('octra_access_token') ?? window.localStorage.getItem('access_token');
  return { 'octra-api-token': token ?? '' };
}

export type DashboardEnvironment = {
  id: string;
  name: string;
  visibility: 'private' | 'public';
  active: boolean;
  created_at: string;
};

export async function createDashboardEnvironment(name: string, visibility: string): Promise<Response> {
  return fetch(API.ENVIRONMENTS, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', ...authHeaders() },
    body: JSON.stringify({ name, visibility }),
  });
}

export async function listDashboardEnvironments(): Promise<Response> {
  return fetch(API.ENVIRONMENTS, { headers: authHeaders(), cache: 'no-cache' });
}

export async function patchDashboardEnvironment(id: string, changes: { active?: boolean; visibility?: string }): Promise<Response> {
  return fetch(`${API.ENVIRONMENTS}/patch`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', ...authHeaders() },
    body: JSON.stringify({ id, ...changes }),
  });
}

export async function deleteDashboardEnvironment(id: string): Promise<Response> {
  return fetch(`${API.ENVIRONMENTS}/${id}`, {
    method: 'DELETE',
    headers: authHeaders(),
  });
}

export type CanvasNodeRequest = {
  item_id: string;
  kind: string;
  name: string;
  detail?: string;
  description?: string;
  meta?: Record<string, string | null>;
  position_x: number;
  position_y: number;
  sort_order: number;
};

export type CanvasNodeResponse = {
  id: string;
  item_id: string;
  kind: string;
  name: string;
  detail?: string;
  description?: string;
  meta?: Record<string, string | null>;
  position_x: number;
  position_y: number;
  sort_order: number;
  created_at: string;
  updated_at: string;
};

export async function getCanvas(envId: string): Promise<Response> {
  return fetch(`${API.ENVIRONMENTS}/${envId}/canvas`, { headers: authHeaders(), cache: 'no-cache' });
}

export async function putCanvas(envId: string, nodes: CanvasNodeRequest[]): Promise<Response> {
  return fetch(`${API.ENVIRONMENTS}/${envId}/canvas`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json', ...authHeaders() },
    body: JSON.stringify({ nodes }),
  });
}
