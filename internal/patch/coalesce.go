package patch

import (
	"sync"
)

// RequestCoalescer prevents duplicate in-flight requests
type RequestCoalescer struct {
	mu    sync.Mutex
	calls map[string]*call
}

type call struct {
	wg  sync.WaitGroup
	val interface{}
	err error
}

func NewRequestCoalescer() *RequestCoalescer {
	return &RequestCoalescer{
		calls: make(map[string]*call),
	}
}

// Do executes the given function only once for the given key at a time.
// If multiple goroutines call Do with the same key, only one will execute fn,
// and the others will wait for the result.
func (c *RequestCoalescer) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	c.mu.Lock()
	if existing, ok := c.calls[key]; ok {
		// Another request is in flight, wait for it
		c.mu.Unlock()
		existing.wg.Wait()
		return existing.val, existing.err
	}

	// Create new call
	newCall := &call{}
	newCall.wg.Add(1)
	c.calls[key] = newCall
	c.mu.Unlock()

	// Execute the function
	newCall.val, newCall.err = fn()

	// Clean up and notify waiters
	c.mu.Lock()
	delete(c.calls, key)
	c.mu.Unlock()
	newCall.wg.Done()

	return newCall.val, newCall.err
}

// Global coalescer for version API calls
var versionCoalescer = NewRequestCoalescer()
