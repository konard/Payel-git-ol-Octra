/**
 * CrewAI - Button Component
 * Компонент кнопки с различными вариантами стилей
 */

import { BaseComponent } from './BaseComponent';

export type ButtonVariant = 'primary' | 'secondary' | 'ghost' | 'danger';
export type ButtonSize = 'sm' | 'md' | 'lg';

export interface ButtonOptions {
    text?: string;
    icon?: string;
    variant?: ButtonVariant;
    size?: ButtonSize;
    disabled?: boolean;
    title?: string;
}

/**
 * Компонент кнопки
 */
export class ButtonComponent extends BaseComponent<HTMLButtonElement> {
    private iconElement: HTMLSpanElement | null = null;
    private textElement: HTMLSpanElement | null = null;
    private onClickCallback?: () => void;

    /**
     * Конструктор кнопки
     * @param options Опции кнопки
     */
    constructor(options: ButtonOptions = {}) {
        super('button', 'crewai-button');

        this.setVariant(options.variant || 'secondary');
        this.setSize(options.size || 'md');

        if (options.icon) {
            this.setIcon(options.icon);
        }

        if (options.text) {
            this.setText(options.text);
        }

        if (options.title) {
            this.setTitle(options.title);
        }

        if (options.disabled) {
            this.setDisabled(true);
        }
    }

    /**
     * Устанавливает вариант кнопки
     * @param variant Вариант кнопки
     */
    public setVariant(variant: ButtonVariant): void {
        this.removeClass('crewai-button--primary');
        this.removeClass('crewai-button--secondary');
        this.removeClass('crewai-button--ghost');
        this.removeClass('crewai-button--danger');
        this.addClass(`crewai-button--${variant}`);
    }

    /**
     * Устанавливает размер кнопки
     * @param size Размер кнопки
     */
    public setSize(size: ButtonSize): void {
        this.removeClass('crewai-button--sm');
        this.removeClass('crewai-button--md');
        this.removeClass('crewai-button--lg');
        this.addClass(`crewai-button--${size}`);
    }

    /**
     * Устанавливает иконку кнопки
     * @param icon Текст иконки (emoji или символ)
     */
    public setIcon(icon: string): void {
        if (!this.iconElement) {
            this.iconElement = document.createElement('span');
            this.iconElement.className = 'crewai-button__icon';
            this.element.prepend(this.iconElement);
        }
        this.iconElement.textContent = icon;
        this.element.classList.add('crewai-button--icon');
    }

    /**
     * Устанавливает текст кнопки
     * @param text Текст кнопки
     */
    public setText(text: string): void {
        if (!this.textElement) {
            this.textElement = document.createElement('span');
            this.textElement.className = 'crewai-button__text';
            this.element.appendChild(this.textElement);
        }
        this.textElement.textContent = text;
    }

    /**
     * Устанавливает заголовок (tooltip)
     * @param title Заголовок
     */
    public setTitle(title: string): void {
        this.element.title = title;
    }

    /**
     * Устанавливает обработчик клика
     * @param callback Функция обратного вызова
     */
    public onClick(callback: () => void): void {
        this.onClickCallback = callback;
        this.element.addEventListener('click', this.handleClick);
    }

    /**
     * Удаляет обработчик клика
     */
    public offClick(): void {
        this.element.removeEventListener('click', this.handleClick);
        this.onClickCallback = undefined;
    }

    /**
     * Симулирует клик
     */
    public click(): void {
        this.element.click();
    }

    /**
     * Уничтожает кнопку
     */
    protected onDestroy(): void {
        this.offClick();
        this.iconElement = null;
        this.textElement = null;
    }

    /**
     * Обработчик клика
     */
    private handleClick = (): void => {
        if (this.onClickCallback && !this.element.disabled) {
            this.onClickCallback();
        }
    };
}

/**
 * Создает группу кнопок
 */
export class ButtonGroupComponent extends BaseComponent<HTMLDivElement> {
    private buttons: ButtonComponent[] = [];

    /**
     * Конструктор группы кнопок
     */
    constructor() {
        super('div', 'crewai-button-group');
    }

    /**
     * Добавляет кнопку в группу
     * @param button Компонент кнопки
     * @returns Эта группа кнопок
     */
    public addButton(button: ButtonComponent): this {
        this.buttons.push(button);
        button.render(this.element);
        return this;
    }

    /**
     * Создает и добавляет кнопку
     * @param options Опции кнопки
     * @returns Созданная кнопка
     */
    public createButton(options: ButtonOptions): ButtonComponent {
        const button = new ButtonComponent(options);
        this.addButton(button);
        return button;
    }

    /**
     * Удаляет кнопку из группы
     * @param button Кнопка для удаления
     */
    public removeButton(button: ButtonComponent): void {
        const index = this.buttons.indexOf(button);
        if (index !== -1) {
            this.buttons.splice(index, 1);
            button.remove();
        }
    }

    /**
     * Очищает все кнопки
     */
    public clear(): void {
        this.buttons.forEach((button) => button.destroy());
        this.buttons = [];
    }

    /**
     * Получает все кнопки
     * @returns Массив кнопок
     */
    public getButtons(): ButtonComponent[] {
        return [...this.buttons];
    }

    /**
     * Уничтожает группу кнопок
     */
    protected onDestroy(): void {
        this.clear();
    }
}
