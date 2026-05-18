import { OpenAPI } from '../api';
import { DefaultService } from '../api/services/DefaultService';
import { auth, setAuth } from './auth';

// Konfigurer OpenAPI client
OpenAPI.BASE = import.meta.env.VITE_API_URL || '';
OpenAPI.TOKEN = async () => {
    return auth.token || '';
};

// Vi kan også legge til en global feilhåndterer hvis ønskelig
// Men for nå beholder vi eksisterende flyt

export const api = {
    signup: (data: any) => DefaultService.postApiAuthSignup(data),
    login: async (data: any) => {
        const res = await DefaultService.postApiAuthLogin(data);
        setAuth(res.user as any, res.access_token);
        return res;
    },
    getMembers: (orgId?: string) => DefaultService.getApiMembers(orgId),
    getMembersAsOf: (date: string, orgId?: string) => DefaultService.getApiReportsAsOf(date, orgId),
    getOrganizations: () => DefaultService.getApiOrganizations(),
    registerMember: (data: any) => DefaultService.postApiCommandsRegisterMember(data),
    updateMember: (id: string, fields: any) => DefaultService.postApiCommandsUpdateMember({ id, updated_fields: fields }),
    shredMember: (id: string) => DefaultService.postApiCommandsShredMember({ id }),
};
