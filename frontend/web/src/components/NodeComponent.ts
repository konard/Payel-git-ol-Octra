/**
 * CrewAI - Agent Node Component
 * Компонент ноды агента для канваса
 */

import { BaseComponent } from './BaseComponent';
import type { NodeData, Point } from '../models/Types';
import { getAgentInfo } from '../models/AgentDefinitions';

export interface NodeComponentOptions {
    nodeData: NodeData;
    draggable?: boolean;
    selectable?: boolean;
    deletable?: boolean;
}

/**
 * Компонент ноды агента
 */
export class NodeComponent extends BaseComponent<HTMLDivElement> {
    private nodeData: NodeData;
    private headerElement: HTMLDivElement | null = null;
    private iconElement: HTMLDivElement | null = null;
    private nameElement: HTMLSpanElement | null = null;
    private typeElement: HTMLSpanElement | null = null;
    private statusElement: HTMLSpanElement | null = null;
    private bodyElement: HTMLDivElement | null = null;
    private inputPortElement: HTMLDivElement | null = null;
    private outputPortElement: HTMLDivElement | null = null;
    private actionsElement: HTMLDivElement | null = null;

    private isDragging = false;
    private dragOffset: Point = { x: 0, y: 0 };
    private onMoveCallbacks: Array<(position: Point) => void> = [];
    private onSelectCallbacks: Array<() => void> = [];
    private onDeleteCallbacks: Array<() => void> = [];
    private onPortClickCallbacks: Array<(portType: 'input' | 'output') => void> = [];

    /**
     * Конструктор ноды
     * @param options Опции ноды
     */
    constructor(options: NodeComponentOptions) {
        super('div', 'crewai-node');

        this.nodeData = { ...options.nodeData };

        this.setData('node-id', this.nodeData.id);
        this.setData('node-type', this.nodeData.type);
        this.setData('node-category', this.nodeData.category);

        this.addClass(`crewai-node--${this.nodeData.category}`);

        this.createStructure();
        this.updateAppearance();
        this.setupEventListeners(options);
    }

    /**
     * Получает данные ноды
     * @returns Данные ноды
     */
    public getNodeData(): NodeData {
        return { ...this.nodeData };
    }

    /**
     * Получает ID ноды
     * @returns ID ноды
     */
    public getNodeId(): string {
        return this.nodeData.id;
    }

    /**
     * Получает тип ноды
     * @returns Тип ноды
     */
    public getNodeType(): string {
        return this.nodeData.type;
    }

    /**
     * Обновляет позицию ноды
     * @param position Новая позиция
     */
    public setPosition(position: Point): void {
        this.nodeData.position = { ...position };
        this.element.style.left = `${position.x}px`;
        this.element.style.top = `${position.y}px`;
    }

    /**
     * Получает текущую позицию
     * @returns Текущая позиция
     */
    public getPosition(): Point {
        return { ...this.nodeData.position };
    }

    /**
     * Обновляет имя ноды
     * @param name Новое имя
     */
    public setName(name: string): void {
        this.nodeData.name = name;
        if (this.nameElement) {
            this.nameElement.textContent = name;
        }
    }

    /**
     * Обновляет статус ноды
     * @param status Новый статус
     */
    public setStatus(status: NodeData['status']): void {
        this.nodeData.status = status;
        this.updateStatusAppearance();
    }

    /**
     * Выделяет ноду
     */
    public select(): void {
        this.nodeData.isSelected = true;
        this.addClass('crewai-node--selected');
    }

    /**
     * Снимает выделение с ноды
     */
    public deselect(): void {
        this.nodeData.isSelected = false;
        this.removeClass('crewai-node--selected');
    }

    /**
     * Проверяет, выделена ли нода
     * @returns True если нода выделена
     */
    public isSelected(): boolean {
        return this.nodeData.isSelected === true;
    }

    /**
     * Подписывается на перемещение ноды
     * @param callback Функция обратного вызова
     */
    public onMove(callback: (position: Point) => void): void {
        this.onMoveCallbacks.push(callback);
    }

    /**
     * Подписывается на выделение ноды
     * @param callback Функция обратного вызова
     */
    public onSelect(callback: () => void): void {
        this.onSelectCallbacks.push(callback);
    }

    /**
     * Подписывается на удаление ноды
     * @param callback Функция обратного вызова
     */
    public onDelete(callback: () => void): void {
        this.onDeleteCallbacks.push(callback);
    }

    /**
     * Подписывается на клик по порту
     * @param callback Функция обратного вызова
     */
    public onPortClick(callback: (portType: 'input' | 'output') => void): void {
        this.onPortClickCallbacks.push(callback);
    }

