/**
 * CrewAI - Node Template Component
 * Компонент шаблона ноды для боковой панели
 */

import { BaseComponent } from './BaseComponent';
import type { AgentType, AgentCategory } from '../models/Types';
import { getAgentInfo } from '../models/AgentDefinitions';

export interface NodeTemplateOptions {
    type: AgentType;
    category: AgentCategory;
    compact?: boolean;
}

/**
 * Компонент шаблона ноды для перетаскивания
 */
export class NodeTemplateComponent extends BaseComponent<HTMLDivElement> {
    private type: AgentType;
    private category: AgentCategory;
    private iconElement: HTMLDivElement | null = null;
    private nameElement: HTMLSpanElement | null = null;
    private descriptionElement: HTMLSpanElement | null = null;
    private isDragging = false;
    private onDragStartCallbacks: Array<(type: AgentType) => void> = [];

    /**
     * Конструктор шаблона ноды
     * @param options Опции шаблона
     */
    constructor(options: NodeTemplateOptions) {
        super('div', 'crewai-node-template');

        this.type = options.type;
        this.category = options.category;

        if (options.compact) {
            this.addClass('crewai-node-template--compact');
        }

        this.setData('agent-type', type);
        this.setData('agent-category', category);

        this.createStructure();
        this.setupEventListeners();
    }

    /**
     * Получает тип агента
     * @returns Тип агента
     */
    public getAgentType(): AgentType {
        return this.type;
    }

    /**
     * Получает категорию агента
     * @returns Категория агента
     */
    public getAgentCategory(): AgentCategory {
        return this.category;
    }

    /**
     * Подписывается на начало перетаскивания
     * @param callback Функция обратного вызова
     */
    public onDragStart(callback: (type: AgentType) => void): void {
        this.onDragStartCallbacks.push(callback);
    }

    /**
     * Уничтожает компонент
     */
    protected onDestroy(): void {
        this.onDragStartCallbacks = [];
        this.iconElement = null;
        this.nameElement = null;
        this.descriptionElement = null;
    }

    /**
     * Создает структуру шаблона
     */
    private createStructure(): void {
        const agentInfo = getAgentInfo(this.type);

        if (!agentInfo) {
            return;
        }

        // Иконка
        this.iconElement = document.createElement('div');
        this.iconElement.className = 'crewai-node-template__icon';
        this.iconElement.textContent = agentInfo.icon;
        this.iconElement.style.background = agentInfo.color;
        this.element.appendChild(this.iconElement);

        // Информация
        const infoElement = document.createElement('div');
        infoElement.className = 'crewai-node-template__info';

        this.nameElement = document.createElement('span');
        this.nameElement.className = 'crewai-node-template__name';
        this.nameElement.textContent = agentInfo.label;
        infoElement.appendChild(this.nameElement);

        this.descriptionElement = document.createElement('span');
        this.descriptionElement.className = 'crewai-node-template__description';
        this.descriptionElement.textContent = agentInfo.description;
        infoElement.appendChild(this.descriptionElement);

        this.element.appendChild(infoElement);
    }

    /**
     * Настраивает обработчики событий
     */
    private setupEventListeners(): void {
        this.element.setAttribute('draggable', 'true');

        this.element.addEventListener('dragstart', this.handleDragStart);
        this.element.addEventListener('dragend', this.handleDragEnd);
        this.element.addEventListener('mouseenter', this.handleMouseEnter);
        this.element.addEventListener('mouseleave', this.handleMouseLeave);
    }

    /**
     * Обработчик начала перетаскивания
     */
    private handleDragStart = (event: DragEvent): void => {
        this.isDragging = true;
        this.addClass('crewai-node-template--dragging');

        // Настраиваем данные для drag-and-drop
        event.dataTransfer?.setData('application/x-crewai-agent-type', this.type);
        event.dataTransfer?.setData('application/x-crewai-agent-category', this.category);
        event.dataTransfer.effectAllowed = 'copy';

        // Создаем кастомный образ для перетаскивания
        const dragImage = this.createDragImage();
        event.dataTransfer?.setDragImage(dragImage, 20, 20);

        // Вызываем колбэки
        this.onDragStartCallbacks.forEach((cb) => cb(this.type));
    };

