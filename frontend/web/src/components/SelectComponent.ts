/**
 * CrewAI - Select Component
 * Компонент выпадающего списка
 */

import { BaseComponent } from './BaseComponent';

export interface SelectOption {
    value: string;
    label: string;
    disabled?: boolean;
    selected?: boolean;
}

export interface SelectOptions {
    options?: SelectOption[];
    placeholder?: string;
    disabled?: boolean;
    multiple?: boolean;
    size?: 'sm' | 'md' | 'lg';
}

/**
 * Компонент селекта
 */
export class SelectComponent extends BaseComponent<HTMLSelectElement> {
    private onChangeCallbacks: Array<(value: string) => void> = [];
    private options: SelectOption[] = [];

    /**
     * Конструктор селекта
     * @param options Опции селекта
     */
    constructor(options: SelectOptions = {}) {
        super('select', 'crewai-select');

        if (options.multiple) {
            this.element.multiple = true;
        }

        if (options.disabled) {
            this.setDisabled(true);
        }

        if (options.options) {
            this.setOptions(options.options);
        }

        this.setupEventListeners();
    }

    /**
     * Устанавливает опции
     * @param newOptions Массив опций
     */
    public setOptions(newOptions: SelectOption[]): void {
        this.options = [...newOptions];
        this.renderOptions();
    }

    /**
     * Добавляет опцию
     * @param option Опция для добавления
     */
    public addOption(option: SelectOption): void {
        this.options.push(option);
        this.renderOption(option);
    }

    /**
     * Удаляет опцию по значению
     * @param value Значение опции для удаления
     */
    public removeOption(value: string): void {
        const index = this.options.findIndex((opt) => opt.value === value);
        if (index !== -1) {
            this.options.splice(index, 1);
            const optionElement = this.element.querySelector(`option[value="${value}"]`);
            if (optionElement) {
                optionElement.remove();
            }
        }
    }

    /**
     * Получает все опции
     * @returns Массив опций
     */
    public getOptions(): SelectOption[] {
        return [...this.options];
    }

    /**
     * Устанавливает выбранное значение
     * @param value Значение для выбора
     */
    public setValue(value: string): void {
        this.element.value = value;
    }

    /**
     * Получает выбранное значение
     * @returns Текущее значение
     */
    public getValue(): string {
        return this.element.value;
    }

    /**
     * Получает выбранную опцию
     * @returns Выбранная опция или null
     */
    public getSelectedOption(): SelectOption | null {
        const selectedValue = this.element.value;
        return this.options.find((opt) => opt.value === selectedValue) || null;
    }

    /**
     * Получает все выбранные опции (для multiple)
     * @returns Массив выбранных опций
     */
    public getSelectedOptions(): SelectOption[] {
        const selectedValues = Array.from(this.element.selectedOptions).map((opt) => opt.value);
        return this.options.filter((opt) => selectedValues.includes(opt.value));
    }

    /**
     * Устанавливает disabled состояние для опции
     * @param value Значение опции
     * @param disabled Состояние
     */
    public setOptionDisabled(value: string, disabled: boolean): void {
        const option = this.options.find((opt) => opt.value === value);
        if (option) {
            option.disabled = disabled;
            const optionElement = this.element.querySelector(`option[value="${value}"]`);
            if (optionElement) {
                optionElement.disabled = disabled;
            }
        }
    }

    /**
     * Очищает все опции
     */
    public clear(): void {
        this.options = [];
        this.element.innerHTML = '';
    }

    /**
     * Подписывается на изменение
     * @param callback Функция обратного вызова
     */
    public onChange(callback: (value: string) => void): void {
        this.onChangeCallbacks.push(callback);
    }

    /**
     * Отписывается от всех событий
     */
    public off(): void {
        this.onChangeCallbacks = [];
    }

    /**
     * Уничтожает компонент
     */
    protected onDestroy(): void {
        this.off();
        this.options = [];
    }

    /**
     * Рендерит все опции
     */
    private renderOptions(): void {
        this.element.innerHTML = '';
        this.options.forEach((option) => this.renderOption(option));
    }

    /**
     * Рендерит одну опцию
     * @param option Опция для рендера
     */
    private renderOption(option: SelectOption): void {
        const optionElement = document.createElement('option');
        optionElement.value = option.value;
        optionElement.textContent = option.label;

        if (option.disabled) {
            optionElement.disabled = true;
        }

        if (option.selected) {
            optionElement.selected = true;
        }

        this.element.appendChild(optionElement);
    }

    /**
     * Настраивает обработчики событий
     */
    private setupEventListeners(): void {
        this.element.addEventListener('change', this.handleChange);
    }

    /**
     * Обработчик изменения
     */
    private handleChange = (): void => {
        const value = this.element.value;
        this.onChangeCallbacks.forEach((callback) => callback(value));
    };
}
