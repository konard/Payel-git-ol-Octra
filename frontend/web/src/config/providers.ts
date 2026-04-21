// Конфигурация всех доступных LLM-провайдеров
import geminiIcon from '../images/gemini-color.png';
import claudeIcon from '../images/Claude_AI_symbol.svg';
import deepseekIcon from '../images/deepseek-color.png';
import grokIcon from '../images/grok.png';
import qwenIcon from '../images/qwen-color.png';
import openaiIcon from '../images/icon.png';
import openrouterIcon from '../images/openrouter.svg';
import zaiIcon from '../images/zai.png';

// Кеш моделей OpenRouter
let cachedModels: ProviderModel[] | null = null;
let cacheTimestamp: number | null = null;
const CACHE_DURATION = 24 * 60 * 60 * 1000; // 24 часа

// Функция для загрузки моделей из OpenRouter API
export async function fetchOpenRouterModels(): Promise<ProviderModel[]> {
  // Проверяем кеш в localStorage
  const cached = localStorage.getItem('openrouter-models');
  const cachedTime = localStorage.getItem('openrouter-models-timestamp');

  if (cached && cachedTime) {
    const timeDiff = Date.now() - parseInt(cachedTime);
    if (timeDiff < CACHE_DURATION) {
      try {
        return JSON.parse(cached);
      } catch (e) {
        // Игнорируем ошибку парсинга
      }
    }
  }

  try {
    const API_KEY = import.meta.env.VITE_OPENROUTER_API_KEY;
    if (!API_KEY) {
      console.warn('VITE_OPENROUTER_API_KEY not set, using fallback models');
      return getFallbackModels();
    }

    const response = await fetch('https://openrouter.ai/api/v1/models', {
      headers: {
        'Authorization': `Bearer ${API_KEY}`,
        'Content-Type': 'application/json'
      }
    });

    if (!response.ok) {
      throw new Error(`HTTP ${response.status}`);
    }

    const data = await response.json();

    // Фильтруем только нужные модели
    const allowedPrefixes = ['openai/', 'google/', 'qwen/', 'z-ai/', 'anthropic/', 'deepseek/'];

    const iconMap = {
      'openai': openaiIcon,
      'google': geminiIcon,
      'qwen': qwenIcon,
      'z-ai': zaiIcon,
      'anthropic': claudeIcon,
      'deepseek': deepseekIcon
    };

    const filteredModels = data.data
      .filter((model: any) => allowedPrefixes.some((prefix: string) => model.id.startsWith(prefix)))
      .map((model: any) => {
        const provider = model.id.split('/')[0];
        const isFree = model.id.includes(':free') || model.pricing?.prompt === '0';

        return {
          id: model.id,
          name: model.name,
          icon: iconMap[provider as keyof typeof iconMap] || openaiIcon,
          free: isFree,
          recommended: false,
          providerId: 'openrouter'
        };
      })
      .sort((a, b) => {
        // Порядок провайдеров
        const providerOrder = {
          'openai': 1,
          'anthropic': 2,
          'google': 3,
          'qwen': 4,
          'deepseek': 5,
          'z-ai': 6
        };

        const getProviderFromId = (modelId: string) => modelId.split('/')[0];
        const aProvider = getProviderFromId(a.id);
        const bProvider = getProviderFromId(b.id);

        const aOrder = providerOrder[aProvider as keyof typeof providerOrder] || 99;
        const bOrder = providerOrder[bProvider as keyof typeof providerOrder] || 99;

        if (aOrder !== bOrder) {
          return aOrder - bOrder;
        }

        // Внутри провайдера сортируем по имени модели
        return a.name.localeCompare(b.name);
      });

    // Кешируем результат
    localStorage.setItem('openrouter-models', JSON.stringify(filteredModels));
    localStorage.setItem('openrouter-models-timestamp', Date.now().toString());

    return filteredModels;
  } catch (error) {
    console.error('Failed to fetch OpenRouter models:', error);
    return getFallbackModels();
  }
}

