/**
 * CrewAI - Label Component
 * Компонент лейбла для форм
 */

import { BaseComponent } from './BaseComponent';

export interface LabelOptions {
    text?: string;
    required?: boolean;
    hint?: string;
    error?: string;
    for?: string;
}

/**
 * Компонент лейбла
 */
export class LabelComponent extends BaseComponent<HTMLLabelElement> {
    private textElement: Text | null = null;
    private hintElement: HTMLSpanElement | null = null;
    private errorElement: HTMLSpanElement | null = null;
    private requiredElement: HTMLSpanElement | null = null;

    /**
     * Конструктор лейбла
     * @param options Опции лейбла
     */
    constructor(options: LabelOptions = {}) {
        super('label', 'crewai-form-label');

        if (options.text) {
            this.setText(options.text);
        }

        if (options.required) {
            this.setRequired(true);
        }

        if (options.hint) {
            this.setHint(options.hint);
        }

        if (options.error) {
            this.setError(options.error);
        }

        if (options.for) {
            this.setFor(options.for);
        }
    }

    /**
     * Устанавливает текст лейбла
     * @param text Текст лейбла
     */
    public setText(text: string): void {
        if (!this.textElement) {
            this.textElement = document.createTextNode('');
            this.element.appendChild(this.textElement);
        }
        this.textElement.textContent = text;
    }

    /**
     * Получает текст лейбла
     * @returns Текст лейбла
     */
    public getText(): string {
        return this.textElement?.textContent || '';
    }

    /**
     * Устанавливает required индикатор
     * @param required Состояние required
     */
    public setRequired(required: boolean): void {
        if (required) {
            if (!this.requiredElement) {
                this.requiredElement = document.createElement('span');
                this.requiredElement.className = 'crewai-form-label__required';
                this.requiredElement.textContent = ' *';
                this.element.appendChild(this.requiredElement);
            }
            this.requiredElement.style.display = '';
        } else if (this.requiredElement) {
            this.requiredElement.style.display = 'none';
        }
    }

    /**
     * Устанавливает подсказку
     * @param hint Текст подсказки
     */
    public setHint(hint: string): void {
        if (!this.hintElement) {
            this.hintElement = document.createElement('span');
            this.hintElement.className = 'crewai-form-label__hint';
            this.element.appendChild(this.hintElement);
        }
        this.hintElement.textContent = hint;
        this.hintElement.style.display = '';
    }

    /**
     * Скрывает подсказку
     */
    public hideHint(): void {
        if (this.hintElement) {
            this.hintElement.style.display = 'none';
        }
    }

    /**
     * Показывает подсказку
     */
    public showHint(): void {
        if (this.hintElement) {
            this.hintElement.style.display = '';
        }
    }

    /**
     * Устанавливает сообщение об ошибке
     * @param message Сообщение об ошибке
     */
    public setError(message: string): void {
        if (!this.errorElement) {
            this.errorElement = document.createElement('span');
            this.errorElement.className = 'crewai-form-label__error';
            this.element.appendChild(this.errorElement);
        }
        this.errorElement.textContent = message;
        this.errorElement.style.display = '';
        this.addClass('crewai-form-label--error');
    }

    /**
     * Очищает ошибку
     */
    public clearError(): void {
        if (this.errorElement) {
            this.errorElement.style.display = 'none';
        }
        this.removeClass('crewai-form-label--error');
    }

    /**
     * Устанавливает связь с input (for атрибут)
     * @param id ID связанного элемента
     */
    public setFor(id: string): void {
        this.element.htmlFor = id;
    }

    /**
     * Получает связанный ID
     * @returns ID связанного элемента
     */
    public getFor(): string {
        return this.element.htmlFor;
    }

    /**
     * Уничтожает компонент
     */
    protected onDestroy(): void {
        this.textElement = null;
        this.hintElement = null;
        this.errorElement = null;
        this.requiredElement = null;
    }
}

/**
 * Компонент группы формы
 */
export class FormGroupComponent extends BaseComponent<HTMLDivElement> {
    private labelComponent: LabelComponent | null = null;
    private contentElement: HTMLDivElement | null = null;

    /**
     * Конструктор группы формы
     */
    constructor() {
        super('div', 'crewai-form-group');
    }

    /**
     * Устанавливает лейбл
     * @param label Компонент лейбла
     */
    public setLabel(label: LabelComponent): void {
        if (this.labelComponent) {
            this.labelComponent.remove();
        }
        this.labelComponent = label;
        label.render(this.element, true);
    }

    /**
     * Создает и устанавливает лейбл
     * @param text Текст лейбла
     * @param required Required состояние
     * @returns Созданный лейбл
     */
    public createLabel(text: string, required = false): LabelComponent {
        const label = new LabelComponent({ text, required });
        this.setLabel(label);
        return label;
    }

    /**
     * Устанавливает контент группы
     * @param content Элемент контента
     */
    public setContent(content: HTMLElement): void {
        if (!this.contentElement) {
            this.contentElement = document.createElement('div');
            this.contentElement.className = 'crewai-form-group__content';
            this.element.appendChild(this.contentElement);
        }

        this.contentElement.innerHTML = '';
        this.contentElement.appendChild(content);
    }

    /**
     * Получает контент элемент
     * @returns Контент элемент
     */
    public getContentElement(): HTMLDivElement | null {
        return this.contentElement;
    }

    /**
     * Уничтожает компонент
     */
    protected onDestroy(): void {
        if (this.labelComponent) {
            this.labelComponent.destroy();
            this.labelComponent = null;
        }
        this.contentElement = null;
    }
}
