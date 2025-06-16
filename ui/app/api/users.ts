export interface AuthResponse {
    username: string
    token: string
}

export async function login(username: string, password: string): Promise<AuthResponse> {
    const response = await fetch(`/users/auth`, {
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