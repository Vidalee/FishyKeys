import {Metadata} from 'next'
import KeyManager from './components/KeyManager'

export const metadata: Metadata = {
  title: 'FishyKeys - Secure Key Management',
  description: 'A secure and ergonomic key management system',
}

export default function Home() {
  return <KeyManager />
} 