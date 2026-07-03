import { API } from '../config/routes';

function authHeaders(): HeadersInit {
  const token = typeof window !== 'undefined'
    ? (window.localStorage.getItem('octra_access_token') ?? window.localStorage.getItem('access_token'))
    : '';
  return { 'octra-api-token': token ?? '' };
}

export type MCPServerSettings = {
  id?: string;
  transport?: 'http' | 'stdio';
  command?: string;
  args?: string[];
  env?: Record<string, string>;
  url?: string;
  bearer_token?: string;
  enabled?: boolean;
};

export type MCPServer = {
  id: string;
  transport: string;
  command?: string;
  args?: string[];
  env?: Record<string, string>;
  url?: string;
  bearer_token?: string;
  enabled: boolean;
  status: string;
  created_at: string;
  updated_at: string;
};

export type MCPTool = {
  name: string;
  description?: string;
  input_schema?: Record<string, unknown>;
};

export type MCPResource = {
  uri: string;
  name: string;
  description?: string;
};

export type MCPPrompt = {
  name: string;
  description?: string;
  arguments?: { name: string; description?: string; required?: boolean }[];
};

export type MCPCatalog = {
  servers: MCPServer[];
  tools: MCPTool[];
  resources: MCPResource[];
  prompts: MCPPrompt[];
};

export async function listMCPServers(): Promise<Response> {
  return fetch(API.MCP_SERVERS, { headers: authHeaders(), cache: 'no-cache' });
}

export async function getMCPServer(id: string): Promise<Response> {
  return fetch(`${API.MCP_SERVERS}/${id}`, { headers: authHeaders(), cache: 'no-cache' });
}

export async function createMCPServer(settings: MCPServerSettings): Promise<Response> {
  return fetch(API.MCP_SERVERS, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', ...authHeaders() },
    body: JSON.stringify(settings),
  });
}

export async function updateMCPServer(id: string, settings: Partial<MCPServerSettings>): Promise<Response> {
  return fetch(`${API.MCP_SERVERS}/${id}`, {
    method: 'PATCH',
    headers: { 'Content-Type': 'application/json', ...authHeaders() },
    body: JSON.stringify(settings),
  });
}

export async function deleteMCPServer(id: string): Promise<Response> {
  return fetch(`${API.MCP_SERVERS}/${id}`, {
    method: 'DELETE',
    headers: authHeaders(),
  });
}

export async function reconnectMCPServer(id: string): Promise<Response> {
  return fetch(`${API.MCP_SERVERS}/${id}/reconnect`, {
    method: 'POST',
    headers: authHeaders(),
  });
}

export async function listMCPCatalog(): Promise<Response> {
  return fetch(API.MCP_CATALOG, { headers: authHeaders(), cache: 'no-cache' });
}
