import { resolveDefaultToken } from '../config/defaultSettings';
import { isGeneratedAgentNodeId, type AgentNode, type Edge, type WorkflowConfig } from '../stores/taskStore';

interface CustomProviderAuth {
  base_url: string;
  api_key?: string | null;
}

interface BuildTaskProviderAuthInput {
  provider: string;
  apiKey?: string | null;
  defaultToken?: string | null;
  customProvider?: CustomProviderAuth | null;
}

interface TaskProviderAuth {
  provider: string;
  tokens: Record<string, string>;
}

const isCustomWorkflowNode = (node: AgentNode): boolean => !isGeneratedAgentNodeId(node.id);

function firstConfiguredToken(...tokens: Array<string | null | undefined>): string {
  for (const token of tokens) {
    const trimmedToken = token?.trim();
    if (trimmedToken) {
      return trimmedToken;
    }
  }

  return resolveDefaultToken();
}

export function buildWorkflowConfigFromGraph(nodes: AgentNode[], edges: Edge[]): WorkflowConfig | null {
  const customNodes = nodes.filter(isCustomWorkflowNode);
  const customNodeIds = new Set(customNodes.map((node) => node.id));
  const customEdges = edges.filter((edge) => customNodeIds.has(edge.from) && customNodeIds.has(edge.to));
  const bossNodes = customNodes.filter((node) => node.type === 'boss');
  const managerNodes = customNodes.filter((node) => node.type === 'manager');
  const workerNodes = customNodes.filter((node) => node.type === 'worker');

  if (managerNodes.length === 0) {
    if (workerNodes.length === 0) {
      return null;
    }

    return {
      useAiPlanning: false,
      managers: [{
        role: 'Direct Worker',
        description: 'Run selected workers directly for this task',
        priority: 1,
        workers: workerNodes.map((worker) => ({
          role: worker.role || 'Worker',
          description: `${worker.role || 'Worker'} worker`,
        })),
      }],
      architecture: `Custom workflow with ${workerNodes.length} direct worker${workerNodes.length === 1 ? '' : 's'}`,
      techStack: bossNodes[0]?.techStack || [],
    };
  }

  const managers = managerNodes.map((manager, index) => {
    const managerWorkers = workerNodes.filter((worker) =>
      customEdges.some((edge) => edge.from === manager.id && edge.to === worker.id),
    );

    return {
      role: manager.role || 'Manager',
      description: `${manager.role || 'Manager'} manager for task execution`,
      priority: index + 1,
      workers: managerWorkers.map((worker) => ({
        role: worker.role || 'Specialist',
        description: `${worker.role || 'Specialist'} worker`,
      })),
    };
  });

  return {
    useAiPlanning: false,
    managers,
    architecture: bossNodes.length > 0
      ? `Custom workflow with ${bossNodes[0]?.role || 'Boss'} and ${managerNodes.length} managers`
      : `Custom workflow with ${managerNodes.length} managers`,
    techStack: bossNodes[0]?.techStack || [],
  };
}

export function buildTaskProviderAuth({
  provider,
  apiKey,
  defaultToken,
  customProvider,
}: BuildTaskProviderAuthInput): TaskProviderAuth {
  if (customProvider) {
    return {
      provider: 'ollama',
      tokens: {
        ollama: firstConfiguredToken(apiKey, customProvider.api_key, defaultToken),
        base_url: customProvider.base_url,
      },
    };
  }

  return {
    provider,
    tokens: {
      [provider]: firstConfiguredToken(apiKey, defaultToken),
    },
  };
}
