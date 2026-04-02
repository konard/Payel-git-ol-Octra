/**
 * CrewAI - Main Application
 * Главное приложение визуального редактора агентов
 */

import { NodeEngine } from './core/NodeEngine';
import { CanvasEngine } from './core/CanvasEngine';
import { HeaderComponent } from './components/HeaderComponent';
import { AgentSidebarComponent } from './components/AgentSidebarComponent';
import { PropertiesPanelComponent } from './components/PropertiesPanelComponent';
import { ToolbarComponent } from './components/ToolbarComponent';
import { NodeComponent } from './components/NodeComponent';
import { ConnectionLayerComponent } from './components/ConnectionLayerComponent';
import { MinimapComponent } from './components/MinimapComponent';
import { NodeTemplateComponent } from './components/NodeTemplateComponent';
import type { NodeData, Point, AgentType, ConnectionData } from './models/Types';
import { saveProject, loadProject, exportProjectToFile, importProjectFromFile } from './utils/StorageUtils';
import { AppEvents } from './utils/EventDispatcher';

/**
 * Главное приложение CrewAI
 */
export class CrewAIApp {
    private container: HTMLElement;
    private nodeEngine: NodeEngine;
    private canvasEngine: CanvasEngine;
    private headerComponent: HeaderComponent | null = null;
    private sidebarComponent: AgentSidebarComponent | null = null;
    private propertiesComponent: PropertiesPanelComponent | null = null;
    private toolbarComponent: ToolbarComponent | null = null;
    private connectionLayer: ConnectionLayerComponent | null = null;
    private minimap: MinimapComponent | null = null;
    private nodeComponents: Map<string, NodeComponent> = new Map();
    private workspaceElement: HTMLDivElement | null = null;
    private canvasElement: HTMLDivElement | null = null;
    private canvasContent: HTMLDivElement | null = null;
    private selectedNodeId: string | null = null;
    private isConnecting = false;
    private connectionSourceNodeId: string | null = null;
    private tempConnectionPath: SVGPathElement | null = null;

    /**
     * Конструктор приложения
     * @param containerId ID контейнера для приложения
     */
    constructor(containerId: string) {
        const container = document.getElementById(containerId);
        if (!container) {
            throw new Error(`Container with id "${containerId}" not found`);
        }
        this.container = container;

        this.nodeEngine = new NodeEngine();
        this.canvasEngine = new CanvasEngine();

        this.setupEventListeners();
    }

    /**
     * Инициализирует приложение
     */
    public initialize(): void {
        this.createStructure();
        this.createComponents();
        this.setupCanvasInteractions();
        this.setupDragAndDrop();
        this.loadExistingProject();

        console.log('[CrewAI] Application initialized');
    }

    /**
     * Уничтожает приложение
     */
    public destroy(): void {
        this.headerComponent?.destroy();
        this.sidebarComponent?.destroy();
        this.propertiesComponent?.destroy();
        this.toolbarComponent?.destroy();
        this.connectionLayer?.destroy();
        this.minimap?.destroy();

        this.nodeComponents.forEach((comp) => comp.destroy());
        this.nodeComponents.clear();

        this.canvasEngine.destroy();
        this.nodeEngine.clear();

        this.container.innerHTML = '';
    }

