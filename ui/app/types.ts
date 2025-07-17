export interface KeyStatus {
  is_locked: boolean
  current_shares: number
  min_shares: number
  total_shares: number
}

export interface Share {
  index: number
  value: string
  isOwned: boolean
}

export interface ShareResponse {
  index: number
  unlocked: boolean
}

export interface SecretOwner {
    id: number;
    username: string;
    created_at: string;
    updated_at: string;
}

export interface SecretInfoSummary {
    path: string;
    owner: SecretOwner;
    created_at: string;
    updated_at: string;
}

export interface SecretInfo extends SecretInfoSummary {
    authorized_users: SecretOwner[];
    authorized_roles: { id: number; name: string }[];
}

export interface FolderNode {
    name: string;
    path: string;
    children: FolderNode[];
    secrets: SecretInfoSummary[];
    isFolder: boolean;
}

export interface ListSecretsResponse {
    secrets: SecretInfoSummary[];
}