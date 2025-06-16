'use client'

import {useEffect} from 'react'
import {useRouter} from 'next/navigation'
import Dashboard from '../components/Dashboard'

export default function DashboardPage() {
    const router = useRouter()

    useEffect(() => {
        // Check if user is authenticated
        const token = localStorage.getItem('authToken')
        if (!token) {
            router.push('/')
        }
    }, [router])

    return <Dashboard/>
} 