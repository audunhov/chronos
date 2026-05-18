/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { AuthResponse } from '../models/AuthResponse';
import type { LoginRequest } from '../models/LoginRequest';
import type { Member } from '../models/Member';
import type { RegisterMemberRequest } from '../models/RegisterMemberRequest';
import type { ShredMemberRequest } from '../models/ShredMemberRequest';
import type { SignupRequest } from '../models/SignupRequest';
import type { UpdateMemberRequest } from '../models/UpdateMemberRequest';
import type { CancelablePromise } from '../core/CancelablePromise';
import { OpenAPI } from '../core/OpenAPI';
import { request as __request } from '../core/request';
export class DefaultService {
    /**
     * Health check
     * @returns string OK
     * @throws ApiError
     */
    public static getApiHealth(): CancelablePromise<string> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/api/health',
        });
    }
    /**
     * User signup
     * @param requestBody
     * @returns any User created
     * @throws ApiError
     */
    public static postApiAuthSignup(
        requestBody: SignupRequest,
    ): CancelablePromise<{
        message?: string;
    }> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/api/auth/signup',
            body: requestBody,
            mediaType: 'application/json',
        });
    }
    /**
     * User login
     * @param requestBody
     * @returns AuthResponse Login successful
     * @throws ApiError
     */
    public static postApiAuthLogin(
        requestBody: LoginRequest,
    ): CancelablePromise<AuthResponse> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/api/auth/login',
            body: requestBody,
            mediaType: 'application/json',
            errors: {
                401: `Invalid credentials`,
            },
        });
    }
    /**
     * Get members
     * @param xOrgId Optional organization filter
     * @returns Member List of members
     * @throws ApiError
     */
    public static getApiMembers(
        xOrgId?: string,
    ): CancelablePromise<Array<Member>> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/api/members',
            headers: {
                'X-Org-ID': xOrgId,
            },
        });
    }
    /**
     * Get unique organization IDs
     * @returns string List of organization IDs
     * @throws ApiError
     */
    public static getApiOrganizations(): CancelablePromise<Array<string>> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/api/organizations',
        });
    }
    /**
     * Register a new member
     * @param requestBody
     * @returns any Member registered
     * @throws ApiError
     */
    public static postApiCommandsRegisterMember(
        requestBody: RegisterMemberRequest,
    ): CancelablePromise<{
        id?: string;
    }> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/api/commands/register-member',
            body: requestBody,
            mediaType: 'application/json',
        });
    }
    /**
     * Update member fields
     * @param requestBody
     * @returns any Member updated
     * @throws ApiError
     */
    public static postApiCommandsUpdateMember(
        requestBody: UpdateMemberRequest,
    ): CancelablePromise<any> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/api/commands/update-member',
            body: requestBody,
            mediaType: 'application/json',
        });
    }
    /**
     * Shred member data (GDPR)
     * @param requestBody
     * @returns any Member shredded
     * @throws ApiError
     */
    public static postApiCommandsShredMember(
        requestBody: ShredMemberRequest,
    ): CancelablePromise<any> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/api/commands/shred-member',
            body: requestBody,
            mediaType: 'application/json',
        });
    }
    /**
     * Get members as of a specific date
     * @param date
     * @param xOrgId
     * @returns Member List of members as of date
     * @throws ApiError
     */
    public static getApiReportsAsOf(
        date: string,
        xOrgId?: string,
    ): CancelablePromise<Array<Member>> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/api/reports/as-of',
            headers: {
                'X-Org-ID': xOrgId,
            },
            query: {
                'date': date,
            },
        });
    }
}
