/**
 * CrewAI - ID Generator Utility
 * Генерация уникальных идентификаторов для нод и соединений
 */

/**
 * Генерирует уникальный ID в формате UUID v4
 * @returns Уникальный идентификатор
 */
export function generateId(): string {
    return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, (char) => {
        const random = (Math.random() * 16) | 0;
        const value = char === 'x' ? random : (random & 0x3) | 0x8;
        return value.toString(16);
    });
}

/**
 * Генерирует короткий ID для отображения
 * @returns Короткий идентификатор (8 символов)
 */
export function generateShortId(): string {
    return Math.random().toString(36).substring(2, 10);
}

/**
 * Генерирует ID с префиксом
 * @param prefix Префикс для ID
 * @returns Уникальный идентификатор с префиксом
 */
export function generateIdWithPrefix(prefix: string): string {
    return `${prefix}_${generateShortId()}`;
}
