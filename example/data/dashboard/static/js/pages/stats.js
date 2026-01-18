import { appState } from '../state.js';
import { apiRequest } from '../api.js';
import { formatCompactNumber, showToast } from '../utils.js';

export async function loadStats(button = null) {
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

export function renderStatsPage() {
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

export function setupStatsEventListeners() {
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
}
