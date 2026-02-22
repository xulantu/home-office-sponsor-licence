export interface Organisation {
  ID: number
  Name: string
  TownCity: string
  County: string
  CreatedAt: string | null
}

export interface Licence {
  ID: number
  OrganisationID: number
  LicenceType: string
  Rating: string
  Route: string
  ValidFrom: string | null
}

export interface DataResponse {
  initial_run_time: string
  total_organisations: number
  from: number
  to: number
  organisations: Organisation[]
  licences: Licence[]
}

export interface User {
  id: number
  username: string
  role: number
}

export interface LoginRequest {
  username: string
  password: string
}
