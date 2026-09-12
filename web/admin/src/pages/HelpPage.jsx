import Box from '@mui/material/Box'
import Container from '@mui/material/Container'
import Stack from '@mui/material/Stack'
import Paper from '@mui/material/Paper'
import Typography from '@mui/material/Typography'
import Button from '@mui/material/Button'
import Header from '../components/Header.jsx'
import Footer from '../components/Footer.jsx'

const FAQS = [
  {
    q: 'How do I set my timezone and location correctly?',
    a: "In Location & calculation settings, pick your city from the Timezone dropdown (or type any standard timezone name, e.g. America/Chicago). Then enter your masjid's Latitude and Longitude — these are used together to calculate accurate prayer times, so both need to be correct.",
  },
  {
    q: 'How do the Azaan and Iqamah times work?',
    a: "In the Prayer times section, type the exact time Azaan is called and the exact time Iqamah starts, for Fajr, Dhuhr, Asr, and Isha — whatever you enter is exactly what shows on the display, every day, until you come back and change it. \"Calculated\" underneath each prayer is only a reference (today's astronomical time from your location) to help you decide what to set — it's never applied automatically. Because sunrise shifts through the year, Fajr especially may need updating every few weeks; the others drift more slowly. Maghrib works differently: since sunset moves too much day to day for a fixed time to stay accurate, its Azaan always matches the calculated sunset automatically — you only set how many minutes after it Iqamah is held.",
  },
  {
    q: "What does the Jumu'ah 1 / 2 toggle do?",
    a: "Most masjids hold one Jumu'ah prayer, but some hold two. Turn on \"This masjid holds a second Jumu'ah\" to add a second time slot to the display. Turning it back off removes the second time from the display immediately — the first Jumu'ah time is unaffected either way.",
  },
  {
    q: 'The prayer timings on the display look wrong — what should I check?',
    a: "First check the Prayer times section — since these are exact times you set yourself, a wrong time there is usually just due for an update (compare against the \"Calculated\" reference shown under each prayer). Next check Location & calculation settings: timezone, latitude/longitude, and calculation method, which drive that calculated reference. If the Hijri date looks off by a day, use the Hijri date adjustment field there too. If a time you set changes on its own with nothing touched, please let support know — that specific symptom is a bug we've fixed but want to keep hearing about.",
  },
  {
    q: 'What does Full Screen vs In Screen mean for flyers?',
    a: "Full Screen fills the entire display with your flyer image for its set duration, then shows the full prayer-times page. In Screen shows your flyer in the upper part of the screen with the prayer-times ribbon visible underneath at the same time, and never switches to a separate full-screen timings page.",
  },
  {
    q: 'What if we forget the admin passphrase?',
    a: 'On the login screen, click "Forgot the passphrase?" for instructions. Whoever manages the computer running Waqti can run Reset-Passphrase.bat (Windows) or start Waqti with --reset-passphrase, which prints a brand-new passphrase without touching any settings, prayer times, or slides.',
  },
]

export default function HelpPage({ onBack, onLogout }) {
  return (
    <Box sx={{ display: 'flex', flexDirection: 'column', minHeight: '100vh' }}>
      <Header onLogout={onLogout} />
      <Container maxWidth="md" sx={{ flex: 1, py: 4 }}>
        <Stack spacing={2}>
          <Button onClick={onBack} sx={{ alignSelf: 'flex-start' }}>
            ← Back to dashboard
          </Button>
          <Typography variant="h5" gutterBottom>
            Help & troubleshooting
          </Typography>
          {FAQS.map(({ q, a }) => (
            <Paper key={q} sx={{ p: 2.5 }} elevation={1}>
              <Typography variant="subtitle1" gutterBottom>
                {q}
              </Typography>
              <Typography variant="body2" color="text.secondary">
                {a}
              </Typography>
            </Paper>
          ))}
        </Stack>
      </Container>
      <Footer />
    </Box>
  )
}
