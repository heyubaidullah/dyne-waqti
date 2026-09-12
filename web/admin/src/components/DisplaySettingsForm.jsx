import { useState } from 'react'
import Paper from '@mui/material/Paper'
import Typography from '@mui/material/Typography'
import Stack from '@mui/material/Stack'
import TextField from '@mui/material/TextField'
import MenuItem from '@mui/material/MenuItem'
import Button from '@mui/material/Button'
import Alert from '@mui/material/Alert'
import FormControlLabel from '@mui/material/FormControlLabel'
import Switch from '@mui/material/Switch'
import { api } from '../api.js'

// A single, growing list of visual/kiosk preferences — deliberately not
// split into sub-sections, so new display prefs can be added here later
// without restructuring anything.
const FONT_SCALES = [
  ['small', 'Small'],
  ['medium', 'Medium'],
  ['large', 'Large'],
]

export default function DisplaySettingsForm({ displaySettings, runGuarded }) {
  const [form, setForm] = useState(() => ({ ...displaySettings }))
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState(null)
  const [saved, setSaved] = useState(false)

  const setField = (key) => (e) => {
    setForm({ ...form, [key]: e.target.value })
    setSaved(false)
  }
  const setToggle = (key) => (e) => {
    setForm({ ...form, [key]: e.target.checked })
    setSaved(false)
  }

  const save = async (e) => {
    e.preventDefault()
    setError(null)
    setBusy(true)
    try {
      await runGuarded(() => api.updateDisplaySettings(form))
      setSaved(true)
    } catch (err) {
      setError(err.message || 'Failed to save display settings.')
    } finally {
      setBusy(false)
    }
  }

  return (
    <Paper sx={{ p: 3 }} elevation={1}>
      <Typography variant="h6" gutterBottom>
        Display settings
      </Typography>
      <form onSubmit={save}>
        <Stack spacing={3}>
          <Stack direction="row" spacing={2} alignItems="center" flexWrap="wrap">
            <TextField
              select
              label="Prayer timings font size"
              value={form.display_font_scale}
              onChange={setField('display_font_scale')}
              size="small"
              sx={{ width: 200 }}
              disabled={busy}
            >
              {FONT_SCALES.map(([value, label]) => (
                <MenuItem key={value} value={value}>
                  {label}
                </MenuItem>
              ))}
            </TextField>

            <TextField
              label="Silence-phone screen duration after prayer (minutes)"
              type="number"
              value={form.silence_duration_after_min}
              onChange={setField('silence_duration_after_min')}
              size="small"
              sx={{ width: 320 }}
              disabled={busy}
              helperText="How long the silence-phone screen stays up after Iqamah. 1-15 minutes."
              slotProps={{ htmlInput: { min: 1, max: 15 } }}
            />
          </Stack>

          <FormControlLabel
            control={<Switch checked={form.show_gregorian_date} onChange={setToggle('show_gregorian_date')} disabled={busy} />}
            label="Show the Gregorian (English calendar) date alongside the Hijri date"
          />

          <Typography variant="subtitle2">"Powered by Waqti" banner — masjid side</Typography>
          <Stack spacing={1}>
            <TextField
              label="Masjid name"
              value={form.masjid_name}
              onChange={setField('masjid_name')}
              size="small"
              sx={{ width: 320 }}
              disabled={busy}
            />
            <FormControlLabel
              control={<Switch checked={form.show_masjid_name} onChange={setToggle('show_masjid_name')} disabled={busy} />}
              label="Show masjid name in the banner"
            />
            <FormControlLabel
              control={<Switch checked={form.show_masjid_logo_banner} onChange={setToggle('show_masjid_logo_banner')} disabled={busy} />}
              label="Show masjid logo in the banner"
            />
            <Typography variant="caption" color="text.secondary">
              The "Powered by Waqti" side of the banner is always shown and can't be turned off.
            </Typography>
          </Stack>

          <Stack spacing={1}>
            <FormControlLabel
              control={<Switch checked={form.weather_enabled} onChange={setToggle('weather_enabled')} disabled={busy} />}
              label="Show current weather on the display (uses your masjid's saved location, no internet = simply hidden)"
            />
            {form.weather_enabled && (
              <TextField
                select
                label="Temperature unit"
                value={form.weather_unit}
                onChange={setField('weather_unit')}
                size="small"
                sx={{ width: 200 }}
                disabled={busy}
              >
                <MenuItem value="F">Fahrenheit (°F)</MenuItem>
                <MenuItem value="C">Celsius (°C)</MenuItem>
              </TextField>
            )}
          </Stack>

          <Stack direction="row" spacing={2} alignItems="center">
            <Button type="submit" variant="contained" disabled={busy}>
              Save display settings
            </Button>
            {saved && <Alert severity="success" sx={{ py: 0 }}>Saved</Alert>}
            {error && <Alert severity="error" sx={{ py: 0 }}>{error}</Alert>}
          </Stack>
        </Stack>
      </form>
    </Paper>
  )
}
