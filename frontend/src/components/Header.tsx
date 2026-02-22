import { AppBar, Toolbar, Typography, Box } from '@mui/material'
import { AuthBar } from './AuthBar'
import type { User } from '../types'

interface HeaderProps {
  user: User | null
  onLogin: (username: string, password: string) => Promise<void>
  onLogout: () => Promise<void>
}

export function Header({ user, onLogin, onLogout }: HeaderProps) {
  return (
    <AppBar position="static">
      <Toolbar>
        <Typography variant="h6" component="div" sx={{ flexGrow: 1 }}>
          UK Sponsor Licence Tracker
        </Typography>
        <Box>
          <AuthBar user={user} onLogin={onLogin} onLogout={onLogout} />
        </Box>
      </Toolbar>
    </AppBar>
  )
}
