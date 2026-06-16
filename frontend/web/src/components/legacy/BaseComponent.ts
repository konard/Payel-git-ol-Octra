/**
 * CrewAI - Base UI Component
 * Базовый класс для всех UI компонентов
 */

/**
 * Базовый класс для UI компонентов
 * Предоставляет общую функциональность для всех компонентов
 */
export abstract class BaseComponent<T extends HTMLElement = HTMLElement> {
    protected element: T;
    private isDestroyed = false;

    /**
     * Конструктор компонента
     * @param tagName Тег элемента
     * @param className CSS класс
     */
    protected constructor(tagName: string, className: string) {
        this.element = document.createElement(tagName) as T;
        if (className) {
            this.element.className = className;
        }
    }

    /**
     * Получает корневой элемент компонента
     * @returns Корневой HTML элемент
     */
    public getElement(): T {
        return this.element;
    }

    /**
     * Рендерит компонент в указанный контейнер
     * @param container Контейнер для рендера
     * @param prepend Рендерить в начало
     */
    public render(container: HTMLElement, prepend = false): void {
        if (prepend) {
            container.prepend(this.element);
        } else {
            container.appendChild(this.element);
        }
    }

    /**
     * Удаляет компонент из DOM
     */
    public remove(): void {
        if (this.element.parentNode) {
            this.element.parentNode.removeChild(this.element);
        }
    }

    /**
     * Уничтожает компонент
     */
    public destroy(): void {
        if (this.isDestroyed) {
            return;
        }

        this.isDestroyed = true;
        this.remove();
        this.onDestroy();
    }

    /**
     * Метод вызываемый при уничтожении компонента
     * Переопределяется в наследниках
     */
    protected onDestroy(): void {
        // Переопределяется в наследниках
    }

    /**
     * Проверяет, уничтожен ли компонент
     * @returns True если компонент уничтожен
     */
    public getIsDestroyed(): boolean {
        return this.isDestroyed;
    }

    /**
     * Добавляет CSS класс
     * @param className Имя класса
     */
    public addClass(className: string): void {
        this.element.classList.add(className);
    }

    /**
     * Удаляет CSS класс
     * @param className Имя класса
     */
    public removeClass(className: string): void {
        this.element.classList.remove(className);
    }

    /**
     * Переключает CSS класс
     * @param className Имя класса
     * @param force Принудительно добавить/удалить
     */
    public toggleClass(className: string, force?: boolean): void {
        this.element.classList.toggle(className, force);
    }

    /**
     * Проверяет наличие CSS класса
     * @param className Имя класса
     * @returns True если класс присутствует
     */
    public hasClass(className: string): boolean {
        return this.element.classList.contains(className);
    }

    /**
     * Устанавливает CSS стили
     * @param styles Объект со стилями
     */
    public setStyles(styles: Partial<CSSStyleDeclaration>): void {
        Object.assign(this.element.style, styles);
    }

    /**
     * Устанавливает атрибут
     * @param name Имя атрибута
     * @param value Значение атрибута
     */
    public setAttribute(name: string, value: string): void {
        this.element.setAttribute(name, value);
    }

    /**
     * Получает атрибут
     * @param name Имя атрибута
     * @returns Значение атрибута или null
     */
    public getAttribute(name: string): string | null {
        return this.element.getAttribute(name);
    }

    /**
     * Удаляет атрибут
     * @param name Имя атрибута
     */
    public removeAttribute(name: string): void {
        this.element.removeAttribute(name);
    }

    /**
     * Устанавливает data атрибут
     * @param name Имя атрибута
     * @param value Значение атрибута
     */
    public setData(name: string, value: string): void {
        this.element.setAttribute(`data-${name}`, value);
    }

    /**
     * Получает data атрибут
     * @param name Имя атрибута
     * @returns Значение атрибута или null
     */
    public getData(name: string): string | null {
        return this.element.getAttribute(`data-${name}`);
    }

    /**
     * Добавляет обработчик события
     * @param event Тип события
     * @param handler Обработчик события
     * @param options Опции слушателя
     */
    public on<K extends keyof HTMLElementEventMap>(
        event: K,
        handler: (this: T, ev: HTMLElementEventMap[K]) => void,
        options?: boolean | AddEventListenerOptions
    ): void {
        this.element.addEventListener(event, handler as EventListener, options);
    }

    /**
     * Удаляет обработчик события
     * @param event Тип события
     * @param handler Обработчик события
     * @param options Опции слушателя
     */
    public off<K extends keyof HTMLElementEventMap>(
        event: K,
        handler: (this: T, ev: HTMLElementEventMap[K]) => void,
        options?: boolean | EventListenerOptions
    ): void {
        this.element.removeEventListener(event, handler as EventListener, options);
    }

    /**
     * Добавляет обработчик события один раз
     * @param event Тип события
     * @param handler Обработчик события
     */
    public once<K extends keyof HTMLElementEventMap>(
        event: K,
        handler: (this: T, ev: HTMLElementEventMap[K]) => void
    ): void {
        this.element.addEventListener(event, handler as EventListener, { once: true });
    }

    /**
     * Эмитит кастомное событие
     * @param name Имя события
     * @param detail Детали события
     */
    public emit(name: string, detail?: unknown): void {
        const event = new CustomEvent(name, {
            detail,
            bubbles: true,
            cancelable: true,
        });
        this.element.dispatchEvent(event);
    }

    /**
     * Устанавливает innerHTML
     * @param html HTML контент
     */
    public setHTML(html: string): void {
        this.element.innerHTML = html;
    }

    /**
     * Устанавливает textContent
     * @param text Текстовый контент
     */
    public setText(text: string): void {
        this.element.textContent = text;
    }

    /**
     * Получает textContent
     * @returns Текстовый контент
     */
    public getText(): string {
        return this.element.textContent || '';
    }

    /**
     * Очищает содержимое элемента
     */
    public clear(): void {
        this.element.innerHTML = '';
    }

    /**
     * Показывает компонент
     */
    public show(): void {
        this.element.style.display = '';
    }

    /**
     * Скрывает компонент
     */
    public hide(): void {
        this.element.style.display = 'none';
    }

    /**
     * Проверяет, виден ли компонент
     * @returns True если компонент виден
     */
    public isVisible(): boolean {
        return this.element.style.display !== 'none';
    }

    /**
     * Устанавливает disabled состояние
     * @param disabled Состояние
     */
    public setDisabled(disabled: boolean): void {
        if (disabled) {
            this.element.setAttribute('disabled', 'true');
            this.addClass('disabled');
        } else {
            this.element.removeAttribute('disabled');
            this.removeClass('disabled');
        }
    }

    /**
     * Получает размеры элемента
     * @returns Размеры элемента
     */
    public getBounds(): DOMRect {
        return this.element.getBoundingClientRect();
    }

    /**
     * Фокусирует элемент
     */
    public focus(): void {
        if ('focus' in this.element) {
            (this.element as HTMLElement).focus();
        }
    }

    /**
     * Убирает фокус с элемента
     */
    public blur(): void {
        if ('blur' in this.element) {
            (this.element as HTMLElement).blur();
        }
    }
}
