import { API } from '../config/routes';

function authHeaders(): HeadersInit {
  const token = typeof window !== 'undefined'
    ? (window.localStorage.getItem('octra_access_token') ?? window.localStorage.getItem('access_token'))
    : '';
  return { 'octra-api-token': token ?? '' };
}

export type HITLRun = {
  workflow_id: string;
  run_id: string;
  resume_labels: string[];
  suspend_payload?: Record<string, unknown>;
  updated_at: string;
};

export type HITLRunDetail = HITLRun & {
  status: string;
  state?: Record<string, unknown>;
};

export type ListHITLRunsResponse = {
  runs: HITLRun[];
};

export async function listHITLRuns(): Promise<Response> {
  return fetch(API.HITL_RUNS, { headers: authHeaders(), cache: 'no-cache' });
}

export async function getHITLRun(workflowId: string, runId: string): Promise<Response> {
  return fetch(API.HITL_RUN(workflowId, runId), { headers: authHeaders(), cache: 'no-cache' });
}

export async function resumeHITLRun(workflowId: string, runId: string, data?: Record<string, unknown>): Promise<Response> {
  return fetch(API.HITL_RUN_ACTIONS(workflowId, runId), {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', ...authHeaders() },
    body: JSON.stringify({ action: 'resume', resume_data: data ?? {} }),
  });
}

export async function cancelHITLRun(workflowId: string, runId: string): Promise<Response> {
  return fetch(API.HITL_RUN_ACTIONS(workflowId, runId), {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', ...authHeaders() },
    body: JSON.stringify({ action: 'cancel' }),
  });
}
