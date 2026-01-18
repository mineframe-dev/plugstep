package utils

import (
	"net/http"
	"time"
)

// HTTPClient is a shared HTTP client with a 30-second timeout for API calls.
var HTTPClient = &http.Client{
	Timeout: 30 * time.Second,
}

// DownloadClient is a shared HTTP client with a 5-minute timeout for file downloads.
var DownloadClient = &http.Client{
	Timeout: 5 * time.Minute,
}
