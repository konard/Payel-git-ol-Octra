// Конфигурация всех доступных LLM-провайдеров
import geminiIcon from '../images/gemini-color.png';
import claudeIcon from '../images/Claude_AI_symbol.svg';
import deepseekIcon from '../images/deepseek-color.png';
import grokIcon from '../images/grok.png';
import qwenIcon from '../images/qwen-color.png';
import openaiIcon from '../images/icon.png';
import openrouterIcon from '../images/openrouter.svg';

// Все модели (для OpenRouter — доступны все)
export const ALL_MODELS: ProviderModel[] = [
  // OpenAI models via OpenRouter
  { id: 'openai/gpt-5.4', name: 'GPT-5.4', icon: openaiIcon, free: false, recommended: false, providerId: 'openrouter' },
  { id: 'openai/gpt-5.4-pro', name: 'GPT-5.4 Pro', icon: openaiIcon, free: false, recommended: false, providerId: 'openrouter' },
  { id: 'openai/gpt-5.4-mini', name: 'GPT-5.4 Mini', icon: openaiIcon, free: false, recommended: false, providerId: 'openrouter' },
  { id: 'openai/gpt-5.4-nano', name: 'GPT-5.4 Nano', icon: openaiIcon, free: false, recommended: false, providerId: 'openrouter' },
  { id: 'openai/gpt-5.4-codex', name: 'GPT-5.4 Codex', icon: openaiIcon, free: false, recommended: false, providerId: 'openrouter' },
  { id: 'openai/gpt-5.3-codex', name: 'GPT-5.3 Codex', icon: openaiIcon, free: false, recommended: false, providerId: 'openrouter' },
  { id: 'openai/gpt-5.2', name: 'GPT-5.2', icon: openaiIcon, free: false, recommended: false, providerId: 'openrouter' },
  { id: 'openai/gpt-5.2-codex', name: 'GPT-5.2 Codex', icon: openaiIcon, free: false, recommended: false, providerId: 'openrouter' },
  { id: 'openai/gpt-5.1', name: 'GPT-5.1 Medium', icon: openaiIcon, free: false, recommended: false, providerId: 'openrouter' },
  { id: 'openai/gpt-5.1-high', name: 'GPT-5.1 High', icon: openaiIcon, free: false, recommended: false, providerId: 'openrouter' },
  { id: 'openai/gpt-5.1-codex', name: 'GPT-5.1 Codex', icon: openaiIcon, free: false, recommended: false, providerId: 'openrouter' },
  { id: 'openai/gpt-5.1-instant', name: 'GPT-5.1 Instant', icon: openaiIcon, free: false, recommended: false, providerId: 'openrouter' },
  { id: 'openai/gpt-5.1-thinking', name: 'GPT-5.1 Thinking', icon: openaiIcon, free: false, recommended: false, providerId: 'openrouter' },
  { id: 'openai/gpt-5', name: 'GPT-5', icon: openaiIcon, free: false, recommended: false, providerId: 'openrouter' },
  { id: 'openai/gpt-5-mini', name: 'GPT-5 Mini', icon: openaiIcon, free: false, recommended: false, providerId: 'openrouter' },
  { id: 'openai/gpt-5-nano', name: 'GPT-5 Nano', icon: openaiIcon, free: false, recommended: false, providerId: 'openrouter' },
  { id: 'openai/gpt-5-high', name: 'GPT-5 High', icon: openaiIcon, free: false, recommended: false, providerId: 'openrouter' },
  { id: 'openai/gpt-5-medium', name: 'GPT-5 Medium', icon: openaiIcon, free: false, recommended: false, providerId: 'openrouter' },
  { id: 'openai/gpt-5-codex', name: 'GPT-5 Codex', icon: openaiIcon, free: false, recommended: false, providerId: 'openrouter' },
  { id: 'openai/gpt-4.1', name: 'GPT-4.1', icon: openaiIcon, free: false, recommended: false, providerId: 'openrouter' },
  { id: 'openai/gpt-4.1-mini', name: 'GPT-4.1 Mini', icon: openaiIcon, free: false, recommended: false, providerId: 'openrouter' },
  { id: 'openai/gpt-4.1-nano', name: 'GPT-4.1 Nano', icon: openaiIcon, free: false, recommended: false, providerId: 'openrouter' },
  { id: 'openai/gpt-4.5-preview', name: 'GPT-4.5', icon: openaiIcon, free: false, recommended: false, providerId: 'openrouter' },
  { id: 'openai/gpt-4o', name: 'GPT-4o', icon: openaiIcon, free: false, recommended: false, providerId: 'openrouter' },
  { id: 'openai/gpt-4o-mini', name: 'GPT-4o Mini', icon: openaiIcon, free: false, recommended: false, providerId: 'openrouter' },
  { id: 'openai/gpt-4-turbo', name: 'GPT-4 Turbo', icon: openaiIcon, free: false, recommended: false, providerId: 'openrouter' },
  { id: 'openai/gpt-4', name: 'GPT-4', icon: openaiIcon, free: false, recommended: false, providerId: 'openrouter' },
  { id: 'openai/gpt-3.5-turbo', name: 'GPT-3.5 Turbo', icon: openaiIcon, free: false, recommended: false, providerId: 'openrouter' },
  // Claude models via OpenRouter
  { id: 'anthropic/claude-opus-4-6', name: 'Claude Opus 4.6', icon: claudeIcon, free: false, recommended: false, providerId: 'openrouter' },
  { id: 'anthropic/claude-opus-4-5', name: 'Claude Opus 4.5', icon: claudeIcon, free: false, recommended: false, providerId: 'openrouter' },
  { id: 'anthropic/claude-sonnet-4-6', name: 'Claude Sonnet 4.6', icon: claudeIcon, free: false, recommended: false, providerId: 'openrouter' },
  { id: 'anthropic/claude-sonnet-4-5', name: 'Claude Sonnet 4.5', icon: claudeIcon, free: false, recommended: false, providerId: 'openrouter' },
  { id: 'anthropic/claude-sonnet-4-20250514', name: 'Claude Sonnet 4', icon: claudeIcon, free: false, recommended: false, providerId: 'openrouter' },
  { id: 'anthropic/claude-haiku-4-5', name: 'Claude Haiku 4.5', icon: claudeIcon, free: false, recommended: false, providerId: 'openrouter' },
  // Gemini models via OpenRouter
  { id: 'google/gemini-3.1-pro', name: 'Gemini 3.1 Pro', icon: geminiIcon, free: false, recommended: false, providerId: 'openrouter' },
  { id: 'google/gemini-3-pro', name: 'Gemini 3 Pro', icon: geminiIcon, free: false, recommended: false, providerId: 'openrouter' },
  { id: 'google/gemini-3-flash', name: 'Gemini 3 Flash', icon: geminiIcon, free: false, recommended: false, providerId: 'openrouter' },
  { id: 'google/gemini-2.5-pro', name: 'Gemini 2.5 Pro', icon: geminiIcon, free: false, recommended: false, providerId: 'openrouter' },
  { id: 'google/gemini-2.5-flash', name: 'Gemini 2.5 Flash', icon: geminiIcon, free: false, recommended: false, providerId: 'openrouter' },
  { id: 'google/gemini-2.5-flash-lite', name: 'Gemini 2.5 Flash-Lite', icon: geminiIcon, free: false, recommended: false, providerId: 'openrouter' },
  { id: 'google/gemini-2.0-flash', name: 'Gemini 2.0 Flash', icon: geminiIcon, free: false, recommended: false, providerId: 'openrouter' },
  { id: 'google/gemini-2.0-flash-lite', name: 'Gemini 2.0 Flash-Lite', icon: geminiIcon, free: false, recommended: false, providerId: 'openrouter' },
  { id: 'google/gemini-1.5-pro', name: 'Gemini 1.5 Pro', icon: geminiIcon, free: false, recommended: false, providerId: 'openrouter' },
  { id: 'google/gemini-1.5-flash', name: 'Gemini 1.5 Flash', icon: geminiIcon, free: false, recommended: false, providerId: 'openrouter' },
  { id: 'google/gemini-1.0-pro', name: 'Gemini 1.0 Pro', icon: geminiIcon, free: false, recommended: false, providerId: 'openrouter' },
  // DeepSeek models via OpenRouter
  { id: 'deepseek/deepseek-v3.2', name: 'DeepSeek V3.2 Thinking', icon: deepseekIcon, free: false, recommended: false, providerId: 'openrouter' },
  { id: 'deepseek/deepseek-v3.2-nonthinking', name: 'DeepSeek V3.2', icon: deepseekIcon, free: false, recommended: false, providerId: 'openrouter' },
  { id: 'deepseek/deepseek-v3.2-speciale', name: 'DeepSeek V3.2 Speciale', icon: deepseekIcon, free: false, recommended: false, providerId: 'openrouter' },
  { id: 'deepseek/deepseek-v3.2-exp', name: 'DeepSeek V3.2 Exp', icon: deepseekIcon, free: false, recommended: false, providerId: 'openrouter' },
  { id: 'deepseek/deepseek-v3.1', name: 'DeepSeek V3.1', icon: deepseekIcon, free: false, recommended: false, providerId: 'openrouter' },
  { id: 'deepseek/deepseek-v3', name: 'DeepSeek V3', icon: deepseekIcon, free: false, recommended: false, providerId: 'openrouter' },
  { id: 'deepseek/deepseek-r1', name: 'DeepSeek R1', icon: deepseekIcon, free: false, recommended: false, providerId: 'openrouter' },
  { id: 'deepseek/deepseek-r1-0528', name: 'DeepSeek R1 0528', icon: deepseekIcon, free: false, recommended: false, providerId: 'openrouter' },
  { id: 'deepseek/deepseek-r1-zero', name: 'DeepSeek R1 Zero', icon: deepseekIcon, free: false, recommended: false, providerId: 'openrouter' },
  { id: 'deepseek/deepseek-r1-distill-llama-70b', name: 'DeepSeek R1 Distill Llama 70B', icon: deepseekIcon, free: false, recommended: false, providerId: 'openrouter' },
  { id: 'deepseek/deepseek-r1-distill-qwen-32b', name: 'DeepSeek R1 Distill Qwen 32B', icon: deepseekIcon, free: false, recommended: false, providerId: 'openrouter' },
  { id: 'deepseek/deepseek-r1-distill-qwen-14b', name: 'DeepSeek R1 Distill Qwen 14B', icon: deepseekIcon, free: false, recommended: false, providerId: 'openrouter' },
  { id: 'deepseek/deepseek-r1-distill-qwen-7b', name: 'DeepSeek R1 Distill Qwen 7B', icon: deepseekIcon, free: false, recommended: false, providerId: 'openrouter' },
  { id: 'deepseek/deepseek-r1-distill-qwen-1.5b', name: 'DeepSeek R1 Distill Qwen 1.5B', icon: deepseekIcon, free: false, recommended: false, providerId: 'openrouter' },
  { id: 'deepseek/deepseek-vl2', name: 'DeepSeek VL2', icon: deepseekIcon, free: false, recommended: false, providerId: 'openrouter' },
  { id: 'deepseek/deepseek-vl2-small', name: 'DeepSeek VL2 Small', icon: deepseekIcon, free: false, recommended: false, providerId: 'openrouter' },
  { id: 'deepseek/deepseek-vl2-tiny', name: 'DeepSeek VL2 Tiny', icon: deepseekIcon, free: false, recommended: false, providerId: 'openrouter' },
  // Grok models via OpenRouter
  { id: 'x-ai/grok-4.20', name: 'Grok 4.20', icon: grokIcon, free: false, recommended: false, providerId: 'openrouter' },
  { id: 'x-ai/grok-4.1', name: 'Grok 4.1 Fast Reasoning', icon: grokIcon, free: false, recommended: false, providerId: 'openrouter' },
  { id: 'x-ai/grok-4.1-nonreasoning', name: 'Grok 4.1 Fast Non-Reasoning', icon: grokIcon, free: false, recommended: false, providerId: 'openrouter' },
  { id: 'x-ai/grok-4-heavy', name: 'Grok 4 Heavy', icon: grokIcon, free: false, recommended: false, providerId: 'openrouter' },
  { id: 'x-ai/grok-4-fast', name: 'Grok 4 Fast Reasoning', icon: grokIcon, free: false, recommended: false, providerId: 'openrouter' },
  { id: 'x-ai/grok-4-fast-nonreasoning', name: 'Grok 4 Fast Non-Reasoning', icon: grokIcon, free: false, recommended: false, providerId: 'openrouter' },
  { id: 'x-ai/grok-code-fast-1', name: 'Grok Code Fast 1', icon: grokIcon, free: false, recommended: false, providerId: 'openrouter' },
  { id: 'x-ai/grok-3', name: 'Grok 3', icon: grokIcon, free: false, recommended: false, providerId: 'openrouter' },
  { id: 'x-ai/grok-3-mini', name: 'Grok 3 Mini', icon: grokIcon, free: false, recommended: false, providerId: 'openrouter' },
  { id: 'x-ai/grok-2', name: 'Grok 2', icon: grokIcon, free: false, recommended: false, providerId: 'openrouter' },
  { id: 'x-ai/grok-2-mini', name: 'Grok 2 Mini', icon: grokIcon, free: false, recommended: false, providerId: 'openrouter' },
  { id: 'x-ai/grok-1.5v', name: 'Grok 1.5V', icon: grokIcon, free: false, recommended: false, providerId: 'openrouter' },
  { id: 'x-ai/grok-1.5', name: 'Grok 1.5', icon: grokIcon, free: false, recommended: false, providerId: 'openrouter' },
  // Qwen models via OpenRouter
  { id: 'qwen/qwen3-max', name: 'Qwen3 Max', icon: qwenIcon, free: false, recommended: false, providerId: 'openrouter' },
  { id: 'qwen/qwen3-vl-32b-thinking', name: 'Qwen3 VL 32B Thinking', icon: qwenIcon, free: false, recommended: false, providerId: 'openrouter' },
  { id: 'qwen/qwen3-next-80b-a3b-instruct', name: 'Qwen3 Next 80B A3B Instruct', icon: qwenIcon, free: false, recommended: false, providerId: 'openrouter' },
  { id: 'qwen/qwen3-235b-a22b-thinking', name: 'Qwen3 235B A22B Thinking', icon: qwenIcon, free: false, recommended: false, providerId: 'openrouter' },
  { id: 'qwen/qwen3-235b-a22b-instruct', name: 'Qwen3 235B A22B Instruct', icon: qwenIcon, free: false, recommended: false, providerId: 'openrouter' },
  { id: 'qwen/qwen3-32b', name: 'Qwen3 32B', icon: qwenIcon, free: false, recommended: false, providerId: 'openrouter' },
  { id: 'qwen/qwen3-30b-a3b', name: 'Qwen3 30B A3B', icon: qwenIcon, free: false, recommended: false, providerId: 'openrouter' },
  { id: 'qwen/qwen3-coder-480b-a35b-instruct', name: 'Qwen3 Coder 480B A35B Instruct', icon: qwenIcon, free: false, recommended: false, providerId: 'openrouter' },
  { id: 'qwen/qwen3-coder', name: 'Qwen3 Coder', icon: qwenIcon, free: false, recommended: true, providerId: 'openrouter' },
  { id: 'qwen/qwen2.5-omni-7b', name: 'Qwen2.5 Omni 7B', icon: qwenIcon, free: false, recommended: false, providerId: 'openrouter' },
  { id: 'qwen/qwen2.5-vl-72b-instruct', name: 'Qwen2.5 VL 72B', icon: qwenIcon, free: false, recommended: false, providerId: 'openrouter' },
  { id: 'qwen/qwen2.5-vl-32b-instruct', name: 'Qwen2.5 VL 32B', icon: qwenIcon, free: false, recommended: false, providerId: 'openrouter' },
  { id: 'qwen/qwen2.5-vl-7b-instruct', name: 'Qwen2.5 VL 7B', icon: qwenIcon, free: false, recommended: false, providerId: 'openrouter' },
  { id: 'qwen/qwen2-vl-72b-instruct', name: 'Qwen2 VL 72B Instruct', icon: qwenIcon, free: false, recommended: false, providerId: 'openrouter' },
  { id: 'qwen/qwq-32b', name: 'QwQ 32B', icon: qwenIcon, free: false, recommended: false, providerId: 'openrouter' },
  { id: 'qwen/qvq-72b-preview', name: 'QvQ 72B Preview', icon: qwenIcon, free: false, recommended: false, providerId: 'openrouter' },
  // Free models
  { id: 'google/gemini-exp-1206:free', name: 'Gemini Exp 1206', icon: geminiIcon, free: true, recommended: false, providerId: 'openrouter' },
  { id: 'google/gemini-2.0-flash-exp:free', name: 'Gemini 2.0 Flash Exp', icon: geminiIcon, free: true, recommended: false, providerId: 'openrouter' },
  { id: 'qwen/qwen3.6-plus:free', name: 'Qwen 3.6 Plus', icon: qwenIcon, free: true, recommended: false, providerId: 'openrouter' },
];

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
    models: ALL_MODELS,
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
