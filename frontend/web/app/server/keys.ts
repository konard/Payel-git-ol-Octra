import { API } from '../config/routes'

function authHeaders(): HeadersInit {
  const token = window.localStorage.getItem('octra_access_token') ?? window.localStorage.getItem('access_token');
  return { Authorization: `Bearer ${token}` };
}

export async function createAPIKey(name: string, expiresAt: string | null): Promise<Response> {
  return fetch(API.KEYS, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', ...authHeaders() },
    body: JSON.stringify({ name, expires_at: expiresAt || null }),
  });
}

export async function listAPIKeys(): Promise<Response> {
  return fetch(API.KEYS, { headers: authHeaders() });
}

export async function deleteAPIKey(id: string): Promise<Response> {
  return fetch(`${API.KEYS}/${id}`, { method: 'DELETE', headers: authHeaders() });
}
