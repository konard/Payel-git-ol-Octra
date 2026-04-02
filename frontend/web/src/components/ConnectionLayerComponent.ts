/**
 * CrewAI - Connection Layer Component
 * Компонент для рендеринга соединений между нодами
 */

import { BaseComponent } from './BaseComponent';
import type { Point, ConnectionData } from '../models/Types';
import { bezierPoint, bezierLength } from '../utils/MathUtils';

export interface ConnectionLayerOptions {
    selectedConnectionId?: string;
}

/**
 * Компонент слоя соединений
 */
export class ConnectionLayerComponent extends BaseComponent<SVGSVGElement> {
    private connections: Map<string, SVGGElement> = new Map();
    private connectionPaths: Map<string, SVGPathElement> = new Map();
    private selectedConnectionId: string | null = null;
    private onClickCallbacks: Array<(connectionId: string) => void> = [];
    private onDblClickCallbacks: Array<(connectionId: string) => void> = [];

    /**
     * Конструктор слоя соединений
     * @param options Опции слоя
     */
    constructor(options: ConnectionLayerOptions = {}) {
        super('svg', 'crewai-connection-layer');

        this.element.setAttribute('xmlns', 'http://www.w3.org/2000/svg');
        this.element.style.position = 'absolute';
        this.element.style.top = '0';
        this.element.style.left = '0';
        this.element.style.width = '100%';
        this.element.style.height = '100%';
        this.element.style.pointerEvents = 'none';
        this.element.style.zIndex = '20';

        if (options.selectedConnectionId) {
            this.selectedConnectionId = options.selectedConnectionId;
        }
    }

    /**
     * Добавляет соединение
     * @param connectionId ID соединения
     * @param fromPoint Начальная точка
     * @param toPoint Конечная точка
     * @param isSelected Выделено ли соединение
     * @returns SVG группа соединения
     */
    public addConnection(
        connectionId: string,
        fromPoint: Point,
        toPoint: Point,
        isSelected = false
    ): SVGGElement {
        // Удаляем существующее соединение с таким ID
        this.removeConnection(connectionId);

        const group = document.createElementNS('http://www.w3.org/2000/svg', 'g');
        group.setAttribute('data-connection-id', connectionId);
        group.style.pointerEvents = 'stroke';
        group.style.cursor = 'pointer';

        // Создаем невидимую толстую линию для лучшего захвата
        const hitPath = document.createElementNS('http://www.w3.org/2000/svg', 'path');
        hitPath.setAttribute('class', 'crewai-connection__hit');
        hitPath.style.stroke = 'transparent';
        hitPath.style.strokeWidth = '20';
        hitPath.style.fill = 'none';
        hitPath.style.pointerEvents = 'stroke';
        group.appendChild(hitPath);

        // Создаем видимую линию соединения
        const path = document.createElementNS('http://www.w3.org/2000/svg', 'path');
        path.setAttribute('class', 'crewai-connection__path');
        path.style.pointerEvents = 'none';
        group.appendChild(path);

        // Добавляем обработчики событий
        group.addEventListener('click', (e) => {
            e.stopPropagation();
            this.onClickCallbacks.forEach((cb) => cb(connectionId));
        });

        group.addEventListener('dblclick', (e) => {
            e.stopPropagation();
            this.onDblClickCallbacks.forEach((cb) => cb(connectionId));
        });

        this.element.appendChild(group);
        this.connections.set(connectionId, group);
        this.connectionPaths.set(connectionId, path);

        // Обновляем путь
        this.updateConnectionPath(connectionId, fromPoint, toPoint);

        // Обновляем выделение
        if (isSelected) {
            this.selectConnection(connectionId);
        }

        return group;
    }

    /**
     * Обновляет путь соединения
     * @param connectionId ID соединения
     * @param fromPoint Начальная точка
     * @param toPoint Конечная точка
     */
    public updateConnectionPath(connectionId: string, fromPoint: Point, toPoint: Point): void {
        const group = this.connections.get(connectionId);
        const path = this.connectionPaths.get(connectionId);

        if (!group || !path) {
            return;
        }

        // Вычисляем кривую Безье
        const deltaX = toPoint.x - fromPoint.x;
        const controlPointOffset = Math.max(Math.abs(deltaX) * 0.5, 50);

        const controlPoint1: Point = {
            x: fromPoint.x + controlPointOffset,
            y: fromPoint.y,
        };
        const controlPoint2: Point = {
            x: toPoint.x - controlPointOffset,
            y: toPoint.y,
        };

        // Создаем путь кривой Безье
        const d = `M ${fromPoint.x} ${fromPoint.y} C ${controlPoint1.x} ${controlPoint1.y}, ${controlPoint2.x} ${controlPoint2.y}, ${toPoint.x} ${toPoint.y}`;

        // Обновляем оба пути (видимый и для захвата)
        const hitPath = group.querySelector('.crewai-connection__hit');
        if (hitPath) {
            hitPath.setAttribute('d', d);
        }
        path.setAttribute('d', d);

        // Вычисляем длину для анимации
        const length = bezierLength(fromPoint, controlPoint1, controlPoint2, toPoint);
        path.style.strokeDasharray = `${length} ${length}`;
    }

