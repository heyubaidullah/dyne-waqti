import { useCallback, useEffect, useState } from 'react'
import Box from '@mui/material/Box'
import Container from '@mui/material/Container'
import Stack from '@mui/material/Stack'
import CircularProgress from '@mui/material/CircularProgress'
import Alert from '@mui/material/Alert'
import { api, AuthError } from '../api.js'
import Header from '../components/Header.jsx'
import Footer from '../components/Footer.jsx'
import EmergencyControls from '../components/EmergencyControls.jsx'
import IqamahTimesForm from '../components/IqamahTimesForm.jsx'
import SettingsForm from '../components/SettingsForm.jsx'
import SlideManager from '../components/SlideManager.jsx'
import LogoUpload from '../components/LogoUpload.jsx'

export default function Dashboard({ onAuthError, onLogout }) {
  const [settings, setSettings] = useState(null)
  const [slides, setSlides] = useState(null)
  const [displayData, setDisplayData] = useState(null)
  const [loadError, setLoadError] = useState(null)

  const refreshAll = useCallback(async () => {
    try {
      const [settingsRes, slidesRes, displayRes] = await Promise.all([
        api.getSettings(),
        api.listSlides(),
        api.getDisplayData(),
      ])
      setSettings(settingsRes)
      setSlides(slidesRes)
      setDisplayData(displayRes)
      setLoadError(null)
    } catch (err) {
      if (err instanceof AuthError && onAuthError(err)) return
      setLoadError('Failed to load dashboard data. Try refreshing the page.')
    }
  }, [onAuthError])

  useEffect(() => {
    refreshAll()
  }, [refreshAll])

  const handleLogout = async () => {
    try {
      await api.logout()
    } catch {
      // Clear local state regardless — cookie may already be gone.
    }
    onLogout()
  }

  // A shared guard for section save handlers: catches an expired session
  // mid-action and bounces to the login screen instead of surfacing a
  // confusing generic error.
  const runGuarded = async (fn) => {
    try {
      await fn()
      await refreshAll()
      return true
    } catch (err) {
      if (err instanceof AuthError && onAuthError(err)) return false
      throw err
    }
  }

  if (loadError) {
    return (
      <Box sx={{ p: 4 }}>
        <Alert severity="error">{loadError}</Alert>
      </Box>
    )
  }

  if (!settings || !slides || !displayData) {
    return (
      <Box sx={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100vh' }}>
        <CircularProgress color="primary" />
      </Box>
    )
  }

  return (
    <Box sx={{ display: 'flex', flexDirection: 'column', minHeight: '100vh' }}>
      <Header onLogout={handleLogout} logoUrl={displayData.logo_url} />
      <Container maxWidth="md" sx={{ flex: 1, py: 4 }}>
        <Stack spacing={3}>
          <LogoUpload logoUrl={displayData.logo_url} runGuarded={runGuarded} />
          <EmergencyControls displayData={displayData} runGuarded={runGuarded} />
          <IqamahTimesForm displayData={displayData} settings={settings} runGuarded={runGuarded} />
          <SettingsForm settings={settings} runGuarded={runGuarded} />
          <SlideManager slides={slides} runGuarded={runGuarded} />
        </Stack>
      </Container>
      <Footer />
    </Box>
  )
}
