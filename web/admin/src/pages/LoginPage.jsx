import { useState } from 'react'
import Box from '@mui/material/Box'
import Paper from '@mui/material/Paper'
import TextField from '@mui/material/TextField'
import Button from '@mui/material/Button'
import Typography from '@mui/material/Typography'
import Alert from '@mui/material/Alert'
import { api, RateLimitError } from '../api.js'
import Footer from '../components/Footer.jsx'

export default function LoginPage({ onLogin }) {
  const [passphrase, setPassphrase] = useState('')
  const [error, setError] = useState(null)
  const [submitting, setSubmitting] = useState(false)

  const handleSubmit = async (e) => {
    e.preventDefault()
    setError(null)
    setSubmitting(true)
    try {
      await api.login(passphrase)
      onLogin()
    } catch (err) {
      if (err instanceof RateLimitError) {
        const minutes = err.retryAfterSeconds ? Math.ceil(err.retryAfterSeconds / 60) : null
        setError(minutes ? `Too many attempts. Try again in about ${minutes} minute(s).` : 'Too many attempts. Try again later.')
      } else {
        setError('Incorrect passphrase.')
      }
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <Box sx={{ display: 'flex', flexDirection: 'column', minHeight: '100vh' }}>
      <Box sx={{ flex: 1, display: 'flex', justifyContent: 'center', alignItems: 'center', p: 2 }}>
        <Paper sx={{ p: 4, width: '100%', maxWidth: 360 }} elevation={3}>
          <Typography variant="h5" component="h1" gutterBottom>
            Waqti Admin
          </Typography>
          <Typography variant="body2" color="text.secondary" sx={{ mb: 3 }}>
            Enter the admin passphrase to continue.
          </Typography>
          <form onSubmit={handleSubmit}>
            <TextField
              type="password"
              label="Passphrase"
              value={passphrase}
              onChange={(e) => setPassphrase(e.target.value)}
              fullWidth
              autoFocus
              margin="normal"
            />
            {error && (
              <Alert severity="error" sx={{ mt: 1 }}>
                {error}
              </Alert>
            )}
            <Button type="submit" variant="contained" fullWidth sx={{ mt: 2 }} disabled={submitting || !passphrase}>
              {submitting ? 'Logging in…' : 'Log in'}
            </Button>
          </form>
        </Paper>
      </Box>
      <Footer />
    </Box>
  )
}
