'use client'

import {useEffect, useState} from 'react'
import {KeyStatus} from '../types'
import LoginScreen from './LoginScreen'
import CreateMasterKey from './CreateMasterKey'
import {getKeyStatus} from '../api/keyManagement'
import styles from './KeyManager.module.css'

export default function KeyManager() {
  const [keyStatus, setKeyStatus] = useState<KeyStatus | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [noKey, setNoKey] = useState(false)
  const [isCreatingKey, setIsCreatingKey] = useState(false)
  const [isAuthenticated, setIsAuthenticated] = useState(false)

  useEffect(() => {
    // Check if user is authenticated
    const token = localStorage.getItem('authToken')
    setIsAuthenticated(!!token)
  }, [])

  const fetchKeyStatus = async () => {
    if (isCreatingKey) return
    try {
      const status = await getKeyStatus()
      setKeyStatus(status)
      setNoKey(false)
      setError(null)
    } catch (err: any) {
      if (err.status === 404 && err.body === 'master key not set') {
        setNoKey(true)
        setKeyStatus(null)
        setError(null)
      } else {
        setError(err?.body || 'Failed to fetch key status')
      }
    }
  }

  useEffect(() => {
    fetchKeyStatus()
    const interval = setInterval(fetchKeyStatus, 5000)
    return () => clearInterval(interval)
  }, [isCreatingKey])

  const handleLoginSuccess = () => {
    setIsAuthenticated(true)
  }

  if (error) {
    return (
      <div className={styles.container}>
        <div style={{ color: 'red', background: '#fff', padding: 16, borderRadius: 8, boxShadow: '0 2px 8px rgba(0,0,0,0.06)' }}>
          <strong>Error: </strong>
          <span>{error}</span>
        </div>
      </div>
    )
  }

  if (noKey) {
    return <CreateMasterKey onCreatingKey={setIsCreatingKey} />
  }

  if (!isAuthenticated) {
    return <LoginScreen onLoginSuccess={handleLoginSuccess}/>
  }

  if (!keyStatus) {
    return (
        <div className={styles.container}>
          <div className={styles.loading}>Loading...</div>
        </div>
    )
  }

  // If we have a key status and user is authenticated, redirect to dashboard
  window.location.href = '/dashboard'
  return null
} 