    /**
     * Обновляет параметры ноды
     * @param parameters Новые параметры
     */
    public updateParameters(parameters: Record<string, unknown>): void {
        this.nodeData.parameters = { ...this.nodeData.parameters, ...parameters };
        this.renderBody();
    }

    /**
     * Уничтожает компонент ноды
     */
    protected onDestroy(): void {
        this.onMoveCallbacks = [];
        this.onSelectCallbacks = [];
        this.onDeleteCallbacks = [];
        this.onPortClickCallbacks = [];
        this.headerElement = null;
        this.iconElement = null;
        this.nameElement = null;
        this.typeElement = null;
        this.statusElement = null;
        this.bodyElement = null;
        this.inputPortElement = null;
        this.outputPortElement = null;
        this.actionsElement = null;
    }

    /**
     * Создает структуру ноды
     */
    private createStructure(): void {
        this.createHeader();
        this.createBody();
        this.createPorts();
        this.createActions();
    }

    /**
     * Создает хедер ноды
     */
    private createHeader(): void {
        this.headerElement = document.createElement('div');
        this.headerElement.className = 'crewai-node__header';

        this.iconElement = document.createElement('div');
        this.iconElement.className = 'crewai-node__icon';
        this.headerElement.appendChild(this.iconElement);

        const infoElement = document.createElement('div');
        infoElement.className = 'crewai-node__info';

        this.nameElement = document.createElement('span');
        this.nameElement.className = 'crewai-node__name';
        infoElement.appendChild(this.nameElement);

        this.typeElement = document.createElement('span');
        this.typeElement.className = 'crewai-node__type';
        infoElement.appendChild(this.typeElement);

        this.headerElement.appendChild(infoElement);

        this.statusElement = document.createElement('span');
        this.statusElement.className = 'crewai-node__status';
        this.headerElement.appendChild(this.statusElement);

        this.element.appendChild(this.headerElement);
    }

    /**
     * Создает тело ноды
     */
    private createBody(): void {
        this.bodyElement = document.createElement('div');
        this.bodyElement.className = 'crewai-node__body';
        this.element.appendChild(this.bodyElement);
        this.renderBody();
    }

    /**
     * Рендерит содержимое тела ноды
     */
    private renderBody(): void {
        if (!this.bodyElement) {
            return;
        }

        this.bodyElement.innerHTML = '';

        const agentInfo = getAgentInfo(this.nodeData.type);
        if (agentInfo) {
            const descriptionElement = document.createElement('div');
            descriptionElement.className = 'crewai-node__description';
            descriptionElement.textContent = agentInfo.description;
            this.bodyElement.appendChild(descriptionElement);
        }

        const paramsElement = document.createElement('div');
        paramsElement.className = 'crewai-node__params';

        const paramKeys = Object.keys(this.nodeData.parameters).slice(0, 3);
        paramKeys.forEach((key) => {
            const value = this.nodeData.parameters[key];
            const paramElement = document.createElement('div');
            paramElement.className = 'crewai-node__param';

            const nameSpan = document.createElement('span');
            nameSpan.className = 'crewai-node__param-name';
            nameSpan.textContent = `${key}: `;

            const valueSpan = document.createElement('span');
            valueSpan.className = 'crewai-node__param-value';
            valueSpan.textContent = String(value);

            paramElement.appendChild(nameSpan);
            paramElement.appendChild(valueSpan);
            paramsElement.appendChild(paramElement);
        });

        this.bodyElement.appendChild(paramsElement);
    }

    /**
     * Создает порты для соединений
     */
    private createPorts(): void {
        const portsContainer = document.createElement('div');
        portsContainer.className = 'crewai-node__ports';

        // Input port
        this.inputPortElement = document.createElement('div');
        this.inputPortElement.className = 'crewai-node__port crewai-node__port--input crewai-node__port-top';
        this.inputPortElement.title = 'Input';
        this.inputPortElement.addEventListener('click', (e) => {
            e.stopPropagation();
            this.onPortClickCallbacks.forEach((cb) => cb('input'));
        });
        portsContainer.appendChild(this.inputPortElement);

        // Output port
        this.outputPortElement = document.createElement('div');
        this.outputPortElement.className = 'crewai-node__port crewai-node__port--output crewai-node__port-top';
        this.outputPortElement.title = 'Output';
        this.outputPortElement.addEventListener('click', (e) => {
            e.stopPropagation();
            this.onPortClickCallbacks.forEach((cb) => cb('output'));
        });
        portsContainer.appendChild(this.outputPortElement);

        this.element.appendChild(portsContainer);
    }

