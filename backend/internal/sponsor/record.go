package sponsor

// Record represents a single sponsor licence entry from the Home Office CSV
type Record struct {
	OrganisationName string
	TownCity         string
	County           string
	LicenceType      string // "Worker" or "Temporary Worker"
	Rating           string // "A rating" or "B rating"
	Route            string // "Skilled Worker", "Creative Worker", etc.
}
