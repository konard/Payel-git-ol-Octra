/**
 * CrewAI - Header Component
 * Верхняя панель приложения с логотипом и навигацией
 */

import { BaseComponent } from './BaseComponent';
import { ButtonComponent, ButtonGroupComponent } from './ButtonComponent';
import { BadgeComponent } from './BadgeComponent';

export interface HeaderOptions {
    projectName?: string;
    showStats?: boolean;
}

/**
 * Компонент хедера приложения
 */
export class HeaderComponent extends BaseComponent<HTMLDivElement> {
    private logoElement: HTMLDivElement | null = null;
    private titleElement: HTMLSpanElement | null = null;
    private navElement: HTMLDivElement | null = null;
    private statsElement: HTMLDivElement | null = null;
    private actionsElement: HTMLDivElement | null = null;
    private nodeCountBadge: BadgeComponent | null = null;
    private connectionCountBadge: BadgeComponent | null = null;
    private newProjectButton: ButtonComponent | null = null;
    private saveButton: ButtonComponent | null = null;
    private exportButton: ButtonComponent | null = null;

    private onNewProjectCallbacks: Array<() => void> = [];
    private onSaveCallbacks: Array<() => void> = [];
    private onExportCallbacks: Array<() => void> = [];

    /**
     * Конструктор хедера
     * @param options Опции хедера
     */
    constructor(options: HeaderOptions = {}) {
        super('header', 'crewai-header');

        this.createStructure(options);
    }

    /**
     * Устанавливает название проекта
     * @param name Название проекта
     */
    public setProjectName(name: string): void {
        if (this.titleElement) {
            this.titleElement.textContent = name;
        }
    }

    /**
     * Обновляет статистику
     * @param nodeCount Количество нод
     * @param connectionCount Количество соединений
     */
    public updateStats(nodeCount: number, connectionCount: number): void {
        if (this.nodeCountBadge) {
            this.nodeCountBadge.setText(`${nodeCount} nodes`);
        }
        if (this.connectionCountBadge) {
            this.connectionCountBadge.setText(`${connectionCount} connections`);
        }
    }

    /**
     * Подписывается на создание нового проекта
     * @param callback Функция обратного вызова
     */
    public onNewProject(callback: () => void): void {
        this.onNewProjectCallbacks.push(callback);
    }

    /**
     * Подписывается на сохранение
     * @param callback Функция обратного вызова
     */
    public onSave(callback: () => void): void {
        this.onSaveCallbacks.push(callback);
    }

    /**
     * Подписывается на экспорт
     * @param callback Функция обратного вызова
     */
    public onExport(callback: () => void): void {
        this.onExportCallbacks.push(callback);
    }

    /**
     * Уничтожает компонент
     */
    protected onDestroy(): void {
        this.onNewProjectCallbacks = [];
        this.onSaveCallbacks = [];
        this.onExportCallbacks = [];

        this.nodeCountBadge?.destroy();
        this.connectionCountBadge?.destroy();
        this.newProjectButton?.destroy();
        this.saveButton?.destroy();
        this.exportButton?.destroy();

        this.logoElement = null;
        this.titleElement = null;
        this.navElement = null;
        this.statsElement = null;
        this.actionsElement = null;
    }

    /**
     * Создает структуру хедера
     * @param options Опции хедера
     */
    private createStructure(options: HeaderOptions): void {
        // Левая часть
        const leftElement = document.createElement('div');
        leftElement.className = 'crewai-header__left';
        this.element.appendChild(leftElement);

        // Логотип
        this.createLogo(leftElement);

        // Название проекта
        this.titleElement = document.createElement('span');
        this.titleElement.className = 'crewai-header__project';
        this.titleElement.textContent = options.projectName || 'Untitled Project';
        this.titleElement.style.fontSize = '14px';
        this.titleElement.style.color = 'var(--color-text-secondary)';
        this.titleElement.style.marginLeft = '12px';
        leftElement.appendChild(this.titleElement);

        // Навигация
        this.navElement = document.createElement('nav');
        this.navElement.className = 'crewai-header__nav';
        this.navElement.style.display = 'flex';
        this.navElement.style.gap = '8px';
        this.navElement.style.marginLeft = '24px';
        this.element.appendChild(this.navElement);

        // Правая часть
        const rightElement = document.createElement('div');
        rightElement.className = 'crewai-header__right';
        this.element.appendChild(rightElement);

        // Статистика
        if (options.showStats !== false) {
            this.createStats(rightElement);
        }

        // Действия
        this.createActions(rightElement);
    }

