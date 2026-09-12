import { useEffect, useState } from 'react'
import Box from '@mui/material/Box'
import CircularProgress from '@mui/material/CircularProgress'
import { api, AuthError } from './api.js'
import LoginPage from './pages/LoginPage.jsx'
import Dashboard from './pages/Dashboard.jsx'
import HelpPage from './pages/HelpPage.jsx'

export default function App() {
  // 'checking' | 'loggedOut' | 'loggedIn'
  const [authState, setAuthState] = useState('checking')
  // 'dashboard' | 'help' — no router in this app; this hand-rolls the one
  // extra view the same way authState already hand-rolls login vs dashboard.
  const [view, setView] = useState('dashboard')

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

  if (view === 'help') {
    return <HelpPage onBack={() => setView('dashboard')} onLogout={() => setAuthState('loggedOut')} />
  }

  return (
    <Dashboard
      onAuthError={handleAuthError}
      onLogout={() => setAuthState('loggedOut')}
      onShowHelp={() => setView('help')}
    />
  )
}
