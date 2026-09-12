import { useState } from 'react'
import Paper from '@mui/material/Paper'
import Typography from '@mui/material/Typography'
import Stack from '@mui/material/Stack'
import Grid from '@mui/material/Grid'
import TextField from '@mui/material/TextField'
import Button from '@mui/material/Button'
import Alert from '@mui/material/Alert'
import Tooltip from '@mui/material/Tooltip'
import IconButton from '@mui/material/IconButton'
import InfoOutlinedIcon from '@mui/icons-material/InfoOutlined'
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

function InfoTip({ text }) {
  return (
    <Tooltip title={text} arrow>
      <IconButton size="small" sx={{ p: 0.25 }} aria-label="more info">
        <InfoOutlinedIcon fontSize="inherit" />
      </IconButton>
    </Tooltip>
  )
}

export default function PrayerTimesForm({ prayerTimes, displayData, runGuarded }) {
  const [form, setForm] = useState(() => ({ ...prayerTimes }))
  const [jumuah2Enabled, setJumuah2Enabled] = useState(prayerTimes.jumuah_count === 2)
  const [jumuah1Auto, setJumuah1Auto] = useState(!prayerTimes.jumuah_1_iqamah)
  const [jumuah2Auto, setJumuah2Auto] = useState(!prayerTimes.jumuah_2_iqamah)
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
      <form onSubmit={save}>
        <Grid container spacing={2}>
          {PRAYERS.map(([key, label]) => (
            <Grid key={key} size={{ xs: 12, sm: 6 }}>
              <Typography variant="subtitle2">{label}</Typography>
              {displayData?.adhan_times?.[key] && (
                <Typography variant="caption" color="text.secondary" display="block" sx={{ mb: 0.5 }}>
                  Azaan (calculated): {formatTime12h(displayData.adhan_times[key])}
                </Typography>
              )}
              <Stack direction="row" spacing={2}>
                <TextField
                  label="Azaan offset (min)"
                  type="number"
                  value={form[`azaan_${key}_min`]}
                  onChange={setField(`azaan_${key}_min`)}
                  size="small"
                  disabled={busy}
                  sx={{ width: 170 }}
                  InputProps={{
                    endAdornment: <InfoTip text="Minutes after the calculated prayer time that Azaan is called. Leave at 0 to announce Azaan exactly at the calculated time." />,
                  }}
                />
                <TextField
                  label="Iqamah offset (min)"
                  type="number"
                  value={form[`iqamah_${key}_min`]}
                  onChange={setField(`iqamah_${key}_min`)}
                  size="small"
                  disabled={busy}
                  sx={{ width: 170 }}
                  InputProps={{
                    endAdornment: <InfoTip text="Minutes after Azaan is called that Iqamah (the second call to start the prayer) is held." />,
                  }}
                />
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
