import { appState, DOM } from './state.js';
import { showToast, toggleButtonLoading, eraseCookie } from './utils.js';

// --- API Wrapper ---
export async function apiRequest(endpoint, options = {}, button = null) {
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

export function logout() {
    eraseCookie('sarr-api-key');
    appState.apiKey = null;
    window.location.reload();
}
