/**
 * User Settings Service
 * Сервис для сохранения и загрузки настроек пользователя
 */

const SETTINGS_API_URL = import.meta.env.VITE_SETTINGS_URL || '/api/settings';

export interface UserSettings {
  language: string;
  theme: 'dark' | 'light';
  hide_api_key: boolean;
  hide_server_status: boolean;
  hide_console: boolean;
}

/**
 * Сохранить настройки пользователя
 */
export async function saveUserSettings(settings: Partial<UserSettings>, token: string): Promise<void> {
  const response = await fetch(`${SETTINGS_API_URL}/update`, {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`,
    },
    body: JSON.stringify(settings),
  });

  if (!response.ok) {
    throw new Error('Failed to save settings');
  }
}

/**
 * Загрузить настройки пользователя
 */
export async function loadUserSettings(token: string): Promise<UserSettings> {
  const response = await fetch(`${SETTINGS_API_URL}/get`, {
    method: 'GET',
    headers: {
      'Authorization': `Bearer ${token}`,
    },
  });

  if (!response.ok) {
    throw new Error('Failed to load settings');
  }

  const data = await response.json();
  return data.settings;
}

/**
 * Обновить язык пользователя
 */
export async function updateUserLanguage(language: string, token: string): Promise<void> {
  await saveUserSettings({ language }, token);
}

/**
 * Обновить тему пользователя
 */
export async function updateUserTheme(theme: 'dark' | 'light', token: string): Promise<void> {
  await saveUserSettings({ theme }, token);
}
