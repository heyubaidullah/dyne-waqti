import AppBar from '@mui/material/AppBar'
import Toolbar from '@mui/material/Toolbar'
import Typography from '@mui/material/Typography'
import Button from '@mui/material/Button'

export default function Header({ onLogout }) {
  return (
    <AppBar position="static" color="transparent" elevation={0} sx={{ borderBottom: 1, borderColor: 'divider' }}>
      <Toolbar>
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
