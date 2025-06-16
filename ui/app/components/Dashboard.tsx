import {useEffect, useState} from 'react'
import {KeyStatus} from '../types'
import ShareManager from './ShareManager'
import {getKeyStatus} from '../api/keyManagement'
import styles from './Dashboard.module.css'

export default function Dashboard() {
    const [keyStatus, setKeyStatus] = useState<KeyStatus | null>(null)
    const [error, setError] = useState<string | null>(null)

    const fetchKeyStatus = async () => {
        try {
            const status = await getKeyStatus()
            setKeyStatus(status)
            setError(null)
        } catch (err: any) {
            setError(err?.body || 'Failed to fetch key status')
        }
    }

    useEffect(() => {
        fetchKeyStatus()
        const interval = setInterval(fetchKeyStatus, 5000)
        return () => clearInterval(interval)
    }, [])

    if (error) {
        return (
            <div className={styles.container}>
                <div className={styles.error}>
                    <strong>Error: </strong>
                    <span>{error}</span>
                </div>
            </div>
        )
    }

    if (!keyStatus) {
        return (
            <div className={styles.container}>
                <div className={styles.loading}>Loading...</div>
            </div>
        )
    }

    return (
        <div className={styles.container}>
            {keyStatus.is_locked ? (
                <ShareManager keyStatus={keyStatus} onStatusChange={fetchKeyStatus}/>
            ) : (
                <div className={styles.unlockedMessage}>
                    <h1>Key is Unlocked</h1>
                    <p>The master key is currently unlocked and ready to use.</p>
                </div>
            )}
        </div>
    )
} 