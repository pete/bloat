package mastodon

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

type lr struct {
	io.ReadCloser
	r *http.Request
}

func (r *lr) Read(p []byte) (n int, err error) {
	n, err = r.ReadCloser.Read(p)
	// override the generic error returned by the MaxBytesReader
	if _, ok := err.(*http.MaxBytesError); ok {
		err = fmt.Errorf("%s \"%s\": response body too large", r.r.Method, r.r.URL)
	}
	return
}

type transport struct {
	t http.RoundTripper
}

func (t *transport) RoundTrip(r *http.Request) (*http.Response, error) {
	resp, err := t.t.RoundTrip(r)
	if resp != nil && resp.Body != nil {
		resp.Body = &lr{http.MaxBytesReader(nil, resp.Body, 8<<20), r}
	}
	return resp, err
}

var httpClient = &http.Client{
	Transport: &transport{
		t: http.DefaultTransport,
	},
	Timeout: 30 * time.Second,
}
