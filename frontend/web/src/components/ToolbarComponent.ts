/**
 * CrewAI - Toolbar Component
 * Компонент панели инструментов для канваса
 */

import { BaseComponent } from './BaseComponent';
import { ButtonComponent, ButtonGroupComponent } from './ButtonComponent';

export interface ToolbarOptions {
    showZoom?: boolean;
    showFit?: boolean;
    showGrid?: boolean;
    showActions?: boolean;
}

/**
 * Компонент тулбара
 */
export class ToolbarComponent extends BaseComponent<HTMLDivElement> {
    private buttonGroup: ButtonGroupComponent | null = null;
    private zoomInButton: ButtonComponent | null = null;
    private zoomOutButton: ButtonComponent | null = null;
    private zoomResetButton: ButtonComponent | null = null;
    private fitButton: ButtonComponent | null = null;
    private gridButton: ButtonComponent | null = null;
    private clearButton: ButtonComponent | null = null;
    private saveButton: ButtonComponent | null = null;
    private loadButton: ButtonComponent | null = null;

    private onZoomInCallbacks: Array<() => void> = [];
    private onZoomOutCallbacks: Array<() => void> = [];
    private onZoomResetCallbacks: Array<() => void> = [];
    private onFitCallbacks: Array<() => void> = [];
    private onGridToggleCallbacks: Array<(show: boolean) => void> = [];
    private onClearCallbacks: Array<() => void> = [];
    private onSaveCallbacks: Array<() => void> = [];
    private onLoadCallbacks: Array<() => void> = [];

    private zoomDisplay: HTMLSpanElement | null = null;
    private currentZoom = 100;

    /**
     * Конструктор тулбара
     * @param options Опции тулбара
     */
    constructor(options: ToolbarOptions = {}) {
        super('div', 'crewai-toolbar');

        this.createStructure(options);
    }

    /**
     * Обновляет отображение зума
     * @param zoom Уровень зума (0.1 - 3)
     */
    public updateZoomDisplay(zoom: number): void {
        this.currentZoom = Math.round(zoom * 100);
        if (this.zoomDisplay) {
            this.zoomDisplay.textContent = `${this.currentZoom}%`;
        }
    }

    /**
     * Подписывается на зум ин
     * @param callback Функция обратного вызова
     */
    public onZoomIn(callback: () => void): void {
        this.onZoomInCallbacks.push(callback);
    }

    /**
     * Подписывается на зум аут
     * @param callback Функция обратного вызова
     */
    public onZoomOut(callback: () => void): void {
        this.onZoomOutCallbacks.push(callback);
    }

    /**
     * Подписывается на сброс зума
     * @param callback Функция обратного вызова
     */
    public onZoomReset(callback: () => void): void {
        this.onZoomResetCallbacks.push(callback);
    }

    /**
     * Подписывается на fit to content
     * @param callback Функция обратного вызова
     */
    public onFit(callback: () => void): void {
        this.onFitCallbacks.push(callback);
    }

    /**
     * Подписывается на переключение сетки
     * @param callback Функция обратного вызова
     */
    public onGridToggle(callback: (show: boolean) => void): void {
        this.onGridToggleCallbacks.push(callback);
    }

    /**
     * Подписывается на очистку канваса
     * @param callback Функция обратного вызова
     */
    public onClear(callback: () => void): void {
        this.onClearCallbacks.push(callback);
    }

    /**
     * Подписывается на сохранение
     * @param callback Функция обратного вызова
     */
    public onSave(callback: () => void): void {
        this.onSaveCallbacks.push(callback);
    }

    /**
     * Подписывается на загрузку
     * @param callback Функция обратного вызова
     */
    public onLoad(callback: () => void): void {
        this.onLoadCallbacks.push(callback);
    }

    /**
     * Уничтожает компонент
     */
    protected onDestroy(): void {
        this.onZoomInCallbacks = [];
        this.onZoomOutCallbacks = [];
        this.onZoomResetCallbacks = [];
        this.onFitCallbacks = [];
        this.onGridToggleCallbacks = [];
        this.onClearCallbacks = [];
        this.onSaveCallbacks = [];
        this.onLoadCallbacks = [];

        this.buttonGroup?.destroy();
        this.zoomInButton = null;
        this.zoomOutButton = null;
        this.zoomResetButton = null;
        this.fitButton = null;
        this.gridButton = null;
        this.clearButton = null;
        this.saveButton = null;
        this.loadButton = null;
        this.zoomDisplay = null;
    }

