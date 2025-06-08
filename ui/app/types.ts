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