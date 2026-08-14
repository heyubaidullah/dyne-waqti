import { createTheme } from '@mui/material/styles'

// Design tokens from the Waqti spec. mode:'dark' gives every MUI component
// sane dark defaults for free; only the brand colors need overriding.
export const theme = createTheme({
  palette: {
    mode: 'dark',
    background: { default: '#0B0F17', paper: '#1E293B' },
    primary: { main: '#F59E0B', dark: '#D97706', contrastText: '#0B0F17' },
    success: { main: '#10B981' },
    error: { main: '#EF4444' },
    text: { primary: '#F8FAFC', secondary: '#94A3B8' },
  },
  typography: {
    fontFamily: '"Plus Jakarta Sans", "Inter", sans-serif',
  },
  shape: { borderRadius: 8 },
})