    /**
     * Создает кнопки действий
     */
    private createActions(): void {
        this.actionsElement = document.createElement('div');
        this.actionsElement.className = 'crewai-node__actions';

        const deleteButton = document.createElement('button');
        deleteButton.className = 'crewai-node__action crewai-node__action--danger';
        deleteButton.innerHTML = '✕';
        deleteButton.title = 'Delete node';
        deleteButton.addEventListener('click', (e) => {
            e.stopPropagation();
            this.onDeleteCallbacks.forEach((cb) => cb());
        });
        this.actionsElement.appendChild(deleteButton);

        this.element.appendChild(this.actionsElement);
    }

    /**
     * Обновляет внешний вид ноды
     */
    private updateAppearance(): void {
        const agentInfo = getAgentInfo(this.nodeData.type);

        if (agentInfo) {
            if (this.iconElement) {
                this.iconElement.textContent = agentInfo.icon;
                this.iconElement.style.background = agentInfo.color;
            }

            if (this.nameElement) {
                this.nameElement.textContent = this.nodeData.name;
            }

            if (this.typeElement) {
                this.typeElement.textContent = agentInfo.label;
            }
        }

        this.setPosition(this.nodeData.position);
        this.updateStatusAppearance();

        if (this.nodeData.isSelected) {
            this.select();
        }
    }

    /**
     * Обновляет внешний вид статуса
     */
    private updateStatusAppearance(): void {
        if (!this.statusElement) {
            return;
        }

        this.removeClass('crewai-node--active');
        this.removeClass('crewai-node--error');

        switch (this.nodeData.status) {
            case 'active':
                this.statusElement.style.background = 'var(--color-success)';
                this.statusElement.style.boxShadow = '0 0 8px var(--color-success)';
                this.addClass('crewai-node--active');
                break;
            case 'error':
                this.statusElement.style.background = 'var(--color-error)';
                this.statusElement.style.boxShadow = '0 0 8px var(--color-error)';
                this.addClass('crewai-node--error');
                break;
            case 'completed':
                this.statusElement.style.background = 'var(--color-success)';
                break;
            default:
                this.statusElement.style.background = 'var(--color-text-disabled)';
                this.statusElement.style.boxShadow = 'none';
        }
    }

    /**
     * Настраивает обработчики событий
     * @param options Опции ноды
     */
    private setupEventListeners(options: NodeComponentOptions): void {
        if (options.draggable !== false) {
            this.setupDragListeners();
        }

        if (options.selectable !== false) {
            this.setupSelectListeners();
        }
    }

    /**
     * Настраивает обработчики перетаскивания
     */
    private setupDragListeners(): void {
        this.headerElement?.addEventListener('mousedown', this.handleMouseDown);
    }

    /**
     * Настраивает обработчики выделения
     */
    private setupSelectListeners(): void {
        this.element.addEventListener('click', this.handleClick);
    }

    /**
     * Обработчик нажатия мыши
     */
    private handleMouseDown = (event: MouseEvent): void => {
        if (event.button !== 0) {
            return;
        }

        event.preventDefault();
        event.stopPropagation();

        this.isDragging = true;
        document.body.classList.add('crewai-dragging');

        const rect = this.element.getBoundingClientRect();
        this.dragOffset = {
            x: event.clientX - rect.left,
            y: event.clientY - rect.top,
        };

        const moveHandler = (e: MouseEvent): void => {
            if (!this.isDragging) {
                return;
            }

            const parent = this.element.parentElement;
            if (!parent) {
                return;
            }

            const parentRect = parent.getBoundingClientRect();
            const newX = e.clientX - parentRect.left - this.dragOffset.x;
            const newY = e.clientY - parentRect.top - this.dragOffset.y;

            const newPosition: Point = {
                x: Math.max(0, newX),
                y: Math.max(0, newY),
            };

            this.setPosition(newPosition);
            this.onMoveCallbacks.forEach((cb) => cb(newPosition));
        };

        const upHandler = (): void => {
            this.isDragging = false;
            document.body.classList.remove('crewai-dragging');
            document.removeEventListener('mousemove', moveHandler);
            document.removeEventListener('mouseup', upHandler);
        };

        document.addEventListener('mousemove', moveHandler);
        document.addEventListener('mouseup', upHandler);
    };

    /**
     * Обработчик клика
     */
    private handleClick = (event: MouseEvent): void => {
        event.stopPropagation();
        this.onSelectCallbacks.forEach((cb) => cb());
    };
}
