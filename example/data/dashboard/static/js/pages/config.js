import { appState } from '../state.js';
import { apiRequest } from '../api.js';
import { showToast } from '../utils.js';

export async function loadConfig(button = null) {
    try {
        const config = await apiRequest('/api/server/config', {}, button);
        appState.dataCache.config = config;
        renderConfigPage();
    } catch (error) {
        // Handle config load error
    }
}

export function renderConfigPage() {
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

export function setupConfigEventListeners() {
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
}