// Fallback модели на случай, если API недоступен
function getFallbackModels(): ProviderModel[] {
  return [
    { id: 'openai/gpt-4o-mini', name: 'OpenAI: GPT-4o Mini', icon: openaiIcon, free: false, recommended: false, providerId: 'openrouter' },
    { id: 'anthropic/claude-sonnet-4', name: 'Anthropic: Claude Sonnet 4', icon: claudeIcon, free: false, recommended: false, providerId: 'openrouter' },
    { id: 'google/gemini-2.5-flash', name: 'Google: Gemini 2.5 Flash', icon: geminiIcon, free: false, recommended: false, providerId: 'openrouter' },
    { id: 'qwen/qwen3-coder', name: 'Qwen: Qwen3 Coder', icon: qwenIcon, free: false, recommended: true, providerId: 'openrouter' },
    { id: 'deepseek/deepseek-chat', name: 'DeepSeek: DeepSeek Chat', icon: deepseekIcon, free: false, recommended: false, providerId: 'openrouter' },
    { id: 'z-ai/glm-4.5-air', name: 'Z.ai: GLM 4.5 Air', icon: zaiIcon, free: false, recommended: false, providerId: 'openrouter' },
  ];
}

// Устаревший экспорт для обратной совместимости (вернет fallback модели)
export const ALL_MODELS: ProviderModel[] = getFallbackModels();

