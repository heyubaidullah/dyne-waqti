import Box from '@mui/material/Box'
import Typography from '@mui/material/Typography'
import Link from '@mui/material/Link'

export default function Footer() {
  return (
    <Box component="footer" sx={{ py: 2, textAlign: 'center' }}>
      <Typography variant="caption" color="text.secondary">
        Developed by Ubaidullah Khan at{' '}
        <Link href="https://www.dynelabs.org" target="_blank" rel="noopener noreferrer" color="inherit">
          Dyne Labs
        </Link>{' '}
        © 2026
      </Typography>
    </Box>
  )
}
