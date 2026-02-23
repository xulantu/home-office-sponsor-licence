package csvfetch

import (
	"net/http"
	"time"
)

// discoveryClient is used for the gov.uk discovery page — small, fast response.
var discoveryClient = &http.Client{
	Timeout: 30 * time.Second,
}

// csvClient is used for the CSV download — large file, slower transfer.
var csvClient = &http.Client{
	Timeout: 30 * time.Minute,
}