    /**
     * Создает структуру приложения
     */
    private createStructure(): void {
        this.container.className = 'crewai-app';
        this.container.innerHTML = '';

        // Хедер
        const headerElement = document.createElement('header');
        this.container.appendChild(headerElement);

        // Рабочая область
        this.workspaceElement = document.createElement('div');
        this.workspaceElement.className = 'crewai-workspace';
        this.container.appendChild(this.workspaceElement);

        // Сайдбар
        const sidebarElement = document.createElement('aside');
        sidebarElement.className = 'crewai-sidebar';
        this.workspaceElement.appendChild(sidebarElement);

        // Канвас
        const canvasContainer = document.createElement('div');
        canvasContainer.className = 'crewai-canvas';
        this.workspaceElement.appendChild(canvasContainer);

        this.canvasElement = canvasContainer;

        // Сетка фона
        const gridElement = document.createElement('div');
        gridElement.className = 'crewai-canvas__grid';
        canvasContainer.appendChild(gridElement);

        // Контент канваса
        this.canvasContent = document.createElement('div');
        this.canvasContent.className = 'crewai-canvas__content';
        canvasContainer.appendChild(this.canvasContent);

        // Слой соединений
        const connectionLayerElement = document.createElement('svg');
        connectionLayerElement.className = 'crewai-canvas__connections';
        canvasContainer.appendChild(connectionLayerElement);

        // Слой нод
        const nodesLayerElement = document.createElement('div');
        nodesLayerElement.className = 'crewai-canvas__nodes';
        canvasContainer.appendChild(nodesLayerElement);

        // Панель свойств
        const propertiesElement = document.createElement('aside');
        propertiesElement.className = 'crewai-properties';
        this.workspaceElement.appendChild(propertiesElement);
    }

    /**
     * Создает компоненты
     */
    private createComponents(): void {
        // Хедер
        this.headerComponent = new HeaderComponent({
            projectName: 'My Crew',
            showStats: true,
        });
        this.headerComponent.getElement().style.gridRow = '1';
        this.headerComponent.getElement().style.gridColumn = '1 / -1';
        this.container.insertBefore(this.headerComponent.getElement(), this.workspaceElement);

        this.headerComponent.onNewProject(() => this.handleNewProject());
        this.headerComponent.onSave(() => this.handleSave());
        this.headerComponent.onExport(() => this.handleExport());

        // Сайдбар
        this.sidebarComponent = new AgentSidebarComponent({
            searchable: true,
            collapsible: false,
        });
        this.sidebarComponent.render(this.workspaceElement!.children[0] as HTMLElement);

        this.sidebarComponent.onDragStart((type) => {
            console.log('[CrewAI] Drag started for agent type:', type);
        });

        // Панель свойств
        this.propertiesComponent = new PropertiesPanelComponent({
            width: 320,
            closable: true,
        });
        this.propertiesComponent.render(this.workspaceElement!.children[2] as HTMLElement);

        this.propertiesComponent.onChange((nodeId, parameters) => {
            this.nodeEngine.updateNodeParameters(nodeId, parameters);
            const nodeComp = this.nodeComponents.get(nodeId);
            if (nodeComp) {
                nodeComp.updateParameters(parameters);
            }
        });

        this.propertiesComponent.onClose(() => {
            this.selectedNodeId = null;
            this.nodeEngine.clearSelection();
        });

        // Тулбар
        this.toolbarComponent = new ToolbarComponent({
            showZoom: true,
            showFit: true,
            showGrid: true,
            showActions: true,
        });
        this.toolbarComponent.render(this.canvasElement!);

        this.toolbarComponent.onZoomIn(() => this.canvasEngine.zoomIn(0.1));
        this.toolbarComponent.onZoomOut(() => this.canvasEngine.zoomOut(0.1));
        this.toolbarComponent.onZoomReset(() => this.canvasEngine.resetZoom());
        this.toolbarComponent.onFit(() => this.handleFitToContent());
        this.toolbarComponent.onGridToggle(() => this.canvasEngine.toggleGrid());
        this.toolbarComponent.onClear(() => this.handleClearCanvas());
        this.toolbarComponent.onSave(() => this.handleSave());
        this.toolbarComponent.onLoad(() => this.handleLoad());

        // Слой соединений
        this.connectionLayer = new ConnectionLayerComponent();
        this.connectionLayer.render(this.canvasElement!);

        this.connectionLayer.onClick((connectionId) => {
            console.log('[CrewAI] Connection clicked:', connectionId);
        });

        this.connectionLayer.onDblClick((connectionId) => {
            this.nodeEngine.deleteConnection(connectionId);
        });

        // Мини-карта
        this.minimap = new MinimapComponent({ width: 200, height: 150 });
        this.minimap.render(this.canvasElement!);

        // Инициализация CanvasEngine
        if (this.canvasElement && this.canvasContent) {
            this.canvasEngine.initialize(this.canvasElement, this.canvasContent);
        }

        // Подписка на события NodeEngine
        this.nodeEngine.on(AppEvents.NODE_CREATED, (node: NodeData) => {
            this.addNodeToCanvas(node);
        });

        this.nodeEngine.on(AppEvents.NODE_DELETED, ({ nodeId }: { nodeId: string }) => {
            this.removeNodeFromCanvas(nodeId);
        });

        this.nodeEngine.on(AppEvents.NODE_MOVED, ({ nodeId, position }: { nodeId: string; position: Point }) => {
            const nodeComp = this.nodeComponents.get(nodeId);
            if (nodeComp) {
                nodeComp.setPosition(position);
            }
            this.updateConnections();
            this.updateMinimap();
        });

        this.nodeEngine.on(AppEvents.NODE_SELECTED, ({ nodeId }: { nodeId: string }) => {
            this.selectNode(nodeId);
        });

        this.nodeEngine.on(AppEvents.CONNECTION_CREATED, (connection: ConnectionData) => {
            this.updateConnections();
        });

        this.nodeEngine.on(AppEvents.CONNECTION_DELETED, () => {
            this.updateConnections();
        });

        this.nodeEngine.on(AppEvents.PROJECT_CHANGED, () => {
            this.updateHeaderStats();
        });
    }

