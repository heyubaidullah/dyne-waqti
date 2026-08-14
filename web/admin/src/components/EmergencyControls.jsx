import { useState } from 'react'
import Paper from '@mui/material/Paper'
import Typography from '@mui/material/Typography'
import Stack from '@mui/material/Stack'
import TextField from '@mui/material/TextField'
import Button from '@mui/material/Button'
import FormControlLabel from '@mui/material/FormControlLabel'
import Switch from '@mui/material/Switch'
import Divider from '@mui/material/Divider'
import Alert from '@mui/material/Alert'
import { api } from '../api.js'

const emptyNotice = { title: '', deceased_name: '', prayer_time: '', location: '' }

export default function EmergencyControls({ displayData, runGuarded }) {
  const [notice, setNotice] = useState(emptyNotice)
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState(null)

  const active = displayData.emergency

  const setField = (field) => (e) => setNotice({ ...notice, [field]: e.target.value })

  const publish = async (e) => {
    e.preventDefault()
    setError(null)
    setBusy(true)
    try {
      await runGuarded(() => api.publishJanazah(notice))
      setNotice(emptyNotice)
    } catch {
      setError('Failed to publish notice. Check all fields are filled in.')
    } finally {
      setBusy(false)
    }
  }

  const dismiss = async () => {
    setBusy(true)
    try {
      await runGuarded(() => api.dismissJanazah())
    } finally {
      setBusy(false)
    }
  }

  const toggleBlackout = async (e) => {
    const nextActive = e.target.checked
    setBusy(true)
    try {
      await runGuarded(() => api.setBlackout(nextActive))
    } finally {
      setBusy(false)
    }
  }

  return (
    <Paper sx={{ p: 3 }} elevation={1}>
      <Typography variant="h6" gutterBottom>
        Emergency controls
      </Typography>

      <FormControlLabel
        control={<Switch checked={displayData.blackout} onChange={toggleBlackout} disabled={busy} />}
        label="Screen blackout (now)"
      />

      <Divider sx={{ my: 2 }} />

      <Typography variant="subtitle1" gutterBottom>
        Janazah notice
      </Typography>

      {active ? (
        <Stack spacing={1} sx={{ mb: 2 }}>
          <Alert severity="warning">
            Active: {active.deceased_name} — {active.prayer_time} at {active.location}
          </Alert>
          <Button variant="outlined" color="error" onClick={dismiss} disabled={busy} sx={{ alignSelf: 'flex-start' }}>
            Dismiss notice
          </Button>
        </Stack>
      ) : (
        <form onSubmit={publish}>
          <Stack spacing={2}>
            <TextField label="Title" value={notice.title} onChange={setField('title')} required size="small" />
            <TextField
              label="Deceased name"
              value={notice.deceased_name}
              onChange={setField('deceased_name')}
              required
              size="small"
            />
            <TextField
              label="Prayer time"
              value={notice.prayer_time}
              onChange={setField('prayer_time')}
              required
              size="small"
              placeholder="e.g. After Dhuhr, 1:30 PM"
            />
            <TextField label="Location" value={notice.location} onChange={setField('location')} required size="small" />
            {error && <Alert severity="error">{error}</Alert>}
            <Button type="submit" variant="contained" color="error" disabled={busy} sx={{ alignSelf: 'flex-start' }}>
              Publish Janazah notice
            </Button>
          </Stack>
        </form>
      )}
    </Paper>
  )
}
