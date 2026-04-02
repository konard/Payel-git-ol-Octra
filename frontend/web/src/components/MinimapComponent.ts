/**
 * CrewAI - Minimap Component
 * Компонент мини-карты для навигации по канвасу
 */

import { BaseComponent } from './BaseComponent';
import type { Point, Rectangle } from '../models/Types';

export interface MinimapOptions {
    width?: number;
    height?: number;
}

/**
 * Компонент мини-карты
 */
export class MinimapComponent extends BaseComponent<HTMLDivElement> {
    private canvas: HTMLCanvasElement | null = null;
    private viewportElement: HTMLDivElement | null = null;
    private options: MinimapOptions;
    private contentBounds: Rectangle | null = null;
    private viewportBounds: Rectangle | null = null;
    private scale = 1;
    private offset: Point = { x: 0, y: 0 };

    private onViewportChangeCallbacks: Array<(viewport: Rectangle) => void> = [];
    private isDraggingViewport = false;
    private dragOffset: Point = { x: 0, y: 0 };

    /**
     * Конструктор мини-карты
     * @param options Опции мини-карты
     */
    constructor(options: MinimapOptions = {}) {
        super('div', 'crewai-minimap');

        this.options = {
            width: options.width || 200,
            height: options.height || 150,
        };

        this.createStructure();
        this.setupEventListeners();
    }

    /**
     * Обновляет содержимое мини-карты
     * @param nodes Данные о позициях нод
     * @param canvasBounds Границы видимой области канваса
     * @param contentBounds Границы всего контента
     */
    public update(
        nodes: Array<{ x: number; y: number; width: number; height: number; color: string }>,
        canvasBounds: Rectangle,
        contentBounds: Rectangle
    ): void {
        this.contentBounds = contentBounds;

        // Вычисляем масштаб
        if (contentBounds.width > 0 && contentBounds.height > 0) {
            const scaleX = this.options.width! / contentBounds.width;
            const scaleY = this.options.height! / contentBounds.height;
            this.scale = Math.min(scaleX, scaleY, 1);

            // Центрируем контент
            const scaledWidth = contentBounds.width * this.scale;
            const scaledHeight = contentBounds.height * this.scale;
            this.offset = {
                x: (this.options.width! - scaledWidth) / 2 - contentBounds.x * this.scale,
                y: (this.options.height! - scaledHeight) / 2 - contentBounds.y * this.scale,
            };
        }

        // Рендерим ноды
        this.renderNodes(nodes);

        // Обновляем viewport
        this.updateViewport(canvasBounds);
    }

    /**
     * Обновляет viewport
     * @param canvasBounds Границы видимой области
     */
    public updateViewport(canvasBounds: Rectangle): void {
        if (!this.viewportElement || !this.contentBounds) {
            return;
        }

        const viewportX = (canvasBounds.x - this.contentBounds.x) * this.scale + this.offset.x;
        const viewportY = (canvasBounds.y - this.contentBounds.y) * this.scale + this.offset.y;
        const viewportWidth = canvasBounds.width * this.scale;
        const viewportHeight = canvasBounds.height * this.scale;

        this.viewportBounds = {
            x: viewportX,
            y: viewportY,
            width: viewportWidth,
            height: viewportHeight,
        };

        this.viewportElement.style.left = `${viewportX}px`;
        this.viewportElement.style.top = `${viewportY}px`;
        this.viewportElement.style.width = `${viewportWidth}px`;
        this.viewportElement.style.height = `${viewportHeight}px`;
    }

    /**
     * Подписывается на изменение viewport
     * @param callback Функция обратного вызова
     */
    public onViewportChange(callback: (viewport: Rectangle) => void): void {
        this.onViewportChangeCallbacks.push(callback);
    }

    /**
     * Уничтожает компонент
     */
    protected onDestroy(): void {
        this.onViewportChangeCallbacks = [];
        this.canvas = null;
        this.viewportElement = null;
    }

