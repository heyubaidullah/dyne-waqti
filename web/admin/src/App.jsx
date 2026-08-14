import { useEffect, useState } from 'react'
import Box from '@mui/material/Box'
import CircularProgress from '@mui/material/CircularProgress'
import { api, AuthError } from './api.js'
import LoginPage from './pages/LoginPage.jsx'
import Dashboard from './pages/Dashboard.jsx'

export default function App() {
  // 'checking' | 'loggedOut' | 'loggedIn'
  const [authState, setAuthState] = useState('checking')

  useEffect(() => {
    api
      .checkSession()
      .then(() => setAuthState('loggedIn'))
      .catch(() => setAuthState('loggedOut'))
  }, [])

  const handleAuthError = (err) => {
    if (err instanceof AuthError) {
      setAuthState('loggedOut')
      return true
    }
    return false
  }

  if (authState === 'checking') {
    return (
      <Box sx={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100vh' }}>
        <CircularProgress color="primary" />
      </Box>
    )
  }

  if (authState === 'loggedOut') {
    return <LoginPage onLogin={() => setAuthState('loggedIn')} />
  }

  return <Dashboard onAuthError={handleAuthError} onLogout={() => setAuthState('loggedOut')} />
}
