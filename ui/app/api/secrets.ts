import {SecretInfoSummary, SecretOwner} from "../types";

const API_BASE = "/secrets"

function getAuthHeaders(): Record<string, string> {
    const token = localStorage.getItem("authToken");
    if (token) {
        return {'Authorization': `Bearer ${token}`};
    }
    return {};
}

export async function listSecrets(): Promise<SecretInfoSummary[]> {
    const baseHeaders: Record<string, string> = {
        'Content-Type': 'application/json',
    };
    const headers: Record<string, string> = {...baseHeaders, ...getAuthHeaders()};
    const response = await fetch(`${API_BASE}`, {
        method: 'GET',
        headers,
        credentials: 'include',
    });
    console.log("resp:", response.ok)
    console.log("resp:", response.status)
    if (!response.ok) {
        let errorBody: any = null;
        try {
            errorBody = await response.text();
            try {
                errorBody = JSON.parse(errorBody);
            } catch {
            }
        } catch {
        }
        throw {status: response.status, body: errorBody};
    }
    return response.json();
}

export interface SecretInfo extends SecretInfoSummary {
    authorized_users: SecretOwner[];
    authorized_roles: { id: number; name: string }[];
}

export async function getSecret(path: string): Promise<SecretInfo> {
    const baseHeaders: Record<string, string> = {
        'Content-Type': 'application/json',
    };
    const headers: Record<string, string> = {...baseHeaders, ...getAuthHeaders()};
    const response = await fetch(`${API_BASE}/${encodeURIComponent(path)}`, {
        method: 'GET',
        headers,
        credentials: 'include',
    });
    if (!response.ok) {
        let errorBody: any = null;
        try {
            errorBody = await response.text();
            try {
                errorBody = JSON.parse(errorBody);
            } catch {
            }
        } catch {
        }
        throw {status: response.status, body: errorBody};
    }
    return response.json();
}

export async function createSecret({path, value, authorized_users, authorized_roles}: {
    path: string;
    value: string;
    authorized_users: number[];
    authorized_roles: number[]
}): Promise<void> {
    const baseHeaders: Record<string, string> = {
        'Content-Type': 'application/json',
    };
    const headers: Record<string, string> = {...baseHeaders, ...getAuthHeaders()};
    // Path must be base64 encoded
    const encodedPath = btoa(path);
    const body = JSON.stringify({
        path: encodedPath,
        value,
        authorized_users,
        authorized_roles,
    });
    const response = await fetch(`${API_BASE}`, {
        method: 'POST',
        headers,
        credentials: 'include',
        body,
    });
    if (!response.ok) {
        let errorBody: any = null;
        try {
            errorBody = await response.text();
            try {
                errorBody = JSON.parse(errorBody);
            } catch {
            }
        } catch {
        }
        throw {status: response.status, body: errorBody};
    }
} 