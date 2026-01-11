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

    // --- Theme Data ---

    const PRESET_THEMES_DATA = {

        'default': {

            '--bg-main': '#0f172a',
            '--bg-surface': '#1e293b',
            '--bg-surface-hover': '#334155',
            '--bg-accent': 'rgba(45, 212, 191, 0.1)',

            '--border-color': '#334155',
            '--border-color-focus': '#2dd4bf',

            '--text-primary': '#e2e8f0',
            '--text-secondary': '#94a3b8',
            '--text-disabled': '#64748b',
            '--text-accent': '#2dd4bf',

            '--accent': '#2dd4bf',
            '--accent-dark': '#14b8a6',
            '--accent-text': '#0f172a',

            '--danger': '#f43f5e',
            '--danger-dark': '#e11d48',
            '--danger-text': '#e2e8f0',

            '--warning': '#facc15',
            '--warning-dark': '#eab308',
            '--warning-text': '#1e293b'

        },

        'light': {

            '--bg-main': '#f1f5f9',
            '--bg-surface': '#ffffff',
            '--bg-surface-hover': '#f8fafc',
            '--bg-accent': 'rgba(45, 212, 191, 0.1)',

            '--border-color': '#e2e8f0',
            '--border-color-focus': '#2dd4bf',

            '--text-primary': '#0f172a',
            '--text-secondary': '#64748b',
            '--text-disabled': '#94a3b8',
            '--text-accent': '#14b8a6',

            '--accent': '#2dd4bf',
            '--accent-dark': '#14b8a6',
            '--accent-text': '#ffffff',

            '--danger': '#f43f5e',
            '--danger-dark': '#e11d48',
            '--danger-text': '#e2e8f0',

            '--warning': '#facc15',
            '--warning-dark': '#eab308',
            '--warning-text': '#1e293b'

        },

        'dracula': {

            '--bg-main': '#282a36',
            '--bg-surface': '#44475a',
            '--bg-surface-hover': '#6272a4',
            '--bg-accent': 'rgba(80, 250, 123, 0.1)',

            '--border-color': '#6272a4',
            '--border-color-focus': '#bd93f9',

            '--text-primary': '#f8f8f2',
            '--text-secondary': '#bd93f9',
            '--text-disabled': '#6272a4',
            '--text-accent': '#50fa7b',

            '--accent': '#bd93f9',
            '--accent-dark': '#ff79c6',
            '--accent-text': '#f8f8f2',

            '--danger': '#ff5555',
            '--danger-dark': '#ffb86c',
            '--danger-text': '#f8f8f2',

            '--warning': '#f1fa8c',
            '--warning-dark': '#ffb86c',
            '--warning-text': '#282a36'

        },

        'solarized': {

            '--bg-main': '#002b36',
            '--bg-surface': '#073642',
            '--bg-surface-hover': '#586e75',
            '--bg-accent': 'rgba(38, 139, 210, 0.1)',

            '--border-color': '#586e75',
            '--border-color-focus': '#268bd2',

            '--text-primary': '#839496',
            '--text-secondary': '#586e75',
            '--text-disabled': '#93a1a1',
            '--text-accent': '#268bd2',

            '--accent': '#268bd2',
            '--accent-dark': '#6c71c4',
            '--accent-text': '#002b36',

            '--danger': '#dc322f',
            '--danger-dark': '#cb4b16',
            '--danger-text': '#002b36',

            '--warning': '#b58900',
            '--warning-dark': '#cb4b16',
            '--warning-text': '#002b36'

        },

        'nord': {

            '--bg-main': '#2e3440',
            '--bg-surface': '#3b4252',
            '--bg-surface-hover': '#434c5e',
            '--bg-accent': 'rgba(129, 161, 193, 0.1)',

            '--border-color': '#4c566a',
            '--border-color-focus': '#88c0d0',

            '--text-primary': '#d8dee9',
            '--text-secondary': '#81a1c1',
            '--text-disabled': '#4c566a',
            '--text-accent': '#88c0d0',

            '--accent': '#81a1c1',
            '--accent-dark': '#8fbcbb',
            '--accent-text': '#2e3440',

            '--danger': '#bf616a',
            '--danger-dark': '#d08770',
            '--danger-text': '#2e3440',

            '--warning': '#ebcb8b',
            '--warning-dark': '#d08770',
            '--warning-text': '#2e3440'

        },

        'gruvbox': {

            '--bg-main': '#282828',
            '--bg-surface': '#3c3836',
            '--bg-surface-hover': '#504945',
            '--bg-accent': 'rgba(254, 128, 25, 0.1)',

            '--border-color': '#504945',
            '--border-color-focus': '#fe8019',

            '--text-primary': '#ebdbb2',
            '--text-secondary': '#bdae93',
            '--text-disabled': '#665c54',
            '--text-accent': '#fe8019',

            '--accent': '#fabd2f',
            '--accent-dark': '#fe8019',
            '--accent-text': '#282828',

            '--danger': '#cc241d',
            '--danger-dark': '#fb4934',
            '--danger-text': '#282828',

            '--warning': '#d79921',
            '--warning-dark': '#fabd2f',
            '--warning-text': '#282828'

        },

        'catppuccin': {

            '--bg-main': '#1e1e2e',
            '--bg-surface': '#313244',
            '--bg-surface-hover': '#45475a',
            '--bg-accent': 'rgba(137, 180, 250, 0.1)',

            '--border-color': '#45475a',
            '--border-color-focus': '#89b4fa',

            '--text-primary': '#cdd6f4',
            '--text-secondary': '#a6adc8',
            '--text-disabled': '#585b70',
            '--text-accent': '#89b4fa',

            '--accent': '#89b4fa',
            '--accent-dark': '#b4befe',
            '--accent-text': '#1e1e2e',

            '--danger': '#f38ba8',
            '--danger-dark': '#eba0ac',
            '--danger-text': '#1e1e2e',

            '--warning': '#fab387',
            '--warning-dark': '#f9e2af',
            '--warning-text': '#1e1e2e'

        },

        'monokai': {

            '--bg-main': '#272822',
            '--bg-surface': '#3e3d32',
            '--bg-surface-hover': '#75715e',
            '--bg-accent': 'rgba(253, 151, 31, 0.1)',

            '--border-color': '#75715e',
            '--border-color-focus': '#fd971f',

            '--text-primary': '#f8f8f2',
            '--text-secondary': '#a6e22e',
            '--text-disabled': '#75715e',
            '--text-accent': '#fd971f',

            '--accent': '#f92672',
            '--accent-dark': '#a6e22e',
            '--accent-text': '#272822',

            '--danger': '#f92672',
            '--danger-dark': '#fd5ff0',
            '--danger-text': '#272822',

            '--warning': '#e6db74',
            '--warning-dark': '#fd971f',
            '--warning-text': '#272822'

        },

        'one-dark': {

            '--bg-main': '#282c34',
            '--bg-surface': '#21252b',
            '--bg-surface-hover': '#3a3f4b',
            '--bg-accent': 'rgba(97, 175, 239, 0.1)',

            '--border-color': '#3a3f4b',
            '--border-color-focus': '#61afef',

            '--text-primary': '#abb2bf',
            '--text-secondary': '#828997',
            '--text-disabled': '#5c6370',
            '--text-accent': '#61afef',

            '--accent': '#c678dd',
            '--accent-dark': '#e06c75',
            '--accent-text': '#282c34',

            '--danger': '#e06c75',
            '--danger-dark': '#be5046',
            '--danger-text': '#282c34',

            '--warning': '#d19a66',
            '--warning-dark': '#e5c07b',
            '--warning-text': '#282c34'

        },

        'tokyo-night': {

            '--bg-main': '#1a1b26',
            '--bg-surface': '#24283b',
            '--bg-surface-hover': '#414868',
            '--bg-accent': 'rgba(122, 162, 247, 0.1)',

            '--border-color': '#414868',
            '--border-color-focus': '#7aa2f7',

            '--text-primary': '#a9b1d6',
            '--text-secondary': '#787c99',
            '--text-disabled': '#414868',
            '--text-accent': '#7aa2f7',

            '--accent': '#bb9af7',
            '--accent-dark': '#f7768e',
            '--accent-text': '#1a1b26',

            '--danger': '#f7768e',
            '--danger-dark': '#ff9e64',
            '--danger-text': '#1a1b26',

            '--warning': '#e0af68',
            '--warning-dark': '#ff9e64',
            '--warning-text': '#1a1b26'

        },

        'synthwave': {

            '--bg-main': '#2b213a',
            '--bg-surface': '#34294f',
            '--bg-surface-hover': '#4e3a71',
            '--bg-accent': 'rgba(255, 121, 198, 0.1)',

            '--border-color': '#4e3a71',
            '--border-color-focus': '#ff79c6',

            '--text-primary': '#f8f8f2',
            '--text-secondary': '#bd93f9',
            '--text-disabled': '#6272a4',
            '--text-accent': '#ff79c6',

            '--accent': '#f1fa8c',
            '--accent-dark': '#50fa7b',
            '--accent-text': '#2b213a',

            '--danger': '#ff5555',
            '--danger-dark': '#ffb86c',
            '--danger-text': '#2b213a',

            '--warning': '#ffb86c',
            '--warning-dark': '#f1fa8c',
            '--warning-text': '#2b213a'

        },

        'solarized-light': {

            '--bg-main': '#fdf6e3',
            '--bg-surface': '#eee8d5',
            '--bg-surface-hover': '#93a1a1',
            '--bg-accent': 'rgba(38, 139, 210, 0.1)',

            '--border-color': '#93a1a1',
            '--border-color-focus': '#268bd2',

            '--text-primary': '#657b83',
            '--text-secondary': '#839496',
            '--text-disabled': '#586e75',
            '--text-accent': '#268bd2',

            '--accent': '#b58900',
            '--accent-dark': '#cb4b16',
            '--accent-text': '#fdf6e3',

            '--danger': '#dc322f',
            '--danger-dark': '#cb4b16',
            '--danger-text': '#fdf6e3',

            '--warning': '#b58900',
            '--warning-dark': '#cb4b16',
            '--warning-text': '#fdf6e3'

        },

        'github-light': {

            '--bg-main': '#ffffff',
            '--bg-surface': '#f6f8fa',
            '--bg-surface-hover': '#e1e4e8',
            '--bg-accent': 'rgba(3, 102, 214, 0.1)',

            '--border-color': '#e1e4e8',
            '--border-color-focus': '#0366d6',

            '--text-primary': '#24292e',
            '--text-secondary': '#586069',
            '--text-disabled': '#959da5',
            '--text-accent': '#0366d6',

            '--accent': '#0366d6',
            '--accent-dark': '#005cc5',
            '--accent-text': '#ffffff',

            '--danger': '#d73a49',
            '--danger-dark': '#cb2431',
            '--danger-text': '#ffffff',

            '--warning': '#f66a0a',
            '--warning-dark': '#e36209',
            '--warning-text': '#ffffff'

        },

        'vscode': {

            '--bg-main': '#1e1e1e',
            '--bg-surface': '#252526',
            '--bg-surface-hover': '#333333',
            '--bg-accent': 'rgba(0, 122, 204, 0.1)',

            '--border-color': '#333333',
            '--border-color-focus': '#007acc',

            '--text-primary': '#cccccc',
            '--text-secondary': '#808080',
            '--text-disabled': '#555555',
            '--text-accent': '#007acc',

            '--accent': '#007acc',
            '--accent-dark': '#005f9e',
            '--accent-text': '#ffffff',

            '--danger': '#f44747',
            '--danger-dark': '#d13438',
            '--danger-text': '#ffffff',

            '--warning': '#f8a529',
            '--warning-dark': '#e69100',
            '--warning-text': '#1e1e1e'

        },

        'cyberpunk': {

            '--bg-main': '#0c0c0c',
            '--bg-surface': '#1a1a1a',
            '--bg-surface-hover': '#2a2a2a',
            '--bg-accent': 'rgba(249, 248, 0, 0.1)',

            '--border-color': '#2a2a2a',
            '--border-color-focus': '#f9f800',

            '--text-primary': '#ffffff',
            '--text-secondary': '#b3b3b3',
            '--text-disabled': '#555555',
            '--text-accent': '#f9f800',

            '--accent': '#00f0ff',
            '--accent-dark': '#f9f800',
            '--accent-text': '#0c0c0c',

            '--danger': '#ff0055',
            '--danger-dark': '#ff00a9',
            '--danger-text': '#0c0c0c',

            '--warning': '#f9f800',
            '--warning-dark': '#ff0055',
            '--warning-text': '#0c0c0c'

        },

        'sky': {

            '--bg-main': '#f0f9ff',
            '--bg-surface': '#e0f2fe',
            '--bg-surface-hover': '#bae6fd',
            '--bg-accent': 'rgba(56, 189, 248, 0.1)',

            '--border-color': '#bae6fd',
            '--border-color-focus': '#38bdf8',

            '--text-primary': '#0c4a6e',
            '--text-secondary': '#38bdf8',
            '--text-disabled': '#7dd3fc',
            '--text-accent': '#0ea5e9',

            '--accent': '#0ea5e9',
            '--accent-dark': '#0284c7',
            '--accent-text': '#ffffff',

            '--danger': '#f43f5e',
            '--danger-dark': '#e11d48',
            '--danger-text': '#ffffff',

            '--warning': '#f59e0b',
            '--warning-dark': '#d97706',
            '--warning-text': '#ffffff'

        }

    };


    // --- Theme System ---

    const ThemeManager = {

        vars: [

            '--bg-main', '--bg-surface', '--bg-surface-hover', '--bg-accent',

            '--border-color', '--border-color-focus',

            '--text-primary', '--text-secondary', '--text-disabled', '--text-accent',

            '--accent', '--accent-dark', '--accent-text',

            '--danger', '--danger-dark', '--danger-text',

            '--warning', '--warning-dark', '--warning-text'

        ],

        presets: {},

        saved: JSON.parse(localStorage.getItem('sarr-saved-themes') || '{}'),


        init() {

            this.loadPresets();

            this.renderLists();

            this.renderEditor();


            const activeThemeId = getCookie('sarr-theme') || 'default';


            if (this.presets[activeThemeId]) {

                this.apply(this.presets[activeThemeId], activeThemeId);

            } else if (this.saved[activeThemeId]) {

                this.apply(this.saved[activeThemeId], activeThemeId);

            } else if (activeThemeId === 'custom') {

                const temp = JSON.parse(localStorage.getItem('sarr-custom-theme') || '{}');

                if (Object.keys(temp).length > 0) this.apply(temp, 'custom');

            } else {

                this.apply(this.presets['default'], 'default');

            }

        },


        loadPresets() {
            for (const [id, data] of Object.entries(PRESET_THEMES_DATA)) {
                // Merge with default values if any missing, but our data is complete
                this.presets[id] = {
                    id: id,
                    name: id
                        .replace(/([A-Z])/g, ' $1') // Add space before capital letters
                        .replace(/-/g, ' ')         // Replace hyphens with spaces
                        .trim()
                        .replace(/\b\w/g, c => c.toUpperCase()), // Capitalize first letter of each word
                    ...data
                };
            }
        },


        apply(themeObj, id) {

            this.vars.forEach(v => {

                const val = themeObj[v] || PRESET_THEMES_DATA['default'][v]; // Fallback to default

                document.documentElement.style.setProperty(v, val);

            });


            if (id) {

                setCookie('sarr-theme', id, 365);

                // We no longer need data-theme for CSS since we apply variables directly

                // but we keep it for potential external CSS hooks or just state inspection

                if (this.presets[id]) {

                    document.documentElement.setAttribute('data-theme', id);

                } else {

                    document.documentElement.removeAttribute('data-theme');

                }

            }

        },

        createPreviewCard(themeObj, isSaved = false) {
            const card = document.createElement('div');
            card.className = 'card theme-card';
            card.style.padding = '1rem'; // Add padding to inset all content
            card.onclick = (e) => {
                if (e.target.closest('button')) return;
                this.apply(themeObj, themeObj.id);
                showToast(`Applied theme: ${themeObj.name || themeObj.id}`);
            };

            // Extract colors
            const bgMain = themeObj['--bg-main'];
            const bgSurface = themeObj['--bg-surface'];
            const textPrimary = themeObj['--text-primary'];
            const textSecondary = themeObj['--text-secondary'];
            const accent = themeObj['--accent'];
            const border = themeObj['--border-color'];
            const danger = themeObj['--danger'];
            const warning = themeObj['--warning'];

            let actionButtons = '';
            if (isSaved) {
                actionButtons = `
                                <div style="display:flex; gap:0.5rem; margin-top:auto; padding-top: 0.75rem; border-top: 1px solid ${border};">
                                    <button class="secondary edit-theme-btn" style="flex:1; font-size:0.8rem; padding:0.4rem;">Edit</button>
                                    <button class="danger delete-theme-btn" style="width:32px; padding:0.4rem;">&times;</button>
                                </div>
                            `;
            } else {
                actionButtons = `
                                <div style="display:flex; gap:0.5rem; margin-top:auto; padding-top: 0.75rem; border-top: 1px solid ${border};">
                                    <button class="secondary edit-theme-btn" style="flex:1; font-size:0.8rem; padding:0.4rem;">Load in Editor</button>
                                </div>
                            `;
            }

            // Mini Mockup HTML
            const mockupHtml = `
                            <div style="background: ${bgMain}; border: 1px solid ${border}; border-radius: 6px; padding: 10px; font-family: sans-serif; height: 120px; overflow: hidden; position: relative; margin-bottom: 0.75rem;">
                                <div style="background: ${bgSurface}; border: 1px solid ${border}; border-radius: 4px; padding: 8px; margin-bottom: 8px;">
                                    <div style="background: ${textPrimary}; height: 6px; width: 60%; border-radius: 3px; margin-bottom: 6px; opacity: 0.9;"></div>
                                    <div style="background: ${textSecondary}; height: 4px; width: 90%; border-radius: 2px; margin-bottom: 4px; opacity: 0.7;"></div>
                                    <div style="background: ${textSecondary}; height: 4px; width: 40%; border-radius: 2px; opacity: 0.7;"></div>
                                </div>
                                <div style="display: flex; gap: 4px;">
                                    <div style="background: ${accent}; height: 16px; flex: 1; border-radius: 3px;"></div>
                                    <div style="background: ${danger}; height: 16px; width: 16px; border-radius: 3px;"></div>
                                    <div style="background: ${warning}; height: 16px; width: 16px; border-radius: 3px;"></div>
                                </div>
                                <div style="position: absolute; bottom: 8px; right: 10px; width: 20px; height: 20px; background: ${bgSurface}; border: 1px solid ${border}; border-radius: 50%; display: flex; align-items: center; justify-content: center;">
                                    <div style="background: ${accent}; width: 8px; height: 8px; border-radius: 50%;"></div>
                                </div>
                            </div>
                        `;

            card.innerHTML = `
                            <div style="display:flex; justify-content:space-between; align-items:center; margin-bottom:0.75rem;">
                                <h3 style="font-size:1rem; margin:0;">${themeObj.name || themeObj.id}</h3>
                                <button class="icon-only secondary copy-theme-btn" title="Copy JSON" style="width:24px; height:24px; padding:2px;">
                                    <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" style="width:14px; height:14px;"><path stroke-linecap="round" stroke-linejoin="round" d="M15.666 3.888A2.25 2.25 0 0 0 13.5 2.25h-3c-1.03 0-1.9.693-2.166 1.638m7.332 0c.055.194.084.4.084.612v0a.75.75 0 0 1-.75.75H9a.75.75 0 0 1-.75-.75v0c0-.212.03-.418.084-.612m7.332 0c.646.049 1.288.11 1.927.184 1.1.128 1.907 1.077 1.907 2.185V19.5a2.25 2.25 0 0 1-2.25 2.25H6.75A2.25 2.25 0 0 1 4.5 19.5V6.257c0-1.108.806-2.057 1.907-2.185a48.208 48.208 0 0 1 1.927-.184" /></svg>
                                </button>
                            </div>
                            ${mockupHtml}
                            ${actionButtons}
                        `;

            // Bind events
            card.querySelector('.copy-theme-btn').onclick = (e) => {
                e.stopPropagation();
                const cleanObj = {...themeObj};
                delete cleanObj.id;
                delete cleanObj.name; // Don't copy name if redundant
                navigator.clipboard.writeText(JSON.stringify(cleanObj, null, 2)).then(() => showToast('Theme JSON copied!'));
            };

            if (isSaved) {
                card.querySelector('.delete-theme-btn').onclick = (e) => {
                    e.stopPropagation();
                    if (confirm(`Delete theme "${themeObj.name}"?`)) {
                        this.deleteSaved(themeObj.id);
                    }
                };
            }

            card.querySelector('.edit-theme-btn').onclick = (e) => {
                e.stopPropagation();
                this.loadIntoEditor(themeObj);
                document.querySelector('[data-target-panel="theme-editor-panel"]').click();
                showToast(`Loaded "${themeObj.name}" into editor.`);
            };

            return card;
        },
        renderLists() {
            const presetGrid = document.getElementById('preset-themes-grid');
            const savedGrid = document.getElementById('saved-themes-grid');
            if (!presetGrid || !savedGrid) return; // Might not be rendered yet

            presetGrid.innerHTML = '';
            savedGrid.innerHTML = '';

            Object.keys(this.presets).forEach(id => {
                // If we haven't hydrated yet (e.g. dynamic elements not ready), skip or hydrate
                if (this.presets[id]) presetGrid.appendChild(this.createPreviewCard(this.presets[id], false));
            });

            const savedIds = Object.keys(this.saved);
            if (savedIds.length === 0) {
                savedGrid.innerHTML = '<p class="list-empty-state" style="grid-column: 1/-1; text-align: left; padding: 0; color: var(--text-secondary);">No custom themes saved yet.</p>';
            } else {
                savedIds.forEach(id => {
                    savedGrid.appendChild(this.createPreviewCard(this.saved[id], true));
                });
            }
        },

        renderEditor() {
            const container = document.getElementById('custom-theme-editor-grid');
            if (!container) return;
            container.innerHTML = '';

            const computed = getComputedStyle(document.documentElement);

            this.vars.forEach(v => {
                const val = computed.getPropertyValue(v).trim();
                const wrapper = document.createElement('div');
                wrapper.style.display = 'flex';
                wrapper.style.flexDirection = 'column';
                wrapper.style.gap = '0.5rem';

                const label = document.createElement('label');
                label.textContent = v.replace('--', '').replace(/-/g, ' ');
                label.style.textTransform = 'capitalize';
                label.style.fontSize = '0.85rem';

                const inputGroup = document.createElement('div');
                inputGroup.style.display = 'flex';
                inputGroup.style.gap = '0.5rem';

                const colorInput = document.createElement('input');
                colorInput.type = 'color';
                colorInput.style.width = '40px';
                colorInput.style.padding = '0';
                colorInput.style.height = '38px';

                const textInput = document.createElement('input');
                textInput.type = 'text';
                textInput.className = 'theme-var-input';
                textInput.dataset.var = v;
                textInput.value = val;

                if (val.match(/^#[0-9A-F]{6}$/i)) colorInput.value = val;

                colorInput.addEventListener('input', e => {
                    textInput.value = e.target.value;
                    document.documentElement.style.setProperty(v, e.target.value);
                });

                textInput.addEventListener('input', e => {
                    if (e.target.value.match(/^#[0-9A-F]{6}$/i)) colorInput.value = e.target.value;
                    document.documentElement.style.setProperty(v, e.target.value);
                });

                inputGroup.appendChild(colorInput);
                inputGroup.appendChild(textInput);
                wrapper.appendChild(label);
                wrapper.appendChild(inputGroup);
                container.appendChild(wrapper);
            });
        },

        loadIntoEditor(themeObj) {
            document.getElementById('customThemeName').value = themeObj.name || '';
            document.querySelectorAll('.theme-var-input').forEach(input => {
                const v = input.dataset.var;
                if (themeObj[v]) {
                    input.value = themeObj[v];
                    input.dispatchEvent(new Event('input'));
                }
            });
        },

        saveCurrent(name) {
            if (!name) return alert('Please enter a theme name.');
            const id = 'custom-' + Date.now();
            const themeObj = {id, name};
            document.querySelectorAll('.theme-var-input').forEach(input => {
                themeObj[input.dataset.var] = input.value;
            });

            this.saved[id] = themeObj;
            localStorage.setItem('sarr-saved-themes', JSON.stringify(this.saved));
            this.renderLists();

            document.querySelector('[data-target-panel="themes-list-panel"]').click();
            showToast(`Theme "${name}" saved.`);
        },

        deleteSaved(id) {
            delete this.saved[id];
            localStorage.setItem('sarr-saved-themes', JSON.stringify(this.saved));
            this.renderLists();
            showToast('Theme deleted.');
        }
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

    function formatCompactNumber(number) {
        if (number === undefined || number === null) return '0';
        return new Intl.NumberFormat(undefined, {
            notation: "compact",
            compactDisplay: "short",
            maximumFractionDigits: 1
        }).format(number);
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
                return null;
            }

            // Check for file downloads FIRST, before attempting to parse JSON.
            const contentDisposition = response.headers.get("content-disposition");
            if (contentDisposition && contentDisposition.includes("attachment")) {
                return response; // It's a file download, return the raw response for the caller to handle.
            }

            const contentType = response.headers.get("content-type");
            if (contentType && contentType.includes("application/json")) {
                return response.json();
            }
            return response;

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
            startPolling();
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

        const totalHits = stats.summary.total_requests || 0;
        const uniqueIps = stats.summary.unique_ips || 0;
        const uniqueAgents = stats.summary.unique_user_agents || 0;

        const totalHitsEl = document.getElementById('total-hits');
        totalHitsEl.textContent = formatCompactNumber(totalHits);
        totalHitsEl.title = totalHits.toLocaleString();

        const uniqueIpsEl = document.getElementById('unique-ips');
        uniqueIpsEl.textContent = formatCompactNumber(uniqueIps);
        uniqueIpsEl.title = uniqueIps.toLocaleString();

        const uniqueAgentsEl = document.getElementById('unique-agents');
        uniqueAgentsEl.textContent = formatCompactNumber(uniqueAgents);
        uniqueAgentsEl.title = uniqueAgents.toLocaleString();

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

        const calculateHitRate = (item) => {
            if (!item.first_seen || !item.last_seen || item.total_hits <= 1) {
                return 0;
            }
            const firstSeen = new Date(item.first_seen).getTime();
            const lastSeen = new Date(item.last_seen).getTime();
            const durationMs = lastSeen - firstSeen;
            if (durationMs < 1000) { // If duration is less than a second, treat as instantaneous
                return Infinity;
            }
            const durationMinutes = durationMs / 60000;
            return (item.total_hits - 1) / durationMinutes;
        };

        const sortedData = [...data].sort((a, b) => {
            if (state.sort.dir === 'none') return 0;

            let valA, valB;
            if (state.sort.key === 'hit_rate') {
                valA = calculateHitRate(a);
                valB = calculateHitRate(b);
            } else {
                valA = a[state.sort.key];
                valB = b[state.sort.key];
            }

            const direction = state.sort.dir === 'asc' ? 1 : -1;
            if (valA < valB) return -1 * direction;
            if (valA > valB) return 1 * direction;
            return 0;
        });

        const filteredData = state.filter
            ? sortedData.filter(item => (item.ip_address || item.user_agent || '').toLowerCase().includes(state.filter.toLowerCase()))
            : sortedData;

        const dataToShow = state.showAll ? filteredData : filteredData.slice(0, 5);

        if (dataToShow.length === 0) {
            tbody.innerHTML = `<tr class="empty-row"><td colspan="5">No data available.</td></tr>`;
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
            const hitRate = calculateHitRate(item);
            const hitRateDisplay = isFinite(hitRate) ? `${hitRate.toFixed(2)}/min` : 'N/A';

            return `
                <tr>
                    <td data-label="Status"><div style="display: flex; align-items: center; gap: 8px;"><span class="recency-dot ${recency.class}"></span>${recency.title}</div></td>
                    <td data-label="${item.ip_address ? 'IP Address' : 'User Agent'}" title="${identity}">
                        <div style="display: flex; align-items: center; gap: 0.5rem; justify-content: space-between;">
                            <span style="overflow: hidden; text-overflow: ellipsis;">${identity}</span>
                            <button type="button" class="icon-only secondary copy-btn" data-copy="${identity}" title="Copy" style="width: 24px; height: 24px; padding: 2px;">
                                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" style="width: 14px; height: 14px;">
                                    <path stroke-linecap="round" stroke-linejoin="round" d="M15.666 3.888A2.25 2.25 0 0 0 13.5 2.25h-3c-1.03 0-1.9.693-2.166 1.638m7.332 0c.055.194.084.4.084.612v0a.75.75 0 0 1-.75.75H9a.75.75 0 0 1-.75-.75v0c0-.212.03-.418.084-.612m7.332 0c.646.049 1.288.11 1.927.184 1.1.128 1.907 1.077 1.907 2.185V19.5a2.25 2.25 0 0 1-2.25 2.25H6.75A2.25 2.25 0 0 1 4.5 19.5V6.257c0-1.108.806-2.057 1.907-2.185a48.208 48.208 0 0 1 1.927-.184" />
                                </svg>
                            </button>
                        </div>
                    </td>
                    <td data-label="Hits" class="text-right">${item.total_hits.toLocaleString()}</td>
                    <td data-label="Hit Rate" class="text-right">${hitRateDisplay}</td>
                    <td data-label="Last Seen">${new Date(item.last_seen).toLocaleString()}</td>
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

    function renderTrainingStatus(status) {
        const indicator = document.getElementById('training-status-indicator');
        if (!indicator) return;

        if (status && status.is_training) {
            indicator.querySelector('span').textContent = `Training ${status.model_name}...`;
            indicator.classList.remove('status-indicator-hidden');
        } else {
            indicator.classList.add('status-indicator-hidden');
        }
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
                <td data-label="ID">${key.id}</td>
                <td data-label="Description">${key.description}</td>
                <td data-label="Scopes" class="scopes-cell">${key.scopes.join(', ')}</td>
                <td data-label="Actions">
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

            // Special handling for "headers" to treat it as a Map (dynamic key-value pairs)
            if (key === 'headers' && typeof value === 'object' && value !== null && !Array.isArray(value)) {
                field.innerHTML = `<label class="config-field-label">${key}</label>`;

                const mapContainer = document.createElement('div');
                mapContainer.className = 'array-field-container map-field-container'; // Reuse array styling but add map marker
                mapContainer.id = `config-${fullKey}`;
                mapContainer.dataset.key = fullKey;
                mapContainer.dataset.type = 'map';

                const createMapItem = (k = "", v = "") => {
                    const item = document.createElement('div');
                    item.className = 'array-item'; // Reuse array-item styling

                    const keyInput = document.createElement('input');
                    keyInput.type = 'text';
                    keyInput.placeholder = 'Header Name';
                    keyInput.value = k;
                    keyInput.className = 'map-key-input';
                    keyInput.style.flex = '1';

                    const valInput = document.createElement('input');
                    valInput.type = 'text';
                    valInput.placeholder = 'Value';
                    valInput.value = v;
                    valInput.className = 'map-val-input';
                    valInput.style.flex = '2';

                    const removeBtn = document.createElement('button');
                    removeBtn.type = 'button';
                    removeBtn.className = 'remove-item-btn danger array-item-remove-btn';
                    removeBtn.innerHTML = '<svg class="btn-icon" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" /></svg>';
                    removeBtn.onclick = () => item.remove();

                    item.appendChild(keyInput);
                    item.appendChild(valInput);
                    item.appendChild(removeBtn);
                    return item;
                };

                // Add existing items
                for (const [hKey, hVal] of Object.entries(value)) {
                    mapContainer.appendChild(createMapItem(hKey, hVal));
                }

                const addBtn = document.createElement('button');
                addBtn.type = 'button';
                addBtn.className = 'secondary array-item-add-btn';
                addBtn.textContent = 'Add Header';
                addBtn.onclick = () => mapContainer.appendChild(createMapItem());

                field.appendChild(mapContainer);
                field.appendChild(addBtn);

            } else {
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
            await previewTemplate(selected, document.getElementById('previewThreat').value);
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
        document.getElementById('model-testing-card').classList.toggle('hidden-view', !isEdit);
    }

    function applyTheme(theme) {
        document.documentElement.setAttribute('data-theme', theme);
        setCookie('sarr-theme', theme, 365);

        if (theme === 'custom') {
            const customJson = localStorage.getItem('sarr-custom-theme');
            if (customJson) {
                const themeObj = JSON.parse(customJson);
                for (const [key, val] of Object.entries(themeObj)) {
                    document.documentElement.style.setProperty(key, val);
                }
            }
        } else {
            // Remove inline overrides to let CSS variables take over
            const themeVars = [
                '--bg-main', '--bg-surface', '--bg-surface-hover', '--bg-accent',
                '--border-color', '--border-color-focus',
                '--text-primary', '--text-secondary', '--text-disabled', '--text-accent',
                '--accent', '--accent-dark', '--accent-text',
                '--danger', '--danger-dark', '--danger-text',
                '--warning', '--warning-dark', '--warning-text'
            ];
            themeVars.forEach(v => document.documentElement.style.removeProperty(v));
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


        // --- Stats Page ---
        document.getElementById('refreshStatsBtn').addEventListener('click', (e) => loadStats(e.currentTarget));

        // Search Inputs
        document.querySelectorAll('.table-search-input').forEach(input => {
            input.addEventListener('input', e => {
                const tableKey = e.target.dataset.table === 'ips' ? 'ipsTable' : 'agentsTable';
                appState.uiState[tableKey].filter = e.target.value;
                renderStatsPage(); // Re-render with filter
            });
        });

        // Copy Buttons (Delegated)
        document.querySelectorAll('.table-container').forEach(container => {
            container.addEventListener('click', e => {
                const btn = e.target.closest('.copy-btn');
                if (!btn) return;
                const text = btn.dataset.copy;
                navigator.clipboard.writeText(text).then(() => {
                    showToast('Copied to clipboard!');
                }).catch(() => showToast('Failed to copy', 'error'));
            });
        });

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
        document.getElementById('generateTextForm').addEventListener('submit', async e => {
            e.preventDefault();
            const button = e.currentTarget.querySelector('button[type="submit"]');
            const modelName = document.getElementById('modelSelector').value;
            const resultArea = document.getElementById('generateResult');

            const payload = {
                maxLength: parseInt(document.getElementById('generateMaxLength').value),
                temperature: parseFloat(document.getElementById('generateTemp').value),
                topK: parseInt(document.getElementById('generateTopK').value),
                startText: document.getElementById('generateStartText').value.trim(),
            };

            resultArea.value = "Generating...";

            try {
                const result = await apiRequest(`/api/markov/models/${modelName}/generate`, {
                    method: 'POST',
                    body: JSON.stringify(payload)
                }, button);
                resultArea.value = result.text;
            } catch (error) {
                resultArea.value = `Error: ${error.message}`;
            }
        });
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
            showToast(`Training job for "${modelName}" has started. Status will update automatically.`);
            startPolling();
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

            if (newKey.id === 1) {
                appState.apiKey = newKey.raw_key;
                setCookie('sarr-api-key', newKey.raw_key, 365);
                showToast('Master key created. You have been automatically logged in.');
            }

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

        // --- Theme Event Listeners ---
        ThemeManager.renderLists();
        ThemeManager.renderEditor();

        document.getElementById('themes-tab-nav').addEventListener('click', e => {
            if (!e.target.matches('.tab-btn')) return;
            const targetPanel = e.target.dataset.targetPanel;
            document.querySelectorAll('#themes-tab-nav .tab-btn').forEach(btn => btn.classList.remove('active'));
            document.querySelectorAll('#themes-content .tab-panel').forEach(panel => panel.classList.remove('active'));
            e.target.classList.add('active');
            document.getElementById(targetPanel).classList.add('active');
        });

        document.getElementById('saveCustomThemeBtn').addEventListener('click', () => {
            ThemeManager.saveCurrent(document.getElementById('customThemeName').value.trim());
        });

        document.getElementById('applyEditorThemeBtn').addEventListener('click', () => {
            const themeObj = {};
            document.querySelectorAll('.theme-var-input').forEach(input => {
                themeObj[input.dataset.var] = input.value;
            });
            ThemeManager.apply(themeObj, 'custom');
            showToast('Theme applied (unsaved).');
        });

        document.getElementById('exportThemeBtn').addEventListener('click', () => {
            const themeObj = {};
            document.querySelectorAll('.theme-var-input').forEach(input => {
                themeObj[input.dataset.var] = input.value;
            });
            navigator.clipboard.writeText(JSON.stringify(themeObj, null, 2)).then(() => showToast('JSON copied to clipboard.'));
        });

        document.getElementById('importThemeBtn').addEventListener('click', async () => {
            try {
                const text = await navigator.clipboard.readText();
                const json = JSON.parse(text);
                ThemeManager.loadIntoEditor(json);
                showToast('Theme imported into editor.');
            } catch (e) {
                showToast('Failed to import JSON.', 'error');
            }
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

            // Handle maps (e.g., headers)
            if (el.dataset.type === 'map') {
                const mapObj = {};
                el.querySelectorAll('.array-item').forEach(item => {
                    const k = item.querySelector('.map-key-input').value.trim();
                    const v = item.querySelector('.map-val-input').value.trim();
                    if (k) mapObj[k] = v;
                });
                setNestedValue(newConfig, key, mapObj);
                return;
            }

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
            await apiRequest(`/api/templates/${name}`, {
                method: 'PUT',
                body: content,
                headers: {'Content-Type': 'text/plain'}
            }, button);
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
            await apiRequest(`/api/templates/${fullName}`, {
                method: 'PUT',
                body: content,
                headers: {'Content-Type': 'text/plain'}
            }, button);
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

    function startPolling() {
        if (appState.timers.trainingStatusPoll) return;
        pollTrainingStatus();
        appState.timers.trainingStatusPoll = setInterval(pollTrainingStatus, 5000);
    }

    function stopPolling() {
        if (appState.timers.trainingStatusPoll) {
            clearInterval(appState.timers.trainingStatusPoll);
            appState.timers.trainingStatusPoll = null;
        }
    }

    async function pollTrainingStatus() {
        try {
            const status = await apiRequest('/api/markov/training/status', {}, null);
            renderTrainingStatus(status);
            if (status && !status.is_training) {
                stopPolling();
            }
        } catch (error) {
            renderTrainingStatus({is_training: false});
            stopPolling();
        }
    }

    // --- Initial Load ---
    DOM.loginForm.addEventListener('submit', e => {
        e.preventDefault();
        attemptLogin(DOM.apiKeyInput.value);
    });

    const savedApiKey = getCookie('sarr-api-key');
    const savedTheme = getCookie('sarr-theme');
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