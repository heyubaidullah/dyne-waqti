import Paper from '@mui/material/Paper'
import Typography from '@mui/material/Typography'
import Divider from '@mui/material/Divider'
import SlideList from './SlideList.jsx'
import SlideUploadForm from './SlideUploadForm.jsx'

export default function SlideManager({ slides, runGuarded }) {
  return (
    <Paper sx={{ p: 3 }} elevation={1}>
      <Typography variant="h6" gutterBottom>
        Flyers & announcements
      </Typography>
      <SlideList slides={slides} runGuarded={runGuarded} />
      <Divider sx={{ my: 3 }} />
      <Typography variant="subtitle1" gutterBottom>
        Add a slide
      </Typography>
      <SlideUploadForm runGuarded={runGuarded} />
    </Paper>
  )
}
