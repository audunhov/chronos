import { auth, setAuth } from './auth';

const API_BASE = '/api';

async function request(path: string, options: RequestInit = {}, orgId?: string) {
    const token = auth.token;
    
    const headers = {
        'Content-Type': 'application/json',
        ...options.headers,
    } as Record<string, string>;

    if (token) {
        headers['Authorization'] = `Bearer ${token}`;
    }

    if (orgId) {
        headers['X-Org-ID'] = orgId;
    }

    const response = await fetch(`${API_BASE}${path}`, {
        ...options,
        headers,
    });

    if (!response.ok) {
        const error = await response.text();
        if (response.status === 401) {
            setAuth(null, null);
        }
        throw new Error(error || response.statusText);
    }

    if (response.status === 204) return null;
    return response.json();
}

export const api = {
    signup: (data: any) => request('/auth/signup', {
        method: 'POST',
        body: JSON.stringify(data),
    }),
    login: async (data: any) => {
        const res = await request('/auth/login', {
            method: 'POST',
            body: JSON.stringify(data),
        });
        setAuth(res.user, res.access_token);
        return res;
    },
    getMembers: (orgId?: string) => request('/members', {}, orgId),
    getMembersAsOf: (date: string, orgId?: string) => request(`/reports/as-of?date=${date}`, {}, orgId),
    getOrganizations: () => request('/organizations'),
    registerMember: (data: { name: string; email: string; org_id?: string; metadata?: any }) => 
        request('/commands/register-member', {
            method: 'POST',
            body: JSON.stringify(data),
        }),
    updateMember: (id: string, fields: any) => 
        request('/commands/update-member', {
            method: 'POST',
            body: JSON.stringify({ id, updated_fields: fields }),
        }),
    shredMember: (id: string) => 
        request('/commands/shred-member', {
            method: 'POST',
            body: JSON.stringify({ id }),
        }),
};
