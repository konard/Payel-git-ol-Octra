/**
 * CrewAI - Properties Panel Component
 * Панель свойств для редактирования параметров ноды
 */

import { BaseComponent } from './BaseComponent';
import { ButtonComponent } from './ButtonComponent';
import { InputComponent, TextareaComponent } from './InputComponent';
import { LabelComponent } from './LabelComponent';
import { SelectComponent } from './SelectComponent';
import { BadgeComponent } from './BadgeComponent';
import type { NodeData, AgentParameter } from '../models/Types';
import { getAgentInfo, getAgentParameters } from '../models/AgentDefinitions';

export interface PropertiesPanelOptions {
    width?: number;
    closable?: boolean;
}

/**
 * Компонент панели свойств
 */
export class PropertiesPanelComponent extends BaseComponent<HTMLDivElement> {
    private headerElement: HTMLDivElement | null = null;
    private titleElement: HTMLHeadingElement | null = null;
    private closeButton: ButtonComponent | null = null;
    private contentElement: HTMLDivElement | null = null;
    private emptyElement: HTMLDivElement | null = null;
    private formElement: HTMLDivElement | null = null;

    private currentNode: NodeData | null = null;
    private inputComponents: Map<string, InputComponent | TextareaComponent | SelectComponent> =
        new Map();
    private onChangeCallbacks: Array<(nodeId: string, parameters: Record<string, unknown>) => void> =
        [];
    private onCloseCallbacks: Array<() => void> = [];

    /**
     * Конструктор панели свойств
     * @param options Опции панели
     */
    constructor(options: PropertiesPanelOptions = {}) {
        super('div', 'crewai-properties');

        this.element.style.width = options.width ? `${options.width}px` : '320px';

        this.createStructure(options);
        this.showEmptyState();
    }

    /**
     * Показывает свойства ноды
     * @param node Данные ноды
     */
    public showNodeProperties(node: NodeData): void {
        this.currentNode = { ...node };
        this.hideEmptyState();
        this.renderNodeProperties(node);
    }

    /**
     * Скрывает свойства ноды
     */
    public hideNodeProperties(): void {
        this.currentNode = null;
        this.showEmptyState();
    }

    /**
     * Обновляет свойства ноды
     * @param node Обновленные данные ноды
     */
    public updateNodeProperties(node: NodeData): void {
        if (this.currentNode?.id === node.id) {
            this.currentNode = { ...node };
            this.updateFormValues(node.parameters);
        }
    }

    /**
     * Подписывается на изменение свойств
     * @param callback Функция обратного вызова
     */
    public onChange(
        callback: (nodeId: string, parameters: Record<string, unknown>) => void
    ): void {
        this.onChangeCallbacks.push(callback);
    }

    /**
     * Подписывается на закрытие панели
     * @param callback Функция обратного вызова
     */
    public onClose(callback: () => void): void {
        this.onCloseCallbacks.push(callback);
    }

    /**
     * Уничтожает компонент
     */
    protected onDestroy(): void {
        this.onChangeCallbacks = [];
        this.onCloseCallbacks = [];
        this.inputComponents.forEach((comp) => comp.destroy());
        this.inputComponents.clear();
        this.closeButton?.destroy();
        this.currentNode = null;
        this.headerElement = null;
        this.titleElement = null;
        this.contentElement = null;
        this.emptyElement = null;
        this.formElement = null;
    }

