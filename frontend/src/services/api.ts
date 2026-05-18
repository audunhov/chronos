import { supabase } from './supabase';

const API_BASE = import.meta.env.VITE_API_URL || 'http://localhost:8080/api';

async function request(path: string, options: RequestInit = {}) {
    const { data: { session } } = await supabase.auth.getSession();
    const token = session?.access_token;
    
    const headers = {
        'Content-Type': 'application/json',
        ...options.headers,
    } as Record<string, string>;

    if (token) {
        headers['Authorization'] = `Bearer ${token}`;
    }

    const response = await fetch(`${API_BASE}${path}`, {
        ...options,
        headers,
    });

    if (!response.ok) {
        const error = await response.text();
        throw new Error(error || response.statusText);
    }

    if (response.status === 204) return null;
    return response.json();
}

export const api = {
    getMembers: () => request('/members'),
    getMembersAsOf: (date: string) => request(`/reports/as-of?date=${date}`),
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
