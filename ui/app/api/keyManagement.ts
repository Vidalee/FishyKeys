export interface KeyStatus {
    is_locked: boolean
    current_shares: number
    min_shares: number
    total_shares: number
}


export interface ShareResponse {
    index: number
    unlocked: boolean
}

const API_BASE = "/key_management"

export async function getKeyStatus(): Promise<KeyStatus> {
  const response = await fetch(`${API_BASE}/status`);
    console.log("resp:", response.ok)
  if (!response.ok) {
    let errorBody: any = null;
    try {
      errorBody = await response.text();
      try {
        errorBody = JSON.parse(errorBody);
      } catch {}
        console.log("errorBody:", errorBody)
    } catch {}
    throw { status: response.status, body: errorBody };
  }
  return response.json();
}

export async function createMasterKey(
  totalShares: number, 
  minShares: number, 
  adminUsername: string, 
  adminPassword: string
): Promise<string[]> {
  const response = await fetch(`${API_BASE}/create_master_key`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ 
      total_shares: totalShares, 
      min_shares: minShares,
      admin_username: adminUsername,
      admin_password: adminPassword
    }),
  });
  if (!response.ok) {
    let errorBody: any = null;
    try {
      errorBody = await response.text();
      try {
        errorBody = JSON.parse(errorBody);
      } catch {}
    } catch {}
    throw { status: response.status, body: errorBody };
  }
  const data = await response.json();
  return data.shares;
}

export async function addShare(share: string): Promise<ShareResponse> {
  const response = await fetch(`${API_BASE}/share`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ share }),
  });
  if (!response.ok) {
    let errorBody: any = null;
    try {
      errorBody = await response.text();
      try {
        errorBody = JSON.parse(errorBody);
      } catch {}
    } catch {}
    throw { status: response.status, body: errorBody };
  }
  return response.json();
}

export async function deleteShare(index: number): Promise<void> {
  const response = await fetch(`${API_BASE}/share`, {
    method: 'DELETE',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ index }),
  });
  if (!response.ok) {
    let errorBody: any = null;
    try {
      errorBody = await response.text();
      try {
        errorBody = JSON.parse(errorBody);
      } catch {}
    } catch {}
    throw { status: response.status, body: errorBody };
  }
} 