    /**
     * Настраивает взаимодействия с канвасом
     */
    private setupCanvasInteractions(): void {
        if (!this.canvasElement) {
            return;
        }

        // Панорамирование
        this.canvasElement.addEventListener('mousedown', (e: MouseEvent) => {
            if (e.target === this.canvasElement || e.target === this.canvasContent) {
                if (e.button === 1 || (e.button === 0 && e.altKey)) {
                    this.canvasEngine.startPan({ x: e.clientX, y: e.clientY });
                } else if (e.button === 0 && !this.isConnecting) {
                    // Клик по пустому месту снимает выделение
                    this.nodeEngine.clearSelection();
                    this.nodeComponents.forEach((comp) => comp.deselect());
                    this.selectedNodeId = null;
                    this.propertiesComponent?.hideNodeProperties();
                }
            }
        });

        this.canvasElement.addEventListener('mousemove', (e: MouseEvent) => {
            if (this.canvasEngine.isPanningActive()) {
                this.canvasEngine.panTo({ x: e.clientX, y: e.clientY });
            }

            if (this.isConnecting && this.connectionSourceNodeId) {
                this.updateTempConnection(e);
            }
        });

        this.canvasElement.addEventListener('mouseup', (e: MouseEvent) => {
            if (this.canvasEngine.isPanningActive()) {
                this.canvasEngine.endPan();
            }

            if (this.isConnecting) {
                this.endConnection(e);
            }
        });

        this.canvasElement.addEventListener('mouseleave', () => {
            if (this.isConnecting) {
                this.cancelConnection();
            }
        });

        // Drag and drop
        this.canvasElement.addEventListener('dragover', (e: DragEvent) => {
            e.preventDefault();
            e.dataTransfer!.dropEffect = 'copy';
        });

        this.canvasElement.addEventListener('drop', (e: DragEvent) => {
            e.preventDefault();
            const agentType = e.dataTransfer?.getData('application/x-crewai-agent-type');
            if (agentType) {
                const canvasPoint = this.canvasEngine.screenToCanvas({
                    x: e.clientX,
                    y: e.clientY,
                });
                this.nodeEngine.createNode(agentType as AgentType, canvasPoint);
            }
        });
    }

    /**
     * Настраивает drag-and-drop
     */
    private setupDragAndDrop(): void {
        // Обработчики уже настроены в setupCanvasInteractions
    }

