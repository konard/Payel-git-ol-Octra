import { CustomProvider, CustomModel } from '../stores/customProvidersStore';

const API_BASE_URL = import.meta.env.VITE_API_URL || '/auth';

export interface CreateCustomProviderRequest {
  name: string;
  base_url: string;
  api_key: string;
  requires_api_key: boolean;
}

export interface UpdateCustomProviderRequest {
  name?: string;
  base_url?: string;
  api_key?: string;
  requires_api_key?: boolean;
}

export interface CreateCustomModelRequest {
  name: string;
  provider_id?: string;
}

export interface UpdateCustomModelRequest {
  name?: string;
  provider_id?: string;
}

class CustomProviderService {
  private getAuthHeaders(): HeadersInit {
    const token = localStorage.getItem('access_token');
    const refreshToken = localStorage.getItem('refresh_token');
    console.log('Auth tokens - access:', token ? 'present' : 'missing', 'refresh:', refreshToken ? 'present' : 'missing');
    return {
      'Content-Type': 'application/json',
      'Authorization': token ? `Bearer ${token}` : '',
    };
  }

  async getUserCustomProviders(): Promise<CustomProvider[]> {
    const response = await fetch(`${API_BASE_URL}/custom-providers`, {
      headers: this.getAuthHeaders(),
    });

    if (!response.ok) {
      throw new Error('Failed to fetch custom providers');
    }

    const data = await response.json();
    return data.data || [];
  }

  async createCustomProvider(provider: CreateCustomProviderRequest): Promise<CustomProvider> {
    const response = await fetch(`${API_BASE_URL}/custom-providers`, {
      method: 'POST',
      headers: this.getAuthHeaders(),
      body: JSON.stringify(provider),
    });

    if (!response.ok) {
      throw new Error('Failed to create custom provider');
    }

    const data = await response.json();
    return data.data;
  }

  async updateCustomProvider(id: string, updates: UpdateCustomProviderRequest): Promise<CustomProvider> {
    const response = await fetch(`${API_BASE_URL}/custom-providers/${id}`, {
      method: 'PUT',
      headers: this.getAuthHeaders(),
      body: JSON.stringify(updates),
    });

    if (!response.ok) {
      throw new Error('Failed to update custom provider');
    }

    const data = await response.json();
    return data.data;
  }

  async deleteCustomProvider(id: string): Promise<void> {
    const response = await fetch(`${API_BASE_URL}/custom-providers/${id}`, {
      method: 'DELETE',
      headers: this.getAuthHeaders(),
    });

    if (!response.ok) {
      throw new Error('Failed to delete custom provider');
    }
  }

  // Custom Models
  async getUserCustomModels(): Promise<CustomModel[]> {
    const response = await fetch(`${API_BASE_URL}/custom-models`, {
      headers: this.getAuthHeaders(),
    });

    if (!response.ok) {
      throw new Error('Failed to fetch custom models');
    }

    const data = await response.json();
    return data.data || [];
  }

  async createCustomModel(model: CreateCustomModelRequest): Promise<CustomModel> {
    const response = await fetch(`${API_BASE_URL}/custom-models`, {
      method: 'POST',
      headers: this.getAuthHeaders(),
      body: JSON.stringify(model),
    });

    if (!response.ok) {
      throw new Error('Failed to create custom model');
    }

    const data = await response.json();
    return data.data;
  }

  async updateCustomModel(id: string, updates: UpdateCustomModelRequest): Promise<CustomModel> {
    const response = await fetch(`${API_BASE_URL}/custom-models/${id}`, {
      method: 'PUT',
      headers: this.getAuthHeaders(),
      body: JSON.stringify(updates),
    });

    if (!response.ok) {
      throw new Error('Failed to update custom model');
    }

    const data = await response.json();
    return data.data;
  }

  async deleteCustomModel(id: string): Promise<void> {
    const response = await fetch(`${API_BASE_URL}/custom-models/${id}`, {
      method: 'DELETE',
      headers: this.getAuthHeaders(),
    });

    if (!response.ok) {
      throw new Error('Failed to delete custom model');
    }
  }
}

export const customProviderService = new CustomProviderService();