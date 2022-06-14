package http

import (
	"log"
	"net/http"
)

type Transport struct {
	// The RoundTripper interface actually used to make requests
	// If nil, http.DefaultTransport is used
	Transport http.RoundTripper
	Cache     *MemoryCache
}

// NewTransport returns a new Transport with the
// provided Cache implementation and MarkCachedResponses set to true
func NewTransport(cache *MemoryCache) *Transport {
	return &Transport{Cache: cache}
}

func cacheKey(req *http.Request) string {
	return req.Method + " " + req.URL.String()
}

// RoundTrip will look at the etags to limit the API usage
func (t *Transport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	transport := t.Transport
	if transport == nil {
		transport = http.DefaultTransport
	}

	if req.Method == "GET" {
		previous, ok := t.Cache.Get(cacheKey(req))
		if ok {
			etag := previous.Header.Get("etag")
			if etag != "" && req.Header.Get("etag") == "" {
				req.Header.Set("if-none-match", etag)
			}
			lastModified := previous.Header.Get("last-modified")
			if lastModified != "" && req.Header.Get("last-modified") == "" {
				req.Header.Set("if-modified-since", lastModified)
			}

			r, e := transport.RoundTrip(req)

			if e != nil {
				log.Println(e)
				return nil, e
			}
			if req.Method == "GET" && r.StatusCode == http.StatusNotModified {
				log.Println("Cached " + req.URL.String())
				return previous, nil
			} else {
				log.Println("Not Cached " + req.URL.String())
				t.Cache.Set(cacheKey(req), r)
				return r, nil
			}

		} else {
			log.Println("Not Cached: " + req.URL.String())
			r, e := transport.RoundTrip(req)
			if e != nil {
				return r, e
			}
			t.Cache.Set(cacheKey(req), r)
			return r, e
		}

	}

	return transport.RoundTrip(req)
}
