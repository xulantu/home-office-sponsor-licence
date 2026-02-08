package csvfetch

import (
	"net/http"
	"time"
)

// DefaultHTTPTimeout is the default timeout for HTTP requests
const DefaultHTTPTimeout = 30 * time.Second

// httpClient is a shared HTTP client with timeout configured.
// Using a shared client enables connection reuse.
var httpClient = &http.Client{
	Timeout: DefaultHTTPTimeout,
}
