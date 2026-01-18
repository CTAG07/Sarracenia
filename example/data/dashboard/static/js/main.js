import { appState, DOM, initDOM } from './state.js';
import { apiRequest, logout } from './api.js';
import { getCookie, setCookie, eraseCookie, showToast, toggleButtonLoading } from './utils.js';
import { ThemeManager } from './themes.js';

import { loadStats, setupStatsEventListeners } from './pages/stats.js';
import { loadWhitelist, setupWhitelistEventListeners } from './pages/whitelist.js';
import { loadTemplates, setupTemplatesEventListeners } from './pages/templates.js';
import { loadMarkovModels, setupMarkovEventListeners } from './pages/markov.js';
import { loadApiKeys, setupAuthEventListeners } from './pages/auth.js';
import { loadConfig, setupConfigEventListeners } from './pages/config.js';
import { setupApplicationEventListeners } from './pages/application.js';

document.addEventListener("DOMContentLoaded", () => {
    initDOM();

    // --- Authentication ---
    async function attemptLogin(apiKey) {
        if (!apiKey) {
            DOM.loginScreen.style.display = 'flex';
            return;
        }

        const loginButton = DOM.loginForm.querySelector('button');
        toggleButtonLoading(loginButton, true);
        DOM.loginError.textContent = '';

        try {
            // We can't use apiRequest here directly because it might redirect to logout if appState.apiKey is not set yet.
            // But apiRequest uses appState.apiKey.
            // So we need to manually fetch or temporarily set appState.apiKey.
            
            // Actually, we can just use fetch directly for the initial auth check.
            const response = await fetch('/api/auth/me', {headers: {'sarr-auth': apiKey}});
            if (!response.ok) throw new Error(`Authentication failed (status ${response.status})`);

            const data = await response.json();
            appState.apiKey = apiKey;
            appState.scopes = new Set(data.scopes);
            setCookie('sarr-api-key', apiKey, 365); // Remember for 1 year

            initializeDashboard();
        } catch (error) {
            DOM.loginError.textContent = error.message;
            DOM.loginScreen.style.display = 'flex';
            eraseCookie('sarr-api-key');
        } finally {
            toggleButtonLoading(loginButton, false);
        }
    }

    function initializeDashboard() {
        DOM.loginScreen.style.display = 'none';
        DOM.dashboard.style.display = 'flex';

        // Hide/show UI elements based on scopes
        const hasMasterScope = appState.scopes.has('*');
        document.querySelectorAll('[data-scope]').forEach(el => {
            const requiredScope = el.dataset.scope;
            el.style.display = (hasMasterScope || appState.scopes.has(requiredScope)) ? '' : 'none';
        });

        setupEventListeners();
        initializeScopesSelector();

        // Navigate to the first visible page
        const firstVisiblePage = document.querySelector('nav div[data-scope]:not([style*="display: none"]) .nav-btn');
        if (firstVisiblePage) {
            navigateTo(firstVisiblePage.dataset.target);
        } else {
            DOM.pageTitle.innerHTML = "No Permissions<span>.</span>";
        }
    }

    // --- Navigation ---
    function navigateTo(pageId) {
        if (appState.activePage === pageId) return;

        appState.activePage = pageId;

        // Stop any running timers from the previous page
        if (appState.timers.statsRefresh) {
            clearInterval(appState.timers.statsRefresh);
            appState.timers.statsRefresh = null;
        }
        if (appState.timers.trainingStatusPoll) {
            clearInterval(appState.timers.trainingStatusPoll);
            appState.timers.trainingStatusPoll = null;
        }

        // Update nav button styles
        document.querySelectorAll('.nav-btn').forEach(btn => btn.classList.toggle('active', btn.dataset.target === pageId));

        // Show the correct page content
        document.querySelectorAll('.page-content').forEach(page => page.classList.toggle('active', page.id === pageId));

        // Update header title
        const navBtn = document.querySelector(`.nav-btn[data-target="${pageId}"]`);
        if (navBtn) {
            DOM.pageTitle.innerHTML = `${navBtn.textContent.trim()}<span>.</span>`;
        }

        // Close mobile nav if open
        DOM.body.classList.remove('nav-open');

        // Load data for the new page
        loadDataForPage(pageId);
    }

    function loadDataForPage(pageId) {
        switch (pageId) {
            case 'stats-content':
                loadStats();
                appState.timers.statsRefresh = setInterval(loadStats, 15000);
                break;
            case 'whitelist-content':
                if (!appState.dataCache.whitelist.ip) loadWhitelist();
                break;
            case 'templates-content':
                if (!appState.dataCache.templates) loadTemplates();
                break;
            case 'markov-content':
                if (!appState.dataCache.models) loadMarkovModels();
                break;
            case 'auth-content':
                if (!appState.dataCache.keys) loadApiKeys();
                break;
            case 'server-content':
                if (!appState.dataCache.config) loadConfig();
                break;
            case 'application-content':
                // No data to load for this page
                break;
        }
    }

    function initializeScopesSelector() {
        const scopesContainer = document.getElementById('scopes-container');
        const scopes = {
            "Master": ["*"],
            "Authentication": ["auth:manage"],
            "Server Control": ["server:config", "server:control"],
            "Statistics": ["stats:read"],
            "Whitelists": ["whitelist:read", "whitelist:write"],
            "Templates": ["templates:read", "templates:write"],
            "Markov Models": ["markov:read", "markov:write"],
        };

        let html = '';
        for (const category in scopes) {
            html += `<fieldset><legend>${category}</legend><div class="scopes-grid">`;
            html += scopes[category].map(scope => `
                <div class="scope-item">
                    <input type="checkbox" id="scope-${scope}" value="${scope}">
                    <label for="scope-${scope}">${scope}</label>
                </div>
            `).join('');
            html += `</div></fieldset>`;
        }
        scopesContainer.innerHTML = html;

        const masterCheck = document.getElementById('scope-*');
        masterCheck.addEventListener('change', (e) => {
            const isChecked = e.currentTarget.checked;
            scopesContainer.querySelectorAll('input[type="checkbox"]').forEach(cb => {
                if (cb !== masterCheck) {
                    cb.disabled = isChecked;
                    if (isChecked) cb.checked = false;
                }
            });
        });
    }

    function setupEventListeners() {
        // --- Global ---
        document.getElementById('logoutBtn').addEventListener('click', logout);
        document.getElementById('menu-toggle').addEventListener('click', () => DOM.body.classList.toggle('nav-open'));
        document.getElementById('backdrop').addEventListener('click', () => DOM.body.classList.remove('nav-open'));

        // --- Navigation ---
        document.querySelectorAll('.nav-btn').forEach(btn => {
            btn.addEventListener('click', (e) => navigateTo(e.currentTarget.dataset.target));
        });

        // --- Page Specific ---
        setupStatsEventListeners();
        setupWhitelistEventListeners();
        setupTemplatesEventListeners();
        setupMarkovEventListeners();
        setupAuthEventListeners();
        setupConfigEventListeners();
        setupApplicationEventListeners();
    }

    // --- Initial Load ---
    DOM.loginForm.addEventListener('submit', e => {
        e.preventDefault();
        attemptLogin(DOM.apiKeyInput.value);
    });

    const savedApiKey = getCookie('sarr-api-key');
    // Theme init logic
    ThemeManager.init();
    
    attemptLogin(savedApiKey || '');

    // Global Shortcuts
    document.addEventListener('keydown', e => {
        if ((e.ctrlKey || e.metaKey) && e.key === 's') {
            e.preventDefault();
            const activePage = appState.activePage;
            let formToSubmit = null;

            if (activePage === 'server-content') {
                // Determine which form is visible
                if (document.getElementById('simple-config-editor-view').classList.contains('hidden-view')) {
                    formToSubmit = document.getElementById('updateRawConfigForm');
                } else {
                    formToSubmit = document.getElementById('updateSimpleConfigForm');
                }
            } else if (activePage === 'templates-content') {
                if (!document.getElementById('edit-template-actions').classList.contains('hidden-view')) {
                    formToSubmit = document.getElementById('templateEditForm');
                } else if (!document.getElementById('create-template-form').classList.contains('hidden-view')) {
                    formToSubmit = document.getElementById('createTemplateFormInternal');
                }
            } else if (activePage === 'auth-content') {
                formToSubmit = document.getElementById('createKeyForm');
            }

            if (formToSubmit) {
                formToSubmit.requestSubmit();
            }
        }
    });
});
