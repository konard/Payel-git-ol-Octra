import { IntegrationConfig } from '../stores/integrationStore';

export interface N8nWorkflow {
  id: string;
  name: string;
  active: boolean;
  createdAt: string;
  updatedAt: string;
  nodes: any[];
  connections: any;
}

export interface N8nWebhook {
  webhookId: string;
  path: string;
  method: string;
  node: string;
}

class N8nService {
  private getHeaders(config: IntegrationConfig) {
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
    };

    if (config.apiKey) {
      headers['X-N8N-API-KEY'] = config.apiKey;
    }

    return headers;
  }

  private getBaseUrl(config: IntegrationConfig): string {
    // Remove /mcp-server/http from the end if present to get base API URL
    const serverUrl = config.activityPubUrl || '';
    return serverUrl.replace('/mcp-server/http', '').replace('/mcp-server', '');
  }

  async getWorkflows(config: IntegrationConfig): Promise<N8nWorkflow[]> {
    try {
      const baseUrl = this.getBaseUrl(config);
      const response = await fetch(`${baseUrl}/api/v1/workflows`, {
        method: 'GET',
        headers: this.getHeaders(config),
      });

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }

      const data = await response.json();
      return data.data || [];
    } catch (error) {
      console.error('Failed to fetch n8n workflows:', error);
      throw error;
    }
  }

  async executeWorkflow(workflowId: string, config: IntegrationConfig, data?: any): Promise<any> {
    try {
      const baseUrl = this.getBaseUrl(config);
      const response = await fetch(`${baseUrl}/api/v1/workflows/${workflowId}/execute`, {
        method: 'POST',
        headers: this.getHeaders(config),
        body: JSON.stringify(data || {}),
      });

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }

      return await response.json();
    } catch (error) {
      console.error('Failed to execute n8n workflow:', error);
      throw error;
    }
  }

  async triggerWebhook(webhookUrl: string, data?: any): Promise<any> {
    try {
      const response = await fetch(webhookUrl, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(data || {}),
      });

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }

      return await response.json();
    } catch (error) {
      console.error('Failed to trigger n8n webhook:', error);
      throw error;
    }
  }

  async testConnection(config: IntegrationConfig): Promise<boolean> {
    try {
      const baseUrl = this.getBaseUrl(config);
      const response = await fetch(`${baseUrl}/api/v1/workflows`, {
        method: 'GET',
        headers: this.getHeaders(config),
      });

      return response.ok;
    } catch (error) {
      console.error('N8n connection test failed:', error);
      return false;
    }
  }
}

export const n8nService = new N8nService();