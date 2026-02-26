import { AppBar, Toolbar, Typography, Box } from '@mui/material'
import { AuthBar } from './AuthBar'
import type { User } from '../types'

interface HeaderProps {
  user: User | null
  onLogout: () => Promise<void>
}

export function Header({ user, onLogout }: HeaderProps) {
  return (
    <AppBar position="static">
      <Toolbar>
        <Typography variant="h6" component="div" sx={{ flexGrow: 1 }}>
          UK Sponsor Licence Tracker
        </Typography>
        <Box>
          <AuthBar user={user} onLogout={onLogout} />
        </Box>
      </Toolbar>
    </AppBar>
  )
}
