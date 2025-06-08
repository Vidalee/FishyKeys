import {useState} from 'react'
import {createMasterKey} from '../api/keyManagement'
import styles from './LoginScreen.module.css'

interface CreateMasterKeyProps {
  onCreatingKey: (isCreating: boolean) => void
}

export default function CreateMasterKey({ onCreatingKey }: CreateMasterKeyProps) {
  const [totalShares, setTotalShares] = useState(5)
  const [minShares, setMinShares] = useState(3)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [shares, setShares] = useState<string[] | null>(null)

  const handleCreateKey = async (e: React.FormEvent) => {
    e.preventDefault()
    setLoading(true)
    setError(null)
    onCreatingKey(true)
    try {
      const generatedShares = await createMasterKey(totalShares, minShares)
      setShares(generatedShares)
    } catch (err: any) {
      setError(err?.body || 'Failed to create master key')
      onCreatingKey(false)
    } finally {
      setLoading(false)
    }
  }

  const handleProceed = () => {
    onCreatingKey(false)
    window.location.reload()
  }

  if (shares) {
    return (
      <div className={styles.container}>
        <div className={styles.card}>
          <div className={styles.title}>Master Key Shares</div>
          <div className={styles.sharesContainer}>
            <p className={styles.sharesInfo}>
              Please save these shares securely. You will need at least {minShares} shares to unlock the master key.
            </p>
            {shares.map((share, index) => (
              <div key={index} className={styles.shareItem}>
                <span className={styles.shareIndex}>Share {index + 1}:</span>
                <code className={styles.shareCode}>{share}</code>
              </div>
            ))}
          </div>
          <button onClick={handleProceed} className={styles.button}>
            I have saved my shares
          </button>
        </div>
      </div>
    )
  }

  return (
    <div className={styles.container}>
      <form className={styles.card} onSubmit={handleCreateKey}>
        <div className={styles.title}>Create Master Key</div>
        <div className={styles.inputGroup}>
          <label htmlFor="total-shares" className={styles.label}>
            Total Number of Shares
          </label>
          <input
            id="total-shares"
            name="total-shares"
            type="number"
            min="2"
            max="10"
            required
            className={styles.input}
            value={totalShares}
            onChange={(e) => setTotalShares(parseInt(e.target.value))}
          />
        </div>
        <div className={styles.inputGroup}>
          <label htmlFor="min-shares" className={styles.label}>
            Minimum Required Shares
          </label>
          <input
            id="min-shares"
            name="min-shares"
            type="number"
            min="2"
            max={totalShares}
            required
            className={styles.input}
            value={minShares}
            onChange={(e) => setMinShares(parseInt(e.target.value))}
          />
        </div>
        {error && (
          <div className={styles.error} role="alert">
            <span>{error}</span>
          </div>
        )}
        <button type="submit" className={styles.button} disabled={loading || minShares > totalShares}>
          {loading ? 'Creating...' : 'Create Master Key'}
        </button>
      </form>
    </div>
  )
} 