import { TableCell, TableRow } from '@mui/material'
import type { Licence, Organisation } from '../types'

function formatDate(isoString: string | null, fallback: string): string {
  if (!isoString) return `Before ${fallback.slice(0, 10)}`
  return isoString.slice(0, 10)
}

interface OrganisationRowProps {
  org: Organisation
  licences: Licence[]
  rowIndex: number
  initialRunTime: string
}

export function OrganisationRow({ org, licences, rowIndex, initialRunTime }: OrganisationRowProps) {
  return licences.map((lic, i) => (
    <TableRow key={lic.ID}>
      {i === 0 && <TableCell rowSpan={licences.length}>{rowIndex}</TableCell>}
      {i === 0 && <TableCell rowSpan={licences.length}>{org.Name}</TableCell>}
      {i === 0 && <TableCell rowSpan={licences.length}>{org.TownCity}</TableCell>}
      {i === 0 && <TableCell rowSpan={licences.length}>{formatDate(org.CreatedAt, initialRunTime)}</TableCell>}
      <TableCell>{lic.LicenceType}</TableCell>
      <TableCell>{lic.Rating}</TableCell>
      <TableCell>{lic.Route}</TableCell>
      <TableCell>{formatDate(lic.ValidFrom, initialRunTime)}</TableCell>
    </TableRow>
  ))
}
