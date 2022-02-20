package periodic

import (
	"fmt"
	"sync/atomic"
	"unsafe"
)

type Cache struct {
	data *unsafe.Pointer
	provider ValueProvider
}

type ValueProvider interface {
	NewValue() (string, error)
}

func NewCache(provider ValueProvider) *Cache {
	ptr := unsafe.Pointer(nil)
	return &Cache{
		data: &ptr,
		provider: provider,
	}
}

func (c *Cache) Get() (string, bool) {
	ptr := atomic.LoadPointer(c.data)
	stringPtr := (*string)(ptr)
	if stringPtr == nil {
		var fallback string
		return fallback, false
	}
	return *stringPtr, true
}

func (c *Cache) Update() error {
	newValue, err := c.provider.NewValue()
	if err != nil {
		return fmt.Errorf("cannot update cache: %v", err)
	}
	atomic.StorePointer(c.data, unsafe.Pointer(&newValue))
	return nil
}
