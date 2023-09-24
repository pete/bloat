package mastodon

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

type lr struct {
	io.ReadCloser
	n int64
	r *http.Request
}

func (r *lr) Read(p []byte) (n int, err error) {
	if r.n <= 0 {
		return 0, fmt.Errorf("%s \"%s\": response body too large", r.r.Method, r.r.URL)
	}
	if int64(len(p)) > r.n {
		p = p[0:r.n]
	}
	n, err = r.ReadCloser.Read(p)
	r.n -= int64(n)
	return
}

type transport struct {
	t http.RoundTripper
}

func (t *transport) RoundTrip(r *http.Request) (*http.Response, error) {
	resp, err := t.t.RoundTrip(r)
	if resp != nil && resp.Body != nil {
		resp.Body = &lr{resp.Body, 8 << 20, r}
	}
	return resp, err
}

var httpClient = &http.Client{
	Transport: &transport{
		t: http.DefaultTransport,
	},
	Timeout: 30 * time.Second,
}