// Создаем PROVIDERS асинхронно
export async function getProviders(): Promise<ProviderConfig[]> {
  const openRouterModels = await fetchOpenRouterModels();

  return [
    {
      id: 'openrouter',
      name: 'OpenRouter',
      color: '#f97316',
      bgColor: 'rgba(249, 115, 22, 0.15)',
      icon: openrouterIcon,
      description: '',
      defaultModel: 'qwen/qwen3-coder',
      pricing: 'Free + Paid',
      models: openRouterModels,
    },
    {
      id: 'gemini',
      name: 'Google Gemini',
      color: '#4285f4',
      bgColor: 'rgba(66, 133, 244, 0.15)',
      icon: geminiIcon,
      description: '',
      defaultModel: 'gemini-2.5-flash',
      pricing: 'Free (20/min)',
      models: [
        { id: 'gemini-2.5-flash', name: 'Gemini 2.5 Flash', icon: geminiIcon, free: true, recommended: true, providerId: 'gemini' },
        { id: 'gemini-2.5-pro', name: 'Gemini 2.5 Pro', icon: geminiIcon, free: true, recommended: false, providerId: 'gemini' },
        { id: 'gemini-2.0-flash', name: 'Gemini 2.0 Flash', icon: geminiIcon, free: true, recommended: false, providerId: 'gemini' },
        { id: 'gemini-2.0-flash-lite', name: 'Gemini 2.0 Flash Lite', icon: geminiIcon, free: true, recommended: false, providerId: 'gemini' },
        { id: 'gemini-2.0-flash-thinking-exp', name: 'Gemini 2.0 Thinking', icon: geminiIcon, free: true, recommended: false, providerId: 'gemini' },
        { id: 'gemini-1.5-flash', name: 'Gemini 1.5 Flash', icon: geminiIcon, free: true, recommended: false, providerId: 'gemini' },
        { id: 'gemini-1.5-pro', name: 'Gemini 1.5 Pro', icon: geminiIcon, free: true, recommended: false, providerId: 'gemini' },
      ],
    },
    {
      id: 'openai',
      name: 'OpenAI',
      color: '#10a37f',
      bgColor: 'rgba(16, 163, 127, 0.15)',
      icon: openaiIcon,
      description: '',
      defaultModel: 'gpt-4o-mini',
      pricing: 'Paid',
      models: [
        { id: 'gpt-4o-mini', name: 'GPT-4o Mini', icon: openaiIcon, free: false, recommended: true, providerId: 'openai' },
        { id: 'gpt-4o', name: 'GPT-4o', icon: openaiIcon, free: false, recommended: false, providerId: 'openai' },
        { id: 'gpt-4-turbo', name: 'GPT-4 Turbo', icon: openaiIcon, free: false, recommended: false, providerId: 'openai' },
        { id: 'o1-mini', name: 'o1 Mini', icon: openaiIcon, free: false, recommended: false, providerId: 'openai' },
        { id: 'o1', name: 'o1', icon: openaiIcon, free: false, recommended: false, providerId: 'openai' },
        { id: 'o3-mini', name: 'o3 Mini', icon: openaiIcon, free: false, recommended: false, providerId: 'openai' },
        { id: 'o4-mini', name: 'o4 Mini', icon: openaiIcon, free: false, recommended: false, providerId: 'openai' },
      ],
    },
    {
      id: 'claude',
      name: 'Anthropic Claude',
      color: '#7c3aed',
      bgColor: 'rgba(124, 58, 237, 0.15)',
      icon: claudeIcon,
      description: '',
      defaultModel: 'anthropic/claude-sonnet-4-20250514',
      pricing: 'Paid',
      models: [
        { id: 'anthropic/claude-opus-4-6', name: 'Claude Opus 4.6', icon: claudeIcon, free: false, recommended: false, providerId: 'claude' },
        { id: 'anthropic/claude-opus-4-5', name: 'Claude Opus 4.5', icon: claudeIcon, free: false, recommended: false, providerId: 'claude' },
        { id: 'anthropic/claude-sonnet-4-6', name: 'Claude Sonnet 4.6', icon: claudeIcon, free: false, recommended: false, providerId: 'claude' },
        { id: 'anthropic/claude-sonnet-4-5', name: 'Claude Sonnet 4.5', icon: claudeIcon, free: false, recommended: false, providerId: 'claude' },
        { id: 'anthropic/claude-sonnet-4-20250514', name: 'Claude Sonnet 4', icon: claudeIcon, free: false, recommended: false, providerId: 'claude' },
        { id: 'anthropic/claude-haiku-4-5', name: 'Claude Haiku 4.5', icon: claudeIcon, free: false, recommended: false, providerId: 'claude' },
      ],
    },
    {
      id: 'deepseek',
      name: 'DeepSeek',
      color: '#3b82f6',
      bgColor: 'rgba(59, 130, 246, 0.15)',
      icon: deepseekIcon,
      description: '',
      defaultModel: 'deepseek-chat',
      pricing: 'Paid (Cheap)',
      models: [
        { id: 'deepseek-chat', name: 'DeepSeek Chat V3', icon: deepseekIcon, free: false, recommended: true, providerId: 'deepseek' },
        { id: 'deepseek-reasoner', name: 'DeepSeek R1', icon: deepseekIcon, free: false, recommended: false, providerId: 'deepseek' },
        { id: 'deepseek-v2.5', name: 'DeepSeek V2.5', icon: deepseekIcon, free: false, recommended: false, providerId: 'deepseek' },
        { id: 'deepseek-coder', name: 'DeepSeek Coder', icon: deepseekIcon, free: false, recommended: false, providerId: 'deepseek' },
      ],
    },
    {
      id: 'grok',
      name: 'xAI Grok',
      color: '#f59e0b',
      bgColor: 'rgba(245, 158, 11, 0.15)',
      icon: grokIcon,
      description: '',
      defaultModel: 'grok-3',
      pricing: 'Paid',
      models: [
        { id: 'grok-3', name: 'Grok 3', icon: grokIcon, free: false, recommended: true, providerId: 'grok' },
        { id: 'grok-2', name: 'Grok 2', icon: grokIcon, free: false, recommended: false, providerId: 'grok' },
        { id: 'grok-beta', name: 'Grok Beta', icon: grokIcon, free: false, recommended: false, providerId: 'grok' },
      ],
    },
    {
      id: 'qwen',
      name: 'Qwen',
      color: '#7c3aed',
      bgColor: 'rgba(124, 58, 237, 0.15)',
      icon: qwenIcon,
      description: '',
      defaultModel: 'qwen/qwen3.6-plus:free',
      pricing: 'Paid',
      models: [
        { id: 'qwen/qwen3.6-plus:free', name: 'Qwen 3.6 Plus (Free)', icon: qwenIcon, free: true, recommended: true, providerId: 'qwen' },
        { id: 'qwen/qwen3-coder', name: 'Qwen3 Coder', icon: qwenIcon, free: false, recommended: false, providerId: 'qwen' },
        { id: 'qwen/qwen3-max', name: 'Qwen3 Max', icon: qwenIcon, free: false, recommended: false, providerId: 'qwen' },
        { id: 'qwen/qwen3-vl-32b-thinking', name: 'Qwen3 VL 32B Thinking', icon: qwenIcon, free: false, recommended: false, providerId: 'qwen' },
        { id: 'qwen/qwen3-next-80b-a3b-instruct', name: 'Qwen3 Next 80B A3B Instruct', icon: qwenIcon, free: false, recommended: false, providerId: 'qwen' },
        { id: 'qwen/qwen3-235b-a22b-thinking', name: 'Qwen3 235B A22B Thinking', icon: qwenIcon, free: false, recommended: false, providerId: 'qwen' },
        { id: 'qwen/qwen3-235b-a22b-instruct', name: 'Qwen3 235B A22B Instruct', icon: qwenIcon, free: false, recommended: false, providerId: 'qwen' },
        { id: 'qwen/qwen3-32b', name: 'Qwen3 32B', icon: qwenIcon, free: false, recommended: false, providerId: 'qwen' },
        { id: 'qwen/qwen3-30b-a3b', name: 'Qwen3 30B A3B', icon: qwenIcon, free: false, recommended: false, providerId: 'qwen' },
        { id: 'qwen/qwen3-coder-480b-a35b-instruct', name: 'Qwen3 Coder 480B A35B Instruct', icon: qwenIcon, free: false, recommended: false, providerId: 'qwen' },
        { id: 'qwen/qwen2.5-omni-7b', name: 'Qwen2.5 Omni 7B', icon: qwenIcon, free: false, recommended: false, providerId: 'qwen' },
        { id: 'qwen/qwen2.5-vl-72b-instruct', name: 'Qwen2.5 VL 72B', icon: qwenIcon, free: false, recommended: false, providerId: 'qwen' },
        { id: 'qwen/qwen2.5-vl-32b-instruct', name: 'Qwen2.5 VL 32B', icon: qwenIcon, free: false, recommended: false, providerId: 'qwen' },
        { id: 'qwen/qwen2.5-vl-7b-instruct', name: 'Qwen2.5 VL 7B', icon: qwenIcon, free: false, recommended: false, providerId: 'qwen' },
        { id: 'qwen/qwen2-vl-72b-instruct', name: 'Qwen2 VL 72B Instruct', icon: qwenIcon, free: false, recommended: false, providerId: 'qwen' },
        { id: 'qwen/qwq-32b', name: 'QwQ 32B', icon: qwenIcon, free: false, recommended: false, providerId: 'qwen' },
        { id: 'qwen/qvq-72b-preview', name: 'QvQ 72B Preview', icon: qwenIcon, free: false, recommended: false, providerId: 'qwen' },
      ],
    },
    {
      id: 'zai',
      name: 'Z.AI',
      color: '#06b6d4',
      bgColor: 'rgba(6, 182, 212, 0.15)',
      icon: zaiIcon,
      description: '',
      defaultModel: 'glm-4.5-air',
      pricing: 'Paid',
      models: [
        { id: 'glm-4.5-air', name: 'GLM 4.5 Air', icon: zaiIcon, free: false, recommended: true, providerId: 'zai' },
        { id: 'glm-4.5', name: 'GLM 4.5', icon: zaiIcon, free: false, recommended: false, providerId: 'zai' },
        { id: 'glm-4', name: 'GLM 4', icon: zaiIcon, free: false, recommended: false, providerId: 'zai' },
      ],
    },
  ];
}

