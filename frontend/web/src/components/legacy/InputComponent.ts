/**
 * CrewAI - Input Component
 * Компонент текстового поля ввода
 */

import { BaseComponent } from './BaseComponent';

export type InputType = 'text' | 'number' | 'password' | 'email' | 'url';
export type InputSize = 'sm' | 'md' | 'lg';

export interface InputOptions {
    type?: InputType;
    placeholder?: string;
    value?: string;
    disabled?: boolean;
    readonly?: boolean;
    required?: boolean;
    size?: InputSize;
    minLength?: number;
    maxLength?: number;
    min?: number;
    max?: number;
    step?: number;
    pattern?: string;
    autocomplete?: string;
}

/**
 * Компонент текстового поля ввода
 */
export class InputComponent extends BaseComponent<HTMLInputElement> {
    private onChangeCallbacks: Array<(value: string) => void> = [];
    private onInputCallbacks: Array<(value: string) => void> = [];

    /**
     * Конструктор input
     * @param options Опции input
     */
    constructor(options: InputOptions = {}) {
        super('input', 'crewai-input');

        this.element.type = options.type || 'text';
        this.setSize(options.size || 'md');

        if (options.placeholder) {
            this.setPlaceholder(options.placeholder);
        }

        if (options.value !== undefined) {
            this.setValue(options.value);
        }

        if (options.disabled) {
            this.setDisabled(true);
        }

        if (options.readonly) {
            this.setReadonly(true);
        }

        if (options.required) {
            this.setRequired(true);
        }

        if (options.minLength !== undefined) {
            this.setMinLength(options.minLength);
        }

        if (options.maxLength !== undefined) {
            this.setMaxLength(options.maxLength);
        }

        if (options.min !== undefined && options.type === 'number') {
            this.setMin(options.min);
        }

        if (options.max !== undefined && options.type === 'number') {
            this.setMax(options.max);
        }

        if (options.step !== undefined && options.type === 'number') {
            this.setStep(options.step);
        }

        if (options.pattern) {
            this.setPattern(options.pattern);
        }

        if (options.autocomplete) {
            this.setAutocomplete(options.autocomplete);
        }

        this.setupEventListeners();
    }

    /**
     * Устанавливает размер input
     * @param size Размер
     */
    public setSize(size: InputSize): void {
        this.removeClass('crewai-input--sm');
        this.removeClass('crewai-input--md');
        this.removeClass('crewai-input--lg');
        this.addClass(`crewai-input--${size}`);
    }

    /**
     * Устанавливает placeholder
     * @param placeholder Текст placeholder
     */
    public setPlaceholder(placeholder: string): void {
        this.element.placeholder = placeholder;
    }

    /**
     * Устанавливает значение
     * @param value Значение
     */
    public setValue(value: string): void {
        this.element.value = value;
    }

    /**
     * Получает значение
     * @returns Текущее значение
     */
    public getValue(): string {
        return this.element.value;
    }

    /**
     * Устанавливает readonly состояние
     * @param readonly Состояние
     */
    public setReadonly(readonly: boolean): void {
        this.element.readOnly = readonly;
        if (readonly) {
            this.addClass('crewai-input--readonly');
        } else {
            this.removeClass('crewai-input--readonly');
        }
    }

    /**
     * Устанавливает required состояние
     * @param required Состояние
     */
    public setRequired(required: boolean): void {
        this.element.required = required;
    }

    /**
     * Устанавливает минимальную длину
     * @param length Минимальная длина
     */
    public setMinLength(length: number): void {
        this.element.minLength = length;
    }

    /**
     * Устанавливает максимальную длину
     * @param length Максимальная длина
     */
    public setMaxLength(length: number): void {
        this.element.maxLength = length;
    }

    /**
     * Устанавливает минимальное значение (для number)
     * @param value Минимальное значение
     */
    public setMin(value: number): void {
        this.element.min = value.toString();
    }

    /**
     * Устанавливает максимальное значение (для number)
     * @param value Максимальное значение
     */
    public setMax(value: number): void {
        this.element.max = value.toString();
    }

    /**
     * Устанавливает шаг (для number)
     * @param value Шаг
     */
    public setStep(value: number): void {
        this.element.step = value.toString();
    }

    /**
     * Устанавливает паттерн валидации
     * @param pattern RegExp паттерн
     */
    public setPattern(pattern: string): void {
        this.element.pattern = pattern;
    }

    /**
     * Устанавливает autocomplete
     * @param value Значение autocomplete
     */
    public setAutocomplete(value: string): void {
        this.element.autocomplete = value;
    }

    /**
     * Устанавливает ошибку валидации
     * @param message Сообщение об ошибке
     */
    public setValidationError(message: string): void {
        this.element.setCustomValidity(message);
        this.addClass('crewai-input--error');
    }

    /**
     * Очищает ошибку валидации
     */
    public clearValidationError(): void {
        this.element.setCustomValidity('');
        this.removeClass('crewai-input--error');
    }

