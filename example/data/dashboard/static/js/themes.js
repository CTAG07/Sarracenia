import { getCookie, setCookie, showToast } from './utils.js';

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
export const ThemeManager = {

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

        // Event Listeners for Theme Manager
        document.getElementById('themes-tab-nav').addEventListener('click', e => {
            if (!e.target.matches('.tab-btn')) return;
            const targetPanel = e.target.dataset.targetPanel;
            document.querySelectorAll('#themes-tab-nav .tab-btn').forEach(btn => btn.classList.remove('active'));
            document.querySelectorAll('#themes-content .tab-panel').forEach(panel => panel.classList.remove('active'));
            e.target.classList.add('active');
            document.getElementById(targetPanel).classList.add('active');
        });

        document.getElementById('saveCustomThemeBtn').addEventListener('click', () => {
            this.saveCurrent(document.getElementById('customThemeName').value.trim());
        });

        document.getElementById('applyEditorThemeBtn').addEventListener('click', () => {
            const themeObj = {};
            document.querySelectorAll('.theme-var-input').forEach(input => {
                themeObj[input.dataset.var] = input.value;
            });
            this.apply(themeObj, 'custom');
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
                this.loadIntoEditor(json);
                showToast('Theme imported into editor.');
            } catch (e) {
                showToast('Failed to import JSON.', 'error');
            }
        });

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
                document.documentElement.setAttribute('data-theme', 'custom');

            }
            
            if (id === 'custom') {
                 localStorage.setItem('sarr-custom-theme', JSON.stringify(themeObj));
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
