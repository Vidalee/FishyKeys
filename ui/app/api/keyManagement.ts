import {KeyStatus, ShareResponse} from '../types'

const API_BASE = process.env.NEXT_PUBLIC_API_BASE
  ? `${process.env.NEXT_PUBLIC_API_BASE}/key_management`
  : '/key_management'

export async function getKeyStatus(): Promise<KeyStatus> {
  const response = await fetch(`${API_BASE}/status`);
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

export async function createMasterKey(totalShares: number, minShares: number): Promise<string[]> {
  const response = await fetch(`${API_BASE}/create_master_key`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ total_shares: totalShares, min_shares: minShares }),
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