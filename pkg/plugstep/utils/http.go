package utils

import (
	"io"
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

// ThrottleBytesPerSecond limits download speed when > 0 (for testing).
var ThrottleBytesPerSecond int

// SetThrottledTransport configures the DownloadClient to use a throttled transport.
func SetThrottledTransport(bytesPerSecond int) {
	ThrottleBytesPerSecond = bytesPerSecond
	DownloadClient.Transport = &throttledTransport{
		base:           http.DefaultTransport,
		bytesPerSecond: bytesPerSecond,
	}
}

type throttledTransport struct {
	base           http.RoundTripper
	bytesPerSecond int
}

func (t *throttledTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := t.base.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	resp.Body = &throttledReader{
		reader:         resp.Body,
		bytesPerSecond: t.bytesPerSecond,
	}
	return resp, nil
}

type throttledReader struct {
	reader         io.ReadCloser
	bytesPerSecond int
	lastRead       time.Time
}

func (r *throttledReader) Read(p []byte) (int, error) {
	// Limit chunk size based on throttle rate
	chunkSize := r.bytesPerSecond / 10 // 100ms worth of data
	if chunkSize < 1024 {
		chunkSize = 1024
	}
	if len(p) > chunkSize {
		p = p[:chunkSize]
	}

	n, err := r.reader.Read(p)

	if n > 0 && r.bytesPerSecond > 0 {
		expectedDuration := time.Duration(float64(n) / float64(r.bytesPerSecond) * float64(time.Second))
		time.Sleep(expectedDuration)
	}

	return n, err
}

func (r *throttledReader) Close() error {
	return r.reader.Close()
}
