import { appState } from '../state.js';
import { apiRequest } from '../api.js';
import { showToast, setCookie } from '../utils.js';

export async function loadApiKeys(button = null) {
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

export function renderApiKeysPage() {
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

export function setupAuthEventListeners() {
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
        document.querySelectorAll('#scopes-container input[type="checkbox"]:not(:disabled):not(#scope-\*').forEach(cb => cb.checked = true);
    });
    document.getElementById('deselectAllScopes').addEventListener('click', () => {
        document.querySelectorAll('#scopes-container input[type="checkbox"]:not(:disabled):not(#scope-\*):checked').forEach(cb => cb.checked = false);
    });

    document.getElementById('copyKeyBtn').addEventListener('click', e => {
        const input = document.getElementById('newKeyValue');
        navigator.clipboard.writeText(input.value).then(() => {
            showToast('Copied to clipboard!');
        }).catch(() => {
            showToast('Failed to copy.', 'error');
        });
    });
}
