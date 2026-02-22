import { Box, Table, TableBody, TableCell, TableContainer, TableHead, TableRow, Typography, Paper } from '@mui/material'
import { OrganisationRow } from './OrganisationRow'
import { Pagination } from './Pagination'
import type { DataResponse } from '../types'

interface DataTableProps {
  data: DataResponse
  loading: boolean
  currentPage: number
  totalPages: number
  atStart: boolean
  atEnd: boolean
  onPrevious: () => void
  onNext: () => void
  onPageChange: (page: number) => void
}

export function DataTable({ data, loading, currentPage, totalPages, atStart, atEnd, onPrevious, onNext, onPageChange }: DataTableProps) {
  const total = data.total_organisations
  const licencesByOrg = new Map(
    data.organisations.map(org => [
      org.ID,
      data.licences.filter(l => l.OrganisationID === org.ID),
    ])
  )

  const paginationProps = { currentPage, totalPages, atStart, atEnd, loading, onPrevious, onNext, onPageChange }

  return (
    <Box>
      <Typography variant="body2" sx={{ mb: 1 }}>
        Showing {data.from}–{Math.min(data.to, total)} of {total} organisations
      </Typography>
      <Pagination {...paginationProps} />
      <TableContainer component={Paper} sx={{ mt: 1, mb: 1 }}>
        <Table size="small">
          <TableHead>
            <TableRow>
              <TableCell>#</TableCell>
              <TableCell>Organisation</TableCell>
              <TableCell>Town/City</TableCell>
              <TableCell>Registered Since</TableCell>
              <TableCell>Type</TableCell>
              <TableCell>Rating</TableCell>
              <TableCell>Route</TableCell>
              <TableCell>Licence Rating Valid From</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {data.organisations.map((org, orgIndex) => (
              <OrganisationRow
                key={org.ID}
                org={org}
                licences={licencesByOrg.get(org.ID) ?? []}
                rowIndex={data.from + orgIndex}
                initialRunTime={data.initial_run_time}
              />
            ))}
          </TableBody>
        </Table>
      </TableContainer>
      <Pagination {...paginationProps} />
    </Box>
  )
}
