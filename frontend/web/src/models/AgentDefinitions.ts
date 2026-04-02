/**
 * CrewAI - Agent Type Definitions
 * Конфигурация всех типов агентов на основе agents.yaml
 */

import type { AgentTypeInfo, AgentParameter } from '../models/Types';

/**
 * Конфигурация всех типов агентов
 */
export const AGENT_DEFINITIONS: Record<string, AgentTypeInfo> = {
    // === Начальники / Руководители ===
    CreateChief: {
        type: 'CreateChief',
        category: 'chief',
        label: 'Chief Agent',
        description: 'Главный агент компании, принимает стратегические решения',
        icon: '👔',
        color: '#ff6d5a',
        defaultParams: {
            name: '',
            department: '',
        },
    },
    CreateManager: {
        type: 'CreateManager',
        category: 'manager',
        label: 'Manager',
        description: 'Управляет командой, распределяет задачи',
        icon: '📋',
        color: '#ff9f4d',
        defaultParams: {
            name: '',
            teamSize: 5,
        },
    },
    CreateTeamLead: {
        type: 'CreateTeamLead',
        category: 'manager',
        label: 'Team Lead',
        description: 'Ведущий разработчик, руководит технической частью',
        icon: '🎯',
        color: '#ff9f4d',
        defaultParams: {
            name: '',
            specialty: '',
        },
    },
    CreateDirector: {
        type: 'CreateDirector',
        category: 'chief',
        label: 'Director',
        description: 'Директор подразделения, стратегическое управление',
        icon: '🏢',
        color: '#ff6d5a',
        defaultParams: {
            name: '',
            division: '',
        },
    },

    // === Дизайнеры ===
    CreateDesigner: {
        type: 'CreateDesigner',
        category: 'designer',
        label: 'Designer',
        description: 'Универсальный дизайнер, создаёт визуальные решения',
        icon: '🎨',
        color: '#d44dff',
        defaultParams: {
            name: '',
            designType: 'general',
        },
    },
    CreateUXDesigner: {
        type: 'CreateUXDesigner',
        category: 'designer',
        label: 'UX Designer',
        description: 'Проектирует пользовательский опыт',
        icon: '🧭',
        color: '#d44dff',
        defaultParams: {
            name: '',
        },
    },
    CreateUIDesigner: {
        type: 'CreateUIDesigner',
        category: 'designer',
        label: 'UI Designer',
        description: 'Создаёт визуальные интерфейсы',
        icon: '🖼️',
        color: '#d44dff',
        defaultParams: {
            name: '',
        },
    },
    CreateGraphicDesigner: {
        type: 'CreateGraphicDesigner',
        category: 'designer',
        label: 'Graphic Designer',
        description: 'Создаёт графический контент',
        icon: '✏️',
        color: '#d44dff',
        defaultParams: {
            name: '',
        },
    },

    // === Frontend разработчики ===
    CreateFrontend: {
        type: 'CreateFrontend',
        category: 'frontend',
        label: 'Frontend Developer',
        description: 'Разрабатывает пользовательские интерфейсы',
        icon: '🌐',
        color: '#00e5ff',
        defaultParams: {
            name: '',
            framework: 'React',
        },
    },
    CreateReactDeveloper: {
        type: 'CreateReactDeveloper',
        category: 'frontend',
        label: 'React Developer',
        description: 'Специалист по React',
        icon: '⚛️',
        color: '#00e5ff',
        defaultParams: {
            name: '',
        },
    },
    CreateVueDeveloper: {
        type: 'CreateVueDeveloper',
        category: 'frontend',
        label: 'Vue Developer',
        description: 'Специалист по Vue.js',
        icon: '💚',
        color: '#00e5ff',
        defaultParams: {
            name: '',
        },
    },
    CreateAngularDeveloper: {
        type: 'CreateAngularDeveloper',
        category: 'frontend',
        label: 'Angular Developer',
        description: 'Специалист по Angular',
        icon: '🅰️',
        color: '#00e5ff',
        defaultParams: {
            name: '',
        },
    },

    // === Backend разработчики ===
    CreateBackend: {
        type: 'CreateBackend',
        category: 'backend',
        label: 'Backend Developer',
        description: 'Разрабатывает серверную логику',
        icon: '🔧',
        color: '#4dff91',
        defaultParams: {
            name: '',
            language: 'Go',
        },
    },
    CreateGoDeveloper: {
        type: 'CreateGoDeveloper',
        category: 'backend',
        label: 'Go Developer',
        description: 'Специалист по Go',
        icon: '🐹',
        color: '#4dff91',
        defaultParams: {
            name: '',
        },
    },
    CreatePythonDeveloper: {
        type: 'CreatePythonDeveloper',
        category: 'backend',
        label: 'Python Developer',
        description: 'Специалист по Python',
        icon: '🐍',
        color: '#4dff91',
        defaultParams: {
            name: '',
        },
    },
    CreateJavaDeveloper: {
        type: 'CreateJavaDeveloper',
        category: 'backend',
        label: 'Java Developer',
        description: 'Специалист по Java',
        icon: '☕',
        color: '#4dff91',
        defaultParams: {
            name: '',
        },
    },
    CreateNodeDeveloper: {
        type: 'CreateNodeDeveloper',
        category: 'backend',
        label: 'Node.js Developer',
        description: 'Специалист по Node.js',
        icon: '📦',
        color: '#4dff91',
        defaultParams: {
            name: '',
        },
    },

    // === Тестировщики ===
    CreateTester: {
        type: 'CreateTester',
        category: 'tester',
        label: 'QA Tester',
        description: 'Тестирует функциональность',
        icon: '🧪',
        color: '#ff4d94',
        defaultParams: {
            name: '',
            testType: 'manual',
        },
    },
    CreateQAAutomation: {
        type: 'CreateQAAutomation',
        category: 'tester',
        label: 'QA Automation',
        description: 'Автоматизирует тестирование',
        icon: '🤖',
        color: '#ff4d94',
        defaultParams: {
            name: '',
        },
    },
    CreateQAManual: {
        type: 'CreateQAManual',
        category: 'tester',
        label: 'QA Manual',
        description: 'Ручное тестирование',
        icon: '📝',
        color: '#ff4d94',
        defaultParams: {
            name: '',
        },
    },
    CreatePerformanceTester: {
        type: 'CreatePerformanceTester',
        category: 'tester',
        label: 'Performance Tester',
        description: 'Тестирует производительность',
        icon: '⚡',
        color: '#ff4d94',
        defaultParams: {
            name: '',
        },
    },

    // === DevOps ===
    CreateDevOps: {
        type: 'CreateDevOps',
        category: 'devops',
        label: 'DevOps Engineer',
        description: 'Автоматизация и инфраструктура',
        icon: '🔄',
        color: '#ffd04d',
        defaultParams: {
            name: '',
            cloudProvider: 'AWS',
        },
    },
    CreateCloudEngineer: {
        type: 'CreateCloudEngineer',
        category: 'devops',
        label: 'Cloud Engineer',
        description: 'Облачная инфраструктура',
        icon: '☁️',
        color: '#ffd04d',
        defaultParams: {
            name: '',
            provider: 'AWS',
        },
    },
    CreateSRE: {
        type: 'CreateSRE',
        category: 'devops',
        label: 'SRE',
        description: 'Site Reliability Engineer',
        icon: '🛡️',
        color: '#ffd04d',
        defaultParams: {
            name: '',
        },
    },

    // === Аналитики ===
    CreateAnalyst: {
        type: 'CreateAnalyst',
        category: 'analyst',
        label: 'Analyst',
        description: 'Анализирует данные и процессы',
        icon: '📊',
        color: '#4d94ff',
        defaultParams: {
            name: '',
            analysisType: 'business',
        },
    },
    CreateBusinessAnalyst: {
        type: 'CreateBusinessAnalyst',
        category: 'analyst',
        label: 'Business Analyst',
        description: 'Бизнес-аналитика',
        icon: '💼',
        color: '#4d94ff',
        defaultParams: {
            name: '',
        },
    },
    CreateDataAnalyst: {
        type: 'CreateDataAnalyst',
        category: 'analyst',
        label: 'Data Analyst',
        description: 'Анализ данных',
        icon: '📈',
        color: '#4d94ff',
        defaultParams: {
            name: '',
        },
    },
    CreateSystemAnalyst: {
        type: 'CreateSystemAnalyst',
        category: 'analyst',
        label: 'System Analyst',
        description: 'Анализ систем и архитектуры',
        icon: '🖥️',
        color: '#4d94ff',
        defaultParams: {
            name: '',
        },
    },
} as const;

