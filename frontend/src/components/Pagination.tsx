import { Button, Stack, TextField, Typography } from '@mui/material'

interface PaginationProps {
  currentPage: number
  totalPages: number
  atStart: boolean
  atEnd: boolean
  loading: boolean
  onPrevious: () => void
  onNext: () => void
  onPageChange: (page: number) => void
}

export function Pagination({ currentPage, totalPages, atStart, atEnd, loading, onPrevious, onNext, onPageChange }: PaginationProps) {
  function handleKeyDown(e: React.KeyboardEvent<HTMLInputElement>) {
    if (e.key !== 'Enter') return
    const page = parseInt((e.target as HTMLInputElement).value)
    if (!isNaN(page) && page >= 1 && page <= totalPages) {
      onPageChange(page)
    }
  }

  return (
    <Stack direction="row" spacing={1} alignItems="center">
      <Button variant="outlined" size="small" onClick={onPrevious} disabled={atStart || loading}>Previous</Button>
      <Typography variant="body2">Page</Typography>
      <TextField key={currentPage} defaultValue={currentPage} size="small" sx={{ width: 64 }} onKeyDown={handleKeyDown} />
      <Typography variant="body2">of {totalPages}</Typography>
      <Button variant="outlined" size="small" onClick={onNext} disabled={atEnd || loading}>Next</Button>
    </Stack>
  )
}
