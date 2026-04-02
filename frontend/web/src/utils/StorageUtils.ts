/**
 * CrewAI - Local Storage Utility
 * Утилиты для сохранения и загрузки данных проекта
 */

import type { ProjectData } from '../models/Types';

const STORAGE_KEY = 'crewai_project';
const STORAGE_KEY_SETTINGS = 'crewai_settings';

/**
 * Сохраняет проект в localStorage
 * @param project Данные проекта для сохранения
 */
export function saveProject(project: ProjectData): void {
    try {
        const serialized = JSON.stringify(project);
        localStorage.setItem(STORAGE_KEY, serialized);
        console.log('[Storage] Project saved successfully');
    } catch (error) {
        console.error('[Storage] Failed to save project:', error);
        throw new Error('Failed to save project to local storage');
    }
}

/**
 * Загружает проект из localStorage
 * @returns Данные проекта или null если проект не найден
 */
export function loadProject(): ProjectData | null {
    try {
        const serialized = localStorage.getItem(STORAGE_KEY);
        if (!serialized) {
            return null;
        }
        const project = JSON.parse(serialized) as ProjectData;
        console.log('[Storage] Project loaded successfully');
        return project;
    } catch (error) {
        console.error('[Storage] Failed to load project:', error);
        return null;
    }
}

/**
 * Удаляет проект из localStorage
 */
export function deleteProject(): void {
    try {
        localStorage.removeItem(STORAGE_KEY);
        console.log('[Storage] Project deleted successfully');
    } catch (error) {
        console.error('[Storage] Failed to delete project:', error);
    }
}

/**
 * Проверяет, существует ли сохраненный проект
 * @returns True если проект существует
 */
export function hasSavedProject(): boolean {
    return localStorage.getItem(STORAGE_KEY) !== null;
}

/**
 * Сохраняет настройки приложения
 * @param settings Настройки для сохранения
 */
export function saveSettings(settings: Record<string, unknown>): void {
    try {
        const serialized = JSON.stringify(settings);
        localStorage.setItem(STORAGE_KEY_SETTINGS, serialized);
        console.log('[Storage] Settings saved successfully');
    } catch (error) {
        console.error('[Storage] Failed to save settings:', error);
    }
}

/**
 * Загружает настройки приложения
 * @returns Настройки или пустой объект если настройки не найдены
 */
export function loadSettings(): Record<string, unknown> {
    try {
        const serialized = localStorage.getItem(STORAGE_KEY_SETTINGS);
        if (!serialized) {
            return {};
        }
        const settings = JSON.parse(serialized) as Record<string, unknown>;
        console.log('[Storage] Settings loaded successfully');
        return settings;
    } catch (error) {
        console.error('[Storage] Failed to load settings:', error);
        return {};
    }
}

/**
 * Экспортирует проект в JSON файл для скачивания
 * @param project Данные проекта для экспорта
 * @param filename Имя файла для экспорта
 */
export function exportProjectToFile(project: ProjectData, filename: string): void {
    try {
        const serialized = JSON.stringify(project, null, 2);
        const blob = new Blob([serialized], { type: 'application/json' });
        const url = URL.createObjectURL(blob);
        const link = document.createElement('a');
        link.href = url;
        link.download = filename.endsWith('.json') ? filename : `${filename}.json`;
        document.body.appendChild(link);
        link.click();
        document.body.removeChild(link);
        URL.revokeObjectURL(url);
        console.log('[Storage] Project exported successfully');
    } catch (error) {
        console.error('[Storage] Failed to export project:', error);
        throw new Error('Failed to export project');
    }
}

/**
 * Импортирует проект из JSON файла
 * @param file Файл для импорта
 * @returns Промис с данными проекта
 */
export function importProjectFromFile(file: File): Promise<ProjectData> {
    return new Promise((resolve, reject) => {
        const reader = new FileReader();

        reader.onload = (event) => {
            try {
                if (!event.target?.result) {
                    reject(new Error('Failed to read file'));
                    return;
                }
                const project = JSON.parse(event.target.result as string) as ProjectData;

                // Валидация структуры проекта
                if (!project.nodes || !project.connections) {
                    reject(new Error('Invalid project structure'));
                    return;
                }

                console.log('[Storage] Project imported successfully');
                resolve(project);
            } catch (error) {
                console.error('[Storage] Failed to parse project file:', error);
                reject(new Error('Failed to parse project file'));
            }
        };

        reader.onerror = () => {
            console.error('[Storage] Failed to read file');
            reject(new Error('Failed to read file'));
        };

        reader.readAsText(file);
    });
}

/**
 * Очищает все данные из localStorage
 */
export function clearAllStorage(): void {
    try {
        localStorage.removeItem(STORAGE_KEY);
        localStorage.removeItem(STORAGE_KEY_SETTINGS);
        console.log('[Storage] All storage cleared');
    } catch (error) {
        console.error('[Storage] Failed to clear storage:', error);
    }
}

/**
 * Получает размер хранилища в байтах
 * @returns Размер хранилища
 */
export function getStorageSize(): number {
    let totalSize = 0;
    for (let i = 0; i < localStorage.length; i++) {
        const key = localStorage.key(i);
        if (key) {
            const value = localStorage.getItem(key);
            if (value) {
                totalSize += (key.length + value.length) * 2; // UTF-16 encoding
            }
        }
    }
    return totalSize;
}

/**
 * Форматирует размер хранилища в человекочитаемый формат
 * @param bytes Размер в байтах
 * @returns Форматированная строка
 */
export function formatStorageSize(bytes: number): string {
    if (bytes < 1024) {
        return `${bytes} B`;
    }
    if (bytes < 1024 * 1024) {
        return `${(bytes / 1024).toFixed(2)} KB`;
    }
    return `${(bytes / (1024 * 1024)).toFixed(2)} MB`;
}
