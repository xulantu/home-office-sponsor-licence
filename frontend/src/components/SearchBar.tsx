import { IconButton, InputAdornment, TextField, Tooltip } from '@mui/material'
import RefreshIcon from '@mui/icons-material/Refresh'
import SearchIcon from '@mui/icons-material/Search'

interface SearchBarProps {
  loading: boolean
  onSearch: (value: string) => void
  onRefresh: () => void
}

export function SearchBar({ loading, onSearch, onRefresh }: SearchBarProps) {
  function handleKeyDown(e: React.KeyboardEvent<HTMLInputElement>) {
    if (e.key === 'Enter') {
      onSearch((e.target as HTMLInputElement).value)
    }
  }

  return (
    <TextField
      placeholder="Search by name or town..."
      size="small"
      disabled={loading}
      onKeyDown={handleKeyDown}
      slotProps={{
        input: {
          startAdornment: (
            <InputAdornment position="start">
              <SearchIcon fontSize="small" />
            </InputAdornment>
          ),
          endAdornment: (
            <InputAdornment position="end">
              <Tooltip title="Refresh">
                <span>
                  <IconButton size="small" onClick={onRefresh} disabled={loading}>
                    <RefreshIcon fontSize="small" />
                  </IconButton>
                </span>
              </Tooltip>
            </InputAdornment>
          ),
        },
      }}
    />
  )
}
