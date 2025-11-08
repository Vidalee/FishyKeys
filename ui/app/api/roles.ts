export interface Role {
    id: number;
    name: string;
    color: string;
    admin: boolean;
    created_at: string;
}

export interface CreateRolePayload {
    name: string;
    color: string;
}

export interface AssignRolePayload {
    user_id: number;
    role_id: number;
}

export interface UnassignRolePayload {
    user_id: number;
    role_id: number;
}

function getAuthHeaders(): Record<string, string> {
    const token = localStorage.getItem("authToken");
    if (token) {
        return {'Authorization': `Bearer ${token}`};
    }
    return {};
}

export async function listRoles(): Promise<Role[]> {
    const headers: Record<string, string> = {
        'Content-Type': 'application/json',
        ...getAuthHeaders(),
    };
    const response = await fetch('/roles', {
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

export async function createRole(payload: CreateRolePayload): Promise<Role[]> {
    const headers: Record<string, string> = {
        'Content-Type': 'application/json',
        ...getAuthHeaders(),
    };
    const response = await fetch('/roles', {
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

export async function deleteRole(roleId: number): Promise<void> {
    const headers: Record<string, string> = {
        'Content-Type': 'application/json',
        ...getAuthHeaders(),
    };
    const response = await fetch(`/roles/${roleId}`, {
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
    return
}

export async function assignRoleToUser(payload: AssignRolePayload): Promise<void> {
    const headers: Record<string, string> = {
        'Content-Type': 'application/json',
        ...getAuthHeaders(),
    };
    const response = await fetch('/roles/assign', {
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
    return;
}

export async function unassignRoleFromUser(payload: UnassignRolePayload): Promise<void> {
    const headers: Record<string, string> = {
        'Content-Type': 'application/json',
        ...getAuthHeaders(),
    };
    const response = await fetch('/roles/unassign', {
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
    return;
}