    /**
     * Создает структуру тулбара
     * @param options Опции тулбара
     */
    private createStructure(options: ToolbarOptions): void {
        this.buttonGroup = new ButtonGroupComponent();
        this.buttonGroup.getElement().style.display = 'flex';
        this.buttonGroup.getElement().style.gap = '4px';
        this.element.appendChild(this.buttonGroup.getElement());

        // Zoom controls
        if (options.showZoom !== false) {
            this.createZoomControls();
        }

        // Grid toggle
        if (options.showGrid !== false) {
            this.createGridButton();
        }

        // Fit button
        if (options.showFit !== false) {
            this.createFitButton();
        }

        // Action buttons
        if (options.showActions !== false) {
            this.createActionButtons();
        }
    }

    /**
     * Создает контролы зума
     */
    private createZoomControls(): void {
        if (!this.buttonGroup) {
            return;
        }

        // Zoom out
        this.zoomOutButton = this.buttonGroup.createButton({
            icon: '−',
            variant: 'ghost',
            size: 'sm',
            title: 'Zoom Out',
        });
        this.zoomOutButton.onClick(() => {
            this.onZoomOutCallbacks.forEach((cb) => cb());
        });

        // Zoom reset
        this.zoomResetButton = this.buttonGroup.createButton({
            text: '100%',
            variant: 'ghost',
            size: 'sm',
            title: 'Reset Zoom',
        });
        this.zoomDisplay = this.zoomResetButton.getElement().querySelector('.crewai-button__text');
        this.zoomResetButton.onClick(() => {
            this.onZoomResetCallbacks.forEach((cb) => cb());
        });

        // Zoom in
        this.zoomInButton = this.buttonGroup.createButton({
            icon: '+',
            variant: 'ghost',
            size: 'sm',
            title: 'Zoom In',
        });
        this.zoomInButton.onClick(() => {
            this.onZoomInCallbacks.forEach((cb) => cb());
        });

        // Divider
        const divider = document.createElement('div');
        divider.className = 'crewai-divider--vertical';
        divider.style.margin = '0 8px';
        this.buttonGroup.getElement().appendChild(divider);
    }

    /**
     * Создает кнопку переключения сетки
     */
    private createGridButton(): void {
        if (!this.buttonGroup) {
            return;
        }

        this.gridButton = this.buttonGroup.createButton({
            icon: '▦',
            variant: 'ghost',
            size: 'sm',
            title: 'Toggle Grid',
        });
        this.gridButton.onClick(() => {
            this.onGridToggleCallbacks.forEach((cb) => cb(true));
        });
    }

    /**
     * Создает кнопку fit to content
     */
    private createFitButton(): void {
        if (!this.buttonGroup) {
            return;
        }

        this.fitButton = this.buttonGroup.createButton({
            icon: '⤢',
            variant: 'ghost',
            size: 'sm',
            title: 'Fit to Content',
        });
        this.fitButton.onClick(() => {
            this.onFitCallbacks.forEach((cb) => cb());
        });
    }

    /**
     * Создает кнопки действий
     */
    private createActionButtons(): void {
        if (!this.buttonGroup) {
            return;
        }

        // Divider
        const divider = document.createElement('div');
        divider.className = 'crewai-divider--vertical';
        divider.style.margin = '0 8px';
        this.buttonGroup.getElement().appendChild(divider);

        // Clear
        this.clearButton = this.buttonGroup.createButton({
            icon: 'trash',
            variant: 'ghost',
            size: 'sm',
            title: 'Clear Canvas',
        });
        this.clearButton.onClick(() => {
            this.onClearCallbacks.forEach((cb) => cb());
        });

        // Save
        this.saveButton = this.buttonGroup.createButton({
            icon: 'save',
            variant: 'ghost',
            size: 'sm',
            title: 'Save Project',
        });
        this.saveButton.onClick(() => {
            this.onSaveCallbacks.forEach((cb) => cb());
        });

        // Load
        this.loadButton = this.buttonGroup.createButton({
            icon: 'folder',
            variant: 'ghost',
            size: 'sm',
            title: 'Load Project',
        });
        this.loadButton.onClick(() => {
            this.onLoadCallbacks.forEach((cb) => cb());
        });
    }
}