    /**
     * Загружает существующий проект
     */
    private loadExistingProject(): void {
        const savedProject = loadProject();
        if (savedProject) {
            this.nodeEngine.loadProject(savedProject);

            // Восстанавливаем ноды
            savedProject.nodes.forEach((node) => {
                this.addNodeToCanvas(node);
            });

            // Восстанавливаем соединения
            this.updateConnections();
        }
    }

    /**
     * Добавляет ноду на канвас
     * @param node Данные ноды
     */
    private addNodeToCanvas(node: NodeData): void {
        const nodeComponent = new NodeComponent({
            nodeData: node,
            draggable: true,
            selectable: true,
            deletable: true,
        });

        nodeComponent.setPosition(node.position);
        nodeComponent.render(this.canvasContent!);

        nodeComponent.onMove((position) => {
            this.nodeEngine.updateNodePosition(node.id, position);
        });

        nodeComponent.onSelect(() => {
            this.selectNode(node.id);
        });

        nodeComponent.onDelete(() => {
            this.nodeEngine.deleteNode(node.id);
        });

        nodeComponent.onPortClick((portType) => {
            if (portType === 'output') {
                this.startConnection(node.id);
            }
        });

        this.nodeComponents.set(node.id, nodeComponent);
        this.updateMinimap();
    }

    /**
     * Удаляет ноду с канваса
     * @param nodeId ID ноды
     */
    private removeNodeFromCanvas(nodeId: string): void {
        const nodeComponent = this.nodeComponents.get(nodeId);
        if (nodeComponent) {
            nodeComponent.destroy();
            this.nodeComponents.delete(nodeId);
        }

        if (this.selectedNodeId === nodeId) {
            this.selectedNodeId = null;
            this.propertiesComponent?.hideNodeProperties();
        }

        this.updateConnections();
        this.updateMinimap();
    }

    /**
     * Выделяет ноду
     * @param nodeId ID ноды
     */
    private selectNode(nodeId: string): void {
        // Снимаем выделение с предыдущей ноды
        if (this.selectedNodeId && this.selectedNodeId !== nodeId) {
            const prevComponent = this.nodeComponents.get(this.selectedNodeId);
            prevComponent?.deselect();
        }

        this.selectedNodeId = nodeId;
        const nodeComponent = this.nodeComponents.get(nodeId);
        nodeComponent?.select();

        const node = this.nodeEngine.getNode(nodeId);
        if (node) {
            this.propertiesComponent?.showNodeProperties(node);
        }
    }

    /**
     * Начинает соединение
     * @param nodeId ID ноды
     */
    private startConnection(nodeId: string): void {
        this.isConnecting = true;
        this.connectionSourceNodeId = nodeId;
        this.canvasElement?.classList.add('crewai-connecting');
    }

    /**
     * Обновляет временное соединение
     * @param e Событие мыши
     */
    private updateTempConnection(e: MouseEvent): void {
        if (!this.connectionLayer || !this.connectionSourceNodeId) {
            return;
        }

        const sourceNode = this.nodeEngine.getNode(this.connectionSourceNodeId);
        if (!sourceNode) {
            return;
        }

        const sourceComp = this.nodeComponents.get(this.connectionSourceNodeId);
        if (!sourceComp) {
            return;
        }

        const sourceRect = sourceComp.getBounds();
        const canvasPoint = this.canvasEngine.screenToCanvas({
            x: e.clientX,
            y: e.clientY,
        });

        const fromPoint: Point = {
            x: sourceRect.right - sourceRect.left,
            y: sourceRect.top + sourceRect.height / 2 - sourceRect.top,
        };

        // Удаляем старое временное соединение
        if (this.tempConnectionPath) {
            this.connectionLayer.removeTempConnection(this.tempConnectionPath);
        }

        // Создаем новое временное соединение
        this.tempConnectionPath = this.connectionLayer.createTempConnection(fromPoint, canvasPoint);
    }

