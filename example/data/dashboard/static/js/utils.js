// --- Utility Functions ---

export function debounce(func, delay) {
    let timeout;
    return function (...args) {
        const context = this;
        clearTimeout(timeout);
        timeout = setTimeout(() => func.apply(context, args), delay);
    };
}

export function showToast(message, type = 'success', duration = 3000) {
    const toast = document.getElementById('toast');
    if (!toast) return;
    toast.textContent = message;
    toast.className = `toast show ${type}`;
    setTimeout(() => {
        toast.className = 'toast';
    }, duration);
}

export function setCookie(name, value, days) {
    let expires = "";
    if (days) {
        const date = new Date();
        date.setTime(date.getTime() + (days * 24 * 60 * 60 * 1000));
        expires = "; expires=" + date.toUTCString();
    }
    document.cookie = name + "=" + (value || "") + expires + "; path=/; SameSite=Strict";
}

export function getCookie(name) {
    const nameEQ = name + "=";
    const ca = document.cookie.split(';');
    for (let i = 0; i < ca.length; i++) {
        let c = ca[i];
        while (c.charAt(0) === ' ') c = c.substring(1, c.length);
        if (c.indexOf(nameEQ) === 0) return c.substring(nameEQ.length, c.length);
    }
    return null;
}

export function eraseCookie(name) {
    document.cookie = name + '=; Path=/; Expires=Thu, 01 Jan 1970 00:00:01 GMT;';
}

export function toggleButtonLoading(button, isLoading) {
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

export function triggerDownload(blob, filename) {
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

export function formatCompactNumber(number) {
    if (number === undefined || number === null) return '0';
    return new Intl.NumberFormat(undefined, {
        notation: "compact",
        compactDisplay: "short",
        maximumFractionDigits: 1
    }).format(number);
}
