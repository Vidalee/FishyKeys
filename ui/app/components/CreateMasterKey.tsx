import {useState} from 'react'
import {createMasterKey} from '../api/keyManagement'
import styles from './LoginScreen.module.css'

interface CreateMasterKeyProps {
  onCreatingKey: (isCreating: boolean) => void
}

export default function CreateMasterKey({ onCreatingKey }: CreateMasterKeyProps) {
  const [totalShares, setTotalShares] = useState(5)
  const [minShares, setMinShares] = useState(3)
  const [adminUsername, setAdminUsername] = useState('')
  const [adminPassword, setAdminPassword] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [shares, setShares] = useState<string[] | null>(null)
  const [copiedIndex, setCopiedIndex] = useState<number | null>(null)

  const handleCopyShare = async (share: string, index: number) => {
    try {
      await navigator.clipboard.writeText(share)
      setCopiedIndex(index)
      setTimeout(() => setCopiedIndex(null), 2000) // Reset after 2 seconds
    } catch (err) {
      console.error('Failed to copy share:', err)
    }
  }

  const handleCreateKey = async (e: React.FormEvent) => {
    e.preventDefault()
    setLoading(true)
    setError(null)
    onCreatingKey(true)
    try {
      const generatedShares = await createMasterKey(totalShares, minShares, adminUsername, adminPassword)
      setShares(generatedShares)
    } catch (err: any) {
      let errorMessage = 'Failed to create master key'
      if (err?.body) {
        console.log(err.body)
        if (err.body.name === 'invalid_parameters') {
          errorMessage = err.body.message
        } else if (err.body === 'key_already_exists') {
          errorMessage = 'A master key already exists'
        } else if (err.body === 'internal_error') {
          errorMessage = 'Internal server error occurred'
        }
      }
      setError(errorMessage)
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
                <button
                  onClick={() => handleCopyShare(share, index)}
                  className={styles.copyButton}
                  title="Copy to clipboard"
                >
                  {copiedIndex === index ? '✓' : '📋'}
                </button>
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
          <label htmlFor="admin-username" className={styles.label}>
            Admin Username
          </label>
          <input
            id="admin-username"
            name="admin-username"
            type="text"
            required
            className={styles.input}
            value={adminUsername}
            onChange={(e) => setAdminUsername(e.target.value)}
          />
        </div>
        <div className={styles.inputGroup}>
          <label htmlFor="admin-password" className={styles.label}>
            Admin Password
          </label>
          <input
            id="admin-password"
            name="admin-password"
            type="password"
            required
            className={styles.input}
            value={adminPassword}
            onChange={(e) => setAdminPassword(e.target.value)}
          />
        </div>
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
        <button 
          type="submit" 
          className={styles.button} 
          disabled={loading || minShares > totalShares || !adminUsername || !adminPassword}
        >
          {loading ? 'Creating...' : 'Create Master Key'}
        </button>
      </form>
    </div>
  )
} 