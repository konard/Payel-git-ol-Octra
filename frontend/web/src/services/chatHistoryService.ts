const AUTH_API_URL = import.meta.env.VITE_AUTH_URL || '';

export interface ChatHistoryItem {
  id: string;
  title: string;
  created_at: Date;
  updated_at: Date;
  last_message: string;
  workflow?: string;
  provider?: string;
  model?: string;
  messages?: ChatMessage[];
}

export interface ChatMessage {
  id: string;
  role: 'user' | 'boss';
  content: string;
  timestamp?: Date;
  created_at?: Date | string;
}

function getAuthHeaders() {
  const token = localStorage.getItem('access_token');
  return token ? { 'Authorization': `Bearer ${token}` } : {};
}

export async function getChatHistory(userId: string): Promise<ChatHistoryItem[]> {
  const response = await fetch(`${AUTH_API_URL}/chat/history?user_id=${userId}`, {
    method: 'GET',
    headers: { 
      'Content-Type': 'application/json',
      ...getAuthHeaders(),
    },
    credentials: 'include',
  });

  if (!response.ok) {
    throw new Error('Failed to fetch chat history');
  }

  const data = await response.json();
  return data.data || [];
}

export async function getChat(chatId: string): Promise<ChatHistoryItem> {
  const response = await fetch(`${AUTH_API_URL}/chat/${chatId}`, {
    method: 'GET',
    headers: {
      'Content-Type': 'application/json',
      ...getAuthHeaders(),
    },
    credentials: 'include',
  });

  if (!response.ok) {
    throw new Error('Failed to fetch chat');
  }

  const data = await response.json();
  return data.data;
}

async function authFetch(url: string, options: RequestInit = {}): Promise<Response> {
  const response = await fetch(url, { ...options, credentials: 'include' });
  if (response.status === 401) {
    const rt = localStorage.getItem('refresh_token');
    if (rt) {
      try {
        const { refreshToken: doRefresh } = await import('./authService');
        const res = await doRefresh(rt);
        localStorage.setItem('access_token', res.data.access_token);
        localStorage.setItem('refresh_token', res.data.refresh_token);
        document.cookie = `access_token=${res.data.access_token}; path=/; SameSite=Lax`;
        const { headers: _, ...rest } = options;
        return fetch(url, {
          ...rest,
          headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${res.data.access_token}`,
          },
          credentials: 'include',
        });
      } catch {
        // refresh failed, return original 401
      }
    }
  }
  return response;
}

export async function createChat(userId: string, title: string): Promise<ChatHistoryItem> {
  const response = await authFetch(`${AUTH_API_URL}/chat/create`, {
    method: 'POST',
    headers: { 
      'Content-Type': 'application/json',
      ...getAuthHeaders(),
    },
    body: JSON.stringify({ user_id: userId, title }),
  });

  if (!response.ok) {
    throw new Error('Failed to create chat');
  }

  const data = await response.json();
  return data.data;
}

export async function updateChatTitle(chatId: string, title: string): Promise<void> {
  const response = await fetch(`${AUTH_API_URL}/chat/${chatId}/title`, {
    method: 'PUT',
    headers: { 
      'Content-Type': 'application/json',
      ...getAuthHeaders(),
    },
    credentials: 'include',
    body: JSON.stringify({ title }),
  });

  if (!response.ok) {
    throw new Error('Failed to update chat title');
  }
}

export async function addMessage(
  chatId: string,
  role: 'user' | 'boss',
  content: string,
  metadata?: {
    provider?: string;
    model?: string;
    apiKey?: string;
  }
): Promise<ChatMessage> {
  const response = await fetch(`${AUTH_API_URL}/chat/${chatId}/messages`, {
    method: 'POST',
    headers: { 
      'Content-Type': 'application/json',
      ...getAuthHeaders(),
    },
    credentials: 'include',
    body: JSON.stringify({ role, content, ...metadata }),
  });

  if (!response.ok) {
    throw new Error('Failed to add message');
  }

  const data = await response.json();
  return data.data;
}

export async function updateChatWorkflow(
  chatId: string,
  workflow: string
): Promise<void> {
  const response = await fetch(`${AUTH_API_URL}/chat/${chatId}/workflow`, {
    method: 'PUT',
    headers: { 
      'Content-Type': 'application/json',
      ...getAuthHeaders(),
    },
    credentials: 'include',
    body: JSON.stringify({ workflow }),
  });

  if (!response.ok) {
    throw new Error(`Failed to update workflow: ${response.status}`);
  }
}

export async function deleteChat(chatId: string): Promise<void> {
  const response = await fetch(`${AUTH_API_URL}/chat/${chatId}`, {
    method: 'DELETE',
    headers: { 
      'Content-Type': 'application/json',
      ...getAuthHeaders(),
    },
    credentials: 'include',
  });

  if (!response.ok) {
    throw new Error('Failed to delete chat');
  }
}
