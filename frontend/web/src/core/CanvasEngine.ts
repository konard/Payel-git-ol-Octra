/**
 * CrewAI - Canvas Engine
 * Движок канваса для рендеринга, зума и панорамирования
 */

import type { Point, CanvasSettings } from '../models/Types';
import { EventDispatcher, AppEvents } from '../utils/EventDispatcher';
import { clamp } from '../utils/MathUtils';

/**
 * Класс управления канвасом
 * Обрабатывает зум, панорамирование и преобразование координат
 */
export class CanvasEngine {
    private container: HTMLElement | null = null;
    private content: HTMLElement | null = null;
    private settings: CanvasSettings;
    private eventDispatcher: EventDispatcher = new EventDispatcher();
    private isPanning = false;
    private panStart: Point = { x: 0, y: 0 };
    private panOrigin: Point = { x: 0, y: 0 };

    /**
     * Конструктор CanvasEngine
     * @param settings Начальные настройки канваса
     */
    constructor(settings?: Partial<CanvasSettings>) {
        this.settings = {
            zoom: settings?.zoom ?? 1,
            pan: settings?.pan ?? { x: 0, y: 0 },
            showGrid: settings?.showGrid ?? true,
            snapToGrid: settings?.snapToGrid ?? true,
            gridSize: settings?.gridSize ?? 20,
        };
    }

    /**
     * Инициализирует канвас
     * @param container Контейнер канваса
     * @param content Контент канваса
     */
    public initialize(container: HTMLElement, content: HTMLElement): void {
        this.container = container;
        this.content = content;
        this.setupEventListeners();
        this.updateTransform();
    }

    /**
     * Уничтожает канвас
     */
    public destroy(): void {
        this.removeEventListeners();
        this.container = null;
        this.content = null;
    }

    // === Zoom API ===

    /**
     * Устанавливает зум
     * @param zoom Уровень зума (0.1 - 3)
     */
    public setZoom(zoom: number): void {
        const clampedZoom = clamp(zoom, 0.1, 3);
        if (this.settings.zoom !== clampedZoom) {
            this.settings.zoom = clampedZoom;
            this.updateTransform();
            this.eventDispatcher.emit(AppEvents.CANVAS_ZOOM_CHANGED, { zoom: clampedZoom });
        }
    }

    /**
     * Получает текущий зум
     * @returns Уровень зума
     */
    public getZoom(): number {
        return this.settings.zoom;
    }

    /**
     * Увеличивает зум
     * @param delta Шаг зума
     */
    public zoomIn(delta: number = 0.1): void {
        this.setZoom(this.settings.zoom + delta);
    }

    /**
     * Уменьшает зум
     * @param delta Шаг зума
     */
    public zoomOut(delta: number = 0.1): void {
        this.setZoom(this.settings.zoom - delta);
    }

    /**
     * Сбрасывает зум к 100%
     */
    public resetZoom(): void {
        this.setZoom(1);
    }

    /**
     * Зумирует к точке
     * @param point Точка для зумирования
     * @param zoom Целевой зум
     */
    public zoomToPoint(point: Point, zoom: number): void {
        if (!this.container || !this.content) {
            return;
        }

        const oldZoom = this.settings.zoom;
        const newZoom = clamp(zoom, 0.1, 3);
        const scale = newZoom / oldZoom;

        // Вычисляем новую позицию панорамирования
        const rect = this.container.getBoundingClientRect();
        const centerX = rect.width / 2;
        const centerY = rect.height / 2;

        const deltaX = (point.x - this.settings.pan.x - centerX) * (1 - scale);
        const deltaY = (point.y - this.settings.pan.y - centerY) * (1 - scale);

        this.settings.zoom = newZoom;
        this.settings.pan.x += deltaX;
        this.settings.pan.y += deltaY;

        this.updateTransform();
        this.eventDispatcher.emit(AppEvents.CANVAS_ZOOM_CHANGED, { zoom: newZoom });
    }

    /**
     * Подгоняет зум для отображения всех элементов
     * @param bounds Границы содержимого
     */
    public fitToContent(bounds: { x: number; y: number; width: number; height: number }): void {
        if (!this.container) {
            return;
        }

        const rect = this.container.getBoundingClientRect();
        const padding = 40;

        const scaleX = (rect.width - padding * 2) / bounds.width;
        const scaleY = (rect.height - padding * 2) / bounds.height;
        const newZoom = Math.min(scaleX, scaleY, 2);

        this.settings.zoom = clamp(newZoom, 0.1, 3);
        this.settings.pan.x = (rect.width - bounds.width * newZoom) / 2 - bounds.x * newZoom;
        this.settings.pan.y = (rect.height - bounds.height * newZoom) / 2 - bounds.y * newZoom;

        this.updateTransform();
        this.eventDispatcher.emit(AppEvents.CANVAS_ZOOM_CHANGED, { zoom: newZoom });
    }

