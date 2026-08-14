import { useState } from 'react'
import Paper from '@mui/material/Paper'
import Typography from '@mui/material/Typography'
import Stack from '@mui/material/Stack'
import Grid from '@mui/material/Grid'
import TextField from '@mui/material/TextField'
import MenuItem from '@mui/material/MenuItem'
import Button from '@mui/material/Button'
import Alert from '@mui/material/Alert'
import { api } from '../api.js'

const CALC_METHODS = [
  ['MWL', 'Muslim World League'],
  ['ISNA', 'ISNA (North America)'],
  ['EGYPTIAN', 'Egyptian General Authority'],
  ['KARACHI', 'University of Islamic Sciences, Karachi'],
  ['UMM_AL_QURA', 'Umm al-Qura, Makkah'],
]

const ASR_METHODS = [
  ['SHAFI', 'Standard (Shafi/Hanbali/Maliki)'],
  ['HANAFI', 'Hanafi'],
]

const OFFSET_FIELDS = [
  ['iqamah_fajr_min', 'Fajr'],
  ['iqamah_dhuhr_min', 'Dhuhr'],
  ['iqamah_asr_min', 'Asr'],
  ['iqamah_maghrib_min', 'Maghrib'],
  ['iqamah_isha_min', 'Isha'],
]

export default function SettingsForm({ settings, runGuarded }) {
  const [form, setForm] = useState(() => ({ ...settings }))
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState(null)
  const [saved, setSaved] = useState(false)

  const setField = (key) => (e) => {
    setForm({ ...form, [key]: e.target.value })
    setSaved(false)
  }

  const save = async (e) => {
    e.preventDefault()
    setError(null)
    setBusy(true)
    try {
      await runGuarded(() => api.updateSettings(form))
      setSaved(true)
    } catch {
      setError('Failed to save. Check the timezone name and that all numbers are valid.')
    } finally {
      setBusy(false)
    }
  }

  return (
    <Paper sx={{ p: 3 }} elevation={1}>
      <Typography variant="h6" gutterBottom>
        Location & calculation settings
      </Typography>
      <form onSubmit={save}>
        <Grid container spacing={2}>
          <Grid size={{ xs: 12, sm: 6 }}>
            <TextField
              label="Timezone (IANA name)"
              value={form.timezone}
              onChange={setField('timezone')}
              placeholder="e.g. America/New_York"
              size="small"
              fullWidth
            />
          </Grid>
          <Grid size={{ xs: 6, sm: 3 }}>
            <TextField label="Latitude" value={form.latitude} onChange={setField('latitude')} size="small" fullWidth />
          </Grid>
          <Grid size={{ xs: 6, sm: 3 }}>
            <TextField label="Longitude" value={form.longitude} onChange={setField('longitude')} size="small" fullWidth />
          </Grid>
          <Grid size={{ xs: 6, sm: 6 }}>
            <TextField select label="Calculation method" value={form.calc_method} onChange={setField('calc_method')} size="small" fullWidth>
              {CALC_METHODS.map(([value, label]) => (
                <MenuItem key={value} value={value}>
                  {label}
                </MenuItem>
              ))}
            </TextField>
          </Grid>
          <Grid size={{ xs: 6, sm: 6 }}>
            <TextField select label="Asr juristic method" value={form.asr_method} onChange={setField('asr_method')} size="small" fullWidth>
              {ASR_METHODS.map(([value, label]) => (
                <MenuItem key={value} value={value}>
                  {label}
                </MenuItem>
              ))}
            </TextField>
          </Grid>
        </Grid>

        <Typography variant="subtitle2" sx={{ mt: 3, mb: 1 }}>
          Default Iqamah offsets (minutes after Adhan, used when no daily override is saved)
        </Typography>
        <Grid container spacing={2}>
          {OFFSET_FIELDS.map(([key, label]) => (
            <Grid key={key} size={{ xs: 6, sm: 2.4 }}>
              <TextField label={label} type="number" value={form[key]} onChange={setField(key)} size="small" fullWidth />
            </Grid>
          ))}
        </Grid>

        <Stack direction="row" spacing={2} alignItems="center" sx={{ mt: 2 }}>
          <Button type="submit" variant="contained" disabled={busy}>
            Save settings
          </Button>
          {saved && <Alert severity="success" sx={{ py: 0 }}>Saved</Alert>}
          {error && <Alert severity="error" sx={{ py: 0 }}>{error}</Alert>}
        </Stack>
      </form>
    </Paper>
  )
}
