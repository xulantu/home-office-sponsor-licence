import { useRef, useState } from 'react'
import { Alert, Button, Stack, TextField, Typography } from '@mui/material'
import type { User } from '../types'

interface AuthBarProps {
  user: User | null
  onLogin: (username: string, password: string) => Promise<void>
  onLogout: () => Promise<void>
}

export function AuthBar({ user, onLogin, onLogout }: AuthBarProps) {
  const usernameRef = useRef<HTMLInputElement>(null)
  const passwordRef = useRef<HTMLInputElement>(null)
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState(false)

  async function handleLogin() {
    const username = usernameRef.current?.value ?? ''
    const password = passwordRef.current?.value ?? ''
    setError(null)
    setLoading(true)
    try {
      await onLogin(username, password)
    } catch {
      setError('Invalid credentials')
    } finally {
      setLoading(false)
    }
  }

  if (user) {
    return (
      <Stack direction="row" spacing={1} alignItems="center">
        <Typography variant="body2">{user.username}</Typography>
        <Button variant="outlined" size="small" color="error" onClick={onLogout}>Logout</Button>
      </Stack>
    )
  }

  return (
    <Stack direction="row" spacing={1} alignItems="center">
      <TextField inputRef={usernameRef} size="small" label="Username" />
      <TextField inputRef={passwordRef} size="small" label="Password" type="password"
        onKeyDown={e => { if (e.key === 'Enter') handleLogin() }} />
      <Button variant="contained" size="small" onClick={handleLogin} disabled={loading}>Login</Button>
      {error && <Alert severity="error" sx={{ py: 0 }}>{error}</Alert>}
    </Stack>
  )
}
