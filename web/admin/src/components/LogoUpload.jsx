import { useState } from 'react'
import Paper from '@mui/material/Paper'
import Typography from '@mui/material/Typography'
import Stack from '@mui/material/Stack'
import Button from '@mui/material/Button'
import Alert from '@mui/material/Alert'
import Box from '@mui/material/Box'
import { api } from '../api.js'

export default function LogoUpload({ logoUrl, runGuarded }) {
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState(null)

  const upload = async (file) => {
    setError(null)
    setBusy(true)
    try {
      const data = new FormData()
      data.set('file', file)
      await runGuarded(() => api.uploadLogo(data))
    } catch (err) {
      setError(err.message || 'Upload failed.')
    } finally {
      setBusy(false)
    }
  }

  const remove = async () => {
    setBusy(true)
    try {
      await runGuarded(() => api.deleteLogo())
    } finally {
      setBusy(false)
    }
  }

  return (
    <Paper sx={{ p: 3 }} elevation={1}>
      <Typography variant="h6" gutterBottom>
        Masjid logo
      </Typography>
      <Stack direction="row" spacing={2} alignItems="center">
        {logoUrl ? (
          <Box
            component="img"
            src={logoUrl}
            alt="Masjid logo"
            sx={{ height: 64, width: 64, objectFit: 'contain', bgcolor: 'background.default', borderRadius: 1, p: 0.5 }}
          />
        ) : (
          <Typography variant="body2" color="text.secondary">
            No logo uploaded yet.
          </Typography>
        )}
        <Button variant="outlined" component="label" disabled={busy}>
          {logoUrl ? 'Replace logo' : 'Upload logo'}
          <input
            type="file"
            accept="image/png,image/jpeg"
            hidden
            onChange={(e) => {
              const file = e.target.files?.[0]
              if (file) upload(file)
              e.target.value = ''
            }}
          />
        </Button>
        {logoUrl && (
          <Button variant="outlined" color="error" disabled={busy} onClick={remove}>
            Remove
          </Button>
        )}
      </Stack>
      {error && (
        <Alert severity="error" sx={{ mt: 2 }}>
          {error}
        </Alert>
      )}
    </Paper>
  )
}
