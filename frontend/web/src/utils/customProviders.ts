import type { CustomModel, CustomProvider } from '../stores/customProvidersStore';

type RawRecord = Record<string, unknown>;

export type CustomProviderInput = Partial<CustomProvider> & RawRecord;
export type CustomModelInput = Partial<CustomModel> & RawRecord;

interface NormalizeOptions {
  fallbackId?: string;
}

function asRecord(value: unknown): RawRecord | null {
  return value && typeof value === 'object' ? (value as RawRecord) : null;
}

function firstString(values: unknown[], allowEmpty = false): string | undefined {
  for (const value of values) {
    if (typeof value !== 'string') {
      continue;
    }

    const normalized = value.trim();
    if (normalized || allowEmpty) {
      return normalized;
    }
  }

  return undefined;
}

function firstBoolean(values: unknown[]): boolean | undefined {
  for (const value of values) {
    if (typeof value === 'boolean') {
      return value;
    }
  }

  return undefined;
}

function firstTimestamp(values: unknown[]): string | undefined {
  for (const value of values) {
    if (typeof value === 'string' && value.trim()) {
      return value;
    }

    if (value instanceof Date) {
      return value.toISOString();
    }
  }

  return undefined;
}

export function normalizeCustomProvider(
  rawProvider: unknown,
  options: NormalizeOptions = {},
): CustomProvider | null {
  const provider = asRecord(rawProvider);
  if (!provider) {
    return null;
  }

  const id = firstString([provider.ID, provider.id, options.fallbackId]);
  const name = firstString([provider.Name, provider.name]);
  const baseUrl = firstString([provider.BaseURL, provider.base_url, provider.baseUrl]);

  if (!id || !name || !baseUrl) {
    return null;
  }

  const createdAt =
    firstTimestamp([provider.CreatedAt, provider.created_at, provider.createdAt]) ||
    new Date().toISOString();

  return {
    id,
    user_id: firstString([provider.UserID, provider.user_id, provider.userId]) || 'local',
    name,
    base_url: baseUrl,
    api_key: firstString([provider.APIKey, provider.api_key, provider.apiKey], true) || '',
    requires_api_key:
      firstBoolean([provider.RequiresApiKey, provider.requires_api_key, provider.requiresApiKey]) ??
      true,
    created_at: createdAt,
    updated_at:
      firstTimestamp([provider.UpdatedAt, provider.updated_at, provider.updatedAt]) || createdAt,
  };
}

export function normalizeCustomModel(
  rawModel: unknown,
  options: NormalizeOptions = {},
): CustomModel | null {
  const model = asRecord(rawModel);
  if (!model) {
    return null;
  }

  const id = firstString([model.ID, model.id, options.fallbackId]);
  const name = firstString([model.Name, model.name]);

  if (!id || !name) {
    return null;
  }

  const createdAt =
    firstTimestamp([model.CreatedAt, model.created_at, model.createdAt]) ||
    new Date().toISOString();

  return {
    id,
    user_id: firstString([model.UserID, model.user_id, model.userId]) || 'local',
    name,
    provider_id: firstString([model.ProviderID, model.provider_id, model.providerId]),
    created_at: createdAt,
    updated_at: firstTimestamp([model.UpdatedAt, model.updated_at, model.updatedAt]) || createdAt,
  };
}

export function normalizeCustomProviderList(rawProviders: unknown): CustomProvider[] {
  if (!Array.isArray(rawProviders)) {
    return [];
  }

  const providersById = new Map<string, CustomProvider>();
  for (const rawProvider of rawProviders) {
    const provider = normalizeCustomProvider(rawProvider);
    if (provider) {
      providersById.set(provider.id, provider);
    }
  }

  return Array.from(providersById.values());
}

export function normalizeCustomModelList(rawModels: unknown): CustomModel[] {
  if (!Array.isArray(rawModels)) {
    return [];
  }

  const modelsById = new Map<string, CustomModel>();
  for (const rawModel of rawModels) {
    const model = normalizeCustomModel(rawModel);
    if (model) {
      modelsById.set(model.id, model);
    }
  }

  return Array.from(modelsById.values());
}

export function upsertCustomProvider(
  providers: CustomProvider[],
  provider: CustomProvider,
): CustomProvider[] {
  const normalizedProviders = normalizeCustomProviderList(providers);
  const existingIndex = normalizedProviders.findIndex((existing) => existing.id === provider.id);

  if (existingIndex === -1) {
    return [...normalizedProviders, provider];
  }

  const nextProviders = [...normalizedProviders];
  nextProviders[existingIndex] = { ...nextProviders[existingIndex], ...provider };
  return nextProviders;
}

export function upsertCustomModel(models: CustomModel[], model: CustomModel): CustomModel[] {
  const normalizedModels = normalizeCustomModelList(models);
  const existingIndex = normalizedModels.findIndex((existing) => existing.id === model.id);

  if (existingIndex === -1) {
    return [...normalizedModels, model];
  }

  const nextModels = [...normalizedModels];
  nextModels[existingIndex] = { ...nextModels[existingIndex], ...model };
  return nextModels;
}

export function normalizeCustomProvidersState(state: unknown): {
  providers: CustomProvider[];
  models: CustomModel[];
} {
  const record = asRecord(state);
  const stateRecord = asRecord(record?.state) || record;

  return {
    providers: normalizeCustomProviderList(stateRecord?.providers),
    models: normalizeCustomModelList(stateRecord?.models),
  };
}