    /**
     * Обновляет все соединения
     * @param connections Данные соединений с координатами
     */
    public updateAllConnections(
        connections: Array<{
            id: string;
            fromPoint: Point;
            toPoint: Point;
            isSelected?: boolean;
        }>
    ): void {
        connections.forEach((conn) => {
            if (this.connections.has(conn.id)) {
                this.updateConnectionPath(conn.id, conn.fromPoint, conn.toPoint);
                if (conn.isSelected !== undefined) {
                    if (conn.isSelected) {
                        this.selectConnection(conn.id);
                    } else {
                        this.deselectConnection(conn.id);
                    }
                }
            } else {
                this.addConnection(conn.id, conn.fromPoint, conn.toPoint, conn.isSelected);
            }
        });
    }

    /**
     * Выделяет соединение
     * @param connectionId ID соединения
     */
    public selectConnection(connectionId: string): void {
        this.selectedConnectionId = connectionId;

        this.connections.forEach((group, id) => {
            const path = this.connectionPaths.get(id);
            if (!path) {
                return;
            }

            if (id === connectionId) {
                path.style.stroke = 'var(--color-primary)';
                path.style.strokeWidth = '3';
                path.style.filter = 'drop-shadow(0 0 4px var(--color-primary))';
                group.style.zIndex = '25';
            } else {
                path.style.stroke = '';
                path.style.strokeWidth = '';
                path.style.filter = '';
                group.style.zIndex = '';
            }
        });
    }

    /**
     * Снимает выделение с соединения
     * @param connectionId ID соединения
     */
    public deselectConnection(connectionId: string): void {
        const path = this.connectionPaths.get(connectionId);
        if (path) {
            path.style.stroke = '';
            path.style.strokeWidth = '';
            path.style.filter = '';
        }

        const group = this.connections.get(connectionId);
        if (group) {
            group.style.zIndex = '';
        }

        if (this.selectedConnectionId === connectionId) {
            this.selectedConnectionId = null;
        }
    }

    /**
     * Снимает выделение со всех соединений
     */
    public deselectAll(): void {
        this.connections.forEach((group, id) => {
            this.deselectConnection(id);
        });
    }

    /**
     * Удаляет соединение
     * @param connectionId ID соединения
     */
    public removeConnection(connectionId: string): void {
        const group = this.connections.get(connectionId);
        if (group) {
            group.remove();
            this.connections.delete(connectionId);
            this.connectionPaths.delete(connectionId);
        }

        if (this.selectedConnectionId === connectionId) {
            this.selectedConnectionId = null;
        }
    }

    /**
     * Удаляет все соединения
     */
    public clear(): void {
        this.connections.forEach((group) => group.remove());
        this.connections.clear();
        this.connectionPaths.clear();
        this.selectedConnectionId = null;
    }

    /**
     * Подписывается на клик по соединению
     * @param callback Функция обратного вызова
     */
    public onClick(callback: (connectionId: string) => void): void {
        this.onClickCallbacks.push(callback);
    }

    /**
     * Подписывается на двойной клик по соединению
     * @param callback Функция обратного вызова
     */
    public onDblClick(callback: (connectionId: string) => void): void {
        this.onDblClickCallbacks.push(callback);
    }

    /**
     * Отписывается от всех событий
     */
    public off(): void {
        this.onClickCallbacks = [];
        this.onDblClickCallbacks = [];
    }

    /**
     * Получает все ID соединений
     * @returns Массив ID соединений
     */
    public getConnectionIds(): string[] {
        return Array.from(this.connections.keys());
    }

    /**
     * Получает количество соединений
     * @returns Количество соединений
     */
    public getConnectionCount(): number {
        return this.connections.size;
    }

    /**
     * Проверяет, существует ли соединение
     * @param connectionId ID соединения
     * @returns True если соединение существует
     */
    public hasConnection(connectionId: string): boolean {
        return this.connections.has(connectionId);
    }

    /**
     * Уничтожает компонент
     */
    protected onDestroy(): void {
        this.off();
        this.clear();
    }

    /**
     * Создает временное соединение для предпросмотра
     * @param fromPoint Начальная точка
     * @param toPoint Конечная точка
     * @returns Временный путь
     */
    public createTempConnection(fromPoint: Point, toPoint: Point): SVGPathElement {
        const path = document.createElementNS('http://www.w3.org/2000/svg', 'path');
        path.setAttribute('class', 'crewai-connection__temp');
        path.style.stroke = 'var(--color-primary)';
        path.style.strokeWidth = '2';
        path.style.strokeDasharray = '10 5';
        path.style.fill = 'none';
        path.style.pointerEvents = 'none';
        path.style.opacity = '0.6';

        // Анимация
        path.style.animation = 'flowAnimation 1s linear infinite';

        this.element.appendChild(path);

        // Обновляем путь
        const deltaX = toPoint.x - fromPoint.x;
        const controlPointOffset = Math.max(Math.abs(deltaX) * 0.5, 50);

        const controlPoint1: Point = {
            x: fromPoint.x + controlPointOffset,
            y: fromPoint.y,
        };
        const controlPoint2: Point = {
            x: toPoint.x - controlPointOffset,
            y: toPoint.y,
        };

        const d = `M ${fromPoint.x} ${fromPoint.y} C ${controlPoint1.x} ${controlPoint1.y}, ${controlPoint2.x} ${controlPoint2.y}, ${toPoint.x} ${toPoint.y}`;
        path.setAttribute('d', d);

        return path;
    }

    /**
     * Удаляет временное соединение
     * @param path Путь для удаления
     */
    public removeTempConnection(path: SVGPathElement): void {
        if (path.parentNode) {
            path.parentNode.removeChild(path);
        }
    }
}
