// Code generated with github.com/kolosovi/periodic, DO NOT EDIT

package caches

import (
	usertype "example/types"
	"fmt"
	"sync/atomic"
	"time"
	"unsafe"
)

type FooCache struct {
	data   *unsafe.Pointer
	stopCh chan struct{}
	doneCh chan struct{}

	provider FooValueProvider
	tracer   FooTracer
	ticks    <-chan time.Time
}

type FooOptions struct {
	Provider FooValueProvider
	Tracer   FooTracer
	Ticks    <-chan time.Time
}

type FooValueProvider interface {
	NewValue() (usertype.Foo, error)
}

type FooTracer interface {
	OnUpdateError(err error)
	OnUpdateSuccess()
}

func NewFooCache(options FooOptions) (*FooCache, error) {
	if options.Provider == nil {
		return nil, fmt.Errorf("provider can't be nil")
	}
	if options.Ticks == nil {
		return nil, fmt.Errorf("ticks can't be nil")
	}
	ptr := unsafe.Pointer(nil)
	return &FooCache{
		data:   &ptr,
		stopCh: make(chan struct{}),
		doneCh: make(chan struct{}),

		provider: options.Provider,
		tracer:   options.Tracer,
		ticks:    options.Ticks,
	}, nil
}

func (c *FooCache) Get() (usertype.Foo, bool) {
	ptr := atomic.LoadPointer(c.data)
	userTypePtr := (*usertype.Foo)(ptr)
	if userTypePtr == nil {
		var fallback usertype.Foo
		return fallback, false
	}
	return *userTypePtr, true
}

func (c *FooCache) Update() error {
	newValue, err := c.provider.NewValue()
	if err != nil {
		return fmt.Errorf("cannot update cache: %v", err)
	}
	atomic.StorePointer(c.data, unsafe.Pointer(&newValue))
	return nil
}

func (c *FooCache) Close() {
	close(c.stopCh)
}

func (c *FooCache) Start() {
	go c.run()
}

func (c *FooCache) run() {
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

func (c *FooCache) onUpdateError(err error) {
	if c.tracer == nil {
		return
	}
	c.tracer.OnUpdateError(err)
}

func (c *FooCache) onUpdateSuccess() {
	if c.tracer == nil {
		return
	}
	c.tracer.OnUpdateSuccess()
}

func (c *FooCache) Done() <-chan struct{} {
	return c.doneCh
}
