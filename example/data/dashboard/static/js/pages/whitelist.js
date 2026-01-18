import { appState } from '../state.js';
import { apiRequest } from '../api.js';
import { showToast } from '../utils.js';

export async function loadWhitelist(button = null) {
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

export function renderWhitelistPage() {
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

export function setupWhitelistEventListeners() {
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
}
