import { appState } from '../state.js';
import { apiRequest } from '../api.js';
import { showToast, triggerDownload } from '../utils.js';

export async function loadMarkovModels(button = null) {
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

export function renderMarkovPage() {
    const selector = document.getElementById('modelSelector');
    selector.innerHTML = `<option value="">-- Select a Model --</option><option value="--new--">-- Create New Model --</option>`;
    appState.dataCache.models.forEach(model => {
        selector.add(new Option(`${model.Name} (Order: ${model.Order})`, model.Name));
    });
    selector.value = appState.uiState.selectedModel || '';
    handleModelSelection();
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

export function setupMarkovEventListeners() {
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
}
