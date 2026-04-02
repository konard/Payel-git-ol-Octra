/**
 * CrewAI - Node Engine Core
 * Ядро системы управления нодами и соединениями
 */

import type { NodeData, ConnectionData, Point, ProjectData, AgentType, AgentCategory } from '../models/Types';
import { generateId } from '../utils/IdGenerator';
import { EventDispatcher, AppEvents } from '../utils/EventDispatcher';
import { getAgentInfo } from '../models/AgentDefinitions';

/**
 * Основной класс движка нод
 * Управляет созданием, удалением, перемещением и соединением нод
 */
export class NodeEngine {
    private nodes: Map<string, NodeData> = new Map();
    private connections: Map<string, ConnectionData> = new Map();
    private eventDispatcher: EventDispatcher = new EventDispatcher();
    private selectedNodeIds: Set<string> = new Set();
    private selectedConnectionIds: Set<string> = new Set();

    /**
     * Конструктор NodeEngine
     * @param projectData Опциональные данные проекта для загрузки
     */
    constructor(projectData?: ProjectData) {
        if (projectData) {
            this.loadProject(projectData);
        }
    }

    // === Public API ===

    /**
     * Создаёт новую ноду
     * @param type Тип агента
     * @param position Позиция на канвасе
     * @param name Опциональное имя ноды
     * @returns Созданная нода
     */
    public createNode(type: AgentType, position: Point, name?: string): NodeData {
        const agentInfo = getAgentInfo(type);
        if (!agentInfo) {
            throw new Error(`Unknown agent type: ${type}`);
        }

        const node: NodeData = {
            id: generateId(),
            type,
            name: name || agentInfo.label,
            position: { ...position },
            category: agentInfo.category as AgentCategory,
            parameters: { ...agentInfo.defaultParams },
            status: 'idle',
            isSelected: false,
            size: { width: 240, height: 120 },
        };

        this.nodes.set(node.id, node);
        this.eventDispatcher.emit(AppEvents.NODE_CREATED, node);
        this.emitProjectChanged();

        return node;
    }

    /**
     * Удаляет ноду по ID
     * @param nodeId ID ноды для удаления
     * @returns True если нода была удалена
     */
    public deleteNode(nodeId: string): boolean {
        const node = this.nodes.get(nodeId);
        if (!node) {
            return false;
        }

        // Удаляем все соединения связанные с этой нодой
        const connectionsToDelete: string[] = [];
        this.connections.forEach((conn, connId) => {
            if (conn.fromNode === nodeId || conn.toNode === nodeId) {
                connectionsToDelete.push(connId);
            }
        });

        connectionsToDelete.forEach((connId) => this.deleteConnection(connId));

        // Снимаем выделение с ноды
        this.deselectNode(nodeId);

        // Удаляем ноду
        this.nodes.delete(nodeId);
        this.eventDispatcher.emit(AppEvents.NODE_DELETED, { nodeId, node });
        this.emitProjectChanged();

        return true;
    }

    /**
     * Обновляет позицию ноды
     * @param nodeId ID ноды
     * @param position Новая позиция
     */
    public updateNodePosition(nodeId: string, position: Point): void {
        const node = this.nodes.get(nodeId);
        if (!node) {
            return;
        }

        node.position = { ...position };
        this.eventDispatcher.emit(AppEvents.NODE_MOVED, { nodeId, position });
        this.emitProjectChanged();
    }

    /**
     * Обновляет параметры ноды
     * @param nodeId ID ноды
     * @param parameters Новые параметры
     */
    public updateNodeParameters(nodeId: string, parameters: Record<string, unknown>): void {
        const node = this.nodes.get(nodeId);
        if (!node) {
            return;
        }

        node.parameters = { ...node.parameters, ...parameters };
        this.eventDispatcher.emit(AppEvents.NODE_UPDATED, { nodeId, parameters });
        this.emitProjectChanged();
    }

    /**
     * Обновляет имя ноды
     * @param nodeId ID ноды
     * @param name Новое имя
     */
    public updateNodeName(nodeId: string, name: string): void {
        const node = this.nodes.get(nodeId);
        if (!node) {
            return;
        }

        node.name = name;
        this.eventDispatcher.emit(AppEvents.NODE_UPDATED, { nodeId, name });
        this.emitProjectChanged();
    }

    /**
     * Обновляет статус ноды
     * @param nodeId ID ноды
     * @param status Новый статус
     */
    public updateNodeStatus(nodeId: string, status: NodeData['status']): void {
        const node = this.nodes.get(nodeId);
        if (!node) {
            return;
        }

        node.status = status;
        this.eventDispatcher.emit(AppEvents.NODE_UPDATED, { nodeId, status });
    }

    /**
     * Выделяет ноду
     * @param nodeId ID ноды для выделения
     * @param addToSelection Добавить к текущему выделению
     */
    public selectNode(nodeId: string, addToSelection = false): void {
        if (!addToSelection) {
            this.clearSelection();
        }

        this.selectedNodeIds.add(nodeId);
        const node = this.nodes.get(nodeId);
        if (node) {
            node.isSelected = true;
        }

        this.eventDispatcher.emit(AppEvents.NODE_SELECTED, { nodeId, node });
    }