    // === Pan API ===

    /**
     * Устанавливает позицию панорамирования
     * @param pan Новая позиция
     */
    public setPan(pan: Point): void {
        this.settings.pan = { ...pan };
        this.updateTransform();
        this.eventDispatcher.emit(AppEvents.CANVAS_PANNED, { pan });
    }

    /**
     * Получает текущую позицию панорамирования
     * @returns Позиция панорамирования
     */
    public getPan(): Point {
        return { ...this.settings.pan };
    }

    /**
     * Сбрасывает панорамирование
     */
    public resetPan(): void {
        this.setPan({ x: 0, y: 0 });
    }

    /**
     * Начинает панорамирование
     * @param startPosition Начальная позиция мыши
     */
    public startPan(startPosition: Point): void {
        this.isPanning = true;
        this.panStart = { ...startPosition };
        this.panOrigin = { ...this.settings.pan };
        document.body.style.cursor = 'grabbing';
    }

    /**
     * Продолжает панорамирование
     * @param currentPosition Текущая позиция мыши
     */
    public panTo(currentPosition: Point): void {
        if (!this.isPanning) {
            return;
        }

        const deltaX = currentPosition.x - this.panStart.x;
        const deltaY = currentPosition.y - this.panStart.y;

        this.settings.pan.x = this.panOrigin.x + deltaX;
        this.settings.pan.y = this.panOrigin.y + deltaY;

        this.updateTransform();
        this.eventDispatcher.emit(AppEvents.CANVAS_PANNED, { pan: this.settings.pan });
    }

    /**
     * Завершает панорамирование
     */
    public endPan(): void {
        this.isPanning = false;
        document.body.style.cursor = '';
    }

    /**
     * Проверяет, идет ли панорамирование
     * @returns True если идет панорамирование
     */
    public isPanningActive(): boolean {
        return this.isPanning;
    }

    // === Coordinate Transformation ===

    /**
     * Преобразует экранные координаты в координаты канваса
     * @param screenPoint Точка на экране
     * @returns Точка на канвасе
     */
    public screenToCanvas(screenPoint: Point): Point {
        if (!this.container) {
            return { ...screenPoint };
        }

        const rect = this.container.getBoundingClientRect();
        return {
            x: (screenPoint.x - rect.left - this.settings.pan.x) / this.settings.zoom,
            y: (screenPoint.y - rect.top - this.settings.pan.y) / this.settings.zoom,
        };
    }

    /**
     * Преобразует координаты канваса в экранные
     * @param canvasPoint Точка на канвасе
     * @returns Точка на экране
     */
    public canvasToScreen(canvasPoint: Point): Point {
        if (!this.container) {
            return { ...canvasPoint };
        }

        const rect = this.container.getBoundingClientRect();
        return {
            x: canvasPoint.x * this.settings.zoom + this.settings.pan.x + rect.left,
            y: canvasPoint.y * this.settings.zoom + this.settings.pan.y + rect.top,
        };
    }

    /**
     * Привязывает точку к сетке
     * @param point Точка для привязки
     * @returns Привязанная точка
     */
    public snapToGrid(point: Point): Point {
        if (!this.settings.snapToGrid) {
            return { ...point };
        }

        const gridSize = this.settings.gridSize;
        return {
            x: Math.round(point.x / gridSize) * gridSize,
            y: Math.round(point.y / gridSize) * gridSize,
        };
    }

    // === Settings API ===

    /**
     * Получает настройки канваса
     * @returns Настройки канваса
     */
    public getSettings(): CanvasSettings {
        return { ...this.settings };
    }

    /**
     * Обновляет настройки канваса
     * @param settings Новые настройки
     */
    public updateSettings(settings: Partial<CanvasSettings>): void {
        this.settings = { ...this.settings, ...settings };
        this.updateTransform();

        if (settings.zoom !== undefined) {
            this.eventDispatcher.emit(AppEvents.CANVAS_ZOOM_CHANGED, { zoom: settings.zoom });
        }
        if (settings.pan !== undefined) {
            this.eventDispatcher.emit(AppEvents.CANVAS_PANNED, { pan: settings.pan });
        }
    }