    /**
     * Создает логотип
     * @param container Контейнер
     */
    private createLogo(container: HTMLElement): void {
        this.logoElement = document.createElement('a');
        this.logoElement.className = 'crewai-header__logo';
        this.logoElement.href = '#';
        this.logoElement.style.display = 'flex';
        this.logoElement.style.alignItems = 'center';
        this.logoElement.style.gap = '10px';
        this.logoElement.style.textDecoration = 'none';
        this.logoElement.style.color = 'var(--color-text-primary)';
        this.logoElement.style.fontSize = '18px';
        this.logoElement.style.fontWeight = '700';

        const iconElement = document.createElement('span');
        iconElement.className = 'crewai-header__logo-icon';
        iconElement.style.display = 'flex';
        iconElement.style.alignItems = 'center';
        iconElement.style.justifyContent = 'center';
        iconElement.style.width = '32px';
        iconElement.style.height = '32px';
        iconElement.style.background = 'var(--gradient-primary)';
        iconElement.style.borderRadius = 'var(--radius-md)';
        iconElement.style.fontSize = '18px';
        iconElement.textContent = '🤖';

        const textElement = document.createElement('span');
        textElement.textContent = 'CrewAI';

        this.logoElement.appendChild(iconElement);
        this.logoElement.appendChild(textElement);
        container.appendChild(this.logoElement);
    }

    /**
     * Создает статистику
     * @param container Контейнер
     */
    private createStats(container: HTMLElement): void {
        this.statsElement = document.createElement('div');
        this.statsElement.className = 'crewai-header__stats';
        this.statsElement.style.display = 'flex';
        this.statsElement.style.gap = '12px';
        this.statsElement.style.marginRight = '16px';

        this.nodeCountBadge = new BadgeComponent({
            text: '0 nodes',
            variant: 'primary',
            size: 'sm',
        });
        this.nodeCountBadge.render(this.statsElement);

        this.connectionCountBadge = new BadgeComponent({
            text: '0 connections',
            variant: 'secondary',
            size: 'sm',
        });
        this.connectionCountBadge.render(this.statsElement);

        container.appendChild(this.statsElement);
    }

    /**
     * Создает кнопки действий
     * @param container Контейнер
     */
    private createActions(container: HTMLElement): void {
        this.actionsElement = document.createElement('div');
        this.actionsElement.className = 'crewai-header__actions';
        this.actionsElement.style.display = 'flex';
        this.actionsElement.style.gap = '8px';
        container.appendChild(this.actionsElement);

        // New Project
        this.newProjectButton = new ButtonComponent({
            text: 'New',
            icon: '📄',
            variant: 'ghost',
            size: 'sm',
        });
        this.newProjectButton.onClick(() => {
            this.onNewProjectCallbacks.forEach((cb) => cb());
        });
        this.newProjectButton.render(this.actionsElement);

        // Save
        this.saveButton = new ButtonComponent({
            text: 'Save',
            icon: '💾',
            variant: 'secondary',
            size: 'sm',
        });
        this.saveButton.onClick(() => {
            this.onSaveCallbacks.forEach((cb) => cb());
        });
        this.saveButton.render(this.actionsElement);

        // Export
        this.exportButton = new ButtonComponent({
            text: 'Export',
            icon: '📤',
            variant: 'primary',
            size: 'sm',
        });
        this.exportButton.onClick(() => {
            this.onExportCallbacks.forEach((cb) => cb());
        });
        this.exportButton.render(this.actionsElement);
    }
}
