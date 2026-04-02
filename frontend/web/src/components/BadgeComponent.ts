/**
 * CrewAI - Badge Component
 * Компонент бейджа для отображения статусов и индикаторов
 */

import { BaseComponent } from './BaseComponent';

export type BadgeVariant = 'primary' | 'secondary' | 'success' | 'warning' | 'error' | 'info';
export type BadgeSize = 'sm' | 'md';

export interface BadgeOptions {
    text?: string;
    variant?: BadgeVariant;
    size?: BadgeSize;
    icon?: string;
    dot?: boolean;
}

/**
 * Компонент бейджа
 */
export class BadgeComponent extends BaseComponent<HTMLSpanElement> {
    private iconElement: HTMLSpanElement | null = null;
    private textElement: Text | null = null;

    /**
     * Конструктор бейджа
     * @param options Опции бейджа
     */
    constructor(options: BadgeOptions = {}) {
        super('span', 'crewai-badge');

        this.setVariant(options.variant || 'primary');
        this.setSize(options.size || 'md');

        if (options.icon) {
            this.setIcon(options.icon);
        }

        if (options.dot) {
            this.setDot(true);
        }

        if (options.text) {
            this.setText(options.text);
        }
    }

    /**
     * Устанавливает вариант бейджа
     * @param variant Вариант бейджа
     */
    public setVariant(variant: BadgeVariant): void {
        this.removeClass('crewai-badge--primary');
        this.removeClass('crewai-badge--secondary');
        this.removeClass('crewai-badge--success');
        this.removeClass('crewai-badge--warning');
        this.removeClass('crewai-badge--error');
        this.removeClass('crewai-badge--info');
        this.addClass(`crewai-badge--${variant}`);
    }

    /**
     * Устанавливает размер бейджа
     * @param size Размер бейджа
     */
    public setSize(size: BadgeSize): void {
        this.removeClass('crewai-badge--sm');
        this.removeClass('crewai-badge--md');
        this.addClass(`crewai-badge--${size}`);
    }

    /**
     * Устанавливает иконку
     * @param icon Текст иконки
     */
    public setIcon(icon: string): void {
        if (!this.iconElement) {
            this.iconElement = document.createElement('span');
            this.iconElement.className = 'crewai-badge__icon';
            this.element.appendChild(this.iconElement);
        }
        this.iconElement.textContent = icon;
    }

    /**
     * Устанавливает текст
     * @param text Текст бейджа
     */
    public setText(text: string): void {
        if (!this.textElement) {
            this.textElement = document.createTextNode('');
            this.element.appendChild(this.textElement);
        }
        this.textElement.textContent = text;
    }

    /**
     * Устанавливает dot индикатор
     * @param show Показывать ли dot
     */
    public setDot(show: boolean): void {
        if (show) {
            this.addClass('crewai-badge--dot');
        } else {
            this.removeClass('crewai-badge--dot');
        }
    }

    /**
     * Уничтожает компонент
     */
    protected onDestroy(): void {
        this.iconElement = null;
        this.textElement = null;
    }
}

/**
 * Компонент группы бейджей
 */
export class BadgeGroupComponent extends BaseComponent<HTMLDivElement> {
    private badges: BadgeComponent[] = [];

    /**
     * Конструктор группы бейджей
     */
    constructor() {
        super('div', 'crewai-badge-group');
    }

    /**
     * Добавляет бейдж в группу
     * @param badge Компонент бейджа
     * @returns Эта группа бейджей
     */
    public addBadge(badge: BadgeComponent): this {
        this.badges.push(badge);
        badge.render(this.element);
        return this;
    }

    /**
     * Создает и добавляет бейдж
     * @param options Опции бейджа
     * @returns Созданный бейдж
     */
    public createBadge(options: BadgeOptions): BadgeComponent {
        const badge = new BadgeComponent(options);
        this.addBadge(badge);
        return badge;
    }

    /**
     * Удаляет бейдж из группы
     * @param badge Бейдж для удаления
     */
    public removeBadge(badge: BadgeComponent): void {
        const index = this.badges.indexOf(badge);
        if (index !== -1) {
            this.badges.splice(index, 1);
            badge.destroy();
        }
    }

    /**
     * Очищает все бейджи
     */
    public clear(): void {
        this.badges.forEach((badge) => badge.destroy());
        this.badges = [];
    }

    /**
     * Получает все бейджи
     * @returns Массив бейджей
     */
    public getBadges(): BadgeComponent[] {
        return [...this.badges];
    }

    /**
     * Уничтожает группу бейджей
     */
    protected onDestroy(): void {
        this.clear();
    }
}