    /**
     * Снимает выделение с ноды
     * @param nodeId ID ноды для снятия выделения
     */
    public deselectNode(nodeId: string): void {
        this.selectedNodeIds.delete(nodeId);
        const node = this.nodes.get(nodeId);
        if (node) {
            node.isSelected = false;
        }

        this.eventDispatcher.emit(AppEvents.NODE_DESELECTED, { nodeId });
    }

    /**
     * Снимает выделение со всех нод
     */
    public clearNodeSelection(): void {
        this.selectedNodeIds.forEach((nodeId) => {
            const node = this.nodes.get(nodeId);
            if (node) {
                node.isSelected = false;
            }
        });
        this.selectedNodeIds.clear();
        this.eventDispatcher.emit(AppEvents.SELECTION_CLEARED, { type: 'nodes' });
    }

    /**
     * Создаёт соединение между нодами
     * @param fromNodeId ID исходной ноды
     * @param toNodeId ID целевой ноды
     * @param fromPort ID порта исхода (опционально)
     * @param toPort ID порта назначения (опционально)
     * @returns Созданное соединение или null если не удалось создать
     */
    public createConnection(
        fromNodeId: string,
        toNodeId: string,
        fromPort?: string,
        toPort?: string
    ): ConnectionData | null {
        // Проверка на существование нод
        const fromNode = this.nodes.get(fromNodeId);
        const toNode = this.nodes.get(toNodeId);

        if (!fromNode || !toNode) {
            return null;
        }

        // Проверка на соединение с самим собой
        if (fromNodeId === toNodeId) {
            return null;
        }

        // Проверка на дублирующееся соединение
        let connectionExists = false;
        this.connections.forEach((conn) => {
            if (
                conn.fromNode === fromNodeId &&
                conn.toNode === toNodeId &&
                conn.fromPort === fromPort &&
                conn.toPort === toPort
            ) {
                connectionExists = true;
            }
        });

        if (connectionExists) {
            return null;
        }

        const connection: ConnectionData = {
            id: generateId(),
            fromNode: fromNodeId,
            toNode: toNodeId,
            fromPort,
            toPort,
            isSelected: false,
        };

        this.connections.set(connection.id, connection);
        this.eventDispatcher.emit(AppEvents.CONNECTION_CREATED, connection);
        this.emitProjectChanged();

        return connection;
    }

    /**
     * Удаляет соединение по ID
     * @param connectionId ID соединения для удаления
     * @returns True если соединение было удалено
     */
    public deleteConnection(connectionId: string): boolean {
        const connection = this.connections.get(connectionId);
        if (!connection) {
            return false;
        }

        this.deselectConnection(connectionId);
        this.connections.delete(connectionId);
        this.eventDispatcher.emit(AppEvents.CONNECTION_DELETED, { connectionId, connection });
        this.emitProjectChanged();

        return true;
    }

    /**
     * Выделяет соединение
     * @param connectionId ID соединения для выделения
     * @param addToSelection Добавить к текущему выделению
     */
    public selectConnection(connectionId: string, addToSelection = false): void {
        if (!addToSelection) {
            this.clearConnectionSelection();
        }

        this.selectedConnectionIds.add(connectionId);
        const connection = this.connections.get(connectionId);
        if (connection) {
            connection.isSelected = true;
        }

        this.eventDispatcher.emit(AppEvents.CONNECTION_SELECTED, { connectionId, connection });
    }

    /**
     * Снимает выделение с соединения
     * @param connectionId ID соединения для снятия выделения
     */
    public deselectConnection(connectionId: string): void {
        this.selectedConnectionIds.delete(connectionId);
        const connection = this.connections.get(connectionId);
        if (connection) {
            connection.isSelected = false;
        }
    }

    /**
     * Снимает выделение со всех соединений
     */
    public clearConnectionSelection(): void {
        this.selectedConnectionIds.forEach((connectionId) => {
            const connection = this.connections.get(connectionId);
            if (connection) {
                connection.isSelected = false;
            }
        });
        this.selectedConnectionIds.clear();
    }

    /**
     * Снимает выделение со всего
     */
    public clearSelection(): void {
        this.clearNodeSelection();
        this.clearConnectionSelection();
    }

    /**
     * Получает ноду по ID
     * @param nodeId ID ноды
     * @returns Нода или undefined
     */
    public getNode(nodeId: string): NodeData | undefined {
        return this.nodes.get(nodeId);
    }

    /**
     * Получает все ноды
     * @returns Массив всех нод
     */
    public getAllNodes(): NodeData[] {
        return Array.from(this.nodes.values());
    }

    /**
     * Получает соединение по ID
     * @param connectionId ID соединения
     * @returns Соединение или undefined
     */
    public getConnection(connectionId: string): ConnectionData | undefined {
        return this.connections.get(connectionId);
    }

