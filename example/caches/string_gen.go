// Code generated with github.com/kolosovi/periodic, DO NOT EDIT

package caches

import (
	"fmt"
	"sync/atomic"
	"time"
	"unsafe"
)

type StringCache struct {
	data   *unsafe.Pointer
	stopCh chan struct{}
	doneCh chan struct{}

	provider StringValueProvider
	tracer   StringTracer
	ticks    <-chan time.Time
}

type StringOptions struct {
	Provider StringValueProvider
	Tracer   StringTracer
	Ticks    <-chan time.Time
}

type StringValueProvider interface {
	NewValue() (string, error)
}

type StringTracer interface {
	OnUpdateError(err error)
	OnUpdateSuccess()
}

func NewStringCache(options StringOptions) (*StringCache, error) {
	if options.Provider == nil {
		return nil, fmt.Errorf("provider can't be nil")
	}
	if options.Ticks == nil {
		return nil, fmt.Errorf("ticks can't be nil")
	}
	ptr := unsafe.Pointer(nil)
	return &StringCache{
		data:   &ptr,
		stopCh: make(chan struct{}),
		doneCh: make(chan struct{}),

		provider: options.Provider,
		tracer:   options.Tracer,
		ticks:    options.Ticks,
	}, nil
}

func (c *StringCache) Get() (string, bool) {
	ptr := atomic.LoadPointer(c.data)
	userTypePtr := (*string)(ptr)
	if userTypePtr == nil {
		var fallback string
		return fallback, false
	}
	return *userTypePtr, true
}

func (c *StringCache) Update() error {
	newValue, err := c.provider.NewValue()
	if err != nil {
		return fmt.Errorf("cannot update cache: %v", err)
	}
	atomic.StorePointer(c.data, unsafe.Pointer(&newValue))
	return nil
}

func (c *StringCache) Close() {
	close(c.stopCh)
}

func (c *StringCache) Start() {
	go c.run()
}

func (c *StringCache) run() {
	for {
		select {
		case <-c.ticks:
			err := c.Update()
			if err != nil {
				c.onUpdateError(err)
			} else {
				c.onUpdateSuccess()
			}
		case <-c.stopCh:
			close(c.doneCh)
			return
		}
	}
}

func (c *StringCache) onUpdateError(err error) {
	if c.tracer == nil {
		return
	}
	c.tracer.OnUpdateError(err)
}

func (c *StringCache) onUpdateSuccess() {
	if c.tracer == nil {
		return
	}
	c.tracer.OnUpdateSuccess()
}

func (c *StringCache) Done() <-chan struct{} {
	return c.doneCh
}