    /**
     * Переключает отображение сетки
     */
    public toggleGrid(): void {
        this.settings.showGrid = !this.settings.showGrid;
        if (this.content) {
            this.content.style.backgroundImage = this.settings.showGrid
                ? `linear-gradient(rgba(58, 58, 92, 0.3) 1px, transparent 1px),
                   linear-gradient(90deg, rgba(58, 58, 92, 0.3) 1px, transparent 1px)`
                : 'none';
            this.content.style.backgroundSize = `${this.settings.gridSize}px ${this.settings.gridSize}px`;
        }
    }

    /**
     * Переключает привязку к сетке
     */
    public toggleSnapToGrid(): void {
        this.settings.snapToGrid = !this.settings.snapToGrid;
    }

    // === Event System ===

    /**
     * Подписывается на событие
     * @param event Имя события
     * @param callback Функция обратного вызова
     */
    public on<T = unknown>(event: string, callback: (data: T) => void): void {
        this.eventDispatcher.on(event, callback);
    }

    /**
     * Отписывается от события
     * @param event Имя события
     * @param callback Функция обратного вызова
     */
    public off<T = unknown>(event: string, callback: (data: T) => void): void {
        this.eventDispatcher.off(event, callback);
    }

    // === Private Methods ===

    /**
     * Обновляет трансформацию контента
     */
    private updateTransform(): void {
        if (!this.content) {
            return;
        }

        this.content.style.transform = `translate(${this.settings.pan.x}px, ${this.settings.pan.y}px) scale(${this.settings.zoom})`;
        this.content.style.transformOrigin = '0 0';

        // Обновляем размер сетки
        if (this.settings.showGrid) {
            const scaledGridSize = this.settings.gridSize * this.settings.zoom;
            this.content.style.backgroundSize = `${scaledGridSize}px ${scaledGridSize}px`;
        }
    }

    /**
     * Настраивает обработчики событий
     */
    private setupEventListeners(): void {
        if (!this.container) {
            return;
        }

        // Обработка колеса мыши для зума
        this.container.addEventListener('wheel', this.handleWheel, { passive: false });

        // Обработка двойного клика для сброса зума
        this.container.addEventListener('dblclick', this.handleDoubleClick);
    }

    /**
     * Удаляет обработчики событий
     */
    private removeEventListeners(): void {
        if (!this.container) {
            return;
        }

        this.container.removeEventListener('wheel', this.handleWheel);
        this.container.removeEventListener('dblclick', this.handleDoubleClick);
    }

    /**
     * Обработчик колеса мыши
     */
    private handleWheel = (event: WheelEvent): void => {
        event.preventDefault();

        const delta = event.deltaY > 0 ? -0.05 : 0.05;
        const rect = this.container!.getBoundingClientRect();
        const mousePoint: Point = {
            x: event.clientX - rect.left,
            y: event.clientY - rect.top,
        };

        const oldZoom = this.settings.zoom;
        const newZoom = clamp(oldZoom + delta, 0.1, 3);
        const scale = newZoom / oldZoom;

        // Зум к точке курсора
        const canvasPoint = this.screenToCanvas(mousePoint);
        this.settings.pan.x = mousePoint.x - canvasPoint.x * newZoom;
        this.settings.pan.y = mousePoint.y - canvasPoint.y * newZoom;
        this.settings.zoom = newZoom;

        this.updateTransform();
        this.eventDispatcher.emit(AppEvents.CANVAS_ZOOM_CHANGED, { zoom: newZoom });
    };

    /**
     * Обработчик двойного клика
     */
    private handleDoubleClick = (): void => {
        this.resetZoom();
        this.resetPan();
    };

    /**
     * Получает размеры видимой области канваса
     * @returns Размеры области
     */
    public getViewportBounds(): { width: number; height: number } {
        if (!this.container) {
            return { width: 0, height: 0 };
        }

        const rect = this.container.getBoundingClientRect();
        return {
            width: rect.width,
            height: rect.height,
        };
    }

    /**
     * Проверяет, видима ли точка на канвасе
     * @param point Точка для проверки
     * @param margin Дополнительный отступ
     * @returns True если точка видима
     */
    public isPointVisible(point: Point, margin: number = 0): boolean {
        if (!this.container) {
            return false;
        }

        const screenPoint = this.canvasToScreen(point);
        const rect = this.container.getBoundingClientRect();

        return (
            screenPoint.x >= rect.left - margin &&
            screenPoint.x <= rect.right + margin &&
            screenPoint.y >= rect.top - margin &&
            screenPoint.y <= rect.bottom + margin
        );
    }
}