    /**
     * Создает структуру панели свойств
     * @param options Опции панели
     */
    private createStructure(options: PropertiesPanelOptions): void {
        // Хедер
        this.headerElement = document.createElement('div');
        this.headerElement.className = 'crewai-properties__header';
        this.element.appendChild(this.headerElement);

        this.titleElement = document.createElement('h3');
        this.titleElement.className = 'crewai-properties__title';
        this.titleElement.textContent = 'Properties';
        this.headerElement.appendChild(this.titleElement);

        if (options.closable !== false) {
            this.closeButton = new ButtonComponent({
                icon: '✕',
                variant: 'ghost',
                size: 'sm',
            });
            this.closeButton.onClick(() => {
                this.onCloseCallbacks.forEach((cb) => cb());
            });
            this.closeButton.render(this.headerElement);
        }

        // Контент
        this.contentElement = document.createElement('div');
        this.contentElement.className = 'crewai-properties__content';
        this.element.appendChild(this.contentElement);

        // Пустое состояние
        this.emptyElement = document.createElement('div');
        this.emptyElement.className = 'crewai-properties__empty';
        this.emptyElement.innerHTML = `
            <div class="crewai-properties__empty-icon">Edit</div>
            <div class="crewai-properties__empty-text">Select a node to view its properties</div>
        `;
        this.contentElement.appendChild(this.emptyElement);

        // Форма
        this.formElement = document.createElement('div');
        this.formElement.className = 'crewai-properties__form';
        this.contentElement.appendChild(this.formElement);
    }

    /**
     * Показывает пустое состояние
     */
    private showEmptyState(): void {
        if (this.emptyElement) {
            this.emptyElement.style.display = 'flex';
        }
        if (this.formElement) {
            this.formElement.style.display = 'none';
        }
    }

    /**
     * Скрывает пустое состояние
     */
    private hideEmptyState(): void {
        if (this.emptyElement) {
            this.emptyElement.style.display = 'none';
        }
        if (this.formElement) {
            this.formElement.style.display = 'block';
        }
    }

    /**
     * Рендерит свойства ноды
     * @param node Данные ноды
     */
    private renderNodeProperties(node: NodeData): void {
        if (!this.formElement || !this.titleElement) {
            return;
        }

        // Очищаем форму
        this.formElement.innerHTML = '';
        this.inputComponents.forEach((comp) => comp.destroy());
        this.inputComponents.clear();

        // Получаем информацию об агенте
        const agentInfo = getAgentInfo(node.type);
        const parameters = getAgentParameters(node.type);

        // Обновляем заголовок
        this.titleElement.textContent = agentInfo?.label || node.name;

        // Бейдж категории
        const categoryBadge = new BadgeComponent({
            text: node.category,
            variant: 'primary',
            size: 'sm',
        });
        categoryBadge.render(this.formElement, true);

        // Описание
        if (agentInfo?.description) {
            const descriptionElement = document.createElement('p');
            descriptionElement.className = 'crewai-properties__description';
            descriptionElement.style.margin = '12px 0';
            descriptionElement.style.fontSize = '12px';
            descriptionElement.style.color = 'var(--color-text-secondary)';
            descriptionElement.style.lineHeight = '1.5';
            descriptionElement.textContent = agentInfo.description;
            this.formElement.appendChild(descriptionElement);
        }

        // Разделитель
        this.createDivider();

        // Поля формы
        // Имя ноды
        this.createNameField(node);

        // Параметры агента
        parameters.forEach((param) => {
            this.createParameterField(param, node.parameters);
        });

        // Разделитель
        this.createDivider();

        // Статус
        this.createStatusField(node);

        // ID ноды
        this.createIdField(node);
    }

    /**
     * Создает поле имени
     * @param node Данные ноды
     */
    private createNameField(node: NodeData): void {
        const label = new LabelComponent({ text: 'Name', required: true });
        label.render(this.formElement!);

        const input = new InputComponent({
            type: 'text',
            value: node.name,
            placeholder: 'Node name',
        });
        input.onChange((value) => {
            if (this.currentNode) {
                this.currentNode.name = value;
                this.emitChange();
            }
        });
        input.render(this.formElement!);
        this.inputComponents.set('name', input);
    }

