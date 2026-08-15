import { useState } from 'react'
import Stack from '@mui/material/Stack'
import Typography from '@mui/material/Typography'
import TextField from '@mui/material/TextField'
import Button from '@mui/material/Button'
import Alert from '@mui/material/Alert'
import RadioGroup from '@mui/material/RadioGroup'
import FormControlLabel from '@mui/material/FormControlLabel'
import Radio from '@mui/material/Radio'
import { api } from '../api.js'

const emptyForm = {
  title: '',
  type: 'text_verse',
  content: '',
  arabic_text: '',
  expiration_date: '',
  display_duration_sec: '10',
}

export default function SlideUploadForm({ runGuarded }) {
  const [form, setForm] = useState(emptyForm)
  const [file, setFile] = useState(null)
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState(null)

  const setField = (key) => (e) => setForm({ ...form, [key]: e.target.value })

  const submit = async (e) => {
    e.preventDefault()
    setError(null)

    if (form.type === 'image' && !file) {
      setError('Choose an image file to upload.')
      return
    }

    const data = new FormData()
    data.set('title', form.title)
    data.set('type', form.type)
    data.set('display_duration_sec', form.display_duration_sec || '10')
    if (form.arabic_text) data.set('arabic_text', form.arabic_text)
    if (form.expiration_date) data.set('expiration_date', form.expiration_date)
    if (form.type === 'image') {
      data.set('file', file)
    } else {
      data.set('content', form.content)
    }

    setBusy(true)
    try {
      await runGuarded(() => api.uploadSlide(data))
      setForm(emptyForm)
      setFile(null)
      e.target.reset()
    } catch (err) {
      setError(err.message || 'Upload failed.')
    } finally {
      setBusy(false)
    }
  }

  return (
    <form onSubmit={submit}>
      <Stack spacing={2}>
        <TextField label="Title" value={form.title} onChange={setField('title')} required size="small" />

        <RadioGroup row value={form.type} onChange={setField('type')}>
          <FormControlLabel value="text_verse" control={<Radio />} label="Verse / Hadith text" />
          <FormControlLabel value="image" control={<Radio />} label="Image flyer" />
        </RadioGroup>

        {form.type === 'image' ? (
          <Stack spacing={0.5} alignItems="flex-start">
            <Button variant="outlined" component="label">
              {file ? file.name : 'Choose image (PNG/JPEG, max 10MB)'}
              <input type="file" accept="image/png,image/jpeg" hidden onChange={(e) => setFile(e.target.files?.[0] ?? null)} />
            </Button>
            <Typography variant="caption" color="text.secondary">
              For best results, use a 16:9 image (1920×1080px recommended) — it fills the whole screen when shown.
            </Typography>
          </Stack>
        ) : (
          <TextField
            label="Verse / Hadith text"
            value={form.content}
            onChange={setField('content')}
            required
            multiline
            minRows={2}
            size="small"
          />
        )}

        <TextField
          label="Arabic text (optional)"
          value={form.arabic_text}
          onChange={setField('arabic_text')}
          size="small"
          slotProps={{ htmlInput: { dir: 'rtl', style: { fontFamily: 'Amiri, "Noto Naskh Arabic", serif' } } }}
        />

        <Stack direction="row" spacing={2}>
          <TextField
            label="Expiration date (optional)"
            type="date"
            value={form.expiration_date}
            onChange={setField('expiration_date')}
            size="small"
            slotProps={{ inputLabel: { shrink: true } }}
          />
          <TextField
            label="Display seconds"
            type="number"
            value={form.display_duration_sec}
            onChange={setField('display_duration_sec')}
            size="small"
            sx={{ width: 140 }}
          />
        </Stack>

        {error && <Alert severity="error">{error}</Alert>}

        <Button type="submit" variant="contained" disabled={busy} sx={{ alignSelf: 'flex-start' }}>
          {busy ? 'Uploading…' : 'Add slide'}
        </Button>
      </Stack>
    </form>
  )
}
