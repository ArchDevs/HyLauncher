package patch

import (
	"sync"
)

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

func (c *RequestCoalescer) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	c.mu.Lock()
	if existing, ok := c.calls[key]; ok {
		c.mu.Unlock()
		existing.wg.Wait()
		return existing.val, existing.err
	}

	newCall := &call{}
	newCall.wg.Add(1)
	c.calls[key] = newCall
	c.mu.Unlock()

	newCall.val, newCall.err = fn()

	c.mu.Lock()
	delete(c.calls, key)
	c.mu.Unlock()
	newCall.wg.Done()

	return newCall.val, newCall.err
}

var versionCoalescer = NewRequestCoalescer()
