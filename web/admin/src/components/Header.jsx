import AppBar from '@mui/material/AppBar'
import Toolbar from '@mui/material/Toolbar'
import Typography from '@mui/material/Typography'
import Button from '@mui/material/Button'
import Box from '@mui/material/Box'

export default function Header({ onLogout, logoUrl }) {
  return (
    <AppBar position="static" color="transparent" elevation={0} sx={{ borderBottom: 1, borderColor: 'divider' }}>
      <Toolbar>
        {logoUrl && (
          <Box component="img" src={logoUrl} alt="Masjid logo" sx={{ height: 32, width: 32, objectFit: 'contain', mr: 1.5 }} />
        )}
        <Typography variant="h6" component="h1" sx={{ flexGrow: 1 }}>
          Waqti Admin
        </Typography>
        <Button color="inherit" onClick={onLogout}>
          Log out
        </Button>
      </Toolbar>
    </AppBar>
  )
}
