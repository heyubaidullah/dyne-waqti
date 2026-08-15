import { useState } from 'react'
import Paper from '@mui/material/Paper'
import Typography from '@mui/material/Typography'
import Stack from '@mui/material/Stack'
import Grid from '@mui/material/Grid'
import TextField from '@mui/material/TextField'
import MenuItem from '@mui/material/MenuItem'
import Autocomplete from '@mui/material/Autocomplete'
import Button from '@mui/material/Button'
import Alert from '@mui/material/Alert'
import { api } from '../api.js'

// A curated set of common IANA timezones so non-technical staff don't need
// to know exact zone naming. Not exhaustive — the field stays freeSolo, so
// any valid IANA name can still be typed directly if a location isn't
// listed here; the backend (time.LoadLocation) is the real validator.
const TIMEZONES = [
  { value: 'America/Los_Angeles', label: 'Pacific Time — Los Angeles' },
  { value: 'America/Denver', label: 'Mountain Time — Denver' },
  { value: 'America/Phoenix', label: 'Arizona (no DST) — Phoenix' },
  { value: 'America/Chicago', label: 'Central Time — Chicago' },
  { value: 'America/New_York', label: 'Eastern Time — New York' },
  { value: 'America/Anchorage', label: 'Alaska Time — Anchorage' },
  { value: 'Pacific/Honolulu', label: 'Hawaii Time (no DST) — Honolulu' },
  { value: 'America/Toronto', label: 'Eastern Time — Toronto' },
  { value: 'America/Vancouver', label: 'Pacific Time — Vancouver' },
  { value: 'America/Regina', label: 'Saskatchewan (no DST) — Regina' },
  { value: 'Europe/London', label: 'UK Time — London' },
  { value: 'Europe/Paris', label: 'Central European Time — Paris' },
  { value: 'Europe/Berlin', label: 'Central European Time — Berlin' },
  { value: 'Europe/Istanbul', label: 'Turkey Time — Istanbul' },
  { value: 'Asia/Dubai', label: 'Gulf Time — Dubai' },
  { value: 'Asia/Riyadh', label: 'Arabia Time — Riyadh' },
  { value: 'Asia/Karachi', label: 'Pakistan Time — Karachi' },
  { value: 'Asia/Kolkata', label: 'India Time — Kolkata' },
  { value: 'Asia/Dhaka', label: 'Bangladesh Time — Dhaka' },
  { value: 'Asia/Jakarta', label: 'Indonesia Western Time — Jakarta' },
  { value: 'Asia/Kuala_Lumpur', label: 'Malaysia Time — Kuala Lumpur' },
  { value: 'Australia/Sydney', label: 'Australia Eastern Time — Sydney' },
  { value: 'UTC', label: 'UTC (no offset)' },
]

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
    } catch (err) {
      setError(err.message || 'Failed to save settings.')
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
            <Autocomplete
              freeSolo
              options={TIMEZONES}
              getOptionLabel={(opt) => (typeof opt === 'string' ? opt : opt.label)}
              inputValue={form.timezone}
              onInputChange={(e, newInputValue, reason) => {
                // Only free typing should write straight through — a
                // dropdown selection fires this too (reason 'reset', with
                // the friendly label as newInputValue), and onChange below
                // is the one that should win for that case.
                if (reason !== 'input') return
                setForm((f) => ({ ...f, timezone: newInputValue }))
                setSaved(false)
              }}
              onChange={(e, newValue) => {
                if (newValue && typeof newValue !== 'string') {
                  setForm((f) => ({ ...f, timezone: newValue.value }))
                  setSaved(false)
                }
              }}
              renderOption={(props, option) => (
                <li {...props} key={option.value}>
                  {option.label}
                  <span style={{ opacity: 0.6, marginLeft: 8 }}>{option.value}</span>
                </li>
              )}
              renderInput={(params) => (
                <TextField {...params} label="Timezone" size="small" fullWidth placeholder="e.g. America/Chicago" />
              )}
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

        <Stack direction="row" spacing={2} sx={{ mt: 2 }}>
          <TextField
            label="Masjid logo height (pixels)"
            type="number"
            value={form.logo_height_px}
            onChange={setField('logo_height_px')}
            size="small"
            sx={{ width: 260 }}
            helperText="Controls how big the uploaded logo appears on the display — every logo's natural proportions differ, so there's no fixed size"
          />
          <TextField
            label="Timings page duration (seconds)"
            type="number"
            value={form.timings_duration_sec}
            onChange={setField('timings_duration_sec')}
            size="small"
            sx={{ width: 260 }}
            helperText="How long the full-screen prayer-times page shows after an image flyer (text slides show their own banner instead and skip this page)"
          />
        </Stack>

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
