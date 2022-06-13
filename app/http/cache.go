package http

import (
	"bufio"
	"bytes"
	"log"
	"net/http"
	"net/http/httputil"
	"sync"
)

// MemoryCache is an implemtation of Cache that stores responses in an in-memory map.
type MemoryCache struct {
	mu    sync.RWMutex
	Responses map[string][]byte
}

func (c *MemoryCache) Get(key string) (resp *http.Response, ok bool) {
	c.mu.RLock()

	byt, ok := c.Responses[key]
	if !ok {
		c.mu.RUnlock()

		return nil, false
	}

	b := bytes.NewBuffer(byt)
	x, err := http.ReadResponse(bufio.NewReader(b), nil)

	c.mu.RUnlock()

	if err != nil {
		log.Println(err)
		return nil, false
	}
	return x, true

}

// Set saves response resp to the cache with key
func (c *MemoryCache) Set(key string, response *http.Response) {
	c.mu.Lock()

	respBytes, err := httputil.DumpResponse(response, true)
	if err != nil {
		log.Println(err)
	}

	c.Responses[key] = respBytes
	c.mu.Unlock()
}

// Delete removes key from the cache
func (c *MemoryCache) Delete(key string) {
	c.mu.Lock()
	delete(c.Responses, key)
	c.mu.Unlock()
}

// NewMemoryCache returns a new Cache that will store items in an in-memory map
func NewMemoryCache() MemoryCache {
	c := MemoryCache{Responses: map[string][]byte{}}
	return c
}
