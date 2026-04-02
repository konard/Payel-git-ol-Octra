/**
 * CrewAI - Core Type Definitions
 * Базовые типы для работы с геометрией и координатами
 */

/**
 * Точка на канвасе
 */
export interface Point {
    x: number;
    y: number;
}

/**
 * Размер элемента
 */
export interface Size {
    width: number;
    height: number;
}

/**
 * Прямоугольник
 */
export interface Rectangle {
    x: number;
    y: number;
    width: number;
    height: number;
}

/**
 * Позиция и размер
 */
export interface Bounds {
    x: number;
    y: number;
    width: number;
    height: number;
}

/**
 * Типы категорий агентов
 */
export type AgentCategory =
    | 'chief'
    | 'manager'
    | 'designer'
    | 'frontend'
    | 'backend'
    | 'tester'
    | 'devops'
    | 'analyst';

/**
 * Типы агентов
 */
export type AgentType =
    // Chiefs
    | 'CreateChief'
    | 'CreateManager'
    | 'CreateTeamLead'
    | 'CreateDirector'
    // Designers
    | 'CreateDesigner'
    | 'CreateUXDesigner'
    | 'CreateUIDesigner'
    | 'CreateGraphicDesigner'
    // Frontend
    | 'CreateFrontend'
    | 'CreateReactDeveloper'
    | 'CreateVueDeveloper'
    | 'CreateAngularDeveloper'
    // Backend
    | 'CreateBackend'
    | 'CreateGoDeveloper'
    | 'CreatePythonDeveloper'
    | 'CreateJavaDeveloper'
    | 'CreateNodeDeveloper'
    // Testers
    | 'CreateTester'
    | 'CreateQAAutomation'
    | 'CreateQAManual'
    | 'CreatePerformanceTester'
    // DevOps
    | 'CreateDevOps'
    | 'CreateCloudEngineer'
    | 'CreateSRE'
    // Analysts
    | 'CreateAnalyst'
    | 'CreateBusinessAnalyst'
    | 'CreateDataAnalyst'
    | 'CreateSystemAnalyst';

/**
 * Информация о типе агента
 */
export interface AgentTypeInfo {
    type: AgentType;
    category: AgentCategory;
    label: string;
    description: string;
    icon: string;
    color: string;
    defaultParams: Record<string, unknown>;
}

/**
 * Порт для соединения нод
 */
export interface Port {
    id: string;
    name: string;
    type: 'input' | 'output';
    label?: string;
}

/**
 * Параметр агента
 */
export interface AgentParameter {
    name: string;
    type: 'string' | 'number' | 'boolean' | 'object' | 'array';
    required: boolean;
    defaultValue?: unknown;
    description?: string;
    options?: string[];
}

/**
 * Данные ноды агента
 */
export interface NodeData {
    id: string;
    type: AgentType;
    name: string;
    position: Point;
    size?: Size;
    category: AgentCategory;
    parameters: Record<string, unknown>;
    status: 'idle' | 'active' | 'error' | 'completed';
    isSelected?: boolean;
}

/**
 * Данные соединения между нодами
 */
export interface ConnectionData {
    id: string;
    fromNode: string;
    toNode: string;
    fromPort?: string;
    toPort?: string;
    isSelected?: boolean;
}

/**
 * Настройки канваса
 */
export interface CanvasSettings {
    zoom: number;
    pan: Point;
    showGrid: boolean;
    snapToGrid: boolean;
    gridSize: number;
}

/**
 * Данные проекта
 */
export interface ProjectData {
    id: string;
    name: string;
    description: string;
    nodes: NodeData[];
    connections: ConnectionData[];
    settings: CanvasSettings;
    createdAt: number;
    updatedAt: number;
    version: string;
}

/**
 * Состояние приложения
 */
export interface AppState {
    project: ProjectData | null;
    selectedNodeIds: string[];
    selectedConnectionIds: string[];
    isDragging: boolean;
    isConnecting: boolean;
    connectionStart: { nodeId: string; portId: string } | null;
    canvasSettings: CanvasSettings;
}

/**
 * Событие перетаскивания
 */
export interface DragEvent {
    type: 'start' | 'move' | 'end';
    position: Point;
    delta: Point;
    node?: NodeData;
}

/**
 * Событие соединения
 */
export interface ConnectEvent {
    type: 'start' | 'move' | 'end' | 'cancel';
    fromNode: string;
    toNode?: string;
    position?: Point;
}

/**
 * Контекст меню
 */
export interface MenuContext {
    type: 'node' | 'connection' | 'canvas';
    position: Point;
    nodeId?: string;
    connectionId?: string;
}

/**
 * Опции экспорта
 */
export interface ExportOptions {
    format: 'json' | 'png' | 'svg';
    includeSettings: boolean;
    prettyPrint: boolean;
}

/**
 * Результат валидации
 */
export interface ValidationResult {
    isValid: boolean;
    errors: string[];
    warnings: string[];
}
