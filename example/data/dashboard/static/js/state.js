// --- Application State ---
export const appState = {
    apiKey: null,
    scopes: new Set(),
    activePage: null,
    dataCache: {
        stats: null,
        version: null,
        config: null,
        templates: null,
        models: null,
        keys: null,
        whitelist: {ip: null, useragent: null},
    },
    uiState: {
        ipsTable: {showAll: false, sort: {key: 'last_seen', dir: 'desc'}, filter: ''},
        agentsTable: {showAll: false, sort: {key: 'total_hits', dir: 'desc'}, filter: ''},
        selectedTemplate: '',
        selectedModel: '',
    },
    timers: {
        statsRefresh: null,
        trainingStatusPoll: null,
    }
};

// --- DOM Elements ---
export const DOM = {
    body: null,
    loginScreen: null,
    dashboard: null,
    loginForm: null,
    apiKeyInput: null,
    loginError: null,
    pageTitle: null,
    toast: null,
};

export function initDOM() {
    DOM.body = document.body;
    DOM.loginScreen = document.getElementById('login-screen');
    DOM.dashboard = document.getElementById('dashboard');
    DOM.loginForm = document.getElementById('login-form');
    DOM.apiKeyInput = document.getElementById('apiKeyInput');
    DOM.loginError = document.getElementById('login-error');
    DOM.pageTitle = document.getElementById('page-title');
    DOM.toast = document.getElementById('toast');
}
