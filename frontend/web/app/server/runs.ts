import { API } from '../config/routes';

function authHeaders(): HeadersInit {
  const token = typeof window !== 'undefined'
    ? (window.localStorage.getItem('octra_access_token') ?? window.localStorage.getItem('access_token'))
    : '';
  return { 'octra-api-token': token ?? '' };
}

export type RunSummary = {
  run_id: string;
  workflow_id: string;
  status: string;
  updated_at: string;
};

export type RunSnapshot = {
  workflow_id: string;
  run_id: string;
  status: string;
  state?: Record<string, unknown>;
  output?: Record<string, unknown>;
  init_data?: Record<string, unknown>;
  node_results?: Record<string, unknown>;
  node_index?: number;
  resume_labels?: string[];
  suspend_payload?: Record<string, unknown>;
  error?: { type: string; message: string };
  resource_id?: string;
  updated_at: string;
};

export type ListRunsResponse = {
  runs: RunSummary[];
};

export async function listRuns(workflowId: string, status?: string): Promise<Response> {
  const params = status ? `?status=${encodeURIComponent(status)}` : '';
  return fetch(`${API.WORKFLOW_RUNS(workflowId)}${params}`, {
    headers: authHeaders(),
    cache: 'no-cache',
  });
}

export async function getRun(workflowId: string, runId: string): Promise<Response> {
  return fetch(API.WORKFLOW_RUN(workflowId, runId), {
    headers: authHeaders(),
    cache: 'no-cache',
  });
}

export async function startRun(workflowId: string, inputData?: Record<string, unknown>): Promise<Response> {
  return fetch(API.WORKFLOW_RUNS(workflowId), {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', ...authHeaders() },
    body: JSON.stringify(inputData ?? {}),
  });
}

export async function resumeRun(workflowId: string, runId: string, data?: Record<string, unknown>): Promise<Response> {
  return fetch(`${API.WORKFLOW_RUN(workflowId, runId)}/resume`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', ...authHeaders() },
    body: JSON.stringify(data ?? {}),
  });
}

export async function cancelRun(workflowId: string, runId: string): Promise<Response> {
  return fetch(`${API.WORKFLOW_RUN(workflowId, runId)}/cancel`, {
    method: 'POST',
    headers: authHeaders(),
  });
}

export async function restartRun(workflowId: string, runId: string): Promise<Response> {
  return fetch(`${API.WORKFLOW_RUN(workflowId, runId)}/restart`, {
    method: 'POST',
    headers: authHeaders(),
  });
}
