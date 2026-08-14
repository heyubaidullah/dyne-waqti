import { useState } from 'react'
import List from '@mui/material/List'
import ListItem from '@mui/material/ListItem'
import ListItemText from '@mui/material/ListItemText'
import Switch from '@mui/material/Switch'
import TextField from '@mui/material/TextField'
import IconButton from '@mui/material/IconButton'
import Chip from '@mui/material/Chip'
import Stack from '@mui/material/Stack'
import Typography from '@mui/material/Typography'
import DeleteIcon from '@mui/icons-material/Delete'
import { api } from '../api.js'

export default function SlideList({ slides, runGuarded }) {
  const [busyId, setBusyId] = useState(null)

  if (slides.length === 0) {
    return (
      <Typography variant="body2" color="text.secondary">
        No slides yet — add one below.
      </Typography>
    )
  }

  const toggleActive = async (slide) => {
    setBusyId(slide.id)
    try {
      await runGuarded(() => api.updateSlide(slide.id, { is_active: !slide.is_active }))
    } finally {
      setBusyId(null)
    }
  }

  const setExpiration = async (slide, value) => {
    setBusyId(slide.id)
    try {
      await runGuarded(() => api.updateSlide(slide.id, { expiration_date: value }))
    } finally {
      setBusyId(null)
    }
  }

  const remove = async (slide) => {
    setBusyId(slide.id)
    try {
      await runGuarded(() => api.deleteSlide(slide.id))
    } finally {
      setBusyId(null)
    }
  }

  return (
    <List disablePadding>
      {slides.map((slide) => (
        <ListItem
          key={slide.id}
          divider
          secondaryAction={
            <IconButton edge="end" aria-label="delete" onClick={() => remove(slide)} disabled={busyId === slide.id}>
              <DeleteIcon />
            </IconButton>
          }
        >
          <Switch checked={slide.is_active} onChange={() => toggleActive(slide)} disabled={busyId === slide.id} />
          <ListItemText
            primary={
              <Stack direction="row" spacing={1} alignItems="center">
                <span>{slide.title}</span>
                <Chip size="small" label={slide.type === 'image' ? 'Image' : 'Text'} />
              </Stack>
            }
            secondary={slide.type === 'text_verse' ? slide.content_url_or_text : slide.content_url_or_text}
          />
          <TextField
            label="Expires"
            type="date"
            size="small"
            value={slide.expiration_date || ''}
            onChange={(e) => setExpiration(slide, e.target.value)}
            disabled={busyId === slide.id}
            slotProps={{ inputLabel: { shrink: true } }}
            sx={{ width: 170, mr: 2 }}
          />
        </ListItem>
      ))}
    </List>
  )
}
