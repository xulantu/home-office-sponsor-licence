import { useState, useEffect } from 'react'
import './App.css'

interface Organisation {
  ID: number
  Name: string
  TownCity: string
  County: string
  CreatedAt: string | null
}

interface Licence {
  ID: number
  OrganisationID: number
  LicenceType: string
  Rating: string
  Route: string
  ValidFrom: string | null
}

interface DataResponse {
  initial_run_time: string
  total_organisations: number
  from: number
  to: number
  organisations: Organisation[]
  licences: Licence[]
}

const PAGE_SIZE = 20

function formatDate(isoString: string | null, fallback: string): string {
  if (!isoString) return `Before ${fallback.slice(0, 10)}`
  return isoString.slice(0, 10)
}

function App() {
  const [data, setData] = useState<DataResponse | null>(null)
  const [from, setFrom] = useState(1)
  const [to, setTo] = useState(PAGE_SIZE)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [syncing, setSyncing] = useState(false)
  const [search, setSearch] = useState('')

  async function fetchData(reqFrom: number, reqTo: number, reqSearch: string) {
    setLoading(true)
    setError(null)
    try {
      const response = await fetch(`/api/data?from=${reqFrom}&to=${reqTo}&search=${encodeURIComponent(reqSearch)}`)
      if (!response.ok) throw new Error('Failed to fetch data')
      const json: DataResponse = await response.json()
      if (json.from === reqFrom && json.to === reqTo) {
        setData(json)
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error')
    } finally {
      setLoading(false)
    }
  }

  async function handleSync() {
    setSyncing(true)
    setError(null)
    try {
      const response = await fetch('/api/sync', { method: 'POST' })
      if (!response.ok) throw new Error('Sync failed')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error')
    } finally {
      setSyncing(false)
    }
  }

  function handlePrevious() {
    const newFrom = Math.max(1, from - PAGE_SIZE)
    const newTo = newFrom + PAGE_SIZE - 1
    setFrom(newFrom)
    setTo(newTo)
    fetchData(newFrom, newTo, search)
  }

  function handleNext() {
    const newFrom = from + PAGE_SIZE
    const newTo = newFrom + PAGE_SIZE - 1
    setFrom(newFrom)
    setTo(newTo)
    fetchData(newFrom, newTo, search)
  }

  useEffect(() => { fetchData(from, to, search) }, [])

  if (loading) return <p>Loading...</p>
  if (error) return <p className="error">Error: {error}</p>
  if (!data) return null

  const total = data.total_organisations
  const totalPages = Math.ceil(total / PAGE_SIZE)
  const currentPage = Math.ceil(from / PAGE_SIZE)
  const atStart = from <= 1
  const atEnd = to >= total

  return (
    <div className="container">
      <h1>UK Sponsor Licence Tracker</h1>
      <div className="controls">
        <button onClick={handleSync} disabled={syncing}>
          {syncing ? 'Syncing...' : 'Sync Now'}
        </button>
        <button onClick={() => fetchData(from, to, search)} disabled={loading}>
          Refresh
        </button>
        <input type="text" placeholder="Search by name or town..." onKeyDown={e => { if (e.key === 'Enter') { const val = (e.target as HTMLInputElement).value; setSearch(val); setFrom(1); setTo(PAGE_SIZE); fetchData(1, PAGE_SIZE, val); }}} size={30} />
      </div>
      <p>Showing {from}–{Math.min(to, total)} of {total} organisations</p>
      <div className="pagination">
        <button onClick={handlePrevious} disabled={atStart || loading}>Previous</button>
        <span>Page <input type="text" key={currentPage} defaultValue={currentPage} onKeyDown={e => { if (e.key === 'Enter') { const page = parseInt((e.target as HTMLInputElement).value); if (!isNaN(page) && page >= 1 && page <= totalPages) { const newFrom = (page - 1) * PAGE_SIZE + 1; const newTo = newFrom + PAGE_SIZE - 1; setFrom(newFrom); setTo(newTo); fetchData(newFrom, newTo, search); }}}} size={4} /> of {totalPages}</span>
        <button onClick={handleNext} disabled={atEnd || loading}>Next</button>
      </div>
      <table>
        <thead>
          <tr>
            <th>#</th>
            <th>Organisation</th>
            <th>Town/City</th>
            <th>Registered Since</th>
            <th>Type</th>
            <th>Rating</th>
            <th>Route</th>
            <th>Licence Rating Valid From</th>
          </tr>
        </thead>
        <tbody>
          {(data.organisations ?? []).map((org, orgIndex) => {
            const orgLicences = (data.licences ?? []).filter(l => l.OrganisationID === org.ID)
            return orgLicences.map((lic, i) => (
              <tr key={lic.ID}>
                {i === 0 && <td rowSpan={orgLicences.length}>{data.from + orgIndex}</td>}
                {i === 0 && <td rowSpan={orgLicences.length}>{org.Name}</td>}
                {i === 0 && <td rowSpan={orgLicences.length}>{org.TownCity}</td>}
                {i === 0 && <td rowSpan={orgLicences.length}>{formatDate(org.CreatedAt, data.initial_run_time)}</td>}
                <td>{lic.LicenceType}</td>
                <td>{lic.Rating}</td>
                <td>{lic.Route}</td>
                <td>{formatDate(lic.ValidFrom, data.initial_run_time)}</td>
              </tr>
            ))
          })}
        </tbody>
      </table>
      <div className="pagination">
        <button onClick={handlePrevious} disabled={atStart || loading}>Previous</button>
        <span>Page <input type="text" key={currentPage} defaultValue={currentPage} onKeyDown={e => { if (e.key === 'Enter') { const page = parseInt((e.target as HTMLInputElement).value); if (!isNaN(page) && page >= 1 && page <= totalPages) { const newFrom = (page - 1) * PAGE_SIZE + 1; const newTo = newFrom + PAGE_SIZE - 1; setFrom(newFrom); setTo(newTo); fetchData(newFrom, newTo, search); }}}} size={4} /> of {totalPages}</span>
        <button onClick={handleNext} disabled={atEnd || loading}>Next</button>
      </div>
    </div>
  )
}

export default App
