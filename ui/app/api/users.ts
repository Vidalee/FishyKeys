import {Role} from "./roles";

export interface User {
    id: number;
    username: string;
    created_at: string;
    updated_at: string;
    roles: Role[];
}

export interface AuthResponse {
    username: string;
    token: string;
}

export interface CreateUserPayload {
    username: string;
    password: string;
}

export interface CreateUserResponse {
    id: number;
    username: string;
}

function getAuthHeaders(): Record<string, string> {
    const token = localStorage.getItem("authToken");
    if (token) {
        return {'Authorization': `Bearer ${token}`};
    }
    return {};
}

const API_BASE = "/users";

export async function login(username: string, password: string): Promise<AuthResponse> {
    const response = await fetch(`${API_BASE}/auth`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({username, password}),
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

export async function listUsers(): Promise<User[]> {
    const headers: Record<string, string> = {
        'Content-Type': 'application/json',
        ...getAuthHeaders(),
    };
    const response = await fetch(API_BASE, {
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

export async function createUser(payload: CreateUserPayload): Promise<CreateUserResponse> {
    const headers: Record<string, string> = {
        'Content-Type': 'application/json',
        ...getAuthHeaders(),
    };
    const response = await fetch(API_BASE, {
        method: 'POST',
        headers,
        credentials: 'include',
        body: JSON.stringify(payload),
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

export async function deleteUser(username: string): Promise<void> {
    const headers: Record<string, string> = {
        'Content-Type': 'application/json',
        ...getAuthHeaders(),
    };
    const response = await fetch(`${API_BASE}/${username}`, {
        method: 'DELETE',
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
    return;
} 