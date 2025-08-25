document.addEventListener("DOMContentLoaded", () => {
    // --- Application State ---
    const appState = {
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
            ipsTable: {showAll: false, sort: {key: 'last_seen', dir: 'desc'}},
            agentsTable: {showAll: false, sort: {key: 'total_hits', dir: 'desc'}},
            selectedTemplate: '',
            selectedModel: '',
        },
        timers: {
            statsRefresh: null,
        }
    };

    // --- DOM Elements ---
    const DOM = {
        body: document.body,
        loginScreen: document.getElementById('login-screen'),
        dashboard: document.getElementById('dashboard'),
        loginForm: document.getElementById('login-form'),
        apiKeyInput: document.getElementById('apiKeyInput'),
        loginError: document.getElementById('login-error'),
        pageTitle: document.getElementById('page-title'),
        toast: document.getElementById('toast'),
    };

    // --- Utility Functions ---

    function debounce(func, delay) {
        let timeout;
        return function (...args) {
            const context = this;
            clearTimeout(timeout);
            timeout = setTimeout(() => func.apply(context, args), delay);
        };
    }

    function showToast(message, type = 'success', duration = 3000) {
        DOM.toast.textContent = message;
        DOM.toast.className = `toast show ${type}`;
        setTimeout(() => {
            DOM.toast.className = 'toast';
        }, duration);
    }

    function setCookie(name, value, days) {
        let expires = "";
        if (days) {
            const date = new Date();
            date.setTime(date.getTime() + (days * 24 * 60 * 60 * 1000));
            expires = "; expires=" + date.toUTCString();
        }
        document.cookie = name + "=" + (value || "") + expires + "; path=/; SameSite=Strict";
    }

    function getCookie(name) {
        const nameEQ = name + "=";
        const ca = document.cookie.split(';');
        for (let i = 0; i < ca.length; i++) {
            let c = ca[i];
            while (c.charAt(0) === ' ') c = c.substring(1, c.length);
            if (c.indexOf(nameEQ) === 0) return c.substring(nameEQ.length, c.length);
        }
        return null;
    }

    function eraseCookie(name) {
        document.cookie = name + '=; Path=/; Expires=Thu, 01 Jan 1970 00:00:01 GMT;';
    }

    function logout() {
        eraseCookie('sarr-api-key');
        appState.apiKey = null;
        window.location.reload();
    }

    function toggleButtonLoading(button, isLoading) {
        if (!button) return;
        button.disabled = isLoading;
        if (isLoading) {
            button.dataset.originalHTML = button.innerHTML;
            button.innerHTML = '<div class="spinner" style="width: 1.2em; height: 1.2em; border-width: 2px; margin: 0;"></div>';
        } else if (button.dataset.originalHTML) {
            button.innerHTML = button.dataset.originalHTML;
            delete button.dataset.originalHTML;
        }
    }

    function triggerDownload(blob, filename) {
        const url = window.URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.style.display = 'none';
        a.href = url;
        a.download = filename;
        document.body.appendChild(a);
        a.click();
        window.URL.revokeObjectURL(url);
        document.body.removeChild(a);
    }

    // --- API Wrapper ---
    async function apiRequest(endpoint, options = {}, button = null) {
        if (!appState.apiKey) {
            logout();
            return Promise.reject(new Error("Not authenticated."));
        }

        toggleButtonLoading(button, true);

        const defaultOptions = {
            headers: {
                'sarr-auth': appState.apiKey,
            },
        };
        // Do not set Content-Type for FormData, browser does it better
        if (!(options.body instanceof FormData)) {
            defaultOptions.headers['Content-Type'] = 'application/json';
        }


        const mergedOptions = {...defaultOptions, ...options};
        mergedOptions.headers = {...defaultOptions.headers, ...options.headers};

        try {
            const response = await fetch(endpoint, mergedOptions);
            if (!response.ok) {
                if (response.status === 401) {
                    logout();
                    throw new Error("Session expired. Please log in again.");
                }
                const errorData = await response.json().catch(() => ({error: `HTTP Error ${response.status}`}));
                throw new Error(errorData.error || `Request failed with status ${response.status}`);
            }

            if (response.status === 204 || response.status === 202) {
                return null; // No content to parse
            }
            // Handle binary file downloads (e.g., model export)
            const contentType = response.headers.get("content-type");
            if (contentType && contentType.includes("application/json")) {
                return response.json();
            }
            return response; // For text, blob, etc.

        } catch (error) {
            showToast(error.message, 'error');
            throw error;
        } finally {
            toggleButtonLoading(button, false);
        }
    }

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

    // --- Data Loading & Rendering ---

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

    async function loadStats(button = null) {
        const ipTbody = document.querySelector('#ips-table tbody');
        const agentTbody = document.querySelector('#agents-table tbody');
        if (!ipTbody.innerHTML) ipTbody.innerHTML = '<tr><td colspan="4"><div class="spinner"></div></td></tr>';
        if (!agentTbody.innerHTML) agentTbody.innerHTML = '<tr><td colspan="4"><div class="spinner"></div></td></tr>';

        try {
            const [summary, ips, agents, version] = await Promise.all([
                apiRequest('/api/stats/summary', {}, button),
                apiRequest('/api/stats/top_ips'),
                apiRequest('/api/stats/top_user_agents'),
                apiRequest('/api/server/version')
            ]);
            appState.dataCache.stats = {summary, ips: ips || [], agents: agents || []};
            appState.dataCache.version = version;
            renderStatsPage();
        } catch (error) {
            ipTbody.innerHTML = `<tr class="empty-row"><td colspan="4">Failed to load stats.</td></tr>`;
            agentTbody.innerHTML = `<tr class="empty-row"><td colspan="4">Failed to load stats.</td></tr>`;
        }
    }

    async function loadWhitelist(button = null) {
        const ipList = document.getElementById('ip-whitelist-list');
        const uaList = document.getElementById('useragent-whitelist-list');
        ipList.innerHTML = `<div class="spinner"></div>`;
        uaList.innerHTML = `<div class="spinner"></div>`;
        try {
            const [ips, useragents] = await Promise.all([
                apiRequest('/api/whitelist/ip', {}, button),
                apiRequest('/api/whitelist/useragent')
            ]);
            appState.dataCache.whitelist.ip = ips || [];
            appState.dataCache.whitelist.useragent = useragents || [];
            renderWhitelistPage();
        } catch (error) {
            ipList.innerHTML = `<li class="list-empty-state">Failed to load IP whitelist.</li>`;
            uaList.innerHTML = `<li class="list-empty-state">Failed to load User Agent whitelist.</li>`;
        }
    }

    async function loadTemplates(button = null) {
        const selector = document.getElementById('templateSelector');
        selector.innerHTML = '<option>Loading...</option>';
        try {
            const templates = await apiRequest('/api/templates', {}, button);
            appState.dataCache.templates = templates;
            renderTemplatesPage();
        } catch (error) {
            selector.innerHTML = '<option>Failed to load templates</option>';
        }
    }

    async function loadMarkovModels(button = null) {
        const selector = document.getElementById('modelSelector');
        selector.innerHTML = '<option>Loading...</option>';
        try {
            const models = await apiRequest('/api/markov/models', {}, button);
            appState.dataCache.models = models;
            renderMarkovPage();
        } catch (error) {
            selector.innerHTML = '<option>Failed to load models</option>';
        }
    }

    async function loadApiKeys(button = null) {
        const tbody = document.querySelector('#keys-table tbody');
        tbody.innerHTML = `<tr><td colspan="4"><div class="spinner"></div></td></tr>`;
        try {
            const keys = await apiRequest('/api/auth/keys', {}, button);
            appState.dataCache.keys = keys || [];
            renderApiKeysPage();
        } catch (error) {
            tbody.innerHTML = `<tr class="empty-row"><td colspan="4">Failed to load API keys.</td></tr>`;
        }
    }

    async function loadConfig(button = null) {
        try {
            const config = await apiRequest('/api/server/config', {}, button);
            appState.dataCache.config = config;
            renderConfigPage();
        } catch (error) {
            // Handle config load error
        }
    }


    function renderStatsPage() {
        const {stats, version} = appState.dataCache;
        if (!stats || !version) return;

        document.getElementById('total-hits').textContent = (stats.summary.total_requests || 0).toLocaleString();
        document.getElementById('unique-ips').textContent = (stats.summary.unique_ips || 0).toLocaleString();
        document.getElementById('unique-agents').textContent = (stats.summary.unique_user_agents || 0).toLocaleString();

        document.getElementById('application-info-list').innerHTML = `
            <li><span class="label">Status</span><span class="value status-indicator">Live</span></li>
            <li><span class="label">Version</span><span class="value">${version.version}</span></li>
            <li><span class="label">Commit</span><span class="value">${version.commit.substring(0, 7)}</span></li>
            <li><span class="label">Build Date</span><span class="value">${new Date(version.build_date).toLocaleString()}</span></li>
        `;

        renderStatsTable('ips-table', stats.ips, appState.uiState.ipsTable);
        renderStatsTable('agents-table', stats.agents, appState.uiState.agentsTable);
    }

    function renderStatsTable(tableId, data, state) {
        const tbody = document.querySelector(`#${tableId} tbody`);
        if (!Array.isArray(data) || data.length === 0) {
            tbody.innerHTML = `<tr class="empty-row"><td colspan="4">No data available.</td></tr>`;
            return;
        }

        const sortedData = [...data].sort((a, b) => {
            if (state.sort.dir === 'none') return 0;
            const valA = a[state.sort.key];
            const valB = b[state.sort.key];
            const direction = state.sort.dir === 'asc' ? 1 : -1;
            if (valA < valB) return -1 * direction;
            if (valA > valB) return 1 * direction;
            return 0;
        });

        const dataToShow = state.showAll ? sortedData : sortedData.slice(0, 5);

        if (dataToShow.length === 0) {
            tbody.innerHTML = `<tr class="empty-row"><td colspan="4">No data available.</td></tr>`;
            return;
        }

        const getRecencyInfo = (lastSeen) => {
            const secondsAgo = (Date.now() - new Date(lastSeen).getTime()) / 1000;
            if (secondsAgo < 600) return {class: "hot", title: "< 10 mins ago"};
            if (secondsAgo < 3600) return {class: "warm", title: "< 1 hour ago"};
            if (secondsAgo < 86400) return {class: "tepid", title: "< 24 hours ago"};
            if (secondsAgo < 259200) return {class: "cold", title: "< 3 days ago"};
            return {class: "frozen", title: "> 3 days ago"};
        };

        tbody.innerHTML = dataToShow.map(item => {
            const identity = item.ip_address || item.user_agent;
            const recency = getRecencyInfo(item.last_seen);
            return `
                <tr>
                    <td><div style="display: flex; align-items: center; gap: 8px;"><span class="recency-dot ${recency.class}"></span>${recency.title}</div></td>
                    <td title="${identity}">${identity}</td>
                    <td>${item.total_hits.toLocaleString()}</td>
                    <td>${new Date(item.last_seen).toLocaleString()}</td>
                </tr>
            `;
        }).join('');
    }

    function renderWhitelistPage() {
        renderWhitelistList('ip');
        renderWhitelistList('useragent');
    }

    function renderWhitelistList(type) {
        const listEl = document.getElementById(`${type}-whitelist-list`);
        const data = appState.dataCache.whitelist[type];

        if (!data) {
            listEl.innerHTML = `<div class="spinner"></div>`;
            return;
        }

        if (data.length === 0) {
            listEl.innerHTML = `<li class="list-empty-state">No ${type === 'ip' ? 'IPs' : 'User Agents'} whitelisted.</li>`;
            return;
        }

        listEl.innerHTML = data.map(item => `
            <li class="whitelist-item" data-value="${item}">
                <span class="value">${item}</span>
                <button type="button" class="remove-item-btn" data-type="${type}" data-value="${item}" title="Remove">
                    <svg class="btn-icon" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" /></svg>
                </button>
            </li>
        `).join('');
    }

    function renderTemplatesPage() {
        const selector = document.getElementById('templateSelector');
        selector.innerHTML = `<option value="">-- Select a Template --</option><option value="--new--">-- Create New Template --</option>`;
        appState.dataCache.templates.forEach(name => {
            selector.add(new Option(name, name));
        });
        selector.value = appState.uiState.selectedTemplate || '';
        handleTemplateSelection();
    }

    function renderMarkovPage() {
        const selector = document.getElementById('modelSelector');
        selector.innerHTML = `<option value="">-- Select a Model --</option><option value="--new--">-- Create New Model --</option>`;
        appState.dataCache.models.forEach(model => {
            selector.add(new Option(`${model.Name} (Order: ${model.Order})`, model.Name));
        });
        selector.value = appState.uiState.selectedModel || '';
        handleModelSelection();
    }

    function renderApiKeysPage() {
        const tbody = document.querySelector('#keys-table tbody');
        const keys = appState.dataCache.keys;

        if (keys.length === 0) {
            tbody.innerHTML = `<tr class="empty-row"><td colspan="4">No API keys exist. The first key created will be a master key.</td></tr>`;
            return;
        }

        const renderKeyRow = key => `
            <tr>
                <td>${key.id}</td>
                <td>${key.description}</td>
                <td class="scopes-cell">${key.scopes.join(', ')}</td>
                <td>
                    <button type="button" class="remove-item-btn danger" data-key-id="${key.id}" ${key.id === 1 ? 'disabled title="Master key cannot be deleted"' : 'title="Delete key"'}>
                        <svg class="btn-icon" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" /></svg>
                    </button>
                </td>
            </tr>
        `;

        tbody.innerHTML = keys.map(renderKeyRow).join('');
    }

    function renderConfigPage() {
        const config = appState.dataCache.config;
        document.getElementById('appConfigText').value = JSON.stringify(config, null, 2);
        buildSimpleConfigEditor(config.server_config, 'server_config', document.getElementById('server_config-editor'));
        buildSimpleConfigEditor(config.template_config, 'template_config', document.getElementById('template_config-editor'));
        buildSimpleConfigEditor(config.threat_config, 'threat_config', document.getElementById('threat_config-editor'));
    }

    function buildSimpleConfigEditor(obj, prefix, container, level = 0) {
        if (level === 0) container.innerHTML = '';
        for (const key in obj) {
            const value = obj[key];
            const fullKey = `${prefix}.${key}`;
            const field = document.createElement('div');
            field.className = 'config-field';
            field.innerHTML = `<label for="config-${fullKey}" class="config-field-label">${key}</label>`;

            if (typeof value === 'object' && !Array.isArray(value) && value !== null) {
                const header = document.createElement('h4');
                header.className = 'config-nested-header';
                header.textContent = key;
                container.appendChild(header);
                buildSimpleConfigEditor(value, fullKey, container, level + 1);
            } else if (Array.isArray(value)) {
                const arrayContainer = document.createElement('div');
                arrayContainer.className = 'array-field-container';
                arrayContainer.id = `config-${fullKey}`;
                arrayContainer.dataset.key = fullKey;

                const createArrayItem = (itemValue = "") => {
                    const item = document.createElement('div');
                    item.className = 'array-item';
                    const input = document.createElement('input');
                    input.type = 'text';
                    input.value = itemValue;
                    const removeBtn = document.createElement('button');
                    removeBtn.type = 'button';
                    removeBtn.className = 'remove-item-btn danger array-item-remove-btn';
                    removeBtn.innerHTML = '<svg class="btn-icon" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" /></svg>';
                    removeBtn.onclick = () => item.remove();
                    item.appendChild(input);
                    item.appendChild(removeBtn);
                    return item;
                };

                value.forEach(item => arrayContainer.appendChild(createArrayItem(item)));

                const addBtn = document.createElement('button');
                addBtn.type = 'button';
                addBtn.className = 'secondary array-item-add-btn';
                addBtn.textContent = 'Add Item';
                addBtn.onclick = () => arrayContainer.appendChild(createArrayItem());

                field.appendChild(arrayContainer);
                field.appendChild(addBtn);
            } else {
                const input = document.createElement('input');
                input.id = `config-${fullKey}`;
                input.dataset.key = fullKey;
                if (typeof value === 'boolean') {
                    input.type = 'checkbox';
                    input.checked = value;
                } else if (typeof value === 'number') {
                    input.type = 'number';
                    input.value = value;
                    input.step = 'any';
                } else {
                    input.type = 'text';
                    input.value = value;
                }
                field.appendChild(input);
            }
            container.appendChild(field);
        }
    }

    // --- Event Handlers ---

    async function handleTemplateSelection() {
        const selected = document.getElementById('templateSelector').value;
        appState.uiState.selectedTemplate = selected;

        const isNew = selected === '--new--';
        const isEdit = selected && !isNew;

        document.getElementById('create-template-form').classList.toggle('hidden-view', !isNew);
        document.getElementById('edit-template-actions').classList.toggle('hidden-view', !isEdit);
        document.getElementById('template-preview-card').classList.toggle('hidden-view', !isEdit);

        if (isEdit) {
            const contentArea = document.getElementById('templateContent');
            contentArea.value = "Loading...";
            const content = await apiRequest(`/api/templates/${selected}`, {headers: {'Content-Type': 'text/plain'}}, null).then(res => res.text());
            contentArea.value = content;
            previewTemplate(selected, document.getElementById('previewThreat').value);
        }
    }

    async function previewTemplate(name, threat, content = null, button = null) {
        const frame = document.getElementById('template-preview-frame');
        const openTabBtn = document.getElementById('openPreviewNewTabBtn');

        try {
            let endpoint = `/api/templates/preview?name=${name}&threat=${threat}`;
            let options = {};
            if (content !== null) {
                endpoint = `/api/templates/test?threat=${threat}`;
                options = {method: 'POST', body: content, headers: {'Content-Type': 'text/plain'}};
            }
            const response = await apiRequest(endpoint, options, button);
            const blob = await response.blob();
            const url = URL.createObjectURL(blob);
            frame.src = url;
            openTabBtn.disabled = false;
            openTabBtn.onclick = () => window.open(url, '_blank');
        } catch (e) {
            frame.src = 'about:blank';
            openTabBtn.disabled = true;
        }
    }

    function handleModelSelection() {
        const selected = document.getElementById('modelSelector').value;
        appState.uiState.selectedModel = selected;
        const isNew = selected === '--new--';
        const isEdit = selected && !isNew;
        document.getElementById('create-model-form').classList.toggle('hidden-view', !isNew);
        document.getElementById('existing-model-actions').classList.toggle('hidden-view', !isEdit);
    }

    function applyTheme(theme) {
        document.documentElement.setAttribute('data-theme', theme);
        setCookie('sarr-theme', theme, 365);
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


        // --- Stats Page ---
        document.getElementById('refreshStatsBtn').addEventListener('click', (e) => loadStats(e.currentTarget));
        document.querySelectorAll('.show-all-btn').forEach(btn => {
            btn.addEventListener('click', e => {
                const tableKey = e.currentTarget.dataset.table === 'ips' ? 'ipsTable' : 'agentsTable';
                appState.uiState[tableKey].showAll = !appState.uiState[tableKey].showAll;
                e.currentTarget.textContent = appState.uiState[tableKey].showAll ? 'Show Top 5' : 'Show All';
                renderStatsPage();
            });
        });
        document.querySelectorAll('.table-container table thead').forEach(header => {
            header.addEventListener('click', e => {
                const th = e.target.closest('th.sortable');
                if (!th) return;

                const tableId = th.closest('table').id;
                const tableKey = tableId === 'ips-table' ? 'ipsTable' : 'agentsTable';
                const sortKey = th.dataset.sort;
                const state = appState.uiState[tableKey];

                if (state.sort.key === sortKey) {
                    if (state.sort.dir === 'desc') state.sort.dir = 'asc';
                    else if (state.sort.dir === 'asc') state.sort.dir = 'none';
                    else state.sort.dir = 'desc';
                } else {
                    state.sort.key = sortKey;
                    state.sort.dir = 'desc';
                }

                th.parentElement.querySelectorAll('th.sortable').forEach(el => el.removeAttribute('data-sort-dir'));
                if (state.sort.dir !== 'none') {
                    th.setAttribute('data-sort-dir', state.sort.dir);
                }

                renderStatsPage();
            });
        });


        // --- Whitelist Page ---
        document.getElementById('refreshWhitelistBtn').addEventListener('click', e => loadWhitelist(e.currentTarget));
        document.getElementById('whitelist-tab-nav').addEventListener('click', e => {
            if (!e.target.matches('.tab-btn')) return;
            const type = e.target.dataset.type;
            document.querySelectorAll('#whitelist-tab-nav .tab-btn').forEach(btn => btn.classList.remove('active'));
            document.querySelectorAll('#whitelist-content .tab-panel').forEach(panel => panel.classList.remove('active'));
            e.target.classList.add('active');
            document.getElementById(`${type}-panel`).classList.add('active');
            if (!appState.dataCache.whitelist[type]) {
                loadWhitelist();
            }
        });

        document.getElementById('add-ip-whitelist-form').addEventListener('submit', handleAddWhitelist);
        document.getElementById('add-useragent-whitelist-form').addEventListener('submit', handleAddWhitelist);

        document.getElementById('whitelist-content').addEventListener('click', e => {
            const removeBtn = e.target.closest('.remove-item-btn');
            if (removeBtn) {
                const {type, value} = removeBtn.dataset;
                if (confirm(`Are you sure you want to remove "${value}"?`)) {
                    handleRemoveWhitelist(type, value, removeBtn);
                }
            }
        });

        document.getElementById('whitelist-import-file-input').addEventListener('change', handleWhitelistImport);
        document.querySelectorAll('.import-whitelist-btn').forEach(btn => {
            btn.addEventListener('click', (e) => {
                const type = e.currentTarget.dataset.type;
                const fileInput = document.getElementById('whitelist-import-file-input');
                fileInput.dataset.importType = type; // Store the type for the change event
                fileInput.click();
            });
        });


        // --- Template Page ---
        document.getElementById('refreshTemplateListBtn').addEventListener('click', e => loadTemplates(e.currentTarget));
        document.getElementById('templateSelector').addEventListener('change', handleTemplateSelection);

        const debouncedPreview = debounce((name, threatValue) => {
            if (name && name !== '--new--') {
                previewTemplate(name, threatValue);
            }
        }, 250);

        document.getElementById('previewThreat').addEventListener('input', e => {
            const name = document.getElementById('templateSelector').value;
            const threatValue = e.target.value;
            document.getElementById('previewThreatValue').textContent = threatValue;
            debouncedPreview(name, threatValue);
        });

        document.getElementById('createTemplateFormInternal').addEventListener('submit', handleCreateTemplate);
        document.getElementById('templateEditForm').addEventListener('submit', handleSaveTemplate);
        document.getElementById('deleteTemplateBtn').addEventListener('click', handleDeleteTemplate);

        document.getElementById('testTemplateBtn').addEventListener('click', e => {
            const name = document.getElementById('templateSelector').value;
            const content = document.getElementById('templateContent').value;
            const threat = document.getElementById('previewThreat').value;
            previewTemplate(name, threat, content, e.currentTarget);
        });

        document.getElementById('testNewTemplateBtn').addEventListener('click', e => {
            const content = document.getElementById('newTemplateContent').value;
            if (!content) {
                showToast('Cannot test an empty template.', 'error');
                return;
            }
            const threat = document.getElementById('previewThreat').value;
            previewTemplate('test.tmpl.html', threat, content, e.currentTarget);
        });


        // --- Markov Page ---
        document.getElementById('refreshModelListBtn').addEventListener('click', e => loadMarkovModels(e.currentTarget));
        document.getElementById('modelSelector').addEventListener('change', handleModelSelection);
        document.getElementById('createModelFormInternal').addEventListener('submit', async e => {
            e.preventDefault();
            const name = document.getElementById('createModelName').value;
            const order = parseInt(document.getElementById('createModelOrder').value);
            try {
                const newModel = await apiRequest('/api/markov/models', {
                    method: 'POST',
                    body: JSON.stringify({name, order})
                }, e.currentTarget.querySelector('button'));
                showToast(`Model "${newModel.Name}" created.`);
                appState.uiState.selectedModel = newModel.Name;
                await loadMarkovModels();
            } catch (error) {
                showToast(error.message, 'error');
            }
        });
        document.getElementById('deleteModelBtn').addEventListener('click', async e => {
            const name = document.getElementById('modelSelector').value;
            if (confirm(`Are you sure you want to permanently delete model "${name}"?`)) {
                try {
                    await apiRequest(`/api/markov/models/${name}`, {method: 'DELETE'}, e.currentTarget);
                    showToast(`Model "${name}" deleted.`);
                    appState.uiState.selectedModel = '';
                    await loadMarkovModels();
                } catch (error) {
                    showToast(error.message, 'error');
                }
            }
        });
        document.getElementById('trainModelForm').addEventListener('submit', async e => {
            e.preventDefault();
            const modelName = document.getElementById('modelSelector').value;
            const fileInput = document.getElementById('trainCorpusFile');
            if (fileInput.files.length === 0) {
                showToast('Please select a corpus file.', 'error');
                return;
            }
            const file = fileInput.files[0];
            await apiRequest(`/api/markov/models/${modelName}/train`, {
                method: 'POST',
                body: file,
                headers: {'Content-Type': 'text/plain'}
            }, e.currentTarget.querySelector('button'));
            showToast(`File accepted for model "${modelName}". This will block database writes until it is done.`);
            fileInput.value = '';
        });

        document.getElementById('pruneModelForm').addEventListener('submit', async e => {
            e.preventDefault();
            const modelName = document.getElementById('modelSelector').value;
            const minFreq = document.getElementById('pruneMinFreq').value;
            await apiRequest(`/api/markov/models/${modelName}/prune`, {
                method: 'POST',
                body: JSON.stringify({minFreq: parseInt(minFreq)})
            }, e.currentTarget.querySelector('button'));
            showToast(`Pruning complete for "${modelName}".`);
        });

        document.getElementById('exportModelBtn').addEventListener('click', async e => {
            const modelName = document.getElementById('modelSelector').value;
            try {
                const response = await apiRequest(`/api/markov/models/${modelName}/export`, {}, e.currentTarget);
                const blob = await response.blob();
                triggerDownload(blob, `${modelName}.json`);
            } catch (error) {
                showToast('Failed to export model.', 'error');
            }
        });

        document.getElementById('importModelForm').addEventListener('submit', async e => {
            e.preventDefault();
            const fileInput = document.getElementById('importModelFile');
            if (fileInput.files.length === 0) {
                showToast('Please select a model file to import.', 'error');
                return;
            }
            const file = fileInput.files[0];
            await apiRequest(`/api/markov/import`, {
                method: 'POST',
                body: file,
                headers: {'Content-Type': 'application/json'}
            }, e.currentTarget.querySelector('button'));
            showToast(`Model import started.`);
            fileInput.value = '';
            await loadMarkovModels(); // Refresh list after import
        });

        document.getElementById('pruneVocabForm').addEventListener('submit', async e => {
            e.preventDefault();
            const minFreq = document.getElementById('pruneVocabMinFreq').value;
            await apiRequest(`/api/markov/vocabulary/prune`, {
                method: 'POST',
                body: JSON.stringify({minFreq: parseInt(minFreq)})
            }, e.currentTarget.querySelector('button'));
            showToast(`Global vocabulary pruning complete.`);
        });


        // --- Auth Page ---
        document.getElementById('refreshKeysListBtn').addEventListener('click', e => loadApiKeys(e.currentTarget));
        document.getElementById('createKeyForm').addEventListener('submit', async e => {
            e.preventDefault();
            const scopes = Array.from(document.querySelectorAll('#scopes-container input:checked')).map(el => el.value);
            const description = document.getElementById('createKeyDesc').value;
            const newKey = await apiRequest('/api/auth/keys', {
                method: 'POST',
                body: JSON.stringify({description, scopes})
            }, e.currentTarget.querySelector('button'));
            document.getElementById('newKeyValue').value = newKey.raw_key;
            document.getElementById('key-created-result').classList.add('visible');
            await loadApiKeys();
        });
        document.getElementById('keys-table').addEventListener('click', async e => {
            const deleteBtn = e.target.closest('.remove-item-btn[data-key-id]');
            if (!deleteBtn) return;
            const keyId = deleteBtn.dataset.keyId;
            if (confirm(`Are you sure you want to delete API key ${keyId}?`)) {
                await apiRequest(`/api/auth/keys/${keyId}`, {method: 'DELETE'}, deleteBtn);
                showToast(`Key ${keyId} deleted.`);
                await loadApiKeys();
            }
        });
        document.getElementById('selectAllScopes').addEventListener('click', () => {
            document.querySelectorAll('#scopes-container input[type="checkbox"]:not(:disabled):not(#scope-\\*').forEach(cb => cb.checked = true);
        });
        document.getElementById('deselectAllScopes').addEventListener('click', () => {
            document.querySelectorAll('#scopes-container input[type="checkbox"]:not(:disabled):not(#scope-\\*):checked').forEach(cb => cb.checked = false);
        });

        document.getElementById('copyKeyBtn').addEventListener('click', e => {
            const input = document.getElementById('newKeyValue');
            navigator.clipboard.writeText(input.value).then(() => {
                showToast('Copied to clipboard!');
            }).catch(() => {
                showToast('Failed to copy.', 'error');
            });
        });


        // --- Config Page ---
        document.getElementById('configEditorToggleBtn').addEventListener('click', (e) => {
            const simpleView = document.getElementById('simple-config-editor-view');
            const rawView = document.getElementById('raw-config-editor-view');
            const isSimpleViewVisible = !simpleView.classList.contains('hidden-view');

            if (isSimpleViewVisible) {
                // Switching to Raw View
                simpleView.classList.add('hidden-view');
                rawView.classList.remove('hidden-view');
                e.target.textContent = 'Switch to Simple Editor';
            } else {
                // Switching to Simple View, validation required
                const rawText = document.getElementById('appConfigText').value;
                try {
                    const parsedConfig = JSON.parse(rawText);
                    appState.dataCache.config = parsedConfig; // Update cache
                    renderConfigPage(); // Re-render simple view with new data
                    simpleView.classList.remove('hidden-view');
                    rawView.classList.add('hidden-view');
                    e.target.textContent = 'Switch to Raw JSON';
                } catch (err) {
                    showToast('Cannot switch: Raw JSON is invalid.', 'error');
                }
            }
        });
        document.getElementById('updateSimpleConfigForm').addEventListener('submit', handleUpdateSimpleConfig);

        document.getElementById('updateRawConfigForm').addEventListener('submit', handleUpdateRawConfig);

        document.getElementById('appConfigText').addEventListener('input', e => {
            const button = document.getElementById('updateConfigBtn');
            try {
                JSON.parse(e.target.value);
                e.target.classList.remove('invalid');
                button.disabled = false;
            } catch (err) {
                e.target.classList.add('invalid');
                button.disabled = true;
            }
        });


        // --- Application Page Listeners ---
        document.getElementById('restartBtn').addEventListener('click', handleRestart);
        document.getElementById('shutdownBtn').addEventListener('click', handleShutdown);
        document.getElementById('resetStatsBtn').addEventListener('click', handleResetStats);

        document.querySelectorAll('.theme-card').forEach(card => {
            card.addEventListener('click', () => {
                applyTheme(card.dataset.theme);
            });
        });

    }

    async function handleRestart(e) {
        if (confirm('Are you sure you want to restart the server?')) {
            await apiRequest('/api/server/restart', {method: 'POST'}, e.currentTarget);
            showToast('Restart command sent. The page will now reload.');
            setTimeout(() => window.location.reload(), 2000);
        }
    }

    async function handleShutdown(e) {
        if (confirm('This will shut down the server. You will need to restart it manually. Are you sure?')) {
            await apiRequest('/api/server/shutdown', {method: 'POST'}, e.currentTarget);
            showToast('Shutdown command sent.');
        }
    }

    async function handleResetStats(e) {
        if (confirm('Are you sure you want to permanently delete all collected statistics? This action cannot be undone.')) {
            await apiRequest('/api/stats/all', {method: 'DELETE'}, e.currentTarget);
            showToast('All statistics have been reset.', 'success');
        }
    }

    async function handleAddWhitelist(e) {
        e.preventDefault();
        const form = e.currentTarget;
        const type = form.id.includes('ip') ? 'ip' : 'useragent';
        const input = form.querySelector('input');
        const value = input.value.trim();
        if (!value) {
            showToast('Value cannot be empty.', 'error');
            return;
        }

        try {
            await apiRequest(`/api/whitelist/${type}`, {
                method: 'POST',
                body: JSON.stringify({value})
            }, form.querySelector('button[type="submit"]'));
            showToast(`${type === 'ip' ? 'IP' : 'User Agent'} added.`);
            input.value = '';
            await loadWhitelist(); // Refresh list
        } catch (error) {
            // Error toast is handled by apiRequest
        }
    }

    async function handleRemoveWhitelist(type, value, button) {
        try {
            await apiRequest(`/api/whitelist/${type}`, {
                method: 'DELETE',
                body: JSON.stringify({value})
            }, button);
            showToast(`${type === 'ip' ? 'IP' : 'User Agent'} removed.`);
            await loadWhitelist(); // Refresh list
        } catch (error) {
            // Error toast is handled by apiRequest
        }
    }

    function reconstructConfigFromSimpleEditor() {
        const newConfig = JSON.parse(JSON.stringify(appState.dataCache.config)); // Deep copy

        const setNestedValue = (obj, path, value) => {
            const keys = path.split('.');
            let current = obj;
            for (let i = 0; i < keys.length - 1; i++) {
                current = current[keys[i]];
            }
            current[keys[keys.length - 1]] = value;
        };

        document.querySelectorAll('#simple-config-editor-view [data-key]').forEach(el => {
            const key = el.dataset.key;

            // Handle arrays
            if (el.classList.contains('array-field-container')) {
                const values = Array.from(el.querySelectorAll('.array-item input'))
                    .map(input => input.value.trim())
                    .filter(v => v !== '');
                setNestedValue(newConfig, key, values);
                return;
            }

            // Handle other inputs
            let value;
            if (el.type === 'checkbox') {
                value = el.checked;
            } else if (el.type === 'number') {
                value = parseFloat(el.value);
            } else {
                value = el.value;
            }
            setNestedValue(newConfig, key, value);
        });

        return newConfig;
    }

    async function handleUpdateSimpleConfig(e) {
        e.preventDefault();
        const button = e.currentTarget.querySelector('button[type="submit"]');
        try {
            const newConfig = reconstructConfigFromSimpleEditor();
            await apiRequest('/api/server/config', {method: 'PUT', body: JSON.stringify(newConfig)}, button);
            appState.dataCache.config = newConfig; // Update local cache
            showToast('Configuration saved successfully.');
        } catch (error) {
            showToast('Failed to save configuration.', 'error');
        }
    }

    async function handleSaveTemplate(e) {
        e.preventDefault();
        const button = e.currentTarget.querySelector('button[type="submit"]');
        const name = document.getElementById('templateSelector').value;
        const content = document.getElementById('templateContent').value;

        if (!name || name === '--new--') {
            showToast('No template selected to save.', 'error');
            return;
        }

        try {
            await apiRequest(`/api/templates/${name}`, {method: 'PUT', body: content, headers: {'Content-Type': 'text/plain'}}, button);
            showToast(`Template "${name}" saved successfully.`);
        } catch (error) {
            // Error is shown by apiRequest
        }
    }

    async function handleCreateTemplate(e) {
        e.preventDefault();
        const button = e.currentTarget.querySelector('button[type="submit"]');
        const baseName = document.getElementById('newTemplateNameInput').value.trim();
        const extension = document.querySelector('input[name="template-extension"]:checked').value;
        const content = document.getElementById('newTemplateContent').value;

        if (!baseName) {
            showToast('Template name cannot be empty.', 'error');
            return;
        }

        const fullName = baseName.replace(/\.tmpl\.html$|\.part\.html$/, '') + extension;

        try {
            await apiRequest(`/api/templates/${fullName}`, {method: 'PUT', body: content, headers: {'Content-Type': 'text/plain'}}, button);
            showToast(`Template "${fullName}" created successfully.`);

            // Reset form
            document.getElementById('newTemplateNameInput').value = '';
            document.getElementById('newTemplateContent').value = '';

            // Refresh list and select the new template
            appState.uiState.selectedTemplate = fullName;
            await loadTemplates();
        } catch (error) {
            // Error is shown by apiRequest
        }
    }

    async function handleDeleteTemplate(e) {
        const name = document.getElementById('templateSelector').value;
        if (!name || name === '--new--') {
            showToast('No template selected to delete.', 'error');
            return;
        }

        if (confirm(`Are you sure you want to permanently delete template "${name}"?`)) {
            try {
                await apiRequest(`/api/templates/${name}`, {method: 'DELETE'}, e.currentTarget);
                showToast(`Template "${name}" deleted.`);
                appState.uiState.selectedTemplate = '';
                await loadTemplates();
            } catch (error) {
                // Error is shown by apiRequest
            }
        }
    }

    async function handleWhitelistImport(e) {
        const file = e.target.files[0];
        const type = e.target.dataset.importType;
        if (!file || !type) return;

        const reader = new FileReader();
        reader.onload = async (event) => {
            const lines = event.target.result.split('\n').map(line => line.trim()).filter(line => line.length > 0);
            if (lines.length === 0) {
                showToast('File is empty or contains no valid lines.', 'error');
                return;
            }

            let successCount = 0;
            let errorCount = 0;

            for (const line of lines) {
                try {
                    await apiRequest(`/api/whitelist/${type}`, {
                        method: 'POST',
                        body: JSON.stringify({value: line})
                    });
                    successCount++;
                } catch (error) {
                    errorCount++;
                }
            }

            showToast(`Import complete. Added: ${successCount}, Failed: ${errorCount}.`, errorCount > 0 ? 'error' : 'success');
            await loadWhitelist(); // Refresh the list
        };
        reader.readAsText(file);
        e.target.value = ''; // Reset file input
    }

    async function handleUpdateRawConfig(e) {
        e.preventDefault();
        const button = e.currentTarget.querySelector('button[type="submit"]');
        const rawConfig = document.getElementById('appConfigText').value;

        try {
            const parsedConfig = JSON.parse(rawConfig); // Validate JSON one last time
            await apiRequest('/api/server/config', {method: 'PUT', body: JSON.stringify(parsedConfig)}, button);
            appState.dataCache.config = parsedConfig; // Update local cache
            showToast('Configuration saved successfully.');
        } catch (error) {
            showToast('Failed to save configuration. Check if JSON is valid.', 'error');
        }
    }


    // --- Initial Load ---
    DOM.loginForm.addEventListener('submit', e => {
        e.preventDefault();
        attemptLogin(DOM.apiKeyInput.value);
    });

    const savedApiKey = getCookie('sarr-api-key');
    const savedTheme = getCookie('sarr-theme');
    if (savedTheme) {
        applyTheme(savedTheme);
    }
    attemptLogin(savedApiKey || '');
});