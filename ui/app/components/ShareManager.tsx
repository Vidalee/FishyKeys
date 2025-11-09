'use client'

import {useEffect, useState} from 'react'
import {addShare, deleteShare, KeyStatus} from '../api/keyManagement'
import styles from './ShareManager.module.css'
import {Share} from '../types'

interface ShareManagerProps {
  keyStatus: KeyStatus
  onStatusChange: () => void
}

export default function ShareManager({ keyStatus, onStatusChange }: ShareManagerProps) {
  const [ownedIndices, setOwnedIndices] = useState<number[]>([])
  const [newShare, setNewShare] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    if (ownedIndices.length > keyStatus.current_shares) {
      setOwnedIndices([])
    }
  }, [keyStatus.current_shares])

  const shares: Share[] = Array.from({ length: keyStatus.current_shares }, (_, i) => ({
    index: i,
    value: '',
    isOwned: ownedIndices.includes(i),
  }))

  const handleAddShare = async () => {
    if (!newShare.trim()) return
    setLoading(true)
    setError(null)
    try {
      const response = await addShare(newShare)
      setOwnedIndices(prev => [...prev, response.index])
      setNewShare('')
      onStatusChange()
    } catch (err: any) {
      setError(err?.body || 'Failed to add share')
    } finally {
      setLoading(false)
    }
  }

  const handleDeleteShare = async (index: number) => {
    setLoading(true)
    setError(null)
    try {
      await deleteShare(index)
      setOwnedIndices(prev => prev.filter(i => i !== index))
      onStatusChange()
    } catch (err: any) {
      setError(err?.body || 'Failed to delete share')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className={styles.card}>
      <div className={styles.sectionTitle}>Key Status</div>
      <div className={styles.statusGrid}>
        <div className={styles.statusBox}>
          <div className={styles.statusLabel}>Current Shares</div>
          <div className={styles.statusValue}>{keyStatus.current_shares}</div>
        </div>
        <div className={styles.statusBox}>
          <div className={styles.statusLabel}>Required Shares</div>
          <div className={styles.statusValue}>{keyStatus.min_shares}</div>
        </div>
      </div>

      <div className={styles.sectionTitle}>Add Share</div>
      <div className={styles.inputRow}>
        <input
          type="password"
          name="share"
          id="share"
          className={styles.input}
          placeholder="Enter your share"
          value={newShare}
          onChange={(e) => setNewShare(e.target.value)}
        />
        <button
          type="button"
          className={styles.button}
          onClick={handleAddShare}
          disabled={loading || !newShare.trim()}
        >
          Add
        </button>
      </div>

      {error && (
        <div className={styles.error} role="alert">
          <span>{error}</span>
        </div>
      )}

      <div className={styles.sectionTitle}>Current Shares</div>
      <div className={styles.sharesList}>
        {shares.map((share) => (
          <div
            key={share.index}
            className={styles.shareItem}
          >
            <div>
              <div className={styles.shareLabel}>Share {share.index + 1}</div>
              <div className={styles.shareType}>{share.isOwned ? 'Your share' : 'Other share'}</div>
            </div>
            <button
              type="button"
              className={styles.deleteBtn}
              onClick={() => handleDeleteShare(share.index)}
              disabled={loading}
            >
              Delete
            </button>
          </div>
        ))}
      </div>
    </div>
  )
} 