    /**
     * Завершает соединение
     * @param e Событие мыши
     */
    private endConnection(e: MouseEvent): void {
        if (!this.connectionSourceNodeId) {
            return;
        }

        // Проверяем, отпустили ли мышь над другой нодой
        const elements = document.elementsFromPoint(e.clientX, e.clientY);
        const nodeElement = elements.find((el) => el.classList.contains('crewai-node'));

        if (nodeElement) {
            const targetNodeId = nodeElement.getAttribute('data-node-id');
            if (targetNodeId && targetNodeId !== this.connectionSourceNodeId) {
                this.nodeEngine.createConnection(this.connectionSourceNodeId, targetNodeId);
            }
        }

        this.cancelConnection();
    }

    /**
     * Отменяет соединение
     */
    private cancelConnection(): void {
        this.isConnecting = false;
        this.connectionSourceNodeId = null;
        this.canvasElement?.classList.remove('crewai-connecting');

        if (this.tempConnectionPath) {
            this.connectionLayer?.removeTempConnection(this.tempConnectionPath);
            this.tempConnectionPath = null;
        }
    }

    /**
     * Обновляет соединения
     */
    private updateConnections(): void {
        if (!this.connectionLayer) {
            return;
        }

        const connections = this.nodeEngine.getAllConnections();
        const connectionData = connections.map((conn) => {
            const fromNode = this.nodeEngine.getNode(conn.fromNode);
            const toNode = this.nodeEngine.getNode(conn.toNode);
            const fromComp = this.nodeComponents.get(conn.fromNode);
            const toComp = this.nodeComponents.get(conn.toNode);

            let fromPoint: Point = { x: 0, y: 0 };
            let toPoint: Point = { x: 0, y: 0 };

            if (fromComp && fromNode) {
                const rect = fromComp.getBounds();
                fromPoint = {
                    x: rect.right - rect.left,
                    y: rect.top + rect.height / 2 - rect.top,
                };
            }

            if (toComp && toNode) {
                const rect = toComp.getBounds();
                toPoint = {
                    x: rect.left - rect.left,
                    y: rect.top + rect.height / 2 - rect.top,
                };
            }

            return {
                id: conn.id,
                fromPoint,
                toPoint,
                isSelected: conn.isSelected,
            };
        });

        this.connectionLayer.updateAllConnections(connectionData);
    }

    /**
     * Обновляет мини-карту
     */
    private updateMinimap(): void {
        if (!this.minimap || !this.canvasElement) {
            return;
        }

        const nodes = this.nodeEngine.getAllNodes();
        const nodeData = nodes.map((node) => {
            const comp = this.nodeComponents.get(node.id);
            const bounds = comp?.getBounds() || { x: node.position.x, y: node.position.y, width: 200, height: 100 };

            return {
                x: node.position.x,
                y: node.position.y,
                width: bounds.width,
                height: bounds.height,
                color: '#ff6d5a',
            };
        });

        const canvasBounds = this.canvasEngine.getViewportBounds();
        const contentBounds = {
            x: 0,
            y: 0,
            width: Math.max(...nodes.map((n) => n.position.x + 240), canvasBounds.width),
            height: Math.max(...nodes.map((n) => n.position.y + 120), canvasBounds.height),
        };

        this.minimap.update(nodeData, canvasBounds, contentBounds);
    }

    /**
     * Обновляет статистику в хедере
     */
    private updateHeaderStats(): void {
        const nodeCount = this.nodeEngine.getNodeCount();
        const connectionCount = this.nodeEngine.getConnectionCount();
        this.headerComponent?.updateStats(nodeCount, connectionCount);
    }

    /**
     * Обработчик нового проекта
     */
    private handleNewProject(): void {
        if (confirm('Create a new project? All unsaved changes will be lost.')) {
            this.nodeEngine.clear();
            this.nodeComponents.forEach((comp) => comp.destroy());
            this.nodeComponents.clear();
            this.selectedNodeId = null;
            this.propertiesComponent?.hideNodeProperties();
            this.updateConnections();
            this.updateMinimap();
            this.updateHeaderStats();
        }
    }

