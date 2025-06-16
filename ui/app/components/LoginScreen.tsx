'use client'

import {useState} from 'react'
import {login} from '../api/users'
import styles from './LoginScreen.module.css'

interface LoginScreenProps {
  onLoginSuccess: () => void
}

export default function LoginScreen({onLoginSuccess}: LoginScreenProps) {
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault()
    setLoading(true)
    setError(null)
    try {
      const data = await login(username, password)
      // Store the auth token
      localStorage.setItem('authToken', data.token)
      // You can store the username too if needed
      localStorage.setItem('username', data.username)

      // Call the onLoginSuccess callback
      onLoginSuccess()
    } catch (err: any) {
      let errorMessage = 'Login failed'
      if (err?.body) {
        if (err.body.name === 'unauthorized') {
          errorMessage = 'Invalid username or password'
        } else if (err.body.name === 'invalid_parameters') {
          errorMessage = err.body.message
        } else if (err.body.name === 'internal_error') {
          errorMessage = 'Internal server error occurred'
        }
      }
      setError(errorMessage)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className={styles.container}>
      <form className={styles.card} onSubmit={handleLogin}>
        <div className={styles.title}>Login</div>
        <div className={styles.inputGroup}>
          <label htmlFor="username" className={styles.label}>Username</label>
          <input
            id="username"
            name="username"
            type="text"
            className={styles.input}
            value={username}
            onChange={e => setUsername(e.target.value)}
            required
          />
        </div>
        <div className={styles.inputGroup}>
          <label htmlFor="password" className={styles.label}>Password</label>
          <input
            id="password"
            name="password"
            type="password"
            className={styles.input}
            value={password}
            onChange={e => setPassword(e.target.value)}
            required
          />
        </div>
        {error && <div className={styles.error} role="alert"><span>{error}</span></div>}
        <button type="submit" className={styles.button} disabled={loading}>
          {loading ? 'Logging in...' : 'Login'}
        </button>
      </form>
    </div>
  )
} 