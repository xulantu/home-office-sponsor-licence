import { Button, Stack } from '@mui/material'
import SyncIcon from '@mui/icons-material/Sync'

interface AdminToolbarProps {
  syncing: boolean
  onSync: () => void
}

export function AdminToolbar({ syncing, onSync }: AdminToolbarProps) {
  return (
    <Stack direction="row" spacing={1} sx={{ mb: 1 }}>
      <Button variant="contained" startIcon={<SyncIcon />} onClick={onSync} disabled={syncing}>
        {syncing ? 'Syncing...' : 'Sync Now'}
      </Button>
    </Stack>
  )
}