    /**
     * Проверяет валидность
     * @returns True если поле валидно
     */
    public isValid(): boolean {
        return this.element.checkValidity();
    }

    /**
     * Фокусирует input
     */
    public focus(): void {
        this.element.focus();
    }

    /**
     * Выделяет весь текст
     */
    public select(): void {
        this.element.select();
    }

    /**
     * Подписывается на изменение значения
     * @param callback Функция обратного вызова
     */
    public onChange(callback: (value: string) => void): void {
        this.onChangeCallbacks.push(callback);
    }

    /**
     * Подписывается на ввод значения
     * @param callback Функция обратного вызова
     */
    public onInput(callback: (value: string) => void): void {
        this.onInputCallbacks.push(callback);
    }

    /**
     * Отписывается от всех событий
     */
    public off(): void {
        this.onChangeCallbacks = [];
        this.onInputCallbacks = [];
    }

    /**
     * Уничтожает компонент
     */
    protected onDestroy(): void {
        this.off();
    }

    /**
     * Настраивает обработчики событий
     */
    private setupEventListeners(): void {
        this.element.addEventListener('change', this.handleChange);
        this.element.addEventListener('input', this.handleInput);
    }

    /**
     * Обработчик изменения
     */
    private handleChange = (): void => {
        const value = this.element.value;
        this.onChangeCallbacks.forEach((callback) => callback(value));
    };

    /**
     * Обработчик ввода
     */
    private handleInput = (): void => {
        const value = this.element.value;
        this.onInputCallbacks.forEach((callback) => callback(value));
    };
}

/**
 * Компонент текстовой области
 */
export class TextareaComponent extends BaseComponent<HTMLTextAreaElement> {
    private onChangeCallbacks: Array<(value: string) => void> = [];
    private onInputCallbacks: Array<(value: string) => void> = [];

    /**
     * Конструктор textarea
     * @param options Опции textarea
     */
    constructor(options: Omit<InputOptions, 'type' | 'min' | 'max' | 'step'> = {}) {
        super('textarea', 'crewai-textarea');

        if (options.placeholder) {
            this.setPlaceholder(options.placeholder);
        }

        if (options.value !== undefined) {
            this.setValue(options.value);
        }

        if (options.disabled) {
            this.setDisabled(true);
        }

        if (options.readonly) {
            this.setReadonly(true);
        }

        if (options.required) {
            this.setRequired(true);
        }

        if (options.minLength !== undefined) {
            this.setMinLength(options.minLength);
        }

        if (options.maxLength !== undefined) {
            this.setMaxLength(options.maxLength);
        }

        this.setupEventListeners();
    }

    /**
     * Устанавливает placeholder
     * @param placeholder Текст placeholder
     */
    public setPlaceholder(placeholder: string): void {
        this.element.placeholder = placeholder;
    }

    /**
     * Устанавливает значение
     * @param value Значение
     */
    public setValue(value: string): void {
        this.element.value = value;
    }

    /**
     * Получает значение
     * @returns Текущее значение
     */
    public getValue(): string {
        return this.element.value;
    }

    /**
     * Устанавливает количество строк
     * @param rows Количество строк
     */
    public setRows(rows: number): void {
        this.element.rows = rows;
    }

    /**
     * Устанавливает readonly состояние
     * @param readonly Состояние
     */
    public setReadonly(readonly: boolean): void {
        this.element.readOnly = readonly;
    }

    /**
     * Устанавливает required состояние
     * @param required Состояние
     */
    public setRequired(required: boolean): void {
        this.element.required = required;
    }

    /**
     * Устанавливает минимальную длину
     * @param length Минимальная длина
     */
    public setMinLength(length: number): void {
        this.element.minLength = length;
    }

    /**
     * Устанавливает максимальную длину
     * @param length Максимальная длина
     */
    public setMaxLength(length: number): void {
        this.element.maxLength = length;
    }

    /**
     * Подписывается на изменение значения
     * @param callback Функция обратного вызова
     */
    public onChange(callback: (value: string) => void): void {
        this.onChangeCallbacks.push(callback);
    }

    /**
     * Подписывается на ввод значения
     * @param callback Функция обратного вызова
     */
    public onInput(callback: (value: string) => void): void {
        this.onInputCallbacks.push(callback);
    }

    /**
     * Уничтожает компонент
     */
    protected onDestroy(): void {
        this.onChangeCallbacks = [];
        this.onInputCallbacks = [];
    }

    /**
     * Настраивает обработчики событий
     */
    private setupEventListeners(): void {
        this.element.addEventListener('change', this.handleChange);
        this.element.addEventListener('input', this.handleInput);
    }

    /**
     * Обработчик изменения
     */
    private handleChange = (): void => {
        const value = this.element.value;
        this.onChangeCallbacks.forEach((callback) => callback(value));
    };

    /**
     * Обработчик ввода
     */
    private handleInput = (): void => {
        const value = this.element.value;
        this.onInputCallbacks.forEach((callback) => callback(value));
    };
}
