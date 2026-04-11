/**
 * Workflow Service
 * API сервис для работы с библиотекой workflow (использует fetch)
 */

const API_URL = import.meta.env.VITE_AUTH_URL || 'http://localhost:3112';

export interface WorkflowNode {
  id: string;
  type: string;
  role: string;
  position?: { x: number; y: number };
  [key: string]: any;
}

export interface WorkflowEdge {
  from: string;
  to: string;
}

export interface Workflow {
  id: string;
  created_at: string;
  updated_at: string;
  user_id: string;
  author_name: string;
  name: string;
  description: string;
  category: string;
  tags: string[];
  nodes: string; // JSON string
  edges: string; // JSON string
  is_public: boolean;
  downloads: number;
}

export interface CreateWorkflowRequest {
  name: string;
  description?: string;
  category?: string;
  tags?: string[];
  nodes: string; // JSON string
  edges: string; // JSON string
  is_public?: boolean;
}

export interface UpdateWorkflowRequest {
  name?: string;
  description?: string;
  category?: string;
  tags?: string[];
  is_public?: boolean;
}

// Get auth token from localStorage
function getAuthToken(): string | null {
  return localStorage.getItem('access_token');
}

function getAuthHeaders(): Record<string, string> {
  const token = getAuthToken();
  return token ? { Authorization: `Bearer ${token}` } : {};
}

async function handleResponse<T>(response: Response): Promise<T> {
  if (!response.ok) {
    const errorData = await response.json().catch(() => ({ error: 'Unknown error' }));
    throw new Error(errorData.error || `HTTP ${response.status}: ${response.statusText}`);
  }
  const data = await response.json();
  return data.data as T;
}

/**
 * Get public workflows from the library
 */
export async function getPublicWorkflows(category?: string, tag?: string): Promise<Workflow[]> {
  const params = new URLSearchParams();
  if (category) params.append('category', category);
  if (tag) params.append('tag', tag);

  const response = await fetch(`${API_URL}/workflows/library?${params.toString()}`);
  return handleResponse<Workflow[]>(response);
}

/**
 * Get current user's workflows
 */
export async function getMyWorkflows(): Promise<Workflow[]> {
  const response = await fetch(`${API_URL}/workflows/my`, {
    headers: getAuthHeaders(),
  });
  return handleResponse<Workflow[]>(response);
}

/**
 * Get a specific workflow by ID
 */
export async function getWorkflowById(id: string): Promise<Workflow> {
  const response = await fetch(`${API_URL}/workflows/${id}`);
  return handleResponse<Workflow>(response);
}

/**
 * Create a new workflow
 */
export async function createWorkflow(data: CreateWorkflowRequest): Promise<Workflow> {
  const response = await fetch(`${API_URL}/workflows`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      ...getAuthHeaders(),
    },
    body: JSON.stringify(data),
  });
  return handleResponse<Workflow>(response);
}

/**
 * Update an existing workflow
 */
export async function updateWorkflow(id: string, data: UpdateWorkflowRequest): Promise<Workflow> {
  const response = await fetch(`${API_URL}/workflows/${id}`, {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json',
      ...getAuthHeaders(),
    },
    body: JSON.stringify(data),
  });
  return handleResponse<Workflow>(response);
}

/**
 * Delete a workflow
 */
export async function deleteWorkflow(id: string): Promise<void> {
  const response = await fetch(`${API_URL}/workflows/${id}`, {
    method: 'DELETE',
    headers: getAuthHeaders(),
  });
  await handleResponse(response);
}

/**
 * Register a download for a workflow
 */
export async function downloadWorkflow(id: string): Promise<void> {
  const response = await fetch(`${API_URL}/workflows/${id}/download`, {
    method: 'POST',
  });
  await handleResponse(response);
}

/**
 * Get all workflow categories
 */
export async function getWorkflowCategories(): Promise<string[]> {
  const response = await fetch(`${API_URL}/workflows/categories`);
  return handleResponse<string[]>(response);
}
