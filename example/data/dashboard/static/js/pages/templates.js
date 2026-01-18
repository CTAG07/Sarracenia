import { appState } from '../state.js';
import { apiRequest } from '../api.js';
import { debounce, showToast } from '../utils.js';

export async function loadTemplates(button = null) {
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

export function renderTemplatesPage() {
    const selector = document.getElementById('templateSelector');
    selector.innerHTML = `<option value="">-- Select a Template --</option><option value="--new--">-- Create New Template --</option>`;
    appState.dataCache.templates.forEach(name => {
        selector.add(new Option(name, name));
    });
    selector.value = appState.uiState.selectedTemplate || '';
    handleTemplateSelection();
}

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

export function setupTemplatesEventListeners() {
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
}