    /**
     * Обработчик окончания перетаскивания
     */
    private handleDragEnd = (): void => {
        this.isDragging = false;
        this.removeClass('crewai-node-template--dragging');
    };

    /**
     * Обработчик наведения мыши
     */
    private handleMouseEnter = (): void => {
        if (!this.isDragging) {
            this.addClass('crewai-node-template--hover');
        }
    };

    /**
     * Обработчик ухода мыши
     */
    private handleMouseLeave = (): void => {
        this.removeClass('crewai-node-template--hover');
    };

    /**
     * Создает изображение для перетаскивания
     * @returns Элемент для drag image
     */
    private createDragImage(): HTMLElement {
        const dragImage = document.createElement('div');
        dragImage.className = 'crewai-node-preview crewai-node--' + this.category;
        dragImage.style.padding = '12px 16px';
        dragImage.style.background = 'var(--color-bg-elevated)';
        dragImage.style.border = '2px solid var(--color-primary)';
        dragImage.style.borderRadius = 'var(--radius-lg)';
        dragImage.style.boxShadow = 'var(--shadow-xl)';
        dragImage.style.display = 'flex';
        dragImage.style.alignItems = 'center';
        dragImage.style.gap = '8px';
        dragImage.style.fontSize = '13px';
        dragImage.style.fontWeight = '600';

        const agentInfo = getAgentInfo(this.type);
        if (agentInfo) {
            dragImage.innerHTML = `
                <span style="background: ${agentInfo.color}; padding: 4px 8px; border-radius: 6px;">
                    ${agentInfo.icon}
                </span>
                <span>${agentInfo.label}</span>
            `;
        }

        document.body.appendChild(dragImage);

        // Удаляем после завершения drag
        setTimeout(() => {
            if (dragImage.parentNode) {
                dragImage.parentNode.removeChild(dragImage);
            }
        }, 100);

        return dragImage;
    }
}

/**
 * Компонент группы шаблонов нод
 */
export class NodeTemplateGroupComponent extends BaseComponent<HTMLDivElement> {
    private titleElement: HTMLDivElement | null = null;
    private templates: NodeTemplateComponent[] = [];

    /**
     * Конструктор группы шаблонов
     * @param title Заголовок группы
     */
    constructor(title: string) {
        super('div', 'crewai-node-template-group');

        this.createTitle(title);
    }

    /**
     * Добавляет шаблон в группу
     * @param template Компонент шаблона
     * @returns Эта группа шаблонов
     */
    public addTemplate(template: NodeTemplateComponent): this {
        this.templates.push(template);
        template.render(this.element);
        return this;
    }

    /**
     * Создает и добавляет шаблон
     * @param type Тип агента
     * @param category Категория агента
     * @returns Созданный шаблон
     */
    public createTemplate(type: AgentType, category: AgentCategory): NodeTemplateComponent {
        const template = new NodeTemplateComponent({ type, category });
        this.addTemplate(template);
        return template;
    }

    /**
     * Получает все шаблоны
     * @returns Массив шаблонов
     */
    public getTemplates(): NodeTemplateComponent[] {
        return [...this.templates];
    }

    /**
     * Очищает все шаблоны
     */
    public clear(): void {
        this.templates.forEach((template) => template.destroy());
        this.templates = [];
    }

    /**
     * Уничтожает группу шаблонов
     */
    protected onDestroy(): void {
        this.clear();
        this.titleElement = null;
    }

    /**
     * Создает заголовок группы
     * @param title Заголовок
     */
    private createTitle(title: string): void {
        this.titleElement = document.createElement('div');
        this.titleElement.className = 'crewai-sidebar__section-title';
        this.titleElement.textContent = title;
        this.element.appendChild(this.titleElement);
    }
}
