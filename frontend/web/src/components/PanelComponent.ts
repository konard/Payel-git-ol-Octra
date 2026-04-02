/**
 * CrewAI - Panel Component
 * Компонент панели для сайдбаров и других контейнеров
 */

import { BaseComponent } from './BaseComponent';

export interface PanelOptions {
    title?: string;
    collapsible?: boolean;
    collapsed?: boolean;
    closable?: boolean;
}

/**
 * Компонент панели
 */
export class PanelComponent extends BaseComponent<HTMLDivElement> {
    private headerElement: HTMLDivElement | null = null;
    private contentElement: HTMLDivElement | null = null;
    private titleElement: HTMLHeadingElement | null = null;
    private collapseButton: HTMLButtonElement | null = null;
    private closeButton: HTMLButtonElement | null = null;
    private isCollapsed = false;
    private onCollapseCallbacks: Array<(collapsed: boolean) => void> = [];
    private onCloseCallbacks: Array<() => void> = [];

    /**
     * Конструктор панели
     * @param options Опции панели
     */
    constructor(options: PanelOptions = {}) {
        super('div', 'crewai-panel');

        this.createHeader();

        if (options.title) {
            this.setTitle(options.title);
        }

        if (options.collapsible) {
            this.setCollapsible(true);
            if (options.collapsed) {
                this.collapse();
            }
        }

        if (options.closable) {
            this.setClosable(true);
        }

        this.createContent();
    }

    /**
     * Устанавливает заголовок панели
     * @param title Заголовок
     */
    public setTitle(title: string): void {
        if (!this.titleElement) {
            this.titleElement = document.createElement('h3');
            this.titleElement.className = 'crewai-panel__title';
            this.headerElement?.appendChild(this.titleElement);
        }
        this.titleElement.textContent = title;
    }

    /**
     * Получает заголовок панели
     * @returns Заголовок
     */
    public getTitle(): string {
        return this.titleElement?.textContent || '';
    }

    /**
     * Устанавливает панель как collapsible
     * @param collapsible Состояние collapsible
     */
    public setCollapsible(collapsible: boolean): void {
        if (collapsible && !this.collapseButton) {
            this.collapseButton = document.createElement('button');
            this.collapseButton.className = 'crewai-panel__collapse-btn';
            this.collapseButton.innerHTML = '▼';
            this.collapseButton.addEventListener('click', this.toggleCollapse);
            this.headerElement?.appendChild(this.collapseButton);
        } else if (!collapsible && this.collapseButton) {
            this.collapseButton.remove();
            this.collapseButton = null;
        }
    }

    /**
     * Устанавливает панель как closable
     * @param closable Состояние closable
     */
    public setClosable(closable: boolean): void {
        if (closable && !this.closeButton) {
            this.closeButton = document.createElement('button');
            this.closeButton.className = 'crewai-panel__close-btn';
            this.closeButton.innerHTML = '✕';
            this.closeButton.addEventListener('click', this.handleClose);
            this.headerElement?.appendChild(this.closeButton);
        } else if (!closable && this.closeButton) {
            this.closeButton.remove();
            this.closeButton = null;
        }
    }

    /**
     * Сворачивает панель
     */
    public collapse(): void {
        this.isCollapsed = true;
        this.contentElement?.classList.add('crewai-panel__content--collapsed');
        if (this.collapseButton) {
            this.collapseButton.innerHTML = '▶';
        }
        this.onCollapseCallbacks.forEach((callback) => callback(true));
    }

    /**
     * Разворачивает панель
     */
    public expand(): void {
        this.isCollapsed = false;
        this.contentElement?.classList.remove('crewai-panel__content--collapsed');
        if (this.collapseButton) {
            this.collapseButton.innerHTML = '▼';
        }
        this.onCollapseCallbacks.forEach((callback) => callback(false));
    }

    /**
     * Переключает состояние сворачивания
     */
    public toggleCollapse(): void {
        if (this.isCollapsed) {
            this.expand();
        } else {
            this.collapse();
        }
    }

    /**
     * Проверяет, свернута ли панель
     * @returns True если панель свернута
     */
    public isCollapsedState(): boolean {
        return this.isCollapsed;
    }

    /**
     * Подписывается на изменение состояния сворачивания
     * @param callback Функция обратного вызова
     */
    public onCollapse(callback: (collapsed: boolean) => void): void {
        this.onCollapseCallbacks.push(callback);
    }

    /**
     * Подписывается на закрытие панели
     * @param callback Функция обратного вызова
     */
    public onClose(callback: () => void): void {
        this.onCloseCallbacks.push(callback);
    }

    /**
     * Получает контент элемент
     * @returns Контент элемент
     */
    public getContentElement(): HTMLDivElement | null {
        return this.contentElement;
    }

    /**
     * Устанавливает контент панели
     * @param content Элемент контента
     */
    public setContent(content: HTMLElement): void {
        if (this.contentElement) {
            this.contentElement.innerHTML = '';
            this.contentElement.appendChild(content);
        }
    }

    /**
     * Добавляет элемент в контент
     * @param element Элемент для добавления
     */
    public appendToContent(element: HTMLElement): void {
        if (this.contentElement) {
            this.contentElement.appendChild(element);
        }
    }

    /**
     * Очищает контент
     */
    public clearContent(): void {
        if (this.contentElement) {
            this.contentElement.innerHTML = '';
        }
    }

    /**
     * Уничтожает панель
     */
    protected onDestroy(): void {
        this.onCollapseCallbacks = [];
        this.onCloseCallbacks = [];
        this.headerElement = null;
        this.contentElement = null;
        this.titleElement = null;
        this.collapseButton = null;
        this.closeButton = null;
    }

    /**
     * Создает хедер панели
     */
    private createHeader(): void {
        this.headerElement = document.createElement('div');
        this.headerElement.className = 'crewai-panel__header';
        this.element.appendChild(this.headerElement);
    }

    /**
     * Создает контент панели
     */
    private createContent(): void {
        this.contentElement = document.createElement('div');
        this.contentElement.className = 'crewai-panel__content';
        this.element.appendChild(this.contentElement);
    }

    /**
     * Обработчик закрытия
     */
    private handleClose = (): void => {
        this.onCloseCallbacks.forEach((callback) => callback());
    };
}

/**
 * Компонент скроллируемой панели
 */
export class ScrollablePanelComponent extends PanelComponent {
    /**
     * Конструктор скроллируемой панели
     * @param options Опции панели
     */
    constructor(options: PanelOptions = {}) {
        super(options);
        this.getContentElement()?.classList.add('crewai-panel__content--scrollable');
    }

    /**
     * Прокручивает к верху
     */
    public scrollToTop(): void {
        const content = this.getContentElement();
        if (content) {
            content.scrollTop = 0;
        }
    }

    /**
     * Прокручивает к низу
     */
    public scrollToBottom(): void {
        const content = this.getContentElement();
        if (content) {
            content.scrollTop = content.scrollHeight;
        }
    }

    /**
     * Получает текущую позицию скролла
     * @returns Позиция скролла
     */
    public getScrollPosition(): number {
        const content = this.getContentElement();
        return content ? content.scrollTop : 0;
    }

    /**
     * Устанавливает позицию скролла
     * @param position Позиция скролла
     */
    public setScrollPosition(position: number): void {
        const content = this.getContentElement();
        if (content) {
            content.scrollTop = position;
        }
    }
}