    /**
     * Создает поле параметра
     * @param param Определение параметра
     * @param parameters Текущие параметры
     */
    private createParameterField(
        param: AgentParameter,
        parameters: Record<string, unknown>
    ): void {
        if (!this.formElement) {
            return;
        }

        const value = parameters[param.name];
        const label = new LabelComponent({
            text: this.capitalizeFirst(param.name),
            required: param.required,
            hint: param.description,
        });
        label.render(this.formElement);

        let input: InputComponent | TextareaComponent | SelectComponent;

        if (param.type === 'boolean') {
            // TODO: Добавить ToggleComponent
            input = new InputComponent({
                type: 'text',
                value: value ? 'true' : 'false',
            });
        } else if (param.type === 'number') {
            input = new InputComponent({
                type: 'number',
                value: String(value ?? param.defaultValue ?? ''),
                min: 0,
            });
        } else if (param.options && param.options.length > 0) {
            // Select с опциями
            const select = new SelectComponent({
                options: param.options.map((opt) => ({ value: opt, label: opt })),
            });
            select.setValue(String(value ?? param.defaultValue ?? param.options[0]));
            select.onChange((selectedValue) => {
                if (this.currentNode) {
                    this.currentNode.parameters[param.name] = selectedValue;
                    this.emitChange();
                }
            });
            select.render(this.formElement);
            this.inputComponents.set(param.name, select);
            return;
        } else if (param.type === 'string' && String(value).length > 50) {
            input = new TextareaComponent({
                value: String(value ?? ''),
                rows: 4,
            });
        } else {
            input = new InputComponent({
                type: 'text',
                value: String(value ?? param.defaultValue ?? ''),
                placeholder: param.name,
            });
        }

        input.onChange((newValue) => {
            if (this.currentNode) {
                let parsedValue: unknown = newValue;

                if (param.type === 'number') {
                    parsedValue = parseFloat(newValue) || 0;
                } else if (param.type === 'boolean') {
                    parsedValue = newValue.toLowerCase() === 'true';
                }

                this.currentNode.parameters[param.name] = parsedValue;
                this.emitChange();
            }
        });

        input.render(this.formElement);
        this.inputComponents.set(param.name, input);
    }

    /**
     * Создает поле статуса
     * @param node Данные ноды
     */
    private createStatusField(node: NodeData): void {
        if (!this.formElement) {
            return;
        }

        const label = new LabelComponent({ text: 'Status' });
        label.render(this.formElement);

        const statusLabels: Record<string, string> = {
            idle: 'Idle',
            active: 'Active',
            error: 'Error',
            completed: 'Completed',
        };

        const statusVariants: Record<string, 'primary' | 'success' | 'warning' | 'error'> = {
            idle: 'primary',
            active: 'success',
            error: 'error',
            completed: 'success',
        };

        const statusBadge = new BadgeComponent({
            text: statusLabels[node.status] || node.status,
            variant: statusVariants[node.status] || 'primary',
            size: 'md',
        });
        statusBadge.render(this.formElement);
    }

    /**
     * Создает поле ID
     * @param node Данные ноды
     */
    private createIdField(node: NodeData): void {
        if (!this.formElement) {
            return;
        }

        const label = new LabelComponent({ text: 'Node ID' });
        label.render(this.formElement);

        const idInput = new InputComponent({
            type: 'text',
            value: node.id,
            readonly: true,
            size: 'sm',
        });
        idInput.render(this.formElement);
    }

    /**
     * Создает разделитель
     */
    private createDivider(): void {
        if (!this.formElement) {
            return;
        }

        const divider = document.createElement('div');
        divider.className = 'crewai-divider';
        divider.style.margin = '16px 0';
        this.formElement.appendChild(divider);
    }

    /**
     * Обновляет значения формы
     * @param parameters Новые параметры
     */
    private updateFormValues(parameters: Record<string, unknown>): void {
        this.inputComponents.forEach((component, key) => {
            const value = parameters[key];
            if (value !== undefined && component instanceof InputComponent) {
                component.setValue(String(value));
            }
        });
    }

    /**
     * Испускает событие изменения
     */
    private emitChange(): void {
        if (this.currentNode) {
            this.onChangeCallbacks.forEach((cb) =>
                cb(this.currentNode!.id, this.currentNode!.parameters)
            );
        }
    }

    /**
     * Капитализирует первую букву
     * @param str Строка
     * @returns Строка с заглавной первой буквой
     */
    private capitalizeFirst(str: string): string {
        return str.charAt(0).toUpperCase() + str.slice(1);
    }
}
