import { create } from 'zustand';
import { persist } from 'zustand/middleware';

export type IntegrationType = 'lefine' | 'telegram' | 'n8n' | 'github';

export interface IntegrationConfig {
  useDefaultKey: boolean;
  apiKey: string;
  workflowId: string;
  // Lefine.pro specific
  activityPubUrl?: string;
  outboxEndpoint?: string;
  inboxEndpoint?: string;
  // GitHub specific
  publishNewProjects?: boolean;
  createPullRequests?: boolean;
}

export interface Integration {
  type: IntegrationType;
  connected: boolean;
  config: IntegrationConfig;
  connectedAt?: string;
}

interface IntegrationState {
  integrations: Record<IntegrationType, Integration>;
  
  // Actions
  setIntegrationConnected: (type: IntegrationType, connected: boolean, config?: Partial<IntegrationConfig>) => void;
  setIntegrationConfig: (type: IntegrationType, config: IntegrationConfig) => void;
  getIntegration: (type: IntegrationType) => Integration;
  disconnectIntegration: (type: IntegrationType) => void;
}

const DEFAULT_INTEGRATION: Integration = {
  type: 'lefine',
  connected: false,
  config: {
    useDefaultKey: true,
    apiKey: '',
    workflowId: '',
  },
};

const DEFAULT_TELEGRAM_INTEGRATION: Integration = {
  type: 'telegram',
  connected: false,
  config: {
    useDefaultKey: true,
    apiKey: '',
    workflowId: '',
  },
};

const DEFAULT_N8N_INTEGRATION: Integration = {
  type: 'n8n',
  connected: false,
  config: {
    useDefaultKey: false,
    apiKey: '',
    workflowId: '',
  },
};

const DEFAULT_GITHUB_INTEGRATION: Integration = {
  type: 'github',
  connected: false,
  config: {
    useDefaultKey: false,
    apiKey: '',
    workflowId: '',
    publishNewProjects: false,
    createPullRequests: false,
  },
};

export const useIntegrationStore = create<IntegrationState>()(
  persist(
    (set, get) => ({
      integrations: {
        lefine: DEFAULT_INTEGRATION,
        telegram: DEFAULT_TELEGRAM_INTEGRATION,
        n8n: DEFAULT_N8N_INTEGRATION,
        github: DEFAULT_GITHUB_INTEGRATION,
      },

      setIntegrationConnected: (type, connected, config = {}) => {
        set((state) => {
          const currentIntegration = state.integrations[type] || {
            type,
            connected: false,
            config: {
              useDefaultKey: true,
              apiKey: '',
              workflowId: '',
            },
          };
          return {
            integrations: {
              ...state.integrations,
              [type]: {
                ...currentIntegration,
                type,
                connected,
                config: {
                  ...currentIntegration.config,
                  ...config,
                },
                connectedAt: connected ? new Date().toISOString() : undefined,
              },
            },
          };
        });
      },

      setIntegrationConfig: (type, config) => {
        set((state) => {
          const currentIntegration = state.integrations[type] || {
            type,
            connected: false,
            config: {
              useDefaultKey: true,
              apiKey: '',
              workflowId: '',
            },
          };
          return {
            integrations: {
              ...state.integrations,
              [type]: {
                ...currentIntegration,
                config,
              },
            },
          };
        });
      },

      getIntegration: (type) => {
        const integration = get().integrations[type];
        if (!integration) {
          return {
            type,
            connected: false,
            config: {
              useDefaultKey: true,
              apiKey: '',
              workflowId: '',
            },
          };
        }
        return integration;
      },

      disconnectIntegration: (type) => {
        set((state) => {
          const currentIntegration = state.integrations[type] || {
            type,
            connected: false,
            config: {
              useDefaultKey: true,
              apiKey: '',
              workflowId: '',
            },
          };
          return {
            integrations: {
              ...state.integrations,
              [type]: {
                ...currentIntegration,
                connected: false,
                config: {
                  useDefaultKey: true,
                  apiKey: '',
                  workflowId: '',
                },
                connectedAt: undefined,
              },
            },
          };
        });
      },
    }),
    {
      name: 'crewai-integrations',
    }
  )
);
