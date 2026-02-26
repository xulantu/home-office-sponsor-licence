import { Button, Stack, Typography } from '@mui/material'
import type { User } from '../types'

interface AuthBarProps {
  user: User | null
  onLogout: () => Promise<void>
}

export function AuthBar({ user, onLogout }: AuthBarProps) {
  if (user) {
    return (
      <Stack direction="row" spacing={1} alignItems="center">
        <Typography variant="body2">{user.username}</Typography>
        <Button variant="outlined" size="small" color="error" onClick={onLogout}>Logout</Button>
      </Stack>
    )
  }

  return (
    <Button variant="outlined" size="small" color="inherit" href="/login.html">Login</Button>
  )
}
