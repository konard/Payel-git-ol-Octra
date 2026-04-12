/**
 * CrewAI - Main Entry Point
 * Точка входа приложения
 */

import './styles/main.css';
import './styles/variables.css';
import './styles/node.css';
import './styles/components.css';

import { CrewAIApp } from './App';

// Обработка глобальных ошибок
window.addEventListener('error', (event) => {
    console.error('[CrewAI] Global error:', event.error);
    const loadingScreen = document.getElementById('loading-screen');
    if (loadingScreen) {
        loadingScreen.innerHTML = `
            <div style="color: #ff4d4d; font-size: 16px; max-width: 600px; padding: 20px;">
                <h2 style="color: #fff; margin-bottom: 12px;">Application Error</h2>
                <p style="margin-bottom: 8px;">${event.error?.message || 'Unknown error'}</p>
                <p style="color: #888; font-size: 12px;">Check console for details</p>
            </div>
        `;
        loadingScreen.style.display = 'flex';
    }
});

// Инициализация приложения после загрузки DOM
document.addEventListener('DOMContentLoaded', () => {
    console.log('[CrewAI] DOMContentLoaded');
    
    try {
        const app = new CrewAIApp('app');
        console.log('[CrewAI] App instance created');
        
        app.initialize();
        console.log('[CrewAI] App initialized');

        // Делаем приложение доступным глобально для отладки
        (window as unknown as Record<string, unknown>).crewAIApp = app;

        console.log('CrewAI Application Started');
        console.log('TIP: Access the app instance via window.crewAIApp');
    } catch (error) {
        console.error('[CrewAI] Initialization error:', error);
        const loadingScreen = document.getElementById('loading-screen');
        if (loadingScreen) {
            loadingScreen.innerHTML = `
                <div style="color: #ff4d4d; font-size: 16px; max-width: 600px; padding: 20px;">
                    <h2 style="color: #fff; margin-bottom: 12px;">Initialization Error</h2>
                    <p style="margin-bottom: 8px;">${error instanceof Error ? error.message : 'Unknown error'}</p>
                    <p style="color: #888; font-size: 12px;">Check console for details</p>
                </div>
            `;
        }
    }
});