// Синхронная версия для обратной совместимости (без OpenRouter моделей)
export const PROVIDERS: ProviderConfig[] = [
  {
    id: 'openrouter',
    name: 'OpenRouter',
    color: '#f97316',
    bgColor: 'rgba(249, 115, 22, 0.15)',
    icon: openrouterIcon,
    description: '',
    defaultModel: 'qwen/qwen3-coder',
    pricing: 'Free + Paid',
    models: [], // Будет заполнено асинхронно
  },
  {
    id: 'gemini',
    name: 'Google Gemini',
    color: '#4285f4',
    bgColor: 'rgba(66, 133, 244, 0.15)',
    icon: geminiIcon,
    description: '',
    defaultModel: 'gemini-2.5-flash',
    pricing: 'Free (20/min)',
    models: [
      { id: 'gemini-2.5-flash', name: 'Gemini 2.5 Flash', icon: geminiIcon, free: true, recommended: true, providerId: 'gemini' },
      { id: 'gemini-2.5-pro', name: 'Gemini 2.5 Pro', icon: geminiIcon, free: true, recommended: false, providerId: 'gemini' },
      { id: 'gemini-2.0-flash', name: 'Gemini 2.0 Flash', icon: geminiIcon, free: true, recommended: false, providerId: 'gemini' },
      { id: 'gemini-2.0-flash-lite', name: 'Gemini 2.0 Flash Lite', icon: geminiIcon, free: true, recommended: false, providerId: 'gemini' },
      { id: 'gemini-2.0-flash-thinking-exp', name: 'Gemini 2.0 Thinking', icon: geminiIcon, free: true, recommended: false, providerId: 'gemini' },
      { id: 'gemini-1.5-flash', name: 'Gemini 1.5 Flash', icon: geminiIcon, free: true, recommended: false, providerId: 'gemini' },
      { id: 'gemini-1.5-pro', name: 'Gemini 1.5 Pro', icon: geminiIcon, free: true, recommended: false, providerId: 'gemini' },
    ],
  },
  {
    id: 'openai',
    name: 'OpenAI',
    color: '#10a37f',
    bgColor: 'rgba(16, 163, 127, 0.15)',
    icon: openaiIcon,
    description: '',
    defaultModel: 'gpt-4o-mini',
    pricing: 'Paid',
    models: [
      { id: 'gpt-4o-mini', name: 'GPT-4o Mini', icon: openaiIcon, free: false, recommended: true, providerId: 'openai' },
      { id: 'gpt-4o', name: 'GPT-4o', icon: openaiIcon, free: false, recommended: false, providerId: 'openai' },
      { id: 'gpt-4-turbo', name: 'GPT-4 Turbo', icon: openaiIcon, free: false, recommended: false, providerId: 'openai' },
      { id: 'o1-mini', name: 'o1 Mini', icon: openaiIcon, free: false, recommended: false, providerId: 'openai' },
      { id: 'o1', name: 'o1', icon: openaiIcon, free: false, recommended: false, providerId: 'openai' },
      { id: 'o3-mini', name: 'o3 Mini', icon: openaiIcon, free: false, recommended: false, providerId: 'openai' },
      { id: 'o4-mini', name: 'o4 Mini', icon: openaiIcon, free: false, recommended: false, providerId: 'openai' },
    ],
  },
  {
    id: 'claude',
    name: 'Anthropic Claude',
    color: '#7c3aed',
    bgColor: 'rgba(124, 58, 237, 0.15)',
    icon: claudeIcon,
    description: '',
    defaultModel: 'anthropic/claude-sonnet-4-20250514',
    pricing: 'Paid',
    models: [
      { id: 'anthropic/claude-opus-4-6', name: 'Claude Opus 4.6', icon: claudeIcon, free: false, recommended: false, providerId: 'claude' },
      { id: 'anthropic/claude-opus-4-5', name: 'Claude Opus 4.5', icon: claudeIcon, free: false, recommended: false, providerId: 'claude' },
      { id: 'anthropic/claude-sonnet-4-6', name: 'Claude Sonnet 4.6', icon: claudeIcon, free: false, recommended: false, providerId: 'claude' },
      { id: 'anthropic/claude-sonnet-4-5', name: 'Claude Sonnet 4.5', icon: claudeIcon, free: false, recommended: false, providerId: 'claude' },
      { id: 'anthropic/claude-sonnet-4-20250514', name: 'Claude Sonnet 4', icon: claudeIcon, free: false, recommended: false, providerId: 'claude' },
      { id: 'anthropic/claude-haiku-4-5', name: 'Claude Haiku 4.5', icon: claudeIcon, free: false, recommended: false, providerId: 'claude' },
    ],
  },
  {
    id: 'deepseek',
    name: 'DeepSeek',
    color: '#3b82f6',
    bgColor: 'rgba(59, 130, 246, 0.15)',
    icon: deepseekIcon,
    description: '',
    defaultModel: 'deepseek-chat',
    pricing: 'Paid (Cheap)',
    models: [
      { id: 'deepseek-chat', name: 'DeepSeek Chat V3', icon: deepseekIcon, free: false, recommended: true, providerId: 'deepseek' },
      { id: 'deepseek-reasoner', name: 'DeepSeek R1', icon: deepseekIcon, free: false, recommended: false, providerId: 'deepseek' },
      { id: 'deepseek-v2.5', name: 'DeepSeek V2.5', icon: deepseekIcon, free: false, recommended: false, providerId: 'deepseek' },
      { id: 'deepseek-coder', name: 'DeepSeek Coder', icon: deepseekIcon, free: false, recommended: false, providerId: 'deepseek' },
    ],
  },
  {
    id: 'grok',
    name: 'xAI Grok',
    color: '#f59e0b',
    bgColor: 'rgba(245, 158, 11, 0.15)',
    icon: grokIcon,
    description: '',
    defaultModel: 'grok-3',
    pricing: 'Paid',
    models: [
      { id: 'grok-3', name: 'Grok 3', icon: grokIcon, free: false, recommended: true, providerId: 'grok' },
      { id: 'grok-2', name: 'Grok 2', icon: grokIcon, free: false, recommended: false, providerId: 'grok' },
      { id: 'grok-beta', name: 'Grok Beta', icon: grokIcon, free: false, recommended: false, providerId: 'grok' },
    ],
  },
  {
    id: 'qwen',
    name: 'Qwen',
    color: '#7c3aed',
    bgColor: 'rgba(124, 58, 237, 0.15)',
    icon: qwenIcon,
    description: '',
    defaultModel: 'qwen/qwen3.6-plus:free',
    pricing: 'Paid',
    models: [
      { id: 'qwen/qwen3.6-plus:free', name: 'Qwen 3.6 Plus (Free)', icon: qwenIcon, free: true, recommended: true, providerId: 'qwen' },
      { id: 'qwen/qwen3-coder', name: 'Qwen3 Coder', icon: qwenIcon, free: false, recommended: false, providerId: 'qwen' },
      { id: 'qwen/qwen3-max', name: 'Qwen3 Max', icon: qwenIcon, free: false, recommended: false, providerId: 'qwen' },
      { id: 'qwen/qwen3-vl-32b-thinking', name: 'Qwen3 VL 32B Thinking', icon: qwenIcon, free: false, recommended: false, providerId: 'qwen' },
      { id: 'qwen/qwen3-next-80b-a3b-instruct', name: 'Qwen3 Next 80B A3B Instruct', icon: qwenIcon, free: false, recommended: false, providerId: 'qwen' },
      { id: 'qwen/qwen3-235b-a22b-thinking', name: 'Qwen3 235B A22B Thinking', icon: qwenIcon, free: false, recommended: false, providerId: 'qwen' },
      { id: 'qwen/qwen3-235b-a22b-instruct', name: 'Qwen3 235B A22B Instruct', icon: qwenIcon, free: false, recommended: false, providerId: 'qwen' },
      { id: 'qwen/qwen3-32b', name: 'Qwen3 32B', icon: qwenIcon, free: false, recommended: false, providerId: 'qwen' },
      { id: 'qwen/qwen3-30b-a3b', name: 'Qwen3 30B A3B', icon: qwenIcon, free: false, recommended: false, providerId: 'qwen' },
      { id: 'qwen/qwen3-coder-480b-a35b-instruct', name: 'Qwen3 Coder 480B A35B Instruct', icon: qwenIcon, free: false, recommended: false, providerId: 'qwen' },
      { id: 'qwen/qwen2.5-omni-7b', name: 'Qwen2.5 Omni 7B', icon: qwenIcon, free: false, recommended: false, providerId: 'qwen' },
      { id: 'qwen/qwen2.5-vl-72b-instruct', name: 'Qwen2.5 VL 72B', icon: qwenIcon, free: false, recommended: false, providerId: 'qwen' },
      { id: 'qwen/qwen2.5-vl-32b-instruct', name: 'Qwen2.5 VL 32B', icon: qwenIcon, free: false, recommended: false, providerId: 'qwen' },
      { id: 'qwen/qwen2.5-vl-7b-instruct', name: 'Qwen2.5 VL 7B', icon: qwenIcon, free: false, recommended: false, providerId: 'qwen' },
      { id: 'qwen/qwen2-vl-72b-instruct', name: 'Qwen2 VL 72B Instruct', icon: qwenIcon, free: false, recommended: false, providerId: 'qwen' },
      { id: 'qwen/qwq-32b', name: 'QwQ 32B', icon: qwenIcon, free: false, recommended: false, providerId: 'qwen' },
      { id: 'qwen/qvq-72b-preview', name: 'QvQ 72B Preview', icon: qwenIcon, free: false, recommended: false, providerId: 'qwen' },
    ],
  },
  {
    id: 'zai',
    name: 'Z.AI',
    color: '#06b6d4',
    bgColor: 'rgba(6, 182, 212, 0.15)',
    icon: zaiIcon,
    description: '',
    defaultModel: 'glm-4.5-air',
    pricing: 'Paid',
    models: [
      { id: 'glm-4.5-air', name: 'GLM 4.5 Air', icon: zaiIcon, free: false, recommended: true, providerId: 'zai' },
      { id: 'glm-4.5', name: 'GLM 4.5', icon: zaiIcon, free: false, recommended: false, providerId: 'zai' },
      { id: 'glm-4', name: 'GLM 4', icon: zaiIcon, free: false, recommended: false, providerId: 'zai' },
    ],
  },
];

export interface ProviderModel {
  id: string;
  name: string;
  icon: string;
  free: boolean;
  recommended: boolean;
  providerId: string;
}

export interface ProviderConfig {
  id: string;
  name: string;
  color: string;
  bgColor: string;
  icon: string;
  description: string;
  defaultModel: string;
  pricing: string;
  models: ProviderModel[];
}

export function getProviderById(id: string): ProviderConfig | undefined {
  return PROVIDERS.find((p) => p.id === id);
}

export function getModelForProvider(providerId: string, modelId: string): ProviderModel | undefined {
  const provider = getProviderById(providerId);
  return provider?.models.find((m) => m.id === modelId);
}