    /**
     * Создает структуру мини-карты
     */
    private createStructure(): void {
        // Canvas для рендеринга
        this.canvas = document.createElement('canvas');
        this.canvas.className = 'crewai-minimap__canvas';
        this.canvas.width = this.options.width!;
        this.canvas.height = this.options.height!;
        this.canvas.style.width = '100%';
        this.canvas.style.height = '100%';
        this.element.appendChild(this.canvas);

        // Viewport индикатор
        this.viewportElement = document.createElement('div');
        this.viewportElement.className = 'crewai-minimap__viewport';
        this.viewportElement.style.position = 'absolute';
        this.viewportElement.style.border = '2px solid var(--color-primary)';
        this.viewportElement.style.background = 'var(--color-primary-light)';
        this.viewportElement.style.pointerEvents = 'none';
        this.element.appendChild(this.viewportElement);
    }

    /**
     * Рендерит ноды на мини-карте
     * @param nodes Данные о нодах
     */
    private renderNodes(
        nodes: Array<{ x: number; y: number; width: number; height: number; color: string }>
    ): void {
        if (!this.canvas || !this.contentBounds) {
            return;
        }

        const ctx = this.canvas.getContext('2d');
        if (!ctx) {
            return;
        }

        // Очищаем canvas
        ctx.clearRect(0, 0, this.canvas.width, this.canvas.height);

        // Рендерим ноды
        nodes.forEach((node) => {
            const x = (node.x - this.contentBounds!.x) * this.scale + this.offset.x;
            const y = (node.y - this.contentBounds!.y) * this.scale + this.offset.y;
            const width = node.width * this.scale;
            const height = node.height * this.scale;

            // Рисуем ноду
            ctx.fillStyle = node.color || 'var(--color-primary)';
            ctx.globalAlpha = 0.7;
            ctx.fillRect(x, y, width, height);
            ctx.globalAlpha = 1;
        });
    }

    /**
     * Настраивает обработчики событий
     */
    private setupEventListeners(): void {
        this.element.addEventListener('mousedown', this.handleMouseDown);
        this.element.addEventListener('mousemove', this.handleMouseMove);
        this.element.addEventListener('mouseup', this.handleMouseUp);
        this.element.addEventListener('mouseleave', this.handleMouseUp);
    }

    /**
     * Обработчик нажатия мыши
     */
    private handleMouseDown = (event: MouseEvent): void => {
        const rect = this.element.getBoundingClientRect();
        const x = event.clientX - rect.left;
        const y = event.clientY - rect.top;

        // Проверяем, кликнули ли внутри viewport
        if (this.viewportBounds) {
            const inViewport =
                x >= this.viewportBounds.x &&
                x <= this.viewportBounds.x + this.viewportBounds.width &&
                y >= this.viewportBounds.y &&
                y <= this.viewportBounds.y + this.viewportBounds.height;

            if (inViewport) {
                this.isDraggingViewport = true;
                this.dragOffset = {
                    x: x - this.viewportBounds.x,
                    y: y - this.viewportBounds.y,
                };
                this.element.style.cursor = 'grabbing';
            }
        }
    };

    /**
     * Обработчик движения мыши
     */
    private handleMouseMove = (event: MouseEvent): void => {
        if (!this.isDraggingViewport || !this.viewportBounds || !this.contentBounds) {
            return;
        }

        const rect = this.element.getBoundingClientRect();
        const x = event.clientX - rect.left;
        const y = event.clientY - rect.top;

        // Вычисляем новую позицию viewport
        let newX = x - this.dragOffset.x;
        let newY = y - this.dragOffset.y;

        // Ограничиваем пределами мини-карты
        newX = Math.max(0, Math.min(newX, this.options.width! - this.viewportBounds.width));
        newY = Math.max(0, Math.min(newY, this.options.height! - this.viewportBounds.height));

        // Преобразуем в координаты канваса
        const canvasX =
            (newX - this.offset.x) / this.scale + this.contentBounds.x;
        const canvasY =
            (newY - this.offset.y) / this.scale + this.contentBounds.y;

        // Вызываем колбэки
        this.onViewportChangeCallbacks.forEach((cb) =>
            cb({
                x: canvasX,
                y: canvasY,
                width: this.viewportBounds.width / this.scale,
                height: this.viewportBounds.height / this.scale,
            })
        );
    };

    /**
     * Обработчик отпускания мыши
     */
    private handleMouseUp = (): void => {
        this.isDraggingViewport = false;
        this.element.style.cursor = '';
    };

    /**
     * Показывает мини-карту
     */
    public show(): void {
        this.element.style.display = '';
    }

    /**
     * Скрывает мини-карту
     */
    public hide(): void {
        this.element.style.display = 'none';
    }
}