    /**
     * Получает все соединения
     * @returns Массив всех соединений
     */
    public getAllConnections(): ConnectionData[] {
        return Array.from(this.connections.values());
    }

    /**
     * Получает выделенные ноды
     * @returns Массив выделенных нод
     */
    public getSelectedNodes(): NodeData[] {
        return Array.from(this.selectedNodeIds)
            .map((id) => this.nodes.get(id))
            .filter((node): node is NodeData => node !== undefined);
    }

    /**
     * Получает выделенные соединения
     * @returns Массив выделенных соединений
     */
    public getSelectedConnections(): ConnectionData[] {
        return Array.from(this.selectedConnectionIds)
            .map((id) => this.connections.get(id))
            .filter((conn): conn is ConnectionData => conn !== undefined);
    }

    /**
     * Получает входящие соединения для ноды
     * @param nodeId ID ноды
     * @returns Массив входящих соединений
     */
    public getIncomingConnections(nodeId: string): ConnectionData[] {
        return Array.from(this.connections.values()).filter((conn) => conn.toNode === nodeId);
    }

    /**
     * Получает исходящие соединения для ноды
     * @param nodeId ID ноды
     * @returns Массив исходящих соединений
     */
    public getOutgoingConnections(nodeId: string): ConnectionData[] {
        return Array.from(this.connections.values()).filter((conn) => conn.fromNode === nodeId);
    }

    /**
     * Экспортирует проект в данные
     * @returns Данные проекта
     */
    public exportProject(): ProjectData {
        const project: ProjectData = {
            id: generateId(),
            name: 'Untitled Project',
            description: '',
            nodes: this.getAllNodes(),
            connections: this.getAllConnections(),
            settings: {
                zoom: 1,
                pan: { x: 0, y: 0 },
                showGrid: true,
                snapToGrid: true,
                gridSize: 20,
            },
            createdAt: Date.now(),
            updatedAt: Date.now(),
            version: '1.0.0',
        };

        return project;
    }

    /**
     * Загружает проект из данных
     * @param projectData Данные проекта
     */
    public loadProject(projectData: ProjectData): void {
        this.clear();

        // Загружаем ноды
        projectData.nodes.forEach((node) => {
            this.nodes.set(node.id, { ...node });
        });

        // Загружаем соединения
        projectData.connections.forEach((conn) => {
            this.connections.set(conn.id, { ...conn });
        });

        this.eventDispatcher.emit(AppEvents.PROJECT_LOADED, projectData);
    }

    /**
     * Очищает все ноды и соединения
     */
    public clear(): void {
        this.nodes.clear();
        this.connections.clear();
        this.clearSelection();
        this.eventDispatcher.emit(AppEvents.PROJECT_CHANGED, { type: 'clear' });
    }

    // === Event System ===

    /**
     * Подписывается на событие
     * @param event Имя события
     * @param callback Функция обратного вызова
     */
    public on<T = unknown>(event: string, callback: (data: T) => void): void {
        this.eventDispatcher.on(event, callback);
    }

    /**
     * Отписывается от события
     * @param event Имя события
     * @param callback Функция обратного вызова
     */
    public off<T = unknown>(event: string, callback: (data: T) => void): void {
        this.eventDispatcher.off(event, callback);
    }

    /**
     * Подписывается на событие один раз
     * @param event Имя события
     * @param callback Функция обратного вызова
     */
    public once<T = unknown>(event: string, callback: (data: T) => void): void {
        this.eventDispatcher.once(event, callback);
    }

    /**
     * Получает количество нод
     * @returns Количество нод
     */
    public getNodeCount(): number {
        return this.nodes.size;
    }

    /**
     * Получает количество соединений
     * @returns Количество соединений
     */
    public getConnectionCount(): number {
        return this.connections.size;
    }

    /**
     * Проверяет, есть ли ноды в проекте
     * @returns True если есть ноды
     */
    public hasNodes(): boolean {
        return this.nodes.size > 0;
    }

    /**
     * Проверяет, можно ли соединить две ноды
     * @param fromNodeId ID исходной ноды
     * @param toNodeId ID целевой ноды
     * @returns True если соединение возможно
     */
    public canConnect(fromNodeId: string, toNodeId: string): boolean {
        if (fromNodeId === toNodeId) {
            return false;
        }

        const fromNode = this.nodes.get(fromNodeId);
        const toNode = this.nodes.get(toNodeId);

        if (!fromNode || !toNode) {
            return false;
        }

        // Проверка на существующее соединение
        let hasConnection = false;
        this.connections.forEach((conn) => {
            if (conn.fromNode === fromNodeId && conn.toNode === toNodeId) {
                hasConnection = true;
            }
        });

        return !hasConnection;
    }

    // === Private Methods ===

    /**
     * Испускает событие изменения проекта
     */
    private emitProjectChanged(): void {
        this.eventDispatcher.emit(AppEvents.PROJECT_CHANGED, {
            nodeCount: this.nodes.size,
            connectionCount: this.connections.size,
        });
    }
}