    /**
     * Обработчик сохранения
     */
    private handleSave(): void {
        const project = this.nodeEngine.exportProject();
        saveProject(project);
        console.log('[CrewAI] Project saved');
    }

    /**
     * Обработчик экспорта
     */
    private handleExport(): void {
        const project = this.nodeEngine.exportProject();
        exportProjectToFile(project, 'crewai-project');
    }

    /**
     * Обработчик загрузки
     */
    private handleLoad(): void {
        const input = document.createElement('input');
        input.type = 'file';
        input.accept = '.json';

        input.onchange = async (e: Event) => {
            const file = (e.target as HTMLInputElement).files?.[0];
            if (file) {
                try {
                    const project = await importProjectFromFile(file);
                    this.nodeEngine.loadProject(project);

                    // Очищаем текущие ноды
                    this.nodeComponents.forEach((comp) => comp.destroy());
                    this.nodeComponents.clear();

                    // Восстанавливаем ноды
                    project.nodes.forEach((node) => {
                        this.addNodeToCanvas(node);
                    });

                    this.updateConnections();
                    this.updateMinimap();
                    this.updateHeaderStats();
                } catch (error) {
                    console.error('[CrewAI] Failed to load project:', error);
                    alert('Failed to load project');
                }
            }
        };

        input.click();
    }

    /**
     * Обработчик fit to content
     */
    private handleFitToContent(): void {
        const nodes = this.nodeEngine.getAllNodes();
        if (nodes.length === 0) {
            return;
        }

        const minX = Math.min(...nodes.map((n) => n.position.x));
        const minY = Math.min(...nodes.map((n) => n.position.y));
        const maxX = Math.max(...nodes.map((n) => n.position.x + 240));
        const maxY = Math.max(...nodes.map((n) => n.position.y + 120));

        this.canvasEngine.fitToContent({
            x: minX - 40,
            y: minY - 40,
            width: maxX - minX + 80,
            height: maxY - minY + 80,
        });
    }

    /**
     * Обработчик очистки канваса
     */
    private handleClearCanvas(): void {
        if (confirm('Clear all nodes and connections?')) {
            this.nodeEngine.clear();
            this.nodeComponents.forEach((comp) => comp.destroy());
            this.nodeComponents.clear();
            this.selectedNodeId = null;
            this.propertiesComponent?.hideNodeProperties();
            this.updateConnections();
            this.updateMinimap();
            this.updateHeaderStats();
        }
    }

    /**
     * Настраивает обработчики событий
     */
    private setupEventListeners(): void {
        // Глобальные обработчики
        window.addEventListener('resize', () => {
            this.updateMinimap();
        });

        // Клавиатурные сокращения
        document.addEventListener('keydown', (e: KeyboardEvent) => {
            // Delete / Backspace - удалить выделенное
            if (e.key === 'Delete' || e.key === 'Backspace') {
                if (this.selectedNodeId) {
                    this.nodeEngine.deleteNode(this.selectedNodeId);
                }
            }

            // Ctrl/Cmd + S - сохранить
            if ((e.ctrlKey || e.metaKey) && e.key === 's') {
                e.preventDefault();
                this.handleSave();
            }

            // Ctrl/Cmd + Z - отменить (TODO)
            if ((e.ctrlKey || e.metaKey) && e.key === 'z') {
                e.preventDefault();
                // TODO: Implement undo
            }

            // Escape - снять выделение
            if (e.key === 'Escape') {
                this.nodeEngine.clearSelection();
                this.nodeComponents.forEach((comp) => comp.deselect());
                this.selectedNodeId = null;
                this.propertiesComponent?.hideNodeProperties();
                this.cancelConnection();
            }
        });
    }
}
