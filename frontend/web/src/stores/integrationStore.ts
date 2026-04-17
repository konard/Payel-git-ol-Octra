import { create } from 'zustand';
import { persist } from 'zustand/middleware';

export type IntegrationType = 'lefine' | 'telegram' | 'n8n';

export interface IntegrationConfig {
  useDefaultKey: boolean;
  apiKey: string;
  workflowId: string;
  // Lefine.pro specific
  activityPubUrl?: string;
  outboxEndpoint?: string;
  inboxEndpoint?: string;
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

export const useIntegrationStore = create<IntegrationState>()(
  persist(
    (set, get) => ({
      integrations: {
        lefine: DEFAULT_INTEGRATION,
        telegram: DEFAULT_TELEGRAM_INTEGRATION,
        n8n: DEFAULT_N8N_INTEGRATION,
      },

      setIntegrationConnected: (type, connected, config = {}) => {
        set((state) => ({
          integrations: {
            ...state.integrations,
            [type]: {
              ...state.integrations[type],
              type,
              connected,
              config: {
                ...state.integrations[type].config,
                ...config,
              },
              connectedAt: connected ? new Date().toISOString() : undefined,
            },
          },
        }));
      },

      setIntegrationConfig: (type, config) => {
        set((state) => ({
          integrations: {
            ...state.integrations,
            [type]: {
              ...state.integrations[type],
              config,
            },
          },
        }));
      },

      getIntegration: (type) => {
        return get().integrations[type];
      },

      disconnectIntegration: (type) => {
        set((state) => ({
          integrations: {
            ...state.integrations,
            [type]: {
              ...state.integrations[type],
              connected: false,
              config: {
                useDefaultKey: true,
                apiKey: '',
                workflowId: '',
              },
              connectedAt: undefined,
            },
          },
        }));
      },
    }),
    {
      name: 'crewai-integrations',
    }
  )
);
