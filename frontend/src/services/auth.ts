import { ref, reactive } from 'vue';

interface User {
    id: string;
    email: string;
    org_id: string;
    role: string;
}

interface AuthState {
    user: User | null;
    token: string | null;
    loading: boolean;
}

const STORAGE_KEY = 'chronos_auth';

const getStoredAuth = (): { user: User | null; token: string | null } => {
    const stored = localStorage.getItem(STORAGE_KEY);
    if (stored) {
        try {
            return JSON.parse(stored);
        } catch (e) {
            console.error('Failed to parse stored auth', e);
        }
    }
    return { user: null, token: null };
};

const stored = getStoredAuth();

export const auth = reactive<AuthState>({
    user: stored.user,
    token: stored.token,
    loading: false,
});

export const setAuth = (user: User | null, token: string | null) => {
    auth.user = user;
    auth.token = token;
    if (user && token) {
        localStorage.setItem(STORAGE_KEY, JSON.stringify({ user, token }));
    } else {
        localStorage.removeItem(STORAGE_KEY);
    }
};

export const logout = () => {
    setAuth(null, null);
};
