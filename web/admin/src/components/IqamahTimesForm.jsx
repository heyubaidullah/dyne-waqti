import { useState } from 'react'
import Paper from '@mui/material/Paper'
import Typography from '@mui/material/Typography'
import Stack from '@mui/material/Stack'
import Grid from '@mui/material/Grid'
import TextField from '@mui/material/TextField'
import Button from '@mui/material/Button'
import Alert from '@mui/material/Alert'
import { api } from '../api.js'
import TimeField12h from './TimeField12h.jsx'

const PRAYERS = [
  ['fajr', 'Fajr'],
  ['dhuhr', 'Dhuhr'],
  ['asr', 'Asr'],
  ['maghrib', 'Maghrib'],
  ['isha', 'Isha'],
  ['jumuah', "Jumu'ah"],
]

export default function IqamahTimesForm({ displayData, settings, runGuarded }) {
  const [times, setTimes] = useState(() => ({ ...displayData.iqamah_times }))
  const [hijriAdjustDays, setHijriAdjustDays] = useState(settings.hijri_adjust_days)
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState(null)
  const [saved, setSaved] = useState(false)

  const setTime = (key) => (newValue) => {
    setTimes({ ...times, [key]: newValue })
    setSaved(false)
  }

  const save = async (e) => {
    e.preventDefault()
    setError(null)
    setBusy(true)
    try {
      await runGuarded(() =>
        api.updatePrayerTimes({
          fajr_iqamah: times.fajr,
          dhuhr_iqamah: times.dhuhr,
          asr_iqamah: times.asr,
          maghrib_iqamah: times.maghrib,
          isha_iqamah: times.isha,
          jumuah_iqamah: times.jumuah,
          hijri_adjust_days: Number(hijriAdjustDays),
        }),
      )
      setSaved(true)
    } catch {
      setError('Failed to save Iqamah times. Please try again.')
    } finally {
      setBusy(false)
    }
  }

  return (
    <Paper sx={{ p: 3 }} elevation={1}>
      <Typography variant="h6" gutterBottom>
        Today's Iqamah times
      </Typography>
      <form onSubmit={save}>
        <Grid container spacing={2}>
          {PRAYERS.map(([key, label]) => (
            <Grid key={key} size={{ xs: 6, sm: 4 }}>
              <TimeField12h label={label} value={times[key]} onChange={setTime(key)} disabled={busy} />
            </Grid>
          ))}
        </Grid>
        <TextField
          label="Hijri date adjustment (days)"
          type="number"
          value={hijriAdjustDays}
          onChange={(e) => setHijriAdjustDays(e.target.value)}
          size="small"
          sx={{ mt: 2, width: 240 }}
          helperText="For local moon-sighting alignment, e.g. -1 or +1"
        />
        <Stack direction="row" spacing={2} alignItems="center" sx={{ mt: 2 }}>
          <Button type="submit" variant="contained" disabled={busy}>
            Save & push to screens
          </Button>
          {saved && <Alert severity="success" sx={{ py: 0 }}>Saved</Alert>}
          {error && <Alert severity="error" sx={{ py: 0 }}>{error}</Alert>}
        </Stack>
      </form>
    </Paper>
  )
}
