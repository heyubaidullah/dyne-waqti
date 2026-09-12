import { useState } from 'react'
import Paper from '@mui/material/Paper'
import Typography from '@mui/material/Typography'
import Stack from '@mui/material/Stack'
import Grid from '@mui/material/Grid'
import Button from '@mui/material/Button'
import Alert from '@mui/material/Alert'
import FormControlLabel from '@mui/material/FormControlLabel'
import Switch from '@mui/material/Switch'
import { api } from '../api.js'
import TimeField12h, { formatTime12h } from './TimeField12h.jsx'

const PRAYERS = [
  ['fajr', 'Fajr'],
  ['dhuhr', 'Dhuhr'],
  ['asr', 'Asr'],
  ['maghrib', 'Maghrib'],
  ['isha', 'Isha'],
]

export default function PrayerTimesForm({ prayerTimes, runGuarded }) {
  const [form, setForm] = useState(() => ({ ...prayerTimes }))
  const [jumuah2Enabled, setJumuah2Enabled] = useState(prayerTimes.jumuah_count === 2)
  const [jumuah1Auto, setJumuah1Auto] = useState(!prayerTimes.jumuah_1_iqamah)
  const [jumuah2Auto, setJumuah2Auto] = useState(!prayerTimes.jumuah_2_iqamah)
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState(null)
  const [saved, setSaved] = useState(false)

  const setTime = (key) => (value) => {
    setForm({ ...form, [key]: value })
    setSaved(false)
  }

  const save = async (e) => {
    e.preventDefault()
    setError(null)
    setBusy(true)
    try {
      await runGuarded(() =>
        api.updatePrayerTimes({
          ...form,
          jumuah_count: jumuah2Enabled ? 2 : 1,
          jumuah_1_iqamah: jumuah1Auto ? '' : form.jumuah_1_iqamah,
          jumuah_2_iqamah: jumuah2Auto ? '' : form.jumuah_2_iqamah,
        }),
      )
      setSaved(true)
    } catch {
      setError('Failed to save prayer times. Please try again.')
    } finally {
      setBusy(false)
    }
  }

  return (
    <Paper sx={{ p: 3 }} elevation={1}>
      <Typography variant="h6" gutterBottom>
        Prayer times
      </Typography>
      <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
        Enter the exact Azaan and Iqamah time for each prayer — whatever you type is exactly what shows up on the display, every day, until you change it here again. "Calculated" below each prayer is today's astronomical time, shown only as a reference (e.g. for Fajr and Maghrib, which shift with sunrise/sunset) — it's never applied automatically.
      </Typography>
      <form onSubmit={save}>
        <Grid container spacing={3}>
          {PRAYERS.map(([key, label]) => (
            <Grid key={key} size={{ xs: 12, sm: 6 }}>
              <Typography variant="subtitle2">{label}</Typography>
              {prayerTimes.calculated?.[key] && (
                <Typography variant="caption" color="text.secondary" display="block" sx={{ mb: 0.5 }}>
                  Calculated: {formatTime12h(prayerTimes.calculated[key])}
                </Typography>
              )}
              <Stack direction="row" spacing={2} flexWrap="wrap">
                <TimeField12h label="Azaan" value={form[`azaan_${key}_time`]} onChange={setTime(`azaan_${key}_time`)} disabled={busy} />
                <TimeField12h label="Iqamah" value={form[`iqamah_${key}_time`]} onChange={setTime(`iqamah_${key}_time`)} disabled={busy} />
              </Stack>
            </Grid>
          ))}
        </Grid>

        <Typography variant="subtitle2" sx={{ mt: 3 }}>
          Jumu'ah
        </Typography>
        <Stack direction="row" spacing={3} alignItems="center" sx={{ mt: 1 }} flexWrap="wrap">
          <Stack spacing={0.5}>
            <FormControlLabel
              control={<Switch size="small" checked={!jumuah1Auto} onChange={(e) => setJumuah1Auto(!e.target.checked)} disabled={busy} />}
              label="Set Jumu'ah 1 time"
            />
            {!jumuah1Auto && (
              <TimeField12h label="Jumu'ah 1 Iqamah" value={form.jumuah_1_iqamah} onChange={(v) => { setForm({ ...form, jumuah_1_iqamah: v }); setSaved(false) }} disabled={busy} />
            )}
          </Stack>

          <FormControlLabel
            control={<Switch size="small" checked={jumuah2Enabled} onChange={(e) => { setJumuah2Enabled(e.target.checked); setSaved(false) }} disabled={busy} />}
            label="This masjid holds a second Jumu'ah"
          />

          {jumuah2Enabled && (
            <Stack spacing={0.5}>
              <FormControlLabel
                control={<Switch size="small" checked={!jumuah2Auto} onChange={(e) => setJumuah2Auto(!e.target.checked)} disabled={busy} />}
                label="Set Jumu'ah 2 time"
              />
              {!jumuah2Auto && (
                <TimeField12h label="Jumu'ah 2 Iqamah" value={form.jumuah_2_iqamah} onChange={(v) => { setForm({ ...form, jumuah_2_iqamah: v }); setSaved(false) }} disabled={busy} />
              )}
            </Stack>
          )}
        </Stack>
        <Typography variant="caption" color="text.secondary" display="block" sx={{ mt: 0.5 }}>
          If a Jumu'ah time isn't set, it defaults to the Dhuhr Iqamah time.
        </Typography>

        <Stack direction="row" spacing={2} alignItems="center" sx={{ mt: 3 }}>
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
