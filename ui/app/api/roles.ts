export interface Role {
    id: number;
    name: string;
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