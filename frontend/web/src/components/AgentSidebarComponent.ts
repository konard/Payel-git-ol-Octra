/**
 * CrewAI - Agent Sidebar Component
 * Боковая панель с типами агентов для перетаскивания
 */

import { BaseComponent } from './BaseComponent';
import { InputComponent } from './InputComponent';
import { NodeTemplateComponent, NodeTemplateGroupComponent } from './NodeTemplateComponent';
import {
    AGENT_DEFINITIONS,
    getAllCategories,
    getAgentsByCategory,
    searchAgents,
} from '../models/AgentDefinitions';
import type { AgentType, AgentCategory } from '../models/Types';

export interface AgentSidebarOptions {
    collapsible?: boolean;
    searchable?: boolean;
}

/**
 * Компонент боковой панели агентов
 */
export class AgentSidebarComponent extends BaseComponent<HTMLDivElement> {
    private searchInput: InputComponent | null = null;
    private contentElement: HTMLDivElement | null = null;
    private groups: Map<string, NodeTemplateGroupComponent> = new Map();
    private templates: Map<string, NodeTemplateComponent> = new Map();
    private options: AgentSidebarOptions;

    private onDragStartCallbacks: Array<(type: AgentType) => void> = [];
    private onSearchCallbacks: Array<(query: string) => void> = [];

    /**
     * Конструктор боковой панели
     * @param options Опции боковой панели
     */
    constructor(options: AgentSidebarOptions = {}) {
        super('div', 'crewai-sidebar');

        this.options = {
            collapsible: options.collapsible ?? false,
            searchable: options.searchable ?? true,
        };

        this.createStructure();
        this.renderAgentTemplates();
    }

    /**
     * Подписывается на начало перетаскивания
     * @param callback Функция обратного вызова
     */
    public onDragStart(callback: (type: AgentType) => void): void {
        this.onDragStartCallbacks.push(callback);
    }

    /**
     * Подписывается на поиск
     * @param callback Функция обратного вызова
     */
    public onSearch(callback: (query: string) => void): void {
        this.onSearchCallbacks.push(callback);
    }

    /**
     * Уничтожает компонент
     */
    protected onDestroy(): void {
        this.onDragStartCallbacks = [];
        this.onSearchCallbacks = [];

        this.searchInput?.destroy();
        this.groups.forEach((group) => group.destroy());
        this.templates.forEach((template) => template.destroy());
        this.groups.clear();
        this.templates.clear();

        this.searchInput = null;
        this.contentElement = null;
    }

    /**
     * Создает структуру боковой панели
     */
    private createStructure(): void {
        // Хедер с поиском
        const headerElement = document.createElement('div');
        headerElement.className = 'crewai-sidebar__header';
        this.element.appendChild(headerElement);

        const titleElement = document.createElement('div');
        titleElement.className = 'crewai-sidebar__title';
        titleElement.textContent = 'Agents';
        headerElement.appendChild(titleElement);

        // Поиск
        if (this.options.searchable) {
            this.searchInput = new InputComponent({
                type: 'text',
                placeholder: 'Search agents...',
                size: 'sm',
            });
            this.searchInput.onInput(this.handleSearch);
            this.searchInput.render(headerElement);
        }

        // Контент
        this.contentElement = document.createElement('div');
        this.contentElement.className = 'crewai-sidebar__content';
        this.element.appendChild(this.contentElement);
    }

    /**
     * Рендерит шаблоны агентов
     */
    private renderAgentTemplates(): void {
        if (!this.contentElement) {
            return;
        }

        const categories = getAllCategories();

        categories.forEach((category) => {
            const group = this.createCategoryGroup(category);
            this.groups.set(category, group);
            group.render(this.contentElement!);
        });
    }