/**
 * Параметры для каждого типа агента
 */
export const AGENT_PARAMETERS: Record<string, AgentParameter[]> = {
    CreateChief: [
        { name: 'name', type: 'string', required: true, description: 'Имя агента' },
        { name: 'department', type: 'string', required: true, description: 'Департамент' },
    ],
    CreateManager: [
        { name: 'name', type: 'string', required: true, description: 'Имя агента' },
        { name: 'teamSize', type: 'number', required: true, defaultValue: 5, description: 'Размер команды' },
    ],
    CreateTeamLead: [
        { name: 'name', type: 'string', required: true, description: 'Имя агента' },
        { name: 'specialty', type: 'string', required: true, description: 'Специализация' },
    ],
    CreateDirector: [
        { name: 'name', type: 'string', required: true, description: 'Имя агента' },
        { name: 'division', type: 'string', required: true, description: 'Подразделение' },
    ],
    CreateDesigner: [
        { name: 'name', type: 'string', required: true, description: 'Имя агента' },
        { name: 'designType', type: 'string', required: true, defaultValue: 'general', description: 'Тип дизайна' },
    ],
    CreateUXDesigner: [
        { name: 'name', type: 'string', required: true, description: 'Имя агента' },
    ],
    CreateUIDesigner: [
        { name: 'name', type: 'string', required: true, description: 'Имя агента' },
    ],
    CreateGraphicDesigner: [
        { name: 'name', type: 'string', required: true, description: 'Имя агента' },
    ],
    CreateFrontend: [
        { name: 'name', type: 'string', required: true, description: 'Имя агента' },
        { name: 'framework', type: 'string', required: true, defaultValue: 'React', description: 'Фреймворк' },
    ],
    CreateReactDeveloper: [
        { name: 'name', type: 'string', required: true, description: 'Имя агента' },
    ],
    CreateVueDeveloper: [
        { name: 'name', type: 'string', required: true, description: 'Имя агента' },
    ],
    CreateAngularDeveloper: [
        { name: 'name', type: 'string', required: true, description: 'Имя агента' },
    ],
    CreateBackend: [
        { name: 'name', type: 'string', required: true, description: 'Имя агента' },
        { name: 'language', type: 'string', required: true, defaultValue: 'Go', description: 'Язык программирования' },
    ],
    CreateGoDeveloper: [
        { name: 'name', type: 'string', required: true, description: 'Имя агента' },
    ],
    CreatePythonDeveloper: [
        { name: 'name', type: 'string', required: true, description: 'Имя агента' },
    ],
    CreateJavaDeveloper: [
        { name: 'name', type: 'string', required: true, description: 'Имя агента' },
    ],
    CreateNodeDeveloper: [
        { name: 'name', type: 'string', required: true, description: 'Имя агента' },
    ],
    CreateTester: [
        { name: 'name', type: 'string', required: true, description: 'Имя агента' },
        { name: 'testType', type: 'string', required: true, defaultValue: 'manual', description: 'Тип тестирования' },
    ],
    CreateQAAutomation: [
        { name: 'name', type: 'string', required: true, description: 'Имя агента' },
    ],
    CreateQAManual: [
        { name: 'name', type: 'string', required: true, description: 'Имя агента' },
    ],
    CreatePerformanceTester: [
        { name: 'name', type: 'string', required: true, description: 'Имя агента' },
    ],
    CreateDevOps: [
        { name: 'name', type: 'string', required: true, description: 'Имя агента' },
        { name: 'cloudProvider', type: 'string', required: true, defaultValue: 'AWS', description: 'Облачный провайдер' },
    ],
    CreateCloudEngineer: [
        { name: 'name', type: 'string', required: true, description: 'Имя агента' },
        { name: 'provider', type: 'string', required: true, defaultValue: 'AWS', description: 'Провайдер' },
    ],
    CreateSRE: [
        { name: 'name', type: 'string', required: true, description: 'Имя агента' },
    ],
    CreateAnalyst: [
        { name: 'name', type: 'string', required: true, description: 'Имя агента' },
        { name: 'analysisType', type: 'string', required: true, defaultValue: 'business', description: 'Тип аналитики' },
    ],
    CreateBusinessAnalyst: [
        { name: 'name', type: 'string', required: true, description: 'Имя агента' },
    ],
    CreateDataAnalyst: [
        { name: 'name', type: 'string', required: true, description: 'Имя агента' },
    ],
    CreateSystemAnalyst: [
        { name: 'name', type: 'string', required: true, description: 'Имя агента' },
    ],
};

