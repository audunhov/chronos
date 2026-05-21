import * as chronos from '../api/chronosComponents';
import { auth, setAuth } from './auth';

// Hjelpefunksjon for å håndtere X-Org-ID
const getHeaders = (orgId?: string) => {
    const headers: Record<string, string> = {};
    if (orgId) {
        headers['X-Org-ID'] = orgId;
    }
    return headers;
};

export const api = {
    signup: (data: any) => chronos.signup({ body: data }),
    login: async (data: any) => {
        const res = await chronos.login({ body: data });
        // @ts-ignore - user structure might slightly differ, but we handle it
        setAuth(res.user, res.access_token);
        return res;
    },
    requestMagicLink: (email: string) => chronos.requestMagicLink({ body: { email } }),
    magicLogin: async (token: string) => {
        const res = await chronos.magicLogin({ queryParams: { token } });
        // @ts-ignore
        setAuth(res.user, res.access_token);
        return res;
    },

    getMyProfile: () => chronos.getMyProfile(),
    updateMyProfile: (data: { name?: string, email?: string }) => chronos.updateMyProfile({ body: data }),
    
    // Admin User Inspection
    getUserProfile: (id: string) => chronos.getUserProfile({ pathParams: { id } }),
    getUserMemberships: (id: string) => chronos.getUserMemberships({ pathParams: { id } }),

    getMyMemberships: () => chronos.getMyMemberships(),
    getMembers: (orgId?: string) => chronos.getMembers({ 
        headers: getHeaders(orgId) 
    }),
    getMembersAsOf: (date: string, orgId?: string) => chronos.getMembersAsOf({ 
        queryParams: { date },
        headers: getHeaders(orgId)
    }),
    
    // Organizations
    getOrganizations: async () => {
        const data = await chronos.getOrganizations();
        return Array.isArray(data) ? data : [];
    },
    getOrganizationHierarchy: () => chronos.getOrganizationHierarchy(),
    createOrganization: (data: any) => (chronos as any).createOrganization({ body: data }),
    deleteOrganization: (id: string) => (chronos as any).deleteOrganization({ queryParams: { id } }),

    // Admin Tools
    getTreasuryReport: () => chronos.getTreasuryReport(),
    getStats: () => chronos.getStats(),
    getAuditLogs: (orgId?: string, actorId?: string, targetId?: string) => 
        chronos.getAuditLogs({ queryParams: { org_id: orgId, actor_id: actorId, target_id: targetId } }),

    // Reactions (Pipelines)
    getReactions: (orgId: string, triggerAggregateId?: string) => 
        chronos.getReactions({ queryParams: { org_id: orgId, trigger_aggregate_id: triggerAggregateId } }),
    createReaction: (data: any) => chronos.createReaction({ body: data }),


    // Organs
    getOrgans: (orgId: string) => chronos.getOrgans({ queryParams: { org_id: orgId } }),
    createOrgan: (data: any) => chronos.createOrgan({ body: data }),
    getOrganMembers: (organId: string) => chronos.getOrganMembers({ queryParams: { organ_id: organId } }),
    assignOrganMember: (data: { organ_id: string, user_id: string, role_type: string }) => chronos.assignOrganMember({ body: data }),
    revokeOrganMember: (id: string) => chronos.revokeOrganMember({ queryParams: { id } }),

    // Forms
    getForms: (orgId?: string) => {
        const queryParams: any = {};
        if (orgId) queryParams.org_id = orgId;
        return chronos.getForms({ queryParams });
    },
    getFormResponses: (formId: string) => chronos.getFormResponses({ queryParams: { form_id: formId } }),
    createForm: (data: any) => chronos.createForm({ body: data }),
    updateForm: (data: any) => chronos.updateForm({ body: data }),
    submitForm: (formId: string, answers: any) => chronos.submitForm({ body: { form_id: formId, answers } }),
    
    // Member Management
    registerMember: (data: any) => chronos.registerMember({ body: data }),
    updateMember: (id: string, fields: any) => chronos.updateMember({ 
        body: { id, updated_fields: fields } 
    }),
    shredMember: (id: string) => chronos.shredMember({ 
        body: { id } 
    }),
};
