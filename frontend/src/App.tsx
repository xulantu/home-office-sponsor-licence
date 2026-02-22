import { useEffect, useState } from 'react'
import { Alert, Box, Container, CssBaseline } from '@mui/material'
import { createTheme, ThemeProvider } from '@mui/material/styles'
import { Header } from './components/Header'
import { SearchBar } from './components/SearchBar'
import { AdminToolbar } from './components/AdminToolbar'
import { DataTable } from './components/DataTable'
import { checkAuth, fetchData, login, logout, syncData } from './api'
import type { DataResponse, User } from './types'

const PAGE_SIZE = 20

const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches
const theme = createTheme({ palette: { mode: prefersDark ? 'dark' : 'light' } })

function App() {
  const [user, setUser] = useState<User | null>(null)
  const [data, setData] = useState<DataResponse | null>(null)
  const [from, setFrom] = useState(1)
  const [to, setTo] = useState(PAGE_SIZE)
  const [search, setSearch] = useState('')
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [syncing, setSyncing] = useState(false)

  async function loadData(reqFrom: number, reqTo: number, reqSearch: string) {
    setLoading(true)
    setError(null)
    try {
      const params = new URLSearchParams({ from: String(reqFrom), to: String(reqTo) })
      if (reqSearch) params.set('search', reqSearch)
      const result = await fetchData(params)
      if (result) setData(result)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    checkAuth().then(setUser).catch(() => {})
    loadData(1, PAGE_SIZE, '')
  }, [])

  async function handleLogin(username: string, password: string) {
    await login({ username, password })
    const u = await checkAuth()
    setUser(u)
  }

  async function handleLogout() {
    await logout()
    setUser(null)
  }

  async function handleSync() {
    setSyncing(true)
    setError(null)
    try {
      await syncData()
      await loadData(from, to, search)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Sync failed')
    } finally {
      setSyncing(false)
    }
  }

  function handleSearch(value: string) {
    setSearch(value)
    setFrom(1)
    setTo(PAGE_SIZE)
    loadData(1, PAGE_SIZE, value)
  }

  function handleRefresh() {
    loadData(from, to, search)
  }

  function handlePrevious() {
    const newFrom = Math.max(1, from - PAGE_SIZE)
    const newTo = newFrom + PAGE_SIZE - 1
    setFrom(newFrom)
    setTo(newTo)
    loadData(newFrom, newTo, search)
  }

  function handleNext() {
    const newFrom = from + PAGE_SIZE
    const newTo = newFrom + PAGE_SIZE - 1
    setFrom(newFrom)
    setTo(newTo)
    loadData(newFrom, newTo, search)
  }

  function handlePageChange(page: number) {
    const newFrom = (page - 1) * PAGE_SIZE + 1
    const newTo = newFrom + PAGE_SIZE - 1
    setFrom(newFrom)
    setTo(newTo)
    loadData(newFrom, newTo, search)
  }

  const total = data?.total_organisations ?? 0
  const totalPages = Math.ceil(total / PAGE_SIZE)
  const currentPage = Math.ceil(from / PAGE_SIZE)
  const atStart = from <= 1
  const atEnd = to >= total

  return (
    <ThemeProvider theme={theme}>
      <CssBaseline />
      <Header user={user} onLogin={handleLogin} onLogout={handleLogout} />
      <Container maxWidth="xl" sx={{ mt: 2 }}>
        <Box sx={{ display: 'flex', gap: 1, alignItems: 'center', mb: 2 }}>
          <SearchBar loading={loading} onSearch={handleSearch} onRefresh={handleRefresh} />
          {user?.role === 10 && <AdminToolbar onSync={handleSync} syncing={syncing} />}
        </Box>
        {error && <Alert severity="error" sx={{ mb: 2 }}>{error}</Alert>}
        {data && (
          <DataTable
            data={data}
            loading={loading}
            currentPage={currentPage}
            totalPages={totalPages}
            atStart={atStart}
            atEnd={atEnd}
            onPrevious={handlePrevious}
            onNext={handleNext}
            onPageChange={handlePageChange}
          />
        )}
      </Container>
    </ThemeProvider>
  )
}

export default App