    /**
     * Создает группу категории
     * @param category Категория
     * @returns Компонент группы
     */
    private createCategoryGroup(category: string): NodeTemplateGroupComponent {
        const categoryLabels: Record<string, string> = {
            chief: '🎯 Chiefs',
            manager: '📋 Managers',
            designer: '🎨 Designers',
            frontend: '🌐 Frontend',
            backend: '🔧 Backend',
            tester: '🧪 Testers',
            devops: '🔄 DevOps',
            analyst: '📊 Analysts',
        };

        const group = new NodeTemplateGroupComponent(categoryLabels[category] || category);
        const agentTypes = getAgentsByCategory(category);

        agentTypes.forEach((type) => {
            const template = new NodeTemplateComponent({
                type: type as AgentType,
                category: category as AgentCategory,
            });
            template.onDragStart((agentType) => {
                this.onDragStartCallbacks.forEach((cb) => cb(agentType));
            });
            group.addTemplate(template);
            this.templates.set(type, template);
        });

        return group;
    }

    /**
     * Фильтрует шаблоны по поисковому запросу
     * @param query Поисковый запрос
     */
    public filterTemplates(query: string): void {
        // Очищаем все группы
        this.groups.forEach((group) => group.clear());
        this.templates.clear();

        if (!query.trim()) {
            // Если запрос пустой, восстанавливаем все шаблоны
            this.renderAgentTemplates();
            return;
        }

        // Ищем агентов
        const results = searchAgents(query);

        if (results.length === 0) {
            // Показываем сообщение "ничего не найдено"
            this.showNoResults();
            return;
        }

        // Группируем результаты по категориям
        const groupedByCategory: Record<string, AgentType[]> = {};

        results.forEach((type) => {
            const info = AGENT_DEFINITIONS[type];
            if (info) {
                if (!groupedByCategory[info.category]) {
                    groupedByCategory[info.category] = [];
                }
                groupedByCategory[info.category].push(type as AgentType);
            }
        });

        // Создаем группы с результатами
        Object.entries(groupedByCategory).forEach(([category, types]) => {
            const group = this.createCategoryGroupWithTypes(category, types);
            group.render(this.contentElement!);
        });
    }

    /**
     * Создает группу с указанными типами
     * @param category Категория
     * @param types Типы агентов
     * @returns Компонент группы
     */
    private createCategoryGroupWithTypes(
        category: string,
        types: AgentType[]
    ): NodeTemplateGroupComponent {
        const categoryLabels: Record<string, string> = {
            chief: '🎯 Chiefs',
            manager: '📋 Managers',
            designer: '🎨 Designers',
            frontend: '🌐 Frontend',
            backend: '🔧 Backend',
            tester: '🧪 Testers',
            devops: '🔄 DevOps',
            analyst: '📊 Analysts',
        };

        const group = new NodeTemplateGroupComponent(categoryLabels[category] || category);

        types.forEach((type) => {
            const template = new NodeTemplateComponent({
                type,
                category: category as AgentCategory,
            });
            template.onDragStart((agentType) => {
                this.onDragStartCallbacks.forEach((cb) => cb(agentType));
            });
            group.addTemplate(template);
            this.templates.set(type, template);
        });

        return group;
    }

    /**
     * Показывает сообщение "ничего не найдено"
     */
    private showNoResults(): void {
        if (!this.contentElement) {
            return;
        }

        const noResultsElement = document.createElement('div');
        noResultsElement.className = 'crewai-sidebar__no-results';
        noResultsElement.style.padding = '20px';
        noResultsElement.style.textAlign = 'center';
        noResultsElement.style.color = 'var(--color-text-tertiary)';
        noResultsElement.style.fontSize = '13px';
        noResultsElement.textContent = 'No agents found';

        this.contentElement.appendChild(noResultsElement);
    }

    /**
     * Обработчик поиска
     */
    private handleSearch = (query: string): void => {
        this.filterTemplates(query);
        this.onSearchCallbacks.forEach((cb) => cb(query));
    };

    /**
     * Сбрасывает фильтр поиска
     */
    public resetSearch(): void {
        if (this.searchInput) {
            this.searchInput.setValue('');
        }
        this.filterTemplates('');
    }

    /**
     * Скроллит к верху
     */
    public scrollToTop(): void {
        if (this.contentElement) {
            this.contentElement.scrollTop = 0;
        }
    }
}
