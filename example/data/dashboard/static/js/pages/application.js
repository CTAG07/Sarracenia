import { apiRequest } from '../api.js';
import { showToast } from '../utils.js';

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

export function setupApplicationEventListeners() {
    document.getElementById('restartBtn').addEventListener('click', handleRestart);
    document.getElementById('shutdownBtn').addEventListener('click', handleShutdown);
    document.getElementById('resetStatsBtn').addEventListener('click', handleResetStats);
}
