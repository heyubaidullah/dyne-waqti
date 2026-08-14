import Box from '@mui/material/Box'
import Typography from '@mui/material/Typography'

export default function Footer() {
  return (
    <Box component="footer" sx={{ py: 2, textAlign: 'center' }}>
      <Typography variant="caption" color="text.secondary">
        Developed by Ubaidullah Khan at Dyne Labs (c) 2026
      </Typography>
    </Box>
  )
}