/**
 * Получает информацию об агенте по типу
 * @param type Тип агента
 * @returns Информация об агенте или undefined
 */
export function getAgentInfo(type: string): AgentTypeInfo | undefined {
    return AGENT_DEFINITIONS[type];
}

/**
 * Получает параметры агента по типу
 * @param type Тип агента
 * @returns Массив параметров
 */
export function getAgentParameters(type: string): AgentParameter[] {
    return AGENT_PARAMETERS[type] || [];
}

/**
 * Получает все агенты указанной категории
 * @param category Категория агентов
 * @returns Массив типов агентов
 */
export function getAgentsByCategory(category: string): string[] {
    return Object.entries(AGENT_DEFINITIONS)
        .filter(([, info]) => info.category === category)
        .map(([type]) => type);
}

/**
 * Получает все доступные категории
 * @returns Массив категорий
 */
export function getAllCategories(): string[] {
    return [...new Set(Object.values(AGENT_DEFINITIONS).map((info) => info.category))];
}

/**
 * Ищет агентов по названию
 * @param query Поисковый запрос
 * @returns Массив типов агентов
 */
export function searchAgents(query: string): string[] {
    const lowerQuery = query.toLowerCase();
    return Object.entries(AGENT_DEFINITIONS)
        .filter(
            ([, info]) =>
                info.label.toLowerCase().includes(lowerQuery) ||
                info.description.toLowerCase().includes(lowerQuery) ||
                info.type.toLowerCase().includes(lowerQuery)
        )
        .map(([type]) => type);
}
