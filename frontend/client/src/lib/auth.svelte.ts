import { browser } from '$app/environment';

const COOKIE_CSRF_USER = 'csrf';
const COOKIE_CSRF_GUEST = 'csrf_guest';
export const HEADER_CSRF_USER = 'X-CSRF-Token';
export const HEADER_CSRF_GUEST = 'X-CSRF-Token-Guest';

/**
 * Gets a cookie value by name, returns an empty string if not found or if not in browser.
 * @param {string} name 
 * @returns {string}
 */
export function getCookie(name: string) {
    if (!browser) return '';
    const match = document.cookie.match(new RegExp('(^| )' + name + '=([^;]+)'));
    return match ? decodeURIComponent(match[2]) : '';
}

export function getGuestToken(): string {
    return getCookie(COOKIE_CSRF_GUEST);
}

export function getUserToken(): string {
    return getCookie(COOKIE_CSRF_USER);
}

export function getUserTokenHeaders(): Record<string, string> {
    const userToken = getUserToken();
    const headers: Record<string, string> = {'Content-Type': 'application/json'};
    headers[HEADER_CSRF_USER] = userToken;
    return headers
}

export function getGuestTokenHeaders(): Record<string, string> {
    const userToken = getGuestToken();
    const headers: Record<string, string> = {'Content-Type': 'application/json'};
    headers[HEADER_CSRF_GUEST] = userToken;
    return headers
}

export function getUserOrGuestTokenHeaders(): Record<string, string> {
    const userToken = getUserToken();
    const headers: Record<string, string> = {'Content-Type': 'application/json'};
    if (!userToken) {
        headers[HEADER_CSRF_GUEST] = getGuestToken();
        return headers
    }
    headers[HEADER_CSRF_USER] = userToken;
    return headers
}

class AuthState {
    get isLoggedIn() {
        return getUserToken() !== '';
    }
    logout() {
    }
}

export const authState = new AuthState();