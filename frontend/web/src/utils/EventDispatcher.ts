/**
 * CrewAI - Event Dispatcher
 * Универсальная система событий для коммуникации между компонентами
 */

type EventCallback<T = unknown> = (data: T) => void;

/**
 * Класс для управления событиями
 */
export class EventDispatcher {
    private events: Map<string, Set<EventCallback>> = new Map();

    /**
     * Подписывается на событие
     * @param event Имя события
     * @param callback Функция обратного вызова
     */
    public on<T = unknown>(event: string, callback: EventCallback<T>): void {
        if (!this.events.has(event)) {
            this.events.set(event, new Set());
        }
        this.events.get(event)!.add(callback as EventCallback);
    }

    /**
     * Отписывается от события
     * @param event Имя события
     * @param callback Функция обратного вызова
     */
    public off<T = unknown>(event: string, callback: EventCallback<T>): void {
        const eventCallbacks = this.events.get(event);
        if (eventCallbacks) {
            eventCallbacks.delete(callback as EventCallback);
            if (eventCallbacks.size === 0) {
                this.events.delete(event);
            }
        }
    }

    /**
     * Подписывается на событие один раз
     * @param event Имя события
     * @param callback Функция обратного вызова
     */
    public once<T = unknown>(event: string, callback: EventCallback<T>): void {
        const onceWrapper: EventCallback<T> = (data: T) => {
            this.off(event, onceWrapper);
            callback(data);
        };
        this.on(event, onceWrapper);
    }

    /**
     * Вызывает событие
     * @param event Имя события
     * @param data Данные события
     */
    public emit<T = unknown>(event: string, data?: T): void {
        const eventCallbacks = this.events.get(event);
        if (eventCallbacks) {
            eventCallbacks.forEach((callback) => {
                try {
                    callback(data as T);
                } catch (error) {
                    console.error(`Error in event callback for "${event}":`, error);
                }
            });
        }
    }

    /**
     * Проверяет, есть ли подписчики на событие
     * @param event Имя события
     * @returns True если есть подписчики
     */
    public hasListeners(event: string): boolean {
        const eventCallbacks = this.events.get(event);
        return eventCallbacks !== undefined && eventCallbacks.size > 0;
    }

    /**
     * Получает количество подписчиков на событие
     * @param event Имя события
     * @returns Количество подписчиков
     */
    public getListenerCount(event: string): number {
        const eventCallbacks = this.events.get(event);
        return eventCallbacks ? eventCallbacks.size : 0;
    }

    /**
     * Удаляет все подписки на событие
     * @param event Имя события
     */
    public removeAllListeners(event?: string): void {
        if (event) {
            this.events.delete(event);
        } else {
            this.events.clear();
        }
    }
}

/**
 * Типы событий приложения
 */
export const AppEvents = {
    // События нод
    NODE_CREATED: 'node:created',
    NODE_DELETED: 'node:deleted',
    NODE_SELECTED: 'node:selected',
    NODE_DESELECTED: 'node:deselected',
    NODE_MOVED: 'node:moved',
    NODE_UPDATED: 'node:updated',

    // События соединений
    CONNECTION_CREATED: 'connection:created',
    CONNECTION_DELETED: 'connection:deleted',
    CONNECTION_SELECTED: 'connection:selected',

    // События канваса
    CANVAS_ZOOM_CHANGED: 'canvas:zoomChanged',
    CANVAS_PANNED: 'canvas:panned',
    CANVAS_CLICKED: 'canvas:clicked',

    // События выделения
    SELECTION_CLEARED: 'selection:cleared',

    // События проекта
    PROJECT_SAVED: 'project:saved',
    PROJECT_LOADED: 'project:loaded',
    PROJECT_CHANGED: 'project:changed',

    // События UI
    SIDEBAR_TOGGLED: 'ui:sidebarToggled',
    PROPERTIES_UPDATED: 'ui:propertiesUpdated',
    TOAST_SHOW: 'ui:toastShow',
} as const;
