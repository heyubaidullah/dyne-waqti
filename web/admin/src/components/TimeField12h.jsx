import Stack from '@mui/material/Stack'
import TextField from '@mui/material/TextField'
import MenuItem from '@mui/material/MenuItem'
import Typography from '@mui/material/Typography'

const HOURS = Array.from({ length: 12 }, (_, i) => i + 1) // 1..12
const MINUTES = Array.from({ length: 60 }, (_, i) => i) // 0..59

// Converts a 24h "HH:MM" string to {hour12, minute, period}. Falls back to
// 12:00 AM for an empty/unparseable value so the dropdowns always have a
// valid selection.
function from24h(hhmm) {
  const match = /^(\d{1,2}):(\d{2})$/.exec(hhmm || '')
  if (!match) return { hour12: 12, minute: 0, period: 'AM' }
  const h = Number(match[1])
  const minute = Number(match[2])
  const period = h >= 12 ? 'PM' : 'AM'
  let hour12 = h % 12
  if (hour12 === 0) hour12 = 12
  return { hour12, minute, period }
}

function to24h(hour12, minute, period) {
  let h = hour12 % 12
  if (period === 'PM') h += 12
  return `${String(h).padStart(2, '0')}:${String(minute).padStart(2, '0')}`
}

// Read-only 12h display string (e.g. "5:03 AM") for a 24h "HH:MM" value —
// used wherever a time needs to be *shown* without the full picker, such
// as the Azaan reference next to each editable Iqamah field.
export function formatTime12h(hhmm) {
  const { hour12, minute, period } = from24h(hhmm)
  return `${hour12}:${String(minute).padStart(2, '0')} ${period}`
}

// Always renders/edits in 12-hour AM/PM form regardless of the device's
// browser/OS locale (unlike native <input type="time">, whose 12h vs 24h
// display isn't controllable from the page) — value/onChange still speak
// the 24h "HH:MM" string the API expects, so callers don't change.
export default function TimeField12h({ label, value, onChange, disabled }) {
  const { hour12, minute, period } = from24h(value)

  const update = (nextHour12, nextMinute, nextPeriod) => {
    onChange(to24h(nextHour12, nextMinute, nextPeriod))
  }

  return (
    <Stack spacing={0.5}>
      <Typography variant="caption" color="text.secondary">
        {label}
      </Typography>
      <Stack direction="row" spacing={1}>
        <TextField
          select
          size="small"
          value={hour12}
          disabled={disabled}
          onChange={(e) => update(Number(e.target.value), minute, period)}
          sx={{ width: 80 }}
        >
          {HOURS.map((h) => (
            <MenuItem key={h} value={h}>
              {h}
            </MenuItem>
          ))}
        </TextField>
        <TextField
          select
          size="small"
          value={minute}
          disabled={disabled}
          onChange={(e) => update(hour12, Number(e.target.value), period)}
          sx={{ width: 90 }}
        >
          {MINUTES.map((m) => (
            <MenuItem key={m} value={m}>
              {String(m).padStart(2, '0')}
            </MenuItem>
          ))}
        </TextField>
        <TextField
          select
          size="small"
          value={period}
          disabled={disabled}
          onChange={(e) => update(hour12, minute, e.target.value)}
          sx={{ width: 90 }}
        >
          <MenuItem value="AM">AM</MenuItem>
          <MenuItem value="PM">PM</MenuItem>
        </TextField>
      </Stack>
    </Stack>
  )